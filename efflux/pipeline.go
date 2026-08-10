package efflux

import (
	"context"
	"errors"
	"fmt"
)

// Package-level error variables for common pipeline errors
var (
	// ErrCycleDetected is returned when a cycle is detected during pipeline execution
	ErrCycleDetected = errors.New("cycle detected in pipeline execution")

	// ErrNoRootStage is returned when attempting to execute or validate a pipeline without a root stage
	ErrNoRootStage = errors.New("pipeline has no root stage")

	// ErrUnreachableStages is returned when the pipeline contains stages that cannot be reached from the root
	ErrUnreachableStages = errors.New("unreachable stages found in pipeline")
)

type PipelineFunc[T any] func(ctx context.Context, v T) error
type ConditionFunc[T any] func(ctx context.Context, v T, stage *Stage[T]) bool

// StageTransition represents a conditional transition to the next stage.
// It contains both the target stage and the condition that must be met
// for the transition to occur.
type StageTransition[T any] struct {
	// Condition determines whether this transition should be taken
	Condition ConditionFunc[T]
	// Target is the next stage to execute if the condition is met
	Target *Stage[T]
	// Name is an optional description of this transition for debugging
	Name string
}

// Pipeline represents a processing pipeline that executes a tree of stages.
// Each stage can have multiple conditional transitions to other stages, allowing
// for complex branching logic based on the processed data and context.
// Example usage:
//
//	pipeline := NewBasePipeline[UserData]()
//	validateStage := &Stage[UserData]{
//		Name: "validate",
//		Func: func(ctx context.Context, user UserData) error {
//			return validateUser(user)
//		},
//	}
//	saveStage := &Stage[UserData]{Name: "save", Func: saveUser}
//	errorStage := &Stage[UserData]{Name: "error", Func: handleError}
//
//	validateStage.AddTransition("success", saveStage, func(ctx context.Context, user UserData) bool {
//		return user.IsValid
//	})
//	validateStage.AddTransition("error", errorStage, func(ctx context.Context, user UserData) bool {
//		return !user.IsValid
//	})
//
//	pipeline.SetRootStage(validateStage)
//	err := pipeline.Execute(ctx, userData)
//
// This interface is useful for implementing complex processing workflows with
// conditional branching, error handling paths, and dynamic execution flows.
type Pipeline[T any] interface {
	// Execute runs the pipeline starting from the root stage, following transitions
	// based on conditions until no more valid transitions are found.
	Execute(ctx context.Context, v T) error

	// SetRootStage sets the starting stage for pipeline execution.
	// This stage will be executed first when Execute is called.
	SetRootStage(stage *Stage[T]) Pipeline[T]

	// GetRootStage returns the current root stage of the pipeline.
	// Returns nil if no root stage has been set.
	GetRootStage() *Stage[T]

	// AddStage adds a stage to the pipeline's stage registry for reference.
	// This doesn't affect execution flow but helps with stage management.
	AddStage(stage *Stage[T]) Pipeline[T]

	// GetStage retrieves a stage by name from the pipeline's registry.
	// Returns nil if no stage with the given name is found.
	GetStage(name string) *Stage[T]

	// GetAllStages returns all stages in the pipeline registry.
	// This method is useful for debugging and pipeline introspection.
	GetAllStages() []*Stage[T]

	// Clear removes all stages and resets the pipeline to empty state.
	// This method is useful for resetting a pipeline for reuse.
	Clear()

	// Validate checks if the pipeline configuration is valid.
	// Returns an error if there are issues like unreachable stages or cycles.
	Validate() error
}

type Stage[T any] struct {
	// Name is the unique identifier for this stage
	Name string
	// Func is the processing function executed by this stage
	Func PipelineFunc[T]
	// Transitions is a list of conditional transitions to other stages
	Transitions []*StageTransition[T]
}

