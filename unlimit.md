# Unlimit - Baseline Documentation for Agents

**Repository:** github.com/nvnamsss/unlimit  
**Go Version:** 1.24.0  
**Purpose:** A comprehensive Go toolkit providing algorithms, data structures, caching, workflow orchestration, event-driven systems, and utility functions for building scalable applications.

---

## Module Overview

| name | description                                                                                                                                           | import path                      |
| ---- | ----------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------- |
| Algo | Generic algorithms and data structures including LRU/LFU caches, trees, graphs, bitmaps, and specialized structures for high-performance applications | github.com/nvnamsss/unlimit/algo |

### Components

| name                     | description                                                                                                                                          | domain                                                                | implementation                                                                    | tests                                                                                       | benchmarks                                                |
| ------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------- | --------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------- | --------------------------------------------------------- |
| LRUCache                 | Least Recently Used cache with automatic eviction when capacity is reached. Thread-safe implementation with O(1) operations                          | caching, memory management, bounded collections                       | /home/namnv/git/unlimit/algo/lru.go                                               | /home/namnv/git/unlimit/algo/lru_test.go                                                    |                                                           |
| LFU                      | Least Frequently Used cache that evicts items based on access frequency. Tracks usage patterns for intelligent eviction                              | caching, frequency-based eviction, hot data retention                 | /home/namnv/git/unlimit/algo/lfu.go                                               |                                                                                             |                                                           |
| ConsistentHasher         | FNV-1a hash function with bucket mapping support for distributed systems. Provides consistent hashing for load balancing                             | distributed systems, load balancing, sharding, partitioning           | /home/namnv/git/unlimit/algo/consistent_hash.go                                   | /home/namnv/git/unlimit/algo/consistent_hash_test.go                                        |                                                           |
| Deque                    | Double-ended queue supporting O(1) push/pop operations at both ends. Generic implementation with slice and linked list variants                      | queue operations, sliding window, breadth-first search                | /home/namnv/git/unlimit/algo/deque.go, /home/namnv/git/unlimit/algo/deque_list.go | /home/namnv/git/unlimit/algo/deque_test.go, /home/namnv/git/unlimit/algo/deque_list_test.go | /home/namnv/git/unlimit/algo/deque_bench_test.go          |
| Stack                    | LIFO (Last-In-First-Out) data structure with generic type support. SliceStack provides array-based implementation                                    | backtracking, parsing, expression evaluation, DFS                     | /home/namnv/git/unlimit/algo/stack.go                                             | /home/namnv/git/unlimit/algo/stack_test.go                                                  | /home/namnv/git/unlimit/algo/stack_bench_test.go          |
| Queue                    | FIFO (First-In-First-Out) data structure with multiple implementations: SliceQueue (array-based), LockFreeQueue (concurrent), ChannelQueue (bounded) | task queues, breadth-first operations, message passing                | /home/namnv/git/unlimit/algo/queue.go                                             | /home/namnv/git/unlimit/algo/queue_test.go                                                  | /home/namnv/git/unlimit/algo/queue_bench_test.go          |
| Trie                     | Prefix tree with Bloom filter support for efficient string operations. Supports autocomplete and prefix matching                                     | autocomplete, prefix search, dictionary, spell checking               | /home/namnv/git/unlimit/algo/trie.go                                              | /home/namnv/git/unlimit/algo/trie_test.go                                                   | /home/namnv/git/unlimit/algo/trie_bench_test.go           |
| UnionFind                | Disjoint-set data structure with path compression and union by rank. Thread-safe implementation for connectivity queries                             | connectivity, graph algorithms, Kruskal's MST, cycle detection        | /home/namnv/git/unlimit/algo/union_find.go                                        | /home/namnv/git/unlimit/algo/union_find_test.go                                             | /home/namnv/git/unlimit/algo/union_find_bench_test.go     |
| RoaringBitmap            | Compressed bitmap using Roaring format with array and bitmap containers. Space-efficient set operations                                              | set operations, bit manipulation, compression, analytics              | /home/namnv/git/unlimit/algo/roaring_bitmap.go                                    | /home/namnv/git/unlimit/algo/roaring_bitmap_test.go                                         | /home/namnv/git/unlimit/algo/roaring_bitmap_bench_test.go |
| OrderedSet               | Sorted set implementation using Red-Black Tree. Maintains elements in sorted order with O(log n) operations                                          | sorted collections, range queries, leaderboards, priority             | /home/namnv/git/unlimit/algo/ordered_set.go                                       | /home/namnv/git/unlimit/algo/ordered_set_test.go                                            | /home/namnv/git/unlimit/algo/ordered_set_bench_test.go    |
| RedBlackTree             | Self-balancing binary search tree guaranteeing O(log n) operations. Generic key-value store                                                          | ordered maps, interval trees, database indexing                       | /home/namnv/git/unlimit/algo/red_black_tree.go                                    | /home/namnv/git/unlimit/algo/red_black_tree_test.go                                         |                                                           |
| Graph                    | Graph structures with DirectedGraph and UndirectedGraph variants. Support for BFS, DFS, cycle detection, topological sort                            | network analysis, pathfinding, dependency resolution, social networks | /home/namnv/git/unlimit/algo/graph.go                                             | /home/namnv/git/unlimit/algo/graph_test.go                                                  |                                                           |
| LinkedList               | Singly linked list with O(1) insertion at head/tail. Generic implementation                                                                          | sequential access, insertions, memory-efficient collections           | /home/namnv/git/unlimit/algo/linkedlist.go                                        |                                                                                             | /home/namnv/git/unlimit/algo/linkedlist_bench_test.go     |
| ReservoirSampleCollector | Random sampling from streams using Algorithm L. Maintains k random samples with O(k(1+log(n/k))) complexity                                          | statistics, sampling, streaming data, randomization                   | /home/namnv/git/unlimit/algo/reservoir_sampling.go                                | /home/namnv/git/unlimit/algo/reservoir_sampling_test.go                                     |                                                           |
| TopK                     | Top-K elements tracking using Redis backend. Interface for frequent items identification                                                             | trending, analytics, hot items, frequency estimation                  | /home/namnv/git/unlimit/algo/topk.go                                              |                                                                                             |                                                           |

