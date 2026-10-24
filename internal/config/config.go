package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/smoolzone/odux-agent/internal/secret"
)

// Config is the CLI configuration. The token is never stored in plaintext:
// the YAML file carries token_enc (AES-256-GCM) and Token is resolved in memory.
type Config struct {
	Hub         string `yaml:"hub"`
	TokenEnc    string `yaml:"token_enc,omitempty"`
	LegacyToken string `yaml:"token,omitempty"` // pre-encryption configs (migration)
	Workspace   string `yaml:"workspace,omitempty"`
	// AuthKind is "key" (read-only workspace API key, sc_...) or "session"
	// (JWT refresh token, stored encrypted). Writes require a session.
	AuthKind     string `yaml:"auth_kind,omitempty"`
	RefreshEnc   string `yaml:"refresh_enc,omitempty"`
	Token        string `yaml:"-"`
	RefreshToken string `yaml:"-"`
}

func HomeDir() string {
	if h := os.Getenv("ODUX_HOME"); h != "" {
		return h
	}
	h, _ := os.UserHomeDir()
	return filepath.Join(h, ".odux")
}

func Path() string {
	if p := os.Getenv("ODUX_CONFIG"); p != "" {
		return p
	}
	return filepath.Join(HomeDir(), "config.yaml")
}

func Load() (*Config, error) {
	c := &Config{}
	if b, err := os.ReadFile(Path()); err == nil {
		if err := yaml.Unmarshal(b, c); err != nil {
			return nil, err
		}
	}
	// Resolve the token: an explicit ODUX_TOKEN env wins; otherwise decrypt
	// token_enc, or fall back to a legacy plaintext token for migration.
	if os.Getenv("ODUX_TOKEN") == "" {
		switch {
		case c.TokenEnc != "":
			plain, err := secret.DecryptString(c.TokenEnc)
			if err != nil {
				return nil, err
			}
			c.Token = plain
		case c.LegacyToken != "":
			c.Token = c.LegacyToken
		}
	}
	if c.RefreshEnc != "" {
		plain, err := secret.DecryptString(c.RefreshEnc)
		if err != nil {
			return nil, err
		}
		c.RefreshToken = plain
	}
	if h := os.Getenv("ODUX_HUB"); h != "" {
		c.Hub = h
	}
	if t := os.Getenv("ODUX_TOKEN"); t != "" {
		c.Token = t
	}
	if w := os.Getenv("ODUX_WORKSPACE"); w != "" {
		c.Workspace = w
	}
	return c, nil
}

func (c *Config) IsSession() bool {
	return c.AuthKind == "session" || c.RefreshToken != ""
}

func Save(c *Config) error {
	if err := os.MkdirAll(HomeDir(), 0o755); err != nil {
		return err
	}
	if c.Token != "" {
		enc, err := secret.EncryptString(c.Token)
		if err != nil {
			return err
		}
		c.TokenEnc = enc
	}
	if c.RefreshToken != "" {
		enc, err := secret.EncryptString(c.RefreshToken)
		if err != nil {
			return err
		}
		c.RefreshEnc = enc
	}
	c.LegacyToken = ""
	toSave := *c
	toSave.Token = ""
	toSave.RefreshToken = ""
	b, err := yaml.Marshal(&toSave)
	if err != nil {
		return err
	}
	return os.WriteFile(Path(), b, 0o600)
}

func Ensure() (*Config, error) {
	c, err := Load()
	if err != nil {
		return nil, err
	}
	if c.Hub == "" || c.Token == "" {
		return nil, fmt.Errorf("odux is not configured — run `odux init`")
	}
	return c, nil
}
