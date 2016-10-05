package main

import "fmt"
import "os"
import "bufio"
import "strings"
import "text/tabwriter"

import (
	"./model"
)

//import "./network"

func read_lnx(filename string) map[model.VirtualIp]model.NodeInterface {
	interfaces := make(map[model.VirtualIp]model.NodeInterface)
	if file, err := os.Open(os.Args[1]); err == nil {

		// make sure it gets closed
		defer file.Close()

		// create a new scanner and read the file line by line
		scanner := bufio.NewScanner(file)
		id_counter := 0
		for scanner.Scan() {
			// fmt.Println(scanner.Text())

			line := scanner.Text()
			tokens := strings.Split(line, " ")

			if len(tokens) == 1 {
				// self address:port pair, ignoring for now
			} else {
				descriptor := tokens[0]
				src := model.VirtualIp{Ip: tokens[1]}
				dest := model.VirtualIp{Ip: tokens[2]}

				node_interface := model.NodeInterface{Id: id_counter, Src: src, Dest: dest, Enabled: true, Descriptor: descriptor}
				interfaces[dest] = node_interface

				id_counter += 1
			}

			// fmt.Println(tokens[0])
		}

	} else {
		fmt.Println("fatal!")
	}

	return interfaces
}

func set_routingtable(interfaces map[model.VirtualIp]model.NodeInterface) model.RoutingTable {
	table := model.MakeRoutingTable()
	for _, v := range interfaces {
		entry := model.MakeRoutingEntry(v.Src, v.Src, v.Src, 0)
		table.PutEntry(entry)
	}
	return table
}

func print_interfaces(interfaces map[model.VirtualIp]model.NodeInterface) {
	w := new(tabwriter.Writer)
	// Format in tab-separated columns with a tab stop of 8.
	w.Init(os.Stdout, 0, 8, 0, '\t', 0)
	fmt.Fprintf(w, "id\tdst\tsrc\tenabled\n")
	for _, v := range interfaces {
		fmt.Fprintf(w, "%d\t%s\t%s\t%t\n", v.Id, v.Dest.Ip, v.Src.Ip, v.Enabled)
	}
	w.Flush()
}

func print_routingtable(table model.RoutingTable) {
	w := new(tabwriter.Writer)
	// Format in tab-separated columns with a tab stop of 8.
	w.Init(os.Stdout, 0, 8, 0, '\t', 0)
	fmt.Fprintf(w, " \t\tdst\tsrc\tcost\n")

	table_map := table.RoutingEntries
	for _, v := range table_map {
		fmt.Fprintf(w, " \t\t%s\t%s\t%d\n", v.ExitIp.Ip, v.ExitIp.Ip, v.Cost)
	}
	w.Flush()
}

func main() {
	IpPacket := model.MakeIpPacket([]byte("hello"), 0, model.VirtualIp{"192.168.0.5"}, model.VirtualIp{"192.168.0.6"})

	buffer := IpPacket.ConvertToBuffer()
	rPacket := model.ConvertToIpPacket(buffer)
	fmt.Println(rPacket.IpPacketString())

	// read link file
	link_file := os.Args[1]
	//fmt.Println(link_file)

	interfaces := read_lnx(link_file)
	print_interfaces(interfaces)

	table := set_routingtable(interfaces)

	print_routingtable(table)
}