---

## Cache Module

| name  | description                                                                                                                                 | import path                       |
| ----- | ------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------- |
| Cache | Unified caching interface with multiple backend implementations (Redis, Memory, LRU). Supports key-value, hash, bitmap, and list operations | github.com/nvnamsss/unlimit/cache |

### Components

| name                     | description                                                                                                           | domain                                                | implementation                                       | tests                                                     |
| ------------------------ | --------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------- | ---------------------------------------------------- | --------------------------------------------------------- |
| Cache Interface          | Abstract cache interface supporting Set, Get, Hash operations, Bitmap operations, and List operations                 | caching, session storage, distributed caching         | /home/namnv/git/unlimit/cache/cache.go               |                                                           |
| TypedCache               | Generic in-memory cache interface with type safety. Supports Get, Set, Delete operations with generic key-value types | type-safe caching, local storage                      | /home/namnv/git/unlimit/cache/cache.go               |                                                           |
| TypedCacheWithExpiration | Cache with TTL (Time-To-Live) support. Extends TypedCache with expiration capabilities                                | expiring caches, rate limiting, temporary storage     | /home/namnv/git/unlimit/cache/cache.go               |                                                           |
| LRUCache                 | LRU cache implementation backed by algo.LRUCache. Bounded cache with automatic eviction                               | bounded caching, memory management                    | /home/namnv/git/unlimit/cache/lru.go                 |                                                           |
| Hashtable                | Thread-safe hash table with TTL support using sync.Map. Generic key-value store with expiration                       | concurrent access, temporary storage, in-memory cache | /home/namnv/git/unlimit/cache/hashtable.go           | /home/namnv/git/unlimit/cache/hashtable_test.go           |
| redisCache               | Redis backend implementation of Cache interface. Distributed caching with persistence                                 | distributed caching, shared state, scalability        | /home/namnv/git/unlimit/cache/redis.go               |                                                           |
| memoryCache              | In-memory backend using go-cache library. Fast local caching with TTL                                                 | local caching, single-instance applications           | /home/namnv/git/unlimit/cache/memory.go              |                                                           |
| DynamicLFU               | Dynamic capacity LFU cache with automatic size adjustment                                                             | adaptive caching, memory pressure management          | /home/namnv/git/unlimit/cache/dynamic_lfu.go         |                                                           |
| CapacityController       | Interface for dynamic capacity management in caches                                                                   | adaptive systems, resource management                 | /home/namnv/git/unlimit/cache/capacity_controller.go | /home/namnv/git/unlimit/cache/capacity_controller_test.go |

