package network

import "../model"

type RipHandler struct {
	routingTable model.RoutingTable

	messageChannel chan<- model.SendMessageRequest
}

func MakeRipHandler(routingTable model.RoutingTable, messageChannel chan<- model.SendMessageRequest) RipHandler {
	return RipHandler{routingTable, messageChannel}
}

func (handler *RipHandler) Handle(packet model.IpPacket) {
}

func (handler *RipHandler) BroadcastAllRoutes() {
}

func (handler *RipHandler) BroadcastUpdatedRoutes() []model.RipMessage {
	return make([]model.RipMessage, 0)
}

func (handler *RipHandler) ExpireRoutes() []model.RipMessage {
	return make([]model.RipMessage, 0)
}
