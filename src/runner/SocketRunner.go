package runner

import (
	"model"
	"transport"
)

type SocketRunner struct {
	SM           transport.SocketManager
	RecvFromIpCh <-chan model.IpPacket
}

func MakeSocketRunner(sm transport.SocketManager,
	recvfromipch <-chan model.IpPacket) SocketRunner {
	return SocketRunner{sm, recvfromipch}
}
func (runner *SocketRunner) run() {
	for {
		select {
		case packet := <-runner.RecvFromIpCh:
			runner.socket.Recv(packet)
		case request := <-runner.RecvFromDriverCh:
			runner.socket.Send(request, runner.SendToIpCh)
		}
	}
}