---

## Efflux Module

| name   | description                                                                                                         | import path                        | documentation                          |
| ------ | ------------------------------------------------------------------------------------------------------------------- | ---------------------------------- | -------------------------------------- |
| Efflux | Background job scheduling, worker pools, event-driven workflows, and polling systems for concurrent task processing | github.com/nvnamsss/unlimit/efflux | /home/namnv/git/unlimit/docs/efflux.md |

### Components

| name                  | description                                                                                                                          | domain                                                        | implementation                                                           | tests                                                       | benchmarks | examples                                                       | documentation                                         |
| --------------------- | ------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------- | ------------------------------------------------------------------------ | ----------------------------------------------------------- | ---------- | -------------------------------------------------------------- | ----------------------------------------------------- |
| Job                   | Periodic job scheduler supporting multiple handlers with interval-based execution. Manages background jobs with cancellation support | background jobs, cron-like tasks, periodic execution          | /home/namnv/git/unlimit/efflux/job.go                                    | /home/namnv/git/unlimit/efflux/job_test.go                  |            |                                                                |                                                       |
| SimpleWorkerPool      | Concurrent task processor with fixed worker count. Processes tasks in parallel with graceful shutdown                                | parallel processing, concurrent execution                     | /home/namnv/git/unlimit/efflux/simple_worker_pool.go                     | /home/namnv/git/unlimit/efflux/simple_worker_pool_test.go   |            |                                                                |                                                       |
| WorkerPool            | Generic worker pool with retry logic, error handling, and context cancellation. Supports dynamic worker scaling                      | task queues, retry logic, fault tolerance                     | /home/namnv/git/unlimit/efflux/worker_pool.go (referenced in benchmarks) | /home/namnv/git/unlimit/efflux/worker_pool_test.go          |            |                                                                | /home/namnv/git/unlimit/docs/worker_pool_benchmark.md |
| Pipeline              | Conditional stage-based workflow processor. Chains processing stages with conditional execution                                      | ETL, data pipelines, transformation chains                    | /home/namnv/git/unlimit/efflux/pipeline.go                               | /home/namnv/git/unlimit/efflux/pipeline_test.go             |            |                                                                |                                                       |
| EventBasedWorkflow    | Event-driven task processing system with state transitions and lifecycle hooks                                                       | state machines, event processing, workflow orchestration      | /home/namnv/git/unlimit/efflux/event_based_workflow.go                   | /home/namnv/git/unlimit/efflux/event_based_workflow_test.go |            |                                                                | /home/namnv/git/unlimit/docs/event_based_workflow.md  |
| PollingManager        | Multi-poller manager for continuous background polling. Manages multiple pollers with independent intervals                          | polling, data synchronization, monitoring                     | /home/namnv/git/unlimit/efflux/polling.go                                | /home/namnv/git/unlimit/efflux/polling_test.go              |            | /home/namnv/git/unlimit/efflux/polling_example_test.go         |                                                       |
| OneTimePollingManager | Transient polling for operations requiring polling until completion or timeout. Single-use polling with retry logic                  | one-time operations, waiting for conditions, async completion | /home/namnv/git/unlimit/efflux/onetime_polling.go                        | /home/namnv/git/unlimit/efflux/onetime_polling_test.go      |            | /home/namnv/git/unlimit/efflux/onetime_polling_example_test.go | /home/namnv/git/unlimit/docs/ONETIME_POLLING.md       |
| FlushBuffer           | Auto-flushing buffer with size and time-based triggers. Thread-safe batch accumulation                                               | batch processing, buffering, bulk operations                  | /home/namnv/git/unlimit/efflux/flush_buffer.go                           | /home/namnv/git/unlimit/efflux/flush_buffer_test.go         |            |                                                                |                                                       |
| Task                  | Generic task interface for scheduled execution. Base abstraction for schedulable work                                                | task scheduling, job abstraction                              | /home/namnv/git/unlimit/efflux/task.go                                   |                                                             |            |                                                                |                                                       |
| ProgressTask          | Task with progress tracking and reporting capabilities                                                                               | long-running tasks, progress monitoring                       | /home/namnv/git/unlimit/efflux/progress_task.go                          | /home/namnv/git/unlimit/efflux/progress_task_test.go        |            |                                                                |                                                       |
| DistributedWorkflow   | Distributed workflow execution with coordinator and worker nodes                                                                     | distributed systems, horizontal scaling, load distribution    | /home/namnv/git/unlimit/efflux/distributed_workflow.go                   | /home/namnv/git/unlimit/efflux/distributed_workflow_test.go |            |                                                                |                                                       |

