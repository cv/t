package codes

import "testing"

func TestIsValidIATA(t *testing.T) {
	tests := []struct {
		code string
		want bool
	}{
		// Valid codes
		{"SFO", true},
		{"JFK", true},
		{"LAX", true},
		{"A1B", true},
		{"Z99", true},
		{"QKL", true}, // Train station (Cologne)

		// Invalid: wrong length
		{"", false},
		{"A", false},
		{"AB", false},
		{"ABCD", false},
		{"ABCDE", false},

		// Invalid: doesn't start with letter
		{"123", false},
		{"1AB", false},
		{"9ZZ", false},
		{"0G6", false},

		// Invalid: non-ASCII
		{"ИКУ", false}, // Cyrillic
		{"%u0", false}, // URL-encoded garbage

		// Invalid: contains non-alphanumeric
		{"A-B", false},
		{"A B", false},
		{"A.B", false},
		{"AB!", false},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			if got := IsValidIATA(tt.code); got != tt.want {
				t.Errorf("IsValidIATA(%q) = %v, want %v", tt.code, got, tt.want)
			}
		})
	}
}

func TestIsASCIILetter(t *testing.T) {
	tests := []struct {
		c    byte
		want bool
	}{
		{'A', true},
		{'Z', true},
		{'a', true},
		{'z', true},
		{'M', true},
		{'0', false},
		{'9', false},
		{' ', false},
		{'-', false},
	}

	for _, tt := range tests {
		t.Run(string(tt.c), func(t *testing.T) {
			if got := isASCIILetter(tt.c); got != tt.want {
				t.Errorf("isASCIILetter(%q) = %v, want %v", tt.c, got, tt.want)
			}
		})
	}
}

func TestIsASCIIAlphanumeric(t *testing.T) {
	tests := []struct {
		c    byte
		want bool
	}{
		{'A', true},
		{'Z', true},
		{'a', true},
		{'z', true},
		{'0', true},
		{'9', true},
		{'5', true},
		{' ', false},
		{'-', false},
		{'_', false},
		{'.', false},
	}

	for _, tt := range tests {
		t.Run(string(tt.c), func(t *testing.T) {
			if got := isASCIIAlphanumeric(tt.c); got != tt.want {
				t.Errorf("isASCIIAlphanumeric(%q) = %v, want %v", tt.c, got, tt.want)
			}
		})
	}
}

// TestAllIATACodesAreValid verifies that every code in the generated IATA map is valid.
func TestAllIATACodesAreValid(t *testing.T) {
	for code := range IATA {
		if !IsValidIATA(code) {
			t.Errorf("IATA map contains invalid code: %q", code)
		}
	}
}
