package main

import "fmt"
import "os"
import "bufio"
import "strings"
import "text/tabwriter"

import "./model"

func main() {

	// read link file
	link_file := os.Args[1]
	fmt.Println(link_file)

	interfaces := make([]model.NodeInterface, 10)

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
				interfaces[id_counter] = node_interface

				id_counter += 1
			}

			// fmt.Println(tokens[0])
		}

	} else {
		fmt.Println("fatal!")
	}

	w := new(tabwriter.Writer)
	// Format in tab-separated columns with a tab stop of 8.
	w.Init(os.Stdout, 0, 8, 0, '\t', 0)
	fmt.Fprintf(w, "id\tdst\tsrc\tenabled\n")

	for k := 0; k < 2; k++ {
		i := interfaces[k]
		fmt.Fprintf(w, "%d\t%s\t\t%s\t\t%t\n", i.Id, i.Dest.Ip, i.Src.Ip, i.Enabled)

	}
	w.Flush()
}
