package transport

import (
	"fmt"
	"model"
	"os"
)

type FileSender struct {
	socketmanager *SocketManager
	socket        *TcpSocket
	file          *os.File
}

func MakeFileReceiver(sm *SocketManager, ip model.VirtualIp, port int, filename string) FileReceiver {
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

	fr.CloseReceiver()
}
