// Command t displays the current time in various timezones using IATA airport codes.
//
// Usage:
//
//	t <IATA>...
//	t -d | --date <IATA>...
//	t -v | --version
//
// Examples:
//
//	$ t sfo jfk
//	SFO: 🕓 16:06:21 (America/Los_Angeles)
//	JFK: 🕖 19:06:21 (America/New_York)
//
//	$ t -d sfo nrt
//	SFO: 🕓 15:12:20 Sun Dec 28 (America/Los_Angeles)
//	NRT: 🕘 08:12:20 Mon Dec 29 (Asia/Tokyo)
//
// Flags:
//
//	-d, --date  Show date alongside time (auto-enabled when dates differ)
//	-v, --version  Show version information
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

// Version information set by goreleaser ldflags
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, "usage: t [-d|--date] <IATA>...\n")
		os.Exit(1)
	}

	// Handle version flag
	if os.Args[1] == "-v" || os.Args[1] == "--version" {
		fmt.Printf("t %s (commit: %s, built: %s)\n", version, commit, date)
		return
	}

	// Parse flags
	args := os.Args[1:]
	showDate := false

	if len(args) > 0 && (args[0] == "-d" || args[0] == "--date") {
		showDate = true
		args = args[1:]
	}

	if len(args) == 0 {
		fmt.Fprint(os.Stderr, "usage: t [-d|--date] <IATA>...\n")
		os.Exit(1)
	}

	ps1Format := os.Getenv("PS1_FORMAT") != ""
	clock.ShowAll(os.Stdout, args, ps1Format, showDate, nil)
}
