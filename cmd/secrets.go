package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/smoolzone/odux-agent/internal/api"
)

var (
	flagSecretValue string
	flagSecretNote  string
)

// secretsCmd manages workspace secrets. Values are encrypted at rest and never
// shown after write except on an explicit `get`.
var secretsCmd = &cobra.Command{
	Use:   "secrets",
	Short: "Manage workspace secrets (encrypted at rest)",
	Long:  "Workspace secrets are encrypted at rest (AES-256-GCM). Values are never shown after write except on an explicit `odux secrets get`.",
}

var secretsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List secret keys (values never shown)",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, at, err := session(cmd)
		if err != nil {
			return err
		}
		out, err := c.ListSecrets(at)
		if err != nil {
			return err
		}
		if flagOutput == "json" {
			printJSON(out)
			return nil
		}
		if len(out) == 0 {
			fmt.Println("(no secrets)")
			return nil
		}
		for _, s := range out {
			fmt.Printf("%s %s%s\n", emoji(shield), s.Key, noteSuffix(s.Note))
		}
		return nil
	},
}

var secretsSetCmd = &cobra.Command{
	Use:   "set <key>",
	Short: "Set a secret (overwrites if present)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, at, err := session(cmd)
		if err != nil {
			return err
		}
		value := flagSecretValue
		if value == "" {
			value = os.Getenv("ODUX_SECRET")
		}
		if value == "" {
			v, err := promptSecret("Value")
			if err != nil {
				return err
			}
			value = v
		}
		if _, err := c.SetSecret(at, args[0], value, flagSecretNote); err != nil {
			return err
		}
		fmt.Printf("%s set secret %q\n", emoji(check), args[0])
		return nil
	},
}

var secretsGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Read a secret's value (explicit decrypt)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, at, err := session(cmd)
		if err != nil {
			return err
		}
		id, err := findSecretID(c, at, args[0])
		if err != nil {
			return err
		}
		got, err := c.GetSecret(at, id)
		if err != nil {
			return err
		}
		fmt.Println(got.Value)
		return nil
	},
}

var secretsRmCmd = &cobra.Command{
	Use:   "rm <key>",
	Short: "Delete a secret",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, at, err := session(cmd)
		if err != nil {
			return err
		}
		id, err := findSecretID(c, at, args[0])
		if err != nil {
			return err
		}
		if err := c.DeleteSecret(at, id); err != nil {
			return err
		}
		fmt.Printf("%s removed secret %q\n", emoji(check), args[0])
		return nil
	},
}

func findSecretID(c *api.Client, at, key string) (string, error) {
	list, err := c.ListSecrets(at)
	if err != nil {
		return "", err
	}
	for _, s := range list {
		if s.Key == key {
			return fmt.Sprintf("%d", s.ID), nil
		}
	}
	return "", fmt.Errorf("secret %q not found", key)
}

func noteSuffix(note string) string {
	if note == "" {
		return ""
	}
	return "  (" + note + ")"
}

func init() {
	secretsSetCmd.Flags().StringVar(&flagSecretValue, "value", "", "secret value (visible in shell history and ps — prefer the hidden prompt or ODUX_SECRET)")
	secretsSetCmd.Flags().StringVar(&flagSecretNote, "note", "", "optional note")
	secretsCmd.AddCommand(secretsListCmd, secretsSetCmd, secretsGetCmd, secretsRmCmd)
	rootCmd.AddCommand(secretsCmd)
}
