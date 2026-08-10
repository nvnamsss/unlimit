package state

import (
	"errors"
	"sync"
	"testing"
)

// MockState implements the State interface for testing
type MockState struct {
	id              StateID
	onEnterCalled   bool
	onExitCalled    bool
	onEnterError    error
	onExitError     error
	onEventResponse StateID
	onEventHandled  bool
	onEventError    error
	receivedData    interface{}
	receivedEvent   Event
}

func NewMockState(id StateID) *MockState {
	return &MockState{
		id:              id,
		onEventResponse: id,
		onEventHandled:  false,
	}
}

func (m *MockState) ID() StateID {
	return m.id
}

func (m *MockState) OnEnter(data interface{}) error {
	m.onEnterCalled = true
	m.receivedData = data
	return m.onEnterError
}

func (m *MockState) OnExit() error {
	m.onExitCalled = true
	return m.onExitError
}

func (m *MockState) OnEvent(event Event, data interface{}) (StateID, bool) {
	m.receivedEvent = event
	m.receivedData = data
	return m.onEventResponse, m.onEventHandled
}

// Reset the mock state for reuse
func (m *MockState) Reset() {
	m.onEnterCalled = false
	m.onExitCalled = false
	m.receivedData = nil
	m.receivedEvent = ""
}

func TestStateMachine_New(t *testing.T) {
	sm := NewStateMachine()

	if sm.states == nil {
		t.Error("Expected states map to be initialized, got nil")
	}

	if len(sm.states) != 0 {
		t.Errorf("Expected empty states map, got map with %d items", len(sm.states))
	}

	if sm.currentState != nil {
		t.Error("Expected currentState to be nil")
	}

	if sm.initialState != "" {
		t.Errorf("Expected empty initialState, got %s", sm.initialState)
	}
}

func TestStateMachine_AddState(t *testing.T) {
	sm := NewStateMachine()
	state := NewMockState("test_state")

	// Test adding valid state
	err := sm.AddState(state)
	if err != nil {
		t.Errorf("Expected no error when adding state, got %v", err)
	}

	if len(sm.states) != 1 {
		t.Errorf("Expected 1 state in machine, got %d", len(sm.states))
	}

	if !sm.HasState("test_state") {
		t.Error("Expected HasState to return true for added state")
	}

	// Test adding nil state
	err = sm.AddState(nil)
	if err == nil {
		t.Error("Expected error when adding nil state, got nil")
	}

	// Test adding duplicate state
	duplicateState := NewMockState("test_state")
	err = sm.AddState(duplicateState)
	if err == nil {
		t.Error("Expected error when adding duplicate state, got nil")
	}
}

func TestStateMachine_SetInitialState(t *testing.T) {
	sm := NewStateMachine()
	stateID := StateID("test_state")
	state := NewMockState(stateID)

	// Add the state first
	err := sm.AddState(state)
	if err != nil {
		t.Errorf("Expected no error when adding state, got %v", err)
	}

	// Set as initial state
	err = sm.SetInitialState(stateID)
	if err != nil {
		t.Errorf("Expected no error when setting initial state, got %v", err)
	}

	if sm.initialState != stateID {
		t.Errorf("Expected initial state %s, got %s", stateID, sm.initialState)
	}

	// Test setting non-existent state
	err = sm.SetInitialState("non_existent")
	if err == nil {
		t.Error("Expected error when setting non-existent initial state, got nil")
	}
}

