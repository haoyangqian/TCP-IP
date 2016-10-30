package transport

import (
	"errors"
	"fmt"
	"model"
	"os"
	"text/tabwriter"
)

type SocketManager struct {
	socketMapByFd   map[int]*TcpSocket
	socketMapByAddr map[SocketAddr]*TcpSocket
	interfacetable  map[model.VirtualIp]bool
	fdcount         int
	portcount       int
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
	for k, v := range manager.socketMapByFd {
		fmt.Fprintf(w, "%d\t%s\t%d\t%s\t\t%d\n", k, v.Addr.LocalIp.Ip, v.Addr.LocalPort, v.Addr.RemoteIp.Ip, v.Addr.RemotePort)
	}
	w.Flush()
}

func MakeSocketManager(interfaces map[model.VirtualIp]*model.NodeInterface) SocketManager {
	socketmapfd := make(map[int]*TcpSocket)
	socketmapaddr := make(map[SocketAddr]*TcpSocket)
	interfacetable := make(map[model.VirtualIp]bool)
	for k, _ := range interfaces {
		interfacetable[k] = true
	}
	return SocketManager{socketmapfd, socketmapaddr, interfacetable, 0, 1024}
}

func (manager *SocketManager) GetSocketByAddr(addr SocketAddr) (*TcpSocket, error) {
	if s, ok := manager.socketMapByAddr[addr]; ok {
		return s, nil
	} else {
		return s, errors.New("No sockets found!")
	}
}

func (manager *SocketManager) GetSocketByFd(fd int) (*TcpSocket, error) {
	if s, ok := manager.socketMapByFd[fd]; ok {
		return s, nil
	} else {
		return s, errors.New("No sockets found!")
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

func (manager *SocketManager) V_socket() int {
	sendch := make(chan SendTcpMessageRequest)
	recvch := make(chan model.IpPacket)
	//create a new socket
	manager.fdcount += 1
	socket := MakeSocket(manager.fdcount, sendch, recvch)
	manager.socketMapByFd[manager.fdcount] = &socket
	return manager.fdcount
}

func (manager *SocketManager) V_bind(socketfd int, addr model.VirtualIp, port int) (int, error) {
	//get the socket from map
	socket, ok := manager.socketMapByFd[socketfd]
	if !ok {
		return -1, errors.New("v_bind() error:Wrong socketfd")
	}
	//if addr is nil/not specified, bind to any available interface
	if addr.Ip == "" {
		vip, err := manager.GetAvailableInterface(port)
		addr = vip
		if err != nil {
			delete(manager.socketMapByFd, socketfd)
			return -1, errors.New("v_bind() error: Cannot assign requested address")
		}
	}

	//set socket addr
	//check if this addr is available
	socketaddr := SocketAddr{addr, port, model.VirtualIp{"0.0.0.0"}, 0}
	socket.SetAddr(socketaddr)
	manager.socketMapByAddr[socketaddr] = socket
	return 0, nil
}

func (manager *SocketManager) V_listen(socket int) int {
	return 0
}

func (manager *SocketManager) V_connect(socket int, addr model.VirtualIp, port int) int {
	return 0
}

func (manager *SocketManager) V_accept(socket int, addr model.VirtualIp) int {
	return 0
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
