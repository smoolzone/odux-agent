//go:build unix

package termio

import (
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

// disableEcho turns terminal echo off with stty and returns a function that turns
// it back on. Echo is also restored if the user presses Ctrl-C mid-prompt, so the
// shell is never left without echo.
func disableEcho() (func(), error) {
	if err := stty("-echo"); err != nil {
		return nil, err
	}
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	done := make(chan struct{})
	go func() {
		select {
		case s := <-sig:
			_ = stty("echo")
			os.Stderr.WriteString("\n")
			signal.Stop(sig)
			if n, ok := s.(syscall.Signal); ok {
				os.Exit(128 + int(n))
			}
			os.Exit(130)
		case <-done:
		}
	}()
	return func() {
		close(done)
		signal.Stop(sig)
		_ = stty("echo")
	}, nil
}

func stty(arg string) error {
	cmd := exec.Command("stty", arg)
	cmd.Stdin = os.Stdin // stty acts on the terminal attached to its stdin
	return cmd.Run()
}
