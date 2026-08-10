package trigger

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestTriggerBasicFunctionality(t *testing.T) {
	// Create a trigger manager
	manager := NewTriggerManager()

	// Start the manager
	if err := manager.Start(); err != nil {
		t.Fatalf("Failed to start trigger manager: %v", err)
	}
	defer manager.Stop()

	// Create a simple trigger
	var executed bool
	var executedMutex sync.Mutex

	trigger := NewTrigger("test-trigger", "Test Trigger").
		AddEventType(EventTypeCustom).
		AddAction(func(event Event) error {
			executedMutex.Lock()
			executed = true
			executedMutex.Unlock()
			return nil
		})

	// Register the trigger
	if err := manager.RegisterTrigger(trigger); err != nil {
		t.Fatalf("Failed to register trigger: %v", err)
	}

	// Fire an event
	event := NewCustomEvent("test", map[string]interface{}{"value": 42}, "test-source")
	if err := manager.FireEvent(event); err != nil {
		t.Fatalf("Failed to fire event: %v", err)
	}

	// Wait a bit for async processing
	time.Sleep(100 * time.Millisecond)

	// Check if the trigger was executed
	executedMutex.Lock()
	if !executed {
		t.Error("Trigger was not executed")
	}
	executedMutex.Unlock()
}

func TestTriggerWithConditions(t *testing.T) {
	manager := NewTriggerManager()
	manager.Start()
	defer manager.Stop()

	var executionCount int

	// Create a trigger with conditions
	trigger := NewTrigger("conditional-trigger", "Conditional Trigger").
		AddEventType(EventTypeUser).
		AddCondition(func(event Event) bool {
			if userEvent, ok := event.(*UserEvent); ok {
				return userEvent.Action == "login"
			}
			return false
		}).
		AddAction(func(event Event) error {
			executionCount++
			return nil
		})

	manager.RegisterTrigger(trigger)

	// Fire events - only one should match the condition
	loginEvent := NewUserEvent("user123", "login", "system", "auth-service")
	logoutEvent := NewUserEvent("user123", "logout", "system", "auth-service")

	manager.FireEvent(loginEvent)
	manager.FireEvent(logoutEvent)

	time.Sleep(100 * time.Millisecond)

	if executionCount != 1 {
		t.Errorf("Expected 1 execution, got %d", executionCount)
	}
}

func TestTriggerPriority(t *testing.T) {
	manager := NewTriggerManager()

	var executionOrder []string
	var mutex sync.Mutex

	// Create triggers with different priorities
	highPriorityTrigger := NewTrigger("high-priority", "High Priority").
		AddEventType(EventTypeSystem).
		SetPriority(PriorityHigh).
		AddAction(func(event Event) error {
			mutex.Lock()
			executionOrder = append(executionOrder, "high")
			mutex.Unlock()
			return nil
		})

	lowPriorityTrigger := NewTrigger("low-priority", "Low Priority").
		AddEventType(EventTypeSystem).
		SetPriority(PriorityLow).
		AddAction(func(event Event) error {
			mutex.Lock()
			executionOrder = append(executionOrder, "low")
			mutex.Unlock()
			return nil
		})

	normalPriorityTrigger := NewTrigger("normal-priority", "Normal Priority").
		AddEventType(EventTypeSystem).
		SetPriority(PriorityNormal).
		AddAction(func(event Event) error {
			mutex.Lock()
			executionOrder = append(executionOrder, "normal")
			mutex.Unlock()
			return nil
		})

	manager.RegisterTrigger(lowPriorityTrigger)
	manager.RegisterTrigger(highPriorityTrigger)
	manager.RegisterTrigger(normalPriorityTrigger)

	// Fire an event synchronously to test priority order
	event := NewSystemEvent("info", "test message", 100, "test")
	manager.FireEvent(event)

	mutex.Lock()
	defer mutex.Unlock()

	expectedOrder := []string{"high", "normal", "low"}
	if len(executionOrder) != len(expectedOrder) {
		t.Errorf("Expected %d executions, got %d", len(expectedOrder), len(executionOrder))
		return
	}

	for i, expected := range expectedOrder {
		if executionOrder[i] != expected {
			t.Errorf("Expected execution order %v, got %v", expectedOrder, executionOrder)
			break
		}
	}
}

