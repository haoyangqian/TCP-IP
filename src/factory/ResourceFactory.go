package factory

import "model"
import "network"
import "runner"

type ResourceFactory struct {
	routingTable       model.RoutingTable
	interfaces         map[model.VirtualIp]*model.NodeInterface
	nodeInterfaceTable model.NodeInterfaceTable

	linkAccessor    network.LinkAccessor
	networkAccessor network.NetworkAccessor

	messageChannel   chan model.SendMessageRequest
	netToLinkChannel chan model.SendPacketRequest
	linkToNetChannel chan model.LinkReceiveResult

	linkReceiveRunner runner.LinkReceiveRunner
	linkSendRunner    runner.LinkSendRunner
	networkRunner     runner.NetworkRunner
	ripRunner         runner.RipRunner
}

func InitializeResourceFactory(routingTable model.RoutingTable, interfaces map[model.VirtualIp]*model.NodeInterface, service string) ResourceFactory {

	nodeInterfaceTable := model.MakeNodeInterfaceTable(interfaces)

	messageChannel := make(chan model.SendMessageRequest)
	netToLinkChannel := make(chan model.SendPacketRequest)
	linkToNetChannel := make(chan model.LinkReceiveResult)
	ipTotcpChannel := make(chan model.IpPacket)

	ripHandler := network.MakeRipHandler(routingTable, messageChannel)
	ipHandler := network.IpHandler{ipTotcpChannel}

	networkAccessor := network.NewNetworkAccessor(routingTable)
	networkAccessor.RegisterHandler(model.RIP_PROTOCOL, ripHandler)
	networkAccessor.RegisterHandler(model.TEST_DATA_PROTOCOL, ipHandler)

	linkAccessor := network.NewLinkAccessor(nodeInterfaceTable, service)
	linkReceiveRunner := runner.MakeLinkReceiveRunner(linkAccessor, linkToNetChannel)
	linkSendRunner := runner.MakeLinkSendRunner(linkAccessor, netToLinkChannel)

	networkRunner := runner.MakeNetworkRunner(networkAccessor, messageChannel, linkToNetChannel, netToLinkChannel)
	ripRunner := runner.MakeRipRunner(ripHandler, messageChannel)

	//socketRunner := runner.MakeSocketRunner()

	factory := ResourceFactory{}
	factory.routingTable = routingTable
	factory.interfaces = interfaces
	factory.nodeInterfaceTable = nodeInterfaceTable
	factory.linkAccessor = linkAccessor
	factory.networkAccessor = networkAccessor
	factory.messageChannel = messageChannel
	factory.netToLinkChannel = netToLinkChannel
	factory.linkToNetChannel = linkToNetChannel
	factory.linkReceiveRunner = linkReceiveRunner
	factory.linkSendRunner = linkSendRunner
	factory.networkRunner = networkRunner
	factory.ripRunner = ripRunner
	return factory
}

func (factory *ResourceFactory) NodeInterfaceTable() model.NodeInterfaceTable {
	return factory.nodeInterfaceTable
}

func (factory *ResourceFactory) LinkAccessor() network.LinkAccessor {
	return factory.linkAccessor
}

func (factory *ResourceFactory) NetworkAccessor() network.NetworkAccessor {
	return factory.networkAccessor
}

func (factory *ResourceFactory) MessageChannel() chan model.SendMessageRequest {
	return factory.messageChannel
}

func (factory *ResourceFactory) NetToLinkChannel() chan model.SendPacketRequest {
	return factory.netToLinkChannel
}

func (factory *ResourceFactory) LinkToNetChannel() chan model.LinkReceiveResult {
	return factory.linkToNetChannel
}

func (factory *ResourceFactory) LinkReceiveRunner() runner.LinkReceiveRunner {
	return factory.linkReceiveRunner
}

func (factory *ResourceFactory) LinkSendRunner() runner.LinkSendRunner {
	return factory.linkSendRunner
}

func (factory *ResourceFactory) NetworkRunner() runner.NetworkRunner {
	return factory.networkRunner
}

func (factory *ResourceFactory) RipRunner() runner.RipRunner {
	return factory.ripRunner
}
