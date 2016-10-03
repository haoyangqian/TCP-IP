package model

type LinkEntry struct {
	Dest_address string
	Local_ip     VirtualIp
	Remote_ip    VirtualIp
	Is_self      bool
}

// func NewLinkEntry(e LinkEntry)
