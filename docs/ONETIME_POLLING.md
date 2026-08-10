# OneTimePollingManager

`OneTimePollingManager` is a transient polling utility designed for scenarios where you need to poll until a specific condition is met, without the overhead of managing persistent background pollers.

## Key Features

- **Immediate First Poll**: Executes the first poll immediately, then continues at configured intervals
- **Self-Stopping**: Polls until the poller function returns `shouldStop=true`
- **Timeout Support**: Optional timeout to prevent infinite polling
- **Retry Logic**: Configurable retry attempts with exponential backoff for failed polls
- **Context Cancellation**: Respects context cancellation for graceful shutdown
- **Comprehensive Results**: Returns all collected data, errors, and metadata

## Usage

### Basic Usage

```go
// Create manager with 100ms interval and 5s timeout
manager := NewOneTimePollingManager[string](100*time.Millisecond, 5*time.Second)

poller := PollerFunc[string](func(ctx context.Context, metrics PollingMetrics) ([]string, bool, error) {
    data := fetchData()
    shouldStop := checkCondition()
    return data, shouldStop, nil
})

result, err := manager.Run(ctx, poller)
if err != nil {
    // Handle error
}

fmt.Printf("Total polls: %d\n", result.TotalPolls)
fmt.Printf("Results: %v\n", result.Results)
```

### Simplified API

For cases where you only need the results:

```go
manager := NewOneTimePollingManager[int](50*time.Millisecond, 2*time.Second)
results, err := manager.RunSimple(ctx, pollerFunc)
// Returns flattened slice of all results
```

### Configuration

```go
manager := NewOneTimePollingManager[T](interval, timeout)
manager.MaxRetries = 3                    // Retry failed polls up to 3 times
manager.BackoffFunc = customBackoffFunc  // Custom backoff strategy
```

## Result Structure

`OneTimePollingResult` contains:
- `Results [][]T` - All batches collected
- `TotalPolls int` - Number of successful polls
- `Errors []error` - Errors encountered (after retries)
- `Duration time.Duration` - Total polling time
- `StoppedByPoller bool` - True if stopped via shouldStop
- `TimedOut bool` - True if stopped via timeout

## Comparison with PollingManager

| Feature | PollingManager | OneTimePollingManager |
|---------|----------------|----------------------|
| Lifecycle | Persistent background pollers | Transient, runs until completion |
| Management | Multiple concurrent pollers | Single synchronous execution |
| API | `RunPoller()`, `StopPoller()` | `Run()`, `RunSimple()` |
| Use Case | Continuous background polling | One-time data collection |

## Common Use Cases

1. **Wait for job completion**: Poll a job status API until job is done
2. **Collect time-bounded data**: Gather metrics over a specific time window
3. **Condition-based polling**: Poll until a specific condition is met
4. **Transient data fetching**: One-off polling without persistent management

## Error Handling

Failed polls are automatically retried based on `MaxRetries` configuration. If all retries fail, the error is recorded in `result.Errors` and polling continues at the next interval.

```go
manager.MaxRetries = 3
manager.BackoffFunc = func(attempt int) time.Duration {
    return time.Second * time.Duration(1<<attempt)
}
```
