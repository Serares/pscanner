/*
Copyright © 2023 rares

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/Serares/pscanner/scan"
	"github.com/spf13/cobra"
)

// scanCmd represents the scan command
var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Run port scanning on existing hosts",
	RunE: func(cmd *cobra.Command, args []string) error {
		hostsFile, err := cmd.Flags().GetString("hosts-file")
		if err != nil {
			return err
		}

		ports, err := cmd.Flags().GetStringSlice("ports")
		if err != nil {
			return err
		}
		isTcp, err := cmd.Flags().GetBool("tcp")
		if err != nil {
			return err
		}
		isUdp, err := cmd.Flags().GetBool("udp")
		if err != nil {
			return err
		}

		if !isTcp && !isUdp {
			return fmt.Errorf("please specify a flag for network scan")
		}

		cfg := &scan.ScanCfg{
			Tcp:   isTcp,
			Udp:   isUdp,
			Ports: ports,
		}

		return scanAction(os.Stdout, hostsFile, cfg)
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)

	scanCmd.Flags().StringSliceP("ports", "p", []string{"22-443"}, "ports to scan")
	scanCmd.Flags().BoolP("tcp", "T", false, "use a TCP scan")
	scanCmd.Flags().BoolP("udp", "U", false, "use a UDP scan")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// scanCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
}

func scanAction(w io.Writer, hostsFile string, cfg *scan.ScanCfg) error {
	hl := &scan.HostsList{}

	if err := hl.Load(hostsFile); err != nil {
		return err
	}

	results := scan.Run(hl, cfg)

	return printResults(w, results, cfg)
}

func printResults(out io.Writer, results []scan.Results, cfg *scan.ScanCfg) error {
	message := ""
	if cfg.Tcp {
		message += fmt.Sprint("TCP scan: \n")
	}
	if cfg.Udp {
		message += fmt.Sprint("UDP scan: \n")
	}

	for _, r := range results {
		message += fmt.Sprintf("%s:", r.Host)

		if r.NotFound {
			message += fmt.Sprintf(" Host not found\n\n")
			continue
		}

		message += fmt.Sprintln()

		for _, p := range r.PortStates {
			message += fmt.Sprintf("\t%d: %s\n", p.Port, p.Open)
		}

		message += fmt.Sprintln()
	}

	_, err := fmt.Fprint(out, message)
	return err
}
