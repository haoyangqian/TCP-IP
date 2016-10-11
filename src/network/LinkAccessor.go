package network

import (
	"../model"
	"../util"
	"errors"
	"fmt"
	"net"
	"strings"
)

/*
struct for LinkLayer
    InterfacesTable:   store the map information about dst virtualip to its corresponding interface
    LocalService   :   LinkLayer Local Address
    UdpSocket      :   UDP socket
*/
type LinkAccessor struct {
	NeighborVipToSelf  map[model.VirtualIp]*model.NodeInterface
	NeighborAddrToSelf map[string]*model.NodeInterface
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
	selfinterface := accessor.NeighborVipToSelf[nextHop]
	//check if interface is down
	if selfinterface.Enabled == false {
		fmt.Println("Sorry,cannot send because interface is down:%s", selfinterface.Src.Ip)
		return
	}
	buffer := packet.ConvertToBuffer()
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

	selfinterface := accessor.NeighborAddrToSelf[addr.String()]
	//fmt.Println(selfinterface.Src, selfinterface.Descriptor)
	if selfinterface.Enabled == false {
		fmt.Println("Sorry,cannot send because interface is down:%s", selfinterface.Src.Ip)
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
func NewLinkAccessor(table map[model.VirtualIp]*model.NodeInterface, service string) LinkAccessor {
	udpAddr, err := net.ResolveUDPAddr("udp", service)
	util.CheckError(err)
	udpSocket, err := net.ListenUDP("udp", udpAddr)
	util.CheckError(err)
	neighborAddrToSelf := make(map[string]*model.NodeInterface)
	for _, v := range table {
		hostname := strings.Split(v.Descriptor, ":")[0]
		port := strings.Split(v.Descriptor, ":")[1]
		remoteAddr, _ := net.LookupIP(hostname)
		neighborAddrToSelf[remoteAddr[0].String()+":"+port] = v
	}
	return LinkAccessor{table, neighborAddrToSelf, service, udpSocket}
}

/*
  function  : Close UDP socket,and upper layer will defer this close function
  parameter : NULL
  return    : NULL
*/
func (accessor *LinkAccessor) CloseConnection() {
	accessor.UdpSocket.Close()
}
