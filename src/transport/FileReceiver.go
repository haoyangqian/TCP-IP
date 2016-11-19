package transport

import (
	"fmt"
	"model"
	"os"
)

type FileReceiver struct {
	socketmanager   *SocketManager
	socket          *TcpSocket
	transportSocket *TcpSocket
	file            *os.File
}

func MakeFileReceiver(sm *SocketManager, port int, filename string) FileReceiver {
	socketfd := sm.V_socket()
	sm.V_bind(socketfd, model.VirtualIp{}, port)
	socket, _ := sm.GetSocketByFd(socketfd)

	file, _ := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	return FileReceiver{socketmanager: sm, socket: socket, file: file}
}

func (fr *FileReceiver) CloseReceiver() {
	fr.socketmanager.V_close(fr.socket.Fd)
	fr.socketmanager.V_close(fr.transportSocket.Fd)
	fr.file.Close()
}
func (fr *FileReceiver) Recv() {
	var addr model.VirtualIp
	var port int
	listenfd := fr.socket.Fd
	fr.socketmanager.V_listen(listenfd)
	newFd, _ := fr.socketmanager.V_accept(listenfd, &addr, &port)
	listenSocket, _ := fr.socketmanager.GetSocketByFd(listenfd)
	newSocket, _ := fr.socketmanager.GetSocketByFd(newFd)
	fr.transportSocket = newSocket
	newrunner, _ := fr.socketmanager.GetRunnerByFd(newFd)
	fr.socketmanager.SetSocketAddr(newFd, SocketAddr{listenSocket.Addr.LocalIp, listenSocket.Addr.LocalPort, addr, port})
	go newrunner.Run()
	for {
		if fr.transportSocket.StateMachine.CurrentState() == TCP_CLOSE_WAIT {
			// read all bytes left in the buffer, then terminate the reading loop
			for {
				buff, size := fr.socketmanager.V_read(newFd, 1024)
				if size == 0 {
					break
				}

				fr.file.Write(buff)
			}
			break
		} else {
			buff, size := fr.socketmanager.V_read(newFd, 1024)
			if size != 0 {
				fr.file.Write(buff)
			}
		}
	}

	fmt.Printf("recvfile on socket %d completed\n", fr.transportSocket.Fd)
	fmt.Printf("ENDING RECVFILE\n")
	//fmt.Println("%d")
	fr.CloseReceiver()
}
