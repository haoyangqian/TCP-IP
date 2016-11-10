package transport

import (
	"errors"
	"fmt"
	"logging"
	"math/rand"
	"model"
	"os"
	"text/tabwriter"
	"time"
)

var (
	TCP_HANDSHAKE_MAX_RETRY int = 3
)

type SocketManager struct {
	socketMapByFd   map[int]*SocketRunner
	socketMapByAddr map[SocketAddr]*SocketRunner
	interfacetable  map[model.VirtualIp]bool
	fsmBuilder      TcpStateMachineBuilder
	fdcount         int
	portcount       int
	sendToIpCh      chan<- model.SendMessageRequest
}

type SocketAddr struct {
	LocalIp    model.VirtualIp
	LocalPort  int
	RemoteIp   model.VirtualIp
	RemotePort int
}

func (manager *SocketManager) PrintSockets() {
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 0, '\t', 0)
	fmt.Fprintf(w, "socket\tlocal-addr\tport\tdst-addr\t\tport\tstatus\n")
	fmt.Fprintf(w, "--------------------------------------------------\n")
	for k, v := range manager.socketMapByFd {
		fmt.Fprintf(w, "%d\t%s\t%d\t%s\t\t%d\t%s\n", k, v.Socket.Addr.LocalIp.Ip, v.Socket.Addr.LocalPort, v.Socket.Addr.RemoteIp.Ip, v.Socket.Addr.RemotePort, v.Socket.StateMachine.CurrentState().Name)
	}
	w.Flush()
}

func MakeSocketManager(interfaces map[model.VirtualIp]*model.NodeInterface, fsmBuilder TcpStateMachineBuilder, sendToIpCh chan<- model.SendMessageRequest) SocketManager {
	socketmapfd := make(map[int]*SocketRunner)
	socketmapaddr := make(map[SocketAddr]*SocketRunner)
	interfacetable := make(map[model.VirtualIp]bool)
	for _, v := range interfaces {
		interfacetable[v.Src] = true
	}
	rand.Seed(time.Now().UnixNano())
	return SocketManager{socketmapfd, socketmapaddr, interfacetable, fsmBuilder, -1, 1024, sendToIpCh}
}

func (manager *SocketManager) GetRunnerByAddr(addr SocketAddr) (*SocketRunner, error) {
	//	for k, _ := range manager.socketMapByAddr {
	//		fmt.Printf("get socketbyaddr() : map key: %+v\n", k)
	//	}
	if r, ok := manager.socketMapByAddr[addr]; ok {
		return r, nil
	} else {
		return r, errors.New("No runner found!")
	}
}

func (manager *SocketManager) GetSocketByAddr(addr SocketAddr) (*TcpSocket, error) {
	runner, err := manager.GetRunnerByAddr(addr)
	if err != nil {
		return nil, errors.New("No sockets found!")
	} else {
		return runner.Socket, nil
	}
}

func (manager *SocketManager) GetRunnerByFd(fd int) (*SocketRunner, error) {
	if r, ok := manager.socketMapByFd[fd]; ok {
		return r, nil
	} else {
		return r, errors.New("No sockets found!")
	}
}

func (manager *SocketManager) GetSocketByFd(fd int) (*TcpSocket, error) {
	runner, err := manager.GetRunnerByFd(fd)
	if err != nil {
		return nil, errors.New("No sockets found!")
	} else {
		return runner.Socket, nil
	}
}

func (manager *SocketManager) GetAvailableInterface(port int) (model.VirtualIp, error) {
	for k, _ := range manager.interfacetable {
		_, used := manager.socketMapByAddr[SocketAddr{k, port, model.VirtualIp{"0.0.0.0"}, 0}]
		if !used {
			return k, nil
		}
	}
	return model.VirtualIp{}, errors.New("GetAvailableInterface() error: No available interfaces")
}

