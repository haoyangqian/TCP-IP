package transport

import (
	"logging"
	"sync"
	//"time"
)

type ArrayBasedReceiverSlidingWindow struct {
	lastByteRead       int
	nextSeqNumExpected int
	nextByteToRead     int

	buffer     []byte
	dirties    []bool
	bufferSize int

	BytesInBuffer int

	Lock *sync.RWMutex
}

func MakeArrayBasedReceiverSlidingWindow(bufferSize int) ArrayBasedReceiverSlidingWindow {
	return ArrayBasedReceiverSlidingWindow{
		buffer:     make([]byte, bufferSize),
		dirties:    make([]bool, bufferSize),
		bufferSize: bufferSize,
		Lock:       &sync.RWMutex{},
	}
}

func (w *ArrayBasedReceiverSlidingWindow) SetNextExpectedSeqNum(seqNum int) {
	w.nextSeqNumExpected = seqNum
	w.nextByteToRead = seqNum
}

func (w *ArrayBasedReceiverSlidingWindow) advertisedWindowSize() int {
	return w.bufferSize - (w.nextSeqNumExpected - w.nextByteToRead)
}

func (w *ArrayBasedReceiverSlidingWindow) AdvertisedWindowSize() int {
	w.Lock.RLock()
	defer w.Lock.RUnlock()
	return w.advertisedWindowSize()
}

func (w *ArrayBasedReceiverSlidingWindow) Receive(seqNum int, payload []byte) int {
	w.Lock.Lock()
	defer w.Lock.Unlock()

	//firstClock := time.Now().UnixNano()
	//logging.Printf("[RECV][TIME][00000] start time %d", firstClock)

	payloadSize := len(payload)

	if w.outOfWindow(seqNum, payloadSize) {
		logging.Printf("[RecvWindow] Dropping out-of-window/too-big packet, seqNum %d, expected seqNum is %d, windowsize is %d, payload size is %d\n", seqNum, w.nextSeqNumExpected, w.advertisedWindowSize(), len(payload))
		return w.nextSeqNumExpected
	}

	targetPos := w.nextSeqNumExpected + (seqNum - w.nextSeqNumExpected)
	if targetPos < 0 {
		logging.Printf("[FATAL] targetPos is calculated as %d, r %d e %d\n", targetPos, seqNum, w.nextSeqNumExpected)
	}

	//secondClock := time.Now().UnixNano()
	//logging.Printf("[RECV][TIME][1] secondClock %d", secondClock-firstClock)

	logging.Printf("Writing %d bytes into the buffer starting at %d(%d)\n", payloadSize, targetPos, targetPos%w.bufferSize)
	for i := 0; i < payloadSize; i++ {
		//		logging.Printf("Writing byte into pos %d(%d)\n", i+targetPos, (i+targetPos)%w.bufferSize)
		w.buffer[(i+targetPos)%w.bufferSize] = payload[i]
		w.dirties[(i+targetPos)%w.bufferSize] = true
		w.BytesInBuffer += 1
	}

	//thirdClock := time.Now().UnixNano()
	//logging.Printf("[RECV][TIME][2] thirdClock %d", thirdClock-secondClock)

	// update next byte expected
	logging.Printf("Expected Seqnum %d, received Seqnum %d\n", w.nextSeqNumExpected, seqNum)
	if seqNum == w.nextSeqNumExpected {
		w.nextSeqNumExpected += payloadSize
		//logging.Printf("added payload size %d to seqNum, now %d(%d), nextByteToRead is %d(%d)\n", payloadSize, w.nextSeqNumExpected, w.nextSeqNumExpected%w.bufferSize, w.nextByteToRead, w.nextByteToRead%w.bufferSize)
		//logging.Printf("%t ", w.dirties[w.nextSeqNumExpected%w.bufferSize])
		//logging.Printf("%t \n", w.dirties[(w.nextSeqNumExpected+1)%w.bufferSize])
		for {
			if !w.dirties[w.nextSeqNumExpected%w.bufferSize] || w.nextSeqNumExpected%w.bufferSize == w.nextByteToRead%w.bufferSize {
				break
			} else {
				w.nextSeqNumExpected += 1
				//				logging.Printf("moving next expected seq number to %d(%d)\n", w.nextSeqNumExpected, w.nextSeqNumExpected%w.bufferSize)
			}
		}
	}

	//fourthClock := time.Now().UnixNano()
	//logging.Printf("[RECV][TIME][3] fourthClock %d", fourthClock-thirdClock)
	//logging.Printf("[RECV][TIME] totalTime %d", fourthClock-firstClock)

	//logging.Printf("Next expected seqNum is %d, window size is %d\n", w.nextSeqNumExpected, w.AdvertisedWindowSize())
	return w.nextSeqNumExpected
}

func (w *ArrayBasedReceiverSlidingWindow) Read(bytes int) ([]byte, int) {
	w.Lock.Lock()
	defer w.Lock.Unlock()

	if w.nextSeqNumExpected-w.nextByteToRead == 0 {
		return []byte{}, 0
	}

	readableBytes := w.nextSeqNumExpected - w.nextByteToRead

	if readableBytes > 0 {
		//		logging.Printf("[DEBUG][Read] %d readable bytes in buffer\n", readableBytes)
	}
	var bytesToRead int
	if readableBytes > bytes {
		bytesToRead = bytes
	} else {
		bytesToRead = readableBytes
	}

	readBuffer := make([]byte, bytesToRead)
	for i := 0; i < bytesToRead; i++ {
		readBuffer[i] = w.buffer[(i+w.nextByteToRead)%w.bufferSize]
		w.dirties[(i+w.nextByteToRead)%w.bufferSize] = false
		w.BytesInBuffer -= 1
	}
	w.nextByteToRead += bytesToRead

	return readBuffer, bytesToRead
}

func (w *ArrayBasedReceiverSlidingWindow) GetAck(seqNum int, payloadSize int) int {
	w.Lock.RLock()
	defer w.Lock.RUnlock()

	if w.outOfWindow(seqNum, payloadSize) {
		return w.nextSeqNumExpected
	}

	calculatedAck := w.nextSeqNumExpected
	if seqNum == w.nextSeqNumExpected {
		calculatedAck += payloadSize
		for {
			if !w.dirties[calculatedAck%w.bufferSize] || calculatedAck%w.bufferSize == w.nextByteToRead%w.bufferSize {
				break
			} else {
				calculatedAck += 1
			}
		}
	}
	return calculatedAck
}

func (w *ArrayBasedReceiverSlidingWindow) outOfWindow(seqNum int, payloadSize int) bool {
	return seqNum < w.nextSeqNumExpected || seqNum > (w.nextSeqNumExpected+w.advertisedWindowSize()) || payloadSize > w.advertisedWindowSize()
}
