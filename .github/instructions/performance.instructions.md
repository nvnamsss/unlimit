---
description: Load when implementing or reviewing performance-sensitive paths, concurrency primitives, rate limiting, pipeline batching, or benchmark instrumentation in this repository.
# applyTo: '**/*.{go,md}'
---

<!-- Tip: Use /create-instructions in chat to generate content with agent assistance -->

## Performance Guardrails

Use these rules for any code that can affect throughput, latency, concurrency safety, or benchmark validity.

### 1) Do not keep runtime configuration thread-safe
- The config struct is immutable, it is initialized once and shared across goroutines without locks.
- If dynamic configuration is needed, use a separate mechanism to reload and atomically swap the entire config struct.

### 2) Use channels for concurrency control, not mutexes
- For coordinating work between goroutines, prefer channels to mutexes for better composability and to avoid deadlocks.
- Use buffered channels to implement worker pools, rate limiters, or task queues.