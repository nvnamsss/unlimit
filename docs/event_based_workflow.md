# Event-Based Workflow Architecture

## Overview

The Event-Based Workflow is a flexible, event-driven task processing system that supports both push (send) and pull patterns for task ingestion, with coordinated worker management and state updates.

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                     EventBasedWorkflow                          │
│                                                                 │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │                    EventQueue                            │  │
│  │                                                          │  │
│  │  ┌────────────────┐         ┌──────────────────┐       │  │
│  │  │  TaskPuller    │────────▶│  Internal Queue  │       │  │
│  │  │  (Pull Tasks)  │         │    (Channel)     │       │  │
│  │  └────────────────┘         └──────────────────┘       │  │
│  │         ▲                            │                  │  │
│  │         │ Periodic Pull              │ Subscribe()      │  │
│  │         │ (Timer-based)              ▼                  │  │
│  │         │                    ┌──────────────────┐       │  │
│  │  ┌──────────────┐            │  Task Channel    │       │  │
│  │  │  TaskSender  │            └──────────────────┘       │  │
│  │  │ Send(task)   │                    │                  │  │
│  │  │ SendBatch()  │                    │                  │  │
│  │  └──────────────┘                    │                  │  │
│  │         ▲                            │                  │  │
│  └─────────┼────────────────────────────┼──────────────────┘  │
│            │                            │                     │
│      External Submit              processEvents()             │
│            │                            │                     │
│            │                            ▼                     │
│  ┌─────────┴────────────────────────────────────────────────┐ │
│  │              EventWorkerManager                          │ │
│  │                                                          │ │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │ │
│  │  │   Worker 1   │  │   Worker 2   │  │   Worker N   │  │ │
│  │  │              │  │              │  │              │  │ │
│  │  │  Process()   │  │  Process()   │  │  Process()   │  │ │
│  │  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘  │ │
│  │         │                 │                 │           │ │
│  │         └─────────────────┼─────────────────┘           │ │
│  │                           │                             │ │
│  │                           ▼                             │ │
│  │                  ┌─────────────────┐                    │ │
│  │                  │  StateUpdater   │                    │ │
│  │                  │   Update()      │                    │ │
│  │                  └────────┬────────┘                    │ │
│  │                           │                             │ │
│  │                           ▼                             │ │
│  │                  ┌─────────────────┐                    │ │
│  │                  │  Return Queue   │  (Optional)        │ │
│  │                  │    Send()       │                    │ │
│  │                  └─────────────────┘                    │ │
│  └──────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

## Component Flow

### 1. Task Ingestion (Two Methods)

#### Pull Method (Automatic)
```
External Source → TaskPuller.Pull() → EventQueue → Subscribe Channel
     │                    ▲
     │                    │
     └─── Periodic Fetch ─┘
          (Timer-based)
```

#### Send Method (Manual)
```
Application → TaskSender.Send() → EventQueue → Subscribe Channel
     │              │
     │              └── SendBatch() for bulk operations
     │
     └── Direct task submission
```

### 2. Task Processing Pipeline

```
EventQueue.Subscribe()
     │
     ▼
Task Channel
     │
     ▼
EventWorkerManager.Assign()
     │
     ▼
Worker Pool (Concurrent)
     │
     ├─── Worker 1: Process(task)
     ├─── Worker 2: Process(task)
     └─── Worker N: Process(task)
     │
     ▼
StateUpdater.Update(task)
     │
     ▼
[Optional] Return to Queue
```

## Interfaces

### Core Interfaces

#### EventTask
```go
type EventTask interface {
    Task
    GetState() string
    SetState(state string)
}
```

#### TaskPuller[T EventTask]
```go
type TaskPuller[T EventTask] interface {
    Pull(ctx context.Context) ([]T, error)
}
```

#### TaskSender[T EventTask]
```go
type TaskSender[T EventTask] interface {
    Send(ctx context.Context, task T) error
    SendBatch(ctx context.Context, tasks []T) error
}
```

#### EventQueue[T EventTask]
```go
type EventQueue[T EventTask] interface {
    TaskPuller[T]
    TaskSender[T]
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Subscribe() <-chan T
}
```

#### EventWorker[T EventTask]
```go
type EventWorker[T EventTask] interface {
    Process(ctx context.Context, task T) error
    GetID() string
}
```