---

## Collections Module

| name        | description                                                                                      | import path                             |
| ----------- | ------------------------------------------------------------------------------------------------ | --------------------------------------- |
| Collections | Design patterns and utilities including iterators, retry mechanisms, and chain of responsibility | github.com/nvnamsss/unlimit/collections |

### Components

| name                              | description                                                                                                 | domain                                                   | implementation                                                 | tests                                                | benchmarks                                                     | examples | documentation |
| --------------------------------- | ----------------------------------------------------------------------------------------------------------- | -------------------------------------------------------- | -------------------------------------------------------------- | ---------------------------------------------------- | -------------------------------------------------------------- | -------- | ------------- |
| Iterator                          | Generic collection iterator with functional methods (Map, Filter, Reduce, ForEach). Lazy evaluation support | data processing, transformations, functional programming | /home/namnv/git/unlimit/collections/iterator.go                | /home/namnv/git/unlimit/collections/iterator_test.go | /home/namnv/git/unlimit/collections/iterator_benchmark_test.go |          |               |
| BatchIterator                     | Iterator for batch processing of collections. Divides data into chunks                                      | bulk operations, batch processing                        | /home/namnv/git/unlimit/collections/iterator.go                | /home/namnv/git/unlimit/collections/iterator_test.go |                                                                |          |               |
| Retrier                           | Retry logic with exponential backoff, jitter, and configurable policies                                     | resilience, error handling, fault tolerance              | /home/namnv/git/unlimit/collections/retrier.go                 | /home/namnv/git/unlimit/collections/retier_test.go   |                                                                |          |               |
| Handler (Chain of Responsibility) | Chain of responsibility pattern implementation. Links handlers for sequential request processing            | request processing chains, middleware, validation chains | /home/namnv/git/unlimit/collections/chain_of_responsibility.go |                                                      |                                                                |          |               |

---

## Database Module

| name     | description                                                                                            | import path                          |
| -------- | ------------------------------------------------------------------------------------------------------ | ------------------------------------ |
| Database | Unified database interfaces for SQL (GORM) and MongoDB with connection pooling and transaction support | github.com/nvnamsss/unlimit/database |

### Components

