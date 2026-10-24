package secret

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// KeyPath returns the path of the local AES-256 key.
// Root on a platform box → /etc/odux/odux.key; elsewhere → ~/.odux/master.key.
// Override with ODUX_KEY_FILE.
func KeyPath() string {
	if p := os.Getenv("ODUX_KEY_FILE"); p != "" {
		return p
	}
	if os.Geteuid() == 0 {
		return "/etc/odux/odux.key"
	}
	h, _ := os.UserHomeDir()
	return filepath.Join(h, ".odux", "master.key")
}

// LoadOrCreateKey loads the 32-byte key, generating and storing it 0600 on first use.
func LoadOrCreateKey() ([]byte, error) {
	path := KeyPath()
	if b, err := os.ReadFile(path); err == nil {
		return b, nil
	}
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, err
	}
	if err := os.WriteFile(path, key, 0o600); err != nil {
		return nil, err
	}
	return key, nil
}

// LoadKey requires the key to already exist (decrypt path).
func LoadKey() ([]byte, error) {
	path := KeyPath()
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("encryption key not found at %s (run `odux init` or `odux box setup`): %w", path, err)
	}
	return b, nil
}

// EncryptString encrypts with the local key, creating the key on first use.
func EncryptString(plain string) (string, error) {
	key, err := LoadOrCreateKey()
	if err != nil {
		return "", err
	}
	return encrypt(key, plain)
}

// DecryptString decrypts ciphertext produced by EncryptString.
func DecryptString(enc string) (string, error) {
	key, err := LoadKey()
	if err != nil {
		return "", err
	}
	return decrypt(key, enc)
}

func encrypt(key []byte, plain string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	out := gcm.Seal(nonce, nonce, []byte(plain), nil)
	return base64.StdEncoding.EncodeToString(out), nil
}

func decrypt(key []byte, enc string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(enc)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(data) < gcm.NonceSize() {
		return "", errors.New("ciphertext too short")
	}
	plain, err := gcm.Open(nil, data[:gcm.NonceSize()], data[gcm.NonceSize():], nil)
	if err != nil {
		return "", fmt.Errorf("decryption failed (wrong key?): %w", err)
	}
	return string(plain), nil
}
