// Package clock provides timezone display functionality using IATA airport codes.
package clock

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/cv/t/codes"
)

const (
	// LayoutFull is the full time format with seconds.
	LayoutFull = "15:04:05"
	// LayoutShort is the short time format without seconds.
	LayoutShort = "15:04"
	// LayoutDate is the date format for display.
	LayoutDate = "Mon Jan 2"
)

var clocksLow = []string{
	"🕛", "🕐", "🕑", "🕒", "🕓", "🕔", "🕕", "🕖", "🕗", "🕘", "🕙", "🕚",
	"🕛", "🕐", "🕑", "🕒", "🕓", "🕔", "🕕", "🕖", "🕗", "🕘", "🕙", "🕚",
}

var clocksHigh = []string{
	"🕧", "🕜", "🕝", "🕞", "🕟", "🕠", "🕡", "🕢", "🕣", "🕤", "🕥", "🕦",
	"🕧", "🕜", "🕝", "🕞", "🕟", "🕠", "🕡", "🕢", "🕣", "🕤", "🕥", "🕦",
}

// TimeResult holds the result of a timezone lookup.
type TimeResult struct {
	IATA     string
	Time     time.Time
	Location string
	Found    bool
}

// RelativeOffset calculates the offset of t's timezone from the local timezone.
// Returns a string like "(+8h)", "(-5h)", "(+5h30m)", or "(+0h)" if same timezone.
func RelativeOffset(t time.Time) string {
	localLoc := time.Local
	localTime := t.In(localLoc)

	// Get the offset in seconds for both timezones at this instant
	_, targetOffset := t.Zone()
	_, localOffset := localTime.Zone()

	// Calculate the difference in seconds
	diffSeconds := targetOffset - localOffset

	// Convert to hours and minutes
	diffMinutes := diffSeconds / 60
	hours := diffMinutes / 60
	minutes := diffMinutes % 60

	// Handle negative minutes (e.g., -5h30m should be -5h -30m -> -5h30m)
	if minutes < 0 {
		minutes = -minutes
	}

	// Format the offset string
	sign := "+"
	if hours < 0 || (hours == 0 && diffSeconds < 0) {
		sign = "-"
		if hours < 0 {
			hours = -hours
		}
	}

	if minutes == 0 {
		return fmt.Sprintf("(%s%dh)", sign, hours)
	}
	return fmt.Sprintf("(%s%dh%dm)", sign, hours, minutes)
}

// ClockEmoji returns the appropriate clock emoji for the given time.
// It uses half-hour emojis when minutes > 30.
func ClockEmoji(t time.Time) string {
	hour := t.Hour()
	if t.Minute() > 30 {
		return clocksHigh[hour]
	}
	return clocksLow[hour]
}

// LookupTime returns the current time for a given IATA airport code.
// If now is nil, the current time is used.
func LookupTime(iata string, now *time.Time) TimeResult {
	iata = strings.ToUpper(iata)

	locName, found := codes.IATA[iata]
	if !found {
		return TimeResult{
			IATA:  iata,
			Found: false,
		}
	}

	loc, err := time.LoadLocation(locName)
	if err != nil {
		return TimeResult{
			IATA:  iata,
			Found: false,
		}
	}

	var t time.Time
	if now != nil {
		t = now.In(loc)
	} else {
		t = time.Now().In(loc)
	}

	return TimeResult{
		IATA:     iata,
		Time:     t,
		Location: locName,
		Found:    true,
	}
}

// FormatResult formats a TimeResult for display.
// If ps1Format is true, outputs a compact format suitable for shell prompts.
// If showDate is true, includes the date alongside the time.
func FormatResult(r TimeResult, ps1Format, showDate bool) string {
	if !r.Found {
		return fmt.Sprintf("%s: ??:??:?? (Unknown)\n", r.IATA)
	}

	if ps1Format {
		return fmt.Sprintf("%s %s", r.IATA, r.Time.Format(LayoutShort))
	}

	emoji := ClockEmoji(r.Time)
	offset := RelativeOffset(r.Time)
	if showDate {
		return fmt.Sprintf("%s: %s %s %s %s (%s)\n", r.IATA, emoji, r.Time.Format(LayoutFull), r.Time.Format(LayoutDate), offset, r.Location)
	}
	return fmt.Sprintf("%s: %s %s %s (%s)\n", r.IATA, emoji, r.Time.Format(LayoutFull), offset, r.Location)
}

// Show writes the time for a given IATA code to the provided writer.
// If ps1Format is true, outputs a compact format suitable for shell prompts.
// If showDate is true, includes the date alongside the time.
// If now is nil, the current time is used.
func Show(w io.Writer, iata string, ps1Format, showDate bool, now *time.Time) {
	result := LookupTime(iata, now)
	_, _ = fmt.Fprint(w, FormatResult(result, ps1Format, showDate))
}

// ShowAll writes the time for multiple IATA codes to the provided writer.
// If ps1Format is true, outputs a compact format suitable for shell prompts.
// If showDate is true, includes the date alongside the time.
// If showDate is false but dates differ across results, date is shown automatically.
// If now is nil, the current time is used.
func ShowAll(w io.Writer, iatas []string, ps1Format, showDate bool, now *time.Time) {
	// Collect all results first
	results := make([]TimeResult, len(iatas))
	for i, iata := range iatas {
		results[i] = LookupTime(iata, now)
	}

	// If showDate is not explicitly requested, check if dates differ
	if !showDate && !ps1Format && len(results) > 1 {
		showDate = datesDiffer(results)
	}

	// Output results
	for i, result := range results {
		_, _ = fmt.Fprint(w, FormatResult(result, ps1Format, showDate))
		if ps1Format && i < len(results)-1 {
			_, _ = fmt.Fprint(w, " ")
		}
	}
}

// datesDiffer returns true if any of the found results have different dates.
func datesDiffer(results []TimeResult) bool {
	var firstDate string
	for _, r := range results {
		if !r.Found {
			continue
		}
		date := r.Time.Format("2006-01-02")
		if firstDate == "" {
			firstDate = date
		} else if date != firstDate {
			return true
		}
	}
	return false
}
