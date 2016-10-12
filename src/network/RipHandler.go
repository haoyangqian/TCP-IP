package network

import (
	"../model"
	"../util"
	"errors"
	// "fmt"
	// "time"
	"math"
)

type RipHandler struct {
	routingTable   model.RoutingTable
	messageChannel chan<- model.SendMessageRequest
}

func MakeRipHandler(routingTable model.RoutingTable, messageChannel chan<- model.SendMessageRequest) RipHandler {
	return RipHandler{routingTable, messageChannel}
}

func (handler RipHandler) Handle(packet model.IpPacket, receivedFrom model.VirtualIp) {
	ripInfo, _ := model.UnmarshalToInfo(packet.Payload)

	command := ripInfo.Command

	selfIp := model.VirtualIp{packet.Ipheader.Dst.String()}
	receivedFromIp := model.VirtualIp{packet.Ipheader.Src.String()}

	//fmt.Println("Selfip:", selfIp)
	//fmt.Println("recevfrom:", receivedFromIp)
	if command == 1 {
		//fmt.Println("Recv Request,recevfrom:", receivedFromIp)
		handler.handleRipRequest(ripInfo, receivedFromIp, handler.messageChannel)
	} else if command == 2 {
		//fmt.Println("Recv Response,recevfrom:", receivedFromIp)
		handler.handleRipResponse(ripInfo, selfIp, receivedFromIp)
	}
}

/*
function :  Convert routing table to RipInfo
            if command = 1,which indicates a request,set num = 0
            if command = 2,which indicates a response,iterate the routing table and add every entry to the RipInfo
parameter:   []RoutingEntry,command(int)
return   :   RipInfo
*/

func RoutingEntries2RipInfo(ripEntries []model.RipEntry, command int) (model.RipInfo, error) {
	if command != 1 && command != 2 {
		return model.RipInfo{}, errors.New("Wrong command type")
	}

	if command == 1 {
		emtpyentries := make([]model.RipEntry, 0)
		ripinfo := model.RipInfo{command, 0, emtpyentries}
		return ripinfo, nil
	}
	emtpyentries := make([]model.RipEntry, 0)
	ripinfo := model.RipInfo{command, 0, emtpyentries}
	for _, v := range ripEntries {
		ripinfo.AddEntry(v)
	}
	return ripinfo, nil
}

func (handler *RipHandler) handleRipRequest(ripInfo model.RipInfo, requester model.VirtualIp, messageChannel chan<- model.SendMessageRequest) {
	neighbor := make([]model.VirtualIp, 0)
	neighbor = append(neighbor, requester)
	//fmt.Println("send response to: ", requester)
	// fmt.Println("[RipHandler] handling RIP request")
	routingentries := handler.routingTable.GetAllEntries()
	handler.SendRoutesTo(neighbor, routingentries, messageChannel, 2)
	// fmt.Println("[RipHandler] RIP request handled")
	//fmt.Println("response sent to: ", requester)
	// entries := handler.routingTable.GetAllEntries()

	// // TODO: marshal
	// responseRipInfo, _ := marshal(entries, model.RIP_RESPONSE_COMMAND)
	// responsePayload := marshal(responseripInfo)

	// sendMessageRequest := model.MakeSendRequest(responsePayload, model.RIP_PROTOCOL, requester)
	// messageChannel <- packet
}

func (handler *RipHandler) handleRipResponse(ripInfo model.RipInfo, selfIp model.VirtualIp, receivedFromIp model.VirtualIp) {
	handler.validateRipInfo(ripInfo)

	for _, ripEntry := range ripInfo.Entries {

		new_cost := int(math.Min(float64(ripEntry.Cost+1), float64(model.RIP_INFINITY)))
		if handler.routingTable.HasEntry(ripEntry.Address) {
			// possible update of existing route

			// calculate new cost
			existing_entry, _ := handler.routingTable.GetEntry(ripEntry.Address)

			// update entry is new cost is cheaper
			if new_cost < existing_entry.Cost {
				existing_entry.Update(new_cost, receivedFromIp)
			}

			if new_cost == existing_entry.Cost && !existing_entry.Expired() {
				existing_entry.ExtendTtl()
			}

			// expire routes if the new cost is inifinity and the existing route is not marked as expired
			if new_cost >= model.RIP_INFINITY && !existing_entry.Expired() && existing_entry.NextHop == receivedFromIp {
				existing_entry.MarkAsExpired()
			}
		} else {
			// func MakeRoutingEntry(dst VirtualIp, exitIp VirtualIp, nextHop VirtualIp, cost int) RoutingEntry
			if new_cost < model.RIP_INFINITY {
				new_entry := model.MakeRoutingEntry(ripEntry.Address, selfIp, receivedFromIp, new_cost, false)
				// fmt.Println("trying to put a new entry")
				handler.routingTable.PutEntry(&new_entry)
				// fmt.Println("new entry added")
			}
		}
	}
}

