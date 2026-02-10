package domain

import (
	"time"
)

type Priority int

const (
	PriorityLow Priority = iota
	PriorityMedium
	PriorityHigh
	PriorityCritical
)

type Event struct {
	ID            string
	Source        string
	EventType     string
	Priority      Priority
	Message       string
	Timestamp     time.Time
	CorrelationID string
	Metadata      map[string]interface{}
}

type IncomingEvent struct {
	Source    string                 `json:"source" binding:"required"`
	EventType string                 `json:"event_type" binding:"required"`
	Severity  string                 `json:"severity" binding:"required"`
	Message   string                 `json:"message" binding:"required"`
	Metadata  map[string]interface{} `json:"metadata"`
}

type EventProcessor interface {
	ProcessEvent(event Event) (err error)
}

type EventMapper interface {
	MapIncomingEvent(incoming IncomingEvent, correlationID string) (event Event, err error)
}
