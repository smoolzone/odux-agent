package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/smoolzone/odux-agent/internal/api"
	"github.com/smoolzone/odux-agent/internal/config"
)

var (
	flagLoginEmail    string
	flagLoginPassword string
	flagLoginOtp      string
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Sign in as a hub operator (stores an encrypted session)",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			cfg = &config.Config{}
		}
		if cfg.Hub == "" {
			if h := os.Getenv("ODUX_HUB"); h != "" {
				cfg.Hub = h
			} else {
				cfg.Hub = prompt("Hub URL", "http://localhost:3200")
			}
		}
		email := flagLoginEmail
		if email == "" {
			email = prompt("Email", "")
		}
		password := flagLoginPassword
		if password == "" {
			password = os.Getenv("ODUX_PASSWORD")
		}
		if password == "" {
			p, err := promptSecret("Password")
			if err != nil {
				return err
			}
			password = p
		}

		pair, err := api.LoginPair(cfg.Hub, email, password, flagLoginOtp)
		if err != nil {
			return err
		}

		cfg.AuthKind = "session"
		cfg.RefreshToken = pair.RefreshToken
		if err := config.Save(cfg); err != nil {
			return err
		}
		fmt.Printf("%s signed in as %s (session stored encrypted; access tokens never touch disk)\n", emoji(check), email)
		return nil
	},
}

func init() {
	loginCmd.Flags().StringVar(&flagLoginEmail, "email", "", "login email")
	loginCmd.Flags().StringVar(&flagLoginPassword, "password", "", "password (visible in shell history and ps — prefer the hidden prompt or ODUX_PASSWORD)")
	loginCmd.Flags().StringVar(&flagLoginOtp, "otp", "", "TOTP code if 2FA is enabled")
	rootCmd.AddCommand(loginCmd)
}
