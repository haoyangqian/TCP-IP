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
	InterfacesTable map[model.VirtualIp]model.NodeInterface
	LocalService    string
	UdpSocket       *net.UDPConn
}

/*
  function  :  send IpPacket to remote address
  parameter :  IpPacket
  return    :  NULL
*/
func (accessor *LinkAccessor) Send(packet model.IpPacket) {
	fmt.Println("link layer Send()")
	buffer := packet.ConvertToBuffer()
	dstVip := model.VirtualIp{packet.Ipheader.Dst.String()}
	remoteService := accessor.InterfacesTable[dstVip].Descriptor
	//fmt.Println("remoteAddr:" + remoteService)
	if remoteService == "" {
		fmt.Println("remoteService is empty")
		return
	}
	remoteAddr, err := net.ResolveUDPAddr("udp", remoteService)
	//fmt.Println("localAddr:" + accessor.LocalService)
	_, err = accessor.UdpSocket.WriteToUDP(buffer, remoteAddr)
	util.CheckError(err)
	fmt.Println("link layer sent")
}

/*
  function  :  receive IpPacket from remote address
  parameter :  NULL
  return    :  IpPacket
*/
func (accessor *LinkAccessor) Receive() model.IpPacket {
	fmt.Println("link Receive()")
	buf := make([]byte, 1400)
	n, addr, err := accessor.UdpSocket.ReadFromUDP(buf)
	util.CheckError(err)
	fmt.Println("Received ", string(buf[0:n]), " from ", addr)
	ipPacket := model.ConvertToIpPacket(buf)
	fmt.Println("link Received, returning packet")
	return ipPacket
}

/*
  function  :  create and initialize an instance of LinkLayer
  parameter :  table, service
  return    :  LinkAccessor
*/
func NewLinkAccessor(table map[model.VirtualIp]model.NodeInterface, service string) LinkAccessor {
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
