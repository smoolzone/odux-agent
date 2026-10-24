package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var inspectCmd = &cobra.Command{
	Use:   "inspect <id|name|host>",
	Short: "🛡️ Deep-dive a single chamber (server)",
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
		s, err := c.ServerByToken(at, id)
		if err != nil {
			return err
		}
		fmt.Printf("%s %s (id %d)\n\n", emoji(shield), s.Name, s.ID)
		fmt.Printf("Host:   %s\n", s.Host)
		fmt.Printf("Role:   %s\n", s.Role)
		fmt.Printf("Status: %s %s\n", statusEmoji(s.Status), s.Status)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(inspectCmd)
}
