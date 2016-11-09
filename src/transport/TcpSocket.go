package transport

import (
	//"fmt"
	"container/heap"
	"logging"
	"model"
	"time"
)

const (
	MAX_WINDOWSIZE = 65535
	MAX_WATINGTIME = 1000000000 // 1s
	MAX_PAYLOAD    = 1024
)

type TcpSocket struct {
	Fd   int
	Addr SocketAddr
	//	Buffer       []byte
	SeqNum       int
	StateMachine TcpStateMachine
	SendToIpCh   chan<- model.SendMessageRequest

	lastSentSeq     int
	lastSentAck     int
	HandshakeSeqnum int
	HandshakeAcknum int
	MaxAckNumRecved int
	packetsQueue    PriorityQueue
	sendWindow      SenderSlidingWindow
	recvWindow      ReceiverSlidingWindow
}

func MakeSocket(fd int, fsm TcpStateMachine, ch chan<- model.SendMessageRequest) TcpSocket {
	//	buffer := make([]byte, MAX_SENDER_BUFFER_SIZE)
	//	recvWindow := MakeReceiverSlidingWindow(MAX_SENDER_BUFFER_SIZE)
	pq := make(PriorityQueue, 0)
	return TcpSocket{
		Fd:   fd,
		Addr: SocketAddr{model.VirtualIp{"0.0.0.0"}, 0, model.VirtualIp{"0.0.0.0"}, 0},
		//		Buffer:       buffer,
		StateMachine:    fsm,
		SendToIpCh:      ch,
		packetsQueue:    pq,
		MaxAckNumRecved: -1,
	}
}

func (socket *TcpSocket) GetSendingWindow() *SenderSlidingWindow {
	return &socket.sendWindow
}

func (socket *TcpSocket) GetReceiverWindow() *ReceiverSlidingWindow {
	return &socket.recvWindow
}

/*
function : send syn to remote addr:port

*/
func (socket *TcpSocket) SendCtrl(Ctrl int, seqnum int, acknum int, laddr model.VirtualIp, lport int, raddr model.VirtualIp, rport int) (int, error) {
	socket.lastSentSeq = seqnum
	socket.lastSentAck = acknum
	socket.SeqNum = seqnum
	socket.HandshakeSeqnum = seqnum
	socket.HandshakeAcknum = acknum
	tcpheader := MakeTcpHeader(lport, rport, seqnum, acknum, Ctrl, socket.recvWindow.AdvertisedWindowSize())
	tcppacket := MakeTcpPacket([]byte{}, tcpheader)
	logging.Logger.Printf("[TcpSocket] send ctrl()--ctrl:%b,window size: %d seqnum: %d acknum: %d\n\n", tcpheader.Ctrl, tcpheader.Window, tcpheader.SeqNum, tcpheader.AckNum)
	//set tcp checksum
	data := tcppacket.ConvertToBuffer()
	tcppacket.Tcpheader.Checksum = int(Csum(data, laddr.Vip2Int(), raddr.Vip2Int()))
	ipPayload := tcppacket.ConvertToBuffer()
	//put tcppacket to channel
	request := model.MakeSendMessageRequestWithSrc(ipPayload, model.TRANSPORT_PROTOCOL, laddr, raddr)
	socket.SendToIpCh <- request
	return 1, nil
}

/*
*     send data packet
*     para: seqnum, acknum, payload
*     return the tcppacket (in favor of retransmission)
 */
func (socket *TcpSocket) SendData(seqnum int, acknum int, payload []byte) (*TcpPacket, error) {
	socket.lastSentSeq = seqnum
	socket.lastSentAck = acknum
	socket.SeqNum = seqnum

	tcpheader := MakeTcpHeader(socket.Addr.LocalPort, socket.Addr.RemotePort, seqnum, acknum, 0, 65535)
	tcppacket := MakeTcpPacket(payload, tcpheader)
	logging.Logger.Printf("[TcpSocket] send Data()--seqnum: %d acknum: %d payloadlen:%d\n", tcpheader.SeqNum, tcpheader.AckNum, len(tcppacket.Payload))
	//set tcp checksum
	data := tcppacket.ConvertToBuffer()
	tcppacket.Tcpheader.Checksum = int(Csum(data, socket.Addr.LocalIp.Vip2Int(), socket.Addr.RemoteIp.Vip2Int()))
	ipPayload := tcppacket.ConvertToBuffer()
	//put tcppacket to channel
	request := model.MakeSendMessageRequestWithSrc(ipPayload, model.TRANSPORT_PROTOCOL, socket.Addr.LocalIp, socket.Addr.RemoteIp)
	socket.SendToIpCh <- request
	return &tcppacket, nil
}

