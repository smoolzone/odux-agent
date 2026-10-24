package cmd

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

var (
	opsProbePort int
	opsProbeHTTP bool
)

var opsCmd = &cobra.Command{
	Use:   "ops",
	Short: "⚔️ Run operational tasks",
}

var opsProbeCmd = &cobra.Command{
	Use:   "probe <host>",
	Short: "Probe a server (TCP reachability or HTTP health)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		host := args[0]
		if opsProbeHTTP {
			url := host
			if !hasScheme(url) {
				url = "https://" + url
			}
			start := time.Now()
			resp, err := http.Get(url)
			if err != nil {
				return fmt.Errorf("http probe failed: %w", err)
			}
			defer resp.Body.Close()
			fmt.Printf("%s %s -> HTTP %s (%dms)\n", emoji(check), host, resp.Status, time.Since(start).Milliseconds())
			return nil
		}
		addr := net.JoinHostPort(host, strconv.Itoa(opsProbePort))
		start := time.Now()
		conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
		if err != nil {
			return fmt.Errorf("%s:%d unreachable: %w", host, opsProbePort, err)
		}
		conn.Close()
		fmt.Printf("%s %s:%d -> open (%dms)\n", emoji(check), host, opsProbePort, time.Since(start).Milliseconds())
		return nil
	},
}

func hasScheme(s string) bool {
	return len(s) >= 8 && (s[:7] == "http://" || s[:8] == "https://")
}

func init() {
	rootCmd.AddCommand(opsCmd)
	opsCmd.AddCommand(opsProbeCmd)
	opsProbeCmd.Flags().IntVarP(&opsProbePort, "port", "p", 22, "port to probe (TCP mode)")
	opsProbeCmd.Flags().BoolVar(&opsProbeHTTP, "http", false, "perform an HTTP health check instead of TCP")
}
