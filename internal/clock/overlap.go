package clock

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/cv/t/codes"
)

// WorkHours represents a working hours range.
type WorkHours struct {
	Start int // Start hour (0-23)
	End   int // End hour (0-23, exclusive)
}

// DefaultWorkHours is the default 9am-5pm work day.
var DefaultWorkHours = WorkHours{Start: 9, End: 17}

// ParseWorkHours parses a work hours string like "8:00-18:00" or "9-17".
// Returns nil if the string is not a valid work hours spec.
func ParseWorkHours(s string) *WorkHours {
	// Try parsing with minutes first: "8:00-18:00"
	var startH, startM, endH, endM int
	if n, _ := fmt.Sscanf(s, "%d:%d-%d:%d", &startH, &startM, &endH, &endM); n == 4 {
		if startH < 0 || startH > 23 || endH < 0 || endH > 24 {
			return nil
		}
		if startM < 0 || startM > 59 || endM < 0 || endM > 59 {
			return nil
		}
		// For simplicity, we only use hours (ignore minutes)
		return &WorkHours{Start: startH, End: endH}
	}

	// Try parsing hours only: "9-17"
	if n, _ := fmt.Sscanf(s, "%d-%d", &startH, &endH); n == 2 {
		if startH < 0 || startH > 23 || endH < 0 || endH > 24 {
			return nil
		}
		return &WorkHours{Start: startH, End: endH}
	}

	return nil
}

// LocationInfo holds timezone information for a location.
type LocationInfo struct {
	IATA     string
	Location *time.Location
	LocName  string
	Offset   int // UTC offset in seconds
}

// OverlapResult holds the result of finding overlapping work hours.
type OverlapResult struct {
	Locations []LocationInfo
	// OverlapHours are the overlapping hours in UTC
	OverlapHoursUTC []int
	WorkHours       WorkHours
}

// FindOverlap finds overlapping work hours across multiple timezones.
// Returns the hours (in UTC) that fall within work hours for all locations.
func FindOverlap(iatas []string, workHours WorkHours, refTime time.Time) (*OverlapResult, error) {
	if len(iatas) < 2 {
		return nil, fmt.Errorf("need at least 2 locations to find overlap")
	}

	locations := make([]LocationInfo, 0, len(iatas))

	for _, iata := range iatas {
		iata = strings.ToUpper(iata)
		locName, found := codes.IATA[iata]
		if !found {
			return nil, fmt.Errorf("unknown IATA code: %s", iata)
		}

		loc, err := time.LoadLocation(locName)
		if err != nil {
			return nil, fmt.Errorf("loading location %s: %w", locName, err)
		}

		// Get offset at reference time
		_, offset := refTime.In(loc).Zone()

		locations = append(locations, LocationInfo{
			IATA:     iata,
			Location: loc,
			LocName:  locName,
			Offset:   offset,
		})
	}

	// Find overlapping hours
	// For each UTC hour, check if it falls within work hours for all locations
	var overlapHours []int
	for utcHour := 0; utcHour < 24; utcHour++ {
		isOverlap := true
		for _, loc := range locations {
			localHour := (utcHour + loc.Offset/3600) % 24
			if localHour < 0 {
				localHour += 24
			}
			if localHour < workHours.Start || localHour >= workHours.End {
				isOverlap = false
				break
			}
		}
		if isOverlap {
			overlapHours = append(overlapHours, utcHour)
		}
	}

	return &OverlapResult{
		Locations:       locations,
		OverlapHoursUTC: overlapHours,
		WorkHours:       workHours,
	}, nil
}

// FormatOverlap formats the overlap result for display.
func FormatOverlap(result *OverlapResult) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Working hours overlap (%d:00-%d:00 local):\n",
		result.WorkHours.Start, result.WorkHours.End))

	if len(result.OverlapHoursUTC) == 0 {
		sb.WriteString("  No overlapping hours found\n")
		return sb.String()
	}

	// Group consecutive hours
	ranges := groupConsecutiveHours(result.OverlapHoursUTC)

	for _, r := range ranges {
		sb.WriteString("  ")
		// Show the range in each timezone
		var parts []string
		for _, loc := range result.Locations {
			startLocal := (r.start + loc.Offset/3600) % 24
			if startLocal < 0 {
				startLocal += 24
			}
			endLocal := (r.end + loc.Offset/3600) % 24
			if endLocal < 0 {
				endLocal += 24
			}
			parts = append(parts, fmt.Sprintf("%02d:00-%02d:00 %s", startLocal, endLocal, loc.IATA))
		}
		sb.WriteString(strings.Join(parts, " = "))
		sb.WriteString("\n")
	}

	// Show total overlap
	hours := len(result.OverlapHoursUTC)
	if hours == 1 {
		sb.WriteString("  (1 hour overlap)\n")
	} else {
		sb.WriteString(fmt.Sprintf("  (%d hours overlap)\n", hours))
	}

	return sb.String()
}

// hourRange represents a range of consecutive hours.
type hourRange struct {
	start int // inclusive
	end   int // exclusive (the hour after the last hour in range)
}

// groupConsecutiveHours groups consecutive UTC hours into ranges.
func groupConsecutiveHours(hours []int) []hourRange {
	if len(hours) == 0 {
		return nil
	}

	var ranges []hourRange
	start := hours[0]
	prev := hours[0]

	for i := 1; i < len(hours); i++ {
		if hours[i] != prev+1 {
			// End current range
			ranges = append(ranges, hourRange{start: start, end: prev + 1})
			start = hours[i]
		}
		prev = hours[i]
	}
	// Add final range
	ranges = append(ranges, hourRange{start: start, end: prev + 1})

	return ranges
}

// ShowOverlap displays overlapping work hours across timezones.
func ShowOverlap(w io.Writer, iatas []string, workHours WorkHours, now *time.Time) {
	var refTime time.Time
	if now != nil {
		refTime = *now
	} else {
		refTime = time.Now()
	}

	result, err := FindOverlap(iatas, workHours, refTime)
	if err != nil {
		_, _ = fmt.Fprintf(w, "Error: %v\n", err)
		return
	}

	_, _ = fmt.Fprint(w, FormatOverlap(result))
}
