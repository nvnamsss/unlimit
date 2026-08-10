package trigger

import (
	"time"
)

// Event represents the base interface for all events in the trigger system
type Event interface {
	// Type returns the event type identifier
	Type() string
	// Data returns the event payload
	Data() interface{}
	// Timestamp returns when the event occurred
	Timestamp() time.Time
	// Source returns the source that generated this event
	Source() string
}

// BaseEvent provides a basic implementation of the Event interface
type BaseEvent struct {
	EventType   string
	EventData   interface{}
	EventTime   time.Time
	EventSource string
}

// Type returns the event type
func (e *BaseEvent) Type() string {
	return e.EventType
}

// Data returns the event data
func (e *BaseEvent) Data() interface{} {
	return e.EventData
}

// Timestamp returns the event timestamp
func (e *BaseEvent) Timestamp() time.Time {
	return e.EventTime
}

// Source returns the event source
func (e *BaseEvent) Source() string {
	return e.EventSource
}

// NewBaseEvent creates a new base event
func NewBaseEvent(eventType string, data interface{}, source string) *BaseEvent {
	return &BaseEvent{
		EventType:   eventType,
		EventData:   data,
		EventTime:   time.Now(),
		EventSource: source,
	}
}

// EventListener is a function that can handle events
type EventListener func(event Event) error

// EventFilter is a function that determines if an event should be processed
type EventFilter func(event Event) bool

// TypedEvent wraps an Event interface with type-safe data access
type TypedEvent[T any] struct {
	Event
	typedData T
}

// NewTypedEvent creates a new typed event wrapper
func NewTypedEvent[T any](eventType string, data T, source string) *TypedEvent[T] {
	return &TypedEvent[T]{
		Event:     NewBaseEvent(eventType, data, source),
		typedData: data,
	}
}

// TypedData returns the event data with type safety
func (e *TypedEvent[T]) TypedData() T {
	return e.typedData
}

// WrapEvent wraps an existing Event with type-safe access
// Returns the typed wrapper and a boolean indicating success
func WrapEvent[T any](event Event) (*TypedEvent[T], bool) {
	if data, ok := event.Data().(T); ok {
		return &TypedEvent[T]{
			Event:     event,
			typedData: data,
		}, true
	}
	return nil, false
}

// TypedEventListener is a type-safe event listener
type TypedEventListener[T any] func(event *TypedEvent[T]) error

// ToEventListener converts a typed listener to a regular listener
func ToEventListener[T any](typedListener TypedEventListener[T]) EventListener {
	return func(event Event) error {
		if typedEvent, ok := WrapEvent[T](event); ok {
			return typedListener(typedEvent)
		}
		// Skip events that don't match the expected type
		return nil
	}
}

// Common event type constants
const (
	EventTypeTimer      = "timer"
	EventTypeCustom     = "custom"
	EventTypeSystem     = "system"
	EventTypeUser       = "user"
	EventTypeLifecycle  = "lifecycle"
	EventTypeValidation = "validation"
	EventTypeWorkflow   = "workflow"
)

// TimerEvent represents a timer-based event
type TimerEvent struct {
	*BaseEvent
	Duration time.Duration
	Periodic bool
}

// NewTimerEvent creates a new timer event
func NewTimerEvent(duration time.Duration, periodic bool, source string) *TimerEvent {
	data := map[string]interface{}{
		"duration": duration,
		"periodic": periodic,
	}

	return &TimerEvent{
		BaseEvent: NewBaseEvent(EventTypeTimer, data, source),
		Duration:  duration,
		Periodic:  periodic,
	}
}

// CustomEvent represents a custom user-defined event
type CustomEvent struct {
	*BaseEvent
	Category   string
	Properties map[string]interface{}
}

