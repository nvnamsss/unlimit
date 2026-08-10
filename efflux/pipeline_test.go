package efflux

import (
	"context"
	"errors"
	"testing"
)

type testData struct {
	Val     int
	IsValid bool
}

func TestBasePipeline_New(t *testing.T) {
	p := NewBasePipeline[testData]()
	if p == nil {
		t.Error("NewBasePipeline should not return nil")
	}
	if len(p.GetAllStages()) != 0 {
		t.Errorf("expected no stages, got %d", len(p.GetAllStages()))
	}
}

func TestBasePipeline_AddAndGetStage(t *testing.T) {
	p := NewBasePipeline[testData]()
	s := &Stage[testData]{Name: "A"}
	p.AddStage(s)
	got := p.GetStage("A")
	if got != s {
		t.Errorf("GetStage did not return the correct stage")
	}
}

func TestBasePipeline_SetRootStage(t *testing.T) {
	p := NewBasePipeline[testData]()
	s := &Stage[testData]{Name: "root"}
	p.SetRootStage(s)
	if p.GetRootStage() != s {
		t.Errorf("GetRootStage did not return the correct root stage")
	}
}

func TestBasePipeline_Execute_Linear(t *testing.T) {
	p := NewBasePipeline[testData]()
	order := make([]string, 0)
	s1 := &Stage[testData]{Name: "A", Func: func(ctx context.Context, v testData) error {
		order = append(order, "A")
		return nil
	}}
	s2 := &Stage[testData]{Name: "B", Func: func(ctx context.Context, v testData) error {
		order = append(order, "B")
		return nil
	}}
	s1.AddTransition("toB", s2, func(ctx context.Context, v testData, stage *Stage[testData]) bool { return true })
	p.SetRootStage(s1).AddStage(s1).AddStage(s2)
	err := p.Execute(context.Background(), testData{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(order) != 2 || order[0] != "A" || order[1] != "B" {
		t.Errorf("unexpected execution order: %v", order)
	}
}

func TestBasePipeline_Execute_Branching(t *testing.T) {
	p := NewBasePipeline[testData]()
	result := ""
	s1 := &Stage[testData]{Name: "A", Func: func(ctx context.Context, v testData) error { return nil }}
	s2 := &Stage[testData]{Name: "B", Func: func(ctx context.Context, v testData) error { result = "B"; return nil }}
	s3 := &Stage[testData]{Name: "C", Func: func(ctx context.Context, v testData) error { result = "C"; return nil }}
	s1.AddTransition("toB", s2, func(ctx context.Context, v testData, stage *Stage[testData]) bool { return v.IsValid })
	s1.AddTransition("toC", s3, func(ctx context.Context, v testData, stage *Stage[testData]) bool { return !v.IsValid })
	p.SetRootStage(s1).AddStage(s1).AddStage(s2).AddStage(s3)

	err := p.Execute(context.Background(), testData{IsValid: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "B" {
		t.Errorf("expected result 'B', got '%s'", result)
	}

	result = ""
	err = p.Execute(context.Background(), testData{IsValid: false})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "C" {
		t.Errorf("expected result 'C', got '%s'", result)
	}
}

func TestBasePipeline_Execute_ErrorHandling(t *testing.T) {
	p := NewBasePipeline[testData]()
	s := &Stage[testData]{Name: "fail", Func: func(ctx context.Context, v testData) error { return errors.New("fail") }}
	p.SetRootStage(s).AddStage(s)
	err := p.Execute(context.Background(), testData{})
	if err == nil {
		t.Error("expected error, got nil")
	}
	if !errors.Is(err, errors.New("fail")) && err.Error() != "stage 'fail' failed: fail" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestBasePipeline_Execute_CycleDetection(t *testing.T) {
	p := NewBasePipeline[testData]()
	s := &Stage[testData]{Name: "cycle", Func: func(ctx context.Context, v testData) error { return nil }}
	s.AddTransition("self", s, func(ctx context.Context, v testData, stage *Stage[testData]) bool { return true })
	p.SetRootStage(s).AddStage(s)
	err := p.Execute(context.Background(), testData{})
	if err == nil {
		t.Error("expected cycle error, got nil")
	}
	if !errors.Is(err, ErrCycleDetected) {
		t.Errorf("expected ErrCycleDetected, got %v", err)
	}
}

func TestBasePipeline_Validate_NoRoot(t *testing.T) {
	p := NewBasePipeline[testData]()
	err := p.Validate()
	if err == nil {
		t.Error("expected ErrNoRootStage, got nil")
	}
	if !errors.Is(err, ErrNoRootStage) {
		t.Errorf("expected ErrNoRootStage, got %v", err)
	}
}

func TestBasePipeline_Validate_Unreachable(t *testing.T) {
	p := NewBasePipeline[testData]()
	s1 := &Stage[testData]{Name: "A"}
	s2 := &Stage[testData]{Name: "B"}
	p.SetRootStage(s1).AddStage(s1).AddStage(s2)
	err := p.Validate()
	if err == nil {
		t.Error("expected ErrUnreachableStages, got nil")
	}
	if !errors.Is(err, ErrUnreachableStages) {
		t.Errorf("expected ErrUnreachableStages, got %v", err)
	}
}

func TestBasePipeline_Clear(t *testing.T) {
	p := NewBasePipeline[testData]()
	s := &Stage[testData]{Name: "A"}
	p.SetRootStage(s).AddStage(s)
	p.Clear()
	if p.GetRootStage() != nil {
		t.Errorf("expected nil root after Clear")
	}
	if len(p.GetAllStages()) != 0 {
		t.Errorf("expected no stages after Clear")
	}
}

func TestStage_AddTransitionAndGetTransitions(t *testing.T) {
	s := &Stage[testData]{Name: "A"}
	s2 := &Stage[testData]{Name: "B"}
	s.AddTransition("toB", s2, func(ctx context.Context, v testData, stage *Stage[testData]) bool { return true })
	transitions := s.GetTransitions()
	if len(transitions) != 1 {
		t.Errorf("expected 1 transition, got %d", len(transitions))
	}
	if transitions[0].Target != s2 {
		t.Errorf("transition target mismatch")
	}
}
