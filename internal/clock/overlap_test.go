package clock

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseWorkHours(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantNil   bool
		wantStart int
		wantEnd   int
	}{
		{
			name:      "full format with minutes",
			input:     "8:00-18:00",
			wantNil:   false,
			wantStart: 8,
			wantEnd:   18,
		},
		{
			name:      "hours only",
			input:     "9-17",
			wantNil:   false,
			wantStart: 9,
			wantEnd:   17,
		},
		{
			name:      "early start",
			input:     "6:00-14:00",
			wantNil:   false,
			wantStart: 6,
			wantEnd:   14,
		},
		{
			name:      "late end at 24",
			input:     "10-24",
			wantNil:   false,
			wantStart: 10,
			wantEnd:   24,
		},
		{
			name:    "invalid start hour",
			input:   "25-17",
			wantNil: true,
		},
		{
			name:    "invalid end hour",
			input:   "9-25",
			wantNil: true,
		},
		{
			name:    "negative start",
			input:   "-1-17",
			wantNil: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantNil: true,
		},
		{
			name:    "no dash",
			input:   "9:00",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseWorkHours(tt.input)
			if tt.wantNil {
				assert.Nil(t, got, "expected nil for input %q", tt.input)
				return
			}
			require.NotNil(t, got, "expected non-nil for input %q", tt.input)
			assert.Equal(t, tt.wantStart, got.Start, "Start mismatch")
			assert.Equal(t, tt.wantEnd, got.End, "End mismatch")
		})
	}
}

func TestFindOverlap(t *testing.T) {
	// Use a fixed reference time in winter to avoid DST complications
	refTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		iatas         []string
		workHours     WorkHours
		wantError     bool
		wantOverlap   int // number of overlapping hours
		errorContains string
	}{
		{
			name:        "SFO and JFK (3 hour difference)",
			iatas:       []string{"SFO", "JFK"},
			workHours:   DefaultWorkHours, // 9-17
			wantError:   false,
			wantOverlap: 5, // JFK 12:00-17:00 = SFO 9:00-14:00 (5 hours)
		},
		{
			name:        "SFO and LON (8 hour difference in winter)",
			iatas:       []string{"SFO", "LON"},
			workHours:   DefaultWorkHours, // 9-17
			wantError:   false,
			wantOverlap: 0, // No overlap with 8 hour diff and 8 hour workday
		},
		{
			name:        "SFO and LON with extended hours",
			iatas:       []string{"SFO", "LON"},
			workHours:   WorkHours{Start: 8, End: 18}, // 10 hour workday
			wantError:   false,
			wantOverlap: 2, // LON 16:00-18:00 = SFO 8:00-10:00
		},
		{
			name:        "three timezones SFO JFK LON",
			iatas:       []string{"SFO", "JFK", "LON"},
			workHours:   DefaultWorkHours,
			wantError:   false,
			wantOverlap: 0, // SFO 9am = JFK 12pm = LON 5pm (just at edge)
		},
		{
			name:          "single location error",
			iatas:         []string{"SFO"},
			workHours:     DefaultWorkHours,
			wantError:     true,
			errorContains: "at least 2 locations",
		},
		{
			name:          "unknown IATA code",
			iatas:         []string{"SFO", "XXX"},
			workHours:     DefaultWorkHours,
			wantError:     true,
			errorContains: "unknown IATA code",
		},
		{
			name:          "empty list",
			iatas:         []string{},
			workHours:     DefaultWorkHours,
			wantError:     true,
			errorContains: "at least 2 locations",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FindOverlap(tt.iatas, tt.workHours, refTime)
			if tt.wantError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantOverlap, len(got.OverlapHoursUTC),
				"expected %d overlapping hours, got %d: %v",
				tt.wantOverlap, len(got.OverlapHoursUTC), got.OverlapHoursUTC)
		})
	}
}

func TestFindOverlapLowercaseIATA(t *testing.T) {
	refTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	result, err := FindOverlap([]string{"sfo", "jfk"}, DefaultWorkHours, refTime)
	require.NoError(t, err)

	// Should normalize to uppercase
	assert.Equal(t, "SFO", result.Locations[0].IATA)
	assert.Equal(t, "JFK", result.Locations[1].IATA)
}

