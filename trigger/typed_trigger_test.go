package trigger

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// Tests for TypedEvent

func TestTypedEvent_New(t *testing.T) {
	type TestData struct {
		Name  string
		Value int
	}

	data := TestData{Name: "test", Value: 42}
	event := NewTypedEvent("test-type", data, "test-source")

	if event.Type() != "test-type" {
		t.Errorf("Expected type 'test-type', got '%s'", event.Type())
	}

	if event.Source() != "test-source" {
		t.Errorf("Expected source 'test-source', got '%s'", event.Source())
	}

	typedData := event.TypedData()
	if typedData.Name != "test" {
		t.Errorf("Expected Name 'test', got '%s'", typedData.Name)
	}

	if typedData.Value != 42 {
		t.Errorf("Expected Value 42, got %d", typedData.Value)
	}
}

func TestTypedEvent_WrapEvent(t *testing.T) {
	t.Run("SuccessfulWrap", func(t *testing.T) {
		type TestData struct {
			Message string
		}

		data := TestData{Message: "hello"}
		baseEvent := NewBaseEvent("test", data, "source")

		typedEvent, ok := WrapEvent[TestData](baseEvent)
		if !ok {
			t.Error("Expected successful wrap")
		}

		if typedEvent.TypedData().Message != "hello" {
			t.Errorf("Expected Message 'hello', got '%s'", typedEvent.TypedData().Message)
		}
	})

	t.Run("FailedWrap", func(t *testing.T) {
		type TestData struct {
			Message string
		}

		// Create event with different data type
		baseEvent := NewBaseEvent("test", 123, "source")

		_, ok := WrapEvent[TestData](baseEvent)
		if ok {
			t.Error("Expected wrap to fail with wrong type")
		}
	})
}

func TestTypedEvent_TypeSafety(t *testing.T) {
	type UserData struct {
		UserID string
		Email  string
	}

	userData := UserData{UserID: "user123", Email: "test@example.com"}
	event := NewTypedEvent("user-event", userData, "test")

	// Type-safe access
	data := event.TypedData()
	if data.UserID != "user123" {
		t.Errorf("Expected UserID 'user123', got '%s'", data.UserID)
	}

	if data.Email != "test@example.com" {
		t.Errorf("Expected Email 'test@example.com', got '%s'", data.Email)
	}
}

// Tests for TypedTrigger

func TestTypedTrigger_New(t *testing.T) {
	type TestData struct {
		Value int
	}

	trigger := NewTypedTrigger[TestData]("test-id", "Test Trigger")

	if trigger.ID() != "test-id" {
		t.Errorf("Expected ID 'test-id', got '%s'", trigger.ID())
	}

	if trigger.Name() != "Test Trigger" {
		t.Errorf("Expected Name 'Test Trigger', got '%s'", trigger.Name())
	}
}

func TestTypedTrigger_AddTypedCondition(t *testing.T) {
	type TestData struct {
		Value int
	}

	trigger := NewTypedTrigger[TestData]("test", "Test").
		AddEventType("test-type").
		AddTypedCondition(func(data TestData) bool {
			return data.Value > 10
		})

	// Test with matching condition
	event1 := NewTypedEvent("test-type", TestData{Value: 20}, "test")
	if !trigger.CanExecute(event1) {
		t.Error("Trigger should execute when condition is met")
	}

	// Test with non-matching condition
	event2 := NewTypedEvent("test-type", TestData{Value: 5}, "test")
	if trigger.CanExecute(event2) {
		t.Error("Trigger should not execute when condition fails")
	}
}

func TestTypedTrigger_AddTypedAction(t *testing.T) {
	type TestData struct {
		Value int
	}

	var executedValue int
	var mutex sync.Mutex

	trigger := NewTypedTrigger[TestData]("test", "Test").
		AddEventType("test-type").
		AddTypedAction(func(data TestData) error {
			mutex.Lock()
			executedValue = data.Value
			mutex.Unlock()
			return nil
		})

	event := NewTypedEvent("test-type", TestData{Value: 42}, "test")
	err := trigger.Execute(event)
	if err != nil {
		t.Errorf("Execute failed: %v", err)
	}

	mutex.Lock()
	if executedValue != 42 {
		t.Errorf("Expected executed value 42, got %d", executedValue)
	}
	mutex.Unlock()
}

