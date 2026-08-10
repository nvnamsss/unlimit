# Trigger Module

A powerful, event-driven trigger system for Go applications, inspired by the Warcraft III World Editor trigger system. This module provides a flexible way to implement event-driven architecture with conditions, actions, and priorities.

## Features

- **Event-driven architecture**: React to events with customizable triggers
- **Conditional execution**: Add conditions to determine when triggers should fire
- **Priority-based execution**: Control the order of trigger execution
- **Asynchronous and synchronous processing**: Choose between async event queuing or sync processing
- **Multiple event types**: Built-in support for common event types (timer, user, system, lifecycle, etc.)
- **Custom events**: Easy to create custom event types
- **Execute-once triggers**: Triggers that only execute once
- **Event listeners**: General event listeners for cross-cutting concerns
- **Thread-safe**: Designed for concurrent use
- **Comprehensive testing**: Well-tested with examples and benchmarks

## Core Concepts

### Events
Events represent something that happened in your system. All events implement the `Event` interface:

```go
type Event interface {
    Type() string
    Data() interface{}
    Timestamp() time.Time
    Source() string
}
```

### Triggers
Triggers are the core unit that responds to events. A trigger consists of:
- **Event types**: Which events it listens for
- **Conditions**: Optional conditions that must be met
- **Actions**: What to do when the trigger fires
- **Priority**: Execution order relative to other triggers

### Trigger Manager
The `TriggerManager` coordinates event dispatch and trigger execution. It handles:
- Registering and managing triggers
- Processing events (async and sync)
- Managing event queues
- Providing statistics and monitoring

## Quick Start

```go
package main

import (
    "fmt"
    "time"
    "your-module/trigger"
)

func main() {
    // Create and start trigger manager
    manager := trigger.NewTriggerManager()
    manager.Start()
    defer manager.Stop()
    
    // Create a simple trigger
    loginTrigger := trigger.NewTrigger("user-login", "Handle User Login").
        AddEventType(trigger.EventTypeUser).
        AddCondition(func(event trigger.Event) bool {
            if userEvent, ok := event.(*trigger.UserEvent); ok {
                return userEvent.Action == "login"
            }
            return false
        }).
        AddAction(func(event trigger.Event) error {
            userEvent := event.(*trigger.UserEvent)
            fmt.Printf("User %s logged in!\n", userEvent.UserID)
            return nil
        })
    
    // Register the trigger
    manager.RegisterTrigger(loginTrigger)
    
    // Fire an event
    loginEvent := trigger.NewUserEvent("user123", "login", "web-app", "auth-service")
    manager.FireEvent(loginEvent)
    
    time.Sleep(100 * time.Millisecond) // Wait for async processing
}
```

## Built-in Event Types

### UserEvent
Represents user actions:
```go
event := trigger.NewUserEvent("user123", "login", "web-app", "auth-service")
```

### SystemEvent
Represents system-level events:
```go
event := trigger.NewSystemEvent("error", "database connection failed", 500, "db-service")
```

### TimerEvent
Represents timer-based events:
```go
event := trigger.NewTimerEvent(5*time.Second, false, "scheduler")
```

### CustomEvent
For application-specific events:
```go
event := trigger.NewCustomEvent("order_placed", map[string]interface{}{
    "order_id": "12345",
    "amount": 99.99,
}, "order-service")
```

### LifecycleEvent
For application lifecycle events:
```go
event := trigger.NewLifecycleEvent("startup", map[string]interface{}{
    "version": "1.0.0",
}, "system")
```

### WorkflowEvent
For workflow state changes:
```go
event := trigger.NewWorkflowEvent("order-processing", "pending", "approved", 
    map[string]interface{}{"order_id": "12345"}, "workflow-engine")
```

### ValidationEvent
For validation results:
```go
event := trigger.NewValidationEvent(false, []string{"email is required"}, "user-form", "validator")
```

## Advanced Usage

### Priority-based Execution
```go
criticalTrigger := trigger.NewTrigger("critical", "Critical Handler").
    SetPriority(trigger.PriorityCritical).
    AddEventType(trigger.EventTypeSystem).
    AddAction(func(event trigger.Event) error {
        // This will execute before normal priority triggers
        return nil
    })
```

### Execute-once Triggers
```go
initTrigger := trigger.NewTrigger("init", "One-time Setup").
    SetExecuteOnce(true).
    AddEventType(trigger.EventTypeLifecycle).
    AddAction(func(event trigger.Event) error {
        // This will only execute once
        return nil
    })
```

### Timer Events
```go
// One-time timer
manager.CreateTimerEvent("backup", 1*time.Hour, nil, "scheduler")

// Periodic timer
manager.CreatePeriodicEvent("health-check", 30*time.Second, nil, "monitor")
```

### Event Listeners
For cross-cutting concerns:
```go
manager.AddEventListener(trigger.EventTypeUser, func(event trigger.Event) error {
    // Log all user events
    log.Printf("User event: %s", event.Type())
    return nil
})
```

### Chained Triggers
Triggers can fire other events to create workflows:
```go
orderTrigger := trigger.NewTrigger("order-placed", "Order Handler").
    AddEventType("order").
    AddAction(func(event trigger.Event) error {
        // Process order, then trigger email
        emailEvent := trigger.NewCustomEvent("send_email", map[string]interface{}{
            "type": "order_confirmation",
        }, "order-service")
        return manager.FireEvent(emailEvent)
    })
```

### Custom Event Types
```go
type GameEvent struct {
    *trigger.BaseEvent
    PlayerID string
    Action   string
}

func NewGameEvent(playerID, action, source string) *GameEvent {
    return &GameEvent{
        BaseEvent: trigger.NewBaseEvent("game", nil, source),
        PlayerID:  playerID,
        Action:    action,
    }
}
```

## Synchronous vs Asynchronous Processing

### Asynchronous (Default)
```go
manager.FireEvent(event) // Returns immediately, event processed in background
```

### Synchronous
```go
err := manager.FireEventSync(event) // Waits for all triggers to complete
```

## Monitoring and Statistics

```go
stats := manager.Stats()
fmt.Printf("Total triggers: %v\n", stats["total_triggers"])
fmt.Printf("Enabled triggers: %v\n", stats["enabled_triggers"])
fmt.Printf("Event queue size: %v\n", stats["event_queue_size"])
```

## Error Handling

Triggers should return errors from their actions:
```go
trigger.AddAction(func(event trigger.Event) error {
    if err := processEvent(event); err != nil {
        return fmt.Errorf("failed to process event: %w", err)
    }
    return nil
})
```

## Best Practices

1. **Use descriptive trigger IDs and names** for easier debugging
2. **Keep actions lightweight** to avoid blocking the event loop
3. **Use conditions** to filter events efficiently
4. **Set appropriate priorities** for critical triggers
5. **Handle errors gracefully** in trigger actions
6. **Use execute-once triggers** for initialization tasks
7. **Monitor event queue size** to detect performance issues
8. **Prefer async processing** unless you need synchronous guarantees

## Thread Safety

All operations are thread-safe:
- Multiple goroutines can fire events concurrently
- Triggers can be registered/unregistered during operation
- Event processing is handled safely in background goroutines

## Testing

The module includes comprehensive tests. Run them with:
```bash
go test ./trigger/...
```

See `trigger_test.go` for detailed examples and `example.go` for usage patterns.

## Performance

- Event processing is asynchronous by default for high throughput
- Events are queued in a buffered channel (default: 1000 events)
- Triggers are executed in priority order
- Memory usage is optimized for high-frequency events
- Benchmark tests are included for performance validation