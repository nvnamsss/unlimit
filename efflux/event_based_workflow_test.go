package efflux

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

// EventTestTask is a sample implementation of EventTask
type EventTestTask struct {
	id         string
	state      string
	retries    int
	maxRetries int
	executeAt  time.Time
	err        error
	data       string
	processLog []string
	mu         sync.Mutex
}

func (t *EventTestTask) GetID() string             { return t.id }
func (t *EventTestTask) GetRetries() int           { return t.retries }
func (t *EventTestTask) SetRetries(retries int)    { t.retries = retries }
func (t *EventTestTask) GetMaxRetries() int        { return t.maxRetries }
func (t *EventTestTask) GetExecuteAt() time.Time   { return t.executeAt }
func (t *EventTestTask) SetExecuteAt(tm time.Time) { t.executeAt = tm }
func (t *EventTestTask) SetError(err error)        { t.err = err }
func (t *EventTestTask) GetError() error           { return t.err }
func (t *EventTestTask) GetState() string          { return t.state }
func (t *EventTestTask) SetState(state string)     { t.state = state }
func (t *EventTestTask) AddLog(entry string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.processLog = append(t.processLog, entry)
}
func (t *EventTestTask) GetLog() []string {
	t.mu.Lock()
	defer t.mu.Unlock()
	result := make([]string, len(t.processLog))
	copy(result, t.processLog)
	return result
}

// MockTaskPuller simulates pulling tasks from an external source
type MockTaskPuller struct {
	tasks     []*EventTestTask
	pullCount int
	mu        sync.Mutex
}

func (p *MockTaskPuller) Pull(ctx context.Context) ([]*EventTestTask, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.pullCount >= len(p.tasks) {
		return nil, nil
	}

	// Return one task per pull
	task := p.tasks[p.pullCount]
	p.pullCount++
	return []*EventTestTask{task}, nil
}

// EventTestWorker is a sample implementation of EventWorker
type EventTestWorker struct {
	id string
}

func (w *EventTestWorker) Process(ctx context.Context, task *EventTestTask) error {
	task.AddLog(fmt.Sprintf("processed-by-%s", w.id))
	task.SetState("processed")
	// Simulate work
	time.Sleep(10 * time.Millisecond)
	return nil
}

func (w *EventTestWorker) GetID() string {
	return w.id
}

// EventTestWorkerFactory creates event test workers
type EventTestWorkerFactory struct{}

func (f *EventTestWorkerFactory) Create(workerID string) EventWorker[*EventTestTask] {
	return &EventTestWorker{id: workerID}
}

// MockStateUpdater simulates state updates
type MockStateUpdater struct {
	updatedTasks []*EventTestTask
	mu           sync.Mutex
}

func (u *MockStateUpdater) Update(ctx context.Context, task *EventTestTask) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	task.AddLog("state-updated")
	u.updatedTasks = append(u.updatedTasks, task)
	return nil
}

func (u *MockStateUpdater) GetUpdatedCount() int {
	u.mu.Lock()
	defer u.mu.Unlock()
	return len(u.updatedTasks)
}

func TestEventBasedWorkflow_BasicFlow(t *testing.T) {
	// Create tasks to be pulled
	tasks := []*EventTestTask{
		{id: "task-1", state: "pending", data: "data-1", processLog: make([]string, 0)},
		{id: "task-2", state: "pending", data: "data-2", processLog: make([]string, 0)},
		{id: "task-3", state: "pending", data: "data-3", processLog: make([]string, 0)},
	}

	// Create puller
	puller := &MockTaskPuller{tasks: tasks}

	// Create queue with puller
	queue := NewBasicEventQueue[*EventTestTask](10, puller, 50*time.Millisecond)

	// Create state updater
	stateUpdater := &MockStateUpdater{
		updatedTasks: make([]*EventTestTask, 0),
	}

	// Create worker manager (no return queue for this test)
	workerManager := NewBasicEventWorkerManager[*EventTestTask](
		&EventTestWorkerFactory{},
		2,
		10,
		stateUpdater,
		nil, // no return queue
	)

	// Create workflow
	workflow := NewEventBasedWorkflow[*EventTestTask](queue, workerManager, stateUpdater)

	// Start workflow
	ctx := context.Background()
	if err := workflow.Start(ctx); err != nil {
		t.Fatalf("failed to start workflow: %v", err)
	}

	// Wait for tasks to be pulled and processed
	time.Sleep(300 * time.Millisecond)

	// Stop workflow
	stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := workflow.Stop(stopCtx); err != nil {
		t.Fatalf("failed to stop workflow: %v", err)
	}

	// Verify all tasks were processed
	for i, task := range tasks {
		if task.GetState() != "processed" {
			t.Errorf("task %d state: expected 'processed', got '%s'", i, task.GetState())
		}

		logs := task.GetLog()
		if len(logs) < 2 {
			t.Errorf("task %d: expected at least 2 log entries, got %d", i, len(logs))
		}
	}

	// Verify state updater was called
	updatedCount := stateUpdater.GetUpdatedCount()
	if updatedCount != len(tasks) {
		t.Errorf("expected %d tasks to be updated, got %d", len(tasks), updatedCount)
	}
}

