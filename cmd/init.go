package cmd

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/smoolzone/odux-agent/internal/config"
	"github.com/smoolzone/odux-agent/internal/termio"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Configure the Odux CLI (hub URL + API key)",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("%s Welcome, your grace\n", emoji(fleur))
		c := &config.Config{}

		// A platform box carries an encrypted service key — use it automatically.
		boxHub, boxToken, boxErr := loadBoxEnv()
		if boxErr != nil {
			return boxErr
		}

		hub := os.Getenv("ODUX_HUB")
		if hub == "" {
			if boxHub != "" {
				hub = boxHub
			} else if localHubReachable() {
				hub = "http://127.0.0.1:3200"
			} else {
				hub = prompt("Hub URL", "http://localhost:3200")
			}
		}
		c.Hub = strings.TrimRight(hub, "/")

		token := os.Getenv("ODUX_TOKEN")
		if token == "" {
			if boxToken != "" {
				token = boxToken
				fmt.Printf("%s using the platform-box service key (%s)\n", emoji(check), boxEnvPath())
			} else {
				t, err := promptSecret("API key (sc_...)")
				if err != nil {
					return err
				}
				token = strings.TrimSpace(t)
			}
		}
		c.Token = token

		if err := config.Save(c); err != nil {
			return err
		}
		fmt.Printf("%s configuration written to %s (token stored encrypted)\n", emoji(check), config.Path())
		fmt.Printf("%s next: odux castle · odux status · odux doctor\n", emoji(fleur))
		return nil
	},
}

func localHubReachable() bool {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("http://127.0.0.1:3200/")
	if err != nil {
		return false
	}
	resp.Body.Close()
	return true
}

func prompt(label, def string) string {
	r := bufio.NewReader(os.Stdin)
	if def != "" {
		fmt.Printf("%s [%s]: ", label, def)
	} else {
		fmt.Printf("%s: ", label)
	}
	s, _ := r.ReadString('\n')
	s = strings.TrimSpace(s)
	if s == "" {
		return def
	}
	return s
}

// promptSecret reads a password, key or token without echoing it to the screen.
// Piped stdin still works for scripts.
func promptSecret(label string) (string, error) {
	return termio.ReadSecret(label)
}

func init() {
	rootCmd.AddCommand(initCmd)
}
