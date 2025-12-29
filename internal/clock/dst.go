package clock

import (
	"fmt"
	"time"
)

// DSTTransition represents an upcoming DST change.
type DSTTransition struct {
	// Date is when the transition occurs
	Date time.Time
	// DaysUntil is how many days until the transition (negative if in the past)
	DaysUntil int
	// OffsetChange is the change in offset (e.g., "+1h" for spring forward, "-1h" for fall back)
	OffsetChange string
	// Description is a human-readable description (e.g., "DST starts", "DST ends")
	Description string
}

// DefaultDSTWindow is the default number of days to look ahead/behind for DST changes.
const DefaultDSTWindow = 5

// FindDSTTransition checks if there's a DST transition within the given window of days
// for the specified time's location. Returns nil if no transition is found.
func FindDSTTransition(t time.Time, windowDays int) *DSTTransition {
	loc := t.Location()
	if loc == time.UTC {
		return nil // UTC has no DST
	}

	// Check each day in the window (both past and future)
	startDay := t.AddDate(0, 0, -windowDays)

	for day := 0; day <= windowDays*2; day++ {
		checkDay := startDay.AddDate(0, 0, day)

		// Check at the start of each hour on this day for a transition
		for hour := 0; hour < 24; hour++ {
			checkTime := time.Date(checkDay.Year(), checkDay.Month(), checkDay.Day(), hour, 0, 0, 0, loc)

			// Skip if this time is too far from our reference time
			hoursDiff := checkTime.Sub(t).Hours()
			if hoursDiff < float64(-windowDays*24) || hoursDiff > float64(windowDays*24) {
				continue
			}

			// Check if there's a transition at this hour by creating the previous hour explicitly.
			// We must use time.Date rather than checkTime.Add(-time.Hour) because Add()
			// preserves the same zone offset, hiding the transition.
			var prevHour time.Time
			if hour == 0 {
				prevDay := checkDay.AddDate(0, 0, -1)
				prevHour = time.Date(prevDay.Year(), prevDay.Month(), prevDay.Day(), 23, 0, 0, 0, loc)
			} else {
				prevHour = time.Date(checkDay.Year(), checkDay.Month(), checkDay.Day(), hour-1, 0, 0, 0, loc)
			}

			_, prevOffset := prevHour.Zone()
			_, checkOffset := checkTime.Zone()

			if prevOffset != checkOffset {
				// Found a transition!
				offsetDiff := checkOffset - prevOffset

				// Calculate days until transition
				var daysUntil int
				if checkTime.After(t) {
					hoursRemaining := checkTime.Sub(t).Hours()
					daysUntil = int(hoursRemaining / 24)
					if hoursRemaining > 0 && int(hoursRemaining)%24 > 0 {
						daysUntil++ // Round up for future transitions
					}
				} else {
					hoursAgo := t.Sub(checkTime).Hours()
					daysUntil = -int(hoursAgo / 24)
					if hoursAgo > 0 && int(hoursAgo)%24 > 0 {
						daysUntil-- // Round down (more negative) for past transitions
					}
				}

				// Only return if we're currently on the "currentOffset" side of the transition
				// and the transition is in the future, OR we just passed it
				if daysUntil >= -windowDays && daysUntil <= windowDays {
					return &DSTTransition{
						Date:         checkTime,
						DaysUntil:    daysUntil,
						OffsetChange: formatOffsetChange(offsetDiff),
						Description:  dstDescription(offsetDiff),
					}
				}
			}
		}
	}

	return nil
}

// formatOffsetChange formats an offset difference in seconds as a human-readable string.
func formatOffsetChange(diffSeconds int) string {
	if diffSeconds == 0 {
		return "+0h"
	}

	sign := "+"
	if diffSeconds < 0 {
		sign = "-"
		diffSeconds = -diffSeconds
	}

	hours := diffSeconds / 3600
	minutes := (diffSeconds % 3600) / 60

	if minutes == 0 {
		return fmt.Sprintf("%s%dh", sign, hours)
	}
	return fmt.Sprintf("%s%dh%dm", sign, hours, minutes)
}

// dstDescription returns a description of the DST transition.
func dstDescription(offsetDiff int) string {
	if offsetDiff > 0 {
		return "DST starts" // Clocks spring forward
	}
	return "DST ends" // Clocks fall back
}

// FormatDSTWarning formats a DST transition as a warning string.
// Returns empty string if transition is nil.
func FormatDSTWarning(transition *DSTTransition) string {
	if transition == nil {
		return ""
	}

	var daysStr string
	switch {
	case transition.DaysUntil == 0:
		daysStr = "today"
	case transition.DaysUntil == 1:
		daysStr = "in 1 day"
	case transition.DaysUntil == -1:
		daysStr = "1 day ago"
	case transition.DaysUntil > 0:
		daysStr = fmt.Sprintf("in %d days", transition.DaysUntil)
	default:
		daysStr = fmt.Sprintf("%d days ago", -transition.DaysUntil)
	}

	return fmt.Sprintf("⚠️ %s %s (%s)", transition.Description, daysStr, transition.OffsetChange)
}
