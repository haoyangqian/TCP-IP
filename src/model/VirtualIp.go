package model

type VirtualIp struct {
	Ip string
}

func MakeVirtualIp(ip string) VirtualIp {
	return VirtualIp{ip}
}
