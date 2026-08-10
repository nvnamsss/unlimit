package trigger

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"
)

// Error variables for TriggerManager
var (
	ErrTriggerAlreadyExists   = errors.New("trigger already exists")
	ErrTriggerNotFound        = errors.New("trigger not found")
	ErrManagerShuttingDown    = errors.New("trigger manager is shutting down")
	ErrTriggerExecutionFailed = errors.New("failed to execute trigger")
	ErrEventListenerFailed    = errors.New("event listener failed")
)

// TriggerManager manages the registration and execution of triggers
type TriggerManager struct {
	triggers       map[string]*Trigger
	eventListeners map[string][]EventListener
	mutex          sync.RWMutex
	ctx            context.Context
	cancel         context.CancelFunc
}

// NewTriggerManager creates a new trigger manager
func NewTriggerManager() *TriggerManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &TriggerManager{
		triggers:       make(map[string]*Trigger),
		eventListeners: make(map[string][]EventListener),
		ctx:            ctx,
		cancel:         cancel,
	}
}

// RegisterTrigger registers a new trigger
func (tm *TriggerManager) RegisterTrigger(trigger *Trigger) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	if _, exists := tm.triggers[trigger.ID]; exists {
		return fmt.Errorf("%w: %s", ErrTriggerAlreadyExists, trigger.ID)
	}

	tm.triggers[trigger.ID] = trigger
	return nil
}

// RegisterTypedTrigger registers a typed trigger
func (tm *TriggerManager) RegisterTypedTrigger(typedTrigger interface{ Trigger() *Trigger }) error {
	return tm.RegisterTrigger(typedTrigger.Trigger())
}

// UnregisterTrigger removes a trigger
func (tm *TriggerManager) UnregisterTrigger(triggerID string) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	if _, exists := tm.triggers[triggerID]; !exists {
		return fmt.Errorf("%w: %s", ErrTriggerNotFound, triggerID)
	}

	delete(tm.triggers, triggerID)
	return nil
}

// GetTrigger retrieves a trigger by ID
func (tm *TriggerManager) GetTrigger(triggerID string) (*Trigger, error) {
	tm.mutex.RLock()
	trigger, exists := tm.triggers[triggerID]
	tm.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrTriggerNotFound, triggerID)
	}

	return trigger, nil
}

// ListTriggers returns all registered triggers
func (tm *TriggerManager) ListTriggers() []*Trigger {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	triggers := make([]*Trigger, 0, len(tm.triggers))
	for _, trigger := range tm.triggers {
		triggers = append(triggers, trigger)
	}

	return triggers
}

// AddEventListener adds a general event listener for a specific event type
func (tm *TriggerManager) AddEventListener(eventType string, listener EventListener) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	if _, exists := tm.eventListeners[eventType]; !exists {
		tm.eventListeners[eventType] = make([]EventListener, 0)
	}

	tm.eventListeners[eventType] = append(tm.eventListeners[eventType], listener)
}

// FireEvent fires an event and processes all matching triggers immediately
func (tm *TriggerManager) FireEvent(event Event) error {
	select {
	case <-tm.ctx.Done():
		return ErrManagerShuttingDown
	default:
		return tm.processEvent(event)
	}
}

// FireTypedEvent fires a typed event
func (tm *TriggerManager) FireTypedEvent(event interface{ Event }) error {
	return tm.FireEvent(event)
}

// Start starts the trigger manager (deprecated - kept for backward compatibility)
// The manager now processes events immediately, no background loop needed
func (tm *TriggerManager) Start() error {
	// No-op: Events are now processed immediately on FireEvent
	return nil
}

// Stop stops the trigger manager and cancels all timer events
func (tm *TriggerManager) Stop() error {
	tm.cancel()
	return nil
}

// processEvent processes a single event against all registered triggers
func (tm *TriggerManager) processEvent(event Event) error {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	// Collect matching triggers
	var matchingTriggers []*Trigger
	for _, trigger := range tm.triggers {
		if trigger.CanExecute(event) {
			matchingTriggers = append(matchingTriggers, trigger)
		}
	}

	// Sort triggers by priority (highest priority first)
	sort.Slice(matchingTriggers, func(i, j int) bool {
		return matchingTriggers[i].Priority > matchingTriggers[j].Priority
	})

	// Execute triggers
	for _, trigger := range matchingTriggers {
		if err := trigger.Execute(event); err != nil {
			return fmt.Errorf("%w %s: %w", ErrTriggerExecutionFailed, trigger.ID, err)
		}
	}

	// Execute general event listeners
	if listeners, exists := tm.eventListeners[event.Type()]; exists {
		for _, listener := range listeners {
			if err := listener(event); err != nil {
				return fmt.Errorf("%w for event type %s: %w", ErrEventListenerFailed, event.Type(), err)
			}
		}
	}

	return nil
}

// CreateTimerEvent creates a timer-based event that fires after a specified duration
func (tm *TriggerManager) CreateTimerEvent(eventType string, duration time.Duration, data interface{}, source string) {
	go func() {
		select {
		case <-time.After(duration):
			event := NewBaseEvent(eventType, data, source)
			tm.FireEvent(event)
		case <-tm.ctx.Done():
			return
		}
	}()
}

// CreatePeriodicEvent creates a periodic event that fires at regular intervals
func (tm *TriggerManager) CreatePeriodicEvent(eventType string, interval time.Duration, data interface{}, source string) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {

			select {
			case <-ticker.C:
				event := NewBaseEvent(eventType, data, source)
				tm.FireEvent(event)
			case <-tm.ctx.Done():
				return
			}
		}
	}()
}

// Stats returns statistics about the trigger manager
func (tm *TriggerManager) Stats() map[string]interface{} {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	return map[string]interface{}{
		"total_triggers":   len(tm.triggers),
		"enabled_triggers": tm.countEnabledTriggers(),
	}
}

// countEnabledTriggers counts the number of enabled triggers
func (tm *TriggerManager) countEnabledTriggers() int {
	count := 0
	for _, trigger := range tm.triggers {
		if trigger.Enabled {
			count++
		}
	}
	return count
}
