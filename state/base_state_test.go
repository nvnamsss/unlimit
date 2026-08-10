package state

import (
	"testing"
)

func TestBaseState_New(t *testing.T) {
	// Create new instance
	stateID := StateID("test_state")
	bs := NewBaseState(stateID)

	// Verify initial state
	if bs.ID() != stateID {
		t.Errorf("Expected state ID %s, got %s", stateID, bs.ID())
	}

	if bs.handlers == nil {
		t.Error("Expected handlers map to be initialized, got nil")
	}

	if len(bs.handlers) != 0 {
		t.Errorf("Expected empty handlers map, got map with %d items", len(bs.handlers))
	}
}

func TestBaseState_ID(t *testing.T) {
	// Setup instance
	stateID := StateID("test_state")
	bs := NewBaseState(stateID)

	// Verify ID method
	if bs.ID() != stateID {
		t.Errorf("Expected state ID %s, got %s", stateID, bs.ID())
	}
}

func TestBaseState_OnEnterExit(t *testing.T) {
	// Setup instance
	bs := NewBaseState("test_state")

	// Test default implementations
	if err := bs.OnEnter(nil); err != nil {
		t.Errorf("Expected nil error from default OnEnter, got %v", err)
	}

	if err := bs.OnExit(); err != nil {
		t.Errorf("Expected nil error from default OnExit, got %v", err)
	}
}

func TestBaseState_AddEventHandler(t *testing.T) {
	// Setup instance
	bs := NewBaseState("test_state")
	testEvent := Event("test_event")
	targetStateID := StateID("target_state")

	// Add handler
	handlerCalled := false
	bs.AddEventHandler(testEvent, func(data interface{}) StateID {
		handlerCalled = true
		return targetStateID
	})

	// Verify handler was added
	if len(bs.handlers) != 1 {
		t.Errorf("Expected 1 handler, got %d", len(bs.handlers))
	}

	// Verify handler works
	nextStateID, handled := bs.OnEvent(testEvent, nil)
	if !handled {
		t.Error("Expected event to be handled, but it was not")
	}

	if nextStateID != targetStateID {
		t.Errorf("Expected next state ID %s, got %s", targetStateID, nextStateID)
	}

	if !handlerCalled {
		t.Error("Expected handler to be called, but it was not")
	}
}

func TestBaseState_OnEvent_Unhandled(t *testing.T) {
	// Setup instance
	stateID := StateID("test_state")
	bs := NewBaseState(stateID)
	testEvent := Event("unhandled_event")

	// Verify unhandled event behavior
	nextStateID, handled := bs.OnEvent(testEvent, nil)
	if handled {
		t.Error("Expected event to be unhandled, but it was handled")
	}

	if nextStateID != stateID {
		t.Errorf("Expected unhandled event to return current state ID %s, got %s", stateID, nextStateID)
	}
}

func TestBaseState_OnEvent_WithData(t *testing.T) {
	// Setup instance
	bs := NewBaseState("test_state")
	testEvent := Event("data_event")
	testData := "test_data"
	var receivedData interface{}

	// Add handler that captures data
	bs.AddEventHandler(testEvent, func(data interface{}) StateID {
		receivedData = data
		return "next_state"
	})

	// Verify data is passed to handler
	bs.OnEvent(testEvent, testData)
	if receivedData != testData {
		t.Errorf("Expected handler to receive data %v, got %v", testData, receivedData)
	}
}

func TestBaseState_MultipleHandlers(t *testing.T) {
	// Setup instance
	bs := NewBaseState("test_state")
	event1 := Event("event1")
	event2 := Event("event2")
	state1 := StateID("state1")
	state2 := StateID("state2")

	// Add multiple handlers
	bs.AddEventHandler(event1, func(data interface{}) StateID {
		return state1
	})
	bs.AddEventHandler(event2, func(data interface{}) StateID {
		return state2
	})

	// Verify handlers
	if len(bs.handlers) != 2 {
		t.Errorf("Expected 2 handlers, got %d", len(bs.handlers))
	}

	nextState1, handled1 := bs.OnEvent(event1, nil)
	if !handled1 || nextState1 != state1 {
		t.Errorf("Expected event1 to return state1, got %s", nextState1)
	}

	nextState2, handled2 := bs.OnEvent(event2, nil)
	if !handled2 || nextState2 != state2 {
		t.Errorf("Expected event2 to return state2, got %s", nextState2)
	}
}

func TestBaseState_OverrideHandler(t *testing.T) {
	// Setup instance
	bs := NewBaseState("test_state")
	testEvent := Event("test_event")
	state1 := StateID("state1")
	state2 := StateID("state2")

	// Add handler
	bs.AddEventHandler(testEvent, func(data interface{}) StateID {
		return state1
	})

	// Override handler
	bs.AddEventHandler(testEvent, func(data interface{}) StateID {
		return state2
	})

	// Verify handler was overridden
	nextStateID, _ := bs.OnEvent(testEvent, nil)
	if nextStateID != state2 {
		t.Errorf("Expected overridden handler to return state2, got %s", nextStateID)
	}
}