func TestTriggerExecuteOnce(t *testing.T) {
	manager := NewTriggerManager()
	manager.Start()
	defer manager.Stop()

	var executionCount int
	var mutex sync.Mutex

	trigger := NewTrigger("once-trigger", "Execute Once Trigger").
		AddEventType(EventTypeTimer).
		SetExecuteOnce(true).
		AddAction(func(event Event) error {
			mutex.Lock()
			executionCount++
			mutex.Unlock()
			return nil
		})

	manager.RegisterTrigger(trigger)

	// Fire the same event multiple times
	for i := 0; i < 3; i++ {
		event := NewTimerEvent(time.Second, false, "test")
		manager.FireEvent(event)
	}

	time.Sleep(100 * time.Millisecond)

	mutex.Lock()
	if executionCount != 1 {
		t.Errorf("Expected 1 execution, got %d", executionCount)
	}
	mutex.Unlock()
}

func TestTimerEvents(t *testing.T) {
	manager := NewTriggerManager()
	manager.Start()
	defer manager.Stop()

	var timerFired bool
	var mutex sync.Mutex

	trigger := NewTrigger("timer-trigger", "Timer Trigger").
		AddEventType(EventTypeTimer).
		AddAction(func(event Event) error {
			mutex.Lock()
			timerFired = true
			mutex.Unlock()
			return nil
		})

	manager.RegisterTrigger(trigger)

	// Create a timer event that fires after 50ms
	manager.CreateTimerEvent(EventTypeTimer, 50*time.Millisecond, nil, "timer-test")

	// Wait for the timer to fire
	time.Sleep(100 * time.Millisecond)

	mutex.Lock()
	if !timerFired {
		t.Error("Timer event did not fire")
	}
	mutex.Unlock()
}

func TestEventListeners(t *testing.T) {
	manager := NewTriggerManager()
	manager.Start()
	defer manager.Stop()

	var listenerCalled bool
	var mutex sync.Mutex

	// Add an event listener
	manager.AddEventListener(EventTypeCustom, func(event Event) error {
		mutex.Lock()
		listenerCalled = true
		mutex.Unlock()
		return nil
	})

	// Fire an event
	event := NewCustomEvent("test", nil, "test")
	manager.FireEvent(event)

	time.Sleep(100 * time.Millisecond)

	mutex.Lock()
	if !listenerCalled {
		t.Error("Event listener was not called")
	}
	mutex.Unlock()
}

func TestTriggerManagerStats(t *testing.T) {
	manager := NewTriggerManager()

	// Register some triggers
	trigger1 := NewTrigger("trigger1", "Trigger 1").SetEnabled(true)
	trigger2 := NewTrigger("trigger2", "Trigger 2").SetEnabled(false)
	trigger3 := NewTrigger("trigger3", "Trigger 3").SetEnabled(true)

	manager.RegisterTrigger(trigger1)
	manager.RegisterTrigger(trigger2)
	manager.RegisterTrigger(trigger3)

	stats := manager.Stats()

	if stats["total_triggers"] != 3 {
		t.Errorf("Expected 3 total triggers, got %v", stats["total_triggers"])
	}

	if stats["enabled_triggers"] != 2 {
		t.Errorf("Expected 2 enabled triggers, got %v", stats["enabled_triggers"])
	}

	if stats["is_running"] != false {
		t.Errorf("Expected is_running to be false, got %v", stats["is_running"])
	}
}