| name            | description                                                                        | domain                                  | implementation                                                                                                                  | tests | benchmarks | examples | documentation |
| --------------- | ---------------------------------------------------------------------------------- | --------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------- | ----- | ---------- | -------- | ------------- |
| GormDBAdapter   | GORM ORM interface wrapper providing unified database operations for SQL databases | SQL databases, ORM, MySQL, PostgreSQL   | /home/namnv/git/unlimit/database/db.go, /home/namnv/git/unlimit/database/mysql.go, /home/namnv/git/unlimit/database/postgres.go |       |            |          |               |
| MongoDBAdapter  | MongoDB interface providing collection and session management                      | NoSQL, document storage, MongoDB        | /home/namnv/git/unlimit/database/db.go, /home/namnv/git/unlimit/database/mongodb.go                                             |       |            |          |               |
| MongoDatabase   | MongoDB database wrapper with collection access                                    | MongoDB operations, database management | /home/namnv/git/unlimit/database/mongodb.go                                                                                     |       |            |          |               |
| MongoCollection | MongoDB collection wrapper with CRUD operations                                    | document operations, queries            | /home/namnv/git/unlimit/database/mongodb.go                                                                                     |       |            |          |               |
| MongoSession    | MongoDB session management for transactions                                        | MongoDB transactions, ACID operations   | /home/namnv/git/unlimit/database/mongodb.go                                                                                     |       |            |          |               |

---

## Trigger Module

| name    | description                                                                                       | import path                         | documentation                           |
| ------- | ------------------------------------------------------------------------------------------------- | ----------------------------------- | --------------------------------------- |
| Trigger | Warcraft III-inspired trigger system for event-driven programming. Event-condition-action pattern | github.com/nvnamsss/unlimit/trigger | /home/namnv/git/unlimit/docs/trigger.md |

### Components

| name           | description                                                                                          | domain                                              | implementation                                   | tests | benchmarks | examples | documentation |
| -------------- | ---------------------------------------------------------------------------------------------------- | --------------------------------------------------- | ------------------------------------------------ | ----- | ---------- | -------- | ------------- |
| Event          | Event interface with BaseEvent and TypedEvent implementations. Carries event data through the system | event sourcing, reactive systems, publish-subscribe | /home/namnv/git/unlimit/trigger/event.go         |       |            |          |               |
| Condition      | Predicate function for trigger evaluation. Returns true if trigger should fire                       | conditional logic, business rules                   | /home/namnv/git/unlimit/trigger/condition.go     |       |            |          |               |
| Action         | Side-effect function executed by triggers when conditions are met                                    | event handling, side effects                        | /home/namnv/git/unlimit/trigger/action.go        |       |            |          |               |
| Trigger        | Event-condition-action rule. Binds events to conditions and actions                                  | business rules, automation, reactive programming    | /home/namnv/git/unlimit/trigger/trigger.go       |       |            |          |               |
| TriggerManager | Central trigger registry and dispatcher. Manages trigger lifecycle and event routing                 | event management, orchestration                     | /home/namnv/git/unlimit/trigger/manager.go       |       |            |          |               |
| TypedTrigger   | Type-safe trigger wrapper providing compile-time type checking                                       | type safety, event handling                         | /home/namnv/git/unlimit/trigger/typed_trigger.go |       |            |          |               |

---

## Storage Module

| name    | description                                                    | import path                         |
| ------- | -------------------------------------------------------------- | ----------------------------------- |
| Storage | Unified interface for object storage systems (S3, MinIO, etc.) | github.com/nvnamsss/unlimit/storage |

### Components

| name        | description                                                                                                    | domain                                      | implementation                             | tests | benchmarks | examples | documentation |
| ----------- | -------------------------------------------------------------------------------------------------------------- | ------------------------------------------- | ------------------------------------------ | ----- | ---------- | -------- | ------------- |
| Storer      | Storage operations interface with methods for CreateCollection, StoreAsset, GetAsset, GetAssetURL, DeleteAsset | file storage, cloud storage, object storage | /home/namnv/git/unlimit/storage/storage.go |       |            |          |               |
| MinioStorer | MinIO/S3 implementation of Storer interface. Compatible with S3-like object stores                             | object storage, S3, MinIO                   | /home/namnv/git/unlimit/storage/minio.go   |       |            |          |               |

---

## Scheduler Module

