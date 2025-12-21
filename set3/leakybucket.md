## Token bucket

### Core Idea:
Requests are accepted, but proccessed at a steady rate like a leak.

### Concept:
No bursts allowed.
Which means fixed output rate.

## Code:
```go
package main

import (
	"fmt"
	"time"
)

type LeakyBucket struct {
	queue chan func()
}

func NewLeakyBucket(rate, capacity int) *LeakyBucket {
	lb := &LeakyBucket{
		queue: make(chan func(), capacity),
	}

	go func() {
		interval := time.Second / time.Duration(rate)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for task := range lb.queue {
			<-ticker.C
			task()
		}
	} ()

	return lb
}

func (lb *LeakyBucket) Submit(task func()) bool {
	select {
	case lb.queue <- task:
		return true
	default:
		return false
	}
}

func main() {
	lb := NewLeakyBucket(
		2,  // 2 tasks per second
		5,  // queue capacity
	)

	for i := range (10) {
		ok := lb.Submit(func() {
			fmt.Printf("Processed job %d at %s\n", i, time.Now().Format(time.StampMilli))
		})

		if !ok {
			fmt.Printf("Job %d dropped (queue full)\n", i)
		}
	}

	// To wait for all to complete
	time.Sleep(6 * time.Second)
}
```

The reason this takes the function and the token the main does the function, is that token bucket tells if we can do the work leaky bucket says it will do the work at the said rate(no bursts).