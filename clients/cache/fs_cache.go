package cache

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const filename = "entry" // Will be a json

type entry string

type FSCache struct {
	cacheDir string
}

func NewEntry(data interface{}) (entry, error) {
	jsonString, err := toJsonString(data)
	if err != nil {
		return entry(""), err
	}
	return entry(jsonString), nil
}

func toJsonString(data interface{}) (string, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func NewFSCache(cacheDir string) *FSCache {
	return &FSCache{
		cacheDir: cacheDir,
	}
}

func (c *FSCache) Set(_ context.Context, key string, data interface{}, _ time.Duration) error {
	entry, err := NewEntry(data)
	if err != nil {
		return err
	}

	// Create directory if it doesn't exist
	dirPath := filepath.Join(c.cacheDir, key)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return err
	}

	location := filepath.Join(dirPath, filename)

	// Write the file
	file, err := os.Create(location)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(string(entry))
	if err != nil {
		return err
	}
	return nil
}

// Get the data from the cache and unmarshal it into the data object
func (c *FSCache) Get(_ context.Context, key string, data interface{}) error {
	location := filepath.Join(c.cacheDir, key, filename)

	// Read the file
	file, err := os.Open(location)
	if err != nil {
		return err
	}
	defer file.Close()

	// Unmarshal the data
	decoder := json.NewDecoder(file)
	err = decoder.Decode(data)
	if err != nil {
		return err
	}
	return nil
}

func (c *FSCache) Delete(_ context.Context, key string) error {
  location := filepath.Join(c.cacheDir, key)
  return os.RemoveAll(location)
}
