package weather

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestParseResponse(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantFound bool
		wantCond  string
		wantTemp  string
	}{
		{
			name:      "valid sunny",
			input:     "☀️|+13°C",
			wantFound: true,
			wantCond:  "☀️",
			wantTemp:  "+13°C",
		},
		{
			name:      "valid cloudy with spaces",
			input:     "☁️   |  +1°C  ",
			wantFound: true,
			wantCond:  "☁️",
			wantTemp:  "+1°C",
		},
		{
			name:      "valid rainy negative temp",
			input:     "🌧️|-5°C",
			wantFound: true,
			wantCond:  "🌧️",
			wantTemp:  "-5°C",
		},
		{
			name:      "valid fahrenheit",
			input:     "⛅|55°F",
			wantFound: true,
			wantCond:  "⛅",
			wantTemp:  "55°F",
		},
		{
			name:      "unknown location",
			input:     "Unknown location",
			wantFound: false,
		},
		{
			name:      "error response",
			input:     "Error: something went wrong",
			wantFound: false,
		},
		{
			name:      "empty response",
			input:     "",
			wantFound: false,
		},
		{
			name:      "missing separator",
			input:     "☀️ +13°C",
			wantFound: false,
		},
		{
			name:      "empty condition",
			input:     "|+13°C",
			wantFound: false,
		},
		{
			name:      "empty temp",
			input:     "☀️|",
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseResponse(tt.input)
			if got.Found != tt.wantFound {
				t.Errorf("parseResponse(%q).Found = %v, want %v", tt.input, got.Found, tt.wantFound)
			}
			if tt.wantFound {
				if got.Condition != tt.wantCond {
					t.Errorf("parseResponse(%q).Condition = %q, want %q", tt.input, got.Condition, tt.wantCond)
				}
				if got.Temp != tt.wantTemp {
					t.Errorf("parseResponse(%q).Temp = %q, want %q", tt.input, got.Temp, tt.wantTemp)
				}
			}
		})
	}
}

func TestInfoFormat(t *testing.T) {
	tests := []struct {
		name string
		info Info
		want string
	}{
		{
			name: "found",
			info: Info{Condition: "☀️", Temp: "+13°C", Found: true},
			want: "☀️ +13°C",
		},
		{
			name: "not found",
			info: Info{Found: false},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.info.Format(); got != tt.want {
				t.Errorf("Info.Format() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestClientFetch(t *testing.T) {
	tests := []struct {
		name       string
		response   string
		statusCode int
		wantFound  bool
		wantCond   string
		wantTemp   string
	}{
		{
			name:       "success",
			response:   "☀️|+20°C",
			statusCode: http.StatusOK,
			wantFound:  true,
			wantCond:   "☀️",
			wantTemp:   "+20°C",
		},
		{
			name:       "server error",
			response:   "",
			statusCode: http.StatusInternalServerError,
			wantFound:  false,
		},
		{
			name:       "not found",
			response:   "Unknown location",
			statusCode: http.StatusOK,
			wantFound:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify the request format
				if !strings.Contains(r.URL.Path, "/SFO") {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				if !strings.Contains(r.URL.RawQuery, "format") {
					t.Errorf("missing format query: %s", r.URL.RawQuery)
				}

				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.response))
			}))
			defer server.Close()

			client := &Client{
				httpClient: &http.Client{Timeout: 5 * time.Second},
				baseURL:    server.URL,
			}

			ctx := context.Background()
			got := client.Fetch(ctx, "SFO")

			if got.Found != tt.wantFound {
				t.Errorf("Fetch().Found = %v, want %v", got.Found, tt.wantFound)
			}
			if tt.wantFound {
				if got.Condition != tt.wantCond {
					t.Errorf("Fetch().Condition = %q, want %q", got.Condition, tt.wantCond)
				}
				if got.Temp != tt.wantTemp {
					t.Errorf("Fetch().Temp = %q, want %q", got.Temp, tt.wantTemp)
				}
			}
		})
	}
}

func TestClientFetchTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("☀️|+20°C"))
	}))
	defer server.Close()

	client := &Client{
		httpClient: &http.Client{Timeout: 50 * time.Millisecond},
		baseURL:    server.URL,
	}

	ctx := context.Background()
	got := client.Fetch(ctx, "SFO")

	if got.Found {
		t.Error("expected timeout to cause Found=false")
	}
}

// mockFetcher implements Fetcher for testing.
type mockFetcher struct {
	results map[string]Info
}

func (m *mockFetcher) Fetch(_ context.Context, iata string) Info {
	if info, ok := m.results[strings.ToUpper(iata)]; ok {
		return info
	}
	return Info{Found: false}
}

func TestFetchAll(t *testing.T) {
	mock := &mockFetcher{
		results: map[string]Info{
			"SFO": {Condition: "☀️", Temp: "+20°C", Found: true},
			"JFK": {Condition: "🌧️", Temp: "+5°C", Found: true},
			"LON": {Found: false},
		},
	}

	ctx := context.Background()
	results := FetchAll(ctx, mock, []string{"sfo", "jfk", "lon"})

	if len(results) != 3 {
		t.Errorf("FetchAll returned %d results, want 3", len(results))
	}

	if !results["SFO"].Found || results["SFO"].Temp != "+20°C" {
		t.Errorf("SFO result incorrect: %+v", results["SFO"])
	}

	if !results["JFK"].Found || results["JFK"].Condition != "🌧️" {
		t.Errorf("JFK result incorrect: %+v", results["JFK"])
	}

	if results["LON"].Found {
		t.Errorf("LON should not be found: %+v", results["LON"])
	}
}
