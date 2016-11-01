package transport

import (
	"fmt"
	"model"
)

type SocketRunner struct {
	Socket        *TcpSocket
	socketManager *SocketManager
	recvFromIpCh  <-chan model.IpPacket
}

func MakeSocketRunner(socket *TcpSocket, sm *SocketManager, recvFromIpCh <-chan model.IpPacket) SocketRunner {
	return SocketRunner{socket, sm, recvFromIpCh}
}

func (runner *SocketRunner) Run() {
	for {
		select {
		case ipPacket := <-runner.recvFromIpCh:
			// do this in IP
			//			localIp := model.Int2Vip(ipPacket.Ipheader.Dst)
			//			remoteIp := model.Int2Vip(ipPacket.Ipheader.Src)
			//
			//			tcpPacket := ConvertToTcpPacket(ipPacket.Payload)
			//
			//			localPort := tcpPacket.Tcpheader.Destination
			//			remotePort := tcpPacket.Tcpheader.Source
			//
			//			socketAddr := SocketAddr{localIp, localPort, remoteIp, remotePort}
			//			tcpSocket, err := runner.socketManager.GetSocketByAddr(socketAddr)
			runner.Socket.Recv(ipPacket)
			//			if err == nil {
			//				tcpSocket.Recv(ipPacket)
			//				fmt.Println("find established")
			//				continue
			//			}
			//
			//			listenSocket, err := runner.socketManager.GetSocketByAddr(SocketAddr{localIp, localPort, model.VirtualIp{"0.0.0.0"}, 0})
			//			if err == nil {
			//				//tcp accept
			//				fmt.Println("find listened")
			//				if tcpPacket.Tcpheader.HasFlag(SYN) && !tcpPacket.Tcpheader.HasFlag(ACK) {
			//					runner.socketManager.V_accept(listenSocket.Fd, remoteIp, remotePort)
			//				}
			//				continue
			//			}
			//			fmt.Printf("Could not find the corresponding socket given %+v/n", socketAddr)
			if true == false {
				fmt.Println("hi")
			}

		}
	}
}
