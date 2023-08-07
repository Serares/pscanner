package scan

import (
	"fmt"
	"net"
	"strconv"
	"strings"
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

func Run(hl *HostsList, ports []string) []Results {
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
			if !checkIfInterval(p) {
				intPort, err := strconv.Atoi(p)
				if err != nil {
					fmt.Println("Error converting port:", p)
					continue
				}
				if !isPortValid(intPort) {
					fmt.Println("port is not valid: ", intPort)
					continue
				}
				r.PortStates = append(r.PortStates, scanPort(h, intPort))
				continue
			}
			intervalPorts, err := processIntervalPorts(p)
			if err != nil {
				fmt.Println("interval is invalid: ", p, err)
				continue
			}
			for i := intervalPorts[0]; i <= intervalPorts[len(intervalPorts)-1]; i++ {
				r.PortStates = append(r.PortStates, scanPort(h, i))
			}
		}

		res = append(res, r)
	}
	return res
}

func checkIfInterval(port string) bool {
	// check if it's an interval of ports
	return strings.Contains(port, "-")
}

func isPortValid(port int) bool {
	if port < 1 || port > 65535 {
		return false
	}

	return true
}

func processIntervalPorts(intervalPorts string) ([]int, error) {
	ports := strings.Split(intervalPorts, "-")

	intPorts := make([]int, 0, len(ports))
	for _, p := range ports {
		intPort, err := strconv.Atoi(p)
		if err != nil {
			return nil, fmt.Errorf("error converting port to int %s", p)
		}
		if !isPortValid(intPort) {
			return nil, fmt.Errorf("port is out of range %s", p)
		}
		intPorts = append(intPorts, intPort)
	}

	return intPorts, nil
}