func (socket *TcpSocket) Send() {
	// put together a TCP packet
	// marshal TCP packet into bytes (message)
	// construct SendMessageRequest
	// ch <- request

	for {
		//retransmission
		if socket.packetsQueue.Len() != 0 {
			n := socket.packetsQueue.Len()
			//check if the expected ACK < max ack, just drop it
			//logging.Logger.Printf("[TcpSocket] length : %d Heap top ExpectedAckNum:%d MaxAckNumRecved:%d ExpireTimeNanos:%d\n", n, socket.packetsQueue[n-1].ExpectedAckNum, socket.MaxAckNumRecved, socket.packetsQueue[n-1].ExpireTimeNanos)
			if socket.packetsQueue[n-1].ExpectedAckNum <= socket.MaxAckNumRecved {
				heap.Pop(&socket.packetsQueue)
				//logging.Logger.Printf("[TcpSocket] Pop Sucess, length : %d", socket.packetsQueue.Len())
			} else {
				//if timeout, retransmit
				if socket.packetsQueue[n-1].ExpireTimeNanos+MAX_WATINGTIME < time.Now().UnixNano() {
					tcppacket := socket.packetsQueue[n-1].Packet
					socket.SendData(tcppacket.Tcpheader.SeqNum, tcppacket.Tcpheader.AckNum, tcppacket.Payload)
					//logging.Logger.Printf("[TcpSocket] Send() retransmition")
					//update the packet in the pq
					socket.packetsQueue.update(socket.packetsQueue[n-1], time.Now().UnixNano())
				}
			}
		}
		//if there are bytes should be sent
		buffer, seqnum := socket.sendWindow.Send()
		if seqnum > 0 && len(buffer) > 0 {
			//logging.Logger.Printf("[TcpSocket] Send() send buffer:%s len : %d seqnum:%d\n", string(buffer), len(buffer), seqnum)
			tcppacket, _ := socket.SendData(seqnum, socket.HandshakeAcknum, buffer)
			//put it into pq
			packetinflight := PacketInFlight{
				Index:           -1,
				ExpireTimeNanos: time.Now().UnixNano(),
				Packet:          tcppacket,
				ExpectedAckNum:  tcppacket.Tcpheader.SeqNum + len(tcppacket.Payload),
			}
			heap.Push(&socket.packetsQueue, &packetinflight)
		}

	}
}

