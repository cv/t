// Package weather provides weather information for locations using wttr.in.
package weather

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// DefaultTimeout is the default HTTP timeout for weather requests.
const DefaultTimeout = 10 * time.Second

// Info holds weather information for a location.
type Info struct {
	Condition string // Weather condition emoji
	Temp      string // Temperature with unit
	Found     bool   // Whether weather was successfully fetched
}

// Fetcher fetches weather information.
type Fetcher interface {
	Fetch(ctx context.Context, iata string) Info
}

// Client fetches weather from wttr.in.
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a new weather client.
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
		baseURL: "https://wttr.in",
	}
}

// Fetch gets weather information for an IATA airport code.
func (c *Client) Fetch(ctx context.Context, iata string) Info {
	// wttr.in format: %c = condition emoji, %t = temperature
	url := fmt.Sprintf("%s/%s?format=%%c|%%t", c.baseURL, strings.ToUpper(iata))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return Info{Found: false}
	}

	// Set User-Agent to avoid getting HTML response
	req.Header.Set("User-Agent", "curl/7.64.1")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Info{Found: false}
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return Info{Found: false}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Info{Found: false}
	}

	return parseResponse(string(body))
}

// parseResponse parses wttr.in response in format "emoji|temp".
func parseResponse(s string) Info {
	s = strings.TrimSpace(s)

	// Check for error responses
	if strings.Contains(s, "Unknown location") || strings.Contains(s, "Error") {
		return Info{Found: false}
	}

	parts := strings.Split(s, "|")
	if len(parts) != 2 {
		return Info{Found: false}
	}

	condition := strings.TrimSpace(parts[0])
	temp := strings.TrimSpace(parts[1])

	// Validate we got something reasonable
	if condition == "" || temp == "" {
		return Info{Found: false}
	}

	return Info{
		Condition: condition,
		Temp:      temp,
		Found:     true,
	}
}

// Format returns a formatted string of the weather info.
func (i Info) Format() string {
	if !i.Found {
		return ""
	}
	return fmt.Sprintf("%s %s", i.Condition, i.Temp)
}

// FetchAll fetches weather for multiple IATA codes concurrently.
// Returns a map of IATA code to weather info.
func FetchAll(ctx context.Context, fetcher Fetcher, iatas []string) map[string]Info {
	results := make(map[string]Info)
	ch := make(chan struct {
		iata string
		info Info
	}, len(iatas))

	for _, iata := range iatas {
		go func(code string) {
			info := fetcher.Fetch(ctx, code)
			ch <- struct {
				iata string
				info Info
			}{code, info}
		}(strings.ToUpper(iata))
	}

	for range iatas {
		result := <-ch
		results[result.iata] = result.info
	}

	return results
}
