## Worker pool with cancel

```go
package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

const NUMJOBS = 10000
const NUMWORKERS = 1

func progress(ctx context.Context) {
	x := 0
	for i := range 10_000_000 {
		if i%1000 == 0 {
			select {
			case <-ctx.Done():
				fmt.Println("  progress: received cancel, stopping early")
				return
			default:
			}
		}
		x += i
	}
	_ = x
}

func worker(ctx context.Context, id int, jobs <-chan int, wg *sync.WaitGroup) {
	defer wg.Done() // ✅ worker lifetime tracked

	fmt.Printf("Worker %d: started\n", id)
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("Worker %d: shutting down: %v\n", id, ctx.Err())
			return
		case job, ok := <-jobs:
			if !ok {
				fmt.Printf("Worker %d: jobs channel closed, exiting\n", id)
				return
			}

			fmt.Printf("Worker %d: working on job %d\n", id, job)

			jobTimeout := 800 * time.Millisecond
			jobCtx, cancel := context.WithTimeout(ctx, jobTimeout)

			progress(jobCtx)
			cancel()

			fmt.Printf("Worker %d: finished job %d (or was canceled)\n", id, job)
		}
	}
}

func main() {
	root := context.Background()
	ctx, cancel := context.WithCancel(root)
	defer cancel()

	jobs := make(chan int, NUMJOBS)

	var wg sync.WaitGroup
	wg.Add(NUMWORKERS)

	for workerId := range NUMWORKERS {
		go worker(ctx, workerId, jobs, &wg)
	}

	for jobId := range NUMJOBS {
		jobs <- jobId
	}
	close(jobs)

	time.AfterFunc(2*time.Second, func() {
		fmt.Println("[main] Cancelling root context (global shutdown)")
		cancel()
	})

	wg.Wait()
	fmt.Println("All workers exited cleanly. Exiting.")
}
```