func TestGroupConsecutiveHours(t *testing.T) {
	tests := []struct {
		name  string
		hours []int
		want  []hourRange
	}{
		{
			name:  "single hour",
			hours: []int{9},
			want:  []hourRange{{start: 9, end: 10}},
		},
		{
			name:  "consecutive hours",
			hours: []int{9, 10, 11, 12},
			want:  []hourRange{{start: 9, end: 13}},
		},
		{
			name:  "two groups",
			hours: []int{9, 10, 14, 15, 16},
			want:  []hourRange{{start: 9, end: 11}, {start: 14, end: 17}},
		},
		{
			name:  "empty",
			hours: []int{},
			want:  nil,
		},
		{
			name:  "single isolated hours",
			hours: []int{9, 12, 15},
			want:  []hourRange{{start: 9, end: 10}, {start: 12, end: 13}, {start: 15, end: 16}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := groupConsecutiveHours(tt.hours)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFormatOverlap(t *testing.T) {
	sfoLoc, _ := time.LoadLocation("America/Los_Angeles")
	jfkLoc, _ := time.LoadLocation("America/New_York")

	// In winter: SFO is UTC-8, JFK is UTC-5
	tests := []struct {
		name      string
		result    *OverlapResult
		wantParts []string
	}{
		{
			name: "with overlap",
			result: &OverlapResult{
				Locations: []LocationInfo{
					{IATA: "SFO", Location: sfoLoc, LocName: "America/Los_Angeles", Offset: -8 * 3600},
					{IATA: "JFK", Location: jfkLoc, LocName: "America/New_York", Offset: -5 * 3600},
				},
				OverlapHoursUTC: []int{17, 18, 19, 20, 21},
				WorkHours:       DefaultWorkHours,
			},
			wantParts: []string{
				"Working hours overlap (9:00-17:00 local):",
				"SFO",
				"JFK",
				"5 hours overlap",
			},
		},
		{
			name: "no overlap",
			result: &OverlapResult{
				Locations: []LocationInfo{
					{IATA: "SFO", Location: sfoLoc, LocName: "America/Los_Angeles", Offset: -8 * 3600},
				},
				OverlapHoursUTC: []int{},
				WorkHours:       DefaultWorkHours,
			},
			wantParts: []string{
				"Working hours overlap (9:00-17:00 local):",
				"No overlapping hours found",
			},
		},
		{
			name: "single hour overlap",
			result: &OverlapResult{
				Locations: []LocationInfo{
					{IATA: "SFO", Location: sfoLoc, LocName: "America/Los_Angeles", Offset: -8 * 3600},
					{IATA: "JFK", Location: jfkLoc, LocName: "America/New_York", Offset: -5 * 3600},
				},
				OverlapHoursUTC: []int{17},
				WorkHours:       DefaultWorkHours,
			},
			wantParts: []string{
				"1 hour overlap",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatOverlap(tt.result)

			for _, part := range tt.wantParts {
				assert.Contains(t, got, part, "output should contain %q", part)
			}
		})
	}
}

func TestShowOverlap(t *testing.T) {
	refTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		iatas     []string
		workHours WorkHours
		wantParts []string
	}{
		{
			name:      "valid overlap",
			iatas:     []string{"SFO", "JFK"},
			workHours: DefaultWorkHours,
			wantParts: []string{"Working hours overlap", "SFO", "JFK"},
		},
		{
			name:      "unknown airport",
			iatas:     []string{"SFO", "XXX"},
			workHours: DefaultWorkHours,
			wantParts: []string{"Error:", "unknown IATA code"},
		},
		{
			name:      "too few locations",
			iatas:     []string{"SFO"},
			workHours: DefaultWorkHours,
			wantParts: []string{"Error:", "at least 2 locations"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			ShowOverlap(&buf, tt.iatas, tt.workHours, &refTime)
			got := buf.String()

			for _, part := range tt.wantParts {
				assert.Contains(t, got, part, "output should contain %q", part)
			}
		})
	}
}

func TestShowOverlapNilTime(t *testing.T) {
	var buf bytes.Buffer
	ShowOverlap(&buf, []string{"SFO", "JFK"}, DefaultWorkHours, nil)
	got := buf.String()

	// Should work with nil time (uses current time)
	assert.Contains(t, got, "Working hours overlap", "should produce valid output with nil time")
}

func TestOverlapSFOLONTYO(t *testing.T) {
	// Test the example from the issue: t --overlap sfo lon tyo
	// In winter:
	// - SFO: UTC-8 (America/Los_Angeles)
	// - LON: UTC+0 (Europe/London)
	// - TYO (NRT): UTC+9 (Asia/Tokyo)

	refTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	result, err := FindOverlap([]string{"SFO", "LON", "NRT"}, DefaultWorkHours, refTime)
	require.NoError(t, err)

	// For 9-17 work hours:
	// SFO 9-17 local = UTC 17-01 (next day)
	// LON 9-17 local = UTC 9-17
	// TYO 9-17 local = UTC 0-8
	// The only overlap is when all three are in 9-17 local time
	// This is impossible with default 9-17 hours across these zones

	// Let's calculate:
	// TYO 9am = UTC 0:00, TYO 5pm = UTC 8:00
	// LON 9am = UTC 9:00, LON 5pm = UTC 17:00
	// SFO 9am = UTC 17:00, SFO 5pm = UTC 01:00 (next day)
	// No UTC hour satisfies all three simultaneously with 9-17 work hours
	assert.Equal(t, 0, len(result.OverlapHoursUTC), "no overlap expected for SFO/LON/TYO with 9-17 hours")
}

func TestOverlapOutputFormat(t *testing.T) {
	refTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	var buf bytes.Buffer
	ShowOverlap(&buf, []string{"SFO", "JFK"}, DefaultWorkHours, &refTime)
	got := buf.String()

	// Verify the output format matches expected structure
	lines := strings.Split(strings.TrimSpace(got), "\n")
	assert.GreaterOrEqual(t, len(lines), 2, "should have at least 2 lines")
	assert.Contains(t, lines[0], "Working hours overlap", "first line should be header")
	assert.True(t, strings.HasPrefix(lines[1], "  "), "second line should be indented")
}

func TestFindOverlapWithDifferentWorkHours(t *testing.T) {
	refTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	// Extended work hours should give more overlap
	extended := WorkHours{Start: 7, End: 19} // 12 hour day

	result, err := FindOverlap([]string{"SFO", "LON"}, extended, refTime)
	require.NoError(t, err)

	// SFO 7am-7pm = UTC 15:00-03:00 (next day)
	// LON 7am-7pm = UTC 7:00-19:00
	// Overlap: UTC 15:00-19:00 = 4 hours
	assert.Equal(t, 4, len(result.OverlapHoursUTC),
		"expected 4 overlapping hours with extended work hours")
}
