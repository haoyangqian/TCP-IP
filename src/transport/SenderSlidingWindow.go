package transport

import (
	//"fmt"
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
	lastByteAcked            int
	lastByteSent             int
	lastByteWritten          int
	BytesToSend              int
	BytesInFlight            int
	lastSeqnumAcked          int
	returnSeqNum             int
	buffer                   []byte
	dirty                    []bool
	bufferSize               int
	lastAdvertisedWindowSize int
	Lock                     *sync.RWMutex
}

func MakeSenderSlidingWindow(bufferSize int, seq int) SenderSlidingWindow {
	newbuffer := make([]byte, bufferSize)
	newdirty := make([]bool, bufferSize)

	return SenderSlidingWindow{
		lastByteAcked:            0,
		lastByteSent:             0,
		lastByteWritten:          0,
		BytesToSend:              0,
		BytesInFlight:            0,
		lastSeqnumAcked:          seq,
		returnSeqNum:             seq,
		buffer:                   newbuffer,
		dirty:                    newdirty,
		bufferSize:               bufferSize,
		lastAdvertisedWindowSize: 65535,
		Lock: &sync.RWMutex{},
	}
}

func (w *SenderSlidingWindow) EffectiveWindowSize() int {
	//effective window size = lastAdvertisedWindowSize - len(bytes in flight)
	return w.lastAdvertisedWindowSize - w.BytesInFlight
}

func (w *SenderSlidingWindow) AvailableWriteSpace() int {

	return w.bufferSize - w.BytesToSend
}

func (w *SenderSlidingWindow) Send() ([]byte, int) {
	w.Lock.RLock()
	defer w.Lock.RUnlock()
	//check the length of bytes should be sent
	if w.BytesToSend == 0 {
		//logging.Printf("[SendWindow] No Bytes to send\n")
		return []byte{}, -1
	}
	logging.Printf("1 BytesToSend:%d EffectiveWindowSize:%d\n", w.BytesToSend, w.EffectiveWindowSize())
	//check the length of bytes can be sent
	if w.EffectiveWindowSize() < 0 {
		return []byte{}, -2
	}
	//logging.Printf("[DEBUG][SendWindow] Send() BytesToSent:%d , EffectiveWindowSize : %d", w.BytesToSend, w.EffectiveWindowSize())
	sendsize := w.EffectiveWindowSize()
	logging.Printf("2 sendsize:%d BytesToSend:%d EffectiveWindowSize:%d\n", sendsize, w.BytesToSend, w.EffectiveWindowSize())
	if w.BytesToSend < w.EffectiveWindowSize() {
		sendsize = w.BytesToSend
	}
	//can only send MAX_PAYLOAD at once in a tcppacket
	if sendsize > MAX_PAYLOAD {
		sendsize = MAX_PAYLOAD
	}

	if w.lastAdvertisedWindowSize == 0 {
		//send probing bytes
		sendsize = 1
	}
	//send bytes
	logging.Printf("3 sendsize:%d BytesToSend:%d EffectiveWindowSize:%d\n", sendsize, w.BytesToSend, w.EffectiveWindowSize())
	buffer := make([]byte, sendsize)

	returnseqnum := w.returnSeqNum
	for i := 0; i < sendsize; i++ {
		w.lastByteSent = w.lastByteSent + 1
		if w.lastByteSent >= w.bufferSize {
			w.lastByteSent = 0
		}
		buffer[i] = w.buffer[w.lastByteSent]
		//		if w.dirty[w.lastByteSent] {
		//			buffer[i] = w.buffer[w.lastByteSent]
		//			w.dirty[w.lastByteSent] = false
		//		}
	}
	//	if len(buffer) > 0 {
	//		logging.Printf("[DEBUG][SendWindow] Send() length:%d \n", len(buffer))
	//	}
	w.BytesToSend -= len(buffer)
	logging.Printf("[DEBUG][SendWindow] Send() BytesToSend:%d\n", w.BytesToSend)
	w.returnSeqNum += sendsize
	return buffer, returnseqnum
}

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

	for i := 0; i < writelength; i++ {
		w.lastByteWritten = w.lastByteWritten + 1
		if w.lastByteWritten >= w.bufferSize {
			w.lastByteWritten = 0
		}
		w.buffer[w.lastByteWritten] = buff[i]
		w.BytesToSend += 1
		//		if w.dirty[w.lastByteWritten] {
		//			logging.Printf("[DEBUG][SendWindow] Overwrite buffer")
		//		} else {
		//			w.buffer[w.lastByteWritten] = buff[i]
		//			w.dirty[w.lastByteWritten] = true
		//			//logging.Printf("[SendWindow] Write buffer:%s,seq:%d\n", string(w.lastByteWritten.Next().Value.(TcpByte).B), w.lastByteWritten.Next().Value.(TcpByte).SeqNum)
		//			w.BytesToSend += 1
		//		}
	}
	//logging.Printf("[DEBUG][SendWindow] write length: %d distance: %d", writelength, w.BytesToSend)
	return writelength
}
