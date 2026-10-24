package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

type metricPoint struct {
	T           string  `json:"t"`
	CPU         float64 `json:"cpu"`
	Load1       float64 `json:"load1"`
	RAMUsedGb   float64 `json:"ramUsedGb"`
	RAMTotalGb  float64 `json:"ramTotalGb"`
	DiskUsedGb  float64 `json:"diskUsedGb"`
	DiskTotalGb float64 `json:"diskTotalGb"`
	NetRxKb     float64 `json:"netRxKb"`
	NetTxKb     float64 `json:"netTxKb"`
}

var metricsCmd = &cobra.Command{
	Use:   "metrics <id|name|host>",
	Short: "📈 Live metrics + CPU sparkline for a chamber",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, at, err := session(cmd)
		if err != nil {
			return err
		}
		id, err := c.ResolveServer(at, args[0])
		if err != nil {
			return err
		}
		var pts []metricPoint
		if err := c.Get("/api/servers/"+id+"/metrics?hours=24&limit=60", at, &pts); err != nil {
			return err
		}
		if len(pts) == 0 {
			fmt.Println("No metrics yet — run a probe first.")
			return nil
		}
		last := pts[len(pts)-1]
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintf(w, "Latest snapshot (chamber %s)\n\n", args[0])
		fmt.Fprintln(w, "CPU\tRAM\tDISK\tLOAD(1m)\tNET(Rx/Tx)")
		fmt.Fprintf(w, "%.1f%%\t%.1f%%\t%.1f%%\t%.2f\t%.0f/%.0f KB/s\n\n",
			last.CPU,
			pct(last.RAMUsedGb, last.RAMTotalGb),
			pct(last.DiskUsedGb, last.DiskTotalGb),
			last.Load1,
			last.NetRxKb, last.NetTxKb)
		w.Flush()
		fmt.Println("CPU sparkline (last 60 samples):")
		fmt.Println(spark(pts))
		return nil
	},
}

func pct(used, total float64) float64 {
	if total == 0 {
		return 0
	}
	return used / total * 100
}

func spark(pts []metricPoint) string {
	blocks := []string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
	min, max := pts[0].CPU, pts[0].CPU
	for _, p := range pts {
		if p.CPU < min {
			min = p.CPU
		}
		if p.CPU > max {
			max = p.CPU
		}
	}
	out := ""
	for _, p := range pts {
		if max == min {
			out += blocks[0]
			continue
		}
		idx := int((p.CPU - min) / (max - min) * 7)
		if idx > 7 {
			idx = 7
		}
		out += blocks[idx]
	}
	return out
}

func init() {
	rootCmd.AddCommand(metricsCmd)
}
