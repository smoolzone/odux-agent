package cmd

import "github.com/spf13/cobra"

func init() {
	ws := &cobra.Command{Use: "workspace", Short: "🏰 Manage castles (workspaces) and members"}
	add(ws,
		jsonCmd("list", "List castles you belong to", "GET", "/api/workspaces", 0),
		bodyCmd("create", "Create a castle (--data '{\"name\":\"acme\"}')", "POST", "/api/workspaces", 0),
		jsonCmd("members", "List members of the active castle", "GET", "/api/workspace/members", 0),
	)
	rootCmd.AddCommand(ws)
}