func (handler *RipHandler) validateRipInfo(ripInfo model.RipInfo) {
	// TODO: do some basic validations here
}

func (handler *RipHandler) SendRoutesTo(neighbors []model.VirtualIp, routingentries []*model.RoutingEntry, messageChannel chan<- model.SendMessageRequest, command int) {
	for _, v := range neighbors {
		ripentries := make([]model.RipEntry, 0)
		/*if command is response*/
		if command == 2 {
			//check routingentries if empty
			if len(routingentries) == 0 {
				return
			}
			//check every entry for poison reverse
			for k, routingv := range routingentries {
				cost := routingv.Cost
				if v == routingv.NextHop {
					cost = model.RIP_INFINITY
					//fmt.Println("Set to infinity")
				}
				routingentries[k].SetIsUpdated(false)
				ripentry := model.RipEntry{cost, routingv.Dest}
				ripentries = append(ripentries, ripentry)
				//every 64 entry warp a rip info and send it to channel
				if (k+1)%64 == 0 || k == len(routingentries)-1 {
					ripinfo, err := RoutingEntries2RipInfo(ripentries, command)
					util.CheckError(err)
					message, err := ripinfo.Marshal()
					util.CheckError(err)
					// fmt.Println("[RipHandler] putting message into channel")
					messageChannel <- model.MakeSendMessageRequest(message, model.RIP_PROTOCOL, v)
					// fmt.Println("[RipHandler] message placed")
				}
			}
			//fmt.Println("ripentres: ", ripentries)
			/*if command is request*/
		} else if command == 1 {
			ripinfo, err := RoutingEntries2RipInfo(ripentries, command)
			util.CheckError(err)
			message, err := ripinfo.Marshal()
			util.CheckError(err)
			// fmt.Println("putting message into channel")
			messageChannel <- model.MakeSendMessageRequest(message, model.RIP_PROTOCOL, v)
			// fmt.Println("message placed")
		}
	}
}

func (handler *RipHandler) BroadcastRoutes(routingentries []*model.RoutingEntry, messageChannel chan<- model.SendMessageRequest, command int) {
	neighbors := handler.routingTable.GetAllNeighbors()
	//fmt.Println("neighbors:", neighbors)
	handler.SendRoutesTo(neighbors, routingentries, messageChannel, command)
}

func (handler *RipHandler) BroadcastRequest(messageChannel chan<- model.SendMessageRequest) {
	routingentries := make([]*model.RoutingEntry, 0)
	handler.BroadcastRoutes(routingentries, messageChannel, 1)

}
func (handler *RipHandler) BroadcastAllRoutes(messageChannel chan<- model.SendMessageRequest) {
	routingentries := handler.routingTable.GetAllEntries()
	handler.BroadcastRoutes(routingentries, messageChannel, 2)
}

func (handler *RipHandler) BroadcastUpdatedRoutes(messageChannel chan<- model.SendMessageRequest) {
	routingentries := handler.routingTable.GetUpdatedEntries()
	handler.BroadcastRoutes(routingentries, messageChannel, 2)
}

func (handler *RipHandler) ExpireRoutes() {
	routes := handler.routingTable.GetExpiredEntries()
	for _, route := range routes {
		if route.ShouldExpire() {
			// fmt.Printf("route to %s should expire ttl left is %d, expired %t shouldGc %t\n", route.Dest, time.Now().Unix() - route.Ttl, route.Expired(), route.ShouldGC())
			route.MarkAsExpired()
			continue
		}

		if route.Expired() && route.ShouldGC() {
			// fmt.Println("calling DeleteEntry")
			handler.routingTable.DeleteEntry(route)
			// fmt.Println("DeleteEntry called")
		}
	}
}

func (handler *RipHandler) UpdateRandom() {
	entry, _ := handler.routingTable.GetEntry(model.VirtualIp{"192.168.0.6"})
	entry.Cost = 8
}
