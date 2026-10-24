package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	Version   = "0.1.1"
	Commit    = "unknown"
	BuildDate = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the Odux CLI version",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("%s odux %s (commit %s, built %s)\n", emoji(fleur), Version, Commit, BuildDate)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
