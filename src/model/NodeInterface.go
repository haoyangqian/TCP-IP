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
	fmt.Printf("interface #%d %s -> %s is now Down\n", i.Id, i.Src, i.Dest)
	i.Enabled = false
}

func (i *NodeInterface) Up() {
	fmt.Printf("interface #%d %s -> %s is now Up\n", i.Id, i.Src, i.Dest)
	i.Enabled = true
}
