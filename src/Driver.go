package main

import "fmt"
import "os"
import "bufio"
import "strings"
import "text/tabwriter"

import (
	"./factory"
	"./model"
	//"./util"
	"strconv"
	"sync"
	"time"
)

//global variable
var service string

func ReadLnx(filename string) map[model.VirtualIp]*model.NodeInterface {
	interfaces := make(map[model.VirtualIp]*model.NodeInterface)
	if file, err := os.Open(os.Args[1]); err == nil {

		// make sure it gets closed
		defer file.Close()

		// create a new scanner and read the file line by line
		scanner := bufio.NewScanner(file)
		id_counter := 0
		for scanner.Scan() {
			line := scanner.Text()
			tokens := strings.Split(line, " ")

			if len(tokens) == 1 {
				service = tokens[0]
				fmt.Printf("servicename: %s\n", service)

			} else {
				descriptor := tokens[0]
				src := model.VirtualIp{Ip: tokens[1]}
				dest := model.VirtualIp{Ip: tokens[2]}

				node_interface := model.NodeInterface{Id: id_counter, Src: src, Dest: dest, Enabled: true, Descriptor: descriptor, ToSelf: false}
				interfaces[dest] = &node_interface
				node_interface2 := model.NodeInterface{Id: id_counter, Src: src, Dest: src, Enabled: true, Descriptor: service, ToSelf: true}
				interfaces[src] = &node_interface2
				id_counter += 1
			}
		}

	} else {
		fmt.Println("fatal!")
	}
	return interfaces
}

func SetRoutingtable(interfaces map[model.VirtualIp]*model.NodeInterface) model.RoutingTable {
	table := model.MakeRoutingTable()
	for _, v := range interfaces {
		entry := model.MakeRoutingEntry(v.Src, v.Src, v.Src, 0)
		table.PutEntry(&entry)

		if v.ToSelf == false {
			table.PutNeighbor(v.Dest, v.Src)
			//fmt.Println("put neighbor:", v.Dest)
		}

	}
	return table
}

func PrintInterfaces(interfaces map[model.VirtualIp]*model.NodeInterface) {
	w := new(tabwriter.Writer)
	// Format in tab-separated columns with a tab stop of 8.
	w.Init(os.Stdout, 0, 8, 0, '\t', 0)
	fmt.Fprintf(w, "id\tdst\tsrc\tenabled\n")
	for _, v := range interfaces {
		if v.ToSelf == false {
			fmt.Fprintf(w, "%d\t%s\t%s\t%t\n", v.Id, v.Dest.Ip, v.Src.Ip, v.Enabled)
		}
	}
	w.Flush()
}

func PrintInterfacesall(interfaces map[model.VirtualIp]*model.NodeInterface) {
	w := new(tabwriter.Writer)
	// Format in tab-separated columns with a tab stop of 8.
	w.Init(os.Stdout, 0, 8, 0, '\t', 0)
	fmt.Fprintf(w, "id\tdst\tsrc\tenabled\ttoself\n")
	for _, v := range interfaces {
		fmt.Fprintf(w, "%d\t%s\t%s\t%t\t%t\n", v.Id, v.Dest.Ip, v.Src.Ip, v.Enabled, v.ToSelf)
	}
	w.Flush()
}

func PrintRoutingtable(table model.RoutingTable) {
	w := new(tabwriter.Writer)
	// Format in tab-separated columns with a tab stop of 8.
	w.Init(os.Stdout, 0, 8, 0, '\t', 0)
	fmt.Fprintf(w, " \t\tdst\tsrc\tcost\n")

	table_map := table.RoutingEntries
	for _, v := range table_map {
		fmt.Fprintf(w, " \t\t%s\t%s\t%d\n", v.NextHop.Ip, v.ExitIp.Ip, v.Cost)
	}
	w.Flush()
}

