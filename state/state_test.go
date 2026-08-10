package state

import (
	"fmt"
	"testing"
)

// Custom state implementation
type DoorState struct {
	*BaseState
	onEnterFunc func(data interface{}) error
	onExitFunc  func() error
}

func NewDoorState(id StateID, onEnter func(data interface{}) error, onExit func() error) *DoorState {
	return &DoorState{
		BaseState:   NewBaseState(id),
		onEnterFunc: onEnter,
		onExitFunc:  onExit,
	}
}

func (d *DoorState) OnEnter(data interface{}) error {
	if d.onEnterFunc != nil {
		return d.onEnterFunc(data)
	}
	return nil
}

func (d *DoorState) OnExit() error {
	if d.onExitFunc != nil {
		return d.onExitFunc()
	}
	return nil
}

func Example_doorStateMachine() {
	// Create a state machine
	doorMachine := NewStateMachine()

	// Define events
	openEvent := Event("open")
	closeEvent := Event("close")
	lockEvent := Event("lock")
	unlockEvent := Event("unlock")

	// Define states
	closedState := NewDoorState(
		"closed",
		func(data interface{}) error {
			fmt.Println("Door is now closed")
			return nil
		},
		nil,
	)

	openedState := NewDoorState(
		"opened",
		func(data interface{}) error {
			fmt.Println("Door is now open")
			return nil
		},
		nil,
	)

	lockedState := NewDoorState(
		"locked",
		func(data interface{}) error {
			fmt.Println("Door is now locked")
			return nil
		},
		nil,
	)

	// Add event handlers
	closedState.AddEventHandler(openEvent, func(data interface{}) StateID {
		return openedState.ID()
	})
	closedState.AddEventHandler(lockEvent, func(data interface{}) StateID {
		return lockedState.ID()
	})

	openedState.AddEventHandler(closeEvent, func(data interface{}) StateID {
		return closedState.ID()
	})

	lockedState.AddEventHandler(unlockEvent, func(data interface{}) StateID {
		return closedState.ID()
	})

	// Add states to the machine
	doorMachine.AddState(closedState)
	doorMachine.AddState(openedState)
	doorMachine.AddState(lockedState)

	// Set initial state
	doorMachine.SetInitialState("closed")

	// Initialize state machine
	doorMachine.Initialize(nil)

	// Perform state transitions
	doorMachine.SendEvent(openEvent, nil)
	doorMachine.SendEvent(closeEvent, nil)
	doorMachine.SendEvent(lockEvent, nil)
	doorMachine.SendEvent(unlockEvent, nil)

	// Output:
	// Door is now closed
	// Door is now open
	// Door is now closed
	// Door is now locked
	// Door is now closed
}

func TestDoorStateMachine(t *testing.T) {
	// This is just a wrapper to run the example as a test
	Example_doorStateMachine()
}
