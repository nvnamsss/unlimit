package main

import (
	"log"
	"time"

	"github.com/voidforge-studios/unlimit/trigger"
)

type LoginData struct {
	UserID    string    `json:"user_id"`
	IPAddress string    `json:"ip_address"`
	Timestamp time.Time `json:"timestamp"`
}

func main() {
	// Create a new trigger manager
	manager := trigger.NewTriggerManager()

	// Start the manager to begin processing events
	if err := manager.Start(); err != nil {
		log.Fatalf("Failed to start trigger manager: %v", err)
	}
	defer manager.Stop()

	// Example 1: Simple event trigger
	simpleLoginTrigger := trigger.NewTrigger("user-login", "User Login Handler").
		AddEventType(trigger.EventTypeUser).
		AddCondition(func(event trigger.Event) bool {
			// Only trigger for login events
			if userEvent, ok := event.(*trigger.UserEvent); ok {
				return userEvent.Action == "login"
			}
			return false
		}).
		AddAction(func(event trigger.Event) error {
			userEvent := event.(*trigger.UserEvent)
			log.Printf("User %s logged in at %s\n", userEvent.UserID, event.Timestamp().Format(time.RFC3339))
			return nil
		})

	// Create a type-safe trigger
	loginTrigger := trigger.NewTypedTrigger[LoginData]("user-login-2", "User Login Handler").
		AddEventType(trigger.EventTypeUser).
		AddTypedCondition(func(data LoginData) bool {
			// Type-safe condition - no type assertions needed!
			return data.UserID != ""
		}).
		AddTypedAction(func(data LoginData) error {
			// Type-safe action - direct access to typed fields
			log.Printf("User %s logged in from %s at %s",
				data.UserID, data.IPAddress, data.Timestamp.Format(time.RFC3339))
			return nil
		})
	_ = loginTrigger
	manager.RegisterTypedTrigger(loginTrigger)
	manager.RegisterTrigger(simpleLoginTrigger)

	manager.FireEvent(trigger.NewUserEvent("user1", "login", "system", "main"))
}
