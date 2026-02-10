package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AlertHandler struct {
	logger AlertLogger
}

type AlertLogger interface {
	InfoContext(ctx context.Context, msg string, args ...any)
	ErrorContext(ctx context.Context, msg string, args ...any)
}

func NewAlertHandler(logger AlertLogger) (handler *AlertHandler) {
	handler = &AlertHandler{
		logger: logger,
	}
	return
}

func (h *AlertHandler) HandleAlert(c *gin.Context) {
	var payload map[string]interface{}
	err := c.ShouldBindJSON(&payload)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "invalid payload", "error", err.Error())
		c.JSON(http.StatusOK, gin.H{"status": "received"})
		return
	}

	correlationID := c.GetHeader("X-Correlation-ID")
	if correlationID != "" {
		ctx := context.WithValue(c.Request.Context(), "correlation_id", correlationID)
		h.logger.InfoContext(ctx, "alert received", "payload", payload)
	} else {
		h.logger.InfoContext(c.Request.Context(), "alert received", "payload", payload)
	}

	c.JSON(http.StatusOK, gin.H{"status": "received"})
}

func (h *AlertHandler) RegisterRoutes(router *gin.Engine) {
	router.POST("/external/alerts", h.HandleAlert)
}
