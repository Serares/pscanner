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

type portScanner func(host string, port int) PortState

type ScanCfg struct {
	Ports []string
	Tcp   bool
	Udp   bool
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

func Run(hl *HostsList, cfg *ScanCfg) []Results {
	res := make([]Results, 0, len(hl.Hosts))
	var scannerFunc portScanner
	if cfg.Tcp {
		scannerFunc = scanTcpPort
	}

	if cfg.Udp {
		scannerFunc = scanUdpPort
	}

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

		for _, p := range cfg.Ports {
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
				r.PortStates = append(r.PortStates, scannerFunc(h, intPort))
				continue
			}
			intervalPorts, err := processIntervalPorts(p)
			if err != nil {
				fmt.Println("interval is invalid: ", p, err)
				continue
			}
			for i := intervalPorts[0]; i <= intervalPorts[len(intervalPorts)-1]; i++ {
				r.PortStates = append(r.PortStates, scannerFunc(h, i))
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

// send a packet and check if you get an error back
// if no error gets back then the port is open
// if an error is sent back then the port is closed
func scanUdpPort(host string, port int) PortState {
	p := PortState{
		Open: false,
		Port: port,
	}
	address := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	adr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return p
	}

	con, err := net.DialUDP("udp", nil, adr)
	if err != nil {
		return p
	}
	defer con.Close()

	packet := []byte("c")
	_, err = con.Write(packet)
	if err != nil {
		return p
	}

	resp := make([]byte, 1024)
	con.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
	n, _, err := con.ReadFromUDP(resp)
	if err != nil {
		fmt.Println("Err: ", fmt.Sprintf("Err:, %v", err))
		return p
	}

	fmt.Printf("Response: %s\n", resp[:n])

	p.Open = true
	return p
}

func scanTcpPort(host string, port int) PortState {
	p := PortState{
		Port: port,
	}

	address := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	// do the network connection attempt
	scanConn, err := net.DialTimeout("tcp", address, time.Second*1)
	if err != nil {
		// assume the port is closed
		return p
	}

	scanConn.Close()
	p.Open = true
	return p
}
