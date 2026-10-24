package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/smoolzone/odux-agent/internal/api"
)

func add(parent *cobra.Command, cmds ...*cobra.Command) {
	parent.AddCommand(cmds...)
}

func init() {
	// ---- chamber (servers) ----
	chamber := &cobra.Command{Use: "chamber", Short: "🛡️ Manage chambers (servers) in full"}
	chamber.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all chambers",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, at, err := session(cmd)
			if err != nil {
				return err
			}
			var servers []api.Server
			if err := c.Get("/api/servers", at, &servers); err != nil {
				return err
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tNAME\tHOST\tROLE\tSTATUS")
			for _, s := range servers {
				fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n", s.ID, s.Name, s.Host, s.Role, s.Status)
			}
			w.Flush()
			return nil
		},
	})
	add(chamber,
		jsonCmd("inspect <id>", "Inspect one chamber", "GET", "/api/servers/%s", 1),
		jsonCmd("overview", "Castle-wide snapshot (all chambers)", "GET", "/api/servers/overview", 0),
		jsonCmd("probe <id>", "Probe a chamber over SSH", "POST", "/api/servers/%s/probe", 1),
		jsonCmd("metrics <id>", "Fetch chamber metrics", "GET", "/api/servers/%s/metrics", 1),
		jsonCmd("logs <id>", "Search collected logs", "GET", "/api/servers/%s/logs", 1),
		jsonCmd("logs-collect <id>", "Collect logs over SSH", "POST", "/api/servers/%s/logs/collect", 1),
		jsonCmd("ssh-keys", "List SSH keys", "GET", "/api/servers/keys/list", 0),
	)
	rootCmd.AddCommand(chamber)

	// ---- dns ----
	dns := &cobra.Command{Use: "dns", Short: "🌐 DNS providers, zones, and records"}
	dnsPro := &cobra.Command{Use: "providers", Short: "Manage DNS providers"}
	add(dnsPro,
		jsonCmd("list", "List providers", "GET", "/api/dns/providers", 0),
		bodyCmd("create", "Add a provider", "POST", "/api/dns/providers", 0),
		jsonCmd("delete <id>", "Remove a provider", "DELETE", "/api/dns/providers/%s", 1),
		jsonCmd("sync <id>", "Sync a provider", "POST", "/api/dns/providers/%s/sync", 1),
	)
	dnsZones := &cobra.Command{Use: "zones", Short: "Manage DNS zones"}
	add(dnsZones,
		jsonCmd("list", "List zones", "GET", "/api/dns/zones", 0),
		bodyCmd("create", "Add a zone", "POST", "/api/dns/zones", 0),
		jsonCmd("records <id>", "List a zone's records", "GET", "/api/dns/zones/%s/records", 1),
		bodyCmd("records add <id>", "Add a record to a zone", "POST", "/api/dns/zones/%s/records", 1),
		jsonCmd("delete <id>", "Delete a zone", "DELETE", "/api/dns/zones/%s", 1),
	)
	dnsRec := &cobra.Command{Use: "records", Short: "Manage DNS records"}
	add(dnsRec,
		bodyCmd("update <id>", "Update a record", "PUT", "/api/dns/records/%s", 1),
		jsonCmd("delete <id>", "Delete a record", "DELETE", "/api/dns/records/%s", 1),
	)
	add(dns, dnsPro, dnsZones, dnsRec)
	rootCmd.AddCommand(dns)

	// ---- ssl ----
	ssl := &cobra.Command{Use: "ssl", Short: "🔐 TLS certificates on a chamber"}
	add(ssl,
		jsonCmd("list <serverId>", "List certificates", "GET", "/api/servers/%s/ssl", 1),
		bodyCmd("issue <serverId>", "Issue a certificate", "POST", "/api/servers/%s/ssl/issue", 1),
		jsonCmd("check <serverId>", "Check certificate health", "POST", "/api/servers/%s/ssl/check", 1),
		jsonCmd("uninstall <serverId>", "Remove SSL", "POST", "/api/servers/%s/ssl/uninstall", 1),
	)
	rootCmd.AddCommand(ssl)

	// ---- nginx ----
	nginx := &cobra.Command{Use: "nginx", Short: "⚙️ nginx vhosts and edge"}
	add(nginx,
		jsonCmd("vhosts <serverId>", "List vhosts", "GET", "/api/servers/%s/nginx", 1),
		bodyCmd("vhost <serverId>", "Create a vhost", "POST", "/api/servers/%s/nginx/vhost", 1),
		jsonCmd("reload <serverId>", "Reload nginx", "POST", "/api/servers/%s/nginx/reload", 1),
		jsonCmd("cert-expiry <serverId>", "Check cert expiry", "GET", "/api/servers/%s/nginx/cert-expiry", 1),
	)
	rootCmd.AddCommand(nginx)

	// ---- domains ----
	domains := &cobra.Command{Use: "domains", Short: "🏷️ White-label domain migration"}
	add(domains,
		bodyCmd("migrate plan", "Plan a migration", "POST", "/api/domains/migrate/plan", 0),
		bodyCmd("migrate", "Run a migration", "POST", "/api/domains/migrate", 0),
	)
	rootCmd.AddCommand(domains)

	// ---- failover ----
	failover := &cobra.Command{Use: "failover", Short: "🔁 HA failover groups"}
	add(failover,
		jsonCmd("list", "List groups", "GET", "/api/failover", 0),
		bodyCmd("create", "Create a group", "POST", "/api/failover", 0),
		jsonCmd("health <id>", "Check group health", "GET", "/api/failover/%s/health", 1),
		jsonCmd("promote <id>", "Promote a standby", "POST", "/api/failover/%s/promote", 1),
	)
	rootCmd.AddCommand(failover)
}
