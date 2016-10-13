package model

import "fmt"

type NodeInterface struct {
	Id         int
	Src        VirtualIp
	Dest       VirtualIp
	Enabled    bool
	Descriptor string
}

func (i *NodeInterface) Down() {
	fmt.Printf("interface #%d Disabled\n", i.Id)
	i.Enabled = false
}

func (i *NodeInterface) Up() {
	fmt.Printf("interface #%d Enabled\n", i.Id)
	i.Enabled = true
}
