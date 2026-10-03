package core

import (
	"encoding/json"
	"io"
	"time"
)

type StepStatus string

const (
	StepRunning StepStatus = "running"
	StepOK      StepStatus = "ok"
	StepFail    StepStatus = "fail"
)

type ProgressEvent struct {
	Step       int        `json:"step"`
	Run        int        `json:"run,omitempty"`
	Action     string     `json:"action"`
	Selector   string     `json:"selector,omitempty"`
	Status     StepStatus `json:"status"`
	Message    string     `json:"message,omitempty"`
	Timestamp  string     `json:"ts"`
	DurationMs int64      `json:"duration_ms,omitempty"`
}

func NewProgressEvent(step, run int, action, selector string, status StepStatus) ProgressEvent {
	return ProgressEvent{
		Step:      step,
		Run:       run,
		Action:    action,
		Selector:  selector,
		Status:    status,
		Timestamp: time.Now().Format(time.RFC3339),
	}
}

// Emitter receives progress events during a run.
type Emitter interface {
	Emit(ProgressEvent)
}

// JSONLinesEmitter writes one JSON object per line to w.
type JSONLinesEmitter struct {
	w io.Writer
}

func NewJSONLinesEmitter(w io.Writer) *JSONLinesEmitter {
	return &JSONLinesEmitter{w: w}
}

func (e *JSONLinesEmitter) Emit(ev ProgressEvent) {
	b, _ := json.Marshal(ev)
	_, _ = e.w.Write(append(b, '\n'))
}

// NoopEmitter discards all events.
type NoopEmitter struct{}

func (NoopEmitter) Emit(ProgressEvent) {}

// CollectingEmitter stores all events and optionally forwards to a wrapped emitter.
type CollectingEmitter struct {
	Events  []ProgressEvent
	wrapped Emitter
}

func NewCollectingEmitter(wrapped Emitter) *CollectingEmitter {
	return &CollectingEmitter{wrapped: wrapped}
}

func (c *CollectingEmitter) Emit(ev ProgressEvent) {
	c.Events = append(c.Events, ev)
	if c.wrapped != nil {
		c.wrapped.Emit(ev)
	}
}

func (c *CollectingEmitter) Reset() {
	c.Events = c.Events[:0]
}
