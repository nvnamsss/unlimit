package efflux

import (
	"context"
	"sync"
	"time"
)

// EventTask represents a task that can be processed in an event-based workflow
type EventTask interface {
	Task
	// GetState returns the current state of the task
	GetState() string
	// SetState updates the state of the task
	SetState(state string)
}

// TaskPuller defines the interface for pulling tasks from an external source
type TaskPuller[T EventTask] interface {
	// Pull retrieves tasks from an external source and returns them
	// Returns a slice of tasks and an error if the operation fails
	Pull(ctx context.Context) ([]T, error)
}

// TaskSender defines the interface for sending tasks to the queue
type TaskSender[T EventTask] interface {
	// Send adds a task to the queue
	Send(ctx context.Context, task T) error
	// SendBatch adds multiple tasks to the queue
	SendBatch(ctx context.Context, tasks []T) error
}

// EventQueue manages the task queue with pulling and sending capabilities
type EventQueue[T EventTask] interface {
	TaskPuller[T]
	TaskSender[T]
	// Start begins queue operations
	Start(ctx context.Context) error
	// Stop gracefully shuts down the queue
	Stop(ctx context.Context) error
	// Subscribe returns a channel to receive tasks from the queue
	Subscribe() <-chan T
}

// EventWorker processes tasks and updates their state
type EventWorker[T EventTask] interface {
	// Process executes the task and updates its state
	Process(ctx context.Context, task T) error
	// GetID returns the worker's unique identifier
	GetID() string
}

// EventWorkerFactory creates workers for event processing
type EventWorkerFactory[T EventTask] interface {
	// Create instantiates a new worker
	Create(workerID string) EventWorker[T]
}

// EventWorkerManager coordinates workers and manages task distribution
type EventWorkerManager[T EventTask] interface {
	// Start initializes the manager and its workers
	Start(ctx context.Context) error
	// Stop gracefully shuts down the manager
	Stop(ctx context.Context) error
	// Assign sends a task to be processed by workers
	Assign(ctx context.Context, task T) error
}

// StateUpdater handles state updates after task processing
type StateUpdater[T EventTask] interface {
	// Update updates the state of a task after processing
	Update(ctx context.Context, task T) error
}

// EventBasedWorkflow orchestrates an event-driven workflow where
// tasks flow from queue to workers and back to queue after state updates
type EventBasedWorkflow[T EventTask] struct {
	queue         EventQueue[T]
	workerManager EventWorkerManager[T]
	stateUpdater  StateUpdater[T]
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
	mu            sync.RWMutex
}

// NewEventBasedWorkflow creates a new event-based workflow
func NewEventBasedWorkflow[T EventTask](
	queue EventQueue[T],
	workerManager EventWorkerManager[T],
	stateUpdater StateUpdater[T],
) *EventBasedWorkflow[T] {
	return &EventBasedWorkflow[T]{
		queue:         queue,
		workerManager: workerManager,
		stateUpdater:  stateUpdater,
	}
}

// Start initializes the workflow and begins processing
func (w *EventBasedWorkflow[T]) Start(ctx context.Context) error {
	w.mu.Lock()
	if w.ctx != nil {
		w.mu.Unlock()
		return nil // already started
	}

	w.ctx, w.cancel = context.WithCancel(ctx)
	w.mu.Unlock()

	// Start the queue
	if err := w.queue.Start(w.ctx); err != nil {
		return err
	}

	// Start the worker manager
	if err := w.workerManager.Start(w.ctx); err != nil {
		return err
	}

	// Start the event loop
	w.wg.Add(1)
	go w.processEvents()

	return nil
}

// processEvents continuously processes tasks from the queue
func (w *EventBasedWorkflow[T]) processEvents() {
	defer w.wg.Done()

	taskChan := w.queue.Subscribe()

	for {
		select {
		case task := <-taskChan:
			// Assign task to worker manager
			if err := w.workerManager.Assign(w.ctx, task); err != nil {
				task.SetError(err)
			}
		case <-w.ctx.Done():
			return
		}
	}
}

// Stop gracefully shuts down the workflow
func (w *EventBasedWorkflow[T]) Stop(ctx context.Context) error {
	w.mu.Lock()
	if w.cancel != nil {
		w.cancel()
	}
	w.mu.Unlock()

	// Stop components
	if err := w.workerManager.Stop(ctx); err != nil {
		return err
	}

	if err := w.queue.Stop(ctx); err != nil {
		return err
	}

	w.wg.Wait()
	return nil
}

// Submit sends a task to the workflow queue
func (w *EventBasedWorkflow[T]) Submit(ctx context.Context, task T) error {
	return w.queue.Send(ctx, task)
}

// SubmitBatch sends multiple tasks to the workflow queue
func (w *EventBasedWorkflow[T]) SubmitBatch(ctx context.Context, tasks []T) error {
	return w.queue.SendBatch(ctx, tasks)
}

// BasicEventQueue is a default implementation of EventQueue
type BasicEventQueue[T EventTask] struct {
	queue        chan T
	puller       TaskPuller[T]
	pullInterval time.Duration
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	mu           sync.RWMutex
	maxQueueSize int
}