func TestEventBasedWorkflow_ManualSend(t *testing.T) {
	// Create queue without puller (manual send only)
	queue := NewBasicEventQueue[*EventTestTask](10, nil, 0)

	// Create state updater
	stateUpdater := &MockStateUpdater{
		updatedTasks: make([]*EventTestTask, 0),
	}

	// Create worker manager
	workerManager := NewBasicEventWorkerManager[*EventTestTask](
		&EventTestWorkerFactory{},
		2,
		10,
		stateUpdater,
		nil,
	)

	// Create workflow
	workflow := NewEventBasedWorkflow[*EventTestTask](queue, workerManager, stateUpdater)

	// Start workflow
	ctx := context.Background()
	if err := workflow.Start(ctx); err != nil {
		t.Fatalf("failed to start workflow: %v", err)
	}

	// Manually send tasks
	tasks := []*EventTestTask{
		{id: "manual-1", state: "pending", processLog: make([]string, 0)},
		{id: "manual-2", state: "pending", processLog: make([]string, 0)},
	}

	for _, task := range tasks {
		if err := workflow.Submit(ctx, task); err != nil {
			t.Fatalf("failed to submit task: %v", err)
		}
	}

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	// Stop workflow
	stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := workflow.Stop(stopCtx); err != nil {
		t.Fatalf("failed to stop workflow: %v", err)
	}

	// Verify processing
	for i, task := range tasks {
		if task.GetState() != "processed" {
			t.Errorf("task %d state: expected 'processed', got '%s'", i, task.GetState())
		}
	}
}

func TestEventBasedWorkflow_WithReturnQueue(t *testing.T) {
	// Create main queue
	mainQueue := NewBasicEventQueue[*EventTestTask](10, nil, 0)

	// Create return queue (tasks go back here after processing)
	returnQueue := NewBasicEventQueue[*EventTestTask](10, nil, 0)

	// Track returned tasks
	var returnedTasks []*EventTestTask
	var returnMu sync.Mutex

	// Create state updater
	stateUpdater := &MockStateUpdater{
		updatedTasks: make([]*EventTestTask, 0),
	}

	// Create worker manager with return queue
	workerManager := NewBasicEventWorkerManager[*EventTestTask](
		&EventTestWorkerFactory{},
		1,
		5,
		stateUpdater,
		returnQueue, // tasks return here after processing
	)

	// Create workflow
	workflow := NewEventBasedWorkflow[*EventTestTask](mainQueue, workerManager, stateUpdater)

	// Start workflow
	ctx := context.Background()
	if err := workflow.Start(ctx); err != nil {
		t.Fatalf("failed to start workflow: %v", err)
	}

	// Start return queue
	if err := returnQueue.Start(ctx); err != nil {
		t.Fatalf("failed to start return queue: %v", err)
	}

	// Monitor return queue
	go func() {
		returnChan := returnQueue.Subscribe()
		for {
			select {
			case task := <-returnChan:
				returnMu.Lock()
				returnedTasks = append(returnedTasks, task)
				returnMu.Unlock()
			case <-ctx.Done():
				return
			}
		}
	}()

	// Submit tasks
	tasks := []*EventTestTask{
		{id: "return-1", state: "pending", processLog: make([]string, 0)},
		{id: "return-2", state: "pending", processLog: make([]string, 0)},
	}

	for _, task := range tasks {
		if err := workflow.Submit(ctx, task); err != nil {
			t.Fatalf("failed to submit task: %v", err)
		}
	}

	// Wait for processing and return
	time.Sleep(200 * time.Millisecond)

	// Stop workflow
	stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := workflow.Stop(stopCtx); err != nil {
		t.Fatalf("failed to stop workflow: %v", err)
	}

	if err := returnQueue.Stop(stopCtx); err != nil {
		t.Fatalf("failed to stop return queue: %v", err)
	}

	// Verify tasks were returned
	returnMu.Lock()
	returnCount := len(returnedTasks)
	returnMu.Unlock()

	if returnCount != len(tasks) {
		t.Errorf("expected %d tasks to be returned, got %d", len(tasks), returnCount)
	}

	// Verify returned tasks have correct state
	returnMu.Lock()
	for i, task := range returnedTasks {
		if task.GetState() != "processed" {
			t.Errorf("returned task %d state: expected 'processed', got '%s'", i, task.GetState())
		}
	}
	returnMu.Unlock()
}

func TestEventBasedWorkflow_BatchSend(t *testing.T) {
	// Create queue
	queue := NewBasicEventQueue[*EventTestTask](20, nil, 0)

	// Create state updater
	stateUpdater := &MockStateUpdater{
		updatedTasks: make([]*EventTestTask, 0),
	}

	// Create worker manager
	workerManager := NewBasicEventWorkerManager[*EventTestTask](
		&EventTestWorkerFactory{},
		3,
		20,
		stateUpdater,
		nil,
	)

	// Create workflow
	workflow := NewEventBasedWorkflow[*EventTestTask](queue, workerManager, stateUpdater)

	// Start workflow
	ctx := context.Background()
	if err := workflow.Start(ctx); err != nil {
		t.Fatalf("failed to start workflow: %v", err)
	}

	// Create batch of tasks
	tasks := make([]*EventTestTask, 10)
	for i := 0; i < 10; i++ {
		tasks[i] = &EventTestTask{
			id:         fmt.Sprintf("batch-%d", i),
			state:      "pending",
			processLog: make([]string, 0),
		}
	}

	// Submit batch
	if err := workflow.SubmitBatch(ctx, tasks); err != nil {
		t.Fatalf("failed to submit batch: %v", err)
	}

	// Wait for processing
	time.Sleep(200 * time.Millisecond)

	// Stop workflow
	stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := workflow.Stop(stopCtx); err != nil {
		t.Fatalf("failed to stop workflow: %v", err)
	}

	// Verify all tasks were processed
	processedCount := 0
	for _, task := range tasks {
		if task.GetState() == "processed" {
			processedCount++
		}
	}

	if processedCount != len(tasks) {
		t.Errorf("expected %d tasks to be processed, got %d", len(tasks), processedCount)
	}
}
