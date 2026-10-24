package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var disconnectCmd = &cobra.Command{
	Use:   "disconnect <id|name|host>",
	Short: "Detach a chamber from your castle",
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
		if err := c.DeleteServer(at, id); err != nil {
			return err
		}
		fmt.Printf("%s chamber %s detached\n", emoji(check), args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(disconnectCmd)
}
