package weather

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCacheGetSet(t *testing.T) {
	// Create a temp cache directory
	tmpDir := t.TempDir()
	cache := &Cache{dir: tmpDir}

	// Initially empty
	_, ok := cache.Get("SFO")
	if ok {
		t.Error("expected cache miss for empty cache")
	}

	// Set a value
	info := Info{Condition: "☀️", Temp: "+20°C", Found: true}
	cache.Set("SFO", info)

	// Should be retrievable
	got, ok := cache.Get("SFO")
	if !ok {
		t.Fatal("expected cache hit after Set")
	}
	if got.Condition != info.Condition {
		t.Errorf("Condition = %q, want %q", got.Condition, info.Condition)
	}
	if got.Temp != info.Temp {
		t.Errorf("Temp = %q, want %q", got.Temp, info.Temp)
	}
}

func TestCacheCaseInsensitive(t *testing.T) {
	tmpDir := t.TempDir()
	cache := &Cache{dir: tmpDir}

	info := Info{Condition: "☀️", Temp: "+20°C", Found: true}
	cache.Set("sfo", info)

	// Should be retrievable with different case
	got, ok := cache.Get("SFO")
	if !ok {
		t.Fatal("expected cache to be case-insensitive")
	}
	if got.Temp != info.Temp {
		t.Errorf("Temp = %q, want %q", got.Temp, info.Temp)
	}
}

func TestCacheDoesNotStoreNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	cache := &Cache{dir: tmpDir}

	// Set a not-found result
	cache.Set("XXX", Info{Found: false})

	// Should not be cached
	_, ok := cache.Get("XXX")
	if ok {
		t.Error("should not cache Found=false entries")
	}
}

func TestCacheExpiration(t *testing.T) {
	tmpDir := t.TempDir()
	cache := &Cache{dir: tmpDir}

	// Set a value
	info := Info{Condition: "☀️", Temp: "+20°C", Found: true}
	cache.Set("SFO", info)

	// Manually expire it by modifying the file
	cacheFile := filepath.Join(tmpDir, "SFO.json")
	data, _ := os.ReadFile(cacheFile)

	// Parse and modify timestamp
	var entry CacheEntry
	_ = json.Unmarshal(data, &entry)
	entry.FetchedAt = time.Now().Add(-2 * CacheTTL)
	data, _ = json.Marshal(entry)
	_ = os.WriteFile(cacheFile, data, 0o644)

	// Should be expired now
	_, ok := cache.Get("SFO")
	if ok {
		t.Error("expected cache miss for expired entry")
	}
}

func TestCachedClientUsesCacheOnHit(t *testing.T) {
	tmpDir := t.TempDir()
	cache := &Cache{dir: tmpDir}

	// Pre-populate cache
	cache.Set("SFO", Info{Condition: "🌧️", Temp: "+5°C", Found: true})

	// Create a mock fetcher that should NOT be called
	mock := &mockFetcher{
		results: map[string]Info{
			"SFO": {Condition: "☀️", Temp: "+20°C", Found: true},
		},
	}

	client := &CachedClient{
		fetcher: mock,
		cache:   cache,
	}

	got := client.Fetch(context.Background(), "SFO")

	// Should return cached value, not mock value
	if got.Condition != "🌧️" {
		t.Errorf("Condition = %q, want cached value 🌧️", got.Condition)
	}
	if got.Temp != "+5°C" {
		t.Errorf("Temp = %q, want cached value +5°C", got.Temp)
	}
}

func TestCachedClientFetchesOnMiss(t *testing.T) {
	tmpDir := t.TempDir()
	cache := &Cache{dir: tmpDir}

	mock := &mockFetcher{
		results: map[string]Info{
			"JFK": {Condition: "☁️", Temp: "+10°C", Found: true},
		},
	}

	client := &CachedClient{
		fetcher: mock,
		cache:   cache,
	}

	got := client.Fetch(context.Background(), "JFK")

	// Should return fetched value
	if got.Condition != "☁️" {
		t.Errorf("Condition = %q, want ☁️", got.Condition)
	}

	// Should be cached now
	cached, ok := cache.Get("JFK")
	if !ok {
		t.Fatal("expected value to be cached after fetch")
	}
	if cached.Temp != "+10°C" {
		t.Errorf("cached Temp = %q, want +10°C", cached.Temp)
	}
}

func TestNewCache(t *testing.T) {
	cache, err := NewCache()
	if err != nil {
		t.Fatalf("NewCache() error = %v", err)
	}
	if cache == nil {
		t.Fatal("NewCache() returned nil")
	}
	if cache.dir == "" {
		t.Error("cache.dir should not be empty")
	}
}

func TestNewCachedClient(t *testing.T) {
	client, err := NewCachedClient()
	if err != nil {
		t.Fatalf("NewCachedClient() error = %v", err)
	}
	if client == nil {
		t.Fatal("NewCachedClient() returned nil")
	}
}
