package clock

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindDSTTransition_NoDST(t *testing.T) {
	// UTC has no DST
	utcTime := time.Date(2024, 3, 10, 12, 0, 0, 0, time.UTC)
	result := FindDSTTransition(utcTime, 5)
	assert.Nil(t, result, "UTC should have no DST transitions")
}

func TestFindDSTTransition_NoTransitionInWindow(t *testing.T) {
	// Pick a date far from any DST transition (mid-summer)
	loc, err := time.LoadLocation("America/Los_Angeles")
	require.NoError(t, err)

	midSummer := time.Date(2024, 7, 15, 12, 0, 0, 0, loc)
	result := FindDSTTransition(midSummer, 5)
	assert.Nil(t, result, "mid-summer should have no DST transitions within 5 days")
}

func TestFindDSTTransition_SpringForward(t *testing.T) {
	// US DST starts second Sunday in March - in 2024 that's March 10
	loc, err := time.LoadLocation("America/Los_Angeles")
	require.NoError(t, err)

	// 3 days before DST starts
	beforeDST := time.Date(2024, 3, 7, 12, 0, 0, 0, loc)
	result := FindDSTTransition(beforeDST, 5)

	require.NotNil(t, result, "should find upcoming DST transition")
	assert.Equal(t, "DST starts", result.Description)
	assert.Equal(t, "+1h", result.OffsetChange)
	assert.True(t, result.DaysUntil > 0, "transition should be in the future")
	assert.True(t, result.DaysUntil <= 5, "transition should be within window")
}

func TestFindDSTTransition_FallBack(t *testing.T) {
	// US DST ends first Sunday in November - in 2024 that's November 3
	loc, err := time.LoadLocation("America/Los_Angeles")
	require.NoError(t, err)

	// 2 days before DST ends
	beforeDSTEnd := time.Date(2024, 11, 1, 12, 0, 0, 0, loc)
	result := FindDSTTransition(beforeDSTEnd, 5)

	require.NotNil(t, result, "should find upcoming DST transition")
	assert.Equal(t, "DST ends", result.Description)
	assert.Equal(t, "-1h", result.OffsetChange)
	assert.True(t, result.DaysUntil > 0, "transition should be in the future")
}

func TestFindDSTTransition_EuropeanDST(t *testing.T) {
	// European DST starts last Sunday in March - in 2024 that's March 31
	loc, err := time.LoadLocation("Europe/London")
	require.NoError(t, err)

	// 3 days before DST starts
	beforeDST := time.Date(2024, 3, 28, 12, 0, 0, 0, loc)
	result := FindDSTTransition(beforeDST, 5)

	require.NotNil(t, result, "should find upcoming DST transition")
	assert.Equal(t, "DST starts", result.Description)
	assert.Equal(t, "+1h", result.OffsetChange)
}

func TestFindDSTTransition_RecentlyPassed(t *testing.T) {
	// Test that we can detect recently passed transitions
	loc, err := time.LoadLocation("America/Los_Angeles")
	require.NoError(t, err)

	// 2 days after DST started (March 10, 2024)
	afterDST := time.Date(2024, 3, 12, 12, 0, 0, 0, loc)
	result := FindDSTTransition(afterDST, 5)

	require.NotNil(t, result, "should find recent DST transition")
	assert.Equal(t, "DST starts", result.Description)
	assert.True(t, result.DaysUntil < 0, "transition should be in the past")
}

func TestFindDSTTransition_NoTransitionZone(t *testing.T) {
	// Arizona doesn't observe DST (except Navajo Nation)
	loc, err := time.LoadLocation("America/Phoenix")
	require.NoError(t, err)

	// Test around when other US states would change
	nearDST := time.Date(2024, 3, 10, 12, 0, 0, 0, loc)
	result := FindDSTTransition(nearDST, 5)
	assert.Nil(t, result, "Phoenix should have no DST transitions")
}

func TestFormatOffsetChange(t *testing.T) {
	tests := []struct {
		name        string
		diffSeconds int
		want        string
	}{
		{name: "plus one hour", diffSeconds: 3600, want: "+1h"},
		{name: "minus one hour", diffSeconds: -3600, want: "-1h"},
		{name: "plus 30 minutes", diffSeconds: 1800, want: "+0h30m"},
		{name: "minus 30 minutes", diffSeconds: -1800, want: "-0h30m"},
		{name: "zero", diffSeconds: 0, want: "+0h"},
		{name: "plus 90 minutes", diffSeconds: 5400, want: "+1h30m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatOffsetChange(tt.diffSeconds)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFormatDSTWarning(t *testing.T) {
	tests := []struct {
		name       string
		transition *DSTTransition
		wantParts  []string
	}{
		{
			name:       "nil transition",
			transition: nil,
			wantParts:  []string{},
		},
		{
			name: "today",
			transition: &DSTTransition{
				DaysUntil:    0,
				OffsetChange: "+1h",
				Description:  "DST starts",
			},
			wantParts: []string{"⚠️", "DST starts", "today", "+1h"},
		},
		{
			name: "in 1 day",
			transition: &DSTTransition{
				DaysUntil:    1,
				OffsetChange: "+1h",
				Description:  "DST starts",
			},
			wantParts: []string{"⚠️", "DST starts", "in 1 day", "+1h"},
		},
		{
			name: "in 3 days",
			transition: &DSTTransition{
				DaysUntil:    3,
				OffsetChange: "-1h",
				Description:  "DST ends",
			},
			wantParts: []string{"⚠️", "DST ends", "in 3 days", "-1h"},
		},
		{
			name: "1 day ago",
			transition: &DSTTransition{
				DaysUntil:    -1,
				OffsetChange: "+1h",
				Description:  "DST starts",
			},
			wantParts: []string{"⚠️", "DST starts", "1 day ago", "+1h"},
		},
		{
			name: "2 days ago",
			transition: &DSTTransition{
				DaysUntil:    -2,
				OffsetChange: "-1h",
				Description:  "DST ends",
			},
			wantParts: []string{"⚠️", "DST ends", "2 days ago", "-1h"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatDSTWarning(tt.transition)
			if tt.transition == nil {
				assert.Empty(t, got)
				return
			}
			for _, part := range tt.wantParts {
				assert.Contains(t, got, part)
			}
		})
	}
}
