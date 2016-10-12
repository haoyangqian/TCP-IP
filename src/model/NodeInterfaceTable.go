package model

import "errors"

type NodeInterfaceTable struct {
	NeighborVipToSelf  map[VirtualIp]*NodeInterface
	NeighborAddrToSelf map[string]*NodeInterface
}

func MakeNodeInterfaceTable(neighborVipToSelfTable map[VirtualIp]*NodeInterface) NodeInterfaceTable {

	neighborAddrToSelf := make(map[string]*NodeInterface)
	for _, v := range neighborVipToSelfTable {
		neighborAddrToSelf[v.Descriptor] = v
	}

	return NodeInterfaceTable{neighborVipToSelfTable, neighborAddrToSelf}
}

func (t *NodeInterfaceTable) HasNextHop(ip VirtualIp) bool {
	if _, ok := t.NeighborVipToSelf[ip]; ok {
		return true
	} else {
		return false
	}
}

func (t *NodeInterfaceTable) GetInterfaceByNextHopVip(ip VirtualIp) (*NodeInterface, error) {
	var i *NodeInterface
	if !t.HasNextHop(ip) {
		return i, errors.New("Invalid State: an interface not in the InterfaceTable was requested " + ip.Ip)
	}

	return t.NeighborVipToSelf[ip], nil
}

func (t *NodeInterfaceTable) HasExitIp(ip VirtualIp) bool {
	for _, v := range t.NeighborVipToSelf {
		if v.Src == ip {
			return true
		}
	}
	return false
}

func (t *NodeInterfaceTable) GetInterfaceByExitIp(ip VirtualIp) (*NodeInterface, error) {
	var i *NodeInterface

	for _, v := range t.NeighborVipToSelf {
		if v.Src == ip {
			return v, nil
		}
	}
	return i, errors.New("Invalid State: an interface not in the InterfaceTable was requested " + ip.Ip)

}

func (t *NodeInterfaceTable) HasNextHopAddr(addr string) bool {
	if _, ok := t.NeighborAddrToSelf[addr]; ok {
		return true
	} else {
		return false
	}
}

func (t *NodeInterfaceTable) GetInterfaceByNextHopAddr(addr string) (*NodeInterface, error) {
	var i *NodeInterface
	if !t.HasNextHopAddr(addr) {
		return i, errors.New("Invalid State: an interface not in the InterfaceTable was requested " + addr)
	}

	return t.NeighborAddrToSelf[addr], nil
}

func (t *NodeInterfaceTable) GetAllInterfaces() []*NodeInterface {
	interfaces := make([]*NodeInterface, 0)
	for _, v := range t.NeighborVipToSelf {
		interfaces = append(interfaces, v)
	}
	return interfaces
}

func (t *NodeInterfaceTable) HasId(id int) bool {
	for _, v := range t.NeighborVipToSelf {
		if v.Id == id {
			return true
		}
	}

	return false
}

func (t *NodeInterfaceTable) GetInterfaceById(id int) (*NodeInterface, error) {
	var result *NodeInterface

	for _, v := range t.NeighborVipToSelf {
		if v.Id == id {
			return v, nil
		}
	}

	return result, errors.New("No interface was found with the requested Id")
}

// func (t *NodeInterfaceTable) GetInterfaceByNextHopAddr(addr string) (*NodeInterface, error) {
// 	var i *NodeInterface
// 	if !t.HasNextHopAddress(addr) {
// 		return i, errors.New("Invalid State: an interface not in the InterfaceTable was requested " + addr)
// 	}

// 	return t.NeighborAddrToSelf[addr], nil
// }

func (t *NodeInterfaceTable) Down(id int) error {
	if id < 0 || id >= len(t.NeighborVipToSelf) {
		return errors.New("Interface id is out of range")
	}

	for _, v := range t.NeighborVipToSelf {
		if v.Id == id {
			v.Down()
		}
	}

	return nil
}

func (t *NodeInterfaceTable) Up(id int) error {
	if id < 0 || id >= len(t.NeighborVipToSelf) {
		return errors.New("Interface id is out of range")
	}

	for _, v := range t.NeighborVipToSelf {
		if v.Id == id {
			v.Up()
		}
	}

	return nil
}
