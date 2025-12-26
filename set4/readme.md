# Set4

## Goal:
- Build HTTP fetch workers  
- Add rate limiting  
- Add cancellation  
- Add visited cache  
- Crawl 50 pages concurrently  

### What is a crawler?
Fetches content from urls, discover new links and repeats this process at scale.

URL → fetch page → extract links → enqueue new URLs → repeat

This i sthe format of data we will get from each page.
```go
type CrawlResult struct {
	URL          string
	StatusCode  int
	FetchedAt   time.Time
	ContentType string
	Body        []byte

	OutLinks    []string
	ContentHash string
}
```


Lets try to see what we want from it and how it will work:
(This is for only when the url is checked and no suburls and found and added)
```go
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
```
- This creare 50 workers all reading onto the jobs channel
- these workers only stop either after completing their job, if the jobs channel is empty or when the context is done
- We ensure no race condition using mutext by locking the visited map so no url is worked on by two workers at the same time.
- All workers are monitered by a waitgroup.

Now lets create a worker that actually fetches the url and goes through to find suburls.

https://chatgpt.com/c/694ad2f7-a528-8321-b36b-4c352a359fea

```go
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
```