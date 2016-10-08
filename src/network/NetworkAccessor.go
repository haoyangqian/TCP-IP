package network

import "../model"
import "fmt"

type NetworkAccessor struct {
	routingTable model.RoutingTable
	handlers     map[int]NetworkHandler
}

func NewNetworkAccessor(routingTable model.RoutingTable) NetworkAccessor {
	handlers := make(map[int]NetworkHandler)
	return NetworkAccessor{routingTable, handlers}
}

func (accessor *NetworkAccessor) RegisterHandler(protocol int, handler NetworkHandler) {
	accessor.handlers[protocol] = handler
}

func (accessor *NetworkAccessor) ReceiveAndHandle(packet model.IpPacket, chToForward chan<- model.SendPacketRequest) {
	//fmt.Println("network receive")
	//fmt.Println("network received")

	if accessor.ShouldDropPacket(packet) {
		dropPacket(packet)
		return
	}

	atDestination, err := accessor.isAtDestination(packet)
	if err != nil {
		dropPacket(packet)
		return
	}

	if !atDestination {
		accessor.ForwardPacket(packet, chToForward)
		return
	}

	protocol := packet.Ipheader.Protocol
	if handler, ok := accessor.handlers[protocol]; ok {
		handler.Handle(packet)
	} else {
		dropPacket(packet)
		return
	}
}

func (accessor *NetworkAccessor) Send(request model.SendMessageRequest, chToLink chan<- model.SendPacketRequest) {
	//fmt.Println("network send test data")
	message := request.Message()
	protocol := request.Protocol()
	dest := request.Dest()

	entry, err := accessor.routingTable.GetEntry(dest)
	if err != nil {
		fmt.Println(request)
		fmt.Println("Cannot reach this destination!")
		return
	}

	packet := convertToIpPacket(message, protocol, entry.NextHop, dest, entry.Cost == 0)

	//fmt.Println("network data sent,NextHop: " + entry.NextHop.Ip)
	chToLink <- model.MakeSendPacketRequest(packet, entry.NextHop)
}

func (accessor *NetworkAccessor) ForwardPacket(packet model.IpPacket, chToForward chan<- model.SendPacketRequest) {
	entry, err := accessor.routingTable.GetEntry(model.VirtualIp{packet.Ipheader.Dst.String()})
	if err != nil {
		println(err)
		return
	}
	packet.Ipheader.TTL -= 1
	packet.Ipheader.Checksum = model.IpPacket.IpSum(packet.Ipheader)

	chToForward <- model.MakeSendPacketRequest(packet, entry.NextHop)
}

func (accessor *NetworkAccessor) isAtDestination(packet model.IpPacket) (bool, error) {
	destIpString := packet.Ipheader.Dst.String()
	dest := model.MakeVirtualIp(destIpString)

	entry, err := accessor.routingTable.GetEntry(dest)
	if err != nil {
		return false, err
	}

	// cost 0 implies that the desired destination is a local interface
	return entry.Cost == 0, nil
}

func (accessor *NetworkAccessor) ShouldDropPacket(packet model.IpPacket) bool {
	if checkSumMismatch(packet) {
		return true
	}

	if packet.Ipheader.TTL < 0 {
		return true
	}

	if !accessor.routingTable.HasEntry(model.VirtualIp{packet.Ipheader.Dst.String()}) {
		return true
	}

	return false
}

func checksumMismatch(packet model.IpPacket) bool {
	receivedSum = packet.Ipheader.Checksum
	packet.Ipheader.Checksum = 0

	return receivedSum == model.IpPacket.IpSum(packet.Ipheader)
}

func dropPacket(packet model.IpPacket) {
	// does nothing, simply drops the packet
	fmt.Println("invalid packet received: " + packet.IpPacketString())
}

func convertToIpPacket(message []byte, protocol int, src model.VirtualIp, dest model.VirtualIp, isToSelf bool) model.IpPacket {
	if isToSelf {
		src = model.VirtualIp{"0.0.0.0"}
	}
	return model.MakeIpPacket(message, protocol, src, dest)
}

// func (accessor *NetworkLayerAccessor) SendRipData(message model.RipMessage, dest model.VirtualIp) {
//     // to be implemented
// }=
