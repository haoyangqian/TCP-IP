package runner

import (
	"model"
	"transport"
)

type SocketRunner struct {
	socket           transport.TcpSocket
	SendToIpCh       chan<- model.SendMessageRequest //send request to IP layer
	RecvFromIpCh     <-chan model.IpPacket
	RecvFromDriverCh <-chan transport.SendTcpMessageRequest
}

func (runner *SocketRunner) run() {
	//	for {
	//		select {
	//		case packet := <-runner.RecvFromIpCh:
	//			//runner.socket.recv(packet)
	//		case request := <-runner.RecvFromDriverCh:
	//			//runner.socket.send(request, runner.SendToIpCh)
	//		}
	//	}
}
