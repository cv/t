package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cv/t/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// captureStdout captures stdout during execution of f and returns the output.
func captureStdout(t *testing.T, f func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	f()

	require.NoError(t, w.Close())
	os.Stdout = old

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)
	return buf.String()
}

// setEnv sets an environment variable and returns a cleanup function.
func setEnv(t *testing.T, key, value string) {
	t.Helper()
	orig := os.Getenv(key)
	require.NoError(t, os.Setenv(key, value))
	t.Cleanup(func() {
		_ = os.Setenv(key, orig)
	})
}

func TestRun_NoArgs(t *testing.T) {
	code := run([]string{})
	assert.Equal(t, 1, code)
}

func TestRun_Version(t *testing.T) {
	var code int
	output := captureStdout(t, func() {
		code = run([]string{"-v"})
	})

	assert.Equal(t, 0, code)
	assert.Contains(t, output, "t dev")
}

func TestRun_VersionLong(t *testing.T) {
	var code int
	_ = captureStdout(t, func() {
		code = run([]string{"--version"})
	})

	assert.Equal(t, 0, code)
}

func TestRun_BasicIATA(t *testing.T) {
	var code int
	output := captureStdout(t, func() {
		code = run([]string{"sfo"})
	})

	assert.Equal(t, 0, code)
	assert.Contains(t, output, "SFO:")
	assert.Contains(t, output, "America/Los_Angeles")
}

func TestRun_MultipleIATA(t *testing.T) {
	var code int
	output := captureStdout(t, func() {
		code = run([]string{"sfo", "jfk"})
	})

	assert.Equal(t, 0, code)
	assert.Contains(t, output, "SFO:")
	assert.Contains(t, output, "JFK:")
}

func TestRun_DateFlag(t *testing.T) {
	var code int
	output := captureStdout(t, func() {
		code = run([]string{"-d", "sfo"})
	})

	assert.Equal(t, 0, code)
	assert.Contains(t, output, "SFO:")
}

func TestRun_SaveListDelete(t *testing.T) {
	tmpDir := t.TempDir()
	setEnv(t, "HOME", tmpDir)

	// Test --save
	var code int
	output := captureStdout(t, func() {
		code = run([]string{"--save", "testteam", "sfo", "jfk", "lon"})
	})
	assert.Equal(t, 0, code)
	assert.Contains(t, output, "Saved alias 'testteam'")

	// Test --list
	output = captureStdout(t, func() {
		code = run([]string{"--list"})
	})
	assert.Equal(t, 0, code)
	assert.Contains(t, output, "testteam: SFO JFK LON")

	// Test @alias
	output = captureStdout(t, func() {
		code = run([]string{"@testteam"})
	})
	assert.Equal(t, 0, code)
	assert.Contains(t, output, "SFO:")
	assert.Contains(t, output, "JFK:")
	assert.Contains(t, output, "LON:")

	// Test --delete
	output = captureStdout(t, func() {
		code = run([]string{"--delete", "testteam"})
	})
	assert.Equal(t, 0, code)
	assert.Contains(t, output, "Deleted alias 'testteam'")

	// Verify deleted with --list
	output = captureStdout(t, func() {
		code = run([]string{"--list"})
	})
	assert.Equal(t, 0, code)
	assert.Contains(t, output, "No aliases saved")
}

func TestRun_SaveMissingArgs(t *testing.T) {
	code := run([]string{"--save"})
	assert.Equal(t, 1, code)

	code = run([]string{"--save", "name"})
	assert.Equal(t, 1, code)
}

func TestRun_DeleteMissingArg(t *testing.T) {
	code := run([]string{"--delete"})
	assert.Equal(t, 1, code)
}

func TestRun_DeleteNonexistent(t *testing.T) {
	tmpDir := t.TempDir()
	setEnv(t, "HOME", tmpDir)

	code := run([]string{"--delete", "nonexistent"})
	assert.Equal(t, 1, code)
}

func TestRun_UnknownAlias(t *testing.T) {
	tmpDir := t.TempDir()
	setEnv(t, "HOME", tmpDir)

	code := run([]string{"@unknownalias"})
	assert.Equal(t, 1, code)
}

func TestRun_AliasWithOverlap(t *testing.T) {
	tmpDir := t.TempDir()
	setEnv(t, "HOME", tmpDir)

	// Save an alias first
	configDir := filepath.Join(tmpDir, ".config", "t")
	require.NoError(t, os.MkdirAll(configDir, 0o755))
	store, err := config.NewAliasStoreWithPath(filepath.Join(configDir, "aliases.json"))
	require.NoError(t, err)
	require.NoError(t, store.Save("team", []string{"sfo", "jfk"}))

	var code int
	output := captureStdout(t, func() {
		code = run([]string{"--overlap", "@team"})
	})

	assert.Equal(t, 0, code)
	assert.Contains(t, output, "Working hours overlap")
}

