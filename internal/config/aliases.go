// Package config provides configuration management for the t CLI,
// including alias storage and retrieval.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// DefaultConfigDir returns the default configuration directory.
// Uses ~/.config/t on Unix-like systems.
func DefaultConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting home directory: %w", err)
	}
	return filepath.Join(home, ".config", "t"), nil
}

// AliasStore manages persistent storage of IATA code aliases.
type AliasStore struct {
	path    string
	aliases map[string][]string
}

// NewAliasStore creates a new AliasStore using the default config directory.
func NewAliasStore() (*AliasStore, error) {
	configDir, err := DefaultConfigDir()
	if err != nil {
		return nil, err
	}
	return NewAliasStoreWithPath(filepath.Join(configDir, "aliases.json"))
}

// NewAliasStoreWithPath creates a new AliasStore with a custom file path.
func NewAliasStoreWithPath(path string) (*AliasStore, error) {
	store := &AliasStore{
		path:    path,
		aliases: make(map[string][]string),
	}

	// Load existing aliases if file exists
	if err := store.load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("loading aliases: %w", err)
	}

	return store, nil
}

// load reads aliases from the JSON file.
func (s *AliasStore) load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		return nil
	}

	return json.Unmarshal(data, &s.aliases)
}

// save writes aliases to the JSON file.
func (s *AliasStore) save() error {
	// Ensure directory exists
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := json.MarshalIndent(s.aliases, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling aliases: %w", err)
	}

	if err := os.WriteFile(s.path, data, 0o644); err != nil {
		return fmt.Errorf("writing aliases file: %w", err)
	}

	return nil
}

// Save stores an alias with the given name and IATA codes.
// Alias names are case-insensitive and stored lowercase.
func (s *AliasStore) Save(name string, codes []string) error {
	name = strings.ToLower(name)
	if name == "" {
		return errors.New("alias name cannot be empty")
	}
	if len(codes) == 0 {
		return errors.New("alias must have at least one code")
	}

	// Normalize codes to uppercase
	normalized := make([]string, len(codes))
	for i, code := range codes {
		normalized[i] = strings.ToUpper(code)
	}

	s.aliases[name] = normalized
	return s.save()
}

// Get retrieves the IATA codes for a given alias name.
// Returns nil if the alias doesn't exist.
func (s *AliasStore) Get(name string) []string {
	name = strings.ToLower(name)
	codes, ok := s.aliases[name]
	if !ok {
		return nil
	}
	// Return a copy to prevent mutation
	result := make([]string, len(codes))
	copy(result, codes)
	return result
}

// Delete removes an alias by name.
// Returns an error if the alias doesn't exist.
func (s *AliasStore) Delete(name string) error {
	name = strings.ToLower(name)
	if _, ok := s.aliases[name]; !ok {
		return fmt.Errorf("alias '%s' not found", name)
	}

	delete(s.aliases, name)
	return s.save()
}

// List returns all aliases sorted by name.
func (s *AliasStore) List() map[string][]string {
	result := make(map[string][]string, len(s.aliases))
	for name, codes := range s.aliases {
		codeCopy := make([]string, len(codes))
		copy(codeCopy, codes)
		result[name] = codeCopy
	}
	return result
}

// ListSorted returns alias names in sorted order.
func (s *AliasStore) ListSorted() []string {
	names := make([]string, 0, len(s.aliases))
	for name := range s.aliases {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Exists checks if an alias exists.
func (s *AliasStore) Exists(name string) bool {
	name = strings.ToLower(name)
	_, ok := s.aliases[name]
	return ok
}
