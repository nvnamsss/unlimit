# Trigger System Architecture

## Overview

The trigger system is an event-driven architecture inspired by the Warcraft III World Editor trigger system. It provides a flexible way to respond to events with configurable conditions and actions. The system consists of five core components that work together to create reactive, rule-based behavior.

## Core Components

### 1. Event
**Location:** `trigger/event.go`

**Purpose:** Represents something that has happened in the system.

**Definition:**
```go
type Event interface {
    Type() string           // Event type identifier
    Data() interface{}      // Event payload
    Timestamp() time.Time   // When it occurred
    Source() string         // Where it came from
}
```

**Key Concepts:**
- Events are **immutable** snapshots of occurrences
- Each event has a **type** (e.g., "user", "system", "timer")
- Events carry **data** (payload) relevant to what happened
- Events are automatically **timestamped** when created
- Events track their **source** for auditing and debugging

**Predefined Event Types:**
- `EventTypeUser` - User actions (login, logout, etc.)
- `EventTypeSystem` - System-level events (errors, warnings)
- `EventTypeTimer` - Time-based events
- `EventTypeCustom` - User-defined events
- `EventTypeLifecycle` - Application lifecycle (startup, shutdown)
- `EventTypeValidation` - Validation results
- `EventTypeWorkflow` - State machine transitions

**Type-Safe Events:**
```go
type TypedEvent[T any] struct {
    Event
    typedData T
}
```
Generic wrapper providing compile-time type safety for event data.

---

### 2. Condition
**Location:** `trigger/condition.go`

**Purpose:** Determines whether a trigger should execute for a given event.

**Definition:**
```go
type Condition func(event Event) bool
```

**Key Concepts:**
- Conditions are **predicate functions** that return true/false
- Multiple conditions on a trigger create an **AND** relationship (all must pass)
- Conditions receive the **full event** for inspection
- Used for **filtering** which events activate a trigger

**Type-Safe Conditions:**
```go
type TypedCondition[T any] func(data T) bool
```
Works directly with typed data instead of Event interface.

**Common Condition Patterns:**
```go
// Check event type
func(event Event) bool {
    return event.Type() == EventTypeUser
}

// Check specific field
func(event Event) bool {
    if userEvent, ok := event.(*UserEvent); ok {
        return userEvent.Action == "login"
    }
    return false
}

// Check multiple criteria
func(event Event) bool {
    if systemEvent, ok := event.(*SystemEvent); ok {
        return systemEvent.Level == "critical" && systemEvent.Code >= 500
    }
    return false
}
```

---

### 3. Action
**Location:** `trigger/action.go`

**Purpose:** Defines what happens when a trigger fires.

**Definition:**
```go
type Action func(event Event) error
```

**Key Concepts:**
- Actions are **side-effect functions** executed when conditions pass
- Multiple actions execute **sequentially** in the order they were added
- Actions can **return errors** to halt trigger execution
- Actions receive the **triggering event** for context

**Type-Safe Actions:**
```go
type TypedAction[T any] func(data T) error
```
Works directly with typed data, avoiding type assertions.

**Common Action Patterns:**
```go
// Logging
func(event Event) error {
    log.Printf("Event occurred: %v", event)
    return nil
}

// Data processing
func(event Event) error {
    userEvent := event.(*UserEvent)
    return processUserLogin(userEvent.UserID)
}

// External notifications
func(event Event) error {
    return sendEmail("admin@example.com", event.Data())
}

// State changes
func(event Event) error {
    return updateDatabase(event)
}
```

---

### 4. Trigger
**Location:** `trigger/trigger.go`

**Purpose:** Binds events to conditions and actions.

**Definition:**
```go
type Trigger struct {
    ID          string        // Unique identifier
    Name        string        // Human-readable name
    Description string        // Optional description
    EventTypes  []string      // Which event types to listen for
    Conditions  []Condition   // Filters (all must pass)
    Actions     []Action      // What to do (executed in order)
    Priority    Priority      // Execution order (Critical > High > Normal > Low)
    Enabled     bool          // Can be disabled without removal
    ExecuteOnce bool          // Run only the first time
    executed    bool          // Tracks execution state
}
```

**Key Concepts:**
- Triggers are **declarative rules**: "When X happens, if Y is true, do Z"
- Triggers subscribe to **one or more event types**
- All conditions must pass for actions to execute (**AND logic**)
- Actions execute in **registration order**
- Triggers support **priorities** for execution ordering
- Triggers can be **enabled/disabled** dynamically
- Triggers support **one-shot execution** mode

