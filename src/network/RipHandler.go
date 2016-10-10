package network

import (
	"../model"
	"../util"
	"errors"
	"fmt"
	"math"
)

type RipHandler struct {
	routingTable model.RoutingTable
}

func MakeRipHandler(routingTable model.RoutingTable) RipHandler {
	return RipHandler{routingTable}
}

func (handler RipHandler) Handle(packet model.IpPacket) {
	ripInfo, _ := model.UnmarshalToInfo(packet.Payload)

	command := ripInfo.Command

	selfIp := model.VirtualIp{packet.Ipheader.Dst.String()}
	receivedFromIp := model.VirtualIp{packet.Ipheader.Src.String()}

	//fmt.Println("Selfip:", selfIp)
	//fmt.Println("recevfrom:", receivedFromIp)
	if command == 1 {
		handler.handleRipRequest(ripInfo, receivedFromIp)
	} else if command == 2 {
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

func RoutingEntries2RipInfo(routingEntries []*model.RoutingEntry, command int) (model.RipInfo, error) {
	if command != 1 && command != 2 {
		return model.RipInfo{}, errors.New("Wrong command type")
	}

	if command == 1 {
		emtpyentries := make([]model.RipEntry, 0)
		ripinfo := model.RipInfo{command, 0, emtpyentries}
		return ripinfo, nil
	}
	if len(routingEntries) == 0 {
		return model.RipInfo{}, errors.New("Empty routingEntries")
	}
	emtpyentries := make([]model.RipEntry, 0)
	ripinfo := model.RipInfo{command, 0, emtpyentries}
	for _, v := range routingEntries {
		ripinfo.AddEntry(model.RipEntry{v.Cost, v.Dest})
	}
	return ripinfo, nil
}

func (handler *RipHandler) handleRipRequest(ripInfo model.RipInfo, requester model.VirtualIp) {
	return
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
				fmt.Println("updaing existing entry")
				existing_entry.Update(new_cost, receivedFromIp)
			}

			// expire routes if the new cost is inifinity and the existing route is not marked as expired
			if new_cost >= model.RIP_INFINITY && !existing_entry.Expired() {
				existing_entry.MarkAsExpired()
			}
		} else {
			// func MakeRoutingEntry(dst VirtualIp, exitIp VirtualIp, nextHop VirtualIp, cost int) RoutingEntry
			new_entry := model.MakeRoutingEntry(ripEntry.Address, selfIp, receivedFromIp, new_cost)
			handler.routingTable.PutEntry(&new_entry)
		}
	}
}

func (handler *RipHandler) validateRipInfo(ripInfo model.RipInfo) {
	// TODO: do some basic validations here
}

func (handler *RipHandler) BroadcastAllRoutes(messageChannel chan<- model.SendMessageRequest) {
	ripinfo, err := RoutingEntries2RipInfo(handler.routingTable.GetAllEntries(), 2)
	util.CheckError(err)
	neighbors := handler.routingTable.GetAllNeighbors()

	for _, v := range neighbors {
		//check learned from
		//fmt.Println(v)
		message, err := ripinfo.Marshal()
		util.CheckError(err)
		messageChannel <- model.MakeSendMessageRequest(message, model.RIP_PROTOCOL, v)
	}
}

func (handler *RipHandler) BroadcastUpdatedRoutes(messageChannel chan<- model.SendMessageRequest) []model.RipEntry {
	return make([]model.RipEntry, 0)
}

func (handler *RipHandler) ExpireRoutes() {
	routes := handler.routingTable.GetExpiredEntries()
	for _, route := range routes {
		if route.ShouldExpire() {
			route.MarkAsExpired()
			continue
		}

		if route.Expired() && route.ShouldGC() {
			fmt.Println("shoudl GC this entry now!!!!!!")
			handler.routingTable.DeleteEntry(route)
		}
	}
}

func (handler *RipHandler) UpdateRandom() {
	entry, _ := handler.routingTable.GetEntry(model.VirtualIp{"192.168.0.6"})
	entry.Cost = 8
}
