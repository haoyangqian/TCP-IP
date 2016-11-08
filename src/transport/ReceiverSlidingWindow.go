package transport

import (
	"container/ring"
	//"fmt"
	"logging"
	"sync"
)

type ReceiverSlidingWindow struct {
	lastByteRead     *ring.Ring
	nextByteExpected *ring.Ring
	lastByteReceived *ring.Ring

	buffer     *ring.Ring
	bufferSize int
	Lock       *sync.RWMutex
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

func (w *ReceiverSlidingWindow) AdvertisedWindowSize() int {
	if w.lastByteReceived == w.lastByteRead {
		return w.bufferSize
	} else {
		return w.bufferSize - (Distance(w.lastByteRead, w.nextByteExpected.Prev()))
	}
}

func (w *ReceiverSlidingWindow) Receive(seqNum int, payload []byte) int {
	// check whether packet is in the window
	//logging.Logger.Printf("[DEBUG][RecvWindow] payload length: %d\n", len(payload))
	w.Lock.Lock()
	defer w.Lock.Unlock()
	if w.nextByteExpected.Prev().Value != nil {
		windowStart := w.nextByteExpected.Prev().Value.(TcpByte)
		if seqNum <= windowStart.SeqNum || seqNum > windowStart.SeqNum+w.AdvertisedWindowSize() || len(payload) > w.AdvertisedWindowSize() {
			// drop packet, bye byte
			logging.Logger.Printf("[RecvWindow] Dropping out-of-window/too-big packet, seqNum %d\n", seqNum)
			return w.nextByteExpected.Prev().Value.(TcpByte).SeqNum + 1
		}
	}

	var moves int
	//count := 0
	if w.nextByteExpected.Prev().Value == nil {
		moves = 0
		//fmt.Println("times:%d", count)
	} else {
		moves = seqNum - w.nextByteExpected.Prev().Value.(TcpByte).SeqNum
		logging.Logger.Printf("[DEBUG][RecvWindow] moves: %d seqNum: %d nextByteExpected prev: %d\n", moves, seqNum, w.nextByteExpected.Prev().Value.(TcpByte).SeqNum)
	}
	//count++
	ringPointer := w.nextByteExpected
	ringPointer.Move(moves)

	// write payload into ring
	logging.Logger.Printf("[DEBUG][RecvWindow] Writing received payload into buffer, seqNum %d payload length: %d\n", seqNum, len(payload))
	for i := 0; i < len(payload); i++ {
		ringPointer.Value = TcpByte{seqNum + i, payload[i]}
		//logging.Logger.Printf("[DEBUG][RecvWindow] Recv window position %d, data %s\n", ringPointer.Value.(TcpByte).SeqNum, string(ringPointer.Value.(TcpByte).B))
		ringPointer = ringPointer.Next()
	}

	// udpate last byte received pointers
	w.lastByteReceived = ringPointer.Prev()

	// update next expected pointer
	for {
		if w.nextByteExpected.Value == nil {
			break
		} else {
			w.nextByteExpected = w.nextByteExpected.Next()
		}
	}
	logging.Logger.Printf("[DEBUG][RecvWindow] Recv window nextByteExpected pointer is right after %d\n", w.nextByteExpected.Prev().Value.(TcpByte).SeqNum)

	// return next expected ACK
	return w.nextByteExpected.Prev().Value.(TcpByte).SeqNum + 1
}

func (w *ReceiverSlidingWindow) Read(bytes int) ([]byte, int) {
	w.Lock.RLock()
	defer w.Lock.RUnlock()
	readableBytes := Distance(w.lastByteRead, w.nextByteExpected.Prev())
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
				logging.Logger.Printf("[DEBUG][RecvWindow] sequence not contiguous!first seq:%d next seq:%d\n", w.lastByteRead.Next().Value.(TcpByte).SeqNum, w.lastByteRead.Next().Next().Value.(TcpByte).SeqNum)
			}
		}
		buffer[i] = w.lastByteRead.Next().Value.(TcpByte).B
		w.lastByteRead.Next().Value = nil
		w.lastByteRead = w.lastByteRead.Next()
		//logging.Logger.Printf("****Window Read****")
	}

	return buffer, bytesToRead
}
