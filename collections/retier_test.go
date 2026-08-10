package collections

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testError = errors.New("test error")
	tempError = errors.New("temporary error")
)

func TestRetrier_New(t *testing.T) {
	t.Run("should create retrier with default isSuccessful", func(t *testing.T) {
		backoffFunc := func(attempt int) time.Duration {
			if attempt >= 2 {
				return 0
			}
			return time.Millisecond * time.Duration(1<<uint(attempt))
		}
		retrier := NewRetrier(backoffFunc, nil)

		require.NotNil(t, retrier)
		assert.NotNil(t, retrier.backoffFunc)
		assert.NotNil(t, retrier.isSuccessful)
		assert.False(t, retrier.surfaceWorkErrors)
	})

	t.Run("should create retrier with custom isSuccessful", func(t *testing.T) {
		backoffFunc := func(attempt int) time.Duration {
			if attempt >= 1 {
				return 0
			}
			return time.Millisecond
		}
		isSuccessful := func(err error) bool { return err == nil || err == tempError }
		retrier := NewRetrier(backoffFunc, isSuccessful)

		require.NotNil(t, retrier)
		assert.NotNil(t, retrier.backoffFunc)
		assert.NotNil(t, retrier.isSuccessful)
	})
}

func TestRetrier_WithSurfaceWorkErrors(t *testing.T) {
	backoffFunc := func(attempt int) time.Duration {
		if attempt >= 1 {
			return 0
		}
		return time.Millisecond
	}
	retrier := NewRetrier(backoffFunc, nil).WithSurfaceWorkErrors()
	assert.True(t, retrier.surfaceWorkErrors)
}

func TestRetrier_Run(t *testing.T) {
	t.Run("should succeed on first try", func(t *testing.T) {
		backoffFunc := func(attempt int) time.Duration {
			if attempt >= 1 {
				return 0
			}
			return time.Millisecond
		}
		retrier := NewRetrier(backoffFunc, nil)
		ctx := context.Background()

		callCount := 0
		work := func(ctx context.Context) error {
			callCount++
			return nil
		}

		err := retrier.Run(ctx, work)
		assert.NoError(t, err)
		assert.Equal(t, 1, callCount)
		assert.Equal(t, 1, retrier.TotalRetryCount)
		assert.Equal(t, 0, retrier.CurrentRetryCount)
	})

	t.Run("should retry and succeed", func(t *testing.T) {
		backoffFunc := func(attempt int) time.Duration {
			if attempt >= 2 {
				return 0
			}
			return time.Millisecond
		}
		retrier := NewRetrier(backoffFunc, nil)
		ctx := context.Background()

		callCount := 0
		work := func(ctx context.Context) error {
			callCount++
			if callCount < 3 {
				return testError
			}
			return nil
		}

		err := retrier.Run(ctx, work)
		assert.NoError(t, err)
		assert.Equal(t, 3, callCount)
		assert.Equal(t, 3, retrier.TotalRetryCount)
		assert.Equal(t, 2, retrier.CurrentRetryCount)
	})

	t.Run("should fail after max retries", func(t *testing.T) {
		backoffFunc := func(attempt int) time.Duration {
			if attempt >= 1 {
				return 0
			}
			return time.Millisecond
		}
		retrier := NewRetrier(backoffFunc, nil)
		ctx := context.Background()

		callCount := 0
		work := func(ctx context.Context) error {
			callCount++
			return testError
		}

		err := retrier.Run(ctx, work)
		assert.Error(t, err)
		assert.Equal(t, testError, err)
		assert.Equal(t, 2, callCount) // initial + 1 retry
		assert.Equal(t, 2, retrier.TotalRetryCount)
	})

	t.Run("should respect custom isSuccessful function", func(t *testing.T) {
		// Custom function that considers tempError as successful
		isSuccessful := func(err error) bool { return err == nil || err == tempError }
		backoffFunc := func(attempt int) time.Duration {
			if attempt >= 1 {
				return 0
			}
			return time.Millisecond
		}
		retrier := NewRetrier(backoffFunc, isSuccessful)
		ctx := context.Background()

		callCount := 0
		work := func(ctx context.Context) error {
			callCount++
			return tempError
		}

		err := retrier.Run(ctx, work)
		assert.Equal(t, tempError, err)
		assert.Equal(t, 1, callCount) // Should not retry
		assert.Equal(t, 1, retrier.TotalRetryCount)
	})
}

