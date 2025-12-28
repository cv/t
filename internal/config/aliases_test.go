package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAliasStore(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "aliases.json")

	store, err := NewAliasStoreWithPath(path)
	require.NoError(t, err)
	assert.NotNil(t, store)
	assert.Empty(t, store.List())
}

func TestAliasStore_SaveAndGet(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "aliases.json")

	store, err := NewAliasStoreWithPath(path)
	require.NoError(t, err)

	// Save an alias
	err = store.Save("team", []string{"sfo", "jfk", "lon"})
	require.NoError(t, err)

	// Get the alias
	codes := store.Get("team")
	assert.Equal(t, []string{"SFO", "JFK", "LON"}, codes)

	// Case insensitive get
	codes = store.Get("TEAM")
	assert.Equal(t, []string{"SFO", "JFK", "LON"}, codes)

	// Non-existent alias
	codes = store.Get("nonexistent")
	assert.Nil(t, codes)
}

func TestAliasStore_Persistence(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "aliases.json")

	// Create store and save alias
	store1, err := NewAliasStoreWithPath(path)
	require.NoError(t, err)
	err = store1.Save("west", []string{"sfo", "lax", "sea"})
	require.NoError(t, err)

	// Create new store from same path - should load saved aliases
	store2, err := NewAliasStoreWithPath(path)
	require.NoError(t, err)
	codes := store2.Get("west")
	assert.Equal(t, []string{"SFO", "LAX", "SEA"}, codes)
}

func TestAliasStore_Delete(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "aliases.json")

	store, err := NewAliasStoreWithPath(path)
	require.NoError(t, err)

	// Save and delete
	err = store.Save("temp", []string{"jfk"})
	require.NoError(t, err)
	assert.True(t, store.Exists("temp"))

	err = store.Delete("temp")
	require.NoError(t, err)
	assert.False(t, store.Exists("temp"))

	// Verify persistence after delete
	store2, err := NewAliasStoreWithPath(path)
	require.NoError(t, err)
	assert.False(t, store2.Exists("temp"))
}

func TestAliasStore_DeleteNonexistent(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "aliases.json")

	store, err := NewAliasStoreWithPath(path)
	require.NoError(t, err)

	err = store.Delete("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestAliasStore_List(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "aliases.json")

	store, err := NewAliasStoreWithPath(path)
	require.NoError(t, err)

	err = store.Save("team", []string{"sfo", "jfk"})
	require.NoError(t, err)
	err = store.Save("west", []string{"lax", "sea"})
	require.NoError(t, err)

	list := store.List()
	assert.Len(t, list, 2)
	assert.Equal(t, []string{"SFO", "JFK"}, list["team"])
	assert.Equal(t, []string{"LAX", "SEA"}, list["west"])
}

func TestAliasStore_ListSorted(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "aliases.json")

	store, err := NewAliasStoreWithPath(path)
	require.NoError(t, err)

	err = store.Save("zebra", []string{"jfk"})
	require.NoError(t, err)
	err = store.Save("alpha", []string{"sfo"})
	require.NoError(t, err)
	err = store.Save("beta", []string{"lax"})
	require.NoError(t, err)

	sorted := store.ListSorted()
	assert.Equal(t, []string{"alpha", "beta", "zebra"}, sorted)
}

func TestAliasStore_SaveValidation(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "aliases.json")

	store, err := NewAliasStoreWithPath(path)
	require.NoError(t, err)

	// Empty name
	err = store.Save("", []string{"sfo"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty")

	// Empty codes
	err = store.Save("test", []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one")
}

func TestAliasStore_Overwrite(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "aliases.json")

	store, err := NewAliasStoreWithPath(path)
	require.NoError(t, err)

	err = store.Save("team", []string{"sfo", "jfk"})
	require.NoError(t, err)

	// Overwrite with new codes
	err = store.Save("team", []string{"lax", "ord", "dfw"})
	require.NoError(t, err)

	codes := store.Get("team")
	assert.Equal(t, []string{"LAX", "ORD", "DFW"}, codes)
}

func TestAliasStore_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	// Nested path that doesn't exist
	path := filepath.Join(tmpDir, "subdir", "nested", "aliases.json")

	store, err := NewAliasStoreWithPath(path)
	require.NoError(t, err)

	err = store.Save("test", []string{"sfo"})
	require.NoError(t, err)

	// Verify file was created
	_, err = os.Stat(path)
	assert.NoError(t, err)
}

func TestAliasStore_GetReturnsCopy(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "aliases.json")

	store, err := NewAliasStoreWithPath(path)
	require.NoError(t, err)

	err = store.Save("team", []string{"sfo", "jfk"})
	require.NoError(t, err)

	// Get codes and modify
	codes := store.Get("team")
	codes[0] = "MODIFIED"

	// Original should be unchanged
	codes2 := store.Get("team")
	assert.Equal(t, []string{"SFO", "JFK"}, codes2)
}

func TestAliasStore_LoadEmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "aliases.json")

	// Create empty file
	err := os.WriteFile(path, []byte{}, 0o644)
	require.NoError(t, err)

	store, err := NewAliasStoreWithPath(path)
	require.NoError(t, err)
	assert.Empty(t, store.List())
}

func TestAliasStore_LoadCorruptFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "aliases.json")

	// Create corrupt file
	err := os.WriteFile(path, []byte("not valid json"), 0o644)
	require.NoError(t, err)

	_, err = NewAliasStoreWithPath(path)
	assert.Error(t, err)
}
