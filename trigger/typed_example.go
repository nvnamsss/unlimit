package trigger

import (
	"fmt"
	"log"
	"time"
)

// ExampleTypedTriggerBasic demonstrates basic typed trigger usage
func ExampleTypedTriggerBasic() {
	type LoginData struct {
		UserID    string
		IPAddress string
		Timestamp time.Time
	}

	manager := NewTriggerManager()
	manager.Start()
	defer manager.Stop()

	// Create a type-safe trigger
	loginTrigger := NewTypedTrigger[LoginData]("user-login", "User Login Handler").
		AddEventType("user-login").
		AddTypedCondition(func(data LoginData) bool {
			// Type-safe condition - no type assertions needed!
			return data.UserID != ""
		}).
		AddTypedAction(func(data LoginData) error {
			// Type-safe action - direct access to typed fields
			fmt.Printf("User %s logged in from %s at %s\n",
				data.UserID, data.IPAddress, data.Timestamp.Format(time.RFC3339))
			return nil
		})

	// Register the typed trigger
	manager.RegisterTypedTrigger(loginTrigger)

	// Fire a typed event
	event := NewTypedEvent("user-login", LoginData{
		UserID:    "user123",
		IPAddress: "192.168.1.1",
		Timestamp: time.Now(),
	}, "auth-service")

	manager.FireTypedEvent(event)

	time.Sleep(100 * time.Millisecond)
}

// ExampleTypedVsNonTyped demonstrates the difference between typed and non-typed approaches
func ExampleTypedVsNonTyped() {
	type UserData struct {
		UserID string
		Email  string
	}

	manager := NewTriggerManager()
	manager.Start()
	defer manager.Stop()

	// OLD WAY: Non-typed (requires type assertions)
	oldTrigger := NewTrigger("old-trigger", "Old Non-Typed Trigger").
		AddEventType("user-event").
		AddCondition(func(event Event) bool {
			// Need type assertion - error-prone!
			if data, ok := event.Data().(UserData); ok {
				return data.UserID != ""
			}
			return false
		}).
		AddAction(func(event Event) error {
			// Need type assertion again
			if data, ok := event.Data().(UserData); ok {
				fmt.Printf("Old way: User %s\n", data.UserID)
			}
			return nil
		})

	// NEW WAY: Typed (type-safe, no assertions needed)
	newTrigger := NewTypedTrigger[UserData]("new-trigger", "New Typed Trigger").
		AddEventType("user-event").
		AddTypedCondition(func(data UserData) bool {
			// Direct access - type-safe!
			return data.UserID != ""
		}).
		AddTypedAction(func(data UserData) error {
			// Direct access - no assertions!
			fmt.Printf("New way: User %s (Email: %s)\n", data.UserID, data.Email)
			return nil
		})

	manager.RegisterTrigger(oldTrigger)
	manager.RegisterTypedTrigger(newTrigger)

	// Fire event
	event := NewTypedEvent("user-event", UserData{
		UserID: "user123",
		Email:  "user@example.com",
	}, "test")

	manager.FireTypedEvent(event)

	time.Sleep(100 * time.Millisecond)
}

// ExampleTypedWorkflow demonstrates a complex workflow using typed triggers
func ExampleTypedWorkflow() {
	type OrderData struct {
		OrderID    string
		CustomerID string
		Amount     float64
		Status     string
	}

	manager := NewTriggerManager()
	manager.Start()
	defer manager.Stop()

	// Step 1: Validate order
	validateTrigger := NewTypedTrigger[OrderData]("validate-order", "Validate Order").
		AddEventType("order-created").
		SetPriority(PriorityCritical).
		AddTypedCondition(func(data OrderData) bool {
			return data.Amount > 0 && data.CustomerID != ""
		}).
		AddTypedAction(func(data OrderData) error {
			fmt.Printf("✓ Order %s validated (Amount: $%.2f)\n", data.OrderID, data.Amount)

			// Chain to next step
			processEvent := NewTypedEvent("order-validated", OrderData{
				OrderID:    data.OrderID,
				CustomerID: data.CustomerID,
				Amount:     data.Amount,
				Status:     "validated",
			}, "validation-service")

			return manager.FireTypedEvent(processEvent)
		})

	// Step 2: Process payment
	paymentTrigger := NewTypedTrigger[OrderData]("process-payment", "Process Payment").
		AddEventType("order-validated").
		SetPriority(PriorityHigh).
		AddTypedAction(func(data OrderData) error {
			fmt.Printf("💳 Processing payment for order %s ($%.2f)\n", data.OrderID, data.Amount)

			// Chain to next step
			fulfillEvent := NewTypedEvent("payment-completed", OrderData{
				OrderID:    data.OrderID,
				CustomerID: data.CustomerID,
				Amount:     data.Amount,
				Status:     "paid",
			}, "payment-service")

			return manager.FireTypedEvent(fulfillEvent)
		})

	// Step 3: Fulfill order
	fulfillmentTrigger := NewTypedTrigger[OrderData]("fulfill-order", "Fulfill Order").
		AddEventType("payment-completed").
		SetPriority(PriorityNormal).
		AddTypedAction(func(data OrderData) error {
			fmt.Printf("📦 Fulfilling order %s for customer %s\n", data.OrderID, data.CustomerID)
			return nil
		})

	// Step 4: Send notification (runs in parallel with fulfillment)
	notificationTrigger := NewTypedTrigger[OrderData]("notify-customer", "Notify Customer").
		AddEventType("payment-completed").
		SetPriority(PriorityNormal).
		AddTypedAction(func(data OrderData) error {
			fmt.Printf("📧 Sending confirmation email to customer %s\n", data.CustomerID)
			return nil
		})

	// Register all triggers
	manager.RegisterTypedTrigger(validateTrigger)
	manager.RegisterTypedTrigger(paymentTrigger)
	manager.RegisterTypedTrigger(fulfillmentTrigger)
	manager.RegisterTypedTrigger(notificationTrigger)

	// Start the workflow
	fmt.Println("=== Order Processing Workflow ===")
	orderEvent := NewTypedEvent("order-created", OrderData{
		OrderID:    "order-12345",
		CustomerID: "customer-789",
		Amount:     99.99,
		Status:     "created",
	}, "order-service")

	manager.FireTypedEvent(orderEvent)

	time.Sleep(200 * time.Millisecond)
}

