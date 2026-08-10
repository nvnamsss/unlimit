package efflux

import (
	"context"
	"fmt"
	"time"
)

// ProgressTask represents a task with multiple steps that can be executed progressively.
// The task repeats at each step until the last step is completed, with progress being
// saved frequently based on the implementation's requirements.
// Example usage:
//
//	task := &MyProgressTask{
//		steps: []string{"init", "process", "finalize"},
//		currentStep: 0,
//	}
//
//	for !task.IsCompleted() {
//		err := task.Execute(ctx)
//		if err != nil {
//			return err
//		}
//		task.SaveProgress()
//		task.NextStep()
//	}
//
// This interface is useful for implementing long-running tasks that need to be
// resumable and trackable across multiple execution cycles.
type ProgressTask interface {
	// GetID returns the unique identifier for this task.
	// This ID is used for persistence and tracking purposes.
	GetID() string

	// GetCurrentStep returns the current step number (0-based index).
	// Returns the index of the step currently being executed.
	GetCurrentStep() int

	// GetTotalSteps returns the total number of steps in this task.
	// This allows calculating progress percentage and determining completion.
	GetTotalSteps() int

	// GetStepName returns the name/description of the specified step.
	// Returns an empty string if the step index is invalid.
	GetStepName(stepIndex int) string

	// GetCurrentStepName returns the name/description of the current step.
	// This is a convenience method equivalent to GetStepName(GetCurrentStep()).
	GetCurrentStepName() string

	// Execute executes the current step of the task.
	// Returns an error if the step execution fails.
	// The implementation should handle the specific logic for each step.
	Execute(ctx context.Context) error

	// NextStep advances to the next step in the task.
	// Returns true if successfully advanced, false if already at the last step.
	NextStep() bool

	// PreviousStep moves back to the previous step in the task.
	// Returns true if successfully moved back, false if already at the first step.
	// This method is useful for error recovery scenarios.
	PreviousStep() bool

	// SetCurrentStep sets the current step to the specified index.
	// Returns an error if the step index is invalid.
	// This method is useful for resuming tasks from a saved state.
	SetCurrentStep(stepIndex int) error

	// IsCompleted returns true if all steps have been completed.
	// A task is considered completed when the current step is beyond the last step.
	IsCompleted() bool

	// GetProgress returns the completion progress as a percentage (0.0 to 1.0).
	// This is calculated as currentStep / totalSteps.
	GetProgress() float64

	// SaveProgress persists the current progress state.
	// The implementation determines how and where the progress is saved.
	// Returns an error if the save operation fails.
	SaveProgress() error

	// LoadProgress restores the progress state from persistent storage.
	// The implementation determines how and where the progress is loaded from.
	// Returns an error if the load operation fails.
	LoadProgress() error

	// Reset resets the task to its initial state (step 0).
	// This method is useful for restarting a task from the beginning.
	Reset()

	// GetStartTime returns the time when the task was started.
	// Returns zero time if the task hasn't been started yet.
	GetStartTime() time.Time

	// GetLastExecutionTime returns the time when the last step was executed.
	// Returns zero time if no step has been executed yet.
	GetLastExecutionTime() time.Time

	// GetEstimatedCompletion returns the estimated completion time based on
	// current progress and execution history. Returns zero time if estimation
	// is not available or reliable.
	GetEstimatedCompletion() time.Time

	// GetMetadata returns task-specific metadata as a key-value map.
	// This can include configuration, intermediate results, or other task-specific data.
	GetMetadata() map[string]interface{}

	// SetMetadata updates the task metadata with the provided key-value pairs.
	// Existing keys will be overwritten, new keys will be added.
	SetMetadata(metadata map[string]interface{})

	// GetErrorHistory returns a list of errors encountered during task execution.
	// This is useful for debugging and monitoring task reliability.
	GetErrorHistory() []error

	// CanRetry returns true if the current step can be retried after a failure.
	// The implementation can define retry logic based on error type or step characteristics.
	CanRetry() bool

	// ShouldSaveProgress returns true if progress should be saved at this point.
	// This allows implementations to control save frequency based on various factors
	// such as time elapsed, steps completed, or resource usage.
	ShouldSaveProgress() bool
}