func TestComplexWorkflowExample(t *testing.T) {
	manager := NewTriggerManager()
	manager.Start()
	defer manager.Stop()

	var results []string
	var mutex sync.Mutex

	// Create a workflow: user registration -> send welcome email -> update metrics
	registrationTrigger := NewTrigger("user-registration", "User Registration Handler").
		AddEventType(EventTypeUser).
		AddCondition(func(event Event) bool {
			if userEvent, ok := event.(*UserEvent); ok {
				return userEvent.Action == "register"
			}
			return false
		}).
		AddAction(func(event Event) error {
			mutex.Lock()
			results = append(results, "user_registered")
			mutex.Unlock()

			// Trigger welcome email event
			welcomeEvent := NewCustomEvent("welcome_email", map[string]interface{}{
				"user_id": event.(*UserEvent).UserID,
			}, "registration-service")
			return manager.FireEvent(welcomeEvent)
		})

	emailTrigger := NewTrigger("welcome-email", "Welcome Email Sender").
		AddEventType(EventTypeCustom).
		AddCondition(func(event Event) bool {
			if customEvent, ok := event.(*CustomEvent); ok {
				return customEvent.Category == "welcome_email"
			}
			return false
		}).
		AddAction(func(event Event) error {
			mutex.Lock()
			results = append(results, "email_sent")
			mutex.Unlock()

			// Trigger metrics update
			metricsEvent := NewSystemEvent("info", "user_registered", 200, "email-service")
			return manager.FireEvent(metricsEvent)
		})

	metricsTrigger := NewTrigger("metrics-update", "Metrics Updater").
		AddEventType(EventTypeSystem).
		AddCondition(func(event Event) bool {
			if systemEvent, ok := event.(*SystemEvent); ok {
				return systemEvent.Message == "user_registered"
			}
			return false
		}).
		AddAction(func(event Event) error {
			mutex.Lock()
			results = append(results, "metrics_updated")
			mutex.Unlock()
			return nil
		})

	// Register all triggers
	manager.RegisterTrigger(registrationTrigger)
	manager.RegisterTrigger(emailTrigger)
	manager.RegisterTrigger(metricsTrigger)

	// Start the workflow with a user registration event
	registrationEvent := NewUserEvent("user123", "register", "system", "web-app")
	manager.FireEvent(registrationEvent)

	// Wait for all async operations to complete
	time.Sleep(200 * time.Millisecond)

	mutex.Lock()
	defer mutex.Unlock()

	expectedResults := []string{"user_registered", "email_sent", "metrics_updated"}
	if len(results) != len(expectedResults) {
		t.Errorf("Expected %d results, got %d: %v", len(expectedResults), len(results), results)
		return
	}

	for i, expected := range expectedResults {
		if results[i] != expected {
			t.Errorf("Expected result sequence %v, got %v", expectedResults, results)
			break
		}
	}
}

// Unit tests for Trigger struct

// TestTrigger_New tests trigger creation
func TestTrigger_New(t *testing.T) {
	trigger := NewTrigger("test-id", "Test Trigger")

	if trigger.ID != "test-id" {
		t.Errorf("Expected ID 'test-id', got '%s'", trigger.ID)
	}

	if trigger.Name != "Test Trigger" {
		t.Errorf("Expected Name 'Test Trigger', got '%s'", trigger.Name)
	}

	if trigger.Priority != PriorityNormal {
		t.Errorf("Expected default priority PriorityNormal, got %v", trigger.Priority)
	}

	if !trigger.Enabled {
		t.Error("Expected trigger to be enabled by default")
	}

	if trigger.ExecuteOnce {
		t.Error("Expected ExecuteOnce to be false by default")
	}

	if trigger.executed {
		t.Error("Expected executed to be false initially")
	}

	if len(trigger.EventTypes) != 0 {
		t.Errorf("Expected empty EventTypes, got %d items", len(trigger.EventTypes))
	}

	if len(trigger.Conditions) != 0 {
		t.Errorf("Expected empty Conditions, got %d items", len(trigger.Conditions))
	}

	if len(trigger.Actions) != 0 {
		t.Errorf("Expected empty Actions, got %d items", len(trigger.Actions))
	}
}

