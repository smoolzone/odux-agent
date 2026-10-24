package cmd

import "github.com/spf13/cobra"

func init() {
	// ---- security ----
	security := &cobra.Command{Use: "security", Short: "🛡️ Hardening: allowlist, mTLS, WireGuard"}
	securityIp := &cobra.Command{Use: "ip-allowlist", Short: "IP allowlist"}
	add(securityIp,
		jsonCmd("get", "Show allowlist", "GET", "/api/security/ip-allowlist", 0),
		bodyCmd("set", "Set allowlist", "PUT", "/api/security/ip-allowlist", 0),
	)
	securityMtls := &cobra.Command{Use: "mtls", Short: "mTLS client certificates"}
	add(securityMtls,
		jsonCmd("status", "Show mTLS status", "GET", "/api/security/mtls/status", 0),
		jsonCmd("clients", "List client certs", "GET", "/api/security/mtls/clients", 0),
		bodyCmd("issue", "Issue a client cert", "POST", "/api/security/mtls/clients", 0),
		jsonCmd("revoke <id>", "Revoke a cert", "PUT", "/api/security/mtls/clients/%s/revoke", 1),
	)
	securityWg := &cobra.Command{Use: "wireguard", Short: "WireGuard peers"}
	add(securityWg,
		jsonCmd("status", "Show WireGuard status", "GET", "/api/security/wireguard/status", 0),
		jsonCmd("init", "Initialize WireGuard", "POST", "/api/security/wireguard/init", 0),
		jsonCmd("peers", "List peers", "GET", "/api/security/wireguard/peers", 0),
		bodyCmd("peer add", "Add a peer", "POST", "/api/security/wireguard/peers", 0),
		jsonCmd("peer revoke <id>", "Revoke a peer", "PUT", "/api/security/wireguard/peers/%s/revoke", 1),
	)
	add(security, securityIp, securityMtls, securityWg)
	rootCmd.AddCommand(security)

	// ---- users ----
	users := &cobra.Command{Use: "users", Short: "👥 Users and RBAC"}
	add(users,
		jsonCmd("list", "List users", "GET", "/api/users", 0),
		bodyCmd("add", "Invite a user", "POST", "/api/users", 0),
		bodyCmd("role <id>", "Change a user's role", "PUT", "/api/users/%s/role", 1),
		bodyCmd("status <id>", "Enable/suspend a user", "PUT", "/api/users/%s/status", 1),
		bodyCmd("permissions <id>", "Set a user's permissions", "PUT", "/api/users/%s/permissions", 1),
	)
	rootCmd.AddCommand(users)

	// ---- audit ----
	audit := &cobra.Command{Use: "audit", Short: "📜 Audit log"}
	add(audit, jsonCmd("list", "List audit events", "GET", "/api/audit", 0))
	rootCmd.AddCommand(audit)

	// ---- approvals ----
	approvals := &cobra.Command{Use: "approvals", Short: "✋ Four-eyes change requests"}
	add(approvals,
		jsonCmd("list", "List change requests", "GET", "/api/approvals", 0),
		bodyCmd("create", "Create a request", "POST", "/api/approvals", 0),
		bodyCmd("update <id>", "Approve/reject a request", "PUT", "/api/approvals/%s", 1),
	)
	rootCmd.AddCommand(approvals)

	// ---- app ----
	app := &cobra.Command{Use: "app", Short: "🧩 App profiles and instances"}
	appPro := &cobra.Command{Use: "profiles", Short: "App profiles (templates)"}
	add(appPro,
		jsonCmd("list", "List profiles", "GET", "/api/app-profiles", 0),
		bodyCmd("create", "Create a profile", "POST", "/api/app-profiles", 0),
	)
	appInst := &cobra.Command{Use: "instances", Short: "App instances"}
	add(appInst,
		jsonCmd("list", "List instances", "GET", "/api/app-instances", 0),
		bodyCmd("create", "Create an instance", "POST", "/api/app-instances", 0),
		jsonCmd("deploy <id>", "Deploy an instance", "POST", "/api/app-instances/%s/deploy", 1),
	)
	add(app, appPro, appInst)
	rootCmd.AddCommand(app)

	// ---- config ----
	config := &cobra.Command{Use: "config", Short: "📝 Config-file editor"}
	add(config,
		jsonCmd("get <serverId>", "Read config files", "GET", "/api/servers/%s/config-files", 1),
		bodyCmd("set <serverId>", "Write config files", "PUT", "/api/servers/%s/config-files", 1),
	)
	rootCmd.AddCommand(config)

	// ---- drift ----
	drift := &cobra.Command{Use: "drift", Short: "🔍 Drift detection"}
	add(drift,
		jsonCmd("status <serverId>", "Show drift", "GET", "/api/servers/%s/drift", 1),
		jsonCmd("check <serverId>", "Run a drift check", "POST", "/api/servers/%s/drift/check", 1),
		jsonCmd("restore <serverId> <fileId>", "Restore a file", "POST", "/api/servers/%s/drift/restore/%s", 2),
	)
	rootCmd.AddCommand(drift)

	// ---- system ----
	system := &cobra.Command{Use: "system", Short: "🖥️ Maintenance mode and backups"}
	add(system,
		jsonCmd("maintenance", "Show maintenance mode", "GET", "/api/system/maintenance", 0),
		bodyCmd("maintenance set", "Toggle maintenance mode", "POST", "/api/system/maintenance", 0),
		jsonCmd("backups", "List backups", "GET", "/api/system/backups", 0),
		jsonCmd("backups create", "Create a backup", "POST", "/api/system/backups", 0),
		bodyCmd("backups restore", "Restore a backup", "POST", "/api/system/backups/restore", 0),
	)
	rootCmd.AddCommand(system)

	// ---- alerts ----
	alerts := &cobra.Command{Use: "alerts", Short: "🚨 Alerts"}
	alertRules := &cobra.Command{Use: "rules", Short: "Threshold alert rules"}
	add(alertRules,
		jsonCmd("list", "List rules", "GET", "/api/alerts/rules", 0),
		bodyCmd("create", "Create a rule (--data '{\"metric\":\"cpu\",\"operator\":\"gt\",\"threshold\":90}')", "POST", "/api/alerts/rules", 0),
		jsonCmd("delete <id>", "Delete a rule", "DELETE", "/api/alerts/rules/%s", 1),
	)
	add(alerts,
		jsonCmd("list", "List alerts", "GET", "/api/alerts", 0),
		jsonCmd("ack <id>", "Acknowledge an alert", "PUT", "/api/alerts/%s/ack", 1),
		jsonCmd("resolve <id>", "Resolve an alert", "PUT", "/api/alerts/%s/resolve", 1),
		alertRules,
	)
	rootCmd.AddCommand(alerts)

	// ---- billing ----
	billing := &cobra.Command{Use: "billing", Short: "💳 Stripe billing"}
	add(billing,
		jsonCmd("status", "Show billing status", "GET", "/api/billing", 0),
		bodyCmd("subscribe", "Subscribe the workspace", "POST", "/api/billing/subscribe", 0),
		jsonCmd("portal", "Open the Stripe portal", "POST", "/api/billing/portal", 0),
	)
	rootCmd.AddCommand(billing)

	// ---- shell ----
	shell := &cobra.Command{Use: "shell", Short: "⌨️ Direct terminal sessions"}
	add(shell,
		jsonCmd("sessions", "List terminal sessions", "GET", "/api/ops/terminal/sessions", 0),
		bodyCmd("open", "Open a terminal session", "POST", "/api/ops/terminal/sessions", 0),
		jsonCmd("close <id>", "Close a session", "POST", "/api/ops/terminal/sessions/%s/close", 1),
	)
	rootCmd.AddCommand(shell)
}
