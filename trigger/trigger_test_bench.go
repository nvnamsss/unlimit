package trigger

import (
	"testing"
)

// BenchmarkTriggerExecution benchmarks the complete trigger execution flow
func BenchmarkTriggerExecution(b *testing.B) {
	manager := NewTriggerManager()

	trigger := NewTrigger("bench-trigger", "Benchmark Trigger").
		AddEventType(EventTypeCustom).
		AddAction(func(event Event) error {
			return nil
		})

	manager.RegisterTrigger(trigger)

	event := NewCustomEvent("benchmark", nil, "bench")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.FireEvent(event)
	}
}

// BenchmarkTrigger_CanExecute benchmarks condition checking
func BenchmarkTrigger_CanExecute(b *testing.B) {
	trigger := NewTrigger("bench", "Bench").
		AddEventType(EventTypeUser).
		AddCondition(func(event Event) bool {
			if userEvent, ok := event.(*UserEvent); ok {
				return userEvent.Action == "login"
			}
			return false
		})

	event := NewUserEvent("user1", "login", "system", "test")

	b.ResetTimer()
	for b.Loop() {
		trigger.CanExecute(event)
	}
}

// BenchmarkTrigger_Execute benchmarks action execution
func BenchmarkTrigger_Execute(b *testing.B) {
	trigger := NewTrigger("bench", "Bench").
		AddEventType(EventTypeUser).
		AddAction(func(event Event) error {
			return nil
		})

	event := NewUserEvent("user1", "login", "system", "test")

	b.ResetTimer()
	for b.Loop() {
		trigger.Execute(event)
	}
}
