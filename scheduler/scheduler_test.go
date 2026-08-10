package scheduler

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestScheduler_New(t *testing.T) {
	scheduler := New()

	if scheduler == nil {
		t.Error("New() should return a non-nil scheduler")
	}

	if scheduler.tasks == nil {
		t.Error("New() should initialize tasks map")
	}

	if scheduler.running {
		t.Error("New() should create scheduler in stopped state")
	}

	if len(scheduler.tasks) != 0 {
		t.Error("New() should create scheduler with empty tasks")
	}
}

func TestScheduler_RegisterIntervalTask(t *testing.T) {
	scheduler := New()
	var executed int32

	handler := func() error {
		atomic.AddInt32(&executed, 1)
		return nil
	}

	scheduler.RegisterIntervalTask("test-task", handler, time.Second)

	if len(scheduler.tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(scheduler.tasks))
	}

	task, exists := scheduler.tasks["test-task"]
	if !exists {
		t.Error("Task 'test-task' should exist")
	}

	if task.ID != "test-task" {
		t.Errorf("Expected task ID 'test-task', got %s", task.ID)
	}

	if task.ScheduleType != Interval {
		t.Errorf("Expected schedule type Interval, got %v", task.ScheduleType)
	}

	if task.Interval != time.Second {
		t.Errorf("Expected interval 1s, got %v", task.Interval)
	}

	if task.timer != nil {
		t.Error("Timer should be nil when scheduler is not running")
	}
}

func TestScheduler_RegisterCronTask(t *testing.T) {
	scheduler := New()
	var executed int32

	handler := func() error {
		atomic.AddInt32(&executed, 1)
		return nil
	}

	scheduler.RegisterCronTask("cron-task", handler, "0 * * * *")

	if len(scheduler.tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(scheduler.tasks))
	}

	task, exists := scheduler.tasks["cron-task"]
	if !exists {
		t.Error("Task 'cron-task' should exist")
	}

	if task.ScheduleType != Cron {
		t.Errorf("Expected schedule type Cron, got %v", task.ScheduleType)
	}

	if task.CronExpr != "0 * * * *" {
		t.Errorf("Expected cron expression '0 * * * *', got %s", task.CronExpr)
	}
}

func TestScheduler_RegisterOneOffTask(t *testing.T) {
	scheduler := New()
	var executed int32

	handler := func() error {
		atomic.AddInt32(&executed, 1)
		return nil
	}

	runAt := time.Now().Add(time.Second)
	scheduler.RegisterOneOffTask("oneoff-task", handler, runAt)

	if len(scheduler.tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(scheduler.tasks))
	}

	task, exists := scheduler.tasks["oneoff-task"]
	if !exists {
		t.Error("Task 'oneoff-task' should exist")
	}

	if task.ScheduleType != OneOff {
		t.Errorf("Expected schedule type OneOff, got %v", task.ScheduleType)
	}

	if !task.OneTime {
		t.Error("OneOff task should have OneTime set to true")
	}
}

func TestScheduler_StartStop(t *testing.T) {
	scheduler := New()

	// Test start
	scheduler.Start()

	if !scheduler.running {
		t.Error("Scheduler should be running after Start()")
	}

	if scheduler.ctx == nil {
		t.Error("Context should be initialized after Start()")
	}

	// Test double start (should not panic)
	scheduler.Start()

	// Test stop
	scheduler.Stop()

	if scheduler.running {
		t.Error("Scheduler should not be running after Stop()")
	}
}

func TestScheduler_IntervalTaskExecution(t *testing.T) {
	scheduler := New()
	var executed int32

	handler := func() error {
		atomic.AddInt32(&executed, 1)
		return nil
	}

	scheduler.RegisterIntervalTask("test-task", handler, 100*time.Millisecond)
	scheduler.Start()
	defer scheduler.Stop()

	// Wait for multiple executions
	time.Sleep(350 * time.Millisecond)

	execCount := atomic.LoadInt32(&executed)
	if execCount < 2 {
		t.Errorf("Expected at least 2 executions, got %d", execCount)
	}
}

