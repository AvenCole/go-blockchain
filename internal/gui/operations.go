package gui

import (
	"context"
	"time"
)

const networkOperationEventName = "network:operation"

type networkOperationTracker struct {
	service   *Service
	operation string
	startedAt time.Time
	totalSteps int
	currentStep int
}

func newNetworkOperationTracker(service *Service, operation string, totalSteps int) *networkOperationTracker {
	return &networkOperationTracker{
		service:     service,
		operation:   operation,
		startedAt:   time.Now().UTC(),
		totalSteps:  totalSteps,
		currentStep: 0,
	}
}

func (t *networkOperationTracker) start(phase, message string) {
	t.emit(NetworkOperationProgress{
		Operation:   t.operation,
		Status:      "started",
		Phase:       phase,
		Message:     message,
		CurrentStep: t.currentStep,
		TotalSteps:  t.totalSteps,
		StartedAt:   t.startedAt.Format(time.RFC3339),
		ElapsedMS:   0,
	})
}

func (t *networkOperationTracker) step(phase, message string) {
	t.currentStep++
	t.emit(NetworkOperationProgress{
		Operation:   t.operation,
		Status:      "progress",
		Phase:       phase,
		Message:     message,
		CurrentStep: t.currentStep,
		TotalSteps:  t.totalSteps,
		StartedAt:   t.startedAt.Format(time.RFC3339),
		ElapsedMS:   time.Since(t.startedAt).Milliseconds(),
	})
}

func (t *networkOperationTracker) complete(phase, message, summary string) {
	finishedAt := time.Now().UTC()
	t.emit(NetworkOperationProgress{
		Operation:   t.operation,
		Status:      "completed",
		Phase:       phase,
		Message:     message,
		CurrentStep: t.totalSteps,
		TotalSteps:  t.totalSteps,
		StartedAt:   t.startedAt.Format(time.RFC3339),
		FinishedAt:  finishedAt.Format(time.RFC3339),
		ElapsedMS:   finishedAt.Sub(t.startedAt).Milliseconds(),
		Summary:     summary,
	})
}

func (t *networkOperationTracker) fail(phase string, err error) {
	finishedAt := time.Now().UTC()
	errorText := ""
	if err != nil {
		errorText = err.Error()
	}
	t.emit(NetworkOperationProgress{
		Operation:   t.operation,
		Status:      "failed",
		Phase:       phase,
		Message:     "流程执行失败",
		CurrentStep: t.currentStep,
		TotalSteps:  t.totalSteps,
		StartedAt:   t.startedAt.Format(time.RFC3339),
		FinishedAt:  finishedAt.Format(time.RFC3339),
		ElapsedMS:   finishedAt.Sub(t.startedAt).Milliseconds(),
		Error:       errorText,
	})
}

func (t *networkOperationTracker) emit(progress NetworkOperationProgress) {
	if t.service == nil {
		return
	}
	t.service.emitRuntimeEvent(networkOperationEventName, progress)
}

func (s *Service) emitRuntimeEvent(eventName string, payload interface{}) {
	if s == nil || s.emit == nil {
		return
	}
	s.emit(s.ctx, eventName, payload)
}

func (s *Service) setRuntimeEmitter(emitter func(ctx context.Context, eventName string, optionalData ...interface{})) {
	s.emit = emitter
}
