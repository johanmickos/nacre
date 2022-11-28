package nacre

import (
	"context"
	"log"
)

// Storage abstraction for handling streamed data.
type Storage interface {
	Push(ctx context.Context, id string, data []byte) error
}

type loggingStorage struct{}

// NewLoggingStorage returns a Storage implementation that prints
// diagnostic information to STDOUT.
func NewLoggingStorage() Storage {
	return loggingStorage{}
}

// Push data by printing it to the standard logger.
func (s loggingStorage) Push(ctx context.Context, id string, data []byte) error {
	log.Printf("[%s] Pushed %d bytes of data", id, len(data))
	return nil
}