func TestRun_MixedAliasAndIATA(t *testing.T) {
	tmpDir := t.TempDir()
	setEnv(t, "HOME", tmpDir)

	// Save an alias
	configDir := filepath.Join(tmpDir, ".config", "t")
	require.NoError(t, os.MkdirAll(configDir, 0o755))
	store, err := config.NewAliasStoreWithPath(filepath.Join(configDir, "aliases.json"))
	require.NoError(t, err)
	require.NoError(t, store.Save("west", []string{"sfo", "lax"}))

	var code int
	output := captureStdout(t, func() {
		code = run([]string{"@west", "jfk"})
	})

	assert.Equal(t, 0, code)
	assert.Contains(t, output, "SFO:")
	assert.Contains(t, output, "LAX:")
	assert.Contains(t, output, "JFK:")
}

func TestRun_TimeConversion(t *testing.T) {
	var code int
	output := captureStdout(t, func() {
		code = run([]string{"sfo@9:00", "jfk"})
	})

	assert.Equal(t, 0, code)
	assert.Contains(t, output, "SFO:")
	assert.Contains(t, output, "→")
	assert.Contains(t, output, "JFK:")
}

func TestRun_TimeConversionMissingTarget(t *testing.T) {
	code := run([]string{"sfo@9:00"})
	assert.Equal(t, 1, code)
}

func TestRun_Overlap(t *testing.T) {
	var code int
	output := captureStdout(t, func() {
		code = run([]string{"--overlap", "sfo", "jfk"})
	})

	assert.Equal(t, 0, code)
	assert.Contains(t, output, "Working hours overlap")
}

func TestRun_OverlapMissingLocations(t *testing.T) {
	code := run([]string{"--overlap", "sfo"})
	assert.Equal(t, 1, code)
}

func TestRun_OverlapWithHours(t *testing.T) {
	var code int
	output := captureStdout(t, func() {
		code = run([]string{"--overlap", "--hours=8-18", "sfo", "jfk"})
	})

	assert.Equal(t, 0, code)
	assert.Contains(t, output, "8:00-18:00")
}

func TestRun_InvalidHoursFormat(t *testing.T) {
	code := run([]string{"--hours=invalid", "sfo", "jfk"})
	assert.Equal(t, 1, code)
}

func TestExpandAliases(t *testing.T) {
	tmpDir := t.TempDir()
	setEnv(t, "HOME", tmpDir)

	// Create an alias
	configDir := filepath.Join(tmpDir, ".config", "t")
	require.NoError(t, os.MkdirAll(configDir, 0o755))
	store, err := config.NewAliasStoreWithPath(filepath.Join(configDir, "aliases.json"))
	require.NoError(t, err)
	require.NoError(t, store.Save("team", []string{"sfo", "jfk", "lon"}))

	tests := []struct {
		name     string
		args     []string
		expected []string
		wantErr  bool
	}{
		{
			name:     "no aliases",
			args:     []string{"sfo", "jfk"},
			expected: []string{"sfo", "jfk"},
			wantErr:  false,
		},
		{
			name:     "single alias",
			args:     []string{"@team"},
			expected: []string{"SFO", "JFK", "LON"},
			wantErr:  false,
		},
		{
			name:     "alias with other codes",
			args:     []string{"@team", "nrt"},
			expected: []string{"SFO", "JFK", "LON", "nrt"},
			wantErr:  false,
		},
		{
			name:    "unknown alias",
			args:    []string{"@unknown"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := expandAliases(tt.args)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestRun_EmptyArgsAfterFlags(t *testing.T) {
	code := run([]string{"-d"})
	assert.Equal(t, 1, code)

	code = run([]string{"--date"})
	assert.Equal(t, 1, code)
}

func TestRun_PS1Format(t *testing.T) {
	setEnv(t, "PS1_FORMAT", "1")

	var code int
	output := captureStdout(t, func() {
		code = run([]string{"sfo", "jfk"})
	})

	assert.Equal(t, 0, code)
	// PS1 format is compact, no newlines between entries
	lines := strings.Split(strings.TrimSpace(output), "\n")
	assert.Equal(t, 1, len(lines))
	assert.Contains(t, output, "SFO")
	assert.Contains(t, output, "JFK")
}
