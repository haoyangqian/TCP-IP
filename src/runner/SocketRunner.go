package runner

import (
	"fmt"
	"model"
	"transport"
)

type SocketRunner struct {
	socketManager *transport.SocketManager
	recvFromIpCh  <-chan model.IpPacket
}

func MakeSocketRunner(sm *transport.SocketManager, recvfromipch <-chan model.IpPacket) SocketRunner {
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
			fmt.Printf("socketaddr : %+v\n", socketAddr)
			if err == nil {
				tcpSocket.Recv(ipPacket)
				fmt.Println("find established")
				continue
			}

			listenSocket, err := runner.socketManager.GetSocketByAddr(transport.SocketAddr{localIp, localPort, model.VirtualIp{"0.0.0.0"}, 0})
			if err == nil {
				//tcp accept
				fmt.Println("find listened")
				if tcpPacket.Tcpheader.HasFlag(transport.SYN) && !tcpPacket.Tcpheader.HasFlag(transport.ACK) {
					runner.socketManager.V_accept(listenSocket.Fd, remoteIp, remotePort)
				}
				continue
			}
			fmt.Printf("Could not find the corresponding socket given %+v/n", socketAddr)
		}
	}
}
