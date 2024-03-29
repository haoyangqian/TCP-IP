package transport

import (
	"container/heap"
	"fmt"
	"logging"
	"model"
	"time"
)

const (
	MAX_WINDOWSIZE = 65535
	MAX_WATINGTIME = 250 * 1000 * 1000 // 250ms
	MAX_PAYLOAD    = 1024
	MAX_RETRANSMIT = 3
)

type TcpSocket struct {
	Fd              int
	Addr            SocketAddr
	SeqNum          int
	StateMachine    TcpStateMachine
	SendToIpCh      chan<- model.SendMessageRequest
	ReadState       bool
	WriteState      bool
	lastSentSeq     int
	lastSentAck     int
	lastRecvAck     int
	dataSentAck     int //for data packet's ACK
	MaxAckNumRecved int
	packetsQueue    PriorityQueue
	SendWindow      SenderSlidingWindow
	RecvWindow      ArrayBasedReceiverSlidingWindow
}

func MakeSocket(fd int, fsm TcpStateMachine, ch chan<- model.SendMessageRequest) TcpSocket {
	pq := make(PriorityQueue, 0)
	return TcpSocket{
		Fd:              fd,
		Addr:            SocketAddr{model.VirtualIp{"0.0.0.0"}, 0, model.VirtualIp{"0.0.0.0"}, 0},
		StateMachine:    fsm,
		SendToIpCh:      ch,
		ReadState:       true,
		WriteState:      true,
		packetsQueue:    pq,
		MaxAckNumRecved: -1,
		RecvWindow:      MakeArrayBasedReceiverSlidingWindow(MAX_WINDOWSIZE),
	}
}

func (socket *TcpSocket) GetSendingWindow() *SenderSlidingWindow {
	return &socket.SendWindow
}

func (socket *TcpSocket) GetReceiverWindow() *ArrayBasedReceiverSlidingWindow {
	return &socket.RecvWindow
}

/*
function : send syn to remote addr:port

*/
func (socket *TcpSocket) SendCtrl(Ctrl int, seqnum int, acknum int, laddr model.VirtualIp, lport int, raddr model.VirtualIp, rport int) (int, error) {
	socket.lastSentSeq = seqnum
	socket.lastSentAck = acknum
	socket.SeqNum = seqnum
	tcpheader := MakeTcpHeader(lport, rport, seqnum, acknum, Ctrl, socket.RecvWindow.AdvertisedWindowSize())
	tcppacket := MakeTcpPacket([]byte{}, tcpheader)
	logging.Printf("[TcpSocket] send ctrl()--ctrl:%b,window size: %d seqnum: %d acknum: %d\n\n", tcpheader.Ctrl, tcpheader.Window, tcpheader.SeqNum, tcpheader.AckNum)
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
	logging.Printf("[TcpSocket] send Data()--seqnum: %d acknum: %d payloadlen:%d\n", tcpheader.SeqNum, tcpheader.AckNum, len(tcppacket.Payload))
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
	//go socket.Retransmit()
	for {
		for socket.packetsQueue.Len() != 0 {
			//retrieve the top element
			item := heap.Pop(&socket.packetsQueue).(*PacketInFlight)
			//logging.Printf("[TcpSocket] pq length : %d top ExpectedAckNum:%d MaxAckNumRecved:%d\n", socket.packetsQueue.Len(), item.ExpectedAckNum, socket.MaxAckNumRecved)
			if item.ExpectedAckNum <= socket.MaxAckNumRecved {
				//logging.Printf("[TcpSocket] Send() discarding a packet in flight, a larger ACK has already been received\n")
				socket.SendWindow.UpdateBytesInFlight(socket.SendWindow.BytesInFlight - len(item.Packet.Payload))
				//logging.Printf("[TcpSocket] Pop Sucess, length : %d", socket.packetsQueue.Len())

			} else {
				//if timeout, retransmit
				if item.ExpireTimeNanos+MAX_WATINGTIME < time.Now().UnixNano() {
					if item.RetransmitTime <= 0 {
						//up to max retransmit time, close
						fmt.Println("v_write() error: Connection reset by peer")
						socket.StateMachine.SetState(TCP_FINAL_CLOSED)
						return
					}
					tcppacket := item.Packet
					socket.SendData(tcppacket.Tcpheader.SeqNum, tcppacket.Tcpheader.AckNum, tcppacket.Payload)
					item.RetransmitTime = item.RetransmitTime - 1
					//logging.Printf("[TcpSocket] Send() retransmition, seqnum:%d acknum:%d", tcppacket.Tcpheader.SeqNum, tcppacket.Tcpheader.AckNum)
					item.ExpireTimeNanos = time.Now().UnixNano()
				}
				//logging.Printf("[TcpSocket] Send() nothing expired, looooooooooop yooooo")
				//push the item back to heap
				heap.Push(&socket.packetsQueue, item)
				break
			}
		}
		//if there are bytes should be sent
		buffer, seqnum := socket.SendWindow.Send()
		//logging.Printf("[TcpSocket] Send() send buffer:%d, seqnum:%d\n", len(buffer), seqnum)
		if seqnum > 0 && len(buffer) > 0 {
			//logging.Printf("[TcpSocket] Send() send buffer:%d, seqnum:%d\n", len(buffer), seqnum)
			//logging.Printf("[TcpSocket] Send() send buffer:%s len : %d seqnum:%d\n", string(buffer), len(buffer), seqnum)

			tcppacket, _ := socket.SendData(seqnum, socket.dataSentAck, buffer)
			socket.SendWindow.UpdateBytesInFlight(socket.SendWindow.BytesInFlight + len(buffer))
			//put it into pq
			packetinflight := PacketInFlight{
				Index:           -1,
				ExpireTimeNanos: time.Now().UnixNano(),
				Packet:          tcppacket,
				ExpectedAckNum:  tcppacket.Tcpheader.SeqNum + len(tcppacket.Payload),
				RetransmitTime:  MAX_RETRANSMIT,
			}
			heap.Push(&socket.packetsQueue, &packetinflight)
			//logging.Printf("[TcpSocket] Send() push to pq time:%d, ExpectedAckNum:%d\n", packetinflight.ExpireTimeNanos, packetinflight.ExpectedAckNum)
		}

	}
}