func (socket *TcpSocket) Recv(packet model.IpPacket) {
	// unmarshall into TCP packet
	// recv
	tcppacket := ConvertToTcpPacket(packet.Payload)

	if socket.recvWindow.bufferSize == 0 {
		socket.recvWindow = MakeReceiverSlidingWindow(MAX_WINDOWSIZE)
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
	} else if len(tcppacket.Payload) == 0 {
		logging.Logger.Printf("[TcpSocket] recv ctrl()--ctrl:%b,laddr:%s,lport,%d,raddr:%s,rport:%d\n", tcppacket.Tcpheader.Ctrl, packet.Ipheader.Dst, tcppacket.Tcpheader.Destination, packet.Ipheader.Src, tcppacket.Tcpheader.Source)
		if tcppacket.Tcpheader.HasFlag(ACK) {
			logging.Logger.Printf("[TcpSocket] recv ACK -- seqnum: %d  acknum: %d\n", tcppacket.Tcpheader.SeqNum, tcppacket.Tcpheader.AckNum)
			//update max ack nubmer
			if tcppacket.Tcpheader.AckNum > socket.MaxAckNumRecved {
				socket.MaxAckNumRecved = tcppacket.Tcpheader.AckNum
			}
			//update lastAckedBytes
			length := tcppacket.Tcpheader.AckNum - socket.sendWindow.lastSeqnumAcked
			//logging.Logger.Printf("[TcpSocket] update length -- %d\n", length)
			for i := 0; i < length; i++ {
				if socket.sendWindow.lastByteAcked.Next().Value != nil {
					socket.sendWindow.lastSeqnumAcked = socket.sendWindow.lastByteAcked.Next().Value.(TcpByte).SeqNum
					socket.sendWindow.lastByteAcked.Next().Value = nil
					socket.sendWindow.lastByteAcked = socket.sendWindow.lastByteAcked.Next()
				}
			}
			//logging.Logger.Printf("[TcpSocket] update MaxAckNumRecved -- %d\n", socket.MaxAckNumRecved)
		}
	}

	// if we are still receiving, we should check the seqnum first before we transit into a new state
	if socket.StateMachine.CurrentState() == TCP_ESTAB || socket.StateMachine.CurrentState().IsActiveClose {
		if tcppacket.Tcpheader.SeqNum != socket.recvWindow.nextSeqNumExpected {
			logging.Logger.Printf("Transition arrived before all data was received, seqNum %d, expected seqNum %d", tcppacket.Tcpheader.SeqNum, socket.recvWindow.nextSeqNumExpected)
			socket.SendCtrl(ACK, socket.lastSentSeq, socket.recvWindow.nextSeqNumExpected, socket.Addr.LocalIp, socket.Addr.LocalPort, socket.Addr.RemoteIp, socket.Addr.RemotePort)
			return
		}
	}

	logging.Logger.Printf("[TcpSocket] recv ctrl()--ctrl:%b,laddr:%s,lport,%d,raddr:%s,rport:%d, window: %d, payload size: %d, currentState %s\n", tcppacket.Tcpheader.Ctrl, packet.Ipheader.Dst, tcppacket.Tcpheader.Destination, packet.Ipheader.Src, tcppacket.Tcpheader.Source, tcppacket.Tcpheader.Window, len(tcppacket.Payload), socket.StateMachine.CurrentState().Name)
	event := MakeTcpTransitionEvent(tcppacket.Tcpheader)
	// state will change, execute statemachine response
	if socket.StateMachine.HasTransition(event) {
		logging.Logger.Printf("transition : %+v\n", event)
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
		if socket.StateMachine.CurrentState() == TCP_ESTAB {
			socket.HandshakeAcknum = tcppacket.Tcpheader.SeqNum
			socket.HandshakeSeqnum = tcppacket.Tcpheader.AckNum
			logging.Logger.Printf("[TcpSocket] trying to initialize Send window, windowsize :%d first seqnum:%d\n", MAX_WINDOWSIZE, socket.HandshakeSeqnum)
			socket.sendWindow = MakeSenderSlidingWindow(MAX_WINDOWSIZE, socket.HandshakeSeqnum)
			logging.Logger.Printf("[TcpSocket] %d InitializeD Send window, ring Size :%d first seqnum:%d\n", socket.Fd, socket.sendWindow.bufferSize, socket.sendWindow.Seqnum)
			if socket.recvWindow.bufferSize == 0 {
				logging.Logger.Printf("[TcpSocket] trying to initialize Recv window, advertised window:%d\n", tcppacket.Tcpheader.Window)
				socket.recvWindow = MakeReceiverSlidingWindow(tcppacket.Tcpheader.Window)
				logging.Logger.Printf("[TcpSocket] %d Initilized Recv Window, ring Size: %d, advertisedWindowSize: %d\n", socket.Fd, socket.recvWindow.bufferSize, socket.recvWindow.AdvertisedWindowSize())
			}
			go socket.Send()
		}

		socket.StateMachine.Transit(event)
		if socket.StateMachine.CurrentState() == TCP_ESTAB {
			socket.recvWindow.SetNextExpectedSeqNum(tcppacket.Tcpheader.SeqNum)
		}
		return
	}

	if len(tcppacket.Payload) > 0 && (socket.StateMachine.CurrentState() == TCP_ESTAB || socket.StateMachine.CurrentState().IsActiveClose) {
		logging.Logger.Printf("[TcpSocket] %d receiving data packet", socket.Fd)
		ack := socket.recvWindow.Receive(tcppacket.Tcpheader.SeqNum, tcppacket.Payload)
		if ack > 0 {
			socket.SendCtrl(ACK, tcppacket.Tcpheader.AckNum, ack, socket.Addr.LocalIp, socket.Addr.LocalPort, socket.Addr.RemoteIp, socket.Addr.RemotePort)
		}
	}
}

/*
*   add bytes to buffer, should block
 */

func (socket *TcpSocket) AddToBuffer(buf []byte, nbyte int) int {
	//block till writing up to nbyte
	var size int
	for {
		size = socket.sendWindow.Write(buf, nbyte)
		logging.Logger.Printf("[TcpSocket] AddToBuffer socketfd:%d write size :%d", socket.Fd, size)
		if size == nbyte {
			break
		} else {
			buf = buf[size+1:]
			nbyte = nbyte - size
		}
	}
	return size
}

/*
*   read bytes from buffer
 */
func (socket *TcpSocket) ReadFromBuffer(bytes int) ([]byte, int) {
	buffer, size := socket.recvWindow.Read(bytes)
	return buffer, size
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
