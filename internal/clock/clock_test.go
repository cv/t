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

func TestRelativeOffset(t *testing.T) {
	// Load fixed locations for deterministic tests
	tokyo, err := time.LoadLocation("Asia/Tokyo")
	require.NoError(t, err)
	london, err := time.LoadLocation("Europe/London")
	require.NoError(t, err)
	newYork, err := time.LoadLocation("America/New_York")
	require.NoError(t, err)
	kolkata, err := time.LoadLocation("Asia/Kolkata") // UTC+5:30
	require.NoError(t, err)

	// Use a fixed base time in winter (no DST complications)
	baseTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		targetLoc  *time.Location
		localLoc   *time.Location
		wantOffset string
	}{
		{
			name:       "same timezone",
			targetLoc:  tokyo,
			localLoc:   tokyo,
			wantOffset: "(+0h)",
		},
		{
			name:       "Tokyo from London (winter)",
			targetLoc:  tokyo,
			localLoc:   london,
			wantOffset: "(+9h)", // Tokyo is UTC+9, London is UTC+0 in winter
		},
		{
			name:       "London from Tokyo (winter)",
			targetLoc:  london,
			localLoc:   tokyo,
			wantOffset: "(-9h)", // London is 9 hours behind Tokyo
		},
		{
			name:       "Tokyo from New York (winter)",
			targetLoc:  tokyo,
			localLoc:   newYork,
			wantOffset: "(+14h)", // Tokyo is UTC+9, NY is UTC-5 in winter
		},
		{
			name:       "New York from Tokyo (winter)",
			targetLoc:  newYork,
			localLoc:   tokyo,
			wantOffset: "(-14h)",
		},
		{
			name:       "Kolkata from London (half hour offset)",
			targetLoc:  kolkata,
			localLoc:   london,
			wantOffset: "(+5h30m)", // Kolkata is UTC+5:30
		},
		{
			name:       "London from Kolkata (negative half hour)",
			targetLoc:  london,
			localLoc:   kolkata,
			wantOffset: "(-5h30m)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Temporarily override time.Local for the test
			origLocal := time.Local
			time.Local = tt.localLoc
			defer func() { time.Local = origLocal }()

			targetTime := baseTime.In(tt.targetLoc)
			got := RelativeOffset(targetTime)

			assert.Equal(t, tt.wantOffset, got, "offset mismatch")
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
		showDate  bool
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
			showDate:  false,
			wantParts: []string{"SFO:", "America/Los_Angeles", "\n"},
		},
		{
			name: "full format with date",
			result: TimeResult{
				IATA:     "SFO",
				Time:     localTime,
				Location: "America/Los_Angeles",
				Found:    true,
			},
			ps1Format: false,
			showDate:  true,
			wantParts: []string{"SFO:", "Sat Jun 15", "America/Los_Angeles", "\n"},
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
			showDate:  false,
			wantParts: []string{"SFO "},
		},
		{
			name: "not found",
			result: TimeResult{
				IATA:  "XXX",
				Found: false,
			},
			ps1Format: false,
			showDate:  false,
			wantParts: []string{"XXX:", "??:??:??", "Unknown", "\n"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatResult(tt.result, tt.ps1Format, tt.showDate)

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

	got := FormatResult(result, false, false)

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

func TestFormatResultContainsOffset(t *testing.T) {
	// Set a fixed local timezone for deterministic testing
	london, err := time.LoadLocation("Europe/London")
	require.NoError(t, err)
	tokyo, err := time.LoadLocation("Asia/Tokyo")
	require.NoError(t, err)

	// Use winter time to avoid DST complications
	baseTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	tokyoTime := baseTime.In(tokyo)

	// Override time.Local to London for deterministic test
	origLocal := time.Local
	time.Local = london
	defer func() { time.Local = origLocal }()

	result := TimeResult{
		IATA:     "NRT",
		Time:     tokyoTime,
		Location: "Asia/Tokyo",
		Found:    true,
	}

	got := FormatResult(result, false, false)

	// Tokyo is UTC+9, London is UTC+0 in winter, so offset should be +9h
	assert.Contains(t, got, "(+9h)", "output should contain relative offset")
	assert.Contains(t, got, "NRT:", "output should contain IATA code")
	assert.Contains(t, got, "Asia/Tokyo", "output should contain location")
}

func TestFormatResultPS1NoOffset(t *testing.T) {
	// PS1 format should NOT include offset (keep it compact)
	fixedTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	result := TimeResult{
		IATA:     "SFO",
		Time:     fixedTime,
		Location: "America/Los_Angeles",
		Found:    true,
	}

	got := FormatResult(result, true, false)

	assert.NotContains(t, got, "(+", "ps1 format should not contain offset")
	assert.NotContains(t, got, "(-", "ps1 format should not contain offset")
}

func TestShow(t *testing.T) {
	fixedTime := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		iata      string
		ps1Format bool
		showDate  bool
		wantParts []string
	}{
		{
			name:      "full format",
			iata:      "SFO",
			ps1Format: false,
			showDate:  false,
			wantParts: []string{"SFO:", "America/Los_Angeles"},
		},
		{
			name:      "full format with date",
			iata:      "SFO",
			ps1Format: false,
			showDate:  true,
			wantParts: []string{"SFO:", "Sat Jun 15", "America/Los_Angeles"},
		},
		{
			name:      "ps1 format",
			iata:      "SFO",
			ps1Format: true,
			showDate:  false,
			wantParts: []string{"SFO "},
		},
		{
			name:      "unknown airport",
			iata:      "XXX",
			ps1Format: false,
			showDate:  false,
			wantParts: []string{"XXX:", "Unknown"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			Show(&buf, tt.iata, tt.ps1Format, tt.showDate, &fixedTime)
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
		showDate  bool
		wantParts []string
	}{
		{
			name:      "multiple airports full format",
			iatas:     []string{"SFO", "JFK"},
			ps1Format: false,
			showDate:  false,
			wantParts: []string{"SFO:", "JFK:", "\n"},
		},
		{
			name:      "multiple airports ps1 format",
			iatas:     []string{"SFO", "JFK"},
			ps1Format: true,
			showDate:  false,
			wantParts: []string{"SFO ", "JFK "},
		},
		{
			name:      "single airport",
			iatas:     []string{"SFO"},
			ps1Format: false,
			showDate:  false,
			wantParts: []string{"SFO:"},
		},
		{
			name:      "empty list",
			iatas:     []string{},
			ps1Format: false,
			showDate:  false,
			wantParts: []string{},
		},
		{
			name:      "with date flag",
			iatas:     []string{"SFO", "JFK"},
			ps1Format: false,
			showDate:  true,
			wantParts: []string{"SFO:", "JFK:", "Sat Jun 15", "\n"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			ShowAll(&buf, tt.iatas, tt.ps1Format, tt.showDate, &fixedTime)
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
	ShowAll(&buf, []string{"SFO", "JFK", "LHR"}, true, false, &fixedTime)
	got := buf.String()

	// Should have spaces between entries: "SFO HH:MM JFK HH:MM LHR HH:MM"
	parts := strings.Split(got, " ")
	assert.GreaterOrEqual(t, len(parts), 6, "ps1 output should have space-separated entries")
}

func TestShowAllAutoDateWhenDatesDiffer(t *testing.T) {
	// Time chosen so SFO (UTC-7) and NRT (UTC+9) are on different days
	// At UTC 2024-06-15 23:00, SFO is 16:00 Jun 15, NRT is 08:00 Jun 16
	fixedTime := time.Date(2024, 6, 15, 23, 0, 0, 0, time.UTC)

	var buf bytes.Buffer
	ShowAll(&buf, []string{"SFO", "NRT"}, false, false, &fixedTime)
	got := buf.String()

	// Should auto-show date because dates differ
	assert.Contains(t, got, "Sat Jun 15", "should show SFO date")
	assert.Contains(t, got, "Sun Jun 16", "should show NRT date")
}

func TestShowAllNoAutoDateWhenSameDate(t *testing.T) {
	// Time chosen so SFO and JFK are on the same day
	fixedTime := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)

	var buf bytes.Buffer
	ShowAll(&buf, []string{"SFO", "JFK"}, false, false, &fixedTime)
	got := buf.String()

	// Should NOT auto-show date because dates are the same
	assert.NotContains(t, got, "Jun 15", "should not show date when dates are same")
}

func TestDatesDiffer(t *testing.T) {
	loc1, _ := time.LoadLocation("America/Los_Angeles")
	loc2, _ := time.LoadLocation("Asia/Tokyo")

	// Same day scenario
	sameDay := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	resultsSameDay := []TimeResult{
		{IATA: "SFO", Time: sameDay.In(loc1), Found: true},
		{IATA: "JFK", Time: sameDay, Found: true},
	}
	assert.False(t, datesDiffer(resultsSameDay), "same day should return false")

	// Different day scenario
	diffDay := time.Date(2024, 6, 15, 23, 0, 0, 0, time.UTC)
	resultsDiffDay := []TimeResult{
		{IATA: "SFO", Time: diffDay.In(loc1), Found: true},
		{IATA: "NRT", Time: diffDay.In(loc2), Found: true},
	}
	assert.True(t, datesDiffer(resultsDiffDay), "different days should return true")

	// With unfound result
	resultsWithUnfound := []TimeResult{
		{IATA: "SFO", Time: sameDay.In(loc1), Found: true},
		{IATA: "XXX", Found: false},
	}
	assert.False(t, datesDiffer(resultsWithUnfound), "unfound results should be skipped")
}
