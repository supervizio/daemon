<!-- updated: 2026-02-15T21:30:00Z -->
# Events Package

Infrastructure adapter implementing lifecycle.Publisher interface.

## Role

Provides event bus for broadcasting lifecycle events to subscribers.

## Files

| File | Purpose |
|------|---------|
| `bus.go` | Bus implementation of lifecycle.Publisher |

## Usage

```go
bus := events.NewBus()
defer bus.Close()

// subscribe
ch := bus.Subscribe()
defer bus.Unsubscribe(ch)

// publish
event := lifecycle.NewEvent(lifecycle.TypeProcessStarted, "service started")
bus.Publish(event)

// receive
select {
case e := <-ch:
    fmt.Printf("Event: %s\n", e.Message)
}
```

## Options

| Option | Default | Description |
|--------|---------|-------------|
| `WithBufferSize(n)` | 64 | Subscriber channel buffer size |

## Thread Safety

- All methods are safe for concurrent use
- Publish is non-blocking (drops events if buffer full)
- Unsubscribe is idempotent

## Related

| Package | Relation |
|---------|----------|
| `domain/lifecycle` | Publisher port interface |
| `bootstrap` | Wire DI injection |
