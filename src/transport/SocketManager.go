package transport

import (
	"errors"
	"fmt"
	"model"
	"os"
	"text/tabwriter"
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
	return SocketManager{socketmapfd, socketmapaddr, interfacetable, fsmBuilder, -1, 1024, sendToIpCh}
}

func (manager *SocketManager) GetSocketByAddr(addr SocketAddr) (*SocketRunner, error) {
	for k, _ := range manager.socketMapByAddr {
		fmt.Printf("get socketbyaddr() : map key: %+v\n", k)
	}
	if r, ok := manager.socketMapByAddr[addr]; ok {
		return r, nil
	} else {
		return r, errors.New("No sockets found!")
	}
}

func (manager *SocketManager) GetSocketByFd(fd int) (*SocketRunner, error) {
	if r, ok := manager.socketMapByFd[fd]; ok {
		return r, nil
	} else {
		return r, errors.New("No sockets found!")
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

func (manager *SocketManager) V_socket() int {
	// create a new socket and socket runner

	//create a new socket
	fmt.Println("****")
	manager.PrintSockets()
	manager.fdcount += 1
	fmt.Println("fdcount:", manager.fdcount)
	stateMachine := manager.fsmBuilder.Build()
	socket := MakeSocket(manager.fdcount, stateMachine, manager.sendToIpCh)
	runner := MakeSocketRunner(&socket, manager, make(chan model.IpPacket))
	manager.socketMapByFd[manager.fdcount] = &runner
	manager.PrintSockets()
	return manager.fdcount
}

func (manager *SocketManager) V_bind(socketfd int, addr model.VirtualIp, port int) (int, error) {
	//get the socket from map
	runner, ok := manager.socketMapByFd[socketfd]
	if !ok {
		return -1, errors.New("v_bind() error:Wrong socketfd")
	}

	socket := runner.Socket
	//if addr is nil/not specified, bind to any available interface
	if addr.Ip == "" {
		//if port is -1, choose portcount;
		if port == -1 {
			port = manager.portcount
			for {
				vip, err := manager.GetAvailableInterface(port)
				//find available port
				if err != nil {
					manager.portcount += 1
					addr = vip
					break
				} else {
					port = manager.portcount
					manager.portcount += 1
				}
			}
			return -1, errors.New("v_bind() error:No available port")
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
	socket.SetAddr(socketaddr)
	fmt.Println("bind addr:", socketaddr)
	//	manager.socketMapByAddr[socketaddr] = socket
	return 0, nil
}

func (manager *SocketManager) V_listen(socketfd int) int {
	// get socket runner
	// transit socket state
	// starts runner

	runner, _ := manager.GetSocketByFd(socketfd)
	runner.Socket.StateMachine.Transit(TCP_PASSIVE_OPEN)
	return 0
}

func (manager *SocketManager) V_connect(socketfd int, addr model.VirtualIp, port int) (int, error) {
	runner, _ := manager.GetSocketByFd(socketfd)
	socket := runner.Socket
	ctrl, _ := socket.StateMachine.GetResponse(TCP_ACTIVE_OPEN)
	_, err := socket.SendCtrl(ctrl.GetCtrlFlags(), socket.Addr.LocalIp, socket.Addr.LocalPort, addr, port)
	if err != nil {
		return -1, errors.New("v_connect() error: sendctrl() went wrong")
	}
	//setaddr and change the key, update record in mapbyaddr
	manager.UpdateRomoteAddr(socket, addr, port)
	socket.StateMachine.Transit(TCP_ACTIVE_OPEN)

	return 0, nil
}

func (manager *SocketManager) V_accept(listenfd int, addr *model.VirtualIp, port *int) (int, error) {
	socketfd := manager.V_socket()
	listenRunner, _ := manager.GetSocketByFd(listenfd)
	listenSocket := listenRunner.Socket
	//manager.V_bind(socketfd, listensocket.Addr.LocalIp, listensocket.Addr.LocalPort)
	runner, _ := manager.GetSocketByFd(socketfd)
	socket := runner.Socket
	//send back ACK and SYN
	fmt.Printf("Accept state is %s\n", socket.StateMachine.CurrentState())

	socket.StateMachine.Transit(TCP_PASSIVE_OPEN)

	ctrl, _ := socket.StateMachine.GetResponse(TCP_RECV_SYN)
	fmt.Printf("Accept returns Ctrl flags as %b\n", ctrl.GetCtrlFlags())
	_, err := socket.SendCtrl(ctrl.GetCtrlFlags(), listenSocket.Addr.LocalIp, listenSocket.Addr.LocalPort, *addr, *port)
	if err != nil {
		return -1, errors.New("v_accept() error: sendctrl() went wrong")
	}
	//set laddr/lport/raddr/rport, insert record into mapbyaddr
	//	manager.InsertRomoteAddr(socket, listensocket.Addr.LocalIp, listensocket.Addr.LocalPort, addr, port)
	for k, _ := range manager.socketMapByAddr {
		fmt.Printf("v_accept() : map key: %+v\n", k)
	}
	socket.StateMachine.Transit(TCP_RECV_SYN)
	return socketfd, nil
}

func (manager *SocketManager) V_read(socket int, buf []byte, nbyte int) int {
	return 0
}

func (manager *SocketManager) V_write(socket int, buf []byte, nbyte int) int {
	return 0
}

func (manager *SocketManager) V_shutdown(socket int, closeType int) int {
	return 0
}

func (manager *SocketManager) V_close(socket int) int {
	return 0
}
