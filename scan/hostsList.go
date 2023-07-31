package scan

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"sort"
)

var (
	ErrExists    = errors.New("host already in the list")
	ErrNotExists = errors.New("host not in the list")
)

// a list o hosts to run port scan

type HostsList struct {
	Hosts []string
}

func (h *HostsList) search(host string) (bool, int) {
	sort.Strings(h.Hosts)

	i := sort.SearchStrings(h.Hosts, host)

	if i < len(h.Hosts) && h.Hosts[i] == host {
		return true, i
	}

	return false, -1
}

func (h *HostsList) Add(host string) error {
	if found, _ := h.search(host); found {
		return fmt.Errorf("%w: %s", ErrExists, host)
	}

	h.Hosts = append(h.Hosts, host)
	return nil
}

func (h *HostsList) Remove(host string) error {
	if found, i := h.search(host); found {
		h.Hosts = append(h.Hosts[:i], h.Hosts[i+1:]...)
		return nil
	}
	return fmt.Errorf("%w, %s", ErrNotExists, host)
}

func (h *HostsList) Load(hostsfile string) error {
	f, err := os.Open(hostsfile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}

	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		h.Hosts = append(h.Hosts, scanner.Text())
	}

	return nil
}

func (h *HostsList) Save(hostsFile string) error {
	output := ""

	for _, host := range h.Hosts {
		output += fmt.Sprintln(host)
	}

	return os.WriteFile(hostsFile, []byte(output), 0644)
}
