package runner

import "../model"
import "../network"

type LinkReceiveRunner struct {
	linkAccessor network.LinkAccessor
	chToNet      chan<- model.LinkReceiveResult
}

func MakeLinkReceiveRunner(linkAccessor network.LinkAccessor, chToNet chan<- model.LinkReceiveResult) LinkReceiveRunner {
	return LinkReceiveRunner{linkAccessor, chToNet}
}

func (runner *LinkReceiveRunner) Run() {
	for {
		packet, receivedFrom, err := runner.linkAccessor.Receive()
		if err == nil {
			runner.chToNet <- model.LinkReceiveResult{packet, receivedFrom}
		}
	}
}
