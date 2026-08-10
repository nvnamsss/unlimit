package state

// Event represents an event that can trigger state transitions
type Event string

// StateID is a unique identifier for a state
type StateID string

// State interface defines the behavior of a state
type State interface {
	// ID returns the state's unique identifier
	ID() StateID

	// OnEnter is called when entering the state
	OnEnter(data interface{}) error

	// OnExit is called when exiting the state
	OnExit() error

	// OnEvent is called when an event is received in the current state
	// Returns the next state ID and whether the event was handled
	OnEvent(event Event, data interface{}) (StateID, bool)
}

// BaseState provides a default implementation of the State interface
type BaseState struct {
	id       StateID
	handlers map[Event]func(data interface{}) StateID
}

// NewBaseState creates a new base state with the given ID
func NewBaseState(id StateID) *BaseState {
	return &BaseState{
		id:       id,
		handlers: make(map[Event]func(data interface{}) StateID),
	}
}

// ID returns the state's unique identifier
func (s *BaseState) ID() StateID {
	return s.id
}

// OnEnter is called when entering the state
func (s *BaseState) OnEnter(data interface{}) error {
	return nil // Default implementation does nothing
}

// OnExit is called when exiting the state
func (s *BaseState) OnExit() error {
	return nil // Default implementation does nothing
}

// AddEventHandler adds a handler for a specific event
func (s *BaseState) AddEventHandler(event Event, handler func(data interface{}) StateID) {
	s.handlers[event] = handler
}

// OnEvent processes events and returns the next state ID if the event is handled
func (s *BaseState) OnEvent(event Event, data interface{}) (StateID, bool) {
	if handler, exists := s.handlers[event]; exists {
		return handler(data), true
	}
	return s.id, false
}
