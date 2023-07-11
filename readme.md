### Package `cch`
This package provides a concurrent-safe cache implementation for Go.

## Basic Usage
### HTTP Server

```go
package main

import (
	"fmt"
	"github.com/aboxofsox/cch"
	"log"
	"net/http"
)

func storeName(cache *cch.Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")

		if err := cache.Add(r.RemoteAddr, name); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write([]byte("ok"))
	}
}

func getName(cache *cch.Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name, err := cache.Get(r.RemoteAddr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		msg := "hello " + name.(string) + "!"

		w.Write([]byte(msg))
	}
}

func main() {
	store := cch.NewStore("primary")
	cache, err := store.NewCache("names", time.MilliSecond*500)
	if err != nil {
		log.Fatal(err.Error())
	}

	http.HandleFunc("/store", storeName(cache))
	http.HandleFunc("/get", getName(cache))
	
	if err := http.ListenAndServe(":3000", nil); err != nil {
		log.Fatal(err.Error())
    }
}
```
### Documentation
- [Types](#types)
- [Cache Functions](#cache-functions)
  - [Add](#add)
  - [Remove](#remove)
  - [Get](#get)
  - [Purge](#purge)
  - [Map](#map)
  - [Size](#size)
- [Store Functions](#store-functions)
  - [NewCache](#newcache)
  - [Namespaces](#namespaces)
  - [UseNamespace](#usenamespace)
  - [Remove](#remove-store)
  - [ExpireCache](#expirecache)

## Types
#### Cache
```go
package cch

type Cache struct {
	namespace string
	storage *sync.Map
	expire time.Time
}

type Store struct {
	sync.Mutex
	id string
	data map[string]*Cache
	expire time.Time
}
```
The `Cache` type represents a cache with a set of utility methods for cache manipulation. It holds the cache namespace, its storage, an expiration interval.

The `Store` type is a storage entity that encapsulates cache data. It extends `sync.Mutex` for providing atomic operations (safe for concurrent use), contains an id type string as a unique identifier, a `map[string]*Cache` where key is of type string (namespace) and value is a pointer to `Cache` and an expiration of type `Time`. 

### Cache Functions
#### Add
Add a new item to the cache.
```go
func (c *Cache) Add(key string, value any) error
```
The function checks if the key already exists in the cache, if so, it will return an error. If not, it stores the value associated with the key.
#### Remove
Remove an ituem from the cache
```go
func (c *Cache) Remove(key string) error
```
The function checks if the key exists in the cache, if so, it deletes the value associated with the key. If not, it will return an error.
#### Get
Get an item from the cache by key.
```go
func (c *Cache) Get(key string) (any, error)
```
The function checks if the key exists in the cache, if so, it retunrs the value associated with the key. If not, it will return an error
#### Purge
Clears the cache.
```go
func (c *Cache) Purge() error
```
The function iterates over the cache and removes all key-value pairs.
#### Map
Returns a `map[string]any` of the given cache.
```go
func (c *Cache) Map() (map[string]any, error)
```
The function creates and returns a map where the keys are strings and the values are of type `any`. This map represents the current state of the cache.
#### Size
Returns the size of the cache.
```go
func (c *Cache) Size() int
```
This function counts and returns the total number of key-value pairs currently in the cache.
### Store Functions
#### NewStore
```go
func NewStore(id string) *Store
````
The function initializes a new cache store with given id and expiration time set to 30 seconds from the current timestamp (`time.Now()`).
#### NewCache
```go
func (s *Store) NewCache(namespace string, expire time.Duration) (*Cache, error)
```
The function creates a new cache in the store under the given namespace. The cache items are set to expire after the given expiration.
#### Namespaces
```go
func (s *Store) Namespaces() []string
```
Returns a slice of strings containing all the namespaces in the store.
#### UseNamespace
```go
func (s *Store) UseNamespace(namespace string) (*Cache, error)
```
Returns a cache within the given namespace. If the namespace does not exist, it will return an error.
#### Remove [Store]
```go
func (s *Store) Remove(namespace string) error
```
Removes a namespace and its associated cache form the store. If the namespace does not exist in the store, it will return an error.
#### ExpireCache
```go
func (s *Store) ExpireCache() error
```
Iterates through all caches in the store, and if a cache is expired (based on the `isCacheExpired` helper method), it removes the cache from teh store. It returns an error if it cannot use a namespace or remove the cache.
#### Helper Methods
```go
func isCacheExpired(cache *Cache) bool
```
Returns true if the cache is expired.
