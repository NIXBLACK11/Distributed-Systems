# Set3

## Goal:
- Build fan-out worker system  
- Build fan-in aggregator  
- Build 3-stage pipeline  
- Add error channel

- Write select with timeout  
- Add quit channel

- Backpressure
- Implement token bucket 
- Implement leaky bucket 
- Build a safe job distributor 
- Add retry-with-backoff 

### [FAN-IN and FAN-OUT](faninout.md)

### Three step pipeline

```go
package main

import (
	"context"
	"fmt"
	"time"
)

func generate(ctx context.Context, n int) <-chan int {
	out := make(chan int)

	go func() {
		defer close(out)

		for i := range n {
			select {
			case <-ctx.Done():
				return
			case out <- i:
			}
		}
	}()

	return out
}

func process(ctx context.Context, in <-chan int) <-chan int {
	out := make(chan int)

	go func() {
		defer close(out)

		for val := range in {
			select {
			case <-ctx.Done():
				return
			case out <- val * 2:
				time.Sleep(200 * time.Millisecond)
			}
		}
	}()

	return out
}

// Normally this is used because the other one depends on the input channel to complete which might not be the case always
// func consume(ctx context.Context, in <-chan int) {
// 	for {
// 		select {
// 		case <-ctx.Done():
// 			return
// 		case val, ok := <-in:
// 			if !ok {
// 				return
// 			}
// 			fmt.Println("consumed:", val)
// 		}
// 	}
// }

func consume(ctx context.Context, in <-chan int) {
	for val := range in {
		select {
		case <-ctx.Done():
			return
		default:
			fmt.Println("consumed:", val)
		}
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	stage1 := generate(ctx, 10)
	stage2 := process(ctx, stage1)
	consume(ctx, stage2)
}
```

### Select theory

- Select helps a goroutine wait on multiple channels at once.

```go
select {
	case v := <-ch1:
		// ch1 received
	case ch2 <- x:
		// sent on ch2
	case <-time.After(1 * time.Second):
    	// timeout
	default:
		// nothing ready
}
```

- Blocks until one case is ready.
- if multiple are read chooses one randomly.
- if we have `default` tho, then it becomes non-blocking as the default is executed.
- select is reevaluated at each iteration.
- closed channels are always ready.
```go
ch := make(chan int)
close(ch)

v, ok := <-ch
```

i.e. this always works even when channel is empty as it returns default value and false.

| Channel state   | Receive `<-ch`                | Send `ch <- x`                       |
| --------------- | ----------------------------- | ------------------------------------ |
| Open + has data | Ready                         | Ready (if buffer/receiver available) |
| Open + empty    | Blocks                        | May block                            |
| **Closed**      | **Always ready** (zero value) | ❌ **PANIC**                          |

What timeouts matter?
- time.After returns a channel after the specified time period
- That channel competes in select and runs if no other condition returns true, meaning it can provide other conditions a specified time period + a graceful exit

### Quit channel
- Like we have discussed multiple times a goroutine does not automatically closes we need to signal it to stop/return.
- For this we use channels that can be specifically used for that.
```go
quit := make(chan struct{})
// why i wrote struct{} because zero allocation

func worker(jobs <- chan int, quit <-chan struct{}) {
	for {
		select {
			case job := <- jobs:
				// do some work 
			case <- quit:
				return
		}
	}
}
// to close we can call close quit()
```

### Backpressure

Backpressure is a flow-control mechanism where a system deliberately slows or blocks producers when consumers cannot keep up, preventing unbounded resource usage and ensuring system stability under load.

Simple example that blocks until we have resources:
```go
func worker(id int, jobs <-chan Job) {
	for job := range jobs {
		process(job)
	}
}

func main() {
	jobs := make(chan Job, 100)

	// start consumers FIRST
	for i := 0; i < 5; i++ {
		go worker(i, jobs)
	}

	// producer
	for i := 0; i < 1000; i++ {
		jobs <- Job{ID: i} // blocks when buffer full
	}
}
```

## ```What should a system do when the demand is more than the capacity?```

We will see two ways we can handle these situations.

The reason this exists:
- Traffic is bursty
- Resources being finite

Two philosophies help us control(as far as i know):
1. Token bucket
2. Leaky Bucket

### [Token Bucket](tokenbucket.md)

### [Leaky Bucket](leakybucket.md)

Things left to do

4️⃣ Safe Job Distributor (bounded + cancelable)
❌ Naive worker pool problem

unbounded job submission

no shutdown

goroutine leaks

✅ Design goals

✔ bounded queue
✔ backpressure
✔ graceful shutdown
✔ no goroutine leaks

🏗 Safe Distributor
type Distributor struct {
	jobs chan Job
	wg   sync.WaitGroup
}

func NewDistributor(workers, capacity int) *Distributor {
	d := &Distributor{
		jobs: make(chan Job, capacity),
	}

	for i := 0; i < workers; i++ {
		d.wg.Add(1)
		go d.worker(i)
	}

	return d
}

Worker
func (d *Distributor) worker(id int) {
	defer d.wg.Done()

	for job := range d.jobs {
		job()
	}
}

Submit with backpressure
func (d *Distributor) Submit(job Job) bool {
	select {
	case d.jobs <- job:
		return true
	default:
		return false // queue full → backpressure
	}
}

Graceful shutdown
func (d *Distributor) Shutdown() {
	close(d.jobs)
	d.wg.Wait()
}


💡 This pattern is used everywhere (HTTP servers, queues, schedulers)

5️⃣ Retry with Exponential Backoff
❌ Problem

Immediate retries under failure:

amplify load

DDOS yourself

thundering herd

✅ Backoff strategy

Wait longer after each failure.

1s → 2s → 4s → 8s → cap

🧠 Rules

exponential growth

jitter (randomness)

max retries

context-aware

✅ Go Implementation
func Retry(
	ctx context.Context,
	maxRetries int,
	baseDelay time.Duration,
	fn func() error,
) error {

	var attempt int
	delay := baseDelay

	for {
		err := fn()
		if err == nil {
			return nil
		}

		attempt++
		if attempt >= maxRetries {
			return err
		}

		select {
		case <-time.After(delay):
			delay *= 2
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

🧪 With jitter (production)
jitter := time.Duration(rand.Int63n(int64(delay / 2)))
time.Sleep(delay + jitter)

🔥 How all this fits together
Client
  ↓
Rate limiter (token bucket)
  ↓
Bounded queue (backpressure)
  ↓
Worker pool
  ↓
Retry w/ backoff


This is exactly how real Go services are built.

🧠 Key takeaways (burn these in)

Blocking is control, not failure

Channels ARE queues, rate limiters, and backpressure tools

Token bucket → admission control

Leaky bucket → execution smoothing

Backoff prevents cascading failure
1️⃣ API rate limiter middleware
2️⃣ Worker pool with bounded jobs
3️⃣ Retry wrapper for DB calls
4️⃣ Load test (100k requests)