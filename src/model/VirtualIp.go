package model

import (
	"net"
)

type VirtualIp struct {
	Ip string
}

func MakeVirtualIp(ip string) VirtualIp {
	return VirtualIp{ip}
}

func (vip *VirtualIp) Vip2Int() []byte {
	return net.ParseIP(vip.Ip).To4()
}

func Int2Vip(vip net.IP) VirtualIp {
	return VirtualIp{vip.String()}
}
