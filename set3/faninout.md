## FAN-IN and FAN-OUT

This is a simple comcept where our operation is divided into two major parts:
FAn-Out does the parallelization as it divided the input to multiple workers so they can parallely work on it.
Fan-in takes the outputs from these multiple workers and combines it into a single result.

```go
package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

const NUMJOBS = 20
const NUMWORKERS = 3

type Result struct {
	WorkerID int
	JobID    int
	Err      error
}

func progress(ctx context.Context) error {
	x := 0
	for i := 0; i < 10_000_000; i++ {
		if i%1000 == 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
		}
		x += i
	}
	_ = x
	return nil
}

/* ---------------- FAN-OUT WORKER ---------------- */

func worker(
	ctx context.Context,
	id int,
	jobs <-chan int,
	results chan<- Result,
	wg *sync.WaitGroup,
) {
	defer wg.Done()
	fmt.Printf("Worker %d started\n", id)

	for {
		select {
		case <-ctx.Done():
			fmt.Printf("Worker %d shutting down: %v\n", id, ctx.Err())
			return

		case job, ok := <-jobs:
			if !ok {
				fmt.Printf("Worker %d: jobs closed\n", id)
				return
			}

			jobCtx, cancel := context.WithTimeout(ctx, 800*time.Millisecond)
			err := progress(jobCtx)
			cancel()

			results <- Result{
				WorkerID: id,
				JobID:    job,
				Err:      err,
			}
		}
	}
}

/* ---------------- FAN-IN AGGREGATOR ---------------- */

func aggregator(results <-chan Result, done chan<- struct{}) {
	for res := range results {
		if res.Err != nil {
			fmt.Printf("[AGG] worker=%d job=%d failed: %v\n",
				res.WorkerID, res.JobID, res.Err)
		} else {
			fmt.Printf("[AGG] worker=%d job=%d done\n",
				res.WorkerID, res.JobID)
		}
	}
	close(done)
}

/* ---------------- MAIN ---------------- */

func main() {
	root := context.Background()
	ctx, cancel := context.WithCancel(root)
	defer cancel()

	jobs := make(chan int, NUMJOBS)
	results := make(chan Result, NUMJOBS)
	done := make(chan struct{})

	var workerWG sync.WaitGroup
	workerWG.Add(NUMWORKERS)

	// FAN-OUT
	for i := 0; i < NUMWORKERS; i++ {
		go worker(ctx, i, jobs, results, &workerWG)
	}

	// FAN-IN
	go aggregator(results, done)

	// send jobs
	for j := 0; j < NUMJOBS; j++ {
		jobs <- j
	}
	close(jobs)

	// global cancel
	time.AfterFunc(2*time.Second, func() {
		fmt.Println("[main] global cancel")
		cancel()
	})

	// wait for workers → then close results
	workerWG.Wait()
	close(results)

	// wait for aggregator
	<-done
	fmt.Println("[main] clean shutdown")
}
```