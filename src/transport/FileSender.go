package transport

import (
	"errors"
	"fmt"
	"io"
	//"logging"
	"model"
	"os"
	"time"
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
	res := fs.socketmanager.V_close(fs.socket.Fd)
	fmt.Printf("v_close() return %d\n", res)
	//close file
	fs.file.Close()
}

func (fs *FileSender) Send() {
	//connect to establish first
	socketFd := fs.socket.Fd
	_, err := fs.socketmanager.V_connect(socketFd, fs.dstIp, fs.dstPort)
	if err != nil {
		fmt.Printf("[MakefileSender] Connect fail")
	}

	starttime := time.Now().UnixNano()
	for {
		//passive_close !!!!!!
		//logging.Printf("[FileSender] : state %s\n buffer size :%d", fs.socket.StateMachine.CurrentState().Name, fs.socket.SendWindow.bufferSize)
		if fs.socket.StateMachine.CurrentState() == TCP_ESTAB && fs.socket.SendWindow.bufferSize != 0 {
			buf := make([]byte, 1024)
			n, err := fs.file.Read(buf)
			if err != nil && err != io.EOF {
				fmt.Printf("[FileSender] read from file fail!")
			}
			if n == 0 {
				break
			}
			size := fs.socketmanager.V_write(socketFd, buf, n)
			if size != 0 {
				//logging.Printf("[FileSender] v_write success! size : %d\n", size)
			}
		}
	}
	//send fin
	//fs.socket.SendCtrl(FIN, fs.socket., acknum, laddr, lport, raddr, rport)
	//	fs.socket.SendCtrl(FIN|ACK, seqnum, acknum, laddr, lport, raddr, rport)
	fmt.Printf("sendfile on socket %d done\n", socketFd)
	fmt.Printf("ENDING SENDFILE\n")
	endtime := time.Now().UnixNano()
	fmt.Printf("total time:%d ms, %d s\n", (endtime-starttime)/1000000, (endtime-starttime)/1000000000)
	//  little bug when close the
	time.Sleep(1000 * time.Millisecond)
	fs.CloseSender()
	//	for {
	//		if fs.socket.MaxAckNumRecved > fs.socket.lastSentSeq {
	//			fs.CloseSender()
	//			break
	//		}
	//	}

}
