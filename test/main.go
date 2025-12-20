package main

import (
	"context"
	"fmt"
	"sync"
)

const NUMJOBS = 300
const NUMWORKERS = 4

func worker(workerId int, ctx context.Context, jobs <-chan int, wg *sync.WaitGroup) {
	defer wg.Done() // Signal completion when worker finishes

	for {
		select {
		case <-ctx.Done():
			// Context cancelled, exit gracefully
			fmt.Printf("Worker %d cancelled\n", workerId)
			return
		case jobId, ok := <-jobs:
			if !ok {
				// Channel closed, no more jobs
				fmt.Printf("Worker %d finished - no more jobs\n", workerId)
				return
			}
			fmt.Printf("Job %d being done by worker %d\n", jobId, workerId)
		}
	}
}

func main() {
	rootContext := context.Background()
	ctx, cancel := context.WithCancel(rootContext)

	defer cancel()

	var wg sync.WaitGroup
	wg.Add(NUMWORKERS)

	jobs := make(chan int, 10)

	// Start workers
	for workerId := range NUMWORKERS {
		go worker(workerId, ctx, jobs, &wg)
	}

	// Send jobs
	go func() {
		defer close(jobs) // Important: close channel when done sending
		for jobId := range NUMJOBS {
			select {
			case <-ctx.Done():
				fmt.Println("Job sender cancelled")
				return
			case jobs <- jobId:
				// Job sent successfully
			}
		}
	}()

	wg.Wait()
	fmt.Println("All workers finished")
}