// NewCustomEvent creates a new custom event
func NewCustomEvent(category string, properties map[string]interface{}, source string) *CustomEvent {
	if properties == nil {
		properties = make(map[string]interface{})
	}

	data := map[string]interface{}{
		"category":   category,
		"properties": properties,
	}

	return &CustomEvent{
		BaseEvent:  NewBaseEvent(EventTypeCustom, data, source),
		Category:   category,
		Properties: properties,
	}
}

// SystemEvent represents a system-level event
type SystemEvent struct {
	*BaseEvent
	Level   string // info, warning, error, critical
	Message string
	Code    int
}

// NewSystemEvent creates a new system event
func NewSystemEvent(level, message string, code int, source string) *SystemEvent {
	data := map[string]interface{}{
		"level":   level,
		"message": message,
		"code":    code,
	}

	return &SystemEvent{
		BaseEvent: NewBaseEvent(EventTypeSystem, data, source),
		Level:     level,
		Message:   message,
		Code:      code,
	}
}

// UserEvent represents a user action event
type UserEvent struct {
	*BaseEvent
	UserID string
	Action string
	Target string
}

// NewUserEvent creates a new user event
func NewUserEvent(userID, action, target, source string) *UserEvent {
	data := map[string]interface{}{
		"user_id": userID,
		"action":  action,
		"target":  target,
	}

	return &UserEvent{
		BaseEvent: NewBaseEvent(EventTypeUser, data, source),
		UserID:    userID,
		Action:    action,
		Target:    target,
	}
}

// LifecycleEvent represents an application lifecycle event
type LifecycleEvent struct {
	*BaseEvent
	Phase string // startup, ready, shutdown, error
	State map[string]interface{}
}

// NewLifecycleEvent creates a new lifecycle event
func NewLifecycleEvent(phase string, state map[string]interface{}, source string) *LifecycleEvent {
	if state == nil {
		state = make(map[string]interface{})
	}

	data := map[string]interface{}{
		"phase": phase,
		"state": state,
	}

	return &LifecycleEvent{
		BaseEvent: NewBaseEvent(EventTypeLifecycle, data, source),
		Phase:     phase,
		State:     state,
	}
}

// ValidationEvent represents a validation event
type ValidationEvent struct {
	*BaseEvent
	Valid   bool
	Errors  []string
	Subject string
}

// NewValidationEvent creates a new validation event
func NewValidationEvent(valid bool, errors []string, subject, source string) *ValidationEvent {
	if errors == nil {
		errors = make([]string, 0)
	}

	data := map[string]interface{}{
		"valid":   valid,
		"errors":  errors,
		"subject": subject,
	}

	return &ValidationEvent{
		BaseEvent: NewBaseEvent(EventTypeValidation, data, source),
		Valid:     valid,
		Errors:    errors,
		Subject:   subject,
	}
}

// WorkflowEvent represents a workflow state change event
type WorkflowEvent struct {
	*BaseEvent
	WorkflowID string
	FromState  string
	ToState    string
	Context    map[string]interface{}
}

// NewWorkflowEvent creates a new workflow event
func NewWorkflowEvent(workflowID, fromState, toState string, context map[string]interface{}, source string) *WorkflowEvent {
	if context == nil {
		context = make(map[string]interface{})
	}

	data := map[string]interface{}{
		"workflow_id": workflowID,
		"from_state":  fromState,
		"to_state":    toState,
		"context":     context,
	}

	return &WorkflowEvent{
		BaseEvent:  NewBaseEvent(EventTypeWorkflow, data, source),
		WorkflowID: workflowID,
		FromState:  fromState,
		ToState:    toState,
		Context:    context,
	}
}

// Typed event data structures

// UserEventData represents the data for user events
type UserEventData struct {
	UserID string
	Action string
	Target string
}

// SystemEventData represents the data for system events
type SystemEventData struct {
	Level   string
	Message string
	Code    int
}

// TimerEventData represents the data for timer events
type TimerEventData struct {
	Duration time.Duration
	Periodic bool
}

// CustomEventData represents the data for custom events
type CustomEventData struct {
	Category   string
	Properties map[string]interface{}
}

