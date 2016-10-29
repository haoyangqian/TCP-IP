package network

import (
	"fmt"
	"model"
	"transport"
)

type IpHandler struct {
	manager transport.SocketManager
}

func (handler IpHandler) Handle(packet model.IpPacket, receivedFrom model.VirtualIp) {
	printPacketInfo(packet)

}

func printPacketInfo(packet model.IpPacket) {
	fmt.Println("driver received packet:")
	fmt.Println(packet.IpPacketString())
	tcppacket := transport.ConvertToTcpPacket(packet.Payload)
	fmt.Println(tcppacket.TcpPacketString())
	fmt.Print("> ")
}

func (handler IpHandler) handlePacket(packet model.IpPacket) {
	//get tcp packet from Ip payload
	tcppacket := transport.ConvertToTcpPacket(packet.Payload)
	//get tcpaddr from packet
	localAddr := model.Int2Vip(packet.Ipheader.Dst)
	localPort := tcppacket.Tcpheader.Destination
	remoteAddr := model.Int2Vip(packet.Ipheader.Src)
	remotePort := tcppacket.Tcpheader.Source
	//get socket from socket manager
	socket, err := handler.manager.GetSocketByAddr(transport.SocketAddr{localAddr, localPort, remoteAddr, remotePort})
	if err != nil {
		fmt.Println(err.Error())
	}
	socket.RecvCh <- packet
}