| name      | description                                                      | import path                           |
| --------- | ---------------------------------------------------------------- | ------------------------------------- |
| Scheduler | Task scheduler with interval, cron, and one-off schedule support | github.com/nvnamsss/unlimit/scheduler |

### Components

| name      | description                                                                         | domain                                         | implementation                                 | tests | benchmarks | examples | documentation |
| --------- | ----------------------------------------------------------------------------------- | ---------------------------------------------- | ---------------------------------------------- | ----- | ---------- | -------- | ------------- |
| Scheduler | Task scheduler supporting multiple schedule types (interval, cron, oneoff)          | cron jobs, periodic tasks, scheduled execution | /home/namnv/git/unlimit/scheduler/scheduler.go |       |            |          |               |
| Task      | Scheduled task with configurable intervals, cron expressions, or one-time execution | background scheduling, time-based execution    | /home/namnv/git/unlimit/scheduler/task.go      |       |            |          |               |

---

## Utility Module

| name    | description                                                                            | import path                         |
| ------- | -------------------------------------------------------------------------------------- | ----------------------------------- |
| Utility | Helper functions for array operations, type conversion, file utilities, and versioning | github.com/nvnamsss/unlimit/utility |

### Components

| name              | description                                                         | domain                                      | implementation                             | tests | benchmarks | examples | documentation                              |
| ----------------- | ------------------------------------------------------------------- | ------------------------------------------- | ------------------------------------------ | ----- | ---------- | -------- | ------------------------------------------ |
| UniqueArray       | Remove duplicate elements from array. Generic implementation        | deduplication, data cleaning                | /home/namnv/git/unlimit/utility/array.go   |       |            |          |                                            |
| Array2Map         | Convert array to map for O(1) lookups. Generic key-value extraction | indexing, fast lookups                      | /home/namnv/git/unlimit/utility/array.go   |       |            |          |                                            |
| BatchArray        | Divide array into batches of specified size                         | batch processing, chunking                  | /home/namnv/git/unlimit/utility/array.go   |       |            |          |                                            |
| Convert           | Generic type conversion with JSON marshaling/unmarshaling           | type safety, serialization, deserialization | /home/namnv/git/unlimit/utility/convert.go |       |            |          |                                            |
| Version Utilities | Semantic versioning comparison and parsing functions                | version comparison, dependency management   | /home/namnv/git/unlimit/utility/version.go |       |            |          | /home/namnv/git/unlimit/docs/versioning.md |

---

## Logger Module

| name   | description                                     | import path                        |
| ------ | ----------------------------------------------- | ---------------------------------- |
| Logger | Zap-based structured logging with field support | github.com/nvnamsss/unlimit/logger |

### Components

| name   | description                                                                                                  | domain                               | implementation                           | tests | benchmarks | examples | documentation |
| ------ | ------------------------------------------------------------------------------------------------------------ | ------------------------------------ | ---------------------------------------- | ----- | ---------- | -------- | ------------- |
| Logger | Logging interface with methods: Debugf, Infof, Warnf, Errorf, Fatalf, WithFields. Structured logging support | observability, debugging, monitoring | /home/namnv/git/unlimit/logger/logger.go |       |            |          |               |

---

## Messages Module

| name     | description                                    | import path                          |
| -------- | ---------------------------------------------- | ------------------------------------ |
| Messages | Message broker integrations for Kafka and NATS | github.com/nvnamsss/unlimit/messages |

### Components

| name               | description                                            | domain                                      | implementation                               | tests | benchmarks | examples | documentation |
| ------------------ | ------------------------------------------------------ | ------------------------------------------- | -------------------------------------------- | ----- | ---------- | -------- | ------------- |
| Producer           | Message producer interface for publishing messages     | message queues, pub/sub, event streaming    | /home/namnv/git/unlimit/messages/producer.go |       |            |          |               |
| Consumer           | Message consumer interface for subscribing to messages | message queues, event processing            | /home/namnv/git/unlimit/messages/consumer.go |       |            |          |               |
| KafkaSyncProducer  | Synchronous Kafka producer with delivery confirmation  | event streaming, Kafka, reliable messaging  | /home/namnv/git/unlimit/messages/kafka.go    |       |            |          |               |
| KafkaAsyncProducer | Asynchronous Kafka producer for high throughput        | event streaming, high-performance messaging | /home/namnv/git/unlimit/messages/kafka.go    |       |            |          |               |
| NATsProducer       | NATS message producer for lightweight messaging        | lightweight messaging, NATS, microservices  | /home/namnv/git/unlimit/messages/nats.go     |       |            |          |               |
| NATsConsumer       | NATS message consumer with subscription management     | message consumption, NATS                   | /home/namnv/git/unlimit/messages/nats.go     |       |            |          |               |

