package scan

import (
	"fmt"
	"net"
	"time"
)

type PortState struct {
	Port int
	Open state
}

type state bool

type Results struct {
	Host       string
	NotFound   bool
	PortStates []PortState
}

// implement the Stringer interface
func (s state) String() string {
	if s {
		return "open"
	}

	return "closed"
}

func scanPort(host string, port int) PortState {
	p := PortState{
		Port: port,
	}

	address := net.JoinHostPort(host, fmt.Sprintf("%d", port))

	// do the network connection attempt
	scanConn, err := net.DialTimeout("tcp", address, 1*time.Second)
	if err != nil {
		// assume the port is closed
		return p
	}

	scanConn.Close()
	p.Open = true
	return p
}

func Run(hl *HostsList, ports []int) []Results {
	res := make([]Results, 0, len(hl.Hosts))

	for _, h := range hl.Hosts {
		r := Results{
			Host: h,
		}
		// do the host checkup and see if it exists
		if _, err := net.LookupHost(h); err != nil {
			r.NotFound = true
			res = append(res, r)
			continue
		}
		for _, p := range ports {
			r.PortStates = append(r.PortStates, scanPort(h, p))
		}

		res = append(res, r)
	}
	return res
}
