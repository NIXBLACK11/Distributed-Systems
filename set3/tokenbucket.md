## Token bucket

### Core Idea:
Not everone is allowed in, but those who are can move fast.

### Concept:
The system produces permission at a steady rate.
With some being at the beggining already.
A task requires persmission to complete.

## Best way we can know is code:
```go
package main

import (
	"fmt"
	"time"
)

type TokenBucket struct {
	tokens chan struct{}
}

func NewTokenBucket(rate, burst int) *TokenBucket {
	tb := &TokenBucket{
		tokens: make(chan struct{}, burst),
	}

	// We initially fill the bucket with the initial permissions
	for _ = range(burst) {
		tb.tokens<-struct{}{}
	}

	go func() {
		// gives the time one will take to do
		interval := time.Second / time.Duration(rate)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			select {
			case tb.tokens <- struct{}{}:
				fmt.Println("Token added")
				// token added at each interval
			default:
				// to keep the select listening to the ticker
			}
		}
	} ()

	return tb
}

func (tb *TokenBucket) Allow() bool {
	select {
	case <- tb.tokens:
		return true
	default:
		return false
	}
}

func main() {
	tb := NewTokenBucket(
		5,  // 5 permissions per second
		10, // 10 permissions burst capacity
	)

	for i := range(30) {
		if tb.Allow() {
			fmt.Printf("Request %d allowed\n", i)
		} else {
			fmt.Printf("Request %d rejected (rate limit)\n", i)	
		}
		time.Sleep(130 * time.Millisecond)
	}
}
```