---

## Middlewares Module

| name        | description                                                                   | import path                             |
| ----------- | ----------------------------------------------------------------------------- | --------------------------------------- |
| Middlewares | HTTP middlewares for Gin framework providing authentication and authorization | github.com/nvnamsss/unlimit/middlewares |

### Components

| name                | description                                              | domain                            | implementation                                    | tests | benchmarks | examples | documentation |
| ------------------- | -------------------------------------------------------- | --------------------------------- | ------------------------------------------------- | ----- | ---------- | -------- | ------------- |
| AuthMiddleware      | JWT authentication middleware for securing API endpoints | API security, JWT, authentication | /home/namnv/git/unlimit/middlewares/auth.go       |       |            |          |               |
| BasicAuthMiddleware | Basic authentication middleware using username/password  | simple auth, basic authentication | /home/namnv/git/unlimit/middlewares/basic_auth.go |       |            |          |               |

---

## Optz Module

| name | description                                                         | import path                      |
| ---- | ------------------------------------------------------------------- | -------------------------------- |
| Optz | High-performance object pooling for resource reuse and optimization | github.com/nvnamsss/unlimit/optz |

### Components

| name              | description                                                                   | domain                                      | implementation                       | tests | benchmarks | examples | documentation |
| ----------------- | ----------------------------------------------------------------------------- | ------------------------------------------- | ------------------------------------ | ----- | ---------- | -------- | ------------- |
| Pool              | Generic object pool interface with bounded and random bounded implementations | object reuse, performance, resource pooling | /home/namnv/git/unlimit/optz/pool.go |       |            |          |               |
| boundedPool       | Bounded object pool with fixed capacity                                       | resource limits, connection pooling         | /home/namnv/git/unlimit/optz/pool.go |       |            |          |               |
| randomBoundedPool | Bounded pool with randomized object selection                                 | load distribution, random access            | /home/namnv/git/unlimit/optz/pool.go |       |            |          |               |

---

## Gen Module

| name | description                                              | import path                     |
| ---- | -------------------------------------------------------- | ------------------------------- |
| Gen  | ID generation and data generators for unique identifiers | github.com/nvnamsss/unlimit/gen |

### Components

| name                  | description                                                                           | domain                                   | implementation                    | tests | benchmarks | examples | documentation |
| --------------------- | ------------------------------------------------------------------------------------- | ---------------------------------------- | --------------------------------- | ----- | ---------- | -------- | ------------- |
| IDGenerator           | ID generation interface with multiple implementations (UUID, RandomToken, Sequential) | unique identifiers, ID generation        | /home/namnv/git/unlimit/gen/id.go |       |            |          |               |
| UUIDGenerator         | UUID v4 generator using google/uuid                                                   | globally unique IDs, distributed systems | /home/namnv/git/unlimit/gen/id.go |       |            |          |               |
| RandomTokenGenerator  | Random token generator for secure tokens                                              | security tokens, authentication          | /home/namnv/git/unlimit/gen/id.go |       |            |          |               |
| SequentialIDGenerator | Sequential ID generator with atomic counter                                           | monotonic IDs, ordering                  | /home/namnv/git/unlimit/gen/id.go |       |            |          |               |

---

## State Module

| name  | description                                                              | import path                       |
| ----- | ------------------------------------------------------------------------ | --------------------------------- |
| State | Finite state machine implementation with transitions and lifecycle hooks | github.com/nvnamsss/unlimit/state |