// TestTrigger_AddEventType tests adding event types
func TestTrigger_AddEventType(t *testing.T) {
	trigger := NewTrigger("test", "Test")

	trigger.AddEventType(EventTypeUser)
	trigger.AddEventType(EventTypeSystem)

	if len(trigger.EventTypes) != 2 {
		t.Errorf("Expected 2 event types, got %d", len(trigger.EventTypes))
	}

	if trigger.EventTypes[0] != EventTypeUser {
		t.Errorf("Expected first event type to be %s, got %s", EventTypeUser, trigger.EventTypes[0])
	}

	if trigger.EventTypes[1] != EventTypeSystem {
		t.Errorf("Expected second event type to be %s, got %s", EventTypeSystem, trigger.EventTypes[1])
	}
}

// TestTrigger_AddCondition tests adding conditions
func TestTrigger_AddCondition(t *testing.T) {
	trigger := NewTrigger("test", "Test")

	condition1 := func(event Event) bool { return true }
	condition2 := func(event Event) bool { return false }

	trigger.AddCondition(condition1)
	trigger.AddCondition(condition2)

	if len(trigger.Conditions) != 2 {
		t.Errorf("Expected 2 conditions, got %d", len(trigger.Conditions))
	}
}

// TestTrigger_AddAction tests adding actions
func TestTrigger_AddAction(t *testing.T) {
	trigger := NewTrigger("test", "Test")

	action1 := func(event Event) error { return nil }
	action2 := func(event Event) error { return nil }

	trigger.AddAction(action1)
	trigger.AddAction(action2)

	if len(trigger.Actions) != 2 {
		t.Errorf("Expected 2 actions, got %d", len(trigger.Actions))
	}
}

// TestTrigger_SetPriority tests setting priority
func TestTrigger_SetPriority(t *testing.T) {
	trigger := NewTrigger("test", "Test")

	trigger.SetPriority(PriorityHigh)
	if trigger.Priority != PriorityHigh {
		t.Errorf("Expected priority PriorityHigh, got %v", trigger.Priority)
	}

	trigger.SetPriority(PriorityLow)
	if trigger.Priority != PriorityLow {
		t.Errorf("Expected priority PriorityLow, got %v", trigger.Priority)
	}

	trigger.SetPriority(PriorityCritical)
	if trigger.Priority != PriorityCritical {
		t.Errorf("Expected priority PriorityCritical, got %v", trigger.Priority)
	}
}

// TestTrigger_SetEnabled tests enabling/disabling trigger
func TestTrigger_SetEnabled(t *testing.T) {
	trigger := NewTrigger("test", "Test")

	if !trigger.Enabled {
		t.Error("Expected trigger to be enabled by default")
	}

	trigger.SetEnabled(false)
	if trigger.Enabled {
		t.Error("Expected trigger to be disabled")
	}

	trigger.SetEnabled(true)
	if !trigger.Enabled {
		t.Error("Expected trigger to be enabled")
	}
}

// TestTrigger_SetExecuteOnce tests execute once flag
func TestTrigger_SetExecuteOnce(t *testing.T) {
	trigger := NewTrigger("test", "Test")

	if trigger.ExecuteOnce {
		t.Error("Expected ExecuteOnce to be false by default")
	}

	trigger.SetExecuteOnce(true)
	if !trigger.ExecuteOnce {
		t.Error("Expected ExecuteOnce to be true")
	}

	trigger.SetExecuteOnce(false)
	if trigger.ExecuteOnce {
		t.Error("Expected ExecuteOnce to be false")
	}
}

// TestTrigger_SetDescription tests setting description
func TestTrigger_SetDescription(t *testing.T) {
	trigger := NewTrigger("test", "Test")

	description := "This is a test trigger"
	trigger.SetDescription(description)

	if trigger.Description != description {
		t.Errorf("Expected description '%s', got '%s'", description, trigger.Description)
	}
}

