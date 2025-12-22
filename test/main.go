package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

func worker(
	ctx context.Context,
	id int,
	jobs <-chan string,
	wg *sync.WaitGroup,
	rateLimiter *time.Ticker,
	visited map[string]bool,
	mu *sync.Mutex,
) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			fmt.Printf("worker %d stopped\n", id)
			return

		case url, ok := <-jobs:
			if !ok {
				return
			}

			// visited check
			mu.Lock()
			if visited[url] {
				mu.Unlock()
				continue
			}
			visited[url] = true
			mu.Unlock()

			// rate limit
			<-rateLimiter.C

			req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				fmt.Println("error:", err)
				continue
			}

			fmt.Printf("worker %d fetched %s [%d]\n", id, url, resp.StatusCode)
			resp.Body.Close()
		}
	}
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	jobs := make(chan string)
	visited := make(map[string]bool)
	var mu sync.Mutex

	rateLimiter := time.NewTicker(200 * time.Millisecond)
	defer rateLimiter.Stop()

	var wg sync.WaitGroup

	// start 50 workers
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go worker(ctx, i, jobs, &wg, rateLimiter, visited, &mu)
	}

	// seed URLs
	go func() {
		urls := []string{
			"https://example.com",
			"https://golang.org",
			"https://httpbin.org/get",
		}

		for _, u := range urls {
			jobs <- u
		}
		close(jobs)
	}()

	wg.Wait()
	fmt.Println("crawl finished")
}
