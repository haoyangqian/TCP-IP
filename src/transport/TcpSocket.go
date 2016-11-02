package transport

import (
	"fmt"
	"model"
)

type TcpSocket struct {
	Fd           int
	Addr         SocketAddr
	Buffer       []byte
	SeqNum       int
	StateMachine TcpStateMachine
	SendToIpCh   chan<- model.SendMessageRequest

	//	  sm StateMachine
	//    sm.CurrentState()
	//    sm.GetNextCtrl(RecvCtrl) -> CtrlToSend
	//    sm.Transit()
	//    sm.Transit(RecvCtrl, SentCtrl)

	//	SendCh chan<- SendTcpMessageRequest
	//	RecvCh chan model.IpPacket //receive Ip packet from Ip layer
}

func MakeSocket(fd int, fsm TcpStateMachine, ch chan<- model.SendMessageRequest) TcpSocket {
	buffer := make([]byte, 0)
	return TcpSocket{fd, SocketAddr{model.VirtualIp{"0.0.0.0"}, 0, model.VirtualIp{"0.0.0.0"}, 0}, buffer, 0, fsm, ch}
}

/*
function : send syn to remote addr:port

*/
func (socket *TcpSocket) SendCtrl(Ctrl int, seqnum int, acknum int, laddr model.VirtualIp, lport int, raddr model.VirtualIp, rport int) (int, error) {
	//fmt.Printf("send ctrl() -- ctrl:%b,laddr:%s,lport,%d,raddr:%s,rport:%d\n", Ctrl, laddr.Ip, lport, raddr.Ip, rport)
	tcpheader := MakeTcpHeader(lport, rport, seqnum, acknum, Ctrl, 0xaaaa)
	socket.SeqNum = seqnum
	tcppacket := MakeTcpPacket([]byte{}, tcpheader)
	data := tcppacket.ConvertToBuffer()
	tcppacket.Tcpheader.Checksum = int(Csum(data, laddr.Vip2Int(), raddr.Vip2Int()))
	data = tcppacket.ConvertToBuffer()

	request := model.MakeSendMessageRequestWithSrc(data, model.TRANSPORT_PROTOCOL, laddr, raddr)
	socket.SendToIpCh <- request
	return 1, nil
}

func (socket *TcpSocket) Send(request SendTcpMessageRequest, ch chan<- model.SendMessageRequest) {
	// put together a TCP packet
	// marshal TCP packet into bytes (message)
	// construct SendMessageRequest
	// ch <- request

	messagerequest := model.MakeSendMessageRequestWithSrc(request.Payload, 0, socket.Addr.LocalIp, socket.Addr.RemoteIp)
	ch <- messagerequest
}

func (socket *TcpSocket) Recv(packet model.IpPacket) {
	// unmarshall into TCP packet
	// recv
	tcppacket := ConvertToTcpPacket(packet.Payload)
	fmt.Println(tcppacket.TcpPacketString())

	if socket.StateMachine.CurrentState() != TCP_ESTAB {
		if tcppacket.Tcpheader.HasFlag(ACK) {
			if tcppacket.Tcpheader.AckNum != socket.SeqNum+1 {
				fmt.Printf("Mismatch AckNum, acknum:%d, seqnum:%d\n", tcppacket.Tcpheader.AckNum, socket.SeqNum)
				return
			}
		}
	}
	event := MakeTcpTransitionEvent(tcppacket.Tcpheader)
	//fmt.Printf("socket.Recv(): event: %+v\n", event)
	//fmt.Printf("socket.Recv(): current state:%s", socket.StateMachine.CurrentState().Name)
	// state will change, execute statemachine response
	if socket.StateMachine.HasTransition(event) {
		//fmt.Printf("transition : %+v\n", event)
		resp, _ := socket.StateMachine.GetResponse(event)
		fmt.Printf("resp: %+v\n", resp)
		if !resp.ShouldDoNothing() {
			if resp.ShouldDeleteSocket {
				// socket clean up
				fmt.Printf("Socket should be deleted\n")
			} else {
				ctrl := resp.GetCtrlFlags()
				fmt.Printf("should send: ctrl : %b\n", ctrl)
				socket.SendCtrl(ctrl, 0, tcppacket.Tcpheader.SeqNum+1, socket.Addr.LocalIp, socket.Addr.LocalPort, socket.Addr.RemoteIp, socket.Addr.RemotePort)
			}

		}

		socket.StateMachine.Transit(event)
	}

	// state does not change
	// if socket.StateMachine.CurrentState() == TCP_ESTABLISHED || socket.StateMachine.CurrentState().IsActiveClose() {
	// sliding window blah blah blah
	// }
}

func (socket *TcpSocket) ReadFromBuffer(bytes int, block bool) []byte {
	if block {
		for len(socket.Buffer) < bytes {
			// if connection closes, break
		}

		return socket.Buffer
	} else {
		return socket.Buffer
	}
}