// TestTrigger_CanExecute tests execution eligibility
func TestTrigger_CanExecute(t *testing.T) {
	t.Run("DisabledTrigger", func(t *testing.T) {
		trigger := NewTrigger("test", "Test").
			AddEventType(EventTypeUser).
			SetEnabled(false)

		event := NewUserEvent("user1", "login", "system", "test")

		if trigger.CanExecute(event) {
			t.Error("Disabled trigger should not be executable")
		}
	})

	t.Run("WrongEventType", func(t *testing.T) {
		trigger := NewTrigger("test", "Test").
			AddEventType(EventTypeUser)

		event := NewSystemEvent("info", "test", 100, "test")

		if trigger.CanExecute(event) {
			t.Error("Trigger should not execute for wrong event type")
		}
	})

	t.Run("MatchingEventType", func(t *testing.T) {
		trigger := NewTrigger("test", "Test").
			AddEventType(EventTypeUser)

		event := NewUserEvent("user1", "login", "system", "test")

		if !trigger.CanExecute(event) {
			t.Error("Trigger should execute for matching event type")
		}
	})

	t.Run("FailedCondition", func(t *testing.T) {
		trigger := NewTrigger("test", "Test").
			AddEventType(EventTypeUser).
			AddCondition(func(event Event) bool {
				return false
			})

		event := NewUserEvent("user1", "login", "system", "test")

		if trigger.CanExecute(event) {
			t.Error("Trigger should not execute when condition fails")
		}
	})

	t.Run("MultipleConditionsAllPass", func(t *testing.T) {
		trigger := NewTrigger("test", "Test").
			AddEventType(EventTypeUser).
			AddCondition(func(event Event) bool { return true }).
			AddCondition(func(event Event) bool { return true })

		event := NewUserEvent("user1", "login", "system", "test")

		if !trigger.CanExecute(event) {
			t.Error("Trigger should execute when all conditions pass")
		}
	})

	t.Run("MultipleConditionsOneFails", func(t *testing.T) {
		trigger := NewTrigger("test", "Test").
			AddEventType(EventTypeUser).
			AddCondition(func(event Event) bool { return true }).
			AddCondition(func(event Event) bool { return false })

		event := NewUserEvent("user1", "login", "system", "test")

		if trigger.CanExecute(event) {
			t.Error("Trigger should not execute when any condition fails")
		}
	})

	t.Run("ExecuteOnceAlreadyExecuted", func(t *testing.T) {
		trigger := NewTrigger("test", "Test").
			AddEventType(EventTypeUser).
			SetExecuteOnce(true)

		// Mark as executed
		trigger.executed = true

		event := NewUserEvent("user1", "login", "system", "test")

		if trigger.CanExecute(event) {
			t.Error("ExecuteOnce trigger should not execute after first execution")
		}
	})
}

// TestTrigger_Execute tests trigger execution
func TestTrigger_Execute(t *testing.T) {
	t.Run("BasicExecution", func(t *testing.T) {
		executed := false
		trigger := NewTrigger("test", "Test").
			AddEventType(EventTypeUser).
			AddAction(func(event Event) error {
				executed = true
				return nil
			})

		event := NewUserEvent("user1", "login", "system", "test")

		err := trigger.Execute(event)
		if err != nil {
			t.Errorf("Execute failed: %v", err)
		}

		if !executed {
			t.Error("Action was not executed")
		}
	})

	t.Run("MultipleActionsExecution", func(t *testing.T) {
		executionOrder := []int{}
		trigger := NewTrigger("test", "Test").
			AddEventType(EventTypeUser).
			AddAction(func(event Event) error {
				executionOrder = append(executionOrder, 1)
				return nil
			}).
			AddAction(func(event Event) error {
				executionOrder = append(executionOrder, 2)
				return nil
			}).
			AddAction(func(event Event) error {
				executionOrder = append(executionOrder, 3)
				return nil
			})

		event := NewUserEvent("user1", "login", "system", "test")

		err := trigger.Execute(event)
		if err != nil {
			t.Errorf("Execute failed: %v", err)
		}

		expected := []int{1, 2, 3}
		if len(executionOrder) != len(expected) {
			t.Errorf("Expected %d actions, got %d", len(expected), len(executionOrder))
		}

		for i, v := range expected {
			if executionOrder[i] != v {
				t.Errorf("Expected execution order %v, got %v", expected, executionOrder)
				break
			}
		}
	})

	t.Run("ExecutionMarksExecuteOnce", func(t *testing.T) {
		trigger := NewTrigger("test", "Test").
			AddEventType(EventTypeUser).
			SetExecuteOnce(true).
			AddAction(func(event Event) error {
				return nil
			})

		event := NewUserEvent("user1", "login", "system", "test")

		if trigger.IsExecuted() {
			t.Error("Trigger should not be marked as executed before execution")
		}

		err := trigger.Execute(event)
		if err != nil {
			t.Errorf("Execute failed: %v", err)
		}

		if !trigger.IsExecuted() {
			t.Error("ExecuteOnce trigger should be marked as executed after execution")
		}
	})

	t.Run("ExecutionReceivesEvent", func(t *testing.T) {
		var receivedEvent Event
		trigger := NewTrigger("test", "Test").
			AddEventType(EventTypeUser).
			AddAction(func(event Event) error {
				receivedEvent = event
				return nil
			})

		event := NewUserEvent("user123", "login", "web", "auth")

		err := trigger.Execute(event)
		if err != nil {
			t.Errorf("Execute failed: %v", err)
		}

		if receivedEvent == nil {
			t.Fatal("Event was not received by action")
		}

		userEvent, ok := receivedEvent.(*UserEvent)
		if !ok {
			t.Fatal("Received event is not UserEvent")
		}

		if userEvent.UserID != "user123" {
			t.Errorf("Expected UserID 'user123', got '%s'", userEvent.UserID)
		}
	})
}

