package transport

import (
	"model"
)

type TcpSocket struct {
	//sm
	fd     int
	addr   SocketAddr
	Buffer []byte

	//	  sm StateMachine
	//    sm.CurrentState()
	//    sm.GetNextCtrl(RecvCtrl) -> CtrlToSend
	//    sm.Transit()
	//    sm.Transit(RecvCtrl, SentCtrl)

	SendCh chan<- transport.SendTcpMessageRequest
	RecvCh <-chan model.IpPacket //receive Ip packet from Ip layer
}

func (socket *TcpSocket) Send(request transport.SendTcpMessageRequest, ch chan<- model.SendMessageRequest) {
	// put together a TCP packet
	// marshal TCP packet into bytes (message)
	// construct SendMessageRequest
	// ch <- request
}

func (socket *TcpSocket) Recv(packet model.IpPacket) {
	// unmarshall into TCP packet
	// recv

}

func (socket *TcpSocket) ReadFromBuffer(bytes int, block bool) []byte {
	if block {
		for len(socket.Buffer) < int {
			// if connection closes, break
		}

		return socket.Buffer
	} else {
		return socket.Buffer
	}
}
