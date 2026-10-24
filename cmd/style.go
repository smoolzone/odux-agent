package cmd

import "strings"

// Royalty-themed branding for the Duke (Odux).
//
//	⚜️ fleur-de-lis = the brand & you (you carry the Duke's symbol)
//	🏰 castle       = all your servers
//	🛡️ shield       = one server (a chamber in the castle)
//	⚔️ sword        = ops
const (
	fleur     = "⚜️"
	castle    = "🏰"
	shield    = "🛡️"
	sword     = "⚔️"
	check     = "✅"
	cross     = "❌"
	warnMark  = "⚠️"
	dotGreen  = "🟢"
	dotRed    = "🔴"
	dotYellow = "🟡"
	dotWhite  = "⚪"
)

// useEmoji is disabled via --no-color so piped/CI output stays clean.
var useEmoji = true

// emoji returns the symbol when emojis are enabled, otherwise an empty string.
func emoji(s string) string {
	if !useEmoji {
		return ""
	}
	return s
}

// statusEmoji maps a server/health status to a colored dot.
func statusEmoji(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "online", "active", "up", "healthy", "ready":
		return dotGreen
	case "offline", "down", "error", "failed", "unreachable":
		return dotRed
	case "pending", "provisioning", "warning", "degraded":
		return dotYellow
	default:
		return dotWhite
	}
}
