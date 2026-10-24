// Package agent persists the on-host agent identity: the hub URL, the enrolled
// agent/server ids, and the long-lived agent key (encrypted at rest with the
// same local master key the operator session uses).
package agent

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/smoolzone/odux-agent/internal/secret"
)

// Config is the on-host agent identity. Key is decrypted into memory only.
type Config struct {
	Hub      string `yaml:"hub"`
	AgentID  int    `yaml:"agent_id"`
	ServerID int    `yaml:"server_id"`
	KeyEnc   string `yaml:"agent_key_enc,omitempty"`
	Key      string `yaml:"-"`
}

// HomeDir is where agent state lives: /etc/odux for root, ~/.odux otherwise.
func HomeDir() string {
	if h := os.Getenv("ODUX_HOME"); h != "" {
		return h
	}
	if os.Geteuid() == 0 {
		return "/etc/odux"
	}
	h, _ := os.UserHomeDir()
	return filepath.Join(h, ".odux")
}

// Path returns the agent config path (override with ODUX_AGENT_CONFIG).
func Path() string {
	if p := os.Getenv("ODUX_AGENT_CONFIG"); p != "" {
		return p
	}
	return filepath.Join(HomeDir(), "agent.yaml")
}

// Load reads the agent config, decrypting the agent key into memory.
func Load() (*Config, error) {
	c := &Config{}
	if b, err := os.ReadFile(Path()); err == nil {
		if err := yaml.Unmarshal(b, c); err != nil {
			return nil, err
		}
	}
	if c.KeyEnc != "" {
		plain, err := secret.DecryptString(c.KeyEnc)
		if err != nil {
			return nil, err
		}
		c.Key = plain
	}
	if h := os.Getenv("ODUX_HUB"); h != "" {
		c.Hub = h
	}
	return c, nil
}

// Save writes the agent config, encrypting the agent key (0600).
func Save(c *Config) error {
	if err := os.MkdirAll(HomeDir(), 0o755); err != nil {
		return err
	}
	if c.Key != "" {
		enc, err := secret.EncryptString(c.Key)
		if err != nil {
			return err
		}
		c.KeyEnc = enc
	}
	toSave := *c
	toSave.Key = ""
	b, err := yaml.Marshal(&toSave)
	if err != nil {
		return err
	}
	return os.WriteFile(Path(), b, 0o600)
}
