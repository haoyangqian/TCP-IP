package network

import (
	"model"
	"util"
	"errors"
	"fmt"
	"net"
)

/*
struct for LinkLayer
    NodeInterfaceTable:   store the map information about dst virtualip to its corresponding interface
    LocalService      :   LinkLayer Local Address
    UdpSocket         :   UDP socket
*/
type LinkAccessor struct {
	NodeInterfaceTable model.NodeInterfaceTable
	LocalService       string
	UdpSocket          *net.UDPConn
}

/*
  function  :  send IpPacket to remote address
  parameter :  SendPacketRequest
  return    :  NULL
*/
func (accessor *LinkAccessor) Send(request model.SendPacketRequest) {
	packet := request.Packet()
	nextHop := request.NextHop()

	selfinterface, err := accessor.NodeInterfaceTable.GetInterfaceByNextHopVip(nextHop)
	if err != nil {
		fmt.Println(err)
		return
	}
	//Check if this interface is down
	if selfinterface.Enabled == false {
		return
	}
	buffer := packet.ConvertToBuffer()
	//Check if udp length is larger than MTU
	if len(buffer) > model.UDP_MTU {
		fmt.Println("udp buffer larger than MTU")
		fmt.Println(buffer)
		// return
	}
	//Check if remoteService is empty
	remoteService := selfinterface.Descriptor
	if remoteService == "" {
		fmt.Println("remoteService is empty")
		return
	}
	remoteAddr, err := net.ResolveUDPAddr("udp", remoteService)
	_, err = accessor.UdpSocket.WriteToUDP(buffer, remoteAddr)
	util.CheckError(err)
}

/*
  function  :  receive IpPacket from remote address
  parameter :  NULL
  return    :  IpPacket
*/
func (accessor *LinkAccessor) Receive() (model.IpPacket, model.VirtualIp, error) {
	//fmt.Println("link Receive()")
	buf := make([]byte, model.UDP_RECEIVE_MAX_BUFFER_SIZE)
	bytesRead, addr, err := accessor.UdpSocket.ReadFromUDP(buf)
	util.CheckError(err)

	//Check if no interface can handle this packet
	if !accessor.NodeInterfaceTable.HasNextHopAddr(addr.String()) {
		return model.IpPacket{}, model.VirtualIp{}, errors.New("No interface matches the sender's address")
	}
	//If exists,get the available interface
	selfInterfaces, _ := accessor.NodeInterfaceTable.GetInterfacesByNextHopAddr(addr.String())
	hasUsableInterface := false
	var usableInterface *model.NodeInterface

	for _, v := range selfInterfaces {
		if v.Enabled {
			hasUsableInterface = true
			usableInterface = v
		}
	}
	//Check if all interfaces are down
	if !hasUsableInterface {
		return model.IpPacket{}, model.VirtualIp{}, errors.New("No interface is up to receive datagram from sender")
	}
	ipPacket := model.ConvertToIpPacket(buf[0:bytesRead])
	receivedFrom := usableInterface.Dest

	return ipPacket, receivedFrom, nil
}

/*
  function  :  create and initialize an instance of LinkLayer
  parameter :  table, service
  return    :  LinkAccessor
*/
func NewLinkAccessor(table model.NodeInterfaceTable, service string) LinkAccessor {
	udpAddr, err := net.ResolveUDPAddr("udp", service)
	util.CheckError(err)
	udpSocket, err := net.ListenUDP("udp", udpAddr)
	util.CheckError(err)

	return LinkAccessor{table, service, udpSocket}
}

/*
  function  : Close UDP socket,and upper layer will defer this close function
  parameter : NULL
  return    : NULL
*/
func (accessor *LinkAccessor) CloseConnection() {
	accessor.UdpSocket.Close()
}
