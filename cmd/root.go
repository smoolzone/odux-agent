package cmd

import (
	"github.com/spf13/cobra"
)

var (
	flagWorkspace string
	flagOutput    string
	flagVerbose   bool
	flagDebug     bool
	flagQuiet     bool
	flagNoColor   bool
	flagTimeout   int
	flagYes       bool
)

var rootCmd = &cobra.Command{
	Use:   "odux",
	Short: "⚜️ Odux — the control plane for your castle",
	Long: `⚜️  Odux ("O Dux" — the Duke) commands your castle of servers.
🏰  Every server is a chamber; you carry the Duke's fleur-de-lis.
⚔️  The Duke is always watching your castle.`,
	SilenceErrors: true,
	SilenceUsage:  true,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		useEmoji = !flagNoColor
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	pf := rootCmd.PersistentFlags()
	pf.StringVar(&flagWorkspace, "workspace", "", "workspace to operate on (default: current context)")
	pf.StringVar(&flagOutput, "output", "text", "output format: text, json, yaml")
	pf.BoolVarP(&flagVerbose, "verbose", "v", false, "verbose output")
	pf.BoolVar(&flagDebug, "debug", false, "debug output (implies --verbose)")
	pf.BoolVarP(&flagQuiet, "quiet", "q", false, "suppress non-error output")
	pf.BoolVar(&flagNoColor, "no-color", false, "disable color output")
	pf.IntVar(&flagTimeout, "timeout", 30, "request timeout in seconds")
	pf.BoolVarP(&flagYes, "yes", "y", false, "assume yes to all prompts")
}
