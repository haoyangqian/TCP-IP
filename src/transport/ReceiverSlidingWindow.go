package transport

import (
	"container/ring"
	"logging"
)

type ReceiverSlidingWindow struct {
	lastByteRead     *ring.Ring
	nextByteExpected *ring.Ring
	lastByteReceived *ring.Ring

	buffer     *ring.Ring
	bufferSize int
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
	}
}

func (w *ReceiverSlidingWindow) AdvertisedWindowSize() int {
	if w.lastByteReceived == w.lastByteRead {
		return w.bufferSize
	} else {
		return Distance(w.lastByteReceived, w.lastByteRead) - 1
	}
}

func (w *ReceiverSlidingWindow) Receive(seqNum int, payload []byte) int {
	// check whether packet is in the window
	if w.lastByteRead.Value != nil {
		windowStart := w.lastByteRead.Value.(TcpByte)
		if seqNum < windowStart.SeqNum || seqNum > windowStart.SeqNum+w.AdvertisedWindowSize() || len(payload) > w.AdvertisedWindowSize() {
			// drop packet, bye byte
			logging.Logger.Printf("[RecvWindow] Dropping out-of-window/too-big packet, seqNum %d\n", seqNum)
			return w.nextByteExpected.Value.(TcpByte).SeqNum
		}
	}

	var moves int
	if w.nextByteExpected.Value == nil {
		moves = 0
	} else {
		moves = seqNum - w.nextByteExpected.Value.(TcpByte).SeqNum
	}
	ringPointer := w.nextByteExpected
	ringPointer.Move(moves)

	// write payload into ring
	logging.Logger.Printf("[DEBUG][RecvWindow] Writing received payload into buffer, seqNum %d\n", seqNum)
	for i := 0; i < len(payload); i++ {
		ringPointer.Value = TcpByte{seqNum + i, payload[i]}
		logging.Logger.Printf("[DEBUG][RecvWindow] Recv window position %d, data %s\n", ringPointer.Value.(TcpByte).SeqNum, string(ringPointer.Value.(TcpByte).B))
		ringPointer = ringPointer.Next()
	}

	// udpate last byte received pointers
	w.lastByteReceived = ringPointer

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
	readableBytes := Distance(w.lastByteRead, w.nextByteExpected.Prev())
	var bytesToRead int
	if readableBytes > bytes {
		bytesToRead = bytes
	} else {
		bytesToRead = readableBytes
	}

	buffer := make([]byte, bytesToRead)
	for i := 0; i < bytesToRead; i++ {
		buffer[i] = w.lastByteRead.Next().Value.(TcpByte).B
		w.lastByteRead = w.lastByteRead.Next()
	}

	return buffer, bytesToRead
}
