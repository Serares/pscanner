package cmd

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"testing"

	"github.com/Serares/pscanner/scan"
)

func setup(t *testing.T, hosts []string, initList bool) (string, func()) {
	// temp file
	tf, err := os.CreateTemp("", "pscan")
	if err != nil {
		t.Fatal(err)
	}

	tf.Close()

	if initList {
		hl := &scan.HostsList{}

		for _, h := range hosts {
			hl.Add(h)
		}

		if err := hl.Save(tf.Name()); err != nil {
			t.Fatal(err)
		}
	}

	return tf.Name(), func() {
		os.Remove(tf.Name())
	}
}

func TestActions(t *testing.T) {
	hosts := []string{
		"host1",
		"host2",
		"host3",
	}
	testCases := []struct {
		name           string
		args           []string
		expectedOut    string
		initList       bool
		actionFunction func(io.Writer, string, []string) error
	}{{
		name:           "AddAction",
		args:           hosts,
		expectedOut:    "Added host: host1\nAdded host: host2\nAdded host: host3\n",
		initList:       false,
		actionFunction: addAction,
	},
		{
			name:           "ListAction",
			expectedOut:    "host1\nhost2\nhost3\n",
			initList:       true,
			actionFunction: listAction,
		},
		{
			name:           "DeleteAction",
			args:           []string{"host1", "host2"},
			expectedOut:    "Deleted host: host1\nDeleted host: host2\n",
			initList:       true,
			actionFunction: deleteAction,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tf, cleanup := setup(t, hosts, tc.initList)
			defer cleanup()
			var stdout bytes.Buffer

			if err := tc.actionFunction(&stdout, tf, tc.args); err != nil {
				t.Fatalf("expected no error, got %q\n", err)
			}

			if stdout.String() != tc.expectedOut {
				t.Errorf("Expected output %q, got %q\n", tc.expectedOut, stdout.String())
			}
		})
	}
}

func TestIntegration(t *testing.T) {
	hosts := []string{
		"host1",
		"host2",
		"host3",
	}
	tf, cleanup := setup(t, hosts, false)
	defer cleanup()

	delHost := "host2"
	hostsEnd := []string{
		"host1",
		"host3",
	}
	var stdout bytes.Buffer

	// define expected output for actions
	expectedOut := ""
	for _, v := range hosts {
		expectedOut += fmt.Sprintf("Added host: %s\n", v)
	}

	expectedOut += strings.Join(hosts, "\n")
	expectedOut += fmt.Sprintln()
	expectedOut += fmt.Sprintf("Deleted host: %s\n", delHost)
	expectedOut += strings.Join(hostsEnd, "\n")
	expectedOut += fmt.Sprintln()
	for _, v := range hostsEnd {
		expectedOut += fmt.Sprintf("%s: Host not found\n", v)
		expectedOut += fmt.Sprintln()
	}
	// Add hosts to the list
	if err := addAction(&stdout, tf, hosts); err != nil {
		t.Fatalf("Expected no error, got %q\n", err)
	}
	// List hosts
	if err := listAction(&stdout, tf, nil); err != nil {
		t.Fatalf("Expected no error, got %q\n", err)
	}
	// Delete host2
	if err := deleteAction(&stdout, tf, []string{delHost}); err != nil {
		t.Fatalf("Expected no error, got %q\n", err)
	}
	// List hosts after delete
	if err := listAction(&stdout, tf, nil); err != nil {
		t.Fatalf("Expected no error, got %q\n", err)
	}

	if err := scanAction(&stdout, tf, nil); err != nil {
		t.Fatalf("expected no error, got %q\n", err)
	}
	// Test integration output
	if stdout.String() != expectedOut {
		t.Errorf("Expected output %q, got %q\n", expectedOut, stdout.String())
	}
}

func TestScanAction(t *testing.T) {
	hosts := []string{
		"localhost",
		"unknownhostoutthere",
	}
	tf, cleanup := setup(t, hosts, true)
	defer cleanup()

	ports := []string{}
	// Init ports, 1 open, 1 closed
	for i := 0; i < 2; i++ {
		ln, err := net.Listen("tcp", net.JoinHostPort("localhost", "0"))
		if err != nil {
			t.Fatal(err)
		}
		defer ln.Close()
		_, portStr, err := net.SplitHostPort(ln.Addr().String())
		if err != nil {
			t.Fatal(err)
		}
		ports = append(ports, portStr)
		if i == 1 {
			ln.Close()
		}
	}

	expectedOut := fmt.Sprintln("localhost:")
	expectedOut += fmt.Sprintf("\t%s: open\n", ports[0])
	expectedOut += fmt.Sprintf("\t%s: closed\n", ports[1])
	expectedOut += fmt.Sprintln()
	expectedOut += fmt.Sprintln("unknownhostoutthere: Host not found")
	expectedOut += fmt.Sprintln()
	// Define var to capture scan output
	var out bytes.Buffer
	// Execute scan and capture output
	if err := scanAction(&out, tf, ports); err != nil {
		t.Fatalf("Expected no error, got %q\n", err)
	}
	// Test scan output
	if out.String() != expectedOut {
		t.Errorf("Expected output %q, got %q\n", expectedOut, out.String())
	}
}
