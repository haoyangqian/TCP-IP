package factory

import "model"
import "network"
import "runner"
import "transport"

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

	socketManager *transport.SocketManager
	socketRunner  runner.SocketRunner
}

func InitializeResourceFactory(routingTable model.RoutingTable, interfaces map[model.VirtualIp]*model.NodeInterface, service string) ResourceFactory {

	nodeInterfaceTable := model.MakeNodeInterfaceTable(interfaces)

	// channels
	messageChannel := make(chan model.SendMessageRequest)
	netToLinkChannel := make(chan model.SendPacketRequest)
	linkToNetChannel := make(chan model.LinkReceiveResult)
	ipToTcpChannel := make(chan model.IpPacket)

	// handlers
	ripHandler := network.MakeRipHandler(routingTable, messageChannel)
	ipHandler := network.IpHandler{ipToTcpChannel}

	// network
	networkAccessor := network.NewNetworkAccessor(routingTable)
	networkAccessor.RegisterHandler(model.RIP_PROTOCOL, ripHandler)
	networkAccessor.RegisterHandler(model.TEST_DATA_PROTOCOL, ipHandler)

	// link
	linkAccessor := network.NewLinkAccessor(nodeInterfaceTable, service)
	linkReceiveRunner := runner.MakeLinkReceiveRunner(linkAccessor, linkToNetChannel)
	linkSendRunner := runner.MakeLinkSendRunner(linkAccessor, netToLinkChannel)

	// runner
	networkRunner := runner.MakeNetworkRunner(networkAccessor, messageChannel, linkToNetChannel, netToLinkChannel)
	ripRunner := runner.MakeRipRunner(ripHandler, messageChannel)

	// transport & sockets
	fmsBuilder := makeTcpFsmBuilder()
	socketManager := transport.MakeSocketManager(interfaces, fmsBuilder, messageChannel)
	socketRunner := runner.MakeSocketRunner(&socketManager, ipToTcpChannel)

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

	factory.socketManager = &socketManager
	factory.socketRunner = socketRunner
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

func (factory *ResourceFactory) SocketManager() *transport.SocketManager {
	return factory.socketManager
}

func (factory *ResourceFactory) SocketRunner() runner.SocketRunner {
	return factory.socketRunner
}

func makeTcpFsmBuilder() transport.TcpStateMachineBuilder {
	builder := transport.MakeTcpStateMachineBuilder(transport.TCP_INITIAL_CLOSED)

	builder.RegisterTransition(transport.TCP_INITIAL_CLOSED, transport.TCP_PASSIVE_OPEN, transport.TCP_RESP_DO_NOTHING, transport.TCP_LISTEN)

	builder.RegisterTransition(transport.TCP_INITIAL_CLOSED, transport.TCP_ACTIVE_OPEN, transport.TCP_RESP_SEND_SYN, transport.TCP_SYN_SENT)

	builder.RegisterTransition(transport.TCP_LISTEN, transport.TCP_CLOSE, transport.TCP_RESP_DEL_SOCK, transport.TCP_INITIAL_CLOSED)
	builder.RegisterTransition(transport.TCP_LISTEN, transport.TCP_RECV_SYN, transport.TCP_RESP_SEND_SYN_ACK, transport.TCP_SYN_RCVD)
	builder.RegisterTransition(transport.TCP_LISTEN, transport.TCP_SEND, transport.TCP_RESP_SEND_SYN, transport.TCP_SYN_SENT)

	builder.RegisterTransition(transport.TCP_SYN_RCVD, transport.TCP_CLOSE, transport.TCP_RESP_SEND_FIN, transport.TCP_FIN_WAIT_1)
	builder.RegisterTransition(transport.TCP_SYN_RCVD, transport.TCP_RECV_ACK, transport.TCP_RESP_DO_NOTHING, transport.TCP_ESTAB)

	builder.RegisterTransition(transport.TCP_SYN_SENT, transport.TCP_CLOSE, transport.TCP_RESP_DEL_SOCK, transport.TCP_INITIAL_CLOSED)
	builder.RegisterTransition(transport.TCP_SYN_SENT, transport.TCP_RECV_SYN, transport.TCP_RESP_SEND_SYN_ACK, transport.TCP_SYN_RCVD)
	builder.RegisterTransition(transport.TCP_SYN_SENT, transport.TCP_RECV_SYN_ACK, transport.TCP_RESP_SEND_ACK, transport.TCP_ESTAB)

	builder.RegisterTransition(transport.TCP_ESTAB, transport.TCP_CLOSE, transport.TCP_RESP_SEND_FIN, transport.TCP_FIN_WAIT_1)
	builder.RegisterTransition(transport.TCP_ESTAB, transport.TCP_RECV_FIN, transport.TCP_RESP_SEND_ACK, transport.TCP_CLOSE_WAIT)

	builder.RegisterTransition(transport.TCP_FIN_WAIT_1, transport.TCP_RECV_ACK, transport.TCP_RESP_DO_NOTHING, transport.TCP_FIN_WAIT_2)
	builder.RegisterTransition(transport.TCP_FIN_WAIT_1, transport.TCP_RECV_FIN, transport.TCP_RESP_SEND_ACK, transport.TCP_CLOSING)

	builder.RegisterTransition(transport.TCP_FIN_WAIT_2, transport.TCP_RECV_FIN, transport.TCP_RESP_SEND_ACK, transport.TCP_TIME_WAIT)

	builder.RegisterTransition(transport.TCP_CLOSING, transport.TCP_RECV_ACK, transport.TCP_RESP_DO_NOTHING, transport.TCP_TIME_WAIT)

	builder.RegisterTransition(transport.TCP_TIME_WAIT, transport.TCP_TIMEOUT_2MSL, transport.TCP_RESP_DEL_SOCK, transport.TCP_FINAL_CLOSED)

	builder.RegisterTransition(transport.TCP_CLOSE_WAIT, transport.TCP_CLOSE, transport.TCP_RESP_SEND_FIN, transport.TCP_LAST_ACK)

	builder.RegisterTransition(transport.TCP_LAST_ACK, transport.TCP_RECV_ACK, transport.TCP_RESP_DO_NOTHING, transport.TCP_FINAL_CLOSED)

	return builder
}
