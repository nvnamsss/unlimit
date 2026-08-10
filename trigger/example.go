package trigger

import (
	"fmt"
	"log"
	"time"
)

// Example demonstrates how to use the trigger system
func Example() {
	// Create a new trigger manager
	manager := NewTriggerManager()

	// Start the manager to begin processing events
	if err := manager.Start(); err != nil {
		log.Fatalf("Failed to start trigger manager: %v", err)
	}
	defer manager.Stop()

	// Example 1: Simple event trigger
	simpleLoginTrigger := NewTrigger("user-login", "User Login Handler").
		AddEventType(EventTypeUser).
		AddCondition(func(event Event) bool {
			// Only trigger for login events
			if userEvent, ok := event.(*UserEvent); ok {
				return userEvent.Action == "login"
			}
			return false
		}).
		AddAction(func(event Event) error {
			userEvent := event.(*UserEvent)
			fmt.Printf("User %s logged in at %s\n", userEvent.UserID, event.Timestamp().Format(time.RFC3339))
			return nil
		})

	manager.RegisterTrigger(simpleLoginTrigger)

	// Example 2: System monitoring trigger with priority
	criticalErrorTrigger := NewTrigger("critical-error", "Critical Error Handler").
		AddEventType(EventTypeSystem).
		SetPriority(PriorityCritical).
		AddCondition(func(event Event) bool {
			if systemEvent, ok := event.(*SystemEvent); ok {
				return systemEvent.Level == "critical"
			}
			return false
		}).
		AddAction(func(event Event) error {
			systemEvent := event.(*SystemEvent)
			fmt.Printf("CRITICAL ERROR: %s (Code: %d)\n", systemEvent.Message, systemEvent.Code)
			// In a real system, you might send alerts, emails, etc.
			return nil
		})

	manager.RegisterTrigger(criticalErrorTrigger)

	// Example 3: One-time initialization trigger
	initTrigger := NewTrigger("system-init", "System Initialization").
		AddEventType(EventTypeLifecycle).
		SetExecuteOnce(true).
		AddCondition(func(event Event) bool {
			if lifecycleEvent, ok := event.(*LifecycleEvent); ok {
				return lifecycleEvent.Phase == "startup"
			}
			return false
		}).
		AddAction(func(event Event) error {
			fmt.Println("System initialized - running one-time setup tasks")
			return nil
		})

	manager.RegisterTrigger(initTrigger)

	// Example 4: Workflow state machine trigger
	workflowTrigger := NewTrigger("order-workflow", "Order Processing Workflow").
		AddEventType(EventTypeWorkflow).
		AddCondition(func(event Event) bool {
			if workflowEvent, ok := event.(*WorkflowEvent); ok {
				return workflowEvent.WorkflowID == "order-processing" && workflowEvent.ToState == "payment-completed"
			}
			return false
		}).
		AddAction(func(event Event) error {
			workflowEvent := event.(*WorkflowEvent)
			fmt.Printf("Order workflow moved to: %s\n", workflowEvent.ToState)

			// Trigger next step in workflow
			nextEvent := NewWorkflowEvent(
				workflowEvent.WorkflowID,
				"payment-completed",
				"fulfillment-started",
				workflowEvent.Context,
				"order-service",
			)
			return manager.FireEvent(nextEvent)
		})

	manager.RegisterTrigger(workflowTrigger)

	// Fire some example events
	fmt.Println("=== Trigger System Example ===")

	// User login event
	loginEvent := NewUserEvent("user123", "login", "web-app", "auth-service")
	manager.FireEvent(loginEvent)

	// System startup event
	startupEvent := NewLifecycleEvent("startup", map[string]interface{}{
		"version": "1.0.0",
		"mode":    "production",
	}, "system")
	manager.FireEvent(startupEvent)

	// Critical system error
	errorEvent := NewSystemEvent("critical", "database connection failed", 500, "db-service")
	manager.FireEvent(errorEvent)

	// Workflow event
	workflowEvent := NewWorkflowEvent("order-processing", "payment-pending", "payment-completed", map[string]interface{}{
		"order_id": "order-12345",
		"amount":   99.99,
	}, "payment-service")
	manager.FireEvent(workflowEvent)

	// Timer-based event
	manager.CreateTimerEvent("scheduled-backup", 2*time.Second, map[string]interface{}{
		"backup_type": "incremental",
	}, "backup-service")

	// Add a general event listener
	manager.AddEventListener(EventTypeTimer, func(event Event) error {
		fmt.Printf("Timer event fired: %s\n", event.Type())
		return nil
	})

	// Wait for async events to process
	time.Sleep(3 * time.Second)

	// Show manager statistics
	stats := manager.Stats()
	fmt.Printf("\n=== Trigger Manager Stats ===\n")
	fmt.Printf("Total triggers: %v\n", stats["total_triggers"])
	fmt.Printf("Enabled triggers: %v\n", stats["enabled_triggers"])
	fmt.Printf("Is running: %v\n", stats["is_running"])
}

