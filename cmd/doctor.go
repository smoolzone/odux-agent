package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/smoolzone/odux-agent/internal/config"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnose the Odux CLI installation and hub connectivity",
	RunE: func(cmd *cobra.Command, args []string) error {
		home := os.Getenv("ODUX_HOME")
		if home == "" {
			if h, err := os.UserHomeDir(); err == nil {
				home = filepath.Join(h, ".odux")
			}
		}
		fmt.Printf("%s  home directory: %s\n", emoji(check), home)

		cfgPath := os.Getenv("ODUX_CONFIG")
		if cfgPath == "" {
			cfgPath = filepath.Join(home, "config.yaml")
		}
		if _, err := os.Stat(cfgPath); err == nil {
			fmt.Printf("%s  config file: %s\n", emoji(check), cfgPath)
		} else {
			fmt.Printf("%s config file missing: %s (run `odux init`)\n", emoji(warnMark), cfgPath)
		}

		if exe, err := os.Executable(); err == nil {
			fmt.Printf("%s  binary: %s\n", emoji(check), exe)
		}

		hub := os.Getenv("ODUX_HUB")
		if cfg, err := config.Load(); err == nil && cfg.Hub != "" {
			hub = cfg.Hub
		}
		if hub == "" {
			fmt.Printf("%s hub not configured: set ODUX_HUB or run `odux init`\n", emoji(warnMark))
		} else {
			fmt.Printf("%s  hub configured: %s\n", emoji(check), hub)
		}

		fmt.Printf("%s The Duke is ready.\n", emoji(fleur))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}
