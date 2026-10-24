// Test helper for termio_test.go: reads one secret and prints what it got.
package main

import (
	"fmt"
	"os"

	"github.com/smoolzone/odux-agent/internal/termio"
)

func main() {
	s, err := termio.ReadSecret("Password")
	if err != nil {
		fmt.Println("ERR", err)
		os.Exit(1)
	}
	fmt.Printf("GOT=[%s]\n", s)
}
