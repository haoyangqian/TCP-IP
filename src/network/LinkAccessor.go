package network

import "../model"

type LinkAccessor struct {
	// interface table
}

func (accessor *LinkAccessor) Send(packet model.IpPacket, dest model.VirtualIp) {
	// to be implemented
}

func (accessor *LinkAccessor) Receive() model.IpPacket {
	// to be implemented
	return model.IpPacket{}
}