**Execution Flow:**
```
Event arrives → Check event type match → Check enabled → Check executed (if ExecuteOnce)
     ↓
Evaluate all conditions (AND)
     ↓
All pass? → Execute all actions sequentially → Mark executed (if ExecuteOnce)
     ↓
Any fail? → Skip trigger
```

**Type-Safe Triggers:**
```go
type TypedTrigger[T any] struct {
    trigger         *Trigger
    typedConditions []TypedCondition[T]
    typedActions    []TypedAction[T]
}
```
Wrapper providing type safety while maintaining compatibility with the manager.

---

### 5. TriggerManager
**Location:** `trigger/manager.go`

**Purpose:** Central orchestrator for the entire trigger system.

**Definition:**
```go
type TriggerManager struct {
    triggers       map[string]*Trigger      // Registered triggers
    eventListeners map[string][]EventListener  // Generic listeners
    eventQueue     chan Event                  // Async event queue
    done           chan struct{}               // Shutdown signal
    wg             sync.WaitGroup              // Goroutine tracking
    mutex          sync.RWMutex                // Thread safety
    running        bool                        // State flag
    ctx            context.Context             // Cancellation
    cancel         context.CancelFunc          // Cancel function
}
```

**Key Responsibilities:**

1. **Trigger Registration**
   - Register/unregister triggers
   - Validate unique trigger IDs
   - Manage trigger lifecycle

2. **Event Processing**
   - Receive events via `FireEvent()` (async) or `FireEventSync()` (sync)
   - Queue events for processing
   - Match events to triggers
   - Execute matching triggers by priority

3. **Concurrency Management**
   - Thread-safe trigger registration
   - Async event processing loop
   - Graceful shutdown handling

4. **Lifecycle Management**
   - `Start()` - Begin event processing
   - `Stop()` - Graceful shutdown
   - Context-based cancellation

**Error Handling:**
```go
var (
    ErrTriggerAlreadyExists   // Duplicate trigger ID
    ErrTriggerNotFound        // Missing trigger
    ErrManagerShuttingDown    // Manager stopping
    ErrEventQueueFull         // Queue at capacity
    ErrManagerAlreadyRunning  // Already started
    ErrManagerNotRunning      // Not started
    ErrTriggerExecutionFailed // Trigger execution error
    ErrEventListenerFailed    // Listener error
)
```

---

## Component Relationships

### Architectural Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                      TriggerManager                         │
│  - Registers triggers                                       │
│  - Receives events                                          │
│  - Matches events to triggers                               │
│  - Executes triggers by priority                            │
└─────────────────────────────────────────────────────────────┘
                            │
                            │ manages
                            ▼
                  ┌──────────────────┐
                  │     Trigger      │
                  │  - ID, Name      │
                  │  - Priority      │
                  │  - Enabled       │
                  └──────────────────┘
                            │
            ┌───────────────┼───────────────┐
            │               │               │
        listens to      filters by     executes
            │               │               │
            ▼               ▼               ▼
      ┌─────────┐    ┌──────────┐    ┌─────────┐
      │  Event  │    │Condition │    │ Action  │
      │ - Type  │    │ (filter) │    │ (do)    │
      │ - Data  │    │  bool    │    │ error   │
      │ - Time  │    └──────────┘    └─────────┘
      │ - Source│
      └─────────┘
```

### Data Flow

```
1. Event Creation
   ┌────────────────┐
   │ System creates │
   │ Event object   │
   └────────┬───────┘
            │
            ▼
2. Event Firing
   ┌────────────────────┐
   │ FireEvent() or     │
   │ FireEventSync()    │
   └────────┬───────────┘
            │
            ▼
3. Event Queuing (async mode)
   ┌────────────────────┐
   │ Event added to     │
   │ eventQueue channel │
   └────────┬───────────┘
            │
            ▼
4. Event Processing Loop
   ┌────────────────────┐
   │ Manager reads from │
   │ queue continuously │
   └────────┬───────────┘
            │
            ▼
5. Trigger Matching
   ┌─────────────────────────────┐
   │ For each registered trigger:│
   │ - Check event type match    │
   │ - Check enabled status      │
   │ - Check ExecuteOnce state   │
   └─────────┬───────────────────┘
             │
             ▼
