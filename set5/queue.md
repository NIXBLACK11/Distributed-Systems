## Queue

API → QUEUE → Worker → DB

Queues will help us:
- Smooth traffic spikes
- Retry failed work
- Survive worker crashes
- Scale consumers independently

### Visibility Timeout
What problem does it solve?
Imagine:

1. Worker receives a message
2. Starts processing
3. Crashes halfway

Without protection → message is lost
Or worse → processed twice concurrently

### SQS SOlution: Visibility TImeout
Message becomes invisible to other consumers
for N seconds (visibility timeout)

If the worker deletes the message -> success
If not meaning the worker crashed > message becomes visible again

### Let's build a simple in-memory Queue
```go
package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"slices"
)

type Message struct {
	ID        string
	Body      string
	VisibleAt time.Time
	Attempts  int
}

type Queue struct {
	mu       sync.Mutex
	messages []*Message
}

func (q *Queue) Enqueue(body string) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.messages = append(q.messages, &Message{
		ID:        uuid.NewString(),
		Body:      body,
		VisibleAt: time.Now(),
	})
}

func (q *Queue) Receive(visTimeout time.Duration) *Message {
	q.mu.Lock()
	defer q.mu.Unlock()

	now := time.Now()

	for _, m := range q.messages {
		if now.After(m.VisibleAt) {
			m.VisibleAt = now.Add(visTimeout)
			m.Attempts++
			return m
		}
	}
	return nil
}

func (q *Queue) Delete(id string) {
	q.mu.Lock()
	defer q.mu.Unlock()

	for i, m := range q.messages {
		if m.ID == id {
			q.messages = slices.Delete(q.messages, i, i+1)
			return
		}
	}
}

func main() {
	queue := &Queue{}

	// Enqueue some messages
	fmt.Println("Enqueueing messages...")
	queue.Enqueue("Hello, World!")
	queue.Enqueue("Second message")
	queue.Enqueue("Third message")

	// Receive messages with a 5-second visibility timeout
	fmt.Println("\nReceiving messages...")
	for i := 0; i < 3; i++ {
		msg := queue.Receive(5 * time.Second)
		if msg != nil {
			fmt.Printf("Received message: ID=%s, Body=%s, Attempts=%d\n",
				msg.ID, msg.Body, msg.Attempts)

			// Simulate processing the message
			time.Sleep(1 * time.Second)

			// Delete the message after processing
			queue.Delete(msg.ID)
			fmt.Printf("Deleted message: %s\n", msg.ID)
		} else {
			fmt.Println("No messages available")
		}
	}

	// Try to receive when queue is empty
	fmt.Println("\nTrying to receive from empty queue...")
	msg := queue.Receive(5 * time.Second)
	if msg == nil {
		fmt.Println("No messages available")
	}

	// Demonstrate visibility timeout
	fmt.Println("\nDemonstrating visibility timeout...")
	queue.Enqueue("Timeout test message")

	// Receive message but don't delete it
	msg = queue.Receive(2 * time.Second)
	if msg != nil {
		fmt.Printf("Received message: %s (will be invisible for 2 seconds)\n", msg.Body)
	}

	// Try to receive immediately - should return nil due to visibility timeout
	msg = queue.Receive(2 * time.Second)
	if msg == nil {
		fmt.Println("Message is still invisible due to timeout")
	}

	// Wait for visibility timeout to expire
	fmt.Println("Waiting 3 seconds for visibility timeout to expire...")
	time.Sleep(3 * time.Second)

	// Now we should be able to receive the message again
	msg = queue.Receive(2 * time.Second)
	if msg != nil {
		fmt.Printf("Message is visible again: %s (Attempts: %d)\n", msg.Body, msg.Attempts)
		queue.Delete(msg.ID)
	}

	fmt.Println("\nQueue demo completed!")
}
```