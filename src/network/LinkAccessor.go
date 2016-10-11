package network

import (
	"../model"
	"../util"
	"errors"
	"fmt"
	"net"
)

/*
struct for LinkLayer
    InterfacesTable:   store the map information about dst virtualip to its corresponding interface
    LocalService   :   LinkLayer Local Address
    UdpSocket      :   UDP socket
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

	selfinterface, _ := accessor.NodeInterfaceTable.GetInterfaceByNextHopVip(nextHop)
	//check if interface is down
	if selfinterface.Enabled == false {
		// fmt.Println("Sorry,cannot send because interface is down:%s", selfinterface.Src.Ip)
		return
	}
	buffer := packet.ConvertToBuffer()
	//check if udp len is larger than MTU
	if len(buffer) > model.UDP_MTU {
		fmt.Println("udp buffer larger than MTU")
		return
	}
	// fmt.Println("nextHop:", nextHop)
	remoteService := selfinterface.Descriptor
	if remoteService == "" {
		fmt.Println("remoteService is empty")
		return
	}
	remoteAddr, err := net.ResolveUDPAddr("udp", remoteService)
	//fmt.Println("localAddr:" + accessor.LocalService)
	_, err = accessor.UdpSocket.WriteToUDP(buffer, remoteAddr)
	util.CheckError(err)
	//fmt.Println("link layer sent")
}

/*
  function  :  receive IpPacket from remote address
  parameter :  NULL
  return    :  IpPacket
*/
func (accessor *LinkAccessor) Receive() (model.IpPacket, error) {
	//fmt.Println("link Receive()")
	// TODO: change 1400 to MTU
	buf := make([]byte, 1400)
	_, addr, err := accessor.UdpSocket.ReadFromUDP(buf)
	util.CheckError(err)

	if !accessor.NodeInterfaceTable.HasNextHopAddr(addr.String()) {
		fmt.Printf("interface table does not have next hop addr %s\n", addr.String())
	}
	selfinterface, _ := accessor.NodeInterfaceTable.GetInterfaceByNextHopAddr(addr.String())
	//fmt.Println(selfinterface.Src, selfinterface.Descriptor)
	if selfinterface.Enabled == false {
		// fmt.Println("Sorry,cannot send because interface is down:%s", selfinterface.Src.Ip)
		return model.IpPacket{}, errors.New("interface down")
	}
	//fmt.Println("Received ", string(buf[0:n]), " from ", addr)
	ipPacket := model.ConvertToIpPacket(buf)
	//fmt.Println("link Received, returning packet")

	return ipPacket, nil
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