6. Condition Evaluation
   ┌─────────────────────────────┐
   │ For matching triggers:      │
   │ - Evaluate all conditions   │
   │ - ALL must return true      │
   └─────────┬───────────────────┘
             │
             ▼
7. Priority Sorting
   ┌─────────────────────────────┐
   │ Sort passing triggers by:   │
   │ Critical > High > Normal    │
   │ > Low                       │
   └─────────┬───────────────────┘
             │
             ▼
8. Action Execution
   ┌─────────────────────────────┐
   │ For each trigger (by order):│
   │ - Execute actions in order  │
   │ - Stop on first error       │
   │ - Mark executed if once     │
   └─────────────────────────────┘
```

### Interaction Patterns

#### 1. Simple Trigger Pattern
```go
// Event → Condition → Action
trigger := NewTrigger("id", "name").
    AddEventType(EventTypeUser).           // Listen for user events
    AddCondition(func(e Event) bool {      // Filter: only login
        return e.(*UserEvent).Action == "login"
    }).
    AddAction(func(e Event) error {        // Do: log the login
        log.Printf("User logged in")
        return nil
    })
```

#### 2. Multi-Condition Pattern
```go
// Event → Condition1 AND Condition2 → Action
trigger := NewTrigger("id", "name").
    AddEventType(EventTypeSystem).
    AddCondition(func(e Event) bool {      // Must be critical
        return e.(*SystemEvent).Level == "critical"
    }).
    AddCondition(func(e Event) bool {      // Must be error code 500+
        return e.(*SystemEvent).Code >= 500
    }).
    AddAction(sendAlert)
```

#### 3. Multi-Action Pattern
```go
// Event → Condition → Action1 → Action2 → Action3
trigger := NewTrigger("id", "name").
    AddEventType(EventTypeUser).
    AddCondition(isVIPUser).
    AddAction(logEvent).                   // Execute first
    AddAction(sendWelcomeEmail).           // Then this
    AddAction(updateStatistics)            // Finally this
```

#### 4. Priority Pattern
```go
// Multiple triggers respond to same event, ordered by priority
criticalTrigger := NewTrigger("crit", "Critical Handler").
    SetPriority(PriorityCritical).         // Executes first
    AddEventType(EventTypeSystem).
    AddAction(handleCritical)

normalTrigger := NewTrigger("norm", "Normal Handler").
    SetPriority(PriorityNormal).           // Executes after critical
    AddEventType(EventTypeSystem).
    AddAction(handleNormal)
```

#### 5. One-Shot Pattern
```go
// Trigger executes only once, then becomes inactive
initTrigger := NewTrigger("init", "Initialization").
    SetExecuteOnce(true).                  // Only first time
    AddEventType(EventTypeLifecycle).
    AddCondition(isStartup).
    AddAction(initializeSystem)

// Can be reset later
initTrigger.Reset()                        // Allow execution again
```

---

## Type Safety Features

### Generic Type System

The trigger system supports both untyped (interface-based) and typed (generic-based) approaches:

```go
// Untyped approach (runtime type assertion)
trigger := NewTrigger("id", "name").
    AddCondition(func(e Event) bool {
        userEvent, ok := e.(*UserEvent)    // Runtime check
        if !ok {
            return false
        }
        return userEvent.Action == "login"
    })

// Typed approach (compile-time safety)
trigger := NewTypedTrigger[UserEventData]("id", "name").
    AddTypedCondition(func(data UserEventData) bool {
        return data.Action == "login"      // No type assertion needed
    })
```

### Benefits of Typed Approach

1. **Compile-time safety** - Invalid field access caught at compile time
2. **Better IDE support** - Autocomplete for data fields
3. **No type assertions** - Cleaner, more readable code
4. **Type documentation** - Function signatures reveal expected data
5. **Refactoring safety** - Compiler catches breaking changes

### Hybrid Compatibility

Typed triggers are compatible with the manager:

```go
manager := NewTriggerManager()

// Regular trigger
regularTrigger := NewTrigger("reg", "Regular")
manager.RegisterTrigger(regularTrigger)

// Typed trigger (automatically unwrapped)
typedTrigger := NewTypedTrigger[UserEventData]("typed", "Typed")
manager.RegisterTypedTrigger(typedTrigger)  // Works seamlessly
```

---

## Concurrency & Thread Safety

### Thread-Safe Operations

The TriggerManager uses `sync.RWMutex` for thread safety:

- **Read operations** (ListTriggers, GetTrigger) use `RLock()`
- **Write operations** (RegisterTrigger, UnregisterTrigger) use `Lock()`
- **Event processing** acquires read lock during trigger matching

### Async vs Sync Execution

```go
// Async: Event added to queue, returns immediately
manager.FireEvent(event)
// Non-blocking, returns quickly
// Events processed by background goroutine

