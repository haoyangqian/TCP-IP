package network

import (
	"fmt"
	"model"
	"transport"
)

type IpHandler struct {
	RecvCh        chan<- model.IpPacket
	SocketManager *transport.SocketManager
}

func (handler IpHandler) Handle(packet model.IpPacket, receivedFrom model.VirtualIp) {
	printPacketInfo(packet)
	handler.handlePacket(packet)

}

func printPacketInfo(packet model.IpPacket) {
	//fmt.Println("driver received packet:")
	//fmt.Println(packet.IpPacketString())
	tcppacket := transport.ConvertToTcpPacket(packet.Payload)
	tcppacket.PrintTcpPacketString()
	//fmt.Printf("ctrl info: %b\n", tcppacket.Tcpheader.Ctrl)
	//fmt.Print("> ")
}

func (handler IpHandler) handlePacket(ipPacket model.IpPacket) {
	localIp := model.Int2Vip(ipPacket.Ipheader.Dst)
	remoteIp := model.Int2Vip(ipPacket.Ipheader.Src)

	tcpPacket := transport.ConvertToTcpPacket(ipPacket.Payload)
	localPort := tcpPacket.Tcpheader.Destination
	remotePort := tcpPacket.Tcpheader.Source
	//check tcp checksum
	recvcheck := tcpPacket.Tcpheader.Checksum
	tcpbytes := ipPacket.Payload
	tcpbytes[16] = 0
	tcpbytes[17] = 0
	//fmt.Printf("tcpbytes length:%d payload:%s", len(tcpbytes), tcpbytes[21])
	//tcpPacket.Tcpheader.Checksum = 0
	//tcpbytes := tcpPacket.ConvertToBuffer()
	if len(tcpbytes) > 20 {
		fmt.Printf("tcpbytes length:%d payload:%s\n", len(tcpbytes), string(tcpbytes[21]))
	}
	calchecksum := transport.Csum(tcpbytes, remoteIp.Vip2Int(), localIp.Vip2Int())
	//fmt.Printf("receive ippacket in Ip layer, localIp: %s, localport : %d, remoteIp: %s , remoteport: %d\n", localIp, localPort, remoteIp, remotePort)
	if recvcheck != calchecksum {
		fmt.Printf("Tcp Checksum Mismatch!recvcheck:%d, calcheck:%d\n", recvcheck, calchecksum)
		return
	}

	socketAddr := transport.SocketAddr{localIp, localPort, remoteIp, remotePort}
	tcprunner, err := handler.SocketManager.GetRunnerByAddr(socketAddr)

	if err == nil {
		//fmt.Println("find established socket\n")
		tcprunner.RecvFromIpCh <- ipPacket
		return
	}

	tcprunner, err = handler.SocketManager.GetRunnerByAddr(transport.SocketAddr{localIp, localPort, model.VirtualIp{"0.0.0.0"}, 0})
	if err == nil {
		//fmt.Println("find listening socket\n")
		tcprunner.RecvFromIpCh <- ipPacket
		return
	}
	//fmt.Printf("Can not find socket runner in map %+v\n", socketAddr)
	return
}