// NewBasicEventQueue creates a new basic event queue
func NewBasicEventQueue[T EventTask](
	maxQueueSize int,
	puller TaskPuller[T],
	pullInterval time.Duration,
) *BasicEventQueue[T] {
	return &BasicEventQueue[T]{
		queue:        make(chan T, maxQueueSize),
		puller:       puller,
		pullInterval: pullInterval,
		maxQueueSize: maxQueueSize,
	}
}

// Start begins queue operations
func (q *BasicEventQueue[T]) Start(ctx context.Context) error {
	q.mu.Lock()
	if q.ctx != nil {
		q.mu.Unlock()
		return nil // already started
	}

	q.ctx, q.cancel = context.WithCancel(ctx)
	q.mu.Unlock()

	// Start pulling tasks if puller is configured
	if q.puller != nil {
		q.wg.Add(1)
		go q.pullTasks()
	}

	return nil
}

// pullTasks periodically pulls tasks from the external source
func (q *BasicEventQueue[T]) pullTasks() {
	defer q.wg.Done()

	ticker := time.NewTicker(q.pullInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			tasks, err := q.puller.Pull(q.ctx)
			if err != nil {
				// Log error but continue
				continue
			}

			for _, task := range tasks {
				select {
				case q.queue <- task:
				case <-q.ctx.Done():
					return
				}
			}
		case <-q.ctx.Done():
			return
		}
	}
}

// Send adds a task to the queue
func (q *BasicEventQueue[T]) Send(ctx context.Context, task T) error {
	select {
	case q.queue <- task:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// SendBatch adds multiple tasks to the queue
func (q *BasicEventQueue[T]) SendBatch(ctx context.Context, tasks []T) error {
	for _, task := range tasks {
		if err := q.Send(ctx, task); err != nil {
			return err
		}
	}
	return nil
}

// Pull retrieves tasks (not typically used in this implementation as pulling is automatic)
func (q *BasicEventQueue[T]) Pull(ctx context.Context) ([]T, error) {
	if q.puller == nil {
		return nil, nil
	}
	return q.puller.Pull(ctx)
}

// Subscribe returns a channel to receive tasks from the queue
func (q *BasicEventQueue[T]) Subscribe() <-chan T {
	return q.queue
}

// Stop gracefully shuts down the queue
func (q *BasicEventQueue[T]) Stop(ctx context.Context) error {
	q.mu.Lock()
	if q.cancel != nil {
		q.cancel()
	}
	q.mu.Unlock()

	// Wait for pulling to stop
	done := make(chan struct{})
	go func() {
		q.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// BasicEventWorkerManager is a default implementation of EventWorkerManager
type BasicEventWorkerManager[T EventTask] struct {
	workers      []EventWorker[T]
	taskQueue    chan T
	stateUpdater StateUpdater[T]
	returnQueue  EventQueue[T]
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	mu           sync.RWMutex
	numWorkers   int
}

// NewBasicEventWorkerManager creates a new basic event worker manager
func NewBasicEventWorkerManager[T EventTask](
	factory EventWorkerFactory[T],
	numWorkers int,
	taskQueueSize int,
	stateUpdater StateUpdater[T],
	returnQueue EventQueue[T],
) *BasicEventWorkerManager[T] {
	workers := make([]EventWorker[T], numWorkers)
	for i := 0; i < numWorkers; i++ {
		workers[i] = factory.Create(string(rune('A' + i)))
	}

	return &BasicEventWorkerManager[T]{
		workers:      workers,
		taskQueue:    make(chan T, taskQueueSize),
		stateUpdater: stateUpdater,
		returnQueue:  returnQueue,
		numWorkers:   numWorkers,
	}
}

// Start initializes the manager and its workers
func (m *BasicEventWorkerManager[T]) Start(ctx context.Context) error {
	m.mu.Lock()
	if m.ctx != nil {
		m.mu.Unlock()
		return nil // already started
	}

	m.ctx, m.cancel = context.WithCancel(ctx)
	m.mu.Unlock()

	// Start worker goroutines
	for i, worker := range m.workers {
		m.wg.Add(1)
		go m.runWorker(i, worker)
	}

	return nil
}

// Assign sends a task to the manager's queue
func (m *BasicEventWorkerManager[T]) Assign(ctx context.Context, task T) error {
	select {
	case m.taskQueue <- task:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// runWorker processes tasks for a single worker
func (m *BasicEventWorkerManager[T]) runWorker(workerID int, worker EventWorker[T]) {
	defer m.wg.Done()

	for {
		select {
		case task := <-m.taskQueue:
			// Process the task
			if err := worker.Process(m.ctx, task); err != nil {
				task.SetError(err)
			}

			// Update state
			if m.stateUpdater != nil {
				if err := m.stateUpdater.Update(m.ctx, task); err != nil {
					task.SetError(err)
				}
			}

			// Send back to queue if configured
			if m.returnQueue != nil {
				_ = m.returnQueue.Send(m.ctx, task)
			}

		case <-m.ctx.Done():
			return
		}
	}
}

// Stop gracefully shuts down the manager
func (m *BasicEventWorkerManager[T]) Stop(ctx context.Context) error {
	m.mu.Lock()
	if m.cancel != nil {
		m.cancel()
	}
	m.mu.Unlock()

	// Wait for all workers to finish
	done := make(chan struct{})
	go func() {
		m.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