// LifecycleEventData represents the data for lifecycle events
type LifecycleEventData struct {
	Phase string
	State map[string]interface{}
}

// ValidationEventData represents the data for validation events
type ValidationEventData struct {
	Valid   bool
	Errors  []string
	Subject string
}

// WorkflowEventData represents the data for workflow events
type WorkflowEventData struct {
	WorkflowID string
	FromState  string
	ToState    string
	Context    map[string]interface{}
}

// Typed event constructors

// NewTypedUserEvent creates a new typed user event
func NewTypedUserEvent(userID, action, target, source string) *TypedEvent[UserEventData] {
	data := UserEventData{
		UserID: userID,
		Action: action,
		Target: target,
	}
	return NewTypedEvent(EventTypeUser, data, source)
}

// NewTypedSystemEvent creates a new typed system event
func NewTypedSystemEvent(level, message string, code int, source string) *TypedEvent[SystemEventData] {
	data := SystemEventData{
		Level:   level,
		Message: message,
		Code:    code,
	}
	return NewTypedEvent(EventTypeSystem, data, source)
}

// NewTypedTimerEvent creates a new typed timer event
func NewTypedTimerEvent(duration time.Duration, periodic bool, source string) *TypedEvent[TimerEventData] {
	data := TimerEventData{
		Duration: duration,
		Periodic: periodic,
	}
	return NewTypedEvent(EventTypeTimer, data, source)
}

// NewTypedCustomEvent creates a new typed custom event
func NewTypedCustomEvent(category string, properties map[string]interface{}, source string) *TypedEvent[CustomEventData] {
	if properties == nil {
		properties = make(map[string]interface{})
	}
	data := CustomEventData{
		Category:   category,
		Properties: properties,
	}
	return NewTypedEvent(EventTypeCustom, data, source)
}

// NewTypedLifecycleEvent creates a new typed lifecycle event
func NewTypedLifecycleEvent(phase string, state map[string]interface{}, source string) *TypedEvent[LifecycleEventData] {
	if state == nil {
		state = make(map[string]interface{})
	}
	data := LifecycleEventData{
		Phase: phase,
		State: state,
	}
	return NewTypedEvent(EventTypeLifecycle, data, source)
}

// NewTypedValidationEvent creates a new typed validation event
func NewTypedValidationEvent(valid bool, errors []string, subject, source string) *TypedEvent[ValidationEventData] {
	if errors == nil {
		errors = make([]string, 0)
	}
	data := ValidationEventData{
		Valid:   valid,
		Errors:  errors,
		Subject: subject,
	}
	return NewTypedEvent(EventTypeValidation, data, source)
}

// NewTypedWorkflowEvent creates a new typed workflow event
func NewTypedWorkflowEvent(workflowID, fromState, toState string, context map[string]interface{}, source string) *TypedEvent[WorkflowEventData] {
	if context == nil {
		context = make(map[string]interface{})
	}
	data := WorkflowEventData{
		WorkflowID: workflowID,
		FromState:  fromState,
		ToState:    toState,
		Context:    context,
	}
	return NewTypedEvent(EventTypeWorkflow, data, source)
}

// Helper functions for converting existing events to typed events

// ToTypedUserEvent converts a UserEvent to TypedEvent
func ToTypedUserEvent(event *UserEvent) *TypedEvent[UserEventData] {
	data := UserEventData{
		UserID: event.UserID,
		Action: event.Action,
		Target: event.Target,
	}
	return &TypedEvent[UserEventData]{
		Event:     event,
		typedData: data,
	}
}

// ToTypedSystemEvent converts a SystemEvent to TypedEvent
func ToTypedSystemEvent(event *SystemEvent) *TypedEvent[SystemEventData] {
	data := SystemEventData{
		Level:   event.Level,
		Message: event.Message,
		Code:    event.Code,
	}
	return &TypedEvent[SystemEventData]{
		Event:     event,
		typedData: data,
	}
}