func TestStateMachine_Initialize(t *testing.T) {
	sm := NewStateMachine()
	stateID := StateID("test_state")
	state := NewMockState(stateID)

	// Test initializing without setting initial state
	err := sm.Initialize(nil)
	if err == nil {
		t.Error("Expected error when initializing without initial state, got nil")
	}

	// Add state and set as initial
	err = sm.AddState(state)
	if err != nil {
		t.Fatalf("Failed to add state: %v", err)
	}

	err = sm.SetInitialState(stateID)
	if err != nil {
		t.Fatalf("Failed to set initial state: %v", err)
	}

	// Test successful initialization
	testData := "init_data"
	err = sm.Initialize(testData)
	if err != nil {
		t.Errorf("Expected no error when initializing, got %v", err)
	}

	if sm.currentState != state {
		t.Error("Expected current state to be set to initial state")
	}

	if !state.onEnterCalled {
		t.Error("Expected OnEnter to be called on initial state")
	}

	if state.receivedData != testData {
		t.Errorf("Expected OnEnter to receive initialization data %v, got %v", testData, state.receivedData)
	}

	// Test initialization with OnEnter error
	sm = NewStateMachine()
	errorState := NewMockState("error_state")
	expectedErr := errors.New("enter error")
	errorState.onEnterError = expectedErr

	sm.AddState(errorState)
	sm.SetInitialState("error_state")

	err = sm.Initialize(nil)
	if err == nil || err.Error() != expectedErr.Error() {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
}

func TestStateMachine_CurrentState(t *testing.T) {
	sm := NewStateMachine()

	// Test before initialization
	_, err := sm.CurrentState()
	if err == nil {
		t.Error("Expected error when getting current state before initialization")
	}

	// Initialize with a state
	state := NewMockState("test_state")
	sm.AddState(state)
	sm.SetInitialState("test_state")
	sm.Initialize(nil)

	// Test after initialization
	currentID, err := sm.CurrentState()
	if err != nil {
		t.Errorf("Expected no error when getting current state after initialization, got %v", err)
	}

	if currentID != "test_state" {
		t.Errorf("Expected current state ID %s, got %s", "test_state", currentID)
	}
}

func TestStateMachine_SendEvent(t *testing.T) {
	sm := NewStateMachine()

	// Test sending event before initialization
	err := sm.SendEvent("event", nil)
	if err == nil {
		t.Error("Expected error when sending event before initialization")
	}

	// Setup state machine with states
	stateA := NewMockState("state_a")
	stateB := NewMockState("state_b")

	sm.AddState(stateA)
	sm.AddState(stateB)
	sm.SetInitialState("state_a")
	sm.Initialize(nil)

	// Setup state A to transition to state B on 'goto_b' event
	stateA.onEventHandled = true
	stateA.onEventResponse = "state_b"

	// Test successful event handling with transition
	testEvent := Event("goto_b")
	testData := "event_data"

	err = sm.SendEvent(testEvent, testData)
	if err != nil {
		t.Errorf("Expected no error when sending handled event, got %v", err)
	}

	if stateA.receivedEvent != testEvent {
		t.Errorf("Expected event %s to be received by state A, got %s", testEvent, stateA.receivedEvent)
	}

	if stateA.receivedData != testData {
		t.Errorf("Expected data %v to be received by state A, got %v", testData, stateA.receivedData)
	}

	if !stateA.onExitCalled {
		t.Error("Expected OnExit to be called on state A during transition")
	}

	if !stateB.onEnterCalled {
		t.Error("Expected OnEnter to be called on state B during transition")
	}

	currentID, _ := sm.CurrentState()
	if currentID != "state_b" {
		t.Errorf("Expected current state to be state_b after transition, got %s", currentID)
	}

	// Test unhandled event
	stateB.onEventHandled = false
	err = sm.SendEvent("unhandled_event", nil)
	if err == nil {
		t.Error("Expected error when sending unhandled event")
	}

	// Test transition to invalid state
	stateB.Reset()
	stateB.onEventHandled = true
	stateB.onEventResponse = "non_existent"

	err = sm.SendEvent("goto_invalid", nil)
	if err == nil {
		t.Error("Expected error when transitioning to non-existent state")
	}

	// Test same-state transition (no actual transition)
	stateB.Reset()
	stateB.onEventHandled = true
	stateB.onEventResponse = "state_b" // Stay in same state

	err = sm.SendEvent("stay", nil)
	if err != nil {
		t.Errorf("Expected no error when staying in same state, got %v", err)
	}

	if stateB.onExitCalled || stateB.onEnterCalled {
		t.Error("Expected OnExit/OnEnter not to be called when staying in same state")
	}

	// Test error during OnExit
	stateB.Reset()
	stateB.onEventHandled = true
	stateB.onEventResponse = "state_a"
	stateB.onExitError = errors.New("exit error")

	err = sm.SendEvent("exit_error", nil)
	if err == nil {
		t.Error("Expected error during OnExit to be propagated")
	}

}

func TestStateMachine_HasState(t *testing.T) {
	sm := NewStateMachine()

	if sm.HasState("any_state") {
		t.Error("HasState should return false for empty state machine")
	}

	sm.AddState(NewMockState("test_state"))

	if !sm.HasState("test_state") {
		t.Error("HasState should return true for added state")
	}

	if sm.HasState("non_existent") {
		t.Error("HasState should return false for non-existent state")
	}
}

func TestStateMachine_ConcurrentAccess(t *testing.T) {
	sm := NewStateMachine()
	state := NewMockState("test_state")

	sm.AddState(state)
	sm.SetInitialState("test_state")
	sm.Initialize(nil)

	var wg sync.WaitGroup
	operationCount := 100
	wg.Add(operationCount)

	// Access HasState and CurrentState concurrently
	for i := 0; i < operationCount; i++ {
		go func() {
			defer wg.Done()
			sm.HasState("test_state")
			sm.CurrentState()
		}()
	}

	wg.Wait()

	// Final state check
	currentID, err := sm.CurrentState()
	if err != nil || currentID != "test_state" {
		t.Errorf("State machine in unexpected state after concurrent access: %v", err)
	}
}
