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
	//	fmt.Printf("I am running! socketfd:%d\n", runner.Socket.Fd)
	for {
		//fmt.Printf("waiting for channel! channel addr:%x\n", runner.RecvFromIpCh)
		select {
		case ipPacket := <-runner.RecvFromIpCh:
			//fmt.Println("***** Runner #%d received data from channel", runner.Socket.Fd)
			runner.Socket.Recv(ipPacket)
			if true == false {
				fmt.Println("hi")
			}
		case <-runner.Socket.StateMachine.TimerChannel():
			if !runner.Socket.StateMachine.CurrentState().CanTimeout() {
				continue
			}

			fmt.Println("socket state timed out , retry!", runner.Socket.StateMachine.RetryCount(), runner.Socket.Fd, runner.Socket.StateMachine.CurrentState())
			if runner.Socket.StateMachine.RetryCount() >= TCP_MAX_RETRY_COUNT {
				// terminate this thread, this socket is literally dead
				return
			}

			runner.Socket.RepeatPreviousStateAction()
			runner.Socket.StateMachine.ResetStateTimer()
		}
	}
}