// TestTrigger_ExecuteError tests error handling during execution
func TestTrigger_ExecuteError(t *testing.T) {
	t.Run("ActionReturnsError", func(t *testing.T) {
		trigger := NewTrigger("test", "Test").
			AddEventType(EventTypeUser).
			AddAction(func(event Event) error {
				return fmt.Errorf("action failed")
			})

		event := NewUserEvent("user1", "login", "system", "test")

		err := trigger.Execute(event)
		if err == nil {
			t.Error("Expected error from failed action")
		}
	})

	t.Run("FirstActionFailsStopsExecution", func(t *testing.T) {
		firstExecuted := false
		secondExecuted := false

		trigger := NewTrigger("test", "Test").
			AddEventType(EventTypeUser).
			AddAction(func(event Event) error {
				firstExecuted = true
				return fmt.Errorf("first action failed")
			}).
			AddAction(func(event Event) error {
				secondExecuted = true
				return nil
			})

		event := NewUserEvent("user1", "login", "system", "test")

		err := trigger.Execute(event)
		if err == nil {
			t.Error("Expected error from failed action")
		}

		if !firstExecuted {
			t.Error("First action should have been executed")
		}

		if secondExecuted {
			t.Error("Second action should not have been executed after first failed")
		}
	})
}

// TestTrigger_IsExecuted tests execution state checking
func TestTrigger_IsExecuted(t *testing.T) {
	trigger := NewTrigger("test", "Test").
		AddEventType(EventTypeUser).
		SetExecuteOnce(true).
		AddAction(func(event Event) error {
			return nil
		})

	if trigger.IsExecuted() {
		t.Error("Trigger should not be executed initially")
	}

	event := NewUserEvent("user1", "login", "system", "test")
	trigger.Execute(event)

	if !trigger.IsExecuted() {
		t.Error("Trigger should be marked as executed")
	}
}

// TestTrigger_Reset tests resetting execution state
func TestTrigger_Reset(t *testing.T) {
	trigger := NewTrigger("test", "Test").
		AddEventType(EventTypeUser).
		SetExecuteOnce(true).
		AddAction(func(event Event) error {
			return nil
		})

	event := NewUserEvent("user1", "login", "system", "test")
	trigger.Execute(event)

	if !trigger.IsExecuted() {
		t.Error("Trigger should be executed")
	}

	trigger.Reset()

	if trigger.IsExecuted() {
		t.Error("Trigger should not be executed after reset")
	}

	// Should be able to execute again after reset
	if !trigger.CanExecute(event) {
		t.Error("Trigger should be executable after reset")
	}
}

