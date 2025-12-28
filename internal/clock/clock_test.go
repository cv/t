package clock

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClockEmoji(t *testing.T) {
	tests := []struct {
		name    string
		hour    int
		minute  int
		wantLow bool // true if expecting low emoji (minute <= 30)
	}{
		{name: "midnight low", hour: 0, minute: 0, wantLow: true},
		{name: "midnight high", hour: 0, minute: 45, wantLow: false},
		{name: "noon low", hour: 12, minute: 15, wantLow: true},
		{name: "noon high", hour: 12, minute: 31, wantLow: false},
		{name: "3pm exactly 30", hour: 15, minute: 30, wantLow: true},
		{name: "3pm just after 30", hour: 15, minute: 31, wantLow: false},
		{name: "11pm low", hour: 23, minute: 10, wantLow: true},
		{name: "11pm high", hour: 23, minute: 55, wantLow: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testTime := time.Date(2024, 1, 1, tt.hour, tt.minute, 0, 0, time.UTC)
			got := ClockEmoji(testTime)

			assert.NotEmpty(t, got, "ClockEmoji() should return a non-empty string")

			// Verify it's from the correct set
			hour := tt.hour % 12
			if tt.wantLow {
				assert.Equal(t, clocksLow[hour], got, "should use low emoji set")
			} else {
				assert.Equal(t, clocksHigh[hour], got, "should use high emoji set")
			}
		})
	}
}

func TestLookupTime(t *testing.T) {
	fixedTime := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name         string
		iata         string
		wantFound    bool
		wantLocation string
	}{
		{
			name:         "SFO valid",
			iata:         "SFO",
			wantFound:    true,
			wantLocation: "America/Los_Angeles",
		},
		{
			name:         "sfo lowercase",
			iata:         "sfo",
			wantFound:    true,
			wantLocation: "America/Los_Angeles",
		},
		{
			name:         "JFK valid",
			iata:         "JFK",
			wantFound:    true,
			wantLocation: "America/New_York",
		},
		{
			name:         "LHR valid",
			iata:         "LHR",
			wantFound:    true,
			wantLocation: "Europe/London",
		},
		{
			name:      "XXX unknown",
			iata:      "XXX",
			wantFound: false,
		},
		{
			name:      "empty string",
			iata:      "",
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LookupTime(tt.iata, &fixedTime)

			assert.Equal(t, tt.wantFound, got.Found, "Found mismatch")

			if tt.wantFound {
				assert.Equal(t, tt.wantLocation, got.Location, "Location mismatch")
				assert.False(t, got.Time.IsZero(), "Time should not be zero")
			}

			// IATA should always be uppercased
			assert.True(t, strings.EqualFold(got.IATA, tt.iata), "IATA should be uppercased version of input")
		})
	}
}

func TestLookupTimeWithNilNow(t *testing.T) {
	result := LookupTime("SFO", nil)

	assert.True(t, result.Found, "should find SFO")
	assert.False(t, result.Time.IsZero(), "should return current time, not zero")
}

func TestFormatResult(t *testing.T) {
	fixedTime := time.Date(2024, 6, 15, 14, 30, 45, 0, time.UTC)
	loc, err := time.LoadLocation("America/Los_Angeles")
	require.NoError(t, err)
	localTime := fixedTime.In(loc)

	tests := []struct {
		name      string
		result    TimeResult
		ps1Format bool
		wantParts []string // parts that should be in the output
	}{
		{
			name: "full format found",
			result: TimeResult{
				IATA:     "SFO",
				Time:     localTime,
				Location: "America/Los_Angeles",
				Found:    true,
			},
			ps1Format: false,
			wantParts: []string{"SFO:", "America/Los_Angeles", "\n"},
		},
		{
			name: "ps1 format found",
			result: TimeResult{
				IATA:     "SFO",
				Time:     localTime,
				Location: "America/Los_Angeles",
				Found:    true,
			},
			ps1Format: true,
			wantParts: []string{"SFO "},
		},
		{
			name: "not found",
			result: TimeResult{
				IATA:  "XXX",
				Found: false,
			},
			ps1Format: false,
			wantParts: []string{"XXX:", "??:??:??", "Unknown", "\n"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatResult(tt.result, tt.ps1Format)

			for _, part := range tt.wantParts {
				assert.Contains(t, got, part, "output should contain expected part")
			}

			// PS1 format should not have newlines
			if tt.ps1Format && tt.result.Found {
				assert.NotContains(t, got, "\n", "ps1 format should not contain newline")
			}
		})
	}
}

func TestFormatResultContainsEmoji(t *testing.T) {
	fixedTime := time.Date(2024, 6, 15, 15, 0, 0, 0, time.UTC)
	result := TimeResult{
		IATA:     "SFO",
		Time:     fixedTime,
		Location: "America/Los_Angeles",
		Found:    true,
	}

	got := FormatResult(result, false)

	// Should contain a clock emoji
	hasEmoji := false
	for _, e := range clocksLow {
		if strings.Contains(got, e) {
			hasEmoji = true
			break
		}
	}
	if !hasEmoji {
		for _, e := range clocksHigh {
			if strings.Contains(got, e) {
				hasEmoji = true
				break
			}
		}
	}
	assert.True(t, hasEmoji, "output should contain a clock emoji")
}

func TestShow(t *testing.T) {
	fixedTime := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		iata      string
		ps1Format bool
		wantParts []string
	}{
		{
			name:      "full format",
			iata:      "SFO",
			ps1Format: false,
			wantParts: []string{"SFO:", "America/Los_Angeles"},
		},
		{
			name:      "ps1 format",
			iata:      "SFO",
			ps1Format: true,
			wantParts: []string{"SFO "},
		},
		{
			name:      "unknown airport",
			iata:      "XXX",
			ps1Format: false,
			wantParts: []string{"XXX:", "Unknown"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			Show(&buf, tt.iata, tt.ps1Format, &fixedTime)
			got := buf.String()

			for _, part := range tt.wantParts {
				assert.Contains(t, got, part, "output should contain expected part")
			}
		})
	}
}

func TestShowAll(t *testing.T) {
	fixedTime := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		iatas     []string
		ps1Format bool
		wantParts []string
	}{
		{
			name:      "multiple airports full format",
			iatas:     []string{"SFO", "JFK"},
			ps1Format: false,
			wantParts: []string{"SFO:", "JFK:", "\n"},
		},
		{
			name:      "multiple airports ps1 format",
			iatas:     []string{"SFO", "JFK"},
			ps1Format: true,
			wantParts: []string{"SFO ", "JFK "},
		},
		{
			name:      "single airport",
			iatas:     []string{"SFO"},
			ps1Format: false,
			wantParts: []string{"SFO:"},
		},
		{
			name:      "empty list",
			iatas:     []string{},
			ps1Format: false,
			wantParts: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			ShowAll(&buf, tt.iatas, tt.ps1Format, &fixedTime)
			got := buf.String()

			for _, part := range tt.wantParts {
				assert.Contains(t, got, part, "output should contain expected part")
			}
		})
	}
}

func TestShowAllPS1Spacing(t *testing.T) {
	fixedTime := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)

	var buf bytes.Buffer
	ShowAll(&buf, []string{"SFO", "JFK", "LHR"}, true, &fixedTime)
	got := buf.String()

	// Should have spaces between entries: "SFO HH:MM JFK HH:MM LHR HH:MM"
	parts := strings.Split(got, " ")
	assert.GreaterOrEqual(t, len(parts), 6, "ps1 output should have space-separated entries")
}