func (manager *SocketManager) UpdateRomoteAddr(socket *TcpSocket, addr model.VirtualIp, port int) {
	//	newsocket, err := manager.GetSocketByAddr(SocketAddr{socket.Addr.LocalIp, socket.Addr.LocalPort, model.VirtualIp{"0.0.0.0"}, 0})
	//	if err != nil {
	//		return
	//	}
	//	if newsocket.Fd == socket.Fd {
	//		delete(manager.socketMapByAddr, socket.Addr)
	//	}
	//	socket.SetAddr(SocketAddr{socket.Addr.LocalIp, socket.Addr.LocalPort, addr, port})
	//	manager.socketMapByAddr[socket.Addr] = socket
}

func (manager *SocketManager) InsertRomoteAddr(socket *TcpSocket, laddr model.VirtualIp, lport int, raddr model.VirtualIp, rport int) {
	//	socket.SetAddr(SocketAddr{laddr, lport, raddr, rport})
	//	manager.socketMapByAddr[socket.Addr] = socket
}

func (manager *SocketManager) SetSocketAddr(socketfd int, addr SocketAddr) {
	runner, err := manager.GetRunnerByFd(socketfd)
	if err != nil {
		fmt.Println("Set socket addr fail")
		return
	}
	runner.Socket.Addr = addr
	newsocket, err := manager.GetSocketByAddr(runner.Socket.Addr)
	//find the same sockefd in map
	if err == nil && socketfd == newsocket.Fd {
		delete(manager.socketMapByAddr, runner.Socket.Addr)
	}
	manager.socketMapByAddr[addr] = runner
}

func (manager *SocketManager) V_socket() int {
	//create a new socket
	//manager.PrintSockets()
	manager.fdcount += 1
	stateMachine := manager.fsmBuilder.Build(manager.fdcount)
	socket := MakeSocket(manager.fdcount, stateMachine, manager.sendToIpCh)
	//create a new runner
	runner := MakeSocketRunner(&socket, manager, make(chan model.IpPacket))
	manager.socketMapByFd[manager.fdcount] = &runner
	//manager.PrintSockets()
	return manager.fdcount
}

func (manager *SocketManager) V_bind(socketfd int, addr model.VirtualIp, port int) (int, error) {
	//get the socket from map
	_, err := manager.GetSocketByFd(socketfd)
	if err != nil {
		return -1, errors.New("v_bind() error:Wrong socketfd")
	}

	//if addr is nil/not specified, bind to any available interface
	if addr.Ip == "" {
		//if port is -1, choose a port from 1024 - 65535;
		if port == -1 {
			port = manager.portcount
			for {
				//fmt.Printf("trying to find a new port: %d\n", port)
				vip, err := manager.GetAvailableInterface(port)
				//find available port
				if err == nil {
					manager.portcount += 1
					addr = vip
					break
				} else {
					port = manager.portcount
					manager.portcount += 1
				}
			}
		} else {
			vip, err := manager.GetAvailableInterface(port)
			addr = vip
			if err != nil {
				delete(manager.socketMapByFd, socketfd)
				return -1, errors.New("v_bind() error: Cannot assign requested address")
			}
		}
	}

	//set socket addr
	//check if this addr is available
	socketaddr := SocketAddr{addr, port, model.VirtualIp{"0.0.0.0"}, 0}
	manager.SetSocketAddr(socketfd, socketaddr)
	return 0, nil
}

func (manager *SocketManager) V_listen(socketfd int) int {
	// get socket runner
	// transit socket state
	// starts runner
	runner, _ := manager.GetRunnerByFd(socketfd)
	socket := runner.Socket
	//if not bind, bind a random ip and port to this scoekts
	if socket.Addr.LocalIp.Ip == "" && socket.Addr.LocalPort == 0 {
		manager.V_bind(socket.Fd, model.VirtualIp{}, -1)
	}
	socket.StateMachine.Transit(TCP_PASSIVE_OPEN)
	//listen socket don't need start running
	//	go runner.Run()
	return 0
}