// ExampleTypedWithPredefinedEvents demonstrates using typed predefined events
func ExampleTypedWithPredefinedEvents() {
	manager := NewTriggerManager()
	manager.Start()
	defer manager.Stop()

	// Using typed UserEventData
	userTrigger := NewTypedTrigger[UserEventData]("user-actions", "User Action Handler").
		AddEventType(EventTypeUser).
		AddTypedCondition(func(data UserEventData) bool {
			// Type-safe access to UserEventData fields
			return data.Action == "login" || data.Action == "logout"
		}).
		AddTypedAction(func(data UserEventData) error {
			fmt.Printf("User %s performed action: %s on %s\n",
				data.UserID, data.Action, data.Target)
			return nil
		})

	// Using typed SystemEventData
	systemTrigger := NewTypedTrigger[SystemEventData]("system-monitor", "System Monitor").
		AddEventType(EventTypeSystem).
		AddTypedCondition(func(data SystemEventData) bool {
			// Type-safe access to SystemEventData fields
			return data.Level == "error" || data.Level == "critical"
		}).
		AddTypedAction(func(data SystemEventData) error {
			fmt.Printf("🚨 System Alert [%s]: %s (Code: %d)\n",
				data.Level, data.Message, data.Code)
			return nil
		})

	manager.RegisterTypedTrigger(userTrigger)
	manager.RegisterTypedTrigger(systemTrigger)

	// Fire typed predefined events
	loginEvent := NewTypedUserEvent("user123", "login", "web-app", "auth")
	systemEvent := NewTypedSystemEvent("error", "Database connection failed", 500, "db")

	manager.FireTypedEvent(loginEvent)
	manager.FireTypedEvent(systemEvent)

	time.Sleep(100 * time.Millisecond)
}

// ExampleMixedTypedAndNonTyped demonstrates mixing typed and non-typed triggers
func ExampleMixedTypedAndNonTyped() {
	type SpecificData struct {
		Value int
	}

	manager := NewTriggerManager()
	manager.Start()
	defer manager.Stop()

	// Typed trigger for specific data type
	typedTrigger := NewTypedTrigger[SpecificData]("typed", "Typed Trigger").
		AddEventType("specific-event").
		AddTypedCondition(func(data SpecificData) bool {
			return data.Value > 10
		}).
		AddTypedAction(func(data SpecificData) error {
			fmt.Printf("Typed: Received value %d\n", data.Value)
			return nil
		})

	// Non-typed trigger for any event type
	generalTrigger := NewTrigger("general", "General Trigger").
		AddEventType("specific-event").
		AddEventType("other-event").
		AddAction(func(event Event) error {
			fmt.Printf("General: Received event of type %s from %s\n",
				event.Type(), event.Source())
			return nil
		})

	manager.RegisterTypedTrigger(typedTrigger)
	manager.RegisterTrigger(generalTrigger)

	// Both triggers will fire for this event
	event := NewTypedEvent("specific-event", SpecificData{Value: 42}, "test")
	manager.FireTypedEvent(event)

	time.Sleep(100 * time.Millisecond)
}

