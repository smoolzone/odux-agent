package cmd

import "github.com/spf13/cobra"

func init() {
	db := &cobra.Command{Use: "database", Short: "🗄️ Provision databases per workspace"}
	add(db,
		jsonCmd("list", "List databases", "GET", "/api/databases", 0),
		bodyCmd("create", "Create a database (--data '{\"name\":\"...\"}')", "POST", "/api/databases", 0),
		jsonCmd("credentials <id>", "Show connection credentials", "GET", "/api/databases/%s/credentials", 1),
		jsonCmd("delete <id>", "Drop a database", "DELETE", "/api/databases/%s", 1),
	)
	rootCmd.AddCommand(db)
}
