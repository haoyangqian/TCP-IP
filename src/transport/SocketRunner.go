package transport

import (
	"fmt"
	"model"
)

type SocketRunner struct {
	Socket        *TcpSocket
	socketManager *SocketManager
	RecvFromIpCh  chan model.IpPacket
}

func MakeSocketRunner(socket *TcpSocket, sm *SocketManager, recvFromIpCh chan model.IpPacket) SocketRunner {
	return SocketRunner{socket, sm, recvFromIpCh}
}

func (runner *SocketRunner) Run() {
	fmt.Printf("I am running! socketfd:%d\n", runner.Socket.Fd)
	for {
		fmt.Printf("waiting for channel! channel addr:%x\n", runner.RecvFromIpCh)
		select {
		case ipPacket := <-runner.RecvFromIpCh:
			//fmt.Println("***** Runner #%d received data from channel", runner.Socket.Fd)
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
