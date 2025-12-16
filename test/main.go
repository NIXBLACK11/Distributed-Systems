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