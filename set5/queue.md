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