func TestTypedTrigger_MultipleTypedConditions(t *testing.T) {
	type TestData struct {
		Value int
		Name  string
	}

	trigger := NewTypedTrigger[TestData]("test", "Test").
		AddEventType("test-type").
		AddTypedCondition(func(data TestData) bool {
			return data.Value > 10
		}).
		AddTypedCondition(func(data TestData) bool {
			return data.Name != ""
		})

	// Both conditions pass
	event1 := NewTypedEvent("test-type", TestData{Value: 20, Name: "test"}, "test")
	if !trigger.CanExecute(event1) {
		t.Error("Trigger should execute when all conditions pass")
	}

	// First condition fails
	event2 := NewTypedEvent("test-type", TestData{Value: 5, Name: "test"}, "test")
	if trigger.CanExecute(event2) {
		t.Error("Trigger should not execute when first condition fails")
	}

	// Second condition fails
	event3 := NewTypedEvent("test-type", TestData{Value: 20, Name: ""}, "test")
	if trigger.CanExecute(event3) {
		t.Error("Trigger should not execute when second condition fails")
	}
}

func TestTypedTrigger_MultipleTypedActions(t *testing.T) {
	type TestData struct {
		Value int
	}

	var executionOrder []int
	var mutex sync.Mutex

	trigger := NewTypedTrigger[TestData]("test", "Test").
		AddEventType("test-type").
		AddTypedAction(func(data TestData) error {
			mutex.Lock()
			executionOrder = append(executionOrder, 1)
			mutex.Unlock()
			return nil
		}).
		AddTypedAction(func(data TestData) error {
			mutex.Lock()
			executionOrder = append(executionOrder, 2)
			mutex.Unlock()
			return nil
		})

	event := NewTypedEvent("test-type", TestData{Value: 42}, "test")
	err := trigger.Execute(event)
	if err != nil {
		t.Errorf("Execute failed: %v", err)
	}

	mutex.Lock()
	if len(executionOrder) != 2 {
		t.Errorf("Expected 2 actions executed, got %d", len(executionOrder))
	}
	if executionOrder[0] != 1 || executionOrder[1] != 2 {
		t.Errorf("Expected execution order [1, 2], got %v", executionOrder)
	}
	mutex.Unlock()
}

func TestTypedTrigger_ErrorHandling(t *testing.T) {
	type TestData struct {
		Value int
	}

	trigger := NewTypedTrigger[TestData]("test", "Test").
		AddEventType("test-type").
		AddTypedAction(func(data TestData) error {
			return fmt.Errorf("action failed")
		})

	event := NewTypedEvent("test-type", TestData{Value: 42}, "test")
	err := trigger.Execute(event)
	if err == nil {
		t.Error("Expected error from failed action")
	}
}

