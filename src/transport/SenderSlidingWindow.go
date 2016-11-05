package transport

import (
	"container/ring"
)

type PacketInFlight struct {
	Index           int
	ExpireTimeNanos int64
	Packet          TcpPacket
	HasAcked        bool
	ExpectedAckNum  int
}

func (p *PacketInFlight) MarkAsAcked() {
	p.HasAcked = true
}

type SenderSlidingWindow struct {
	lastByteAcked   int
	lastByteSent    int
	lastByteWritten int

	buffer     *ring.Ring
	bufferSize int

	lastAdvertisedWindowSize int

	packetsInFlight []PacketInFlight
}

func (w *SenderSlidingWindow) EffectiveWindowSize() int {
	return 0
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
