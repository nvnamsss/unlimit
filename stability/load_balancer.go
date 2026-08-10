package stability

import (
	"context"
	"errors"
	"sync"
	"time"
)

// LoadBalancerStats contains statistics about the load balancer
type LoadBalancerStats struct {
	TotalServers    int
	HealthyServers  int
	TotalRequests   int64
	FailedRequests  int64
	LastHealthCheck time.Time
}

// LoadBalancer defines a generic interface for load balancing requests to any data structure
type LoadBalancer[T any] interface {
	// SelectTarget selects the next target based on the load balancing strategy
	SelectTarget(ctx context.Context) (string, T, error)

	// GetTargets returns all targets in the pool
	GetTargets() map[string]Target[T]

	// GetAvailableTargets returns only available targets
	GetAvailableTargets() map[string]Target[T]

	// SetTargetAvailability updates the availability status of a target
	SetTargetAvailability(id string, available bool) error

	// GetStats returns load balancing statistics
	GetStats() LoadBalancerStats

	// Reset resets all statistics and state
	Reset()
}

// Target represents a generic target in the load balancer pool
type Target[T any] struct {
	ID        string
	Data      T
	Weight    int
	Available bool
	Requests  int64
	LastUsed  time.Time
}

// LoadBalancingStrategy defines the strategy for selecting targets
type LoadBalancingStrategy int

const (
	RoundRobin LoadBalancingStrategy = iota
	WeightedRoundRobin
	LeastConnections
	Random
	WeightedRandom
)

// GenericLoadBalancerStats contains statistics about the generic load balancer
type GenericLoadBalancerStats struct {
	TotalTargets     int
	AvailableTargets int
	TotalRequests    int64
	Strategy         LoadBalancingStrategy
	LastSelection    time.Time
}

// RoundRobinLoadBalancer implements LoadBalancer using round-robin strategy
type RoundRobinLoadBalancer[T any] struct {
	mu           sync.RWMutex
	targets      map[string]Target[T]
	targetOrder  []string
	currentIndex int
	stats        LoadBalancerStats
}

// NewRoundRobinLoadBalancer creates a new round-robin load balancer
func NewRoundRobinLoadBalancer[T any]() *RoundRobinLoadBalancer[T] {
	return &RoundRobinLoadBalancer[T]{
		targets:     make(map[string]Target[T]),
		targetOrder: make([]string, 0),
	}
}

// SelectTarget selects the next target using round-robin strategy
func (rb *RoundRobinLoadBalancer[T]) SelectTarget(ctx context.Context) (string, T, error) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	availableTargets := rb.getAvailableTargetIDs()
	if len(availableTargets) == 0 {
		var zero T
		return "", zero, ErrNoTargetsAvailable
	}

	// Find next available target in round-robin order
	for i := 0; i < len(availableTargets); i++ {
		targetID := availableTargets[rb.currentIndex%len(availableTargets)]
		rb.currentIndex = (rb.currentIndex + 1) % len(availableTargets)

		if target, exists := rb.targets[targetID]; exists && target.Available {
			// Update target statistics
			target.Requests++
			target.LastUsed = time.Now()
			rb.targets[targetID] = target

			// Update stats
			rb.stats.TotalRequests++

			return targetID, target.Data, nil
		}
	}

	var zero T
	return "", zero, ErrNoTargetsAvailable
}

// GetTargets returns all targets in the pool
func (rb *RoundRobinLoadBalancer[T]) GetTargets() map[string]Target[T] {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	result := make(map[string]Target[T])
	for k, v := range rb.targets {
		result[k] = v
	}
	return result
}

// GetAvailableTargets returns only available targets
func (rb *RoundRobinLoadBalancer[T]) GetAvailableTargets() map[string]Target[T] {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	result := make(map[string]Target[T])
	for k, v := range rb.targets {
		if v.Available {
			result[k] = v
		}
	}
	return result
}

// SetTargetAvailability updates the availability status of a target
func (rb *RoundRobinLoadBalancer[T]) SetTargetAvailability(id string, available bool) error {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	target, exists := rb.targets[id]
	if !exists {
		return ErrTargetNotFound
	}

	target.Available = available
	rb.targets[id] = target
	return nil
}

// GetStats returns load balancing statistics
func (rb *RoundRobinLoadBalancer[T]) GetStats() LoadBalancerStats {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	totalTargets := len(rb.targets)
	healthyTargets := 0
	for _, target := range rb.targets {
		if target.Available {
			healthyTargets++
		}
	}

	return LoadBalancerStats{
		TotalServers:   totalTargets,
		HealthyServers: healthyTargets,
		TotalRequests:  rb.stats.TotalRequests,
		FailedRequests: rb.stats.FailedRequests,
	}
}

// Reset resets all statistics and state
func (rb *RoundRobinLoadBalancer[T]) Reset() {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	rb.currentIndex = 0
	rb.stats = LoadBalancerStats{}
	for id, target := range rb.targets {
		target.Requests = 0
		target.LastUsed = time.Time{}
		rb.targets[id] = target
	}
}

// AddTarget adds a new target to the pool
func (rb *RoundRobinLoadBalancer[T]) AddTarget(id string, data T, weight int) error {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	if _, exists := rb.targets[id]; exists {
		return ErrTargetExists
	}

	rb.targets[id] = Target[T]{
		ID:        id,
		Data:      data,
		Weight:    weight,
		Available: true,
		Requests:  0,
		LastUsed:  time.Time{},
	}

	rb.targetOrder = append(rb.targetOrder, id)
	return nil
}

// RemoveTarget removes a target from the pool
func (rb *RoundRobinLoadBalancer[T]) RemoveTarget(id string) error {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	if _, exists := rb.targets[id]; !exists {
		return ErrTargetNotFound
	}

	delete(rb.targets, id)

	// Remove from target order
	for i, targetID := range rb.targetOrder {
		if targetID == id {
			rb.targetOrder = append(rb.targetOrder[:i], rb.targetOrder[i+1:]...)
			break
		}
	}

	// Adjust current index if needed
	if rb.currentIndex >= len(rb.targetOrder) {
		rb.currentIndex = 0
	}

	return nil
}

// getAvailableTargetIDs returns slice of available target IDs in order
func (rb *RoundRobinLoadBalancer[T]) getAvailableTargetIDs() []string {
	var available []string
	for _, id := range rb.targetOrder {
		if target, exists := rb.targets[id]; exists && target.Available {
			available = append(available, id)
		}
	}
	return available
}

// Common errors
var (
	ErrNoTargetsAvailable = errors.New("no targets available")
	ErrTargetNotFound     = errors.New("target not found")
	ErrTargetExists       = errors.New("target already exists")
	ErrInvalidWeight      = errors.New("invalid weight value")
)