func PrintRoutingtableall(table model.RoutingTable) {
	w := new(tabwriter.Writer)
	// Format in tab-separated columns with a tab stop of 8.
	w.Init(os.Stdout, 0, 8, 0, '\t', 0)
	fmt.Fprintf(w, " \t\tdst\texit\tnexthop\tcost\texpired\tshould_expire\tshould_gc\n")
	fmt.Printf("time now is %d\n", time.Now().Unix())

	table_map := table.RoutingEntries
	for _, v := range table_map {
		fmt.Fprintf(w, " \t\t%s\t%s\t%s\t%d\t%t\t%t\t%t\n", v.Dest.Ip, v.ExitIp.Ip, v.NextHop, v.Cost, v.Expired(), v.ShouldExpire(), v.ShouldGC())
	}
	w.Flush()
}

func PrintHelp() {
	println("Commands:")
	println("up <id>                         - enable interface with id")
	println("down <id>                       - disable interface with id")
	println("send <dst_ip> <prot> <payload>  - send ip packet to <dst_ip> using prot <prot>")
	println("interfaces                      - list interfaces")
	println("routes                          - list routing table rows")
	println("help                            - show this help")
}

func main() {
	//IpPacket := model.MakeIpPacket([]byte("hello"), 0, model.VirtualIp{"192.168.0.6"}, model.VirtualIp{"192.168.0.5"})

	//buffer := IpPacket.ConvertToBuffer()
	//rPacket := model.ConvertToIpPacket(buffer)
	//fmt.Println(rPacket.IpPacketString())

	link_file := os.Args[1]
	//fmt.Println(link_file)

	interfaces := ReadLnx(link_file)

	table := SetRoutingtable(interfaces)

	// ripinfo, err := model.RoutingTable2RipInfo(table, 2)
	// util.CheckError(err)
	// fmt.Println(ripinfo.String())
	// b, err := ripinfo.Marshal()
	// util.CheckError(err)
	// returnrip, err := model.UnmarshalForInfo(b)
	// util.CheckError(err)
	// fmt.Println(returnrip.String())

	factory := factory.InitializeResourceFactory(table, interfaces, service)

	linkReceiveRunner := factory.LinkReceiveRunner()
	linkSendRunner := factory.LinkSendRunner()
	networkRunner := factory.NetworkRunner()
	ripRunner := factory.RipRunner()

	var wg sync.WaitGroup
	wg.Add(4)
	go networkRunner.Run()
	go linkSendRunner.Run()
	go linkReceiveRunner.Run()
	go ripRunner.Run()

	// linklayer := network.NewLinkAccessor(interfaces, service)
	// defer linklayer.CloseConnection()
	// fmt.Println("im sending!")
	// linklayer.Send(IpPacket)
	// for {
	// 	// wait for UDP client to connect
	// 	ReceivePacket := linklayer.Receive()
	// 	fmt.Println(ReceivePacket.IpPacketString())
	// }

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("> ")
	for scanner.Scan() {
		line := scanner.Text()
		tokens := strings.Split(line, " ")
		command := strings.ToLower(tokens[0])
		switch command {
		case "up":
		case "down":
		case "send":
			{
				if len(tokens) != 4 {
					fmt.Println("invalid args: send <dst_ip> <prot> <payload> ")
					break
				}
				dstIp := tokens[1]
				prot, _ := strconv.Atoi(tokens[2])
				payload := tokens[3]

				request := model.MakeSendMessageRequest([]byte(payload), prot, model.VirtualIp{dstIp})
				factory.MessageChannel() <- request
			}
		case "interfaces":
			PrintInterfaces(interfaces)
		case "di":
			PrintInterfacesall(interfaces)
		case "routes":
			PrintRoutingtable(table)
		case "dr":
			PrintRoutingtableall(table)
		case "help":
			PrintHelp()
		default:
			PrintHelp()

		}
		fmt.Print("> ")
	}
	wg.Wait()
}
