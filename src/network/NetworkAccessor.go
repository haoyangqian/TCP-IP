package network

import "model"
import "fmt"
import "logging"
import "transport"

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

func (accessor *NetworkAccessor) ReceiveAndHandle(result model.LinkReceiveResult, chToForward chan<- model.SendPacketRequest) {
	packet := result.Packet
	receivedFrom := result.ReceivedFrom
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
	//	if protocol != 200 {
	//		fmt.Println("Network accessor:\n", packet.IpPacketString())
	//	}
	if handler, ok := accessor.handlers[protocol]; ok {
		go handler.Handle(packet, receivedFrom)
	} else {
		fmt.Println("no handler")
		dropPacket(packet, "no handler found")
		return
	}
}

func (accessor *NetworkAccessor) Send(request model.SendMessageRequest, chToLink chan<- model.SendPacketRequest) {
	message := request.Message()
	protocol := request.Protocol()
	dest := request.Dest()

	var nextHop model.VirtualIp
	var exitIp model.VirtualIp
	var packet model.IpPacket

	toSelf := false
	localDelivery := false

	if accessor.routingTable.HasEntry(dest) {
		entry, _ := accessor.routingTable.GetEntry(dest)
		nextHop = entry.NextHop
		toSelf = entry.IsLocal
		localDelivery = entry.Cost == 0
		exitIp = entry.ExitIp
	} else if accessor.routingTable.HasNeighbor(dest) {
		nextHop = dest
		exitIp, _ = accessor.routingTable.GetNeighbor(dest)
	} else {
		return
	}

	if toSelf && !localDelivery {
		// the packet is intended for the local node but the destination VIP is unreachable
		// this might happen if a interface is down but other nodes managed to send in requests before their routes were updated
		dropPacket(packet, "A packet is inteded for a local VIP but that VIP is not reachable locally")
		return
	}

	// override the source of the outgoing IP packet if the request has a specified source
	if request.HasSrc() {
		exitIp = request.Src()
	}

	packet = convertToIpPacket(message, protocol, exitIp, dest, toSelf)

	// handle the packet locally if the packet was for a local VIP and the VIP is reachable
	if handler, ok := accessor.handlers[protocol]; ok && toSelf && localDelivery {
		go handler.Handle(packet, nextHop)
		return
	}

	// cannot handle locally, forwarding the packet
	chToLink <- model.MakeSendPacketRequest(packet, nextHop)
}

func (accessor *NetworkAccessor) ForwardPacket(packet model.IpPacket, chToForward chan<- model.SendPacketRequest) {
	entry, err := accessor.routingTable.GetEntry(model.VirtualIp{packet.Ipheader.Dst.String()})
	if err != nil {
		println(err)
		return
	}

	if !entry.Reachable() {
		return
	}

	if packet.Ipheader.Protocol == 6 {
		tcppacket := transport.ConvertToTcpPacket(packet.Payload)
		logging.Logger.Printf("[IpHandler][TcpPacket] Seqnum: %d, Acknum: %d, from %+v to %+v, window size: %d, payload size: %d, ctrlFlag: %b\n",
			tcppacket.Tcpheader.SeqNum, tcppacket.Tcpheader.AckNum, tcppacket.Tcpheader.Source, tcppacket.Tcpheader.Destination,
			tcppacket.Tcpheader.Window, len(tcppacket.Payload), tcppacket.Tcpheader.Ctrl)
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
		fmt.Printf("invalid packet received: because TTL < 0")
		return true, "TTL is less than 0"
	}

	if !accessor.routingTable.HasEntry(model.VirtualIp{packet.Ipheader.Dst.String()}) {
		return true, "destination " + packet.Ipheader.Dst.String() + " is not in the routign table"
	}

	routingEntry, _ := accessor.routingTable.GetEntry(model.VirtualIp{packet.Ipheader.Dst.String()})
	if !routingEntry.Reachable() {
		return true, "destionation " + packet.Ipheader.Dst.String() + " is in the routing table but not reachable"
	}

	return false, ""
}

func checksumMismatch(packet model.IpPacket) bool {
	receivedSum := packet.Ipheader.Checksum
	packet.Ipheader.Checksum = 0
	return receivedSum != model.IpSum(packet.Ipheader)
}

func dropPacket(packet model.IpPacket, reason string) {
	// does nothing, simply drops the packet
}

func convertToIpPacket(message []byte, protocol int, src model.VirtualIp, dest model.VirtualIp, isToSelf bool) model.IpPacket {
	//	if isToSelf {
	//		src = model.VirtualIp{"0.0.0.0"}
	//	}
	return model.MakeIpPacket(message, protocol, src, dest)
}
