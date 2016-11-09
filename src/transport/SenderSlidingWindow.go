package transport

import (
	"container/ring"
	"logging"
	"sync"
)

type PacketInFlight struct {
	Index           int
	ExpireTimeNanos int64
	Packet          *TcpPacket
	ExpectedAckNum  int
}

type SenderSlidingWindow struct {
	lastByteAcked            *ring.Ring
	lastByteSent             *ring.Ring
	lastByteWritten          *ring.Ring
	Seqnum                   int
	lastSeqnumAcked          int
	buffer                   *ring.Ring
	bufferSize               int
	lastAdvertisedWindowSize int
	Lock                     *sync.RWMutex
}

func MakeSenderSlidingWindow(bufferSize int, seq int) SenderSlidingWindow {
	ring := ring.New(bufferSize)
	lastByteAcked := ring
	lastByteSent := ring
	lastByteWritten := ring

	return SenderSlidingWindow{
		lastByteAcked:            lastByteAcked,
		lastByteSent:             lastByteSent,
		lastByteWritten:          lastByteWritten,
		Seqnum:                   seq,
		lastSeqnumAcked:          seq,
		buffer:                   ring,
		bufferSize:               bufferSize,
		lastAdvertisedWindowSize: 65535,
		Lock: &sync.RWMutex{},
	}
}

/*
*   return the EffectiveWindowSize
 */
func (w *SenderSlidingWindow) EffectiveWindowSize() int {
	//effective window size = lastAdvertisedWindowSize - len(bytes in flight)
	return w.lastAdvertisedWindowSize - Distance(w.lastByteSent, w.lastByteAcked)
}

/*
*   return the AvailableWriteSpace
 */
func (w *SenderSlidingWindow) AvailableWriteSpace() int {
	var space int
	//the whole buffer
	if w.lastByteAcked == w.lastByteWritten {
		space = w.bufferSize
	} else {
		space = Distance(w.lastByteWritten, w.lastByteAcked)
	}
	return space
}

/*
*   return the BytesToSent
 */
func (w *SenderSlidingWindow) BytesToSent() int {
	var length int
	//the whole buffer
	if w.lastByteWritten == w.lastByteSent {
		length = 0
	} else {
		length = Distance(w.lastByteSent, w.lastByteWritten)
		//logging.Logger.Printf("[SendWindow] Have Bytes to send: %d\n", length)
	}
	return length
}

/*
*     send bytes to the receiver
*     return send buffer and seqnum
 */
func (w *SenderSlidingWindow) Send() ([]byte, int) {
	w.Lock.RLock()
	defer w.Lock.RUnlock()
	//check the length of bytes should be sent
	if w.BytesToSent() == 0 {
		//logging.Logger.Printf("[SendWindow] No Bytes to send\n")
		return []byte{}, -1
	}
	//check the length of bytes can be sent
	if w.EffectiveWindowSize() == 0 {
		//send 1-byte probing packet
		//logging.Logger.Printf("[SendWindow] EffectiveWindowSize == 0 \n")
		return []byte{}, -2
	}
	logging.Logger.Printf("[DEBUG][SendWindow] Send() BytesToSent:%d , EffectiveWindowSize : %d", w.BytesToSent(), w.EffectiveWindowSize())
	sendsize := w.EffectiveWindowSize()
	if w.BytesToSent() < w.EffectiveWindowSize() {
		sendsize = w.BytesToSent()
	}
	//can only send MAX_PAYLOAD at once in a tcppacket
	if sendsize > MAX_PAYLOAD {
		sendsize = MAX_PAYLOAD
	}

	//send bytes
	buffer := make([]byte, sendsize)
	seqnum := w.lastByteSent.Next().Value.(TcpByte).SeqNum
	for i := 0; i < sendsize; i++ {
		if w.lastByteSent.Next().Value != nil {
			buffer[i] = w.lastByteSent.Next().Value.(TcpByte).B
			w.lastByteSent.Next().Value = nil
			w.lastByteSent = w.lastByteSent.Next()
		}
	}
	if len(buffer) > 0 {
		logging.Logger.Printf("[DEBUG][SendWindow] Send() length:%d buffer:%s \n", len(buffer), string(buffer))
	}
	return buffer, seqnum
}

/*
*    write bytes into ringbuffer, do not block
*    return the number of bytes which is written into buffer successfully
 */
func (w *SenderSlidingWindow) Write(buff []byte, nbytes int) int {
	w.Lock.Lock()
	defer w.Lock.Unlock()
	var writelength int
	//check if there are enough space
	if nbytes <= w.AvailableWriteSpace() {
		writelength = nbytes
	} else {
		writelength = w.AvailableWriteSpace()
	}

	//write bytes into buffer starting from lastbyteswritten.next()
	for i := 0; i < writelength; i++ {
		if w.lastByteWritten.Next().Value != nil {
			logging.Logger.Printf("[DEBUG][SendWindow] Overwrite buffer seqnum:%d byte:%s\n", w.lastByteWritten.Next().Value.(TcpByte).SeqNum, w.lastByteWritten.Next().Value.(TcpByte).B)
		} else {
			w.lastByteWritten.Next().Value = TcpByte{w.Seqnum, buff[i]}
			//logging.Logger.Printf("[SendWindow] Write buffer:%s,seq:%d\n", string(w.lastByteWritten.Next().Value.(TcpByte).B), w.lastByteWritten.Next().Value.(TcpByte).SeqNum)
			w.Seqnum += 1
			w.lastByteWritten = w.lastByteWritten.Next()
		}
	}
	logging.Logger.Printf("[DEBUG][SendWindow] write length: %d distance: %d", writelength, w.BytesToSent())
	return writelength
}

/*
 * return whether there are unsent data in the buffer
 */
func (w *SenderSlidingWindow) HasUnsentData() bool {
	return false
}

/*
 * return a byte array representing data to be sent while respecting the effective window size
 */
func (w *SenderSlidingWindow) GetUnsentData() []byte {
	return []byte{}
}

/*
 * register a new packet that has been sent out but have not been ACKed yet
 */
func (w *SenderSlidingWindow) AddPacketInFlight(tcpPacket TcpPacket, expectedAckNum int) {
}

/*
 * notify the sliding window with a received ACK
 */
func (w *SenderSlidingWindow) Acknowledge(ackNum int, lastAdvertisedWindow int) {
}

/*
 * returns true if there exist a packet in flight that has timed out
 */
func (w *SenderSlidingWindow) HasTimedoutInFlightPacket() bool {
	return false
}

/*
 * get the next expired packet in flight
 */
func (w *SenderSlidingWindow) GetNextExpiredPacket() (PacketInFlight, bool) {
	var p PacketInFlight
	return p, false
}
