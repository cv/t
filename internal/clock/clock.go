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
	if showDate {
		return fmt.Sprintf("%s: %s %s %s (%s)\n", r.IATA, emoji, r.Time.Format(LayoutFull), r.Time.Format(LayoutDate), r.Location)
	}
	return fmt.Sprintf("%s: %s %s (%s)\n", r.IATA, emoji, r.Time.Format(LayoutFull), r.Location)
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
