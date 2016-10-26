package runner

import "model"
import "network"

type LinkSendRunner struct {
	linkAccessor network.LinkAccessor
	chFromNet    <-chan model.SendPacketRequest
}

func MakeLinkSendRunner(linkAccessor network.LinkAccessor, chFromNet <-chan model.SendPacketRequest) LinkSendRunner {
	return LinkSendRunner{linkAccessor, chFromNet}
}

func (runner *LinkSendRunner) Run() {
	for {
		request := <-runner.chFromNet
		runner.linkAccessor.Send(request)
	}
}
