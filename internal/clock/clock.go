// Package clock provides timezone display functionality using IATA airport codes.
package clock

import (
	"fmt"
	"io"
	"regexp"
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

// timeSpecRegex matches patterns like "SFO@9:00", "jfk@14:30", "lhr@9"
var timeSpecRegex = regexp.MustCompile(`(?i)^([A-Z]{3})@(\d{1,2}):?(\d{2})?$`)

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
	Weather  string // Optional weather info (e.g., "☀️ +13°C")
}

// DisplayOptions controls how time results are formatted.
type DisplayOptions struct {
	PS1Format bool // Compact format for shell prompts
	ShowDate  bool // Show date alongside time
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
func FormatResult(r *TimeResult, ps1Format, showDate bool) string {
	if !r.Found {
		return fmt.Sprintf("%s: ??:??:?? (Unknown)\n", r.IATA)
	}

	if ps1Format {
		if r.Weather != "" {
			return fmt.Sprintf("%s %s %s", r.IATA, r.Time.Format(LayoutShort), r.Weather)
		}
		return fmt.Sprintf("%s %s", r.IATA, r.Time.Format(LayoutShort))
	}

	emoji := ClockEmoji(r.Time)
	offset := RelativeOffset(r.Time)
	weather := ""
	if r.Weather != "" {
		weather = " " + r.Weather
	}
	if showDate {
		return fmt.Sprintf("%s: %s %s %s%s %s (%s)\n", r.IATA, emoji, r.Time.Format(LayoutFull), r.Time.Format(LayoutDate), weather, offset, r.Location)
	}
	return fmt.Sprintf("%s: %s %s%s %s (%s)\n", r.IATA, emoji, r.Time.Format(LayoutFull), weather, offset, r.Location)
}

// Show writes the time for a given IATA code to the provided writer.
// If ps1Format is true, outputs a compact format suitable for shell prompts.
// If showDate is true, includes the date alongside the time.
// If now is nil, the current time is used.
func Show(w io.Writer, iata string, ps1Format, showDate bool, now *time.Time) {
	result := LookupTime(iata, now)
	_, _ = fmt.Fprint(w, FormatResult(&result, ps1Format, showDate))
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
	for i := range results {
		_, _ = fmt.Fprint(w, FormatResult(&results[i], ps1Format, showDate))
		if ps1Format && i < len(results)-1 {
			_, _ = fmt.Fprint(w, " ")
		}
	}
}

// ShowAllWithWeather writes the time for multiple IATA codes with weather info.
// weatherData is a map of IATA code to weather string (e.g., "☀️ +13°C").
func ShowAllWithWeather(w io.Writer, iatas []string, ps1Format, showDate bool, now *time.Time, weatherData map[string]string) {
	// Collect all results first
	results := make([]TimeResult, len(iatas))
	for i, iata := range iatas {
		results[i] = LookupTime(iata, now)
		if weather, ok := weatherData[strings.ToUpper(iata)]; ok {
			results[i].Weather = weather
		}
	}

	// If showDate is not explicitly requested, check if dates differ
	if !showDate && !ps1Format && len(results) > 1 {
		showDate = datesDiffer(results)
	}

	// Output results
	for i := range results {
		_, _ = fmt.Fprint(w, FormatResult(&results[i], ps1Format, showDate))
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

// TimeSpec represents a parsed time specification like "SFO@9:00".
type TimeSpec struct {
	IATA   string
	Hour   int
	Minute int
}

// ParseTimeSpec parses a time specification string like "SFO@9:00" or "jfk@14:30".
// Returns nil if the string is not a valid time spec (just a plain IATA code).
func ParseTimeSpec(s string) *TimeSpec {
	matches := timeSpecRegex.FindStringSubmatch(s)
	if matches == nil {
		return nil
	}

	iata := strings.ToUpper(matches[1])
	hour := 0
	minute := 0

	// Parse hour
	if _, err := fmt.Sscanf(matches[2], "%d", &hour); err != nil {
		return nil
	}
	if hour < 0 || hour > 23 {
		return nil
	}

	// Parse minute if present
	if matches[3] != "" {
		if _, err := fmt.Sscanf(matches[3], "%d", &minute); err != nil {
			return nil
		}
		if minute < 0 || minute > 59 {
			return nil
		}
	}

	return &TimeSpec{
		IATA:   iata,
		Hour:   hour,
		Minute: minute,
	}
}

// ResolveTime creates a time.Time for this TimeSpec based on a reference time.
// The resulting time will be at the specified hour:minute in the IATA location's timezone.
func (ts *TimeSpec) ResolveTime(ref time.Time) (time.Time, error) {
	locName, found := codes.IATA[ts.IATA]
	if !found {
		return time.Time{}, fmt.Errorf("unknown IATA code: %s", ts.IATA)
	}

	loc, err := time.LoadLocation(locName)
	if err != nil {
		return time.Time{}, fmt.Errorf("loading location %s: %w", locName, err)
	}

	// Get reference time in the target location
	refInLoc := ref.In(loc)

	// Create a new time with the specified hour and minute
	result := time.Date(
		refInLoc.Year(), refInLoc.Month(), refInLoc.Day(),
		ts.Hour, ts.Minute, 0, 0, loc,
	)

	return result, nil
}

// ConversionResult holds the result of a time conversion.
type ConversionResult struct {
	Source  TimeResult
	Targets []TimeResult
}

// FormatConversion formats a conversion result for display.
func FormatConversion(c *ConversionResult, ps1Format bool) string {
	if !c.Source.Found {
		return fmt.Sprintf("%s: Unknown airport code\n", c.Source.IATA)
	}

	if ps1Format {
		var parts []string
		parts = append(parts, fmt.Sprintf("%s %s", c.Source.IATA, c.Source.Time.Format(LayoutShort)))
		for _, t := range c.Targets {
			if t.Found {
				parts = append(parts, fmt.Sprintf("%s %s", t.IATA, t.Time.Format(LayoutShort)))
			} else {
				parts = append(parts, fmt.Sprintf("%s ??:??", t.IATA))
			}
		}
		return strings.Join(parts, " ")
	}

	// Check if dates differ to auto-show dates
	allResults := append([]TimeResult{c.Source}, c.Targets...)
	showDate := datesDiffer(allResults)

	var sb strings.Builder

	// Format source
	emoji := ClockEmoji(c.Source.Time)
	if showDate {
		sb.WriteString(fmt.Sprintf("%s: %s %s %s", c.Source.IATA, emoji, c.Source.Time.Format(LayoutShort), c.Source.Time.Format(LayoutDate)))
	} else {
		sb.WriteString(fmt.Sprintf("%s: %s %s", c.Source.IATA, emoji, c.Source.Time.Format(LayoutShort)))
	}

	sb.WriteString("  →  ")

	// Format targets
	var targetParts []string
	for _, t := range c.Targets {
		if t.Found {
			tEmoji := ClockEmoji(t.Time)
			if showDate {
				targetParts = append(targetParts, fmt.Sprintf("%s: %s %s %s", t.IATA, tEmoji, t.Time.Format(LayoutShort), t.Time.Format(LayoutDate)))
			} else {
				targetParts = append(targetParts, fmt.Sprintf("%s: %s %s", t.IATA, tEmoji, t.Time.Format(LayoutShort)))
			}
		} else {
			targetParts = append(targetParts, fmt.Sprintf("%s: ??:??", t.IATA))
		}
	}
	sb.WriteString(strings.Join(targetParts, ", "))
	sb.WriteString("\n")

	return sb.String()
}

// ShowConversion displays a time conversion from a source location to multiple targets.
// sourceSpec is a time specification like "SFO@9:00".
// targets are IATA codes to convert to.
func ShowConversion(w io.Writer, sourceSpec TimeSpec, targets []string, ps1Format bool, now *time.Time) {
	var refTime time.Time
	if now != nil {
		refTime = *now
	} else {
		refTime = time.Now()
	}

	sourceTime, err := sourceSpec.ResolveTime(refTime)
	if err != nil {
		_, _ = fmt.Fprintf(w, "%s: Unknown airport code\n", sourceSpec.IATA)
		return
	}

	sourceResult := TimeResult{
		IATA:     sourceSpec.IATA,
		Time:     sourceTime,
		Location: codes.IATA[sourceSpec.IATA],
		Found:    true,
	}

	var targetResults []TimeResult
	for _, target := range targets {
		targetResults = append(targetResults, LookupTime(target, &sourceTime))
	}

	result := &ConversionResult{
		Source:  sourceResult,
		Targets: targetResults,
	}

	_, _ = fmt.Fprint(w, FormatConversion(result, ps1Format))
}
