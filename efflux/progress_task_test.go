package efflux

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestBaseProgressTask_New(t *testing.T) {
	task := NewBaseProgressTask("test-id", []string{"step1", "step2"})
	if task.GetID() != "test-id" {
		t.Errorf("expected id 'test-id', got %v", task.GetID())
	}
	if task.GetTotalSteps() != 2 {
		t.Errorf("expected 2 steps, got %d", task.GetTotalSteps())
	}
	if task.GetCurrentStep() != 0 {
		t.Errorf("expected current step 0, got %d", task.GetCurrentStep())
	}
}

func TestBaseProgressTask_ExecuteAndProgress(t *testing.T) {
	task := NewBaseProgressTask("exec-id", []string{"a", "b", "c"})
	executed := make([]string, 0)
	task.SetStepExecutor(func(ctx context.Context, stepIndex int, stepName string) error {
		executed = append(executed, stepName)
		return nil
	})

	for !task.IsCompleted() {
		err := task.Execute(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		task.NextStep()
	}

	if len(executed) != 3 {
		t.Errorf("expected 3 executed steps, got %d", len(executed))
	}
	if !task.IsCompleted() {
		t.Errorf("task should be completed")
	}
	if task.GetProgress() != 1.0 {
		t.Errorf("expected progress 1.0, got %f", task.GetProgress())
	}
}

func TestBaseProgressTask_EmptySteps(t *testing.T) {
	task := NewBaseProgressTask("empty-id", []string{})
	if !task.IsCompleted() {
		t.Errorf("empty task should be completed")
	}
	if task.GetProgress() != 1.0 {
		t.Errorf("expected progress 1.0 for empty, got %f", task.GetProgress())
	}
}

func TestBaseProgressTask_InvalidStepIndex(t *testing.T) {
	task := NewBaseProgressTask("invalid-step", []string{"x"})
	err := task.SetCurrentStep(-1)
	if err == nil {
		t.Error("SetCurrentStep(-1) should return error")
	}
	err = task.SetCurrentStep(2)
	if err == nil {
		t.Error("SetCurrentStep(2) should return error")
	}
}

func TestBaseProgressTask_ErrorHandling(t *testing.T) {
	task := NewBaseProgressTask("err-id", []string{"fail"})
	task.SetStepExecutor(func(ctx context.Context, stepIndex int, stepName string) error {
		return errors.New("step failed")
	})
	err := task.Execute(context.Background())
	if err == nil {
		t.Error("Execute should return error")
	}
	if len(task.GetErrorHistory()) != 1 {
		t.Errorf("expected 1 error in history, got %d", len(task.GetErrorHistory()))
	}
}

func TestBaseProgressTask_Metadata(t *testing.T) {
	task := NewBaseProgressTask("meta-id", []string{"a"})
	meta := map[string]interface{}{"foo": 42, "bar": "baz"}
	task.SetMetadata(meta)
	got := task.GetMetadata()
	if got["foo"] != 42 || got["bar"] != "baz" {
		t.Errorf("metadata not set or retrieved correctly: %v", got)
	}
}

func TestBaseProgressTask_SaveLoadProgress(t *testing.T) {
	task := NewBaseProgressTask("persist-id", []string{"a", "b"})
	var savedStep int
	task.SetProgressSaver(func() error {
		savedStep = task.GetCurrentStep()
		return nil
	})
	task.SetProgressLoader(func() error {
		return task.SetCurrentStep(savedStep)
	})
	task.NextStep()
	err := task.SaveProgress()
	if err != nil {
		t.Errorf("SaveProgress should not error: %v", err)
	}
	task.SetCurrentStep(0)
	err = task.LoadProgress()
	if err != nil {
		t.Errorf("LoadProgress should not error: %v", err)
	}
	if task.GetCurrentStep() != 1 {
		t.Errorf("expected current step 1 after load, got %d", task.GetCurrentStep())
	}
}

func TestBaseProgressTask_Reset(t *testing.T) {
	task := NewBaseProgressTask("reset-id", []string{"a", "b"})
	task.NextStep()
	task.SetMetadata(map[string]interface{}{"x": 1})
	task.Reset()
	if task.GetCurrentStep() != 0 {
		t.Errorf("expected step 0 after reset, got %d", task.GetCurrentStep())
	}
	if len(task.GetErrorHistory()) != 0 {
		t.Errorf("expected no errors after reset")
	}
}

func TestBaseProgressTask_CanRetry(t *testing.T) {
	task := NewBaseProgressTask("retry-id", []string{"a"})
	if !task.CanRetry() {
		t.Errorf("CanRetry should be true before completion")
	}
	task.NextStep()
	if task.CanRetry() {
		t.Errorf("CanRetry should be false after completion")
	}
}

func TestBaseProgressTask_ShouldSaveProgress(t *testing.T) {
	task := NewBaseProgressTask("savefreq-id", []string{"a"})
	called := false
	task.SetSaveFrequencyChecker(func() bool {
		called = true
		return false
	})
	if task.ShouldSaveProgress() {
		t.Errorf("ShouldSaveProgress should be false")
	}
	if !called {
		t.Errorf("SaveFrequencyChecker should be called")
	}
}

func TestBaseProgressTask_ConcurrentOperations(t *testing.T) {
	task := NewBaseProgressTask("concurrent-id", []string{"a", "b", "c"})
	task.SetStepExecutor(func(ctx context.Context, stepIndex int, stepName string) error {
		time.Sleep(10 * time.Millisecond)
		return nil
	})
	var wg sync.WaitGroup
	wg.Add(3)
	for i := 0; i < 3; i++ {
		go func() {
			defer wg.Done()
			task.Execute(context.Background())
		}()
	}
	wg.Wait()
	// Not thread-safe, but should not panic or deadlock
}
