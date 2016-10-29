package transport

import (
	"errors"
	"model"
)

type SocketManager struct {
	socketMapByFd   map[int]*TcpSocket
	socketMapByAddr map[SocketAddr]*TcpSocket
}

type SocketAddr struct {
	LocalIp    model.VirtualIp
	LocalPort  int
	RemoteIp   model.VirtualIp
	RemotePort int
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

func (manager *SocketManager) V_socket() int {

	return 0
}

func (manager *SocketManager) V_bind(socket int, addr model.VirtualIp, port int) int {
	return 0
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
