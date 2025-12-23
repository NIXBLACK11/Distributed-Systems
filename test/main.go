package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

const NUMWORKERS = 50

func worker(
	workerId int,
	ctx context.Context,
	jobs <-chan string,
	visited map[string]bool,
	wg *sync.WaitGroup,
	mu *sync.Mutex,
	rateLimiter *time.Ticker,
) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			fmt.Printf("Worker %d stopped\n", workerId)
			return
		case url, ok := <- jobs:
			if !ok {
				fmt.Printf("worker %d exiting (jobs channel closed)\n", workerId)
				return
			}

			mu.Lock()
			if visited[url] {
				mu.Unlock()
				continue
			}

			visited[url] = true
			mu.Unlock()

			<-rateLimiter.C

			req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				fmt.Println("error:", err)
				continue
			}

			fmt.Printf("worker %d fetched %s [%d]\n", workerId, url, resp.StatusCode)
			resp.Body.Close()
		}
	}
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
	defer cancel()

	jobs := make(chan string)
	visited := make(map[string]bool)

	var wg sync.WaitGroup
	var mu sync.Mutex

	rateLimiter := time.NewTicker(200 * time.Millisecond)
	defer rateLimiter.Stop()

	for workerId := range(NUMWORKERS) {
		wg.Add(1)
		go worker(workerId, ctx, jobs, visited, &wg, &mu, rateLimiter)
	}

	go func() {
		urls := []string {
			"https://example.com",
			"https://golang.org",
			"https://httpbin.org/get",
		}

		for _, url := range(urls) {
			jobs <- url
		}

		close(jobs)
	} ()

	wg.Wait()
	fmt.Println("crawl finished")
}