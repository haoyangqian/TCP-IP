package model

type SendMessageRequest struct {
	message  []byte
	protocol int
	dest     VirtualIp
}

func MakeSendMessageRequest(message []byte, protocol int, dest VirtualIp) SendMessageRequest {
	return SendMessageRequest{message, protocol, dest}
}

func (r *SendMessageRequest) Message() []byte {
	return r.message
}

func (r *SendMessageRequest) Protocol() int {
	return r.protocol
}

func (r *SendMessageRequest) Dest() VirtualIp {
	return r.dest
}
