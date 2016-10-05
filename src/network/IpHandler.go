package network

import "fmt"
import "../model"

type IpHandler struct {
}

func (handler IpHandler) Handle(packet model.IpPacket) {
	printPacketInfo(packet)
}

func printPacketInfo(packet model.IpPacket) {
	fmt.Println("driver received packet:")
	packet.IpPacketString()
}
