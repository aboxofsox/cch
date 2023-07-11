package cch

import (
	"crypto/rand"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"
)

func Test_Store(t *testing.T) {
	store := NewStore(uuid())

	namespaces := []string{
		"namespace test 1",
		"namespace test 2",
		"namespace test 3",
	}

	tests := map[string]int{
		"foo": 1,
		"bar": 2,
		"baz": 3,
	}

	for _, namespace := range namespaces {
		cache, err := store.NewCache(namespace, time.Millisecond*5000)
		if err != nil {
			t.Error(err)
		}

		for k, v := range tests {
			if err := cache.Add(k, v); err != nil {
				t.Error(err)
			}
		}
	}

	for _, namespace := range namespaces {
		cache, err := store.UseNamespace(namespace)
		if err != nil {
			t.Error(err)
		}

		for k, v := range tests {
			got, err := cache.Get(k)
			if err != nil {
				t.Error(err)
			}

			if got != v {
				t.Errorf("expected %v but got %v", v, got)
			}
		}
	}

	for _, namespace := range namespaces {
		cache, err := store.UseNamespace(namespace)
		if err != nil {
			t.Error(err)
		}

		for k := range tests {
			if err := cache.Remove(k); err != nil {
				t.Error(err)
			}
		}
	}

	for _, namespace := range namespaces {
		cache, err := store.UseNamespace(namespace)
		if err != nil {
			t.Error(err)
		}

		if cache.Size() != 0 {
			t.Errorf("exepcted cache to be empty but got a size of %d", cache.Size())
		}
	}

	for _, namespace := range namespaces {
		cache, err := store.UseNamespace(namespace)
		if err != nil {
			t.Error(err)
		}

		if cache.Size() == 0 {
			if err := store.Remove(namespace); err != nil {
				t.Error(err)
			}
		}
	}

	if store.Size() != 0 {
		t.Errorf("expected store to be empty but got a size of %d", store.Size())
	}
}

func Test_StoreConcurrency(t *testing.T) {
	wg := new(sync.WaitGroup)
	store := NewStore(uuid())

	namespaces := []string{
		"namespace test 1",
		"namespace test 2",
		"namespace test 3",
	}

	tests := map[string]int{
		"foo": 1,
		"bar": 2,
		"baz": 3,
	}

	exp := time.Millisecond * 5000

	for _, namespace := range namespaces {
		wg.Add(1)
		namespace := namespace
		go func() {
			defer wg.Done()
			cache, err := store.NewCache(namespace, exp)
			if err != nil {
				t.Error(err)
			}
			for k, v := range tests {
				if err := cache.Add(k, v); err != nil {
					t.Error(err)
				}
			}
		}()
	}
	wg.Wait()

	if store.Size() != len(namespaces) {
		t.Errorf("expected a store size of %d but got %d", len(namespaces), store.Size())
	}

	for _, namespace := range namespaces {
		cache, err := store.UseNamespace(namespace)
		if err != nil {
			t.Error(err)
		}
		for k, v := range tests {
			got, err := cache.Get(k)
			if err != nil {
				t.Error(err)
			}
			if got != v {
				t.Errorf("expected %v got %v", v, got)
			}
		}
	}
}

func Test_Expire(t *testing.T) {
	store := NewStore(uuid())

	namespaces := []string{
		"namespace test 1",
		"namespace test 2",
		"namespace test 3",
	}

	tests := map[string]int{
		"foo": 1,
		"bar": 2,
		"baz": 3,
	}

	expire := time.Second * 10

	for _, namespace := range namespaces {
		cache, err := store.NewCache(namespace, expire)
		if err != nil {
			t.Error(err)
		}
		for k, v := range tests {
			if err := cache.Add(k, v); err != nil {
				t.Error(err)
			}
		}
	}

	ticker := time.NewTicker(time.Millisecond * 1000)
	done := make(chan bool)

	i := 0
	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				if i == 5 {
					cache, err := store.UseNamespace(namespaces[0])
					if err != nil {
						t.Error(err)
					}
					if err := cache.Purge(); err != nil {
						t.Error(err)
					}
				}
				if err := store.ExpireCache(); err != nil {
					t.Error(err)
				}
				i++
			}
		}
	}()

	time.Sleep(time.Second * 15)
	done <- true
	if store.Size() != 2 {
		t.Errorf("expected store size of 2 but got %d", store.Size())
	}
}

func Test_Generics(t *testing.T) {
	namespaces := []string{
		"namespace test 1",
		"namespace test 2",
		"namespace test 3",
	}
	tests := map[string]any{
		"foo": '!',
		"bar": 1.3,
		"baz": []int{1, 2, 3},
	}

	store := NewStore(uuid())
	expire := time.Second * 5

	for _, namespace := range namespaces {
		cache, err := store.NewCache(namespace, expire)
		if err != nil {
			t.Error(err)
		}
		for k, v := range tests {
			if err := cache.Add(k, v); err != nil {
				t.Error(err)
			}
		}
	}

	for _, namespace := range namespaces {
		cache, err := store.UseNamespace(namespace)
		if err != nil {
			t.Error(err)
		}

		for k, v := range tests {
			got, err := cache.Get(k)
			if err != nil {
				t.Error(err)
			}
			if !reflect.DeepEqual(v, got) {
				t.Errorf("expected %v but got %v", v, got)
			}
		}
	}
}

func Test_Update(t *testing.T) {
	store := NewStore("test store")
	cache, err := store.NewCache("test cache", time.Millisecond*500)
	if err != nil {
		t.Error(err)
	}

	tests := map[string]int{
		"foo": 1,
		"bar": 2,
		"baz": 3,
	}

	for k, v := range tests {
		if err := cache.Add(k, v); err != nil {
			t.Error(err)
		}
	}

	if err := cache.Replace("foo", 7); err != nil {
		t.Error(err)
	}

	v, err := cache.Get("foo")
	if err != nil {
		t.Error(err)
	}

	if err := cache.Replace("foo", v.(int)+1); err != nil {
		t.Error(err)
	}

	v, err = cache.Get("foo")
	if err != nil {
		t.Error(err)
	}

	if v.(int) != 8 {
		t.Errorf("exepcted %d but got %d", 7, v.(int))
	}
}

func rando(size int) string {
	buff := make([]byte, size)

	if _, err := rand.Read(buff); err != nil {
		return ""
	}

	return string(buff)
}

func Test_Load(t *testing.T) {
	store := NewStore("load store")
	cache, _ := store.NewCache("load", time.Minute)

	tests := 100
	mb := 1 << 12

	for i := 0; i < tests; i++ {
		k := fmt.Sprintf("test %d", i)
		s := rando(i * mb)
		if err := cache.Add(k, s); err != nil {
			t.Error(err)
		}
	}

	for i := 0; i < tests; i++ {
		k := fmt.Sprintf("test %d", i)
		s, _ := cache.Get(k)
		fmt.Println(len(s.(string)))
	}
	time.Sleep(time.Second * 10)
}
