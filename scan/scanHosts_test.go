package scan_test

import (
	"net"
	"strconv"
	"testing"

	"github.com/Serares/pscanner/scan"
)

func TestStateString(t *testing.T) {
	ps := scan.PortState{}

	if ps.Open.String() != "closed" {
		t.Errorf("expected %q, got %q instead\n", "closed", ps.Open.String())
	}

	ps.Open = true

	if ps.Open.String() != "open" {
		t.Errorf("Expected %q, got %q instead\n", "open", ps.Open.String())
	}
}

func TestRunHostsFound(t *testing.T) {
	testCases := []struct {
		name        string
		expectState string
	}{
		{"OpenPort", "open"},
		{"ClosedPort", "closed"},
	}
	host := "localhost"
	hl := &scan.HostsList{}
	hl.Add(host)

	ports := []string{}
	// initialize ports 1 open, 1 closed
	for _, tc := range testCases {
		// searching on port '0' is going to return a random port
		// that is available on the machine
		ln, err := net.Listen("tcp", net.JoinHostPort(host, "0"))
		if err != nil {
			t.Fatal(err)
		}

		defer ln.Close()
		_, portStr, err := net.SplitHostPort(ln.Addr().String())
		if err != nil {
			t.Fatal(err)
		}

		ports = append(ports, portStr)
		if tc.name == "ClosedPort" {
			ln.Close()
		}
	}

	res := scan.Run(hl, ports)
	if len(res) != 1 {
		t.Fatalf("Expected 1 results, got %d instead\n", len(res))
	}
	if res[0].Host != host {
		t.Errorf("Expected host %q, got %q instead\n", host, res[0].Host)
	}
	if res[0].NotFound {
		t.Errorf("Expected host %q to be found\n", host)
	}

	if len(res[0].PortStates) != 2 {
		t.Fatalf("Expected 2 port states, got %d instead\n", len(res[0].PortStates))
	}
	for i, tc := range testCases {
		portToCheck, err := strconv.Atoi(ports[i])
		if err != nil {
			t.Fatalf("error converting port to int %v", ports[i])
		}
		if res[0].PortStates[i].Port != portToCheck {
			t.Errorf("Expected port %s, got %d instead\n", ports[0],
				res[0].PortStates[i].Port)
		}
		if res[0].PortStates[i].Open.String() != tc.expectState {
			t.Errorf("Expected port %s to be %s\n", ports[i], tc.expectState)
		}
	}
}

func TestRunHostNotFound(t *testing.T) {
	host := "389.389.389.389"
	hl := &scan.HostsList{}
	hl.Add(host)
	res := scan.Run(hl, []string{})
	// Verify results for HostNotFound test
	if len(res) != 1 {
		t.Fatalf("Expected 1 results, got %d instead\n", len(res))
	}
	if res[0].Host != host {
		t.Errorf("Expected host %q, got %q instead\n", host, res[0].Host)
	}
	if !res[0].NotFound {
		t.Errorf("Expected host %q NOT to be found\n", host)
	}
	if len(res[0].PortStates) != 0 {
		t.Fatalf("Expected 0 port states, got %d instead\n", len(res[0].PortStates))
	}
}
