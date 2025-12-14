```go
package main

import (
	"fmt"
	"strconv"
	"sync"
)

type KVStore struct {
	mu sync.RWMutex
	data map[string]string
}

func (k *KVStore) Get(key string) string {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.data[key]
}

func (k *KVStore) Set(key, value string) {
	k.mu.Lock()
	k.data[key] = value
	k.mu.Unlock()
}

func (k *KVStore) Delete(key string) {
	k.mu.Lock()
	delete(k.data, key)
	k.mu.Unlock()
}

func main() {
	store := &KVStore{
		data: make(map[string]string),
	}

	var wg sync.WaitGroup

	// 50 writers
	for i := range 50 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := "key-" + strconv.Itoa(i)
			store.Set(key, strconv.Itoa(i))
		}(i)
	}

	// 50 readers
	for i := range 50 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := "key-" + strconv.Itoa(i)
			fmt.Println("GET:", key, "=", store.Get(key))
		}(i)
	}

	// 20 deleters
	for i := range 20 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := "key-" + strconv.Itoa(i)
			store.Delete(key)
		}(i)
	}

	wg.Wait()

	fmt.Println("Done. Final store size:", len(store.data))
}
```