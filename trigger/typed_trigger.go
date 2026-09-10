package trigger

import (
	"fmt"

	"github.com/nvnamsss/unlimit/caster"
)

// TypedTrigger wraps a Trigger with type-safe methods
type TypedTrigger[T any] struct {
	trigger         *Trigger
	typedConditions []TypedCondition[T]
	typedActions    []TypedAction[T]
}

// NewTypedTrigger creates a new type-safe trigger
func NewTypedTrigger[T any](id, name string) *TypedTrigger[T] {
	return &TypedTrigger[T]{
		trigger:         NewTrigger(id, name),
		typedConditions: make([]TypedCondition[T], 0),
		typedActions:    make([]TypedAction[T], 0),
	}
}

// AddEventType adds an event type that this trigger should listen for
func (tt *TypedTrigger[T]) AddEventType(eventType string) *TypedTrigger[T] {
	tt.trigger.AddEventType(eventType)
	return tt
}

// AddTypedCondition adds a type-safe condition
func (tt *TypedTrigger[T]) AddTypedCondition(condition TypedCondition[T]) *TypedTrigger[T] {
	tt.typedConditions = append(tt.typedConditions, condition)

	// Add a wrapper condition to the underlying trigger
	tt.trigger.AddCondition(func(event Event) bool {
		if data, ok := event.Data().(T); ok {
			return condition(data)
		}

		// type mismatch, marshal if needed
		if data, err := caster.MarshalCast[T](event.Data()); err == nil {
			return condition(data)
		}
		return false
	})

	return tt
}

// AddTypedAction adds a type-safe action
func (tt *TypedTrigger[T]) AddTypedAction(action TypedAction[T]) *TypedTrigger[T] {
	tt.typedActions = append(tt.typedActions, action)

	// Add a wrapper action to the underlying trigger
	tt.trigger.AddAction(func(event Event) error {
		if data, ok := event.Data().(T); ok {
			return action(data)
		}

		if data, err := caster.MarshalCast[T](event.Data()); err == nil {
			return action(data)
		}
		return fmt.Errorf("event data type mismatch")
	})

	return tt
}

// AddCondition adds a regular (non-typed) condition for backward compatibility
func (tt *TypedTrigger[T]) AddCondition(condition Condition) *TypedTrigger[T] {
	tt.trigger.AddCondition(condition)
	return tt
}

// AddAction adds a regular (non-typed) action for backward compatibility
func (tt *TypedTrigger[T]) AddAction(action Action) *TypedTrigger[T] {
	tt.trigger.AddAction(action)
	return tt
}

// SetPriority sets the execution priority
func (tt *TypedTrigger[T]) SetPriority(priority Priority) *TypedTrigger[T] {
	tt.trigger.SetPriority(priority)
	return tt
}

// SetEnabled enables or disables the trigger
func (tt *TypedTrigger[T]) SetEnabled(enabled bool) *TypedTrigger[T] {
	tt.trigger.SetEnabled(enabled)
	return tt
}

// SetExecuteOnce sets whether the trigger should only execute once
func (tt *TypedTrigger[T]) SetExecuteOnce(once bool) *TypedTrigger[T] {
	tt.trigger.SetExecuteOnce(once)
	return tt
}

// SetDescription sets the trigger description
func (tt *TypedTrigger[T]) SetDescription(description string) *TypedTrigger[T] {
	tt.trigger.SetDescription(description)
	return tt
}

// Trigger returns the underlying Trigger for use with TriggerManager
func (tt *TypedTrigger[T]) Trigger() *Trigger {
	return tt.trigger
}

// ID returns the trigger ID
func (tt *TypedTrigger[T]) ID() string {
	return tt.trigger.ID
}

// Name returns the trigger name
func (tt *TypedTrigger[T]) Name() string {
	return tt.trigger.Name
}

// IsExecuted returns whether the trigger has been executed
func (tt *TypedTrigger[T]) IsExecuted() bool {
	return tt.trigger.IsExecuted()
}

// Reset resets the execution state
func (tt *TypedTrigger[T]) Reset() {
	tt.trigger.Reset()
}

// Execute executes the trigger with a typed event
func (tt *TypedTrigger[T]) Execute(event *TypedEvent[T]) error {
	return tt.trigger.Execute(event)
}

// CanExecute checks if the trigger can execute for the given typed event
func (tt *TypedTrigger[T]) CanExecute(event *TypedEvent[T]) bool {
	return tt.trigger.CanExecute(event)
}

// Helper functions for creating typed triggers with common patterns

// NewTypedTriggerWithFilter creates a typed trigger with a data filter
func NewTypedTriggerWithFilter[T any](id, name, eventType string, filter TypedCondition[T]) *TypedTrigger[T] {
	return NewTypedTrigger[T](id, name).
		AddEventType(eventType).
		AddTypedCondition(filter)
}

// NewTypedTriggerWithAction creates a typed trigger with a single action
func NewTypedTriggerWithAction[T any](id, name, eventType string, action TypedAction[T]) *TypedTrigger[T] {
	return NewTypedTrigger[T](id, name).
		AddEventType(eventType).
		AddTypedAction(action)
}

// NewTypedTriggerComplete creates a fully configured typed trigger
func NewTypedTriggerComplete[T any](id, name, eventType string, condition TypedCondition[T], action TypedAction[T]) *TypedTrigger[T] {
	return NewTypedTrigger[T](id, name).
		AddEventType(eventType).
		AddTypedCondition(condition).
		AddTypedAction(action)
}
