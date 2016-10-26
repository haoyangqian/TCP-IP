package runner

// import "fmt"
import "model"
import "network"

type NetworkRunner struct {
	networkAccessor network.NetworkAccessor

	messageReceiver <-chan model.SendMessageRequest
	chFromLink      <-chan model.LinkReceiveResult
	chToLink        chan<- model.SendPacketRequest
}

func (runner *NetworkRunner) Run() {
	for {
		select {
		case request := <-runner.messageReceiver:
			runner.networkAccessor.Send(request, runner.chToLink)
		case result := <-runner.chFromLink:
			runner.networkAccessor.ReceiveAndHandle(result, runner.chToLink)
		}
	}
}

func (runner *NetworkRunner) GetNetworkAccess() network.NetworkAccessor {
	return runner.networkAccessor
}

func MakeNetworkRunner(
	networkAccessor network.NetworkAccessor,
	messageReceiver <-chan model.SendMessageRequest,
	chFromLink <-chan model.LinkReceiveResult,
	chToLink chan<- model.SendPacketRequest) NetworkRunner {

	return NetworkRunner{networkAccessor, messageReceiver, chFromLink, chToLink}
}
