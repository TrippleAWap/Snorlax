package cache

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"unsafe"
)

type Cache struct {
	Path      string
	data      map[string]interface{}
	dataMutex sync.Mutex
	fStream   *os.File
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
	c.dataMutex.Lock()
	defer c.dataMutex.Unlock()
	if err := decoder.Decode(&c.data); err != nil {
		fmt.Printf("cache - loadFromFile | Error decoding file: %s\n", err)
	}
	return nil
}

func (c *Cache) saveToFile() error {
	log.Printf("cache - saveToFile | Saving cache to file: %s\n", c.Path)
	file, err := os.OpenFile(c.Path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return fmt.Errorf("cache - saveToFile | Error opening file: %s", err)
	}
	log.Printf("cache - saveToFile | Opened file: %s\n", c.Path)
	defer file.Close()

	log.Print("cache - saveToFile | creating encoder\n")
	encoder := json.NewEncoder(file)
	log.Print("cache - saveToFile | created encoder\n")

	log.Println("cache - saveToFile | convert data to JSON")
	if err := encoder.Encode(c.data); err != nil {
		return fmt.Errorf("cache - saveToFile | Error encoding JSON: %s", err)
	}

	log.Printf("cache - saveToFile | saved to: %s\n", c.Path)

	return nil
}

func (c *Cache) Set(key string, value interface{}) error {
	fmt.Println("Set called.")

	c.dataMutex.Lock()
	defer c.dataMutex.Unlock()

	log.Printf("cache - Set | Setting key: %s, sizeOf: %v\n", key, unsafe.Sizeof(value))
	c.data[key] = value
	log.Printf("cache - Set | Set key: %s, sizeOf: %v\n", key, unsafe.Sizeof(value))
	fmt.Println("Set finished.")
	return c.saveToFile()
}

func (c *Cache) Get(key string) interface{} {
	c.dataMutex.Lock()
	defer c.dataMutex.Unlock()

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