// AddTransition adds a conditional transition to another stage.
// The transition will be evaluated when this stage completes successfully.
// Example usage:
//
//	stage.AddTransition("success", nextStage, func(ctx context.Context, v T) bool {
//		return v.IsValid()
//	})
//
// Multiple transitions can be added to create branching logic.
func (s *Stage[T]) AddTransition(name string, target *Stage[T], condition ConditionFunc[T]) *Stage[T] {
	transition := &StageTransition[T]{
		Name:      name,
		Target:    target,
		Condition: condition,
	}
	s.Transitions = append(s.Transitions, transition)
	return s
}

// GetNextStage evaluates all transitions and returns the first stage whose condition is met.
// Returns nil if no transition conditions are satisfied.
// Example usage:
//
//	nextStage := currentStage.GetNextStage(ctx, data)
//	if nextStage != nil {
//		// Continue to next stage
//	}
//
// This method enables dynamic stage selection based on runtime conditions.
func (s *Stage[T]) GetNextStage(ctx context.Context, v T) *Stage[T] {
	for _, transition := range s.Transitions {
		if transition.Condition(ctx, v, s) {
			return transition.Target
		}
	}
	return nil
}

// GetTransitions returns all transitions defined for this stage.
// This method is useful for pipeline introspection and debugging.
func (s *Stage[T]) GetTransitions() []*StageTransition[T] {
	transitions := make([]*StageTransition[T], len(s.Transitions))
	copy(transitions, s.Transitions)
	return transitions
}

// BasePipeline is a concrete implementation of the Pipeline interface.
// It manages a tree of stages with conditional transitions, allowing for
// complex branching execution flows.
// Example usage:
//
//	pipeline := NewBasePipeline[OrderData]()
//
//	validateStage := &Stage[OrderData]{Name: "validate", Func: validateOrder}
//	processStage := &Stage[OrderData]{Name: "process", Func: processOrder}
//	errorStage := &Stage[OrderData]{Name: "error", Func: handleError}
//
//	validateStage.AddTransition("valid", processStage, func(ctx context.Context, order OrderData) bool {
//		return order.IsValid
//	})
//	validateStage.AddTransition("invalid", errorStage, func(ctx context.Context, order OrderData) bool {
//		return !order.IsValid
//	})
//
//	pipeline.SetRootStage(validateStage)
//	err := pipeline.Execute(ctx, orderData)
//
// This implementation provides conditional branching, cycle detection,
// and comprehensive stage management capabilities.
type BasePipeline[T any] struct {
	root   *Stage[T]
	stages map[string]*Stage[T]
}

// NewBasePipeline creates a new empty pipeline with tree-like stage support.
// The pipeline is ready to have stages added and a root stage configured.
// Example usage:
//
//	pipeline := NewBasePipeline[MyDataType]()
//	pipeline.AddStage(myStage)
//	pipeline.SetRootStage(myStage)
//	pipeline.Execute(ctx, myData)
//
// This function initializes an empty pipeline with no stages or root.
func NewBasePipeline[T any]() *BasePipeline[T] {
	return &BasePipeline[T]{
		stages: make(map[string]*Stage[T]),
	}
}

// Execute runs the pipeline starting from the root stage and following transitions
// based on conditions until no more valid transitions are found or an error occurs.
// Example usage:
//
//	err := pipeline.Execute(ctx, userData)
//	if err != nil {
//		log.Printf("Pipeline execution failed: %v", err)
//	}
//
// This method implements the Pipeline interface's Execute method with tree traversal.
func (p *BasePipeline[T]) Execute(ctx context.Context, v T) error {
	if p.root == nil {
		return nil // Empty pipeline executes successfully
	}

	current := p.root
	visited := make(map[*Stage[T]]bool) // Prevent infinite cycles

	for current != nil {
		// Check for cycles
		if visited[current] {
			return fmt.Errorf("%w at stage: %s", ErrCycleDetected, current.Name)
		}
		visited[current] = true

		// Execute current stage
		if err := current.Func(ctx, v); err != nil {
			return fmt.Errorf("stage '%s' failed: %w", current.Name, err)
		}

		// Find next stage based on conditions
		current = current.GetNextStage(ctx, v)
	}

	return nil
}

