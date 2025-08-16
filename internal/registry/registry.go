package registry

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type StickyRegistry struct {
	StickyPatterns []string `json:"sticky_patterns"`
}

type Registry interface {
	// AddPattern adds a pattern to the sticky registry
	AddPattern(pattern string) error
	
	// RemovePattern removes a pattern from the sticky registry
	RemovePattern(pattern string) error
	
	// HasPattern checks if a pattern exists in the registry
	HasPattern(pattern string) bool
	
	// GetPatterns returns all patterns in the registry
	GetPatterns() []string
	
	// IsEmpty returns true if there are no sticky patterns
	IsEmpty() bool
	
	// Save persists the registry to disk
	Save() error
	
	// Load reads the registry from disk
	Load() error
}

type FileRegistry struct {
	registry StickyRegistry
	filePath string
}

func NewFileRegistry() (*FileRegistry, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return nil, fmt.Errorf("unable to get config directory: %w", err)
	}
	
	filePath := filepath.Join(configDir, "registry.json")
	
	registry := &FileRegistry{
		registry: StickyRegistry{StickyPatterns: []string{}},
		filePath: filePath,
	}
	
	// Try to load existing registry
	if err := registry.Load(); err != nil {
		// If file doesn't exist, that's ok, we'll start with empty registry
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("unable to load registry: %w", err)
		}
	}
	
	return registry, nil
}

func (f *FileRegistry) AddPattern(pattern string) error {
	if f.HasPattern(pattern) {
		return nil // Already exists
	}
	
	f.registry.StickyPatterns = append(f.registry.StickyPatterns, pattern)
	return f.Save()
}

func (f *FileRegistry) RemovePattern(pattern string) error {
	for i, p := range f.registry.StickyPatterns {
		if p == pattern {
			f.registry.StickyPatterns = append(f.registry.StickyPatterns[:i], f.registry.StickyPatterns[i+1:]...)
			return f.Save()
		}
	}
	return nil // Pattern not found, nothing to remove
}

func (f *FileRegistry) HasPattern(pattern string) bool {
	for _, p := range f.registry.StickyPatterns {
		if p == pattern {
			return true
		}
	}
	return false
}

func (f *FileRegistry) GetPatterns() []string {
	return f.registry.StickyPatterns
}

func (f *FileRegistry) IsEmpty() bool {
	return len(f.registry.StickyPatterns) == 0
}

func (f *FileRegistry) Save() error {
	// Ensure config directory exists
	configDir := filepath.Dir(f.filePath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("unable to create config directory: %w", err)
	}
	
	data, err := json.MarshalIndent(f.registry, "", "  ")
	if err != nil {
		return fmt.Errorf("unable to marshal registry: %w", err)
	}
	
	if err := os.WriteFile(f.filePath, data, 0644); err != nil {
		return fmt.Errorf("unable to write registry file: %w", err)
	}
	
	return nil
}

func (f *FileRegistry) Load() error {
	data, err := os.ReadFile(f.filePath)
	if err != nil {
		return err
	}
	
	if err := json.Unmarshal(data, &f.registry); err != nil {
		return fmt.Errorf("unable to unmarshal registry: %w", err)
	}
	
	return nil
}

func getConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	
	return filepath.Join(homeDir, ".config", "aerospace-sticky"), nil
}