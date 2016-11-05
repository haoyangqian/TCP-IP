package transport

import ()

type PacketInFlight struct {
	ExpireTimeNanos int64
	Packet          TcpPacket
	HasAcked        bool
}

func (p *PacketInFlight) MarkAsAcked() {
	p.HasAcked = true
}

//type SenderSlidingWindow struct {
//	lastByteAcked   int
//	lastByteSent    int
//	lastByteWritten int
//
//	buffer     []byte
//	bufferSize int
//
//	lastAdvertisedWindowSize int
//	effectiveWindowSize      int
//
//	packetsInFlight []PacketInFlight
//}
//
///*
// * return whether there are unsent data in the buffer
// */
//func (w *SenderSlidingWindow) HasUnsentData() bool {
//}
//
//func (w *SenderSlidingWindow) GetUnsentData() []byte {
//}
//
//func (w *SenderSlidingWindow) AddPacketInFlight(packet PacketInFlight) {
//}