// ExampleCustomEventTypes shows how to create custom event types
func ExampleCustomEventTypes() {
	// You can create your own event types by implementing the Event interface
	type GameEvent struct {
		*BaseEvent
		PlayerID string
		Action   string
		Location string
	}

	newGameEvent := func(playerID, action, location, source string) *GameEvent {
		data := map[string]interface{}{
			"player_id": playerID,
			"action":    action,
			"location":  location,
		}

		return &GameEvent{
			BaseEvent: NewBaseEvent("game", data, source),
			PlayerID:  playerID,
			Action:    action,
			Location:  location,
		}
	}

	// Example usage
	manager := NewTriggerManager()
	manager.Start()
	defer manager.Stop()

	// Create a trigger for game events
	gameTrigger := NewTrigger("player-death", "Player Death Handler").
		AddEventType("game").
		AddCondition(func(event Event) bool {
			if gameEvent, ok := event.(*GameEvent); ok {
				return gameEvent.Action == "death"
			}
			return false
		}).
		AddAction(func(event Event) error {
			gameEvent := event.(*GameEvent)
			fmt.Printf("Player %s died at location %s\n", gameEvent.PlayerID, gameEvent.Location)
			return nil
		})

	manager.RegisterTrigger(gameTrigger)

	// Fire a game event
	deathEvent := newGameEvent("player123", "death", "dungeon-level-5", "game-engine")
	manager.FireEvent(deathEvent)

	time.Sleep(100 * time.Millisecond)
}

// ExampleChainedTriggers demonstrates how triggers can chain events
func ExampleChainedTriggers() {
	manager := NewTriggerManager()
	manager.Start()
	defer manager.Stop()

	// Step 1: User registration trigger
	registrationTrigger := NewTrigger("user-register", "User Registration").
		AddEventType(EventTypeUser).
		AddCondition(func(event Event) bool {
			if userEvent, ok := event.(*UserEvent); ok {
				return userEvent.Action == "register"
			}
			return false
		}).
		AddAction(func(event Event) error {
			userEvent := event.(*UserEvent)
			fmt.Printf("Step 1: User %s registered\n", userEvent.UserID)

			// Chain to welcome email
			emailEvent := NewCustomEvent("send_email", map[string]interface{}{
				"user_id": userEvent.UserID,
				"type":    "welcome",
			}, "user-service")
			return manager.FireEvent(emailEvent)
		})

	// Step 2: Welcome email trigger
	emailTrigger := NewTrigger("welcome-email", "Welcome Email").
		AddEventType(EventTypeCustom).
		AddCondition(func(event Event) bool {
			if customEvent, ok := event.(*CustomEvent); ok {
				return customEvent.Category == "send_email"
			}
			return false
		}).
		AddAction(func(event Event) error {
			customEvent := event.(*CustomEvent)
			userID := customEvent.Properties["user_id"]
			fmt.Printf("Step 2: Sending welcome email to user %v\n", userID)

			// Chain to analytics
			analyticsEvent := NewCustomEvent("track_event", map[string]interface{}{
				"user_id": userID,
				"event":   "user_onboarded",
			}, "email-service")
			return manager.FireEvent(analyticsEvent)
		})

	// Step 3: Analytics trigger
	analyticsTrigger := NewTrigger("analytics", "Analytics Tracking").
		AddEventType(EventTypeCustom).
		AddCondition(func(event Event) bool {
			if customEvent, ok := event.(*CustomEvent); ok {
				return customEvent.Category == "track_event"
			}
			return false
		}).
		AddAction(func(event Event) error {
			customEvent := event.(*CustomEvent)
			userID := customEvent.Properties["user_id"]
			eventName := customEvent.Properties["event"]
			fmt.Printf("Step 3: Tracked event '%v' for user %v\n", eventName, userID)
			return nil
		})

	// Register all triggers
	manager.RegisterTrigger(registrationTrigger)
	manager.RegisterTrigger(emailTrigger)
	manager.RegisterTrigger(analyticsTrigger)

	// Start the chain
	fmt.Println("=== Chained Triggers Example ===")
	registrationEvent := NewUserEvent("newuser456", "register", "signup-form", "web-app")
	manager.FireEvent(registrationEvent)

	time.Sleep(200 * time.Millisecond)
}
