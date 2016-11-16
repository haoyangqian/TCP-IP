package main

import "fmt"
import "os"
import "bufio"
import "strings"
import "text/tabwriter"

import (
	"factory"
	"logging"
	"model"
	"net"
	"sort"
	"strconv"
	"sync"
	"transport"
)

//global variable
var service string

func ReadLnx(filename string) map[model.VirtualIp]*model.NodeInterface {
	interfaces := make(map[model.VirtualIp]*model.NodeInterface)
	if file, err := os.Open(filename); err == nil {

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
				id_counter += 1
			}
		}

	} else {
		fmt.Println("[FATAL] Failed to read input to initialize the server")
		os.Exit(-1)
	}
	return interfaces
}

func SetRoutingtable(interfaces map[model.VirtualIp]*model.NodeInterface) model.RoutingTable {
	table := model.MakeRoutingTable()
	for _, v := range interfaces {
		entry := model.MakeRoutingEntry(v.Src, v.Src, v.Src, 0, true)
		table.PutEntry(&entry)
		table.PutNeighbor(v.Dest, v.Src)

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
	// Format in tab-separated columns with a tab stop of 8.
	fmt.Println("table len:", len(table.RoutingEntries))
	//fmt.Println("neighbor len:", len(table.Neighbors))
	w.Init(os.Stdout, 0, 8, 0, '\t', 0)
	fmt.Fprintf(w, " \t\tdst\texit\tnexthop\tcost\texpired\tshould_expire\tshould_gc\n")

	for _, k := range keys {
		v := table_map[model.VirtualIp{k}]
		fmt.Fprintf(w, " \t\t%s\t%s\t%s\t%d\t%t\t%t\t%t\n", v.Dest.Ip, v.ExitIp.Ip, v.NextHop.Ip, v.Cost, v.Expired(), v.ShouldExpire(), v.ShouldGC())
	}
	w.Flush()
}

func TcpSocketAccept(socketManager *transport.SocketManager, port int) {
	listenfd := socketManager.V_socket()
	_, err := socketManager.V_bind(listenfd, model.VirtualIp{}, port)
	if err != nil {
		fmt.Println(err)
		return
	}
	socketManager.V_listen(listenfd)
	for {
		var addr model.VirtualIp
		var port int
		//wait for connection...blocking
		newFd, _ := socketManager.V_accept(listenfd, &addr, &port)
		//fmt.Println("Accept returned from V_accept")
		listenSocket, _ := socketManager.GetSocketByFd(listenfd)
		newrunner, _ := socketManager.GetRunnerByFd(newFd)
		socketManager.SetSocketAddr(newFd, transport.SocketAddr{listenSocket.Addr.LocalIp, listenSocket.Addr.LocalPort, addr, port})
		go newrunner.Run()
	}
}

func PrintHelp() {
	fmt.Println("Commands:")
	fmt.Println("accept [port]                        - Spawn a socket, bind it to the given port,")
	fmt.Println("                                       and start accepting connections on that port.")
	fmt.Println("connect [ip] [port]                  - Attempt to connect to the given ip address,")
	fmt.Println("                                       in dot notation, on the given port.")
	fmt.Println("send [socket] [data]                 - Send a string on a socket.")
	fmt.Println("recv [socket] [numbytes] [y/n]       - Try to read data from a given socket. If")
	fmt.Println("                                       the last argument is y, then you should")
	fmt.Println("                                       block until numbytes is received, or the")
	fmt.Println("                                       connection closes. If n, then don.t block;")
	fmt.Println("                                       return whatever recv returns. Default is n.")
	fmt.Println("sendfile [filename] [ip] [port]      - Connect to the given ip and port, send the")
	fmt.Println("                                       entirety of the specified file, and close")
	fmt.Println("                                       the connection.")
	fmt.Println("recvfile [filename] [port]           - Listen for a connection on the given port.")
	fmt.Println("                                       Once established, write everything you can")
	fmt.Println("                                       read from the socket to the given file.")
	fmt.Println("                                       Once the other side closes the connection,")
	fmt.Println("                                       close the connection as well.")
	fmt.Println("shutdown [socket] [read/write/both]  - v_shutdown on the given socket.")
	fmt.Println("close [socket]                       - v_close on the given socket.")
	fmt.Println("up [id]                              - enable interface with id")
	fmt.Println("down [id]                            - disable interface with id")
	fmt.Println("interfaces                           - list interfaces")
	fmt.Println("routes                               - list routing table rows")
	fmt.Println("sockets                              - list sockets (fd, ip, port, state)")
	fmt.Println("window [socket]                      - lists window sizes for socket")
	fmt.Println("quit                                 - no cleanup, exit(0)")
	fmt.Println("help                                 - show this help")
}

func main() {
	link_file := os.Args[1]
	//link_file := "src.lnx"
	interfaces := ReadLnx(link_file)

	logging.Init(link_file, true)
	logging.Logger.Println("*********************************************************")

	table := SetRoutingtable(interfaces)

	factory := factory.InitializeResourceFactory(table, interfaces, service)

	interfaceTable := factory.NodeInterfaceTable()

	linkReceiveRunner := factory.LinkReceiveRunner()
	linkSendRunner := factory.LinkSendRunner()
	networkRunner := factory.NetworkRunner()
	ripRunner := factory.RipRunner()

	socketmanager := factory.SocketManager()

	var wg sync.WaitGroup
	wg.Add(4)
	go networkRunner.Run()
	go linkSendRunner.Run()
	go linkReceiveRunner.Run()
	go ripRunner.Run()
	//go
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("> ")
	//transport.Heaptest()
	//transport.Ringtest()
	for scanner.Scan() {
		line := scanner.Text()
		tokens := strings.Split(line, " ")
		command := strings.ToLower(tokens[0])
		switch command {
		case "vsocket":
			socketmanager.V_socket()
		case "vbind":
			listenfd, _ := strconv.Atoi(tokens[1])
			ip := tokens[2]
			port, _ := strconv.Atoi(tokens[3])
			port = port % 65535
			socketmanager.V_bind(listenfd, model.VirtualIp{ip}, port)
		case "vlisten":
			listenfd, _ := strconv.Atoi(tokens[1])
			socketmanager.V_listen(listenfd)
		case "vconnect":
			socketfd, _ := strconv.Atoi(tokens[1])
			ip := tokens[2]
			port, _ := strconv.Atoi(tokens[3])
			socketmanager.V_connect(socketfd, model.VirtualIp{ip}, port)
		case "vaccept":
			listenfd, _ := strconv.Atoi(tokens[1])
			var addr model.VirtualIp
			var port int
			//wait for connection...blocking
			newFd, _ := socketmanager.V_accept(listenfd, &addr, &port)
			//fmt.Println("Accept returned from V_accept")
			listenSocket, _ := socketmanager.GetSocketByFd(listenfd)
			newrunner, _ := socketmanager.GetRunnerByFd(newFd)
			socketmanager.SetSocketAddr(newFd, transport.SocketAddr{listenSocket.Addr.LocalIp, listenSocket.Addr.LocalPort, addr, port})
			go newrunner.Run()
		case "sockets":
			socketmanager.PrintSockets()
		case "accept":
			if len(tokens) != 2 || tokens[1] == "\n" {
				fmt.Println("syntax error(usage: accept [port])")
				break
			}
			//if port is larger than 65535
			port, _ := strconv.Atoi(tokens[1])
			port = port % 65535
			go TcpSocketAccept(socketmanager, port)

		case "connect":
			if len(tokens) != 3 || tokens[1] == "\n" {
				fmt.Println("syntax error(usage: connect [ip] [port])")
				break
			}
			dstIp := tokens[1]
			dstPort, _ := strconv.Atoi(tokens[2])
			socketfd := socketmanager.V_socket()
			_, err := socketmanager.V_bind(socketfd, model.VirtualIp{}, -1)
			if err != nil {
				fmt.Println(err)
				break
			}
			socketmanager.V_connect(socketfd, model.VirtualIp{dstIp}, dstPort)
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
				if len(tokens) != 3 {
					fmt.Println("invalid args: send <socket> <data>")
					break
				}
				socketFd, _ := strconv.Atoi(tokens[1])
				payload := tokens[2]
				socketmanager.V_write(socketFd, []byte(payload), len(payload))

			}
		case "sendfile":
			if len(tokens) != 4 {
				fmt.Println("invalid args: sendfile [filename] [ip] [port] ")
			}
			filename := tokens[1]
			dstIp := tokens[2]
			dstPort, _ := strconv.Atoi(tokens[3])
			dstPort = dstPort % 65535
			filesender, err := transport.MakeFileSender(socketmanager, model.VirtualIp{dstIp}, dstPort, filename)
			if err != nil {
				fmt.Println(err)
			} else {
				go filesender.Send()
			}
		case "recv":
			if len(tokens) != 4 {
				fmt.Println("invalid args: recv [socket] [numbytes] [y/n] ")
				break
			}

			socketFd, _ := strconv.Atoi(tokens[1])
			nBytes, _ := strconv.Atoi(tokens[2])
			shouldBlockStr := tokens[3]
			var shouldBlock bool
			if strings.ToLower(shouldBlockStr) == "y" {
				shouldBlock = true
			} else if strings.ToLower(shouldBlockStr) == "n" {
				shouldBlock = false
			} else {
				fmt.Println("invalid args: recv [socket] [numbytes] [y/n] ")
				break
			}

			bytesRead := 0
			result := make([]byte, 0)
			if shouldBlock {
				for {
					buff, size := socketmanager.V_read(socketFd, nBytes-bytesRead)
					bytesRead += size
					result = append(result, buff...)

					if bytesRead == nBytes {
						break
					}
				}
			} else {
				result, bytesRead = socketmanager.V_read(socketFd, nBytes)
			}

			fmt.Printf("%d bytes read from V_read\n", bytesRead)
			fmt.Println(string(result))
		case "recvfile":
			if len(tokens) != 3 {
				fmt.Println("invalid args: recvfile [filename] [port] ")
			}
			filename := tokens[1]
			port, _ := strconv.Atoi(tokens[2])
			port = port % 65535
			filereceiver := transport.MakeFileReceiver(socketmanager, port, filename)
			go filereceiver.Recv()
		case "window":
			if len(tokens) != 2 || tokens[1] == "\n" {
				fmt.Println("invalid args: window [sockets]")
			}
			socketfd, _ := strconv.Atoi(tokens[1])
			socket, _ := socketmanager.GetSocketByFd(socketfd)
			fmt.Printf("effective sending window:%d receiver advertise window:%d\n", socket.GetSendingWindow().EffectiveWindowSize(), socket.GetReceiverWindow().AdvertisedWindowSize())
		case "rr":
			filereceiver := transport.MakeFileReceiver(socketmanager, 9999, "rr")
			go filereceiver.Recv()
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
		case "tsm":
			transport.TestStateMachine()
		default:
			PrintHelp()

		}
		fmt.Print("> ")
	}
	wg.Wait()
}
