package runner

import "../model"
import "../network"

type LinkReceiveRunner struct {
	linkAccessor network.LinkAccessor
	chToNet      chan<- model.IpPacket
}

func MakeLinkReceiveRunner(linkAccessor network.LinkAccessor, chToNet chan<- model.IpPacket) LinkReceiveRunner {
	return LinkReceiveRunner{linkAccessor, chToNet}
}

func (runner *LinkReceiveRunner) Run() {
	for {
		packet, err := runner.linkAccessor.Receive()
		if err == nil {
			runner.chToNet <- packet
		}
	}
}