// ExampleTypedConditionalChaining demonstrates conditional event chaining
func ExampleTypedConditionalChaining() {
	type TransactionData struct {
		TransactionID string
		Amount        float64
		IsFraud       bool
	}

	manager := NewTriggerManager()
	manager.Start()
	defer manager.Stop()

	// Check for fraud
	fraudCheckTrigger := NewTypedTrigger[TransactionData]("fraud-check", "Fraud Detection").
		AddEventType("transaction-submitted").
		AddTypedAction(func(data TransactionData) error {
			fmt.Printf("Checking transaction %s for fraud...\n", data.TransactionID)

			// Simulate fraud detection
			isFraud := data.Amount > 10000

			nextEvent := NewTypedEvent("fraud-check-complete", TransactionData{
				TransactionID: data.TransactionID,
				Amount:        data.Amount,
				IsFraud:       isFraud,
			}, "fraud-service")

			return manager.FireTypedEvent(nextEvent)
		})

	// Handle legitimate transactions
	approveTrigger := NewTypedTrigger[TransactionData]("approve", "Approve Transaction").
		AddEventType("fraud-check-complete").
		AddTypedCondition(func(data TransactionData) bool {
			return !data.IsFraud
		}).
		AddTypedAction(func(data TransactionData) error {
			fmt.Printf("✓ Transaction %s approved ($%.2f)\n",
				data.TransactionID, data.Amount)
			return nil
		})

	// Handle fraudulent transactions
	blockTrigger := NewTypedTrigger[TransactionData]("block", "Block Transaction").
		AddEventType("fraud-check-complete").
		AddTypedCondition(func(data TransactionData) bool {
			return data.IsFraud
		}).
		AddTypedAction(func(data TransactionData) error {
			fmt.Printf("⛔ Transaction %s blocked due to fraud ($%.2f)\n",
				data.TransactionID, data.Amount)
			return nil
		})

	manager.RegisterTypedTrigger(fraudCheckTrigger)
	manager.RegisterTypedTrigger(approveTrigger)
	manager.RegisterTypedTrigger(blockTrigger)

	// Test with normal transaction
	normalTransaction := NewTypedEvent("transaction-submitted", TransactionData{
		TransactionID: "txn-001",
		Amount:        99.99,
	}, "payment-gateway")

	// Test with suspicious transaction
	suspiciousTransaction := NewTypedEvent("transaction-submitted", TransactionData{
		TransactionID: "txn-002",
		Amount:        15000.00,
	}, "payment-gateway")

	fmt.Println("=== Transaction Processing ===")
	manager.FireTypedEvent(normalTransaction)
	time.Sleep(50 * time.Millisecond)

	manager.FireTypedEvent(suspiciousTransaction)
	time.Sleep(50 * time.Millisecond)
}

// ExampleTypedHelperFunctions demonstrates using helper functions
func ExampleTypedHelperFunctions() {
	type AlertData struct {
		Level   string
		Message string
	}

	manager := NewTriggerManager()
	manager.Start()
	defer manager.Stop()

	// Using NewTypedTriggerComplete for concise creation
	alertTrigger := NewTypedTriggerComplete(
		"critical-alert",
		"Critical Alert Handler",
		"alert",
		func(data AlertData) bool { return data.Level == "critical" },
		func(data AlertData) error {
			fmt.Printf("🚨 CRITICAL: %s\n", data.Message)
			return nil
		},
	)

	// Using NewTypedTriggerWithFilter
	warningTrigger := NewTypedTriggerWithFilter(
		"warning-alert",
		"Warning Alert Handler",
		"alert",
		func(data AlertData) bool { return data.Level == "warning" },
	).AddTypedAction(func(data AlertData) error {
		fmt.Printf("⚠️  WARNING: %s\n", data.Message)
		return nil
	})

	manager.RegisterTypedTrigger(alertTrigger)
	manager.RegisterTypedTrigger(warningTrigger)

	// Fire events
	criticalEvent := NewTypedEvent("alert", AlertData{
		Level:   "critical",
		Message: "System overload detected",
	}, "monitor")

	warningEvent := NewTypedEvent("alert", AlertData{
		Level:   "warning",
		Message: "High memory usage",
	}, "monitor")

	manager.FireTypedEvent(criticalEvent)
	manager.FireTypedEvent(warningEvent)

	time.Sleep(100 * time.Millisecond)
}

// ExampleTypedMigrationPath demonstrates how to gradually migrate from non-typed to typed
func ExampleTypedMigrationPath() {
	type LegacyData struct {
		ID   string
		Data map[string]interface{}
	}

	manager := NewTriggerManager()
	manager.Start()
	defer manager.Stop()

	// Phase 1: Keep existing non-typed triggers running
	legacyTrigger := NewTrigger("legacy", "Legacy Trigger").
		AddEventType("legacy-event").
		AddAction(func(event Event) error {
			log.Println("Legacy trigger still works!")
			return nil
		})

	// Phase 2: Add new typed triggers alongside
	modernTrigger := NewTypedTrigger[LegacyData]("modern", "Modern Typed Trigger").
		AddEventType("legacy-event").
		AddTypedAction(func(data LegacyData) error {
			fmt.Printf("Modern trigger with type-safe access: ID=%s\n", data.ID)
			return nil
		})

	manager.RegisterTrigger(legacyTrigger)
	manager.RegisterTypedTrigger(modernTrigger)

	// Both triggers work with the same event
	event := NewTypedEvent("legacy-event", LegacyData{
		ID:   "item-123",
		Data: map[string]interface{}{"key": "value"},
	}, "service")

	manager.FireTypedEvent(event)

	time.Sleep(100 * time.Millisecond)

	// Phase 3: Eventually remove legacy trigger and keep only typed version
	fmt.Println("\n✓ Migration complete: Both typed and non-typed triggers coexist!")
}
