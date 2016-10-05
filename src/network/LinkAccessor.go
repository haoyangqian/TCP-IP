package network

import "fmt"
import "../model"

type LinkAccessor struct {
	Interfaces map[model.VirtualIp]model.NodeInterface
}

func (accessor *LinkAccessor) Send(packet model.IpPacket) {
	// to be implemented
	fmt.Println("calling #Send on LinkAccessor")

}

func (accessor *LinkAccessor) Receive() model.IpPacket {
	// to be implemented
	fmt.Println("calling #Receive on LinkAccessor")
	return model.IpPacket{}
}

func MakeLinkAccessor(interfaces map[model.VirtualIp]model.NodeInterface) LinkAccessor {
	return LinkAccessor{interfaces}
}
