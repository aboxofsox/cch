package cch

import (
	"crypto/rand"
	"fmt"
	"sort"
	"sync"
	"time"
)

type Store struct {
	sync.Mutex
	id     string
	data   map[string]*Cache
	expire time.Time
}

// NewStore creates a new namespace cache store
func NewStore(id string) *Store {
	return &Store{
		id:     id,
		data:   make(map[string]*Cache),
		expire: time.Now().Add(time.Second * 30),
	}
}

// NewCache creates a new cachen in the given namespace
func (s *Store) NewCache(namespace string, expire time.Duration) (*Cache, error) {
	if s == nil {
		return nil, nilStore(namespace)
	}
	s.Lock()
	defer s.Unlock()

	if cache, exists := s.data[namespace]; !exists && cache != nil {
		return cache, fmt.Errorf("cache %s already exists", namespace)
	}

	cache := &Cache{
		namespace: namespace,
		storage:   new(sync.Map),
		expire:    time.Now().Add(expire),
	}
	s.data[namespace] = cache

	return cache, nil
}

func (s *Store) Namespaces() []string {
	if s == nil {
		return nil
	}
	var namespaces []string

	s.Lock()
	defer s.Unlock()

	for k := range s.data {
		namespaces = append(namespaces, k)
	}
	sort.Strings(namespaces)
	return namespaces
}

func (s *Store) Size() int {
	return len(s.Namespaces())
}

// UseNamespace returns a cache within the given namespace
func (s *Store) UseNamespace(namespace string) (*Cache, error) {
	if s == nil {
		return nil, nilCache(namespace)
	}

	s.Lock()
	defer s.Unlock()

	if s.data[namespace] == nil {
		return nil, fmt.Errorf("cache with %s namespace does not exist", namespace)
	}
	return s.data[namespace], nil
}

func (s *Store) Remove(namespace string) error {
	s.Lock()
	defer s.Unlock()

	if s == nil {
		return nilStore(namespace)
	}

	if _, exists := s.data[namespace]; !exists {
		return namespaceNotFound(namespace)
	}

	delete(s.data, namespace)

	return nil
}

func (s *Store) ExpireCache() error {
	for _, namespace := range s.Namespaces() {
		cache, err := s.UseNamespace(namespace)
		if err != nil {
			return err
		}
		if isCacheExpired(cache) {
			if err := s.Remove(namespace); err != nil {
				return err
			}
		}
	}
	return nil
}

func isCacheExpired(cache *Cache) bool {
	return cache.expire.After(time.Now()) && cache.Size() == 0
}

func nilStore(namespace string) error {
	return fmt.Errorf("store cannot be nil\n\tnamespace: %s", namespace)
}

func namespaceNotFound(namespace string) error {
	return fmt.Errorf("namespace not found: %s", namespace)
}

func uuid() string {
	b := make([]byte, 16)

	if _, err := rand.Read(b); err != nil {
		return ""
	}

	return fmt.Sprintf("%x-%x-%x-%x-%x", b[:4], b[4:6], b[6:8], b[8:10], b[10:])
}