#### EventWorkerManager[T EventTask]
```go
type EventWorkerManager[T EventTask] interface {
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Assign(ctx context.Context, task T) error
}
```

#### StateUpdater[T EventTask]
```go
type StateUpdater[T EventTask] interface {
    Update(ctx context.Context, task T) error
}
```

## Execution Flow Sequence

```
┌───────┐     ┌──────────┐     ┌─────────────┐     ┌────────┐     ┌──────────┐
│ Start │────▶│  Queue   │────▶│   Manager   │────▶│ Worker │────▶│ Updater  │
└───────┘     │  Start() │     │   Start()   │     │Process │     │ Update() │
              └─────┬────┘     └──────┬──────┘     └────┬───┘     └─────┬────┘
                    │                 │                 │               │
                    ▼                 ▼                 ▼               ▼
              ┌─────────┐       ┌──────────┐     ┌──────────┐   ┌──────────┐
              │ Puller  │       │  Assign  │     │  Task    │   │  State   │
              │ Starts  │       │  Tasks   │     │ Complete │   │ Updated  │
              └─────────┘       └──────────┘     └──────────┘   └─────┬────┘
                    │                                                   │
                    └───────────────────────────────────────────────────┘
                                    [Optional Return to Queue]
```

## Usage Patterns

### Pattern 1: Pull-Based Processing
```
Database/Queue → Puller → EventQueue → Workers → StateUpdater → Database
                   ▲                                               │
                   └───────────────── Periodic Poll ──────────────┘
```

### Pattern 2: Push-Based Processing
```
API/Service → Send() → EventQueue → Workers → StateUpdater → Response
```

### Pattern 3: Hybrid with Return Queue
```
External → EventQueue → Workers → StateUpdater → Return Queue → Next Stage
    │                                                    │
    └────────────── Feedback Loop for Retry ────────────┘
```

## Key Features

### 1. Dual Ingestion Model
- **Pull**: Automatic periodic fetching from external sources
- **Send**: Manual submission via Send() or SendBatch()

### 2. Concurrent Processing
- Multiple workers process tasks in parallel
- Configurable worker pool size
- Work stealing via shared channel

### 3. State Management
- Tasks maintain state throughout processing
- StateUpdater handles persistence
- Optional return queue for multi-stage workflows

### 4. Graceful Shutdown
- Context-based cancellation
- Wait for in-flight tasks
- Clean resource cleanup

### 5. Flexibility
- Generic type support for any EventTask
- Pluggable components (workers, updaters, queues)
- Optional return queue for complex workflows

## Example Use Cases

### 1. Order Processing System
```
Order DB → Pull Orders → Process Payment → Update Status → Return Queue
                            ↓
                      Inventory Update
                            ↓
                      Notification Send
```

### 2. Data Pipeline
```
API → Send Raw Data → Transform → Validate → Save → Next Stage
```

### 3. Message Processing
```
Message Queue → Pull Messages → Process → Update State → ACK/NACK
```

### 4. Background Job System
```
Job Queue → Pull Jobs → Execute → Update Progress → Complete/Retry
```

## Implementation Components

### BasicEventQueue
- Manages internal channel-based queue
- Periodic pulling with configurable interval
- Thread-safe send operations
- Subscription-based consumption

### BasicEventWorkerManager
- Pool of concurrent workers
- Task assignment via channel
- Automatic state updates
- Optional return queue routing

### State Flow
```
pending → assigned → processing → processed → [returned/completed]
   ↓         ↓           ↓            ↓              ↓
[Initial] [Queued]  [Working]   [Updated]      [Final State]
```

## Configuration Parameters

| Parameter | Description | Typical Values |
|-----------|-------------|----------------|
| maxQueueSize | Internal queue buffer size | 100-1000 |
| pullInterval | Time between pulls | 1s-60s |
| numWorkers | Concurrent worker count | 1-100 |
| taskQueueSize | Worker manager queue size | 10-100 |

## Thread Safety

- All operations are thread-safe
- Mutex protection for shared state
- Channel-based communication
- Context-based cancellation

## Error Handling

- Errors stored in task via SetError()
- Pull errors logged and continued
- Worker errors captured but don't stop workflow
- StateUpdater errors prevent return queue submission
