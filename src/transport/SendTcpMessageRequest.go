package transport

import ()

type SendTcpMessageRequest struct {
	SocketFd int
	Payload  []byte
}
