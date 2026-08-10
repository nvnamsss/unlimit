# WorkerPool Benchmark with pprof

This program demonstrates the `efflux.WorkerPool` with integrated pprof profiling for CPU and memory benchmarking.

## Features

- Configurable number of workers
- Configurable task queue size
- Time-based or task count-based benchmarking
- Integrated pprof server for profiling
- Real-time metrics reporting
- Graceful shutdown handling
- Task retry mechanism with exponential backoff

## Usage

### Build and Run

```bash
go run main.go [flags]
```

### Command Line Flags

- `-workers`: Number of worker goroutines (default: 10)
- `-queue`: Task queue size (default: 100)
- `-tasks`: Number of tasks to process (default: 1000)
- `-pprof`: pprof server address (default: ":6060")
- `-duration`: Benchmark duration in seconds, 0 for task-based mode (default: 60)

### Examples

#### Task-based benchmark (process 5000 tasks)
```bash
go run main.go -workers=20 -queue=200 -tasks=5000 -duration=0
```

#### Time-based benchmark (run for 120 seconds)
```bash
go run main.go -workers=50 -queue=500 -duration=120
```

#### High-load stress test
```bash
go run main.go -workers=100 -queue=1000 -duration=300
```

## Profiling with pprof

The program starts a pprof HTTP server on port 6060 by default. Use the following commands to collect profiles:

### CPU Profiling

```bash
# Collect 30-second CPU profile
curl http://localhost:6060/debug/pprof/profile?seconds=30 > cpu.prof

# Analyze CPU profile
go tool pprof cpu.prof
```

In pprof interactive mode:
```
(pprof) top10        # Show top 10 CPU consumers
(pprof) list main.   # Show source code with CPU usage
(pprof) web          # Generate graph visualization (requires graphviz)
```

### Memory Profiling

```bash
# Collect heap profile
curl http://localhost:6060/debug/pprof/heap > heap.prof

# Analyze heap profile
go tool pprof heap.prof
```

In pprof interactive mode:
```
(pprof) top10        # Show top 10 memory allocators
(pprof) list main.   # Show source code with allocations
(pprof) web          # Generate graph visualization
```

### Goroutine Profiling

```bash
# Collect goroutine profile
curl http://localhost:6060/debug/pprof/goroutine > goroutine.prof

# Analyze goroutine profile
go tool pprof goroutine.prof
```

### Block Profiling

```bash
# Collect block profile
curl http://localhost:6060/debug/pprof/block > block.prof

# Analyze block profile
go tool pprof block.prof
```

### Mutex Profiling

```bash
# Collect mutex profile
curl http://localhost:6060/debug/pprof/mutex > mutex.prof

# Analyze mutex profile
go tool pprof mutex.prof
```

## Web UI for pprof

You can also use the web interface for interactive profiling:

```bash
# Start the program
go run main.go -workers=50 -duration=300

# In another terminal, open the web UI
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/profile?seconds=30
```

This will open a browser with an interactive flame graph and other visualizations.

## Monitoring Metrics

The program outputs metrics every 5 seconds:

```
metrics: QueueLen=45/200, Goroutines=25, Memory: Alloc=12 MB, TotalAlloc=150 MB, Sys=35 MB, NumGC=8
```

- **QueueLen**: Current tasks in queue / Total queue capacity
- **Goroutines**: Number of active goroutines
- **Alloc**: Currently allocated memory
- **TotalAlloc**: Cumulative allocated memory
- **Sys**: Memory obtained from OS
- **NumGC**: Number of garbage collection cycles

## Graceful Shutdown

Press `Ctrl+C` to trigger graceful shutdown:

```
received shutdown signal, stopping worker pool...
worker pool stopped gracefully
```

The program will:
1. Stop accepting new tasks
2. Wait for all in-flight tasks to complete
3. Print final memory statistics
4. Exit cleanly

## Understanding Results

### Memory Analysis

- High **Alloc** indicates current memory usage
- Rapidly increasing **TotalAlloc** suggests many allocations
- Large **Sys** compared to **Alloc** indicates fragmentation
- High **NumGC** frequency may impact performance

### CPU Analysis

- Look for hotspots in `taskHandler` or worker loops
- Check if context switching overhead is high
- Identify expensive operations in task processing

### Optimization Tips

1. **Adjust worker count**: Balance between parallelism and overhead
2. **Tune queue size**: Larger queues buffer bursts but use more memory
3. **Optimize task handler**: Reduce allocations in hot paths
4. **Batch operations**: Process multiple items together when possible
5. **Profile regularly**: Use pprof to find bottlenecks before optimizing
