package model

type SendPacketRequest struct {
	packet  IpPacket
	nextHop VirtualIp
}

func MakeSendPacketRequest(packet IpPacket, nextHop VirtualIp) SendPacketRequest {
	return SendPacketRequest{packet, nextHop}
}

func (r *SendPacketRequest) Packet() IpPacket {
	return r.packet
}
func (r *SendPacketRequest) NextHop() VirtualIp {
	return r.nextHop
}
