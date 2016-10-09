package network

import (
	"../model"
	"../util"
	"fmt"
)

type RipHandler struct {
	routingTable model.RoutingTable
}

func MakeRipHandler(routingTable model.RoutingTable) RipHandler {
	return RipHandler{routingTable}
}

/*
function :  Convert routing table to RipInfo
            if command = 1,which indicates a request,set num = 0
            if command = 2,which indicates a response,iterate the routing table and add every entry to the RipInfo
parameter:   []RoutingEntry,command(int)
return   :   RipInfo
*/

func RoutingEntries2RipInfo(ripentries []RoutingEntry, command int) (RipInfo, error) {
	if command != 1 && command != 2 {
		return RipInfo{}, errors.New("Wrong command type")
	}

	if command == 1 {
        emtpyentries := make([]RipEntry, 0)
		ripinfo := RipInfo{command, 0, emtpyentries}
		return ripinfo, nil
	}

	if len(ripentries) == 0 {
		return RipInfo{}, errors.New("Empty ripentries")
	}
	emtpyentries := make([]RipEntry, 0)
	ripinfo := RipInfo{command, 0, emtpyentries}
	for _, v := range ripentries{
		ripinfo.AddEntry(RipEntry{v.Cost, v.Dest})
	}

	return ripinfo, nil
}

func (handler RipHandler) Handle(packet model.IpPacket) {
}

func (handler *RipHandler) BroadcastAllRoutes(messageChannel chan<- model.SendMessageRequest) {
    ripinfo RoutingTable2RipInfo(routingTable, 2)
}

func (handler *RipHandler) BroadcastUpdatedRoutes(messageChannel chan<- model.SendMessageRequest) []model.RipEntry {
	return make([]model.RipEntry, 0)
}

func (handler *RipHandler) ExpireRoutes() []model.RipEntry {
	return make([]model.RipEntry, 0)
}

func (handler *RipHandler) UpdateRandom() {
	entry, _ := handler.routingTable.GetEntry(model.VirtualIp{"192.168.0.6"})
	entry.Cost = 8
}
