package cache

import (
	"encoding/json"
	"fmt"
	"os"
)

type Cache struct {
	Path    string
	data    map[string]interface{}
	fStream *os.File
}

func New(path string) *Cache {
	return &Cache{
		Path: path,
		data: make(map[string]interface{}),
	}
}

func (c *Cache) loadFromFile() error {
	file, err := os.Open(c.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File does not exist, nothing to load
		}
		return fmt.Errorf("cache - loadFromFile | Error opening file: %s", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&c.data); err != nil {
		return fmt.Errorf("cache - loadFromFile | Error decoding JSON: %s", err)
	}
	return nil
}

func (c *Cache) saveToFile() error {
	file, err := os.OpenFile(c.Path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return fmt.Errorf("cache - saveToFile | Error opening file: %s", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(c.data); err != nil {
		return fmt.Errorf("cache - saveToFile | Error encoding JSON: %s", err)
	}
	return nil
}

func (c *Cache) Set(key string, value interface{}) error {
	c.data[key] = value
	return c.saveToFile()
}

func (c *Cache) Get(key string) interface{} {
	value, exists := c.data[key]
	if !exists {
		return nil // Key does not exist
	}
	return value
}

func (c *Cache) Load() error {
	return c.loadFromFile()
}

func (c *Cache) Close() error {
	return c.saveToFile()
}
