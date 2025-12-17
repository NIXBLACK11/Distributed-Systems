# Set3

## Goal:
- Build fan-out worker system  
- Build fan-in aggregator  
- Build 3-stage pipeline  
- Add error channel

- Write select with timeout  
- Add quit channel  
- Build a multi-priority worker using select  
- Implement graceful shutdown   

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