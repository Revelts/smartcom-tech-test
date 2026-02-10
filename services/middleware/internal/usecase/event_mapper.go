package usecase

import (
	"fmt"
	"strings"
	"time"

	"github.com/smartcom/integration-platform/services/middleware/internal/domain"
)

type eventMapper struct {
	idGenerator IDGenerator
}

type IDGenerator interface {
	Generate() (id string, err error)
}

func NewEventMapper(idGen IDGenerator) (mapper domain.EventMapper) {
	mapper = &eventMapper{
		idGenerator: idGen,
	}
	return
}

func (m *eventMapper) MapIncomingEvent(incoming domain.IncomingEvent, correlationID string) (event domain.Event, err error) {
	var id string
	id, err = m.idGenerator.Generate()
	if err != nil {
		err = fmt.Errorf("failed to generate event ID: %w", err)
		return
	}

	var priority domain.Priority
	priority = m.mapSeverityToPriority(incoming.Severity)

	event = domain.Event{
		ID:            id,
		Source:        incoming.Source,
		EventType:     incoming.EventType,
		Priority:      priority,
		Message:       incoming.Message,
		Timestamp:     time.Now().UTC(),
		CorrelationID: correlationID,
		Metadata:      incoming.Metadata,
	}

	return
}

func (m *eventMapper) mapSeverityToPriority(severity string) (priority domain.Priority) {
	normalized := strings.ToLower(strings.TrimSpace(severity))

	switch normalized {
	case "critical", "fatal", "emergency":
		priority = domain.PriorityCritical
	case "high", "error":
		priority = domain.PriorityHigh
	case "medium", "warning", "warn":
		priority = domain.PriorityMedium
	default:
		priority = domain.PriorityLow
	}

	return
}
