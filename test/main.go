package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

func busy() {
	x := 0
	for i := 0; i < 10000000; i++ {
		x += i
	}
	_ = x
}

func main() {
	for _, procs := range []int{1, 2, 4, runtime.NumCPU()} {
		runtime.GOMAXPROCS(procs)
		start := time.Now()

		var wg sync.WaitGroup
		wg.Add(runtime.NumCPU())
		for i := 0; i < runtime.NumCPU(); i++ {
			go func() {
				busy() // CPU-bound
				wg.Done()
			}()
		}
		wg.Wait()
		fmt.Printf("GOMAXPROCS=%d took %v\n", procs, time.Since(start))
	}
}
