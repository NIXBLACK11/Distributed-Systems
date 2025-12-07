### This folder has differnet things I came across that might help or just a place i can store things to remember not necessarily in a order:

- ## Race flag

When running code if you want to check if an race condition is encountered, but the output seems just fine try running using:
```
go run -race main.go
```

This is not a static analysis tool but a tool that checks the memory at runtime and points to parts of code that access a shared memory at the same time(race).

- ## M:N scheduler

Go runtime uses an M:N scheduler, this means:
- M OS threads SERVE N goroutines

The scheduler maps multiple goroutines into a small pool of OS threads.

This is when gomaxprocs comes into picture