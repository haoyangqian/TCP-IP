package runner

// import "fmt"
import "../model"
import "../network"

type NetworkRunner struct {
	RoutingTable    model.RoutingTable
	Interfaces      map[model.VirtualIp]model.NodeInterface
	NetworkAccessor network.NetworkAccessor
}

func (runner *NetworkRunner) Run() {
	for {
		runner.NetworkAccessor.ReceiveAndHandle()
	}
}

func MakeNetworkRunner(table model.RoutingTable, interfaces map[model.VirtualIp]model.NodeInterface) NetworkRunner {
	ipHandler := network.IpHandler{}
	linkAccessor := network.LinkAccessor{interfaces}
	networkAccessor := network.NewNetworkAccessor(linkAccessor, table, ipHandler)

	return NetworkRunner{table, interfaces, networkAccessor}
}