// SetRootStage sets the starting stage for pipeline execution.
// Returns the pipeline instance for method chaining.
// Example usage:
//
//	pipeline.SetRootStage(validateStage)
//
// This method implements the Pipeline interface's SetRootStage method.
func (p *BasePipeline[T]) SetRootStage(stage *Stage[T]) Pipeline[T] {
	p.root = stage
	if stage != nil {
		p.stages[stage.Name] = stage // Ensure root stage is in registry
	}
	return p
}

// GetRootStage returns the current root stage of the pipeline.
// Returns nil if no root stage has been set.
func (p *BasePipeline[T]) GetRootStage() *Stage[T] {
	return p.root
}

// AddStage adds a stage to the pipeline's stage registry for reference.
// This doesn't affect execution flow but helps with stage management.
// Returns the pipeline instance for method chaining.
// Example usage:
//
//	pipeline.AddStage(validateStage).AddStage(processStage)
//
// This method implements the Pipeline interface's AddStage method.
func (p *BasePipeline[T]) AddStage(stage *Stage[T]) Pipeline[T] {
	if stage != nil {
		p.stages[stage.Name] = stage
	}
	return p
}

// GetStage retrieves a stage by name from the pipeline's registry.
// Returns nil if no stage with the given name is found.
// Example usage:
//
//	validateStage := pipeline.GetStage("validate")
//	if validateStage != nil {
//		// Use the stage
//	}
//
// This method implements the Pipeline interface's GetStage method.
func (p *BasePipeline[T]) GetStage(name string) *Stage[T] {
	return p.stages[name]
}

// GetAllStages returns all stages in the pipeline registry.
// This method is useful for debugging and pipeline introspection.
// Example usage:
//
//	stages := pipeline.GetAllStages()
//	for _, stage := range stages {
//		fmt.Printf("Stage: %s\n", stage.Name)
//	}
//
// This method implements the Pipeline interface's GetAllStages method.
func (p *BasePipeline[T]) GetAllStages() []*Stage[T] {
	stages := make([]*Stage[T], 0, len(p.stages))
	for _, stage := range p.stages {
		stages = append(stages, stage)
	}
	return stages
}

// Clear removes all stages and resets the pipeline to empty state.
// This method is useful for resetting a pipeline for reuse.
// Example usage:
//
//	pipeline.Clear()
//	pipeline.AddStage(newStage)
//
// After calling Clear, the pipeline can be used as if it were newly created.
func (p *BasePipeline[T]) Clear() {
	p.root = nil
	p.stages = make(map[string]*Stage[T])
}

// Validate checks if the pipeline configuration is valid.
// Returns an error if there are issues like unreachable stages or cycles.
// Example usage:
//
//	if err := pipeline.Validate(); err != nil {
//		log.Printf("Pipeline validation failed: %v", err)
//	}
//
// This method implements the Pipeline interface's Validate method.
func (p *BasePipeline[T]) Validate() error {
	if p.root == nil {
		return ErrNoRootStage
	}

	// Check for unreachable stages
	reachable := make(map[*Stage[T]]bool)
	p.markReachable(p.root, reachable)

	unreachable := make([]string, 0)
	for name, stage := range p.stages {
		if !reachable[stage] {
			unreachable = append(unreachable, name)
		}
	}

	if len(unreachable) > 0 {
		return fmt.Errorf("%w: %v", ErrUnreachableStages, unreachable)
	}

	return nil
}

// markReachable is a helper method for Validate that marks all reachable stages
func (p *BasePipeline[T]) markReachable(stage *Stage[T], reachable map[*Stage[T]]bool) {
	if stage == nil || reachable[stage] {
		return
	}

	reachable[stage] = true
	for _, transition := range stage.Transitions {
		p.markReachable(transition.Target, reachable)
	}
}
