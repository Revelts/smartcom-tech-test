package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/smartcom/integration-platform/pkg/correlation"
	"github.com/smartcom/integration-platform/services/middleware/internal/domain"
	"github.com/smartcom/integration-platform/services/middleware/internal/repository"
)

type EventHandler struct {
	queue  *repository.EventQueue
	mapper domain.EventMapper
	logger HandlerLogger
}

type HandlerLogger interface {
	InfoContext(ctx context.Context, msg string, args ...any)
	ErrorContext(ctx context.Context, msg string, args ...any)
}

func NewEventHandler(queue *repository.EventQueue, mapper domain.EventMapper, logger HandlerLogger) (handler *EventHandler) {
	handler = &EventHandler{
		queue:  queue,
		mapper: mapper,
		logger: logger,
	}
	return
}

func (h *EventHandler) HandleEvent(c *gin.Context) {
	var incoming domain.IncomingEvent
	err := c.ShouldBindJSON(&incoming)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "invalid request payload", "error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	var correlationID string
	correlationID, err = correlation.GenerateID()
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to generate correlation ID", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	ctx := correlation.WithID(c.Request.Context(), correlationID)

	var event domain.Event
	event, err = h.mapper.MapIncomingEvent(incoming, correlationID)
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to map event", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	err = h.queue.Enqueue(ctx, event)
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to enqueue event", "error", err.Error())
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service temporarily unavailable"})
		return
	}

	h.logger.InfoContext(ctx, "event accepted",
		"event_id", event.ID,
		"source", event.Source,
		"type", event.EventType,
	)

	c.JSON(http.StatusOK, gin.H{
		"status":         "accepted",
		"event_id":       event.ID,
		"correlation_id": correlationID,
	})
}

func (h *EventHandler) HandleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "middleware",
	})
}

func (h *EventHandler) RegisterRoutes(router *gin.Engine) {
	router.GET("/health", h.HandleHealth)
	router.POST("/integrations/events", h.HandleEvent)
}
