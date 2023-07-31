package cmd

import (
	"bytes"
	"fmt"
	"io"
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
	// Test integration output
	if stdout.String() != expectedOut {
		t.Errorf("Expected output %q, got %q\n", expectedOut, stdout.String())
	}
}
