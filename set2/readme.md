# Set 2

## Goal:
- Learn cancellation propagation 
- Experiment with context.WithTimeout 
- Build nested goroutines and test cancellation 
-  Wrap goroutine pools with context 
- Build a cancellable job processor

- Read about critical sections  
- Use go test -race on a buggy program  
- Fix the race using Mutex  
- Add RWMutex for performance  

## Notes:

### Quick principles
- Context are not used for passing state, they are used to cancellation and deadlines.
- A context can foram a tree where each child context cancels automaticaly when the parent cancels.
- A goroutine does not automatically end when a context is cancelled it must actually listen for Done using select.
- We should always aim to cancel contexts using the cancel function returned to free resources.


### [Nested goroutines and cancellation propogation](nestedGoroutines.md)

### [Worker pool with cancel](workerpoolcancel.md)

### [Concurrency](concurrency.md)