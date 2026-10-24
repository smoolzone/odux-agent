package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "🏰 Live overview of your castle",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, at, err := session(cmd)
		if err != nil {
			return err
		}
		servers, err := c.Servers(at, "")
		if err != nil {
			return err
		}
		online := 0
		for _, s := range servers {
			if s.Status == "online" {
				online++
			}
		}
		offline := len(servers) - online
		noun := "chamber"
		if len(servers) != 1 {
			noun = "chambers"
		}
		fmt.Printf("%s %d %s · %s %d online · %s %d offline\n",
			emoji(castle), len(servers), noun, emoji(dotGreen), online, emoji(dotRed), offline)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
