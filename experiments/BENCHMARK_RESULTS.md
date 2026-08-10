# Mutex Map vs sync.Map Performance Comparison

This document contains the results and analysis of benchmarking a mutex-protected map versus Go's built-in `sync.Map`.

## Implementations Compared

### MutexMap
- Regular Go map protected by `sync.RWMutex`
- Uses read locks for Get operations (allowing concurrent reads)
- Uses write locks for Set/Delete operations (exclusive access)
- Simple and predictable memory layout

### SyncMapWrapper
- Wraps Go's built-in `sync.Map`
- Lock-free implementation optimized for concurrent access
- Uses atomic operations and copy-on-write semantics
- More complex internal structure

## Benchmark Results Summary

### Single-threaded Operations

| Operation | MutexMap | sync.Map | Winner | Performance Difference |
|-----------|----------|----------|---------|----------------------|
| **Set** | 145.1 ns/op, 24 B/op | 252.6 ns/op, 88 B/op | **MutexMap** | ~74% faster, 73% less memory |
| **Get** | 32.36 ns/op, 0 B/op | 34.45 ns/op, 0 B/op | **MutexMap** | ~6% faster |
| **Mixed Operations** | 48.60 ns/op, 0 B/op | 95.66 ns/op, 32 B/op | **MutexMap** | ~97% faster |

### Concurrent Operations

| Operation | MutexMap | sync.Map | Winner | Performance Difference |
|-----------|----------|----------|---------|----------------------|
| **Parallel Read/Write** | 34.72 ns/op, 0 B/op | 19.85 ns/op, 6 B/op | **sync.Map** | ~75% faster |
| **High Contention** | 98.95 ns/op, 0 B/op | 253.2 ns/op, 64 B/op | **MutexMap** | ~156% faster |

### Different Data Sizes (Get Operations)

| Size | MutexMap | sync.Map | Winner |
|------|----------|----------|---------|
| **10 items** | 21.09 ns/op | 25.71 ns/op | **MutexMap** (~22% faster) |
| **100 items** | 28.98 ns/op | 36.26 ns/op | **MutexMap** (~25% faster) |
| **1,000 items** | 34.77 ns/op | 36.92 ns/op | **MutexMap** (~6% faster) |
| **10,000 items** | 32.21 ns/op | 53.22 ns/op | **MutexMap** (~65% faster) |

## Key Findings

### ✅ MutexMap Advantages

1. **Superior Write Performance**: 74% faster Set operations
2. **Better Memory Efficiency**: 73% less memory allocation per Set operation
3. **Excellent for High Contention**: 156% faster when many goroutines compete for same keys
4. **Consistent Performance**: Performance remains stable across different data sizes
5. **Mixed Workloads**: 97% faster for typical read/write mixtures
6. **Predictable Behavior**: Simple locking semantics

### ✅ sync.Map Advantages

1. **Parallel Read-Heavy Workloads**: 75% faster when mostly reading with occasional writes
2. **Lock-Free Reads**: No blocking for concurrent read operations
3. **Standard Library**: Part of Go's standard library, well-tested
4. **Copy-on-Write**: Efficient for scenarios with many readers, few writers

### 📊 Performance Patterns

1. **Single-threaded**: MutexMap wins across all operations
2. **Write-heavy**: MutexMap significantly outperforms
3. **Read-heavy with high concurrency**: sync.Map shows advantages
4. **High contention**: MutexMap handles better
5. **Large datasets**: MutexMap scales better

## When to Use Each

### Choose MutexMap When:
- ✅ Write operations are frequent (>10% of total operations)
- ✅ High contention scenarios (many goroutines accessing same keys)
- ✅ Memory efficiency is important
- ✅ Predictable performance is required
- ✅ Mixed read/write workloads
- ✅ Small to medium-sized maps

### Choose sync.Map When:
- ✅ Read-heavy workloads (>90% reads)
- ✅ Many concurrent readers with occasional writes
- ✅ Keys are written once and read many times
- ✅ You need the guarantees of the standard library
- ✅ Lock-free read operations are critical

## Real-World Implications

### For TriggerManager Use Case

In the context of the TriggerManager:
- **Operations**: Frequent trigger registration (writes) + event processing (reads)
- **Concurrency**: Multiple goroutines registering triggers and processing events
- **Contention**: Moderate, as different events may access different triggers
- **Performance**: Event processing speed is critical

**Recommendation**: **MutexMap (current implementation) is optimal** because:
1. Trigger registration happens frequently enough to benefit from faster writes
2. Mixed read/write workload favors MutexMap
3. Better memory efficiency for long-running services
4. More predictable performance characteristics

## Code Examples

### Running the Benchmarks

```bash
# Run all benchmarks
go test -bench=. -benchmem ./experiments/

# Compare specific operations
go test -bench="BenchmarkMutexMap_|BenchmarkSyncMap_" -benchmem ./experiments/

# Test with different data sizes
go test -bench="BenchmarkComparison_DifferentSizes" -benchmem ./experiments/
```

### Usage Patterns

```go
// High-write scenario - MutexMap wins
for i := 0; i < 10000; i++ {
    map.Set(fmt.Sprintf("key-%d", i), value)
}

// Read-heavy scenario - sync.Map might win in high concurrency
go func() {
    for {
        map.Get("frequently-read-key")
    }
}()
```

## Conclusion

The benchmark results clearly show that **mutex-protected maps are superior for most general-purpose use cases**, especially those involving frequent writes or mixed workloads. The `sync.Map` shines only in very specific scenarios with extremely read-heavy concurrent access patterns.

For the TriggerManager implementation, the current mutex-based approach is well-justified and provides the best performance characteristics for the expected usage patterns.