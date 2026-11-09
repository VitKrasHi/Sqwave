//go:build !windows

package main

import (
	"fmt"
	"os"
)

// showErrorDialog на не-Windows платформах просто пишет в stderr.
func showErrorDialog(title, text string) {
	fmt.Fprintf(os.Stderr, "\n=== %s ===\n%s\n", title, text)
}
