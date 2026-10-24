package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/smoolzone/odux-agent/internal/api"
	"github.com/smoolzone/odux-agent/internal/config"
)

// session resolves a fresh access token from the stored encrypted session, and
// persists the rotated refresh token the server returns.
func session(cmd *cobra.Command) (*api.Client, string, error) {
	cfg, err := config.Ensure()
	if err != nil {
		return nil, "", err
	}
	c := api.New(cfg)
	pair, err := api.RefreshPair(cfg.Hub, cfg.RefreshToken)
	if err != nil {
		return nil, "", err
	}
	if pair.RefreshToken != "" && pair.RefreshToken != cfg.RefreshToken {
		cfg.RefreshToken = pair.RefreshToken
		if err := config.Save(cfg); err != nil {
			return nil, "", err
		}
	}
	return c, pair.AccessToken, nil
}

func printJSON(v any) {
	b, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(b))
}

// jsonCmd builds a command that hits a path (with nargs positional params
// filled into the path via Sprintf) and prints the JSON response.
func jsonCmd(use, short, method, path string, nargs int) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != nargs {
				return fmt.Errorf("expected %d argument(s), got %d", nargs, len(args))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			c, at, err := session(cmd)
			if err != nil {
				return err
			}
			p := path
			if nargs > 0 {
				p = fmt.Sprintf(path, iface(args)...)
			}
			var out any
			if method == "GET" {
				err = c.Get(p, at, &out)
			} else {
				err = c.Write(method, p, at, nil, &out)
			}
			if err != nil {
				return err
			}
			if out == nil {
				fmt.Println("ok")
				return nil
			}
			printJSON(out)
			return nil
		},
	}
}

// bodyCmd builds a write command that accepts a `--data '{"k":...}'` JSON body.
func bodyCmd(use, short, method, path string, nargs int) *cobra.Command {
	var data string
	c := &cobra.Command{
		Use:   use,
		Short: short,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != nargs {
				return fmt.Errorf("expected %d argument(s), got %d", nargs, len(args))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, at, err := session(cmd)
			if err != nil {
				return err
			}
			p := path
			if nargs > 0 {
				p = fmt.Sprintf(path, iface(args)...)
			}
			var body any
			if data != "" {
				if err := json.Unmarshal([]byte(data), &body); err != nil {
					return fmt.Errorf("--data is not valid JSON: %w", err)
				}
			}
			var out any
			if err := cl.Write(method, p, at, body, &out); err != nil {
				return err
			}
			if out == nil {
				fmt.Println("ok")
				return nil
			}
			printJSON(out)
			return nil
		},
	}
	c.Flags().StringVar(&data, "data", "", "JSON body for the request")
	return c
}

func iface(args []string) []any {
	out := make([]any, len(args))
	for i, a := range args {
		out[i] = a
	}
	return out
}
