package network

import (
	"../model"
	"../util"
	"errors"
	//"fmt"
	"math"
	//"time"
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
	if command == 1 {
		handler.handleRipRequest(ripInfo, receivedFromIp, handler.messageChannel)
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

/*
function :  handle Ripinfo Request, send all RIP entries to the requester
parameter:  ripInfo,requester,messageChannel
return   :  NULL
*/
func (handler *RipHandler) handleRipRequest(ripInfo model.RipInfo, requester model.VirtualIp, messageChannel chan<- model.SendMessageRequest) {
	neighbor := make([]model.VirtualIp, 0)
	neighbor = append(neighbor, requester)
	routingentries := handler.routingTable.GetAllEntries()
	handler.SendRoutesTo(neighbor, routingentries, messageChannel, 2)
}

func (handler *RipHandler) handleRipResponse(ripInfo model.RipInfo, selfIp model.VirtualIp, receivedFromIp model.VirtualIp) {
	err := handler.validateRipInfo(ripInfo, receivedFromIp)
	if err != nil {
		return
	}

	for _, ripEntry := range ripInfo.Entries {

		newCost := int(math.Min(float64(ripEntry.Cost+1), float64(model.RIP_INFINITY)))

		if !handler.routingTable.HasEntry(ripEntry.Address) {
			handler.addNewRoute(newCost, ripEntry.Address, selfIp, receivedFromIp)
		} else {
			// possible update of existing route
			// calculate new cost
			existingEntry, _ := handler.routingTable.GetEntry(ripEntry.Address)
			handler.updateExistingRoute(newCost, existingEntry, receivedFromIp)
		}
	}
}

func (handler *RipHandler) addNewRoute(newCost int, newAddress model.VirtualIp, selfIp model.VirtualIp, receivedFromIp model.VirtualIp) {
	if newCost >= model.RIP_INFINITY {
		return
	}

	if handler.routingTable.HasNeighbor(newAddress) {
		exitIpToNeighbor, _ := handler.routingTable.GetNeighbor(newAddress)
		selfIp = exitIpToNeighbor // override the selfIp from the neighbor table to avoid transient errors
	}

	new_entry := model.MakeRoutingEntry(newAddress, selfIp, receivedFromIp, newCost, false)
	handler.routingTable.PutEntry(&new_entry)
}

func (handler *RipHandler) updateExistingRoute(newCost int, existingEntry *model.RoutingEntry, receivedFromIp model.VirtualIp) {
	// update route if the new cost is lower
	if newCost < existingEntry.Cost {
		existingEntry.Update(newCost, receivedFromIp)
		return 
	}

	// decide whether to extend or expire a route learned from the same node
	if !existingEntry.Expired() && existingEntry.NextHop == receivedFromIp {
		if newCost == existingEntry.Cost {
			existingEntry.ExtendTtl()
		}

		if newCost >= model.RIP_INFINITY {
			existingEntry.MarkAsExpired()
		}
	}
}

func (handler *RipHandler) validateRipInfo(ripInfo model.RipInfo, receivedFromIp model.VirtualIp) error {
	if !handler.routingTable.HasNeighbor(receivedFromIp) {
		return errors.New("RIP response received is not from a neighbor")
	}

	return nil
}

/*
function :  Send Routing Eentries to neighbors
parameter:  neighbors, routingentries, messageChannel, command
return   :  NULL
*/
func (handler *RipHandler) SendRoutesTo(neighbors []model.VirtualIp, routingentries []*model.RoutingEntry, messageChannel chan<- model.SendMessageRequest, command int) {
	for _, v := range neighbors {
		ripentries := make([]model.RipEntry, 0)
		if command == 2 {
			/*if command is response*/
			//Check routingentries if empty
			if len(routingentries) == 0 {
				return
			}
			//Check every entry for poison reverse
			for k, routingv := range routingentries {
				cost := routingv.Cost
				if v == routingv.NextHop {
					cost = model.RIP_INFINITY
				}
				routingentries[k].SetIsUpdated(false)
				ripentry := model.RipEntry{cost, routingv.Dest}
				ripentries = append(ripentries, ripentry)
				//Every 64 entry warp a rip info and send it to channel
				if (k+1)%64 == 0 || k == len(routingentries)-1 {
					ripinfo, err := RoutingEntries2RipInfo(ripentries, command)
					util.CheckError(err)
					message, err := ripinfo.Marshal()
					util.CheckError(err)
					messageChannel <- model.MakeSendMessageRequest(message, model.RIP_PROTOCOL, v)
				}
			}
		} else if command == 1 {
			/*if command is request*/
			ripinfo, err := RoutingEntries2RipInfo(ripentries, command)
			util.CheckError(err)
			message, err := ripinfo.Marshal()
			util.CheckError(err)
			messageChannel <- model.MakeSendMessageRequest(message, model.RIP_PROTOCOL, v)
		}
	}
}

func (handler *RipHandler) BroadcastRoutes(routingentries []*model.RoutingEntry, messageChannel chan<- model.SendMessageRequest, command int) {
	neighbors := handler.routingTable.GetAllNeighbors()
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
			route.MarkAsExpired()
			continue
		}

		if route.Expired() && route.ShouldGC() {
			handler.routingTable.DeleteEntry(route)
		}
	}
}