// TestTrigger_ConcurrentAccess tests thread safety
func TestTrigger_ConcurrentAccess(t *testing.T) {
	trigger := NewTrigger("test", "Test").
		AddEventType(EventTypeUser)

	var wg sync.WaitGroup
	operationCount := 100

	// Concurrent writes
	wg.Add(operationCount)
	for i := 0; i < operationCount; i++ {
		go func(val int) {
			defer wg.Done()
			trigger.AddAction(func(event Event) error {
				return nil
			})
		}(i)
	}
	wg.Wait()

	// Verify all actions were added
	if len(trigger.Actions) != operationCount {
		t.Errorf("Expected %d actions, got %d", operationCount, len(trigger.Actions))
	}

	// Concurrent reads
	wg.Add(operationCount)
	for i := 0; i < operationCount; i++ {
		go func() {
			defer wg.Done()
			trigger.IsExecuted()
			_ = trigger.Enabled
		}()
	}
	wg.Wait()
}

// TestTrigger_ChainedMethods tests method chaining
func TestTrigger_ChainedMethods(t *testing.T) {
	trigger := NewTrigger("test", "Test").
		AddEventType(EventTypeUser).
		AddEventType(EventTypeSystem).
		SetPriority(PriorityHigh).
		SetEnabled(true).
		SetExecuteOnce(true).
		SetDescription("Test Description").
		AddCondition(func(event Event) bool { return true }).
		AddAction(func(event Event) error { return nil })

	if len(trigger.EventTypes) != 2 {
		t.Errorf("Expected 2 event types, got %d", len(trigger.EventTypes))
	}

	if trigger.Priority != PriorityHigh {
		t.Errorf("Expected PriorityHigh, got %v", trigger.Priority)
	}

	if !trigger.Enabled {
		t.Error("Expected trigger to be enabled")
	}

	if !trigger.ExecuteOnce {
		t.Error("Expected ExecuteOnce to be true")
	}

	if trigger.Description != "Test Description" {
		t.Errorf("Expected description 'Test Description', got '%s'", trigger.Description)
	}

	if len(trigger.Conditions) != 1 {
		t.Errorf("Expected 1 condition, got %d", len(trigger.Conditions))
	}

	if len(trigger.Actions) != 1 {
		t.Errorf("Expected 1 action, got %d", len(trigger.Actions))
	}
}

// TestTrigger_EmptyActions tests trigger with no actions
func TestTrigger_EmptyActions(t *testing.T) {
	trigger := NewTrigger("test", "Test").
		AddEventType(EventTypeUser)

	event := NewUserEvent("user1", "login", "system", "test")

	err := trigger.Execute(event)
	if err != nil {
		t.Errorf("Execute with no actions should succeed: %v", err)
	}
}

// TestTrigger_DifferentEventTypes tests trigger with multiple event types
func TestTrigger_DifferentEventTypes(t *testing.T) {
	executed := false
	trigger := NewTrigger("test", "Test").
		AddEventType(EventTypeUser).
		AddEventType(EventTypeSystem).
		AddEventType(EventTypeCustom).
		AddAction(func(event Event) error {
			executed = true
			return nil
		})

	tests := []struct {
		name  string
		event Event
		match bool
	}{
		{"UserEvent", NewUserEvent("u1", "login", "sys", "test"), true},
		{"SystemEvent", NewSystemEvent("info", "msg", 100, "test"), true},
		{"CustomEvent", NewCustomEvent("cat", nil, "test"), true},
		{"TimerEvent", NewTimerEvent(time.Second, false, "test"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executed = false
			canExec := trigger.CanExecute(tt.event)

			if canExec != tt.match {
				t.Errorf("CanExecute = %v, want %v", canExec, tt.match)
			}

			if tt.match {
				err := trigger.Execute(tt.event)
				if err != nil {
					t.Errorf("Execute failed: %v", err)
				}
				if !executed {
					t.Error("Action was not executed")
				}
			}
		})
	}
}