// BaseProgressTask is a concrete implementation of the ProgressTask interface.
// It provides a basic implementation that can be used directly or embedded
// in custom task implementations.
// Example usage:
//
//	task := NewBaseProgressTask("task-001", []string{"init", "process", "cleanup"})
//	task.SetStepExecutor(func(ctx context.Context, stepIndex int, stepName string) error {
//		switch stepName {
//		case "init":
//			return initializeTask()
//		case "process":
//			return processData()
//		case "cleanup":
//			return cleanupResources()
//		default:
//			return fmt.Errorf("unknown step: %s", stepName)
//		}
//	})
//
// This implementation provides persistence, error tracking, and progress monitoring
// out of the box while allowing customization through the step executor function.
type BaseProgressTask struct {
	id                   string
	steps                []string
	currentStep          int
	startTime            time.Time
	lastExecutionTime    time.Time
	metadata             map[string]interface{}
	errorHistory         []error
	stepExecutor         func(ctx context.Context, stepIndex int, stepName string) error
	progressSaver        func() error
	progressLoader       func() error
	saveFrequencyChecker func() bool
}

// NewBaseProgressTask creates a new BaseProgressTask with the specified ID and steps.
// The task is initialized at step 0 and ready for execution.
// Example usage:
//
//	task := NewBaseProgressTask("data-migration-001", []string{
//		"validate_input",
//		"backup_data",
//		"transform_data",
//		"migrate_data",
//		"verify_migration",
//	})
//
// This function initializes all required fields with sensible defaults.
func NewBaseProgressTask(id string, steps []string) *BaseProgressTask {
	stepsCopy := make([]string, len(steps))
	copy(stepsCopy, steps)

	return &BaseProgressTask{
		id:           id,
		steps:        stepsCopy,
		currentStep:  0,
		metadata:     make(map[string]interface{}),
		errorHistory: make([]error, 0),
	}
}

// GetID returns the unique identifier for this task.
func (t *BaseProgressTask) GetID() string {
	return t.id
}

// GetCurrentStep returns the current step number (0-based index).
func (t *BaseProgressTask) GetCurrentStep() int {
	return t.currentStep
}

// GetTotalSteps returns the total number of steps in this task.
func (t *BaseProgressTask) GetTotalSteps() int {
	return len(t.steps)
}

// GetStepName returns the name/description of the specified step.
func (t *BaseProgressTask) GetStepName(stepIndex int) string {
	if stepIndex < 0 || stepIndex >= len(t.steps) {
		return ""
	}
	return t.steps[stepIndex]
}

// GetCurrentStepName returns the name/description of the current step.
func (t *BaseProgressTask) GetCurrentStepName() string {
	return t.GetStepName(t.currentStep)
}

// Execute executes the current step of the task.
func (t *BaseProgressTask) Execute(ctx context.Context) error {
	if t.IsCompleted() {
		return nil
	}

	if t.startTime.IsZero() {
		t.startTime = time.Now()
	}

	stepName := t.GetCurrentStepName()
	if t.stepExecutor == nil {
		return fmt.Errorf("no step executor configured for task %s", t.id)
	}

	err := t.stepExecutor(ctx, t.currentStep, stepName)
	t.lastExecutionTime = time.Now()

	if err != nil {
		t.errorHistory = append(t.errorHistory, err)
		return err
	}

	return nil
}

// NextStep advances to the next step in the task.
func (t *BaseProgressTask) NextStep() bool {
	if t.currentStep < len(t.steps)-1 {
		t.currentStep++
		return true
	}
	if t.currentStep == len(t.steps)-1 {
		t.currentStep++ // Move beyond last step to mark as completed
		return true
	}
	return false
}

// PreviousStep moves back to the previous step in the task.
func (t *BaseProgressTask) PreviousStep() bool {
	if t.currentStep > 0 {
		t.currentStep--
		return true
	}
	return false
}

// SetCurrentStep sets the current step to the specified index.
func (t *BaseProgressTask) SetCurrentStep(stepIndex int) error {
	if stepIndex < 0 || stepIndex > len(t.steps) {
		return fmt.Errorf("invalid step index %d, must be between 0 and %d", stepIndex, len(t.steps))
	}
	t.currentStep = stepIndex
	return nil
}

// IsCompleted returns true if all steps have been completed.
func (t *BaseProgressTask) IsCompleted() bool {
	return t.currentStep >= len(t.steps)
}

// GetProgress returns the completion progress as a percentage (0.0 to 1.0).
func (t *BaseProgressTask) GetProgress() float64 {
	if len(t.steps) == 0 {
		return 1.0
	}
	progress := float64(t.currentStep) / float64(len(t.steps))
	if progress > 1.0 {
		return 1.0
	}
	return progress
}

// SaveProgress persists the current progress state.
func (t *BaseProgressTask) SaveProgress() error {
	if t.progressSaver != nil {
		return t.progressSaver()
	}
	// Default implementation: no-op
	return nil
}