func (manager *SocketManager) V_connect(socketfd int, addr model.VirtualIp, port int) (int, error) {
	runner, _ := manager.GetRunnerByFd(socketfd)
	socket := runner.Socket
	ctrl, _ := socket.StateMachine.GetResponse(TCP_ACTIVE_OPEN)

	//setaddr ,  update record in mapbyaddr
	manager.SetSocketAddr(socketfd, SocketAddr{socket.Addr.LocalIp, socket.Addr.LocalPort, addr, port})
	socket.StateMachine.Transit(TCP_ACTIVE_OPEN)

	//send syn
	_, err := socket.SendCtrl(ctrl.GetCtrlFlags(), int(rand.Int31()), 0, socket.Addr.LocalIp, socket.Addr.LocalPort, addr, port)
	if err != nil {
		return -1, errors.New("v_connect() error: sendctrl() went wrong")
	}

	go runner.Run()

	startTimeMillis := int(time.Now().UnixNano() / int64(time.Millisecond))
	for {
		if socket.StateMachine.CurrentState() == TCP_ESTAB {
			fmt.Println("v_connect() return 0")
			return 0, nil
		}

		// if current time is ahead of start time + 3 timeouts + a small jitter, consider the connection timed out
		if int(time.Now().UnixNano()/int64(time.Millisecond)) > (startTimeMillis + TCP_STATE_DEFAULT_TIMEOUT_MILLIS*TCP_MAX_RETRY_COUNT + 100) {
			return -1, errors.New("v_connect() error: timed out")
		}
	}
}

func (manager *SocketManager) V_accept(listenfd int, addr *model.VirtualIp, port *int) (int, error) {
	//get ip header from channel
	runner, _ := manager.GetRunnerByFd(listenfd)
	//fmt.Println("reading channel... ")
	ipPacket := <-runner.RecvFromIpCh

	localIp := model.Int2Vip(ipPacket.Ipheader.Dst)
	remoteIp := model.Int2Vip(ipPacket.Ipheader.Src)

	tcpPacket := ConvertToTcpPacket(ipPacket.Payload)
	localPort := tcpPacket.Tcpheader.Destination
	remotePort := tcpPacket.Tcpheader.Source

	//fmt.Println("receive ippacket in v_accept(), localIp: %s, localport : %d, remoteIp: %s , remoteport: %d", localIp, localPort, remoteIp, remotePort)
	//create a new socket
	socketfd := manager.V_socket()

	socket, _ := manager.GetSocketByFd(socketfd)
	//send back ACK and SYN
	//fmt.Printf("Accept state is %s\n", socket.StateMachine.CurrentState())
	socket.StateMachine.Transit(TCP_PASSIVE_OPEN)
	ctrl, _ := socket.StateMachine.GetResponse(TCP_RECV_SYN)
	//fmt.Printf("Accept returns Ctrl flags as %b\n", ctrl.GetCtrlFlags())
	_, err := socket.SendCtrl(ctrl.GetCtrlFlags(), int(rand.Int31()), tcpPacket.Tcpheader.SeqNum+1, localIp, localPort, remoteIp, remotePort)
	if err != nil {
		return -1, errors.New("v_accept() error: sendctrl() went wrong")
	}
	*addr = remoteIp
	*port = remotePort
	socket.StateMachine.Transit(TCP_RECV_SYN)
	fmt.Printf("v_accept() on socket %d returned %d\n", listenfd, socketfd)
	return socketfd, nil
}

func (manager *SocketManager) V_read(socketFd int, nbyte int) ([]byte, int) {
	socket, _ := manager.GetSocketByFd(socketFd)
	buff, buffSize := socket.ReadFromBuffer(nbyte)

	return buff, buffSize
}

func (manager *SocketManager) V_write(socketfd int, buf []byte, nbyte int) int {
	//put data into buffer
	//if full, blocking
	socket, _ := manager.GetSocketByFd(socketfd)
	size := socket.AddToBuffer(buf, nbyte)
	fmt.Printf("v_write() on %d bytes returned %d\n", nbyte, size)
	logging.Logger.Printf("[SocketManager] V_write socketfd:%d write size :%d", socket.Fd, size)
	return size
}

func (manager *SocketManager) V_shutdown(socket int, closeType int) int {
	return 0
}

func (manager *SocketManager) V_close(socket int) int {
	return 0
}