func TestTypedTrigger_WithManager(t *testing.T) {
	type TestData struct {
		UserID string
		Action string
	}

	manager := NewTriggerManager()
	manager.Start()
	defer manager.Stop()

	var executed bool
	var executedUserID string
	var mutex sync.Mutex

	trigger := NewTypedTrigger[TestData]("user-action", "User Action Handler").
		AddEventType("user-action").
		AddTypedCondition(func(data TestData) bool {
			return data.Action == "login"
		}).
		AddTypedAction(func(data TestData) error {
			mutex.Lock()
			executed = true
			executedUserID = data.UserID
			mutex.Unlock()
			return nil
		})

	// Register typed trigger with manager
	err := manager.RegisterTypedTrigger(trigger)
	if err != nil {
		t.Fatalf("Failed to register typed trigger: %v", err)
	}

	// Fire typed event
	event := NewTypedEvent("user-action", TestData{UserID: "user123", Action: "login"}, "auth")
	err = manager.FireTypedEvent(event)
	if err != nil {
		t.Fatalf("Failed to fire typed event: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	mutex.Lock()
	if !executed {
		t.Error("Typed trigger was not executed")
	}
	if executedUserID != "user123" {
		t.Errorf("Expected UserID 'user123', got '%s'", executedUserID)
	}
	mutex.Unlock()
}

func TestTypedTrigger_ChainedMethods(t *testing.T) {
	type TestData struct {
		Value int
	}

	trigger := NewTypedTrigger[TestData]("test", "Test").
		AddEventType("type1").
		AddEventType("type2").
		SetPriority(PriorityHigh).
		SetEnabled(true).
		SetExecuteOnce(true).
		SetDescription("Test Description").
		AddTypedCondition(func(data TestData) bool { return true }).
		AddTypedAction(func(data TestData) error { return nil })

	underlyingTrigger := trigger.Trigger()

	if len(underlyingTrigger.EventTypes) != 2 {
		t.Errorf("Expected 2 event types, got %d", len(underlyingTrigger.EventTypes))
	}

	if underlyingTrigger.Priority != PriorityHigh {
		t.Errorf("Expected PriorityHigh, got %v", underlyingTrigger.Priority)
	}

	if !underlyingTrigger.ExecuteOnce {
		t.Error("Expected ExecuteOnce to be true")
	}
}

// Tests for Typed Event Constructors

func TestTypedUserEvent(t *testing.T) {
	event := NewTypedUserEvent("user123", "login", "web-app", "auth-service")

	if event.Type() != EventTypeUser {
		t.Errorf("Expected type %s, got %s", EventTypeUser, event.Type())
	}

	data := event.TypedData()
	if data.UserID != "user123" {
		t.Errorf("Expected UserID 'user123', got '%s'", data.UserID)
	}

	if data.Action != "login" {
		t.Errorf("Expected Action 'login', got '%s'", data.Action)
	}

	if data.Target != "web-app" {
		t.Errorf("Expected Target 'web-app', got '%s'", data.Target)
	}
}

func TestTypedSystemEvent(t *testing.T) {
	event := NewTypedSystemEvent("error", "database connection failed", 500, "db-service")

	data := event.TypedData()
	if data.Level != "error" {
		t.Errorf("Expected Level 'error', got '%s'", data.Level)
	}

	if data.Code != 500 {
		t.Errorf("Expected Code 500, got %d", data.Code)
	}
}

func TestTypedCustomEvent(t *testing.T) {
	props := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	}

	event := NewTypedCustomEvent("order", props, "order-service")

	data := event.TypedData()
	if data.Category != "order" {
		t.Errorf("Expected Category 'order', got '%s'", data.Category)
	}

	if data.Properties["key1"] != "value1" {
		t.Errorf("Expected key1 'value1', got '%v'", data.Properties["key1"])
	}
}

// Test real-world scenario: User registration workflow

func TestTypedTrigger_UserRegistrationWorkflow(t *testing.T) {
	type UserRegistrationData struct {
		UserID   string
		Email    string
		Username string
	}

	manager := NewTriggerManager()
	manager.Start()
	defer manager.Stop()

	var steps []string
	var mutex sync.Mutex

	// Step 1: Validate registration
	validationTrigger := NewTypedTrigger[UserRegistrationData]("validate", "Validate Registration").
		AddEventType("user-register").
		AddTypedCondition(func(data UserRegistrationData) bool {
			return data.Email != "" && data.Username != ""
		}).
		AddTypedAction(func(data UserRegistrationData) error {
			mutex.Lock()
			steps = append(steps, "validated")
			mutex.Unlock()
			return nil
		})

	// Step 2: Send welcome email
	emailTrigger := NewTypedTrigger[UserRegistrationData]("email", "Send Welcome Email").
		AddEventType("user-register").
		SetPriority(PriorityNormal).
		AddTypedAction(func(data UserRegistrationData) error {
			mutex.Lock()
			steps = append(steps, fmt.Sprintf("email_sent_%s", data.Email))
			mutex.Unlock()
			return nil
		})

	manager.RegisterTypedTrigger(validationTrigger)
	manager.RegisterTypedTrigger(emailTrigger)

	// Fire registration event
	event := NewTypedEvent("user-register", UserRegistrationData{
		UserID:   "user123",
		Email:    "test@example.com",
		Username: "testuser",
	}, "registration-service")

	manager.FireTypedEvent(event)

	time.Sleep(100 * time.Millisecond)

	mutex.Lock()
	if len(steps) != 2 {
		t.Errorf("Expected 2 steps, got %d: %v", len(steps), steps)
	}
	mutex.Unlock()
}

// Benchmark tests

func BenchmarkTypedEvent_Creation(b *testing.B) {
	type TestData struct {
		Value int
		Name  string
	}

	data := TestData{Value: 42, Name: "test"}

	b.ResetTimer()
	for b.Loop() {
		NewTypedEvent("test", data, "source")
	}
}

func BenchmarkTypedTrigger_Execute(b *testing.B) {
	type TestData struct {
		Value int
	}

	trigger := NewTypedTrigger[TestData]("test", "Test").
		AddEventType("test").
		AddTypedAction(func(data TestData) error {
			return nil
		})

	event := NewTypedEvent("test", TestData{Value: 42}, "source")

	b.ResetTimer()
	for b.Loop() {
		trigger.Execute(event)
	}
}

func BenchmarkTypedTrigger_ConditionCheck(b *testing.B) {
	type TestData struct {
		Value int
	}

	trigger := NewTypedTrigger[TestData]("test", "Test").
		AddEventType("test").
		AddTypedCondition(func(data TestData) bool {
			return data.Value > 10
		})

	event := NewTypedEvent("test", TestData{Value: 42}, "source")

	b.ResetTimer()
	for b.Loop() {
		trigger.CanExecute(event)
	}
}
