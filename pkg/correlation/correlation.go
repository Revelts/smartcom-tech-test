package correlation

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

type contextKey string

const (
	CorrelationIDKey contextKey = "correlation_id"
)

func GenerateID() (id string, err error) {
	bytes := make([]byte, 16)
	var n int
	n, err = rand.Read(bytes)
	if err != nil {
		err = fmt.Errorf("failed to generate correlation ID: %w", err)
		return
	}
	if n != 16 {
		err = fmt.Errorf("insufficient random bytes")
		return
	}

	id = hex.EncodeToString(bytes)
	return
}

func WithID(ctx context.Context, correlationID string) (newCtx context.Context) {
	newCtx = context.WithValue(ctx, CorrelationIDKey, correlationID)
	return
}

func FromContext(ctx context.Context) (id string) {
	value := ctx.Value(CorrelationIDKey)
	if correlationID, ok := value.(string); ok {
		id = correlationID
		return
	}
	id = ""
	return
}
