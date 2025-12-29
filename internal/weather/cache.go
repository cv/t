package weather

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// CacheTTL is how long cached weather data is valid.
const CacheTTL = 1 * time.Hour

// CacheEntry holds cached weather data with timestamp.
type CacheEntry struct {
	Condition string    `json:"condition"`
	Temp      string    `json:"temp"`
	FetchedAt time.Time `json:"fetched_at"`
}

// Cache provides file-based caching for weather data.
type Cache struct {
	dir string
	mu  sync.RWMutex
}

// NewCache creates a cache in the user's cache directory.
func NewCache() (*Cache, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		// Fall back to temp dir
		cacheDir = os.TempDir()
	}

	dir := filepath.Join(cacheDir, "t", "weather")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}

	return &Cache{dir: dir}, nil
}

// cacheFile returns the path to the cache file for an IATA code.
func (c *Cache) cacheFile(iata string) string {
	return filepath.Join(c.dir, strings.ToUpper(iata)+".json")
}

// Get retrieves cached weather info if it exists and is not expired.
func (c *Cache) Get(iata string) (Info, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	data, err := os.ReadFile(c.cacheFile(iata))
	if err != nil {
		return Info{}, false
	}

	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return Info{}, false
	}

	// Check if expired
	if time.Since(entry.FetchedAt) > CacheTTL {
		return Info{}, false
	}

	return Info{
		Condition: entry.Condition,
		Temp:      entry.Temp,
		Found:     true,
	}, true
}

// Set stores weather info in the cache.
func (c *Cache) Set(iata string, info Info) {
	if !info.Found {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	entry := CacheEntry{
		Condition: info.Condition,
		Temp:      info.Temp,
		FetchedAt: time.Now(),
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return
	}

	_ = os.WriteFile(c.cacheFile(iata), data, 0o644)
}

// CachedClient wraps a Fetcher with caching.
type CachedClient struct {
	fetcher Fetcher
	cache   *Cache
}

// NewCachedClient creates a new cached weather client.
func NewCachedClient() (*CachedClient, error) {
	cache, err := NewCache()
	if err != nil {
		return nil, err
	}

	return &CachedClient{
		fetcher: NewClient(),
		cache:   cache,
	}, nil
}

// Fetch gets weather info, using cache if available.
func (c *CachedClient) Fetch(ctx context.Context, iata string) Info {
	iata = strings.ToUpper(iata)

	// Check cache first
	if info, ok := c.cache.Get(iata); ok {
		return info
	}

	// Fetch from API
	info := c.fetcher.Fetch(ctx, iata)

	// Cache the result
	c.cache.Set(iata, info)

	return info
}
