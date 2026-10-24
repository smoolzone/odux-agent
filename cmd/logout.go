package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/smoolzone/odux-agent/internal/api"
	"github.com/smoolzone/odux-agent/internal/config"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Revoke the stored session and clear it locally",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Ensure()
		if err != nil {
			return err
		}
		if cfg.RefreshToken != "" {
			_ = api.Logout(cfg.Hub, cfg.RefreshToken)
		}
		cfg.AuthKind = ""
		cfg.RefreshToken = ""
		if err := config.Save(cfg); err != nil {
			return err
		}
		fmt.Printf("%s session cleared and revoked\n", emoji(check))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}
