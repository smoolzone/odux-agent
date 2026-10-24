//go:build !unix

package termio

import "errors"

// Releases target Linux and macOS only. Elsewhere, refuse to prompt (ReadSecret
// then asks for the environment variable) rather than echo a secret.
func disableEcho() (func(), error) {
	return nil, errors.New("hidden input is not supported on this OS")
}
