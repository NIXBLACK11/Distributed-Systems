package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

var urls = []string{
	"https://google.com",
	"https://google.in",
}

const NUMWORKERS = 5

func fetchUrlDetails(url string) []string {
	return []string{
		"https://google.us",
		"https://google.can",
		url, 
	}
}

func worker(
	workerId int,
	ctx context.Context,
	jobs chan string,
	visited map[string]bool,
	wg *sync.WaitGroup,
	mu *sync.Mutex,
	rateLimiter *time.Ticker,
) {
	for {
		select {
		case <-ctx.Done():
			return

		case url, ok := <-jobs:
			if !ok {
				return
			}

			// mark job done when function exits
			func() {
				defer wg.Done()

				mu.Lock()
				if visited[url] {
					mu.Unlock()
					return
				}
				visited[url] = true
				mu.Unlock()

				<-rateLimiter.C

				subUrls := fetchUrlDetails(url)

				for _, subUrl := range subUrls {
					wg.Add(1)
					jobs <- subUrl
				}

				fmt.Printf("worker %d processed %s\n", workerId, url)
			}()
		}
	}
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	jobs := make(chan string, 100)
	visited := make(map[string]bool)

	var wg sync.WaitGroup
	var mu sync.Mutex

	rateLimiter := time.NewTicker(200 * time.Millisecond)
	defer rateLimiter.Stop()

	for i := range NUMWORKERS {
		go worker(i, ctx, jobs, visited, &wg, &mu, rateLimiter)
	}

	for _, url := range urls {
		wg.Add(1)
		jobs <- url
	}

	go func() {
		wg.Wait()
		close(jobs)
	}()

	// wait for context or completion
	<-ctx.Done()
	fmt.Println("JOB FINISHED")
}
