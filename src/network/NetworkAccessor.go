package network

import "../model"
import "fmt"

const TEST_DATA_PROTOCOL = 0
const RIP_PROTOCOL = 200

type NetworkAccessor struct {
	linkAccessor LinkAccessor
	routingTable model.RoutingTable
	handlers     map[int]NetworkHandler
}

func NewNetworkAccessor(linkAccessor LinkAccessor, routingTable model.RoutingTable) NetworkAccessor {
	handlers := make(map[int]NetworkHandler)
	return NetworkAccessor{linkAccessor, routingTable, handlers}
}

func (accessor *NetworkAccessor) RegisterHandler(protocol int, handler NetworkHandler) {
	accessor.handlers[protocol] = handler
}

func (accessor *NetworkAccessor) ReceiveAndHandle() {
	//fmt.Println("network receive")
	packet := accessor.linkAccessor.Receive()
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

	if atDestination {
		protocol := packet.Ipheader.Protocol
		if handler, ok := accessor.handlers[protocol]; ok {
			handler.Handle(packet)
		} else {
			dropPacket(packet)
			return
		}
	} else {
		accessor.ForwardPacket(packet)
	}
}

func (accessor *NetworkAccessor) SendMessage(message string, protocol int, dest model.VirtualIp) {
	//fmt.Println("network send test data")
	entry, err := accessor.routingTable.GetEntry(dest)
	if err != nil {
		println("Cannot reach this destination!")
		return
	}

	packet := convertToIpPacket(message, protocol, entry.ExitIp, dest)
	accessor.linkAccessor.Send(packet, entry.ExitIp)
	//fmt.Println("network data sent")
}

func (accessor *NetworkAccessor) ForwardPacket(packet model.IpPacket) {
	entry, err := accessor.routingTable.GetEntry(model.VirtualIp{packet.Ipheader.Dst.String()})
	if err != nil {
		println(err)
		return
	}

	packet.Ipheader.TTL -= 1
	accessor.linkAccessor.Send(packet, entry.ExitIp)
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
	if !accessor.routingTable.HasEntry(model.VirtualIp{packet.Ipheader.Dst.String()}) {
		return true
	}

	if packet.Ipheader.TTL == 0 {
		return true
	}

	return false
}

func (accessor *NetworkAccessor) CloseConnection() {
	accessor.linkAccessor.CloseConnection()
}

func dropPacket(packet model.IpPacket) {
	// does nothing, simply drops the packet
	fmt.Println("invalid packet received: " + packet.IpPacketString())
}

func convertToIpPacket(message string, protocol int, src model.VirtualIp, dest model.VirtualIp) model.IpPacket {
	payload := []byte(message)
	return model.MakeIpPacket(payload, protocol, src, dest)
}

// func (accessor *NetworkLayerAccessor) SendRipData(message model.RipMessage, dest model.VirtualIp) {
//     // to be implemented
// }
