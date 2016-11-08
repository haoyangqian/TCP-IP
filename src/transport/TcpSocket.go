package transport

import (
	//"fmt"
	"logging"
	"model"
)

const (
	MAX_SENDER_BUFFER_SIZE = 512
)

type TcpSocket struct {
	Fd   int
	Addr SocketAddr
	//	Buffer       []byte
	SeqNum       int
	StateMachine TcpStateMachine
	SendToIpCh   chan<- model.SendMessageRequest

	lastSentSeq int
	lastSentAck int

	//	sendWindow SenderSlidingWindow
	recvWindow ReceiverSlidingWindow
}

func MakeSocket(fd int, fsm TcpStateMachine, ch chan<- model.SendMessageRequest) TcpSocket {
	//	buffer := make([]byte, MAX_SENDER_BUFFER_SIZE)
	//	recvWindow := MakeReceiverSlidingWindow(MAX_SENDER_BUFFER_SIZE)
	return TcpSocket{
		Fd:   fd,
		Addr: SocketAddr{model.VirtualIp{"0.0.0.0"}, 0, model.VirtualIp{"0.0.0.0"}, 0},
		//		Buffer:       buffer,
		StateMachine: fsm,
		SendToIpCh:   ch,
	}
}

func (socket *TcpSocket) AddToBuffer(buf []byte, nbyte int) {

}

/*
function : send syn to remote addr:port

*/
func (socket *TcpSocket) SendCtrl(Ctrl int, seqnum int, acknum int, laddr model.VirtualIp, lport int, raddr model.VirtualIp, rport int) (int, error) {
	socket.lastSentSeq = seqnum
	socket.lastSentAck = acknum
	socket.SeqNum = seqnum

	tcpheader := MakeTcpHeader(lport, rport, seqnum, acknum, Ctrl, socket.recvWindow.AdvertisedWindowSize())
	tcppacket := MakeTcpPacket([]byte{}, tcpheader)
	logging.Logger.Printf("[TcpSocket] send ctrl()--ctrl:%b,window size: %d seqnum: %d acknum: %d\n", tcpheader.Ctrl, tcpheader.Window, tcpheader.SeqNum, tcpheader.AckNum)
	//set tcp checksum
	data := tcppacket.ConvertToBuffer()
	tcppacket.Tcpheader.Checksum = int(Csum(data, laddr.Vip2Int(), raddr.Vip2Int()))
	ipPayload := tcppacket.ConvertToBuffer()
	//put tcppacket to channel
	request := model.MakeSendMessageRequestWithSrc(ipPayload, model.TRANSPORT_PROTOCOL, laddr, raddr)
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

	if socket.recvWindow.bufferSize == 0 {
		socket.recvWindow = MakeReceiverSlidingWindow(tcppacket.Tcpheader.Window)
		logging.Logger.Printf("[TcpSocket] %d Initilized Recv Window, ring Size: %d, advertisedWindowSize: %d\n", socket.Fd, socket.recvWindow.bufferSize, socket.recvWindow.AdvertisedWindowSize())
	}
	//fmt.Println(tcppacket.TcpPacketString())

	if socket.StateMachine.CurrentState() != TCP_ESTAB {
		if tcppacket.Tcpheader.HasFlag(ACK) {
			if tcppacket.Tcpheader.AckNum != socket.SeqNum+1 {
				logging.Logger.Printf("[TcpSocket] Mismatch AckNum -- acknum:%d, seqnum:%d\n", tcppacket.Tcpheader.AckNum, socket.SeqNum)
				return
			}
		}
	}
	logging.Logger.Printf("[TcpSocket] recv ctrl()--ctrl:%b,laddr:%s,lport,%d,raddr:%s,rport:%d\n", tcppacket.Tcpheader.Ctrl, packet.Ipheader.Dst, tcppacket.Tcpheader.Destination, packet.Ipheader.Src, tcppacket.Tcpheader.Source)
	event := MakeTcpTransitionEvent(tcppacket.Tcpheader)
	// state will change, execute statemachine response
	if socket.StateMachine.HasTransition(event) {
		//fmt.Printf("transition : %+v\n", event)
		resp, _ := socket.StateMachine.GetResponse(event)
		//fmt.Printf("resp: %+v\n", resp)
		if !resp.ShouldDoNothing() {
			if resp.ShouldDeleteSocket {
				// socket clean up
				//fmt.Printf("Socket should be deleted\n")
			} else {
				ctrl := resp.GetCtrlFlags()
				//fmt.Printf("should send: ctrl : %b\n", ctrl)
				socket.SendCtrl(ctrl, socket.lastSentSeq, tcppacket.Tcpheader.SeqNum+1, socket.Addr.LocalIp, socket.Addr.LocalPort, socket.Addr.RemoteIp, socket.Addr.RemotePort)
			}

		}

		socket.StateMachine.Transit(event)
		return
	}

	if len(tcppacket.Payload) > 0 && (socket.StateMachine.CurrentState() == TCP_ESTAB || socket.StateMachine.CurrentState().IsActiveClose) {
		logging.Logger.Printf("[TcpSocket] %d receiving data packet", socket.Fd)
		ack := socket.recvWindow.Receive(tcppacket.Tcpheader.SeqNum, tcppacket.Payload)

		socket.SendCtrl(ACK, 0, ack, socket.Addr.LocalIp, socket.Addr.LocalPort, socket.Addr.RemoteIp, socket.Addr.RemotePort)
	}
}

func (socket *TcpSocket) ReadFromBuffer(bytes int, block bool) []byte {
	// blocking
	if block {
		bytesRead := 0
		buffer := make([]byte, bytes)

		for {
			if bytesRead == bytes {
				break
			}

			readBuffer, read := socket.recvWindow.Read(bytes - bytesRead)
			buffer = append(buffer, readBuffer...)
			bytesRead += read
		}
		return buffer

		// non-blocking
	} else {
		buffer, _ := socket.recvWindow.Read(bytes)
		return buffer
	}
}

func (socket *TcpSocket) RepeatPreviousStateAction() {
	previousResponse := socket.StateMachine.GetPreviousResponse()

	if previousResponse.ShouldDoNothing() {
		return
	}

	if previousResponse.ShouldDeleteSocket {
		// socket clean up
	}

	ctrlFlags := previousResponse.GetCtrlFlags()
	socket.SendCtrl(ctrlFlags, socket.lastSentSeq, socket.lastSentAck, socket.Addr.LocalIp, socket.Addr.LocalPort, socket.Addr.RemoteIp, socket.Addr.RemotePort)

	socket.StateMachine.IncrementRetryCount()
}
