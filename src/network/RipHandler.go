package network

import "../model"

type RipHandler struct {
	routingTable model.RoutingTable
}

func MakeRipHandler(routingTable model.RoutingTable) RipHandler {
	return RipHandler{routingTable}
}

func (handler *RipHandler) Handle(packet model.IpPacket) {
}

func (handler *RipHandler) BroadcastAllRoutes(messageChannel chan<- model.SendMessageRequest) {
}

func (handler *RipHandler) BroadcastUpdatedRoutes(messageChannel chan<- model.SendMessageRequest) []model.RipMessage {
	return make([]model.RipMessage, 0)
}

func (handler *RipHandler) ExpireRoutes() []model.RipMessage {
	return make([]model.RipMessage, 0)
}

func (handler *RipHandler) UpdateRandom() {
	entry, _ := handler.routingTable.GetEntry(model.VirtualIp{"192.168.0.6"})
	entry.Cost = 8
}