### Components

| name         | description                                                       | domain                                        | implementation                                 | tests | benchmarks | examples | documentation |
| ------------ | ----------------------------------------------------------------- | --------------------------------------------- | ---------------------------------------------- | ----- | ---------- | -------- | ------------- |
| StateMachine | FSM with state transitions, guards, and event handling            | workflow states, game logic, state management | /home/namnv/git/unlimit/state/state_machine.go |       |            |          |               |
| State        | State interface with OnEnter, OnExit, and OnEvent lifecycle hooks | state management, lifecycle                   | /home/namnv/git/unlimit/state/state.go         |       |            |          |               |

---

## Orchestrator Module

| name         | description                                                                 | import path                              |
| ------------ | --------------------------------------------------------------------------- | ---------------------------------------- |
| Orchestrator | Complex workflow orchestration with task dependencies and execution control | github.com/nvnamsss/unlimit/orchestrator |

---

## Cross-Cutting Concerns

### Error Handling
- Error strings should NOT be capitalized (per .github/copilot-instructions.md)
- Custom error types defined throughout modules
- Context-aware error propagation

### Concurrency
- Thread-safe implementations using sync.RWMutex
- Lock-free structures (LockFreeQueue in algo)
- Context-aware cancellation throughout
- Atomic operations for high-performance scenarios

### Generics
- Extensive use of Go 1.18+ generics for type safety
- Generic algorithms, caches, and data structures
- Type parameters [T any], [K comparable, V any]

### Testing
- Comprehensive unit tests (*_test.go)
- Benchmark tests (*_bench_test.go) for performance-critical code
- Example tests (*_example_test.go) for documentation

### Documentation
- Inline documentation in code
- Separate docs/ folder with detailed guides:
  - efflux.md - Workflow documentation
  - event_based_workflow.md - Event workflow architecture
  - trigger.md - Trigger system guide
  - ONETIME_POLLING.md - Polling patterns
  - error.md - Error handling
  - versioning.md - Version management
  - worker_pool_benchmark.md - Performance benchmarks

---

## Key Dependencies

- **IBM/sarama** - Kafka client
- **redis/go-redis/v9** - Redis client
- **mongodb/mongo-driver** - MongoDB driver
- **gorm.io/gorm** - ORM for SQL databases
- **gin-gonic/gin** - Web framework
- **nats-io/nats.go** - NATS messaging
- **minio/minio-go/v7** - Object storage
- **uber-go/zap** - Structured logging
- **google/uuid** - UUID generation
- **tylertreat/BoomFilters** - Bloom filters and probabilistic structures

---

## Usage Patterns

### Caching Pattern
```go
// LRU cache for bounded memory usage
cache, _ := algo.NewLRU[string, *User](1000)
cache.Add("user:123", user)
user, found := cache.Get("user:123")
```

### Worker Pool Pattern
```go
// Parallel task processing
pool := efflux.NewSimpleWorkerPool[Task](10)
pool.Submit(task)
pool.Shutdown()
```

### Event-Driven Pattern
```go
// Trigger system for reactive programming
tm := trigger.NewTriggerManager()
tm.RegisterTrigger(event, condition, action)
tm.Fire(event)
```

### Polling Pattern
```go
// One-time polling until condition met
manager := efflux.NewOneTimePollingManager[Data](time.Second, time.Minute)
results, _ := manager.Run(ctx, pollerFunc)
```

---

## Architecture Principles

1. **Generic First** - Use generics for type safety
2. **Context Aware** - All long-running operations accept context.Context
3. **Thread Safe** - Concurrent access patterns use proper synchronization
4. **Interface Driven** - Define interfaces for swappable implementations
5. **Test Coverage** - Unit tests, benchmarks, and examples
6. **Documentation** - Inline docs and separate guides

---

**Last Updated:** January 28, 2026  
**Go Version:** 1.24.0  
**Repository:** github.com/nvnamsss/unlimit