func TestRetrier_RunFn(t *testing.T) {
	t.Run("should pass retry count to work function", func(t *testing.T) {
		backoffFunc := func(attempt int) time.Duration {
			if attempt >= 2 {
				return 0
			}
			return time.Millisecond
		}
		retrier := NewRetrier(backoffFunc, nil)
		ctx := context.Background()

		var retryCounts []int
		work := func(ctx context.Context, retries int) error {
			retryCounts = append(retryCounts, retries)
			if retries < 2 {
				return testError
			}
			return nil
		}

		err := retrier.RunFn(ctx, work)
		assert.NoError(t, err)
		assert.Equal(t, []int{0, 1, 2}, retryCounts)
		assert.Equal(t, 3, retrier.TotalRetryCount)
		assert.Equal(t, 2, retrier.CurrentRetryCount)
	})

	t.Run("should handle context cancellation", func(t *testing.T) {
		backoffFunc := func(attempt int) time.Duration {
			if attempt >= 5 {
				return 0
			}
			return time.Second
		}
		retrier := NewRetrier(backoffFunc, nil)
		ctx, cancel := context.WithCancel(context.Background())

		work := func(ctx context.Context, retries int) error {
			if retries == 0 {
				cancel() // Cancel after first attempt
			}
			return testError
		}

		err := retrier.RunFn(ctx, work)
		assert.Error(t, err)
		assert.Equal(t, context.Canceled, err)
	})

	t.Run("should surface work errors when configured", func(t *testing.T) {
		backoffFunc := func(attempt int) time.Duration {
			if attempt >= 5 {
				return 0
			}
			return time.Second
		}
		retrier := NewRetrier(backoffFunc, nil).WithSurfaceWorkErrors()
		ctx, cancel := context.WithCancel(context.Background())

		work := func(ctx context.Context, retries int) error {
			if retries == 0 {
				cancel()
			}
			return testError
		}

		err := retrier.RunFn(ctx, work)
		assert.Error(t, err)
		assert.Equal(t, testError, err)
	})
}

func TestRetrier_TrackingStats(t *testing.T) {
	t.Run("should track retry count and sleep duration", func(t *testing.T) {
		backoffFunc := func(attempt int) time.Duration {
			if attempt >= 3 {
				return 0
			}
			return time.Millisecond * time.Duration(attempt+1)
		}
		retrier := NewRetrier(backoffFunc, nil)
		ctx := context.Background()

		callCount := 0
		work := func(ctx context.Context, retries int) error {
			callCount++
			if callCount <= 3 {
				return testError
			}
			return nil
		}

		err := retrier.RunFn(ctx, work)
		assert.NoError(t, err)
		assert.Equal(t, 4, callCount)
		assert.Equal(t, 4, retrier.TotalRetryCount)
		assert.Equal(t, 3, retrier.CurrentRetryCount)
		// TotalSleepDuration should be 1ms + 2ms + 3ms = 6ms
		assert.Equal(t, 6*time.Millisecond, retrier.TotalSleepDuration)
	})
}

func TestRetrier_EdgeCases(t *testing.T) {
	t.Run("should handle immediate stop from backoffFunc", func(t *testing.T) {
		backoffFunc := func(attempt int) time.Duration {
			return 0 // Stop immediately after first failure
		}
		retrier := NewRetrier(backoffFunc, nil)
		ctx := context.Background()

		callCount := 0
		work := func(ctx context.Context) error {
			callCount++
			return testError
		}

		err := retrier.Run(ctx, work)
		assert.Error(t, err)
		assert.Equal(t, 1, callCount) // Only initial call, no retries
		assert.Equal(t, 1, retrier.TotalRetryCount)
	})

	t.Run("should handle successful operation without retries", func(t *testing.T) {
		backoffFunc := func(attempt int) time.Duration {
			if attempt >= 1 {
				return 0
			}
			return time.Millisecond
		}
		retrier := NewRetrier(backoffFunc, nil)
		ctx := context.Background()

		work := func(ctx context.Context) error {
			return nil
		}

		err := retrier.Run(ctx, work)
		assert.NoError(t, err)
		assert.Equal(t, 0, retrier.CurrentRetryCount)
		assert.Equal(t, 1, retrier.TotalRetryCount)
	})
}
