package transport

import (
	"errors"
	"fmt"
	"io"
	"logging"
	"model"
	"os"
)

type FileSender struct {
	socketmanager *SocketManager
	socket        *TcpSocket
	file          *os.File
	dstIp         model.VirtualIp
	dstPort       int
}

func MakeFileSender(sm *SocketManager, dstIp model.VirtualIp, dstPort int, filename string) (FileSender, error) {
	socketfd := sm.V_socket()
	sm.V_bind(socketfd, model.VirtualIp{}, -1)
	socket, _ := sm.GetSocketByFd(socketfd)
	file, err2 := os.OpenFile(filename, os.O_RDONLY, 0666)
	if err2 != nil {
		return FileSender{}, errors.New("[MakefileSender] Open file fail No such file")
	}
	return FileSender{sm, socket, file, dstIp, dstPort}, nil
}

func (fs *FileSender) CloseSender() {
	//close socket
	//close file
}

func (fs *FileSender) Send() {
	//connect to establish first
	socketFd := fs.socket.Fd
	_, err := fs.socketmanager.V_connect(socketFd, fs.dstIp, fs.dstPort)
	if err != nil {
		fmt.Printf("[MakefileSender] Connect fail")
	}
	buf := make([]byte, 512)
	for {
		if fs.socket.StateMachine.CurrentState() == TCP_ESTAB && fs.socket.sendWindow.bufferSize != 0 {
			n, err := fs.file.Read(buf)
			if err != nil && err != io.EOF {
				fmt.Printf("[FileSender] read from file fail!")
			}
			if n == 0 {
				break
			}
			size := fs.socketmanager.V_write(socketFd, buf, n)
			if size != 0 {
				logging.Logger.Printf("[FileSender] v_write success! size : %d\n", size)
			}
		}
	}
	fmt.Printf("sendfile on socket %d done\n", socketFd)
	fmt.Printf("ENDING SENDFILE\n")
	fs.CloseSender()
}
