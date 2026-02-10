package usecase

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/smartcom/integration-platform/pkg/httpclient"
	"github.com/smartcom/integration-platform/services/middleware/internal/domain"
)

type eventProcessor struct {
	httpClient  *httpclient.Client
	targetURL   string
	eventLogger EventLogger
}

type EventLogger interface {
	InfoContext(ctx context.Context, msg string, args ...any)
	ErrorContext(ctx context.Context, msg string, args ...any)
}

func NewEventProcessor(client *httpclient.Client, targetURL string, logger EventLogger) (processor domain.EventProcessor) {
	processor = &eventProcessor{
		httpClient:  client,
		targetURL:   targetURL,
		eventLogger: logger,
	}
	return
}

func (p *eventProcessor) ProcessEvent(event domain.Event) (err error) {
	ctx := context.Background()

	if event.CorrelationID != "" {
		ctx = context.WithValue(ctx, "correlation_id", event.CorrelationID)
	}

	p.eventLogger.InfoContext(ctx, "processing event",
		"event_id", event.ID,
		"source", event.Source,
		"type", event.EventType,
		"priority", event.Priority,
	)

	var payload map[string]interface{}
	payload = p.buildPayload(event)

	headers := make(map[string]string)
	if event.CorrelationID != "" {
		headers["X-Correlation-ID"] = event.CorrelationID
	}

	var statusCode int
	var body []byte
	statusCode, body, err = p.httpClient.PostJSON(ctx, p.targetURL, payload, headers)

	if err != nil {
		p.eventLogger.ErrorContext(ctx, "failed to send event",
			"event_id", event.ID,
			"error", err.Error(),
		)
		err = fmt.Errorf("failed to send event to external endpoint: %w", err)
		return
	}

	p.eventLogger.InfoContext(ctx, "event sent successfully",
		"event_id", event.ID,
		"status_code", statusCode,
		"response_body", string(body),
	)

	return
}

func (p *eventProcessor) buildPayload(event domain.Event) (payload map[string]interface{}) {
	payload = map[string]interface{}{
		"event_id":       event.ID,
		"source":         event.Source,
		"event_type":     event.EventType,
		"priority":       p.priorityToString(event.Priority),
		"message":        event.Message,
		"timestamp":      event.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
		"correlation_id": event.CorrelationID,
	}

	if event.Metadata != nil {
		var metadataBytes []byte
		metadataBytes, _ = json.Marshal(event.Metadata)
		payload["metadata"] = string(metadataBytes)
	}

	return
}

func (p *eventProcessor) priorityToString(priority domain.Priority) (result string) {
	switch priority {
	case domain.PriorityCritical:
		result = "critical"
	case domain.PriorityHigh:
		result = "high"
	case domain.PriorityMedium:
		result = "medium"
	case domain.PriorityLow:
		result = "low"
	default:
		result = "unknown"
	}
	return
}
