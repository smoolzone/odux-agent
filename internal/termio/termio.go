// Package termio reads secrets (passwords, API keys, tokens) from the terminal
// without echoing them. Standard library only, so it adds no module dependency.
package termio

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// IsTerminal reports whether f is an interactive terminal (not a pipe or file).
func IsTerminal(f *os.File) bool {
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

// ReadSecret prints label to stderr and reads one line from stdin. On a terminal
// the typed characters are hidden; when stdin is piped (scripts, CI) the line is
// read as-is, so `echo "$PW" | odux login` keeps working.
func ReadSecret(label string) (string, error) {
	fmt.Fprintf(os.Stderr, "%s: ", label)
	if !IsTerminal(os.Stdin) {
		return readLine(os.Stdin)
	}
	restore, err := disableEcho()
	if err != nil {
		// Could not hide input: refuse rather than silently echo a secret.
		fmt.Fprintln(os.Stderr)
		return "", fmt.Errorf("cannot hide input on this terminal (%v); pass it via the environment variable instead", err)
	}
	s, rerr := readLine(os.Stdin)
	restore()
	fmt.Fprintln(os.Stderr) // the user's Enter was not echoed either
	return s, rerr
}

func readLine(r io.Reader) (string, error) {
	s, err := bufio.NewReader(r).ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	// Secrets are taken verbatim except for the line ending.
	return strings.TrimRight(s, "\r\n"), nil
}
