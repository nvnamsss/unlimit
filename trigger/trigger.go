package trigger

import (
	"fmt"
)

// Priority defines the execution priority of triggers
type Priority int

const (
	PriorityLow Priority = iota
	PriorityNormal
	PriorityHigh
	PriorityCritical
)

// Trigger represents a single trigger with events, conditions, and actions
type Trigger struct {
	ID          string
	Name        string
	Description string
	EventTypes  []string
	Conditions  []Condition
	Actions     []Action
	Priority    Priority
	Enabled     bool
	ExecuteOnce bool
	executed    bool
}

// NewTrigger creates a new trigger
func NewTrigger(id, name string) *Trigger {
	return &Trigger{
		ID:          id,
		Name:        name,
		EventTypes:  make([]string, 0),
		Conditions:  make([]Condition, 0),
		Actions:     make([]Action, 0),
		Priority:    PriorityNormal,
		Enabled:     true,
		ExecuteOnce: false,
		executed:    false,
	}
}

// AddEventType adds an event type that this trigger should listen for
func (t *Trigger) AddEventType(eventType string) *Trigger {
	t.EventTypes = append(t.EventTypes, eventType)
	return t
}

// AddCondition adds a condition that must be met for the trigger to execute
func (t *Trigger) AddCondition(condition Condition) *Trigger {
	t.Conditions = append(t.Conditions, condition)
	return t
}

// AddAction adds an action to be executed when the trigger fires
func (t *Trigger) AddAction(action Action) *Trigger {
	t.Actions = append(t.Actions, action)
	return t
}

// SetPriority sets the execution priority of the trigger
func (t *Trigger) SetPriority(priority Priority) *Trigger {
	t.Priority = priority
	return t
}

// SetEnabled enables or disables the trigger
func (t *Trigger) SetEnabled(enabled bool) *Trigger {
	t.Enabled = enabled
	return t
}

// SetExecuteOnce sets whether the trigger should only execute once
func (t *Trigger) SetExecuteOnce(once bool) *Trigger {
	t.ExecuteOnce = once
	return t
}

// SetDescription sets the trigger description
func (t *Trigger) SetDescription(description string) *Trigger {
	t.Description = description
	return t
}

// CanExecute checks if the trigger can be executed for the given event
func (t *Trigger) CanExecute(event Event) bool {
	// t.mutex.RLock()
	// defer t.mutex.RUnlock()

	// Check if trigger is enabled
	if !t.Enabled {
		return false
	}

	// Check if already executed and should only execute once
	if t.ExecuteOnce && t.executed {
		return false
	}

	// Check if event type matches
	eventTypeMatches := false
	for _, eventType := range t.EventTypes {
		if eventType == event.Type() {
			eventTypeMatches = true
			break
		}
	}
	if !eventTypeMatches {
		return false
	}

	// Check all conditions
	for _, condition := range t.Conditions {
		if !condition(event) {
			return false
		}
	}

	return true
}

// Execute executes the trigger's actions for the given event
func (t *Trigger) Execute(event Event) error {
	if !t.CanExecute(event) {
		return fmt.Errorf("trigger %s cannot be executed for event %s", t.ID, event.Type())
	}

	// Execute all actions
	for i, action := range t.Actions {
		if err := action(event); err != nil {
			return fmt.Errorf("action %d in trigger %s failed: %w", i, t.ID, err)
		}
	}

	// Mark as executed if it should only execute once
	if t.ExecuteOnce {
		t.executed = true
	}

	return nil
}

// IsExecuted returns whether the trigger has been executed (only relevant for ExecuteOnce triggers)
func (t *Trigger) IsExecuted() bool {
	return t.executed
}

// Reset resets the execution state of the trigger
func (t *Trigger) Reset() {
	t.executed = false
}
