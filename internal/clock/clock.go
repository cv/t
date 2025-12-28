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
func FormatResult(r TimeResult, ps1Format bool) string {
	if !r.Found {
		return fmt.Sprintf("%s: ??:??:?? (Unknown)\n", r.IATA)
	}

	if ps1Format {
		return fmt.Sprintf("%s %s", r.IATA, r.Time.Format(LayoutShort))
	}

	emoji := ClockEmoji(r.Time)
	return fmt.Sprintf("%s: %s  %s (%s)\n", r.IATA, emoji, r.Time.Format(LayoutFull), r.Location)
}

// Show writes the time for a given IATA code to the provided writer.
// If ps1Format is true, outputs a compact format suitable for shell prompts.
// If now is nil, the current time is used.
func Show(w io.Writer, iata string, ps1Format bool, now *time.Time) {
	result := LookupTime(iata, now)
	_, _ = fmt.Fprint(w, FormatResult(result, ps1Format))
}

// ShowAll writes the time for multiple IATA codes to the provided writer.
// If ps1Format is true, outputs a compact format suitable for shell prompts.
// If now is nil, the current time is used.
func ShowAll(w io.Writer, iatas []string, ps1Format bool, now *time.Time) {
	for i, iata := range iatas {
		Show(w, iata, ps1Format, now)
		if ps1Format && i < len(iatas)-1 {
			_, _ = fmt.Fprint(w, " ")
		}
	}
}
