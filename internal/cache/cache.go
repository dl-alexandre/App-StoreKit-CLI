package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Cache provides simple file-based caching
type Cache struct {
	dir string
	ttl time.Duration
}

// CacheEntry represents a cached item
type CacheEntry struct {
	Data      []byte    `json:"data"`
	CreatedAt time.Time `json:"created_at"`
	TTL       int       `json:"ttl"`
}

// New creates a new cache instance
func New(dir string, ttl time.Duration) *Cache {
	return &Cache{
		dir: dir,
		ttl: ttl,
	}
}

// Get retrieves an item from the cache
func (c *Cache) Get(key string) (any, bool) {
	path := c.filePath(key)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}

	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, false
	}

	// Check if expired
	if time.Since(entry.CreatedAt) > c.ttl {
		_ = os.Remove(path)
		return nil, false
	}

	// Unmarshal the actual data
	var result any
	if err := json.Unmarshal(entry.Data, &result); err != nil {
		return nil, false
	}

	return result, true
}

// Set stores an item in the cache
func (c *Cache) Set(key string, value any, ttl time.Duration) error {
	path := c.filePath(key)

	// Ensure cache directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Marshal the value
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal cache data: %w", err)
	}

	entry := CacheEntry{
		Data:      data,
		CreatedAt: time.Now(),
		TTL:       int(ttl.Seconds()),
	}

	encoded, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to encode cache entry: %w", err)
	}

	if err := os.WriteFile(path, encoded, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

// Delete removes an item from the cache
func (c *Cache) Delete(key string) error {
	path := c.filePath(key)
	return os.Remove(path)
}

// Clear removes all cached items
func (c *Cache) Clear() error {
	return os.RemoveAll(c.dir)
}

// filePath returns the file path for a cache key
func (c *Cache) filePath(key string) string {
	// Use hash of key as filename
	hash := sha256.Sum256([]byte(key))
	hashStr := hex.EncodeToString(hash[:])
	// Use first 2 chars of hash as subdirectory to avoid too many files in one dir
	return filepath.Join(c.dir, hashStr[:2], hashStr)
}
