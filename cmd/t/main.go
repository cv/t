// Command t displays the current time in various timezones using IATA airport codes.
//
// Usage:
//
//	t <IATA>...
//	t <IATA>@<time> <IATA>...
//	t @alias
//	t -d | --date <IATA>...
//	t --overlap [--hours=H-H] <IATA> <IATA>...
//	t --save <name> <IATA>...
//	t --list
//	t --delete <name>
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
//	$ t sfo@9:00 jfk lon
//	SFO: 🕘 09:00  →  JFK: 🕛 12:00, LON: 🕔 17:00
//
//	$ t --overlap sfo lon nrt
//	Working hours overlap (9:00-17:00 local):
//	  No overlapping hours found
//
//	$ t --overlap --hours=8-18 sfo jfk
//	Working hours overlap (8:00-18:00 local):
//	  09:00-15:00 SFO = 12:00-18:00 JFK
//	  (6 hours overlap)
//
//	$ t --save team sfo jfk lon
//	Saved alias 'team'
//
//	$ t @team
//	SFO: 🕓 16:06:21 (America/Los_Angeles)
//	JFK: 🕖 19:06:21 (America/New_York)
//	LON: 🕛 00:06:21 (Europe/London)
//
//	$ t --list
//	team: SFO JFK LON
//
//	$ t --delete team
//	Deleted alias 'team'
//
// Time Conversion:
//
//	Use IATA@HH:MM to specify a time at a location and see the equivalent
//	time in other timezones. Useful for scheduling meetings across timezones.
//
// Meeting Overlap:
//
//	Use --overlap to find overlapping work hours across timezones.
//	Default work hours are 9:00-17:00 local time. Use --hours=H-H or
//	--hours=HH:MM-HH:MM to customize (e.g., --hours=8:00-18:00).
//
// Aliases:
//
//	Save frequently used city groups with --save and recall them with @alias.
//	Aliases are stored in ~/.config/t/aliases.json.
//
// Flags:
//
//	-d, --date     Show date alongside time (auto-enabled when dates differ)
//	--dst          Show DST warnings when a transition is within 5 days
//	--dst=N        Show DST warnings when a transition is within N days
//	--overlap      Find overlapping work hours across timezones
//	--hours=H-H    Custom work hours for overlap calculation (default: 9-17)
//	--save <name>  Save following IATA codes as named alias
//	--list         List all saved aliases
//	--delete <name> Delete a saved alias
//	-v, --version  Show version information
//
// Environment:
//
//	PS1_FORMAT  If set, output is compact with no decorations (for shell prompts)
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/cv/t/internal/clock"
	"github.com/cv/t/internal/config"
)

// Version information set by goreleaser ldflags
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) < 1 {
		fmt.Fprint(os.Stderr, "usage: t [-d|--date] [--dst[=N]] [--overlap [--hours=H-H]] <IATA>...\n")
		fmt.Fprint(os.Stderr, "       t --save <name> <IATA>...\n")
		fmt.Fprint(os.Stderr, "       t --list | --delete <name>\n")
		return 1
	}

	// Handle version flag
	if args[0] == "-v" || args[0] == "--version" {
		fmt.Printf("t %s (commit: %s, built: %s)\n", version, commit, date)
		return 0
	}

	// Handle alias management flags
	if args[0] == "--list" {
		return handleList()
	}
	if args[0] == "--delete" {
		if len(args) < 2 {
			fmt.Fprint(os.Stderr, "usage: t --delete <name>\n")
			return 1
		}
		return handleDelete(args[1])
	}
	if args[0] == "--save" {
		if len(args) < 3 {
			fmt.Fprint(os.Stderr, "usage: t --save <name> <IATA>...\n")
			return 1
		}
		return handleSave(args[1], args[2:])
	}

	// Parse flags
	showDate := false
	showDST := false
	dstWindow := clock.DefaultDSTWindow
	overlapMode := false
	workHours := clock.DefaultWorkHours

	for len(args) > 0 {
		switch {
		case args[0] == "-d" || args[0] == "--date":
			showDate = true
			args = args[1:]
		case args[0] == "--dst":
			showDST = true
			args = args[1:]
		case len(args[0]) > 6 && args[0][:6] == "--dst=":
			showDST = true
			var n int
			if _, err := fmt.Sscanf(args[0][6:], "%d", &n); err != nil || n < 1 {
				fmt.Fprintf(os.Stderr, "invalid DST window: %s (use a positive number)\n", args[0][6:])
				return 1
			}
			dstWindow = n
			args = args[1:]
		case args[0] == "--overlap":
			overlapMode = true
			args = args[1:]
		case len(args[0]) > 8 && args[0][:8] == "--hours=":
			hoursStr := args[0][8:]
			if parsed := clock.ParseWorkHours(hoursStr); parsed != nil {
				workHours = *parsed
			} else {
				fmt.Fprintf(os.Stderr, "invalid work hours format: %s (use H-H or HH:MM-HH:MM)\n", hoursStr)
				return 1
			}
			args = args[1:]
		default:
			goto done
		}
	}
