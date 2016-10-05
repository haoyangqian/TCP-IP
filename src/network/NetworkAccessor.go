package network

import "../model"

const FORWARD_PROTOCOL = 0
const RIP_PROTOCOL = 200

type NetworkAccessor struct {
	linkAccessor LinkAccessor
	routingTable model.RoutingTable
	handlers     map[int]NetworkHandler
}

func NewNetworkAccessor(linkAccessor LinkAccessor, routingTable model.RoutingTable, ipHandler NetworkHandler) NetworkAccessor {
	handlers := make(map[int]NetworkHandler)
	handlers[FORWARD_PROTOCOL] = ipHandler
	return NetworkAccessor{linkAccessor, routingTable, handlers}
}

func (accessor *NetworkAccessor) ReceiveAndHandle() {
	packet := accessor.linkAccessor.Receive()
	dest := model.VirtualIp{packet.Ipheader.Dst.String()}

	if !accessor.routingTable.HasEntry(dest) {
		// does not have routing for the destination, dropping packet
		return
	}

	atFinalDest, err := accessor.isAtFinalDest(packet)
	if err != nil {
		// handle error
	}

	if atFinalDest == true {
		protocol := packet.Ipheader.Protocol
		if handler, ok := accessor.handlers[protocol]; ok {
			handler.Handle(packet)
		} else {
			// invalid protocol, dropping packet
			return
		}
	} else {
		accessor.ForwardPacket(packet)
	}
}

func (accessor *NetworkAccessor) SendTestData(message string, src model.VirtualIp, dest model.VirtualIp) {
	packet := convertToIpPacket(message, FORWARD_PROTOCOL, src, dest)
	accessor.linkAccessor.Send(packet)
}

func (accessor *NetworkAccessor) ForwardPacket(packet model.IpPacket) {
	accessor.linkAccessor.Send(packet)
}

func convertToIpPacket(message string, protocol int, src model.VirtualIp, dest model.VirtualIp) model.IpPacket {
	payload := []byte(message)
	return model.MakeIpPacket(payload, protocol, src, dest)
}

func (accessor *NetworkAccessor) isAtFinalDest(packet model.IpPacket) (bool, error) {
	destIpString := packet.Ipheader.Dst.String()
	dest := model.MakeVirtualIp(destIpString)

	entry, err := accessor.routingTable.GetEntry(dest)
	if err != nil {
		return false, err
	}

	// cost 0 implies that the desired destination is a local interface
	return entry.Cost == 0, nil
}

// func (accessor *NetworkLayerAccessor) SendRipData(message model.RipMessage, dest model.VirtualIp) {
//     // to be implemented
// }
