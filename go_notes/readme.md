### This folder has differnet things I came across that might help or just a place i can store things to remember not necessarily in a order:

- ## Race flag

When running code if you want to check if an race condition is encountered, but the output seems just fine try running using:
```
go run -race main.go
```

This is not a static analysis tool but a tool that checks the memory at runtime and points to parts of code that access a shared memory at the same time(race).

- ## Little bit about contexts

A context is a lightweight object that lets us:
- Cancel work around goruotines
- Lets us cancel using deadlines and timeouts

we have:
- context.Background()
An empty root context, never cancels.

- context.WithCancel()
Can be cancelled manually

- context.Timeout(parent, duration)
same as withCancel
but cancels after the timeout is reached what ever happens first

- context.Deadline(parent, duration)
same as withCancel
but cancels after the deadline is reached what ever happens first