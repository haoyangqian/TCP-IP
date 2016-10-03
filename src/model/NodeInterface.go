package model

type NodeInterface struct {
	Id         int
	Src        VirtualIp
	Dest       VirtualIp
	Enabled    bool
	Descriptor string
}
