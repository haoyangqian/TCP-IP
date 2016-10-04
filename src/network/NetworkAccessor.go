package network

import "../model"

type NetworkAccessor struct {
}

func (accessor *NetworkAccessor) Receive() {
	// to be implemented
}

func (accessor *NetworkAccessor) SendTestData(message string, dest model.VirtualIp) {
	// to be implemented
}

// func (accessor *NetworkLayerAccessor) SendRipData(message model.RipMessage, dest model.VirtualIp) {
//     // to be implemented
// }
