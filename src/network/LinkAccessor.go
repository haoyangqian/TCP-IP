package network

import (
	"../model"
	"../util"
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
	InterfacesTable map[model.VirtualIp]*model.NodeInterface
	LocalService    string
	UdpSocket       *net.UDPConn
}

/*
  function  :  send IpPacket to remote address
  parameter :  SendPacketRequest
  return    :  NULL
*/
func (accessor *LinkAccessor) Send(request model.SendPacketRequest) {
	packet := request.Packet()
	nextHop := request.Dest()

	//fmt.Println("link layer Send()")
	buffer := packet.ConvertToBuffer()
	// fmt.Println("nextHop:", nextHop)
	remoteService := accessor.InterfacesTable[nextHop].Descriptor
	//fmt.Println("remoteAddr:" + remoteService)
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
func (accessor *LinkAccessor) Receive() model.IpPacket {
	//fmt.Println("link Receive()")
	// TODO: change 1400 to MTU
	buf := make([]byte, 1400)
	_, _, err := accessor.UdpSocket.ReadFromUDP(buf)
	util.CheckError(err)
	//fmt.Println("Received ", string(buf[0:n]), " from ", addr)
	ipPacket := model.ConvertToIpPacket(buf)
	//fmt.Println("link Received, returning packet")

	return ipPacket
}

/*
  function  :  create and initialize an instance of LinkLayer
  parameter :  table, service
  return    :  LinkAccessor
*/
func NewLinkAccessor(table map[model.VirtualIp]*model.NodeInterface, service string) LinkAccessor {
	udpAddr, err := net.ResolveUDPAddr("udp", service)
	util.CheckError(err)
	//fmt.Println("Listen UDP!!!!!" + service)
	udpSocket, err := net.ListenUDP("udp", udpAddr)
	util.CheckError(err)
	//fmt.Println("HERE!!")
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
