package model

type SendPacketRequest struct {
	packet IpPacket
	dest   VirtualIp
}

func MakeSendPacketRequest(packet IpPacket, dest VirtualIp) SendPacketRequest {
	return SendPacketRequest{packet, dest}
}

func (r *SendPacketRequest) Packet() IpPacket {
	return r.packet
}
func (r *SendPacketRequest) Dest() VirtualIp {
	return r.dest
}
