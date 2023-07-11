package cch

import (
	"fmt"
	"sync"
	"time"
)

type Cache struct {
	namespace string
	storage   *sync.Map
	expire    time.Time
}

// Add adds a new item to the cache
func (c *Cache) Add(key string, value any) error {
	if c == nil {
		return nilCache(c.namespace)
	}
	if _, exists := c.storage.Load(key); exists {
		return fmt.Errorf("key already exists: %s", key)
	}
	c.storage.Store(key, value)
	return nil
}

// Remove removes an item from the cache
func (c *Cache) Remove(key string) error {
	if c == nil {
		return nilCache(c.namespace)
	}
	if _, exists := c.storage.Load(key); !exists {
		return fmt.Errorf("key does not exist: %s", key)
	}
	c.storage.Delete(key)
	return nil
}

// Get gets an item from the cache by key
func (c *Cache) Get(key string) (any, error) {
	if c == nil {
		return nil, nilCache(c.namespace)
	}
	value, exists := c.storage.Load(key)
	if !exists {
		return nil, fmt.Errorf("key not found: %s", key)
	}
	return value, nil
}

// Replace removes the value and replaces it with a new one
func (c *Cache) Replace(key string, newValue any) error {
	if c == nil {
		return nilCache(c.namespace)
	}
	if _, exists := c.storage.Load(key); !exists {
		return keyNotExists(key, c.namespace)
	}
	if err := c.Remove(key); err != nil {
		return err
	}
	if err := c.Add(key, newValue); err != nil {
		return err
	}
	return nil
}

// Purge clears the cache
func (c *Cache) Purge() error {
	c.storage.Range(func(key, value any) bool {
		if err := c.Remove(key.(string)); err != nil {
			fmt.Println(err.Error())
			return false
		}
		return true
	})
	return nil
}

// Map returns a map[string]any of the given cache
func (c *Cache) Map() (map[string]any, error) {
	if c == nil {
		return nil, nilCache(c.namespace)
	}
	mp := make(map[string]any)
	c.storage.Range(func(key, value any) bool {
		mp[key.(string)] = value
		return true
	})
	return mp, nil
}

// Size returns the size of the given cache
func (c *Cache) Size() int {
	if c == nil {
		return 0
	}
	i := 0
	c.storage.Range(func(key, value any) bool {
		i += 1
		return true
	})
	return i
}

func nilCache(namespace string) error {
	return fmt.Errorf("cache cannot be nil\n\tnamespace: %s", namespace)
}

func keyNotExists(key, namespace string) error {
	return fmt.Errorf("key not found:\n\tkey: %s\nnamespace:%s\n", key, namespace)
}