func TestScheduler_OneOffTaskExecution(t *testing.T) {
	scheduler := New()
	var executed int32

	handler := func() error {
		atomic.AddInt32(&executed, 1)
		return nil
	}

	scheduler.Start()
	defer scheduler.Stop()

	runAt := time.Now().Add(100 * time.Millisecond)
	scheduler.RegisterOneOffTask("oneoff-task", handler, runAt)

	// Wait for execution
	time.Sleep(200 * time.Millisecond)

	execCount := atomic.LoadInt32(&executed)
	if execCount != 1 {
		t.Errorf("Expected exactly 1 execution, got %d", execCount)
	}

	// Check task was removed
	if len(scheduler.tasks) != 0 {
		t.Errorf("OneOff task should be removed after execution, but %d tasks remain", len(scheduler.tasks))
	}
}

func TestScheduler_CronTaskExecution(t *testing.T) {
	scheduler := New()
	var executed int32

	handler := func() error {
		atomic.AddInt32(&executed, 1)
		return nil
	}

	// Use a cron expression that runs every minute
	beforeRegister := time.Now()
	scheduler.RegisterCronTask("cron-task", handler, "* * * * *")

	scheduler.Start()
	defer scheduler.Stop()

	// Verify the task is scheduled
	time.Sleep(100 * time.Millisecond)

	task, exists := scheduler.tasks["cron-task"]
	if !exists {
		t.Error("Cron task should still exist")
	}

	if task.timer == nil {
		t.Error("Cron task should have an active timer")
	}

	// Verify NextRun is scheduled for the next minute
	currentMinute := beforeRegister.Minute()
	nextRunMinute := task.NextRun.Minute()

	// Calculate expected next minute (handling hour rollover)
	expectedMinute := (currentMinute + 1) % 60

	if nextRunMinute != expectedMinute {
		t.Errorf("NextRun should be scheduled for the next minute. Current minute: %d, Expected next minute: %d, Got: %d (NextRun: %v)",
			currentMinute, expectedMinute, nextRunMinute, task.NextRun)
	}

	// Verify NextRun is in the future
	if task.NextRun.Before(beforeRegister) {
		t.Errorf("NextRun should be in the future, got %v", task.NextRun)
	}
}

func TestScheduler_EmptyOperations(t *testing.T) {
	scheduler := New()

	// Start/stop empty scheduler should not panic
	scheduler.Start()
	scheduler.Stop()

	// Starting already stopped scheduler should not panic
	scheduler.Stop()
}

func TestScheduler_InvalidCronExpression(t *testing.T) {
	scheduler := New()
	var executed int32

	handler := func() error {
		atomic.AddInt32(&executed, 1)
		return nil
	}

	// Register task with invalid cron expression
	scheduler.RegisterCronTask("invalid-cron", handler, "invalid-expression")

	task, exists := scheduler.tasks["invalid-cron"]
	if !exists {
		t.Error("Task should be registered even with invalid cron expression")
	}

	if task.CronExpr != "invalid-expression" {
		t.Error("Invalid cron expression should be stored as-is")
	}
}

func TestScheduler_ConcurrentRegistration(t *testing.T) {
	scheduler := New()
	var wg sync.WaitGroup

	taskCount := 10
	wg.Add(taskCount)

	handler := func() error { return nil }

	// Register tasks concurrently
	for i := 0; i < taskCount; i++ {
		go func(id int) {
			defer wg.Done()
			scheduler.RegisterIntervalTask(
				fmt.Sprintf("task-%d", id),
				handler,
				time.Second,
			)
		}(i)
	}

	wg.Wait()

	if len(scheduler.tasks) != taskCount {
		t.Errorf("Expected %d tasks, got %d", taskCount, len(scheduler.tasks))
	}
}

