package cmd

import "github.com/spf13/cobra"

func init() {
	u := &cobra.Command{Use: "usage", Short: "💳 Usage & billing (per castle)"}
	add(u,
		jsonCmd("list", "List usage records", "GET", "/api/usage", 0),
		jsonCmd("snapshot", "Snapshot this month's usage", "POST", "/api/usage/snapshot", 0),
	)
	rootCmd.AddCommand(u)
}