func (socket *TcpSocket) Recv(packet model.IpPacket) {
	// unmarshall into TCP packet
	// recv
	tcppacket := ConvertToTcpPacket(packet.Payload)
	socket.lastRecvAck = tcppacket.Tcpheader.AckNum
	payloadSize := len(tcppacket.Payload)
	//fmt.Println(tcppacket.TcpPacketString())
	if socket.StateMachine.CurrentState() != TCP_ESTAB {
		if tcppacket.Tcpheader.HasFlag(ACK) {
			logging.Printf("[TcpSocket] recv ACK -- seqnum: %d  acknum: %d ctrl:%d\n", tcppacket.Tcpheader.SeqNum, tcppacket.Tcpheader.AckNum, tcppacket.Tcpheader.Ctrl)
			if tcppacket.Tcpheader.AckNum != socket.SeqNum+1 {
				logging.Printf("[TcpSocket] Mismatch AckNum -- acknum:%d, seqnum:%d\n", tcppacket.Tcpheader.AckNum, socket.SeqNum)
				//return
			}
		}
		if tcppacket.Tcpheader.HasFlag(FIN) {
			logging.Printf("[TcpSocket] recv FIN -- seqnum: %d  acknum: %d\n", tcppacket.Tcpheader.SeqNum, tcppacket.Tcpheader.AckNum)
		}
	} else if payloadSize == 0 {
		//logging.Printf("[TcpSocket] recv ctrl()--ctrl:%b,laddr:%s,lport,%d,raddr:%s,rport:%d\n", tcppacket.Tcpheader.Ctrl, packet.Ipheader.Dst, tcppacket.Tcpheader.Destination, packet.Ipheader.Src, tcppacket.Tcpheader.Source)
		socket.SendWindow.UpdateLastAdvertisedWindow(tcppacket.Tcpheader.Window)
		if tcppacket.Tcpheader.HasFlag(ACK) {
			logging.Printf("[TcpSocket] recv ACK -- seqnum: %d  acknum: %d\n", tcppacket.Tcpheader.SeqNum, tcppacket.Tcpheader.AckNum)
			//update max ack nubmer
			if tcppacket.Tcpheader.AckNum > socket.MaxAckNumRecved {
				socket.MaxAckNumRecved = tcppacket.Tcpheader.AckNum
			}
		}
		if tcppacket.Tcpheader.HasFlag(FIN) {
			logging.Printf("[TcpSocket] recv FIN -- seqnum: %d  acknum: %d\n", tcppacket.Tcpheader.SeqNum, tcppacket.Tcpheader.AckNum)
		}
	}

	//logging.Printf("[TcpSocket] recv ctrl()--ctrl:%b,laddr:%s,lport,%d,raddr:%s,rport:%d, window: %d, payload size: %d, currentState %s\n", tcppacket.Tcpheader.Ctrl, packet.Ipheader.Dst, tcppacket.Tcpheader.Destination, packet.Ipheader.Src, tcppacket.Tcpheader.Source, tcppacket.Tcpheader.Window, len(tcppacket.Payload), socket.StateMachine.CurrentState().Name)
	event := MakeTcpTransitionEvent(tcppacket.Tcpheader)
	// state will change, execute statemachine response
	if socket.StateMachine.HasTransition(event) {
		logging.Printf("transition : %+v\n", event)
		resp, _ := socket.StateMachine.GetResponse(event)
		//fmt.Printf("resp: %+v\n", resp)
		if !resp.ShouldDoNothing() {
			if resp.ShouldDeleteSocket {
				// socket clean up
				//fmt.Printf("Socket should be deleted\n")
			} else {
				ctrl := resp.GetCtrlFlags()
				//fmt.Printf("should send: ctrl : %b\n", ctrl)
				if socket.StateMachine.CurrentState() == TCP_CLOSING {
					socket.SendCtrl(ctrl, socket.lastSentSeq+1, tcppacket.Tcpheader.SeqNum, socket.Addr.LocalIp, socket.Addr.LocalPort, socket.Addr.RemoteIp, socket.Addr.RemotePort)
				} else {
					socket.SendCtrl(ctrl, socket.lastSentSeq+1, tcppacket.Tcpheader.SeqNum+1, socket.Addr.LocalIp, socket.Addr.LocalPort, socket.Addr.RemoteIp, socket.Addr.RemotePort)
				}
			}
     	}
			// if we are still receiving, we should check the seqnum first before we transit into a new state
		if socket.StateMachine.CurrentState() == TCP_ESTAB || socket.StateMachine.CurrentState().IsActiveClose {
			if tcppacket.Tcpheader.SeqNum != socket.RecvWindow.nextSeqNumExpected {
				logging.Printf("Transition arrived before all data was received, seqNum %d, expected seqNum %d", tcppacket.Tcpheader.SeqNum, socket.RecvWindow.nextSeqNumExpected)
				socket.SendCtrl(ACK, socket.lastSentSeq, socket.RecvWindow.nextSeqNumExpected, socket.Addr.LocalIp, socket.Addr.LocalPort, socket.Addr.RemoteIp, socket.Addr.RemotePort)
				return
			}
		}

		socket.StateMachine.Transit(event)
		if socket.StateMachine.CurrentState() == TCP_ESTAB {
			socket.dataSentAck = tcppacket.Tcpheader.AckNum
			logging.Printf("[TcpSocket] trying to initialize Send window, windowsize :%d first seqnum:%d\n", MAX_WINDOWSIZE, socket.lastRecvAck)
			socket.SendWindow = MakeSenderSlidingWindow(MAX_WINDOWSIZE, socket.lastRecvAck)
			logging.Printf("[TcpSocket] %d InitializeD Send window, ring Size :%d first seqnum:%d\n", socket.Fd, socket.SendWindow.bufferSize, socket.SendWindow.returnSeqNum)
			//			if socket.RecvWindow.bufferSize == 0 {
			logging.Printf("[TcpSocket] trying to initialize Recv window, advertised window:%d\n", tcppacket.Tcpheader.Window)
			//				socket.RecvWindow = MakeReceiverSlidingWindow(MAX_WINDOWSIZE)
			socket.RecvWindow.SetNextExpectedSeqNum(socket.lastSentAck) // FIXME: add 1 or not add 1
			logging.Printf("[TcpSocket] %d Initilized Recv Window, ring Size: %d, advertisedWindowSize: %d\n", socket.Fd, socket.RecvWindow.bufferSize, socket.RecvWindow.AdvertisedWindowSize())
			//			}
			go socket.Send()
			//go socket.Retransmit()
		}

		return
	}

	if payloadSize > 0 && (socket.StateMachine.CurrentState() == TCP_ESTAB || socket.StateMachine.CurrentState().IsActiveClose) {
		//logging.Printf("[TcpSocket] %d receiving data packet", socket.Fd)
		if socket.ReadState {
			ack := socket.RecvWindow.Receive(tcppacket.Tcpheader.SeqNum, tcppacket.Payload)
			if ack > 0 {
				go socket.SendCtrl(ACK, tcppacket.Tcpheader.AckNum, ack, socket.Addr.LocalIp, socket.Addr.LocalPort, socket.Addr.RemoteIp, socket.Addr.RemotePort)
			}
		} else { // if the read has been shutdown, just return ack
			go socket.SendCtrl(ACK, tcppacket.Tcpheader.AckNum, tcppacket.Tcpheader.SeqNum+len(tcppacket.Payload), socket.Addr.LocalIp, socket.Addr.LocalPort, socket.Addr.RemoteIp, socket.Addr.RemotePort)
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
		size = socket.SendWindow.Write(buf, nbyte)
		//logging.Printf("[TcpSocket] AddToBuffer socketfd:%d write size :%d", socket.Fd, size)
		if size > 0 {
			//logging.Printf("[TcpSocket] AddToBuffer socketfd:%d write size :%d", socket.Fd, size)
		}
		if size == nbyte {
			break
		} else {
			buf = buf[size:]
			nbyte = nbyte - size
		}
	}
	return size
}

/*
*   read bytes from buffer
 */
func (socket *TcpSocket) ReadFromBuffer(bytes int) ([]byte, int) {
	buffer, size := socket.RecvWindow.Read(bytes)
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
