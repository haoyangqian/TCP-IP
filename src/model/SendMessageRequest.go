package model

type SendMessageRequest struct {
	message  []byte
	protocol int
	src      VirtualIp
	dest     VirtualIp
}

func MakeSendMessageRequest(message []byte, protocol int, dest VirtualIp) SendMessageRequest {
	return SendMessageRequest{message, protocol, EMPTY_VIRTUAL_IP, dest}
}

func MakeSendMessageRequestWithSrc(message []byte, protocol int, src VirtualIp, dest VirtualIp) SendMessageRequest {
	return SendMessageRequest{message, protocol, src, dest}
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

func (r *SendMessageRequest) Src() VirtualIp {
	return r.src
}

func (r *SendMessageRequest) HasSrc() bool {
	return r.src != EMPTY_VIRTUAL_IP
}
