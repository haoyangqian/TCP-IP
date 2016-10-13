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
	"net"
	"sort"
	"strconv"
	"sync"
	//"time"
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
				untranslatedService := tokens[0]
				hostname := strings.Split(untranslatedService, ":")[0]
				port := strings.Split(untranslatedService, ":")[1]
				remoteAddr, _ := net.LookupIP(hostname)

				service = remoteAddr[0].String() + ":" + port

				fmt.Printf("servicename: %s\n", service)

			} else {

				descriptor := tokens[0]
				src := model.VirtualIp{Ip: tokens[1]}
				dest := model.VirtualIp{Ip: tokens[2]}

				hostname := strings.Split(descriptor, ":")[0]
				port := strings.Split(descriptor, ":")[1]
				remoteAddr, _ := net.LookupIP(hostname)

				fullAddr := remoteAddr[0].String() + ":" + port

				node_interface := model.NodeInterface{Id: id_counter, Src: src, Dest: dest, Enabled: true, Descriptor: fullAddr}
				interfaces[dest] = &node_interface
				// node_interface2 := model.NodeInterface{Id: id_counter, Src: src, Dest: src, Enabled: true, Descriptor: service}
				// interfaces[src] = &node_interface2
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
		entry := model.MakeRoutingEntry(v.Src, v.Src, v.Src, 0, true)
		table.PutEntry(&entry)
		table.PutNeighbor(v.Dest, v.Src)
		//fmt.Println("put neighbor:", v.Dest)

	}
	return table
}

func PrintInterfaces(table model.NodeInterfaceTable) {
	interfaces := table.GetAllInterfaces()
	w := new(tabwriter.Writer)
	// Format in tab-separated columns with a tab stop of 8.
	w.Init(os.Stdout, 0, 8, 0, '\t', 0)
	fmt.Fprintf(w, "id\tdst\tsrc\tenabled\n")
	for _, v := range interfaces {

		fmt.Fprintf(w, "%d\t%s\t%s\t%t\n", v.Id, v.Dest.Ip, v.Src.Ip, v.Enabled)
	}
	w.Flush()
}

func PrintInterfacesall(table model.NodeInterfaceTable) {
	interfaces := table.GetAllInterfaces()
	w := new(tabwriter.Writer)
	// Format in tab-separated columns with a tab stop of 8.
	w.Init(os.Stdout, 0, 8, 0, '\t', 0)
	w.Init(os.Stdout, 0, 8, 0, '\t', 0)
	fmt.Fprintf(w, "id\tdst\tsrc\tenabled\taddress\n")
	for _, v := range interfaces {
		fmt.Fprintf(w, "%d\t%s\t%s\t%t\t%s\n", v.Id, v.Dest.Ip, v.Src.Ip, v.Enabled, v.Descriptor)
	}
	w.Flush()
}

func PrintRoutingtable(table model.RoutingTable) {
	w := new(tabwriter.Writer)
	table_map := table.RoutingEntries
	keys := make([]string, 0)
	for k, _ := range table_map {
		keys = append(keys, k.Ip)
	}
	sort.Strings(keys)
	fmt.Println(keys)
	//fmt.Println(interfaces)
	// Format in tab-separated columns with a tab stop of 8.
	w.Init(os.Stdout, 0, 8, 0, '\t', 0)
	fmt.Fprintf(w, " \t\tdst\tsrc\tcost\n")

	for _, k := range keys {
		v := table_map[model.VirtualIp{k}]
		fmt.Fprintf(w, " \t\t%s\t%s\t%d\n", v.Dest.Ip, v.ExitIp.Ip, v.Cost)
	}

	w.Flush()
}

func PrintNeighbors(table model.RoutingTable) {
	w := new(tabwriter.Writer)
	// Format in tab-separated columns with a tab stop of 8.
	w.Init(os.Stdout, 0, 8, 0, '\t', 0)
	fmt.Fprintf(w, " \t\tNeighborVip\t\tLocalVip\n")

	neighbor_map := table.GetAllNeighbors()
	fmt.Fprintf(w, "len of neighbor is %d\n", len(neighbor_map))
	for _, neighbor := range neighbor_map {
		fmt.Fprintf(w, " \t\t%s\n", neighbor.Ip)
	}

	w.Flush()
}

func PrintRoutingtableall(table model.RoutingTable) {
	w := new(tabwriter.Writer)
	table_map := table.RoutingEntries
	keys := make([]string, 0)
	for k, _ := range table_map {
		keys = append(keys, k.Ip)
	}
	sort.Strings(keys)
	//fmt.Println(keys)
	// Format in tab-separated columns with a tab stop of 8.
	fmt.Println("table len:", len(table.RoutingEntries))
	//fmt.Println("neighbor len:", len(table.Neighbors))
	w.Init(os.Stdout, 0, 8, 0, '\t', 0)
	fmt.Fprintf(w, " \t\tdst\texit\tnexthop\tcost\texpired\tshould_expire\tshould_gc\n")
	//fmt.Printf("time now is %d\n", time.Now().Unix())

	for _, k := range keys {
		v := table_map[model.VirtualIp{k}]
		fmt.Fprintf(w, " \t\t%s\t%s\t%s\t%d\t%t\t%t\t%t\n", v.Dest.Ip, v.ExitIp.Ip, v.NextHop.Ip, v.Cost, v.Expired(), v.ShouldExpire(), v.ShouldGC())
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
	interfaceTable := factory.NodeInterfaceTable()

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
			id, _ := strconv.Atoi(tokens[1])

			if interfaceTable.HasId(id) {
				interfaceTable.Up(id)
				uppedInterface, _ := interfaceTable.GetInterfaceById(id)
				newRoute := model.MakeRoutingEntry(uppedInterface.Src, uppedInterface.Src, uppedInterface.Src, 0, true)
				newRoute.SetIsUpdated(true)
				table.PutEntry(&newRoute)
				table.PutNeighbor(uppedInterface.Dest, uppedInterface.Src)
			}

		case "down":
			id, _ := strconv.Atoi(tokens[1])

			if interfaceTable.HasId(id) {
				interfaceTable.Down(id)
				downedInterface, _ := interfaceTable.GetInterfaceById(id)
				table.ExpireRoutesByExitIp(downedInterface.Src)
				table.DeleteNeighbor(downedInterface.Dest)
			}

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
			PrintInterfaces(interfaceTable)
		case "di":
			PrintInterfacesall(interfaceTable)
		case "routes":
			PrintRoutingtable(table)
		case "dr":
			PrintRoutingtableall(table)
		case "dn":
			PrintNeighbors(table)
		case "help":
			PrintHelp()
		default:
			PrintHelp()

		}
		fmt.Print("> ")
	}
	wg.Wait()
}
