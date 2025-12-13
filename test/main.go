package main

import (
	"fmt"
	"strconv"
	"sync"
)

var mu sync.RWMutex
var data string

func reader(wg *sync.WaitGroup) {
	defer wg.Done()

	mu.RLock()
	fmt.Println(data)
	mu.RUnlock()
}

func writer(wg *sync.WaitGroup, val string) {
	defer wg.Done()

	mu.Lock()
	data = data + " " + val
	mu.Unlock()
}

func main() {
	data = "Hello random number insert"

	var wg sync.WaitGroup
	wg.Add(120)

	for i := 0; i < 130; i++ {
		if i%60 == 0 {
			go writer(&wg, strconv.Itoa(i))
		} else {
			go reader(&wg)
		}
	}
	
	wg.Wait()
}