func TestScheduler_TaskReplacementSameID(t *testing.T) {
	scheduler := New()
	var executed1, executed2 int32

	handler1 := func() error {
		atomic.AddInt32(&executed1, 1)
		return nil
	}

	handler2 := func() error {
		atomic.AddInt32(&executed2, 1)
		return nil
	}

	// Register first task
	scheduler.RegisterIntervalTask("same-id", handler1, time.Second)

	// Replace with second task
	scheduler.RegisterIntervalTask("same-id", handler2, time.Second)

	if len(scheduler.tasks) != 1 {
		t.Errorf("Expected 1 task after replacement, got %d", len(scheduler.tasks))
	}

	// Verify the task was replaced by checking the handler reference
	task := scheduler.tasks["same-id"]
	if task.Handler == nil {
		t.Error("Task handler should not be nil")
	}
}

func TestScheduler_StopCancelsTimers(t *testing.T) {
	scheduler := New()
	var executed int32

	handler := func() error {
		atomic.AddInt32(&executed, 1)
		return nil
	}

	scheduler.RegisterIntervalTask("test-task", handler, 50*time.Millisecond)
	scheduler.Start()

	// Let it run briefly
	time.Sleep(25 * time.Millisecond)

	scheduler.Stop()

	// Wait longer than the interval
	time.Sleep(100 * time.Millisecond)

	// Task should not execute after stop
	initialCount := atomic.LoadInt32(&executed)
	time.Sleep(100 * time.Millisecond)
	finalCount := atomic.LoadInt32(&executed)

	if finalCount > initialCount {
		t.Error("Task should not execute after scheduler is stopped")
	}
}

func TestScheduler_RemoveTask(t *testing.T) {
	scheduler := New()
	var executed int32

	handler := func() error {
		atomic.AddInt32(&executed, 1)
		return nil
	}

	// Register a task
	scheduler.RegisterIntervalTask("test-task", handler, time.Second)

	if len(scheduler.tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(scheduler.tasks))
	}

	// Remove the task
	removed := scheduler.RemoveTask("test-task")
	if !removed {
		t.Error("RemoveTask should return true for existing task")
	}

	if len(scheduler.tasks) != 0 {
		t.Errorf("Expected 0 tasks after removal, got %d", len(scheduler.tasks))
	}
}

func TestScheduler_RemoveTaskNonExistent(t *testing.T) {
	scheduler := New()

	// Try to remove non-existent task
	removed := scheduler.RemoveTask("non-existent")
	if removed {
		t.Error("RemoveTask should return false for non-existent task")
	}
}

func TestScheduler_RemoveTaskWithActiveTimer(t *testing.T) {
	scheduler := New()
	var executed int32

	handler := func() error {
		atomic.AddInt32(&executed, 1)
		return nil
	}

	scheduler.RegisterIntervalTask("test-task", handler, 100*time.Millisecond)
	scheduler.Start()
	defer scheduler.Stop()

	// Verify task is scheduled
	time.Sleep(50 * time.Millisecond)

	// Remove the task while it has an active timer
	removed := scheduler.RemoveTask("test-task")
	if !removed {
		t.Error("RemoveTask should return true for existing task with active timer")
	}

	// Wait past the original interval
	time.Sleep(200 * time.Millisecond)

	// Task should not have executed since it was removed
	execCount := atomic.LoadInt32(&executed)
	if execCount > 0 {
		t.Errorf("Task should not execute after removal, but executed %d times", execCount)
	}
}

func TestScheduler_RemoveTaskConcurrent(t *testing.T) {
	scheduler := New()
	var wg sync.WaitGroup

	handler := func() error { return nil }

	// Register multiple tasks
	taskCount := 10
	for i := 0; i < taskCount; i++ {
		scheduler.RegisterIntervalTask(fmt.Sprintf("task-%d", i), handler, time.Second)
	}

	wg.Add(taskCount)

	// Remove tasks concurrently
	for i := 0; i < taskCount; i++ {
		go func(id int) {
			defer wg.Done()
			removed := scheduler.RemoveTask(fmt.Sprintf("task-%d", id))
			if !removed {
				t.Errorf("Task task-%d should have been removed", id)
			}
		}(i)
	}

	wg.Wait()

	if len(scheduler.tasks) != 0 {
		t.Errorf("Expected 0 tasks after concurrent removal, got %d", len(scheduler.tasks))
	}
}
