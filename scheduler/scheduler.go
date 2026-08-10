package scheduler

import (
	"context"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

type TaskHandler func() error

type ScheduleType int

const (
	Interval ScheduleType = iota
	Cron
	OneOff
)

type Task struct {
	ID           string
	Handler      TaskHandler
	ScheduleType ScheduleType
	Interval     time.Duration // for interval-based
	CronExpr     string        // for cron-like (simplified)
	NextRun      time.Time     // next execution time for debugging
	timer        *time.Timer   // timer for execution
	OneTime      bool          // for one-off jobs
}

type Scheduler struct {
	tasks   map[string]*Task
	mu      sync.RWMutex
	running bool
	ctx     context.Context
	cancel  context.CancelFunc
}

func New() *Scheduler {
	return &Scheduler{
		tasks: make(map[string]*Task),
	}
}

func (s *Scheduler) RegisterIntervalTask(id string, handler TaskHandler, interval time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	nextRun := time.Now().Add(interval)
	task := &Task{
		ID:           id,
		Handler:      handler,
		ScheduleType: Interval,
		Interval:     interval,
		NextRun:      nextRun,
	}

	if s.running {
		task.timer = time.AfterFunc(interval, func() {
			s.executeTask(task)
		})
	}

	s.tasks[id] = task
}

func (s *Scheduler) RegisterCronTask(id string, handler TaskHandler, cronExpr string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	duration := s.getDurationUntilCron(cronExpr)
	nextRun := time.Now().Add(duration)
	task := &Task{
		ID:           id,
		Handler:      handler,
		ScheduleType: Cron,
		CronExpr:     cronExpr,
		NextRun:      nextRun,
	}

	if s.running {
		task.timer = time.AfterFunc(duration, func() {
			s.executeTask(task)
		})
	}

	s.tasks[id] = task
}

func (s *Scheduler) RegisterOneOffTask(id string, handler TaskHandler, runAt time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	duration := time.Until(runAt)
	if duration < 0 {
		duration = 0
	}

	task := &Task{
		ID:           id,
		Handler:      handler,
		ScheduleType: OneOff,
		NextRun:      runAt,
		OneTime:      true,
	}

	if s.running {
		task.timer = time.AfterFunc(duration, func() {
			s.executeTask(task)
		})
	}

	s.tasks[id] = task
}

func (s *Scheduler) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return
	}
	s.running = true
	s.ctx, s.cancel = context.WithCancel(context.Background())

	// Start timers for all registered tasks
	for _, task := range s.tasks {
		s.startTaskTimer(task)
	}
}

func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		s.cancel()
		s.running = false

		// Stop all timers
		for _, task := range s.tasks {
			if task.timer != nil {
				task.timer.Stop()
			}
		}
	}
}

func (s *Scheduler) RemoveTask(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.tasks[id]
	if !exists {
		return false
	}

	// Stop the timer if it exists
	if task.timer != nil {
		task.timer.Stop()
	}

	// Remove the task from the map
	delete(s.tasks, id)

	return true
}

func (s *Scheduler) startTaskTimer(task *Task) {
	var duration time.Duration

	switch task.ScheduleType {
	case Interval:
		duration = task.Interval
		task.NextRun = time.Now().Add(duration)
	case Cron:
		duration = s.getDurationUntilCron(task.CronExpr)
		task.NextRun = time.Now().Add(duration)
	case OneOff:
		// For one-off tasks, NextRun is already set during registration
		duration = time.Until(task.NextRun)
		if duration < 0 {
			duration = 0
		}
	}

	task.timer = time.AfterFunc(duration, func() {
		s.executeTask(task)
	})
}

func (s *Scheduler) executeTask(task *Task) {
	// Check if scheduler is still running
	s.mu.RLock()
	running := s.running
	s.mu.RUnlock()

	if !running {
		return
	}

	// Execute the task handler
	go task.Handler()

	s.mu.Lock()
	defer s.mu.Unlock()

	// Handle task cleanup and rescheduling
	if task.OneTime {
		delete(s.tasks, task.ID)
	} else if s.running {
		// Reschedule the task
		s.startTaskTimer(task)
	}
}

func (s *Scheduler) getDurationUntilCron(expr string) time.Duration {
	schedule, err := cron.ParseStandard(expr)
	if err != nil {
		// Default to next hour for invalid expressions
		return time.Hour
	}

	now := time.Now()
	next := schedule.Next(now)
	return next.Sub(now)
}