// LoadProgress restores the progress state from persistent storage.
func (t *BaseProgressTask) LoadProgress() error {
	if t.progressLoader != nil {
		return t.progressLoader()
	}
	// Default implementation: no-op
	return nil
}

// Reset resets the task to its initial state (step 0).
func (t *BaseProgressTask) Reset() {
	t.currentStep = 0
	t.startTime = time.Time{}
	t.lastExecutionTime = time.Time{}
	t.errorHistory = make([]error, 0)
}

// GetStartTime returns the time when the task was started.
func (t *BaseProgressTask) GetStartTime() time.Time {
	return t.startTime
}

// GetLastExecutionTime returns the time when the last step was executed.
func (t *BaseProgressTask) GetLastExecutionTime() time.Time {
	return t.lastExecutionTime
}

// GetEstimatedCompletion returns the estimated completion time based on
// current progress and execution history.
func (t *BaseProgressTask) GetEstimatedCompletion() time.Time {
	if t.startTime.IsZero() || t.currentStep == 0 || len(t.steps) == 0 {
		return time.Time{}
	}

	elapsed := time.Since(t.startTime)
	avgTimePerStep := elapsed / time.Duration(t.currentStep)
	remainingSteps := len(t.steps) - t.currentStep
	estimatedRemainingTime := avgTimePerStep * time.Duration(remainingSteps)

	return time.Now().Add(estimatedRemainingTime)
}

// GetMetadata returns task-specific metadata as a key-value map.
func (t *BaseProgressTask) GetMetadata() map[string]interface{} {
	// Return a copy to prevent external modification
	metadata := make(map[string]interface{})
	for k, v := range t.metadata {
		metadata[k] = v
	}
	return metadata
}

// SetMetadata updates the task metadata with the provided key-value pairs.
func (t *BaseProgressTask) SetMetadata(metadata map[string]interface{}) {
	if t.metadata == nil {
		t.metadata = make(map[string]interface{})
	}
	for k, v := range metadata {
		t.metadata[k] = v
	}
}

// GetErrorHistory returns a list of errors encountered during task execution.
func (t *BaseProgressTask) GetErrorHistory() []error {
	// Return a copy to prevent external modification
	errors := make([]error, len(t.errorHistory))
	copy(errors, t.errorHistory)
	return errors
}

// CanRetry returns true if the current step can be retried after a failure.
func (t *BaseProgressTask) CanRetry() bool {
	// Default implementation: allow retry if not completed
	return !t.IsCompleted()
}

// ShouldSaveProgress returns true if progress should be saved at this point.
func (t *BaseProgressTask) ShouldSaveProgress() bool {
	if t.saveFrequencyChecker != nil {
		return t.saveFrequencyChecker()
	}
	// Default implementation: save after each step
	return true
}

// SetStepExecutor sets the function that will be called to execute each step.
// The executor function receives the context, step index, and step name.
// Example usage:
//
//	task.SetStepExecutor(func(ctx context.Context, stepIndex int, stepName string) error {
//		log.Printf("Executing step %d: %s", stepIndex, stepName)
//		// Implementation-specific logic here
//		return nil
//	})
//
// This method allows customization of step execution logic while reusing
// the base progress tracking functionality.
func (t *BaseProgressTask) SetStepExecutor(executor func(ctx context.Context, stepIndex int, stepName string) error) {
	t.stepExecutor = executor
}

// SetProgressSaver sets the function that will be called to save progress.
// Example usage:
//
//	task.SetProgressSaver(func() error {
//		return database.SaveTaskProgress(task.GetID(), task.GetCurrentStep())
//	})
//
// This method allows customization of progress persistence while reusing
// the base task management functionality.
func (t *BaseProgressTask) SetProgressSaver(saver func() error) {
	t.progressSaver = saver
}

// SetProgressLoader sets the function that will be called to load progress.
// Example usage:
//
//	task.SetProgressLoader(func() error {
//		step, err := database.LoadTaskProgress(task.GetID())
//		if err != nil {
//			return err
//		}
//		return task.SetCurrentStep(step)
//	})
//
// This method allows customization of progress restoration while reusing
// the base task management functionality.
func (t *BaseProgressTask) SetProgressLoader(loader func() error) {
	t.progressLoader = loader
}

// SetSaveFrequencyChecker sets the function that determines when to save progress.
// Example usage:
//
//	task.SetSaveFrequencyChecker(func() bool {
//		return time.Since(task.GetLastExecutionTime()) > 5*time.Minute
//	})
//
// This method allows customization of save frequency based on time, steps,
// or other criteria while reusing the base progress functionality.
func (t *BaseProgressTask) SetSaveFrequencyChecker(checker func() bool) {
	t.saveFrequencyChecker = checker
}
