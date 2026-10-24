package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/smoolzone/odux-agent/internal/api"
	"github.com/smoolzone/odux-agent/internal/secret"
)

// boxEnvPath is where the platform box keeps its encrypted service key.
func boxEnvPath() string {
	if p := os.Getenv("ODUX_BOX_ENV"); p != "" {
		return p
	}
	return "/etc/odux/box.env"
}

// boxEnvHub returns the hub address for box operations.
func boxEnvHub() string {
	if h := os.Getenv("ODUX_HUB"); h != "" {
		return h
	}
	return "http://127.0.0.1:3200"
}

// loadBoxEnv reads the box's encrypted service key. Returns ("", "", nil)
// when the box env file does not exist (a normal state on non-box machines).
func loadBoxEnv() (hub, token string, err error) {
	b, err := os.ReadFile(boxEnvPath())
	if err != nil {
		return "", "", nil
	}
	vars := map[string]string{}
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if i := strings.Index(line, "="); i > 0 {
			vars[strings.TrimSpace(line[:i])] = strings.TrimSpace(line[i+1:])
		}
	}
	hub = vars["ODUX_HUB"]
	if enc := vars["ODUX_TOKEN_ENC"]; enc != "" {
		token, err = secret.DecryptString(enc)
		if err != nil {
			return "", "", fmt.Errorf("cannot decrypt %s: %w", boxEnvPath(), err)
		}
	}
	return hub, token, nil
}

var boxCmd = &cobra.Command{
	Use:   "box",
	Short: "🏰 Platform box commands",
}

var boxSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Bootstrap this platform box (encrypted service key — no admin password stored)",
	RunE: func(cmd *cobra.Command, args []string) error {
		hub := boxEnvHub()
		email := os.Getenv("ODUX_ADMIN_EMAIL")
		password := os.Getenv("ODUX_ADMIN_PASSWORD")
		if email == "" || password == "" {
			return fmt.Errorf("box setup needs ODUX_ADMIN_EMAIL and ODUX_ADMIN_PASSWORD (given at setup time, never stored)")
		}
		host, _ := os.Hostname()
		if host == "" {
			host = "box"
		}
		name := fmt.Sprintf("odux-cli-%s", host)

		fmt.Printf("%s logging into %s as %s\n", emoji(fleur), hub, email)
		jwt, err := api.Login(hub, email, password)
		if err != nil {
			return err
		}

		fmt.Printf("%s creating service key %s\n", emoji(fleur), name)
		key, err := api.CreateWorkspaceAPIKey(hub, jwt, name)
		if err != nil {
			return err
		}

		enc, err := secret.EncryptString(key)
		if err != nil {
			return err
		}
		path := boxEnvPath()
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			return err
		}
		content := fmt.Sprintf("ODUX_HUB=%s\nODUX_TOKEN_ENC=%s\n", hub, enc)
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			return err
		}

		fmt.Printf("%s service key encrypted and stored in %s\n", emoji(check), path)
		fmt.Printf("%s the admin password was used once and never written to disk\n", emoji(check))
		fmt.Printf("%s next: odux init · odux castle\n", emoji(fleur))
		return nil
	},
}

var boxStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the platform-box key status",
	RunE: func(cmd *cobra.Command, args []string) error {
		path := boxEnvPath()
		if _, err := os.Stat(path); err != nil {
			fmt.Printf("%s no box key at %s (run `odux box setup`)\n", emoji(warnMark), path)
			return nil
		}
		hub, _, err := loadBoxEnv()
		if err != nil {
			return err
		}
		fmt.Printf("%s box key present: %s\n", emoji(check), path)
		fmt.Printf("%s hub: %s\n", emoji(check), hub)
		fmt.Printf("%s key file: %s\n", emoji(check), secret.KeyPath())
		return nil
	},
}

func init() {
	rootCmd.AddCommand(boxCmd)
	boxCmd.AddCommand(boxSetupCmd)
	boxCmd.AddCommand(boxStatusCmd)
}
