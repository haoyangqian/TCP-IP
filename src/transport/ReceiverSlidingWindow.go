package transport

import (
	"container/ring"
	//	"fmt"
	"logging"
	"strconv"
	"sync"
)

type ReceiverSlidingWindow struct {
	lastByteRead     *ring.Ring
	nextByteExpected *ring.Ring
	lastByteReceived *ring.Ring

	buffer     *ring.Ring
	bufferSize int
	Lock       *sync.RWMutex

	nextSeqNumExpected int
	totalBytesRecevied int
	bytesInBuffer      int
}

func MakeReceiverSlidingWindow(bufferSize int) ReceiverSlidingWindow {
	ring := ring.New(bufferSize)
	lastByteRead := ring
	nextByteExpected := ring.Next()
	lastByteReceived := ring

	return ReceiverSlidingWindow{
		lastByteRead:     lastByteRead,
		nextByteExpected: nextByteExpected,
		lastByteReceived: lastByteReceived,
		buffer:           ring,
		bufferSize:       bufferSize,
		Lock:             &sync.RWMutex{},
	}
}

func (w *ReceiverSlidingWindow) SetNextExpectedSeqNum(seqNum int) {
	w.nextSeqNumExpected = seqNum
}

func (w *ReceiverSlidingWindow) AdvertisedWindowSize() int {
	//	if w.bufferSize == 0 {
	//		return 65535
	//	} else {
	//		//		return 0
	//		return w.bufferSize - Distance(w.lastByteRead, w.nextByteExpected) + 1
	//	}

	return w.bufferSize - w.bytesInBuffer
}

func (w *ReceiverSlidingWindow) Receive(seqNum int, payload []byte) int {
	// check whether packet is in the window
	//logging.Printf("[DEBUG][RecvWindow] payload length: %d\n", len(payload))
	w.Lock.Lock()
	defer w.Lock.Unlock()

	if seqNum != w.nextSeqNumExpected {
		logging.Printf("[WARN] Received out-of-order packet, received %d, expected %d\n", seqNum, w.nextSeqNumExpected)
	}

	if seqNum < w.nextSeqNumExpected || seqNum > (w.nextSeqNumExpected+w.AdvertisedWindowSize()) || len(payload) > w.AdvertisedWindowSize() {
		// drop packet, bye byte
		logging.Printf("[RecvWindow] Dropping out-of-window/too-big packet, seqNum %d, expected seqNum is %d\n", seqNum, w.nextSeqNumExpected)
		return w.nextSeqNumExpected
	}

	// find the appropriate spot of the put the received block of data
	moves := seqNum - w.nextSeqNumExpected
	if moves < 0 {
		logging.Printf("[FATAL] moves is calculated as %d, r %d e %d\n", moves, seqNum, w.nextSeqNumExpected)
	}

	//count++
	ringPointer := w.nextByteExpected
	ringPointer = ringPointer.Move(moves)

	var expectedAfter string
	if w.nextByteExpected.Prev().Value == nil {
		expectedAfter = "nil pointer"
	} else {
		expectedAfter = strconv.Itoa(w.nextByteExpected.Prev().Value.(TcpByte).SeqNum + 1)
	}
	logging.Printf("[DEBUG] Received SeqNum %d, nextExpected at %s, moves calculated as %d, distance form lastRead to nextExpected is %d",
		seqNum, expectedAfter, moves, Distance(w.lastByteRead, w.nextByteExpected))

	// write payload into ring
	logging.Printf("[DEBUG][RecvWindow] Writing received payload into buffer, seqNum %d payload length: %d\n", seqNum, len(payload))
	for i := 0; i < len(payload); i++ {

		ringPointer.Value = TcpByte{seqNum + i, payload[i]}
		if i == 0 {
			logging.Printf("First byte recieved is written at %d, nextExpectedSeqNum was %d\n", ringPointer.Value.(TcpByte).SeqNum, w.nextSeqNumExpected)
		}
		//logging.Printf("[DEBUG][RecvWindow] Recv window position %d, data %s\n", ringPointer.Value.(TcpByte).SeqNum, string(ringPointer.Value.(TcpByte).B))
		ringPointer = ringPointer.Next()
	}
	w.totalBytesRecevied += len(payload)

	// udpate last byte received pointers
	w.lastByteReceived = ringPointer.Prev()

	if seqNum != w.nextSeqNumExpected {
		// do not move!
	} else {
		for {
			if w.nextByteExpected.Value == nil {
				break
			} else if w.nextByteExpected == w.lastByteRead {
				w.nextByteExpected = w.nextByteExpected.Next()
				w.nextSeqNumExpected += 1
				w.bytesInBuffer += 1
				break
			} else {
				if w.nextByteExpected.Value != nil && w.nextByteExpected.Next().Value != nil {
					if w.nextByteExpected.Value.(TcpByte).SeqNum+1 != w.nextByteExpected.Next().Value.(TcpByte).SeqNum {
						logging.Printf("[FATAL] pointer seqnum is %d, next seqnum is %d\n", w.nextByteExpected.Value.(TcpByte).SeqNum, w.nextByteExpected.Next().Value.(TcpByte).SeqNum)
					}
				}
				w.nextByteExpected = w.nextByteExpected.Next()
				w.nextSeqNumExpected += 1
				w.bytesInBuffer += 1
			}
		}
	}

	if w.nextByteExpected.Prev().Value == nil {
		expectedAfter = "nil pointer"
	} else {
		expectedAfter = strconv.Itoa(w.nextByteExpected.Prev().Value.(TcpByte).SeqNum + 1)
	}
	logging.Printf("[DEBUG][RecvWindow] Recv window nextByteExpected pointer is at %s, nextExpectedSeqNum is %d\n", expectedAfter, w.nextSeqNumExpected)

	// return next expected ACK
	return w.nextSeqNumExpected
}

func (w *ReceiverSlidingWindow) Read(bytes int) ([]byte, int) {
	if w.totalBytesRecevied == 0 {
		return []byte{}, 0
	}

	w.Lock.RLock()
	defer w.Lock.RUnlock()
	readableBytes := w.bytesInBuffer

	if readableBytes > 0 {
		logging.Printf("[DEBUG][Read] %d readable bytes in buffer\n", readableBytes)
	}
	var bytesToRead int
	if readableBytes > bytes {
		bytesToRead = bytes
	} else {
		bytesToRead = readableBytes
	}

	buffer := make([]byte, bytesToRead)
	for i := 0; i < bytesToRead; i++ {
		if w.lastByteRead.Next().Next().Value != nil {
			if w.lastByteRead.Next().Value.(TcpByte).SeqNum != w.lastByteRead.Next().Next().Value.(TcpByte).SeqNum-1 {
				logging.Printf("[DEBUG][RecvWindow] sequence not contiguous!first seq:%d next seq:%d\n", w.lastByteRead.Next().Value.(TcpByte).SeqNum, w.lastByteRead.Next().Next().Value.(TcpByte).SeqNum)
			}
		}
		buffer[i] = w.lastByteRead.Next().Value.(TcpByte).B
		w.lastByteRead.Next().Value = nil
		w.bytesInBuffer -= 1
		w.lastByteRead = w.lastByteRead.Next()
		//logging.Printf("****Window Read****")
	}

	return buffer, bytesToRead
}
