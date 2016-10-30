package runner

import (
	"fmt"
	"model"
	"transport"
)

type SocketRunner struct {
	socketManager transport.SocketManager
	recvFromIpCh  <-chan model.IpPacket
}

func MakeSocketRunner(sm transport.SocketManager, recvfromipch <-chan model.IpPacket) SocketRunner {
	return SocketRunner{sm, recvfromipch}
}
func (runner *SocketRunner) Run() {
	for {
		select {
		case ipPacket := <-runner.recvFromIpCh:
			localIp := model.Int2Vip(ipPacket.Ipheader.Dst)
			remoteIp := model.Int2Vip(ipPacket.Ipheader.Src)

			tcpPacket := transport.ConvertToTcpPacket(ipPacket.Payload)

			localPort := tcpPacket.Tcpheader.Destination
			remotePort := tcpPacket.Tcpheader.Source

			socketAddr := transport.SocketAddr{localIp, localPort, remoteIp, remotePort}
			tcpSocket, err := runner.socketManager.GetSocketByAddr(socketAddr)
			if err != nil {
				fmt.Printf("Could not find the corresponding socket given %+v/n", socketAddr)
			}
			tcpSocket.Recv(ipPacket)
		}
	}
}
