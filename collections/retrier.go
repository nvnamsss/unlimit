package collections

// Retrier implements the "retriable" resiliency pattern for Go.

import (
	"context"
	"time"
)

// Retrier implements the "retriable" resiliency pattern, abstracting out the process of retrying a failed action
// a certain number of times with an optional back-off between each retry.
// It is a generic retry mechanism that can be configured with custom backoff strategies and success criteria.
// Example usage:
//
//	backoffFunc := func(attempt int) time.Duration {
//		if attempt >= 3 { return 0 } // Stop after 3 retries
//		return time.Duration(1<<uint(attempt)) * time.Second // Exponential backoff
//	}
//	isSuccessful := func(err error) bool { return err == nil }
//	retrier := New(backoffFunc, isSuccessful)
//
//	err := retrier.Run(ctx, func(ctx context.Context) error {
//		return someOperation()
//	})
//
// This struct is useful for implementing resilient operations that may fail temporarily.
type Retrier struct {
	CurrentRetryCount  int
	TotalRetryCount    int
	TotalSleepDuration time.Duration

	backoffFunc       func(int) time.Duration
	surfaceWorkErrors bool
	isSuccessful      func(error) bool
}

// NewRetrier constructs a Retrier with the given backoff function and isSuccessful function.
// It creates a new retry mechanism that will attempt operations according to the backoff function.
// Example usage:
//
//	backoffFunc := func(attempt int) time.Duration { return time.Duration(attempt) * time.Second }
//	isSuccessful := func(err error) bool { return err == nil || isTemporaryError(err) }
//	retrier := NewRetrier(backoffFunc, isSuccessful)
//
// The backoff function receives the attempt number (starting from 0) and returns the duration to wait.
// Return 0 from the backoff function to stop retrying.
// The isSuccessful function is used to determine if an error should be retried.
// If isSuccessful is nil, only nil errors are considered successful.
func NewRetrier(backoffFunc func(int) time.Duration, isSuccessful func(error) bool) *Retrier {
	if isSuccessful == nil {
		isSuccessful = func(err error) bool { return err == nil }
	}

	return &Retrier{
		backoffFunc:  backoffFunc,
		isSuccessful: isSuccessful,
	}
}

// WithSurfaceWorkErrors configures the retrier to always return the last error from the work function.
// It modifies error handling to return work errors even if a context timeout/deadline is hit.
// Example usage:
//
//	retrier := New(backoff, isSuccessful).WithSurfaceWorkErrors()
//	err := retrier.Run(ctx, work) // Returns work error instead of context.DeadlineExceeded
//
// This is useful when you want to distinguish between operation failures and context cancellations.
func (r *Retrier) WithSurfaceWorkErrors() *Retrier {
	r.surfaceWorkErrors = true
	return r
}

// Run executes the given work function with retry logic based on the configured parameters.
// It repeatedly calls the work function until it succeeds or the retry limit is reached.
// Example usage:
//
//	err := retrier.Run(ctx, func(ctx context.Context) error {
//		response, err := httpClient.Get("https://api.example.com/data")
//		if err != nil {
//			return err
//		}
//		defer response.Body.Close()
//		return processResponse(response)
//	})
//
// The work function is called with the provided context and should return an error.
// If the error is considered successful by the isSuccessful function, the operation completes.
// Otherwise, Run sleeps according to the backoff policy before retrying.
func (r *Retrier) Run(ctx context.Context, work func(ctx context.Context) error) error {
	return r.RunFn(ctx, func(c context.Context, r int) error {
		return work(c)
	})
}

// RunFn executes the given work function with retry logic, providing retry count information.
// It is similar to Run but passes the current retry count to the work function.
// Example usage:
//
//	err := retrier.RunFn(ctx, func(ctx context.Context, retries int) error {
//		if retries > 0 {
//			log.Printf("Retrying operation, attempt %d", retries+1)
//		}
//		return performOperation()
//	})
//
// The work function receives the context and the number of retry attempts (starting from 0).
// This is useful when the work function needs to adjust its behavior based on retry count.
// Retrying stops when backoffFunc returns 0 or when the operation succeeds.
func (r *Retrier) RunFn(ctx context.Context, work func(ctx context.Context, retries int) error) error {
	r.CurrentRetryCount = 0
	r.TotalRetryCount = 0
	r.TotalSleepDuration = 0

	for {
		ret := work(ctx, r.CurrentRetryCount)
		r.TotalRetryCount++

		if r.isSuccessful(ret) {
			return ret
		}

		// Get backoff duration for this retry attempt
		sleepDuration := r.backoffFunc(r.CurrentRetryCount)
		if sleepDuration == 0 {
			// backoffFunc returns 0 to signal no more retries
			return ret
		}

		timer := time.NewTimer(sleepDuration)
		if err := r.sleep(ctx, timer); err != nil {
			if r.surfaceWorkErrors {
				return ret
			}
			return err
		}

		r.TotalSleepDuration += sleepDuration
		r.CurrentRetryCount++
	}
}

func (r *Retrier) sleep(ctx context.Context, timer *time.Timer) error {
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		timer.Stop()
		return ctx.Err()
	}
}