// Sync: Event processed immediately, waits for completion
manager.FireEventSync(event)
// Blocking, waits for all triggers to execute
// Returns when all actions complete
```

### Graceful Shutdown

```go
manager.Start()
// ... application runs ...

// Graceful shutdown
manager.Stop()
// - Closes event queue
// - Waits for pending events to complete (WaitGroup)
// - Cancels timer events (Context)
// - Clean exit
```

---

## Best Practices

### 1. Trigger Design

✅ **DO:**
- Use descriptive IDs and names
- Keep conditions simple and focused
- Make actions idempotent when possible
- Use appropriate priorities
- Handle errors in actions

❌ **DON'T:**
- Create circular trigger chains
- Block in actions (use async operations)
- Mutate event data
- Ignore action errors
- Create very complex conditions

### 2. Event Design

✅ **DO:**
- Include all relevant data in events
- Use consistent event types
- Timestamp events at creation
- Track event sources
- Use typed events when possible

❌ **DON'T:**
- Include mutable references
- Create overly large event payloads
- Reuse event instances
- Modify events after creation

### 3. Manager Usage

✅ **DO:**
- Start manager before firing events
- Stop manager during shutdown
- Use FireEventSync for critical operations
- Monitor queue size via Stats()
- Handle registration errors

❌ **DON'T:**
- Fire events before Start()
- Ignore queue full errors
- Register duplicate trigger IDs
- Block the event processing loop

### 4. Performance Considerations

- **Queue size**: Default 1000 events, tune based on load
- **Condition complexity**: Simple conditions are faster
- **Action duration**: Long-running actions block processing
- **Trigger count**: More triggers = slower matching
- **Priority usage**: Only when execution order matters

---

## Common Use Cases

### 1. User Activity Tracking
```go
trigger := NewTrigger("track-login", "Login Tracker").
    AddEventType(EventTypeUser).
    AddCondition(isLoginAction).
    AddAction(recordLoginTime).
    AddAction(updateUserStats)
```

### 2. Error Monitoring & Alerting
```go
trigger := NewTrigger("alert-critical", "Critical Alert").
    AddEventType(EventTypeSystem).
    SetPriority(PriorityCritical).
    AddCondition(isCriticalError).
    AddAction(sendPagerDutyAlert).
    AddAction(logToErrorTracker)
```

### 3. Workflow Orchestration
```go
trigger := NewTrigger("order-complete", "Order Completion").
    AddEventType(EventTypeWorkflow).
    AddCondition(isOrderPaid).
    AddAction(sendConfirmationEmail).
    AddAction(scheduleShipment).
    AddAction(updateInventory)
```

### 4. Application Lifecycle
```go
trigger := NewTrigger("app-startup", "Startup Tasks").
    AddEventType(EventTypeLifecycle).
    SetExecuteOnce(true).
    AddCondition(isStartupPhase).
    AddAction(initializeDatabase).
    AddAction(loadConfiguration).
    AddAction(startBackgroundJobs)
```

### 5. Validation & Business Rules
```go
trigger := NewTrigger("validate-order", "Order Validation").
    AddEventType(EventTypeValidation).
    AddCondition(isOrderValidation).
    AddCondition(hasErrors).
    AddAction(notifyUser).
    AddAction(logValidationFailure)
```

---

## Summary

The trigger system provides a **declarative, event-driven architecture** with five interconnected components:

1. **Event** - What happened (immutable data)
2. **Condition** - Should we respond? (filter logic)
3. **Action** - What to do (side effects)
4. **Trigger** - Binding of event → conditions → actions
5. **Manager** - Central orchestrator and executor

This architecture enables:
- ✅ **Loose coupling** - Components communicate via events
- ✅ **Flexibility** - Add/remove triggers without code changes
- ✅ **Testability** - Each component testable in isolation
- ✅ **Scalability** - Async processing with priority queuing
- ✅ **Type safety** - Optional generic types for compile-time checks
- ✅ **Maintainability** - Clear separation of concerns

The system is **inspired by game engines** but designed for **production applications** requiring reactive, rule-based behavior.