done:

	if len(args) == 0 {
		fmt.Fprint(os.Stderr, "usage: t [-d|--date] [--dst[=N]] [--overlap [--hours=H-H]] <IATA>...\n")
		return 1
	}

	// Expand any @alias references in args
	expandedArgs, err := expandAliases(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}
	args = expandedArgs

	// Handle overlap mode
	if overlapMode {
		if len(args) < 2 {
			fmt.Fprint(os.Stderr, "usage: t --overlap [--hours=H-H] <IATA> <IATA>...\n")
			return 1
		}
		clock.ShowOverlap(os.Stdout, args, workHours, nil)
		return 0
	}

	ps1Format := os.Getenv("PS1_FORMAT") != ""

	// Check if first argument is a time spec (e.g., "SFO@9:00")
	if spec := clock.ParseTimeSpec(args[0]); spec != nil {
		if len(args) < 2 {
			fmt.Fprint(os.Stderr, "usage: t <IATA>@<time> <IATA>...\n")
			return 1
		}
		clock.ShowConversion(os.Stdout, *spec, args[1:], ps1Format, nil)
		return 0
	}

	clock.ShowAllWithDST(os.Stdout, args, ps1Format, showDate, showDST, dstWindow, nil)
	return 0
}

// handleSave saves an alias with the given name and IATA codes.
func handleSave(name string, codes []string) int {
	store, err := config.NewAliasStore()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}

	if err := store.Save(name, codes); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}

	fmt.Printf("Saved alias '%s'\n", name)
	return 0
}

// handleList lists all saved aliases.
func handleList() int {
	store, err := config.NewAliasStore()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}

	names := store.ListSorted()
	if len(names) == 0 {
		fmt.Println("No aliases saved")
		return 0
	}

	aliases := store.List()
	for _, name := range names {
		codes := aliases[name]
		fmt.Printf("%s: %s\n", name, strings.Join(codes, " "))
	}
	return 0
}

// handleDelete deletes an alias by name.
func handleDelete(name string) int {
	store, err := config.NewAliasStore()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}

	if err := store.Delete(name); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}

	fmt.Printf("Deleted alias '%s'\n", name)
	return 0
}

// expandAliases expands any @alias references in the argument list.
// Returns the expanded list of IATA codes.
func expandAliases(args []string) ([]string, error) {
	var result []string
	var store *config.AliasStore

	for _, arg := range args {
		if strings.HasPrefix(arg, "@") {
			// Lazy initialization of store
			if store == nil {
				var err error
				store, err = config.NewAliasStore()
				if err != nil {
					return nil, fmt.Errorf("error loading aliases: %w", err)
				}
			}

			aliasName := arg[1:] // Remove @ prefix
			codes := store.Get(aliasName)
			if codes == nil {
				return nil, fmt.Errorf("unknown alias: %s", aliasName)
			}
			result = append(result, codes...)
		} else {
			result = append(result, arg)
		}
	}

	return result, nil
}
