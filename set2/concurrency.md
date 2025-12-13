## Concurrency

- Concurrency bugs don't crash immediately, They lie.
- Critical section: any piece of code where shared data is read or written concurrently

Go can't protect us automatically.
But we can atleast use -race flag that is our X-ray vision.

Lets test our buggy code first:
```go
package main

import (
	"fmt"
	"sync"
)

func main() {
	var counter int
	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			counter++	
			wg.Done()
		}()
	}

	wg.Wait()
	fmt.Println("Final counter:", counter)
}
```

run using:
```sh
go run main.go
```

then try this:
```sh
go run -race main.go
```

Notice how the second output shows that two goroutines try to access the same memory.
But the program still runs.
And how the value is never 1000

That's a problem.

### But we can fix these problems using mutex.
The allows only one goroutine inside a critical section, and blocks others.

Now lets fix our previous buggy code using mutex.
```go
package main

import (
	"fmt"
	"sync"
)

func main() {
	var counter int
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			mu.Lock()
			counter++
			mu.Unlock()	
			wg.Done()
		}()
	}

	wg.Wait()
	fmt.Println("Final counter:", counter)
}
```

This is an easy example, but these are some discpline we should follow:
- Lock as late as possible.
- Unlock as early as possible.

Now when we rerun this code with the same commands.
There are no errors and the output is correct.

And no race warnings, this means we successfully fixed the problem with mutex.

