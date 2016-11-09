package transport

import (
	"fmt"
	"logging"
	"model"
	"os"
)

type FileReceiver struct {
	socketmanager *SocketManager
	socket        *TcpSocket
	file          *os.File
}

func MakeFileReceiver(sm *SocketManager, port int, filename string) FileReceiver {
	socketfd := sm.V_socket()
	sm.V_bind(socketfd, model.VirtualIp{}, port)
	socket, _ := sm.GetSocketByFd(socketfd)

	file, _ := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	return FileReceiver{sm, socket, file}
}

func (fr *FileReceiver) CloseReceiver() {
	//close socket
	//close file
}
func (fr *FileReceiver) Recv() {
	var addr model.VirtualIp
	var port int
	listenfd := fr.socket.Fd
	fr.socketmanager.V_listen(listenfd)
	newFd, _ := fr.socketmanager.V_accept(listenfd, &addr, &port)
	listenSocket, _ := fr.socketmanager.GetSocketByFd(listenfd)
	newsocket, _ := fr.socketmanager.GetSocketByFd(newFd)
	newrunner, _ := fr.socketmanager.GetRunnerByFd(newFd)
	fr.socketmanager.SetSocketAddr(newFd, SocketAddr{listenSocket.Addr.LocalIp, listenSocket.Addr.LocalPort, addr, port})
	go newrunner.Run()
	for {
		if newsocket.StateMachine.CurrentState() == TCP_ESTAB {
			buff, size := fr.socketmanager.V_read(newFd, 1024)
			if size != 0 {
				fr.file.Write(buff)
				logging.Logger.Printf("[FileReceiver] read success! size : %d buffsize : %d\n", size, len(buff))
			} else if newsocket.StateMachine.CurrentState() == TCP_CLOSE_WAIT {
				break
			}
		}
	}
	fmt.Printf("recvfile on socket %d done", newsocket.Fd)
	fmt.Printf("ENDING RECVFILE\n")
	//fmt.Println("%d")
	fr.CloseReceiver()
}
