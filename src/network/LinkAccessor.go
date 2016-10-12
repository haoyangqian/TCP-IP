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

	//fmt.Println("nextHop:", nextHop)
	selfinterface, err := accessor.NodeInterfaceTable.GetInterfaceByNextHopVip(nextHop)
	//fmt.Println("interface:", selfinterface)
	if err != nil {
		fmt.Println(err)
		return
	}
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
func (accessor *LinkAccessor) Receive() (model.IpPacket, model.VirtualIp, error) {
	//fmt.Println("link Receive()")
	// TODO: change 1400 to MTU\
	buf := make([]byte, 1400)
	_, addr, err := accessor.UdpSocket.ReadFromUDP(buf)
	util.CheckError(err)

	if !accessor.NodeInterfaceTable.HasNextHopAddr(addr.String()) {
		fmt.Printf("interface table does not have next hop addr %s\n", addr.String())
		return model.IpPacket{}, model.VirtualIp{}, errors.New("No interface matches the sender's address")
	}
	selfInterfaces, _ := accessor.NodeInterfaceTable.GetInterfacesByNextHopAddr(addr.String())
	//fmt.Println(selfinterface.Src, selfinterface.Descriptor)
	hasUsableInterface := false
	var usableInterface *model.NodeInterface

	for _, v := range(selfInterfaces) {
		if v.Enabled {
			hasUsableInterface = true
			usableInterface = v
		}
	}

	if !hasUsableInterface {
		return model.IpPacket{}, model.VirtualIp{}, errors.New("No interface is up to receive datagram from sender")
	}

	//fmt.Println("Received ", string(buf[0:n]), " from ", addr)
	ipPacket := model.ConvertToIpPacket(buf)
	receivedFrom := usableInterface.Dest
	//fmt.Println("link Received, returning packet")

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
