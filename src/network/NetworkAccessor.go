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

	shouldDrop, reason := accessor.ShouldDropPacket(packet)
	if shouldDrop {
		dropPacket(packet, reason)
		return
	}

	atDestination, err := accessor.isAtDestination(packet)
	if err != nil {
		dropPacket(packet, "error encountered while checking destination")
		return
	}

	if !atDestination {
		accessor.ForwardPacket(packet, chToForward)
		return
	}

	protocol := packet.Ipheader.Protocol
	if handler, ok := accessor.handlers[protocol]; ok {
		go handler.Handle(packet)
	} else {
		fmt.Println("no handler")
		dropPacket(packet, "no handler found")
		return
	}
	//fmt.Println("done handling")
}

func (accessor *NetworkAccessor) Send(request model.SendMessageRequest, chToLink chan<- model.SendPacketRequest) {
	//fmt.Println("network send test data")
	message := request.Message()
	protocol := request.Protocol()
	dest := request.Dest()
	var nextHop model.VirtualIp
	var toSelf = false
	var ExitIp model.VirtualIp

	if accessor.routingTable.HasEntry(dest) {
		entry, _ := accessor.routingTable.GetEntry(dest)
		nextHop = entry.NextHop
		toSelf = entry.Cost == 0
		ExitIp = entry.ExitIp
		// fmt.Printf("table has entry, sending to %s via exitIp %s, through next hop %s", dest, ExitIp, nextHop)
	} else if accessor.routingTable.HasNeighbor(dest) {
		//fmt.Println("hop neighbor")
		nextHop = dest
		ExitIp, _ = accessor.routingTable.GetNeighbor(dest)
		// fmt.Printf("neighbor found, sending to %s via exitIp %s, through next hop %s", dest, ExitIp, nextHop)
	} else {
		fmt.Println(request)
		fmt.Println("Cannot reach this destination!")
		return
	}

	packet := convertToIpPacket(message, protocol, ExitIp, dest, toSelf)

	//fmt.Println("network data sent,NextHop: " + entry.NextHop.Ip)
	chToLink <- model.MakeSendPacketRequest(packet, nextHop)
}

func (accessor *NetworkAccessor) ForwardPacket(packet model.IpPacket, chToForward chan<- model.SendPacketRequest) {
	entry, err := accessor.routingTable.GetEntry(model.VirtualIp{packet.Ipheader.Dst.String()})
	if err != nil {
		println(err)
		return
	}
	packet.Ipheader.TTL -= 1
	packet.Ipheader.Checksum = 0
	packet.Ipheader.Checksum = model.IpSum(packet.Ipheader)

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

func (accessor *NetworkAccessor) ShouldDropPacket(packet model.IpPacket) (bool, string) {
	if checksumMismatch(packet) {
		fmt.Println("checksum mismatch!")
		return true, "checksum mismatch"
	}

	if packet.Ipheader.TTL < 0 {
		return true, "TTL is less than 0"
	}

	if !accessor.routingTable.HasEntry(model.VirtualIp{packet.Ipheader.Dst.String()}) {
		return true, "destination " + packet.Ipheader.Dst.String() + "is not in the routign table"
	}

	return false, ""
}

func checksumMismatch(packet model.IpPacket) bool {
	receivedSum := packet.Ipheader.Checksum
	packet.Ipheader.Checksum = 0
	//fmt.Println("recv sum:", receivedSum)
	//fmt.Println("cal sum:", model.IpSum(packet.Ipheader))
	return receivedSum != model.IpSum(packet.Ipheader)
}

func dropPacket(packet model.IpPacket, reason string) {
	// does nothing, simply drops the packet
	fmt.Printf("invalid packet received: %s\n", reason)
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
