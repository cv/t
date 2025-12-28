// Command t displays the current time in various timezones using IATA airport codes.
//
// Usage:
//
//	t <IATA>...
//
// Examples:
//
//	$ t sfo jfk
//	SFO: 🕓  16:06:21 (America/Los_Angeles)
//	JFK: 🕖  19:06:21 (America/New_York)
//
// Environment:
//
//	PS1_FORMAT  If set, output is compact with no decorations (for shell prompts)
package main

import (
	"fmt"
	"os"

	"github.com/cv/t/internal/clock"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, "usage: t <IATA>...\n")
		os.Exit(1)
	}

	ps1Format := os.Getenv("PS1_FORMAT") != ""
	clock.ShowAll(os.Stdout, os.Args[1:], ps1Format, nil)
}
