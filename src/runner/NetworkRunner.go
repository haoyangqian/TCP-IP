package runner

// import "fmt"
import "../model"
import "../network"

const TEST_DATA_PROTOCOL = 0
const RIP_PROTOCOL = 200

type NetworkRunner struct {
	routingTable    model.RoutingTable
	interfaces      map[model.VirtualIp]model.NodeInterface
	networkAccessor network.NetworkAccessor
}

func (runner *NetworkRunner) Run() {
	defer runner.networkAccessor.CloseConnection()
	for {
		runner.networkAccessor.ReceiveAndHandle()
	}
}

func (runner *NetworkRunner) GetNetworkAccess() network.NetworkAccessor {
	return runner.networkAccessor
}

func MakeNetworkRunner(table model.RoutingTable, interfaces map[model.VirtualIp]model.NodeInterface, service string) NetworkRunner {
	ipHandler := network.IpHandler{}
	linkAccessor := network.NewLinkAccessor(interfaces, service)

	networkAccessor := network.NewNetworkAccessor(linkAccessor, table)
	networkAccessor.RegisterHandler(TEST_DATA_PROTOCOL, ipHandler)

	return NetworkRunner{table, interfaces, networkAccessor}
}
