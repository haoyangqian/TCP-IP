package network

import "../model"

type NetworkHandler interface {
	handle(packet model.IpPacket)
}
