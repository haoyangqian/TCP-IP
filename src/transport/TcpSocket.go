package transport

import (
	"fmt"
	"math/rand"
	"model"
)

type TcpSocket struct {
	//sm
	Fd     int
	Addr   SocketAddr
	Buffer []byte

	//	  sm StateMachine
	//    sm.CurrentState()
	//    sm.GetNextCtrl(RecvCtrl) -> CtrlToSend
	//    sm.Transit()
	//    sm.Transit(RecvCtrl, SentCtrl)

	SendCh chan<- SendTcpMessageRequest
	RecvCh chan model.IpPacket //receive Ip packet from Ip layer
}

func MakeSocket(fd int, SendCh chan<- SendTcpMessageRequest, RecvCh chan model.IpPacket) TcpSocket {
	buffer := make([]byte, 0)
	return TcpSocket{fd, SocketAddr{model.VirtualIp{"0.0.0.0"}, 0, model.VirtualIp{"0.0.0.0"}, 0}, buffer, SendCh, RecvCh}
}

func (socket *TcpSocket) SetAddr(addr SocketAddr) {
	socket.Addr = addr
}

/*
function : send syn to remote addr:port

*/
func (socket *TcpSocket) SendSyn(laddr, raddr model.VirtualIp, lport, rport int) (int, error) {

	tcpheader := MakeTcpHeader(lport, rport, int(rand.Uint32()), 0, 2, 0xaaaa)
	tcppacket := MakeTcpPacket([]byte{}, tcpheader)
	data := tcppacket.ConvertToBuffer()
	tcppacket.Tcpheader.Checksum = int(Csum(data, laddr.Vip2Int(), raddr.Vip2Int()))
	data = tcppacket.ConvertToBuffer()

	request := SendTcpMessageRequest{socket.Fd, data}
	socket.SendCh <- request

	return 1, nil
}

func (socket *TcpSocket) Send(request SendTcpMessageRequest, ch chan<- model.SendMessageRequest) {
	// put together a TCP packet
	// marshal TCP packet into bytes (message)
	// construct SendMessageRequest
	// ch <- request

	messagerequest := model.MakeSendMessageRequest(request.Payload, 0, socket.Addr.RemoteIp)
	ch <- messagerequest
}

func (socket *TcpSocket) Recv(packet model.IpPacket) {
	// unmarshall into TCP packet
	// recv
	tcppacket := ConvertToTcpPacket(packet.Payload)
	fmt.Println(tcppacket.TcpPacketString())
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
