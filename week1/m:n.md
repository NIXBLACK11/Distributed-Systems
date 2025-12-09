## M:N scheduler

Go runtime uses an M:N scheduler, this means:
- M OS threads SERVE N goroutines

The scheduler maps multiple goroutines into a small pool of OS threads.

This is when GOMAXPROCS comes into picture, this helps Go limits how many OS threads run Go code in parallel.

Simple code to check the usage of GOMAXPROCS
```go
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

```

The execution time decreases as we increase the number of cps
![output](../images/gomaxprocs.png)