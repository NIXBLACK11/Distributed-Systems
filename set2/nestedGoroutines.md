## Nested goroutines and cancellation propogation

```go
package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

func childWorker(ctx context.Context, id int, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Printf("[worker %d] started\n", id)
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("[worker %d] stopped: %v\n", id, ctx.Err())
			return
		default:
			// simulate work
			time.Sleep(200 * time.Millisecond)
		}
	}
}

func parent(ctx context.Context) {
	// create child context (cancellable)
	ctxChild, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go childWorker(ctxChild, i, &wg)
	}

	// spawn a nested goroutine that cancels child after 1s
	go func() {
		time.Sleep(2 * time.Second)
		fmt.Println("[parent] cancelling child context")
		cancel()
	}()

	wg.Wait()
	fmt.Println("[parent] all workers done")
}

func main() {
	root := context.Background()

	ctx, cancel := context.WithTimeout(root, 3 * time.Second)
	defer cancel()

	parent(ctx)

	fmt.Println("Main done")
}
```