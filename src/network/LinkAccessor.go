package network

import "../model"

type LinkAccessor struct {
	interfaces_table map[model.VirtualIp]model.NodeInterface
}

func (accessor *LinkAccessor) Send(packet model.IpPacket, dest model.VirtualIp) {
	// to be implemented
}

func (accessor *LinkAccessor) Receive() model.IpPacket {
	// to be implemented
	return model.IpPacket{}
}

func MakeLinkAccessor(table map[model.VirtualIp]model.NodeInterface) LinkAccessor {
	return LinkAccessor(table)
}
