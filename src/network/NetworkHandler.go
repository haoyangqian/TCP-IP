package network

import "../model"

type NetworkHandler interface {
	Handle(packet model.IpPacket)
}
