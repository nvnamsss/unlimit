package state

import (
	"errors"
	"fmt"
	"sync"
)

// StateMachine manages states and transitions between them
type StateMachine struct {
	states       map[StateID]State
	currentState State
	initialState StateID
	mutex        sync.RWMutex
}

// NewStateMachine creates a new state machine with the given initial state
func NewStateMachine() *StateMachine {
	return &StateMachine{
		states: make(map[StateID]State),
	}
}

// AddState adds a state to the state machine
func (sm *StateMachine) AddState(state State) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if state == nil {
		return errors.New("cannot add nil state")
	}

	id := state.ID()
	if _, exists := sm.states[id]; exists {
		return fmt.Errorf("state with id %q already exists", id)
	}

	sm.states[id] = state
	return nil
}

// SetInitialState sets the initial state of the state machine
func (sm *StateMachine) SetInitialState(stateID StateID) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if _, exists := sm.states[stateID]; !exists {
		return fmt.Errorf("state with id %q does not exist", stateID)
	}

	sm.initialState = stateID
	return nil
}

// Initialize initializes the state machine by entering the initial state
func (sm *StateMachine) Initialize(data interface{}) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if sm.initialState == "" {
		return errors.New("initial state not set")
	}

	if len(sm.states) == 0 {
		return errors.New("no states added to state machine")
	}

	initialState, exists := sm.states[sm.initialState]
	if !exists {
		return fmt.Errorf("initial state %q not found", sm.initialState)
	}

	sm.currentState = initialState
	return initialState.OnEnter(data)
}

// SendEvent sends an event to the current state and handles any resulting state transition
func (sm *StateMachine) SendEvent(event Event, data interface{}) error {
	sm.mutex.Lock()

	if sm.currentState == nil {
		sm.mutex.Unlock()
		return errors.New("state machine not initialized")
	}

	currentState := sm.currentState
	nextStateID, handled := currentState.OnEvent(event, data)
	sm.mutex.Unlock()

	if !handled {
		return fmt.Errorf("event %q not handled by current state %q", event, currentState.ID())
	}

	// If the state doesn't change, we're done
	if nextStateID == currentState.ID() {
		return nil
	}

	// Look up the next state
	sm.mutex.Lock()
	nextState, exists := sm.states[nextStateID]
	sm.mutex.Unlock()
	if !exists {
		return fmt.Errorf("next state %q not found", nextStateID)
	}

	// Exit the current state
	if err := currentState.OnExit(); err != nil {
		return fmt.Errorf("error exiting state %q: %w", currentState.ID(), err)
	}

	// Enter the new state
	if err := nextState.OnEnter(data); err != nil {
		return fmt.Errorf("error entering state %q: %w", nextState.ID(), err)
	}

	// Update current state
	sm.mutex.Lock()
	sm.currentState = nextState
	sm.mutex.Unlock()

	return nil
}

// CurrentState returns the ID of the current state
func (sm *StateMachine) CurrentState() (StateID, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	if sm.currentState == nil {
		return "", errors.New("state machine not initialized")
	}

	return sm.currentState.ID(), nil
}

// HasState checks if a state with the given ID exists in the state machine
func (sm *StateMachine) HasState(stateID StateID) bool {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	_, exists := sm.states[stateID]
	return exists
}
