package nacre

import (
	"context"
	"log"
	"time"
)

// Storage abstraction for handling streamed data.
type Storage interface {
	Push(ctx context.Context, id string, data []byte) error
	Listen(ctx context.Context, id string) (<-chan []byte, error)
	GetAll(ctx context.Context, id string) ([][]byte, error)
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

// Listen returns a channel that is populated with dummy data by a separate goroutine.
func (s loggingStorage) Listen(ctx context.Context, id string) (<-chan []byte, error) {
	ch := make(chan []byte)
	go func() {
		defer close(ch)
		defer func() { log.Printf("Ending listen loop for dummy data\n") }()
		for {
			select {
			case <-ctx.Done():
				return
			default: // OK
			}
			time.Sleep(time.Second * 1)
			select {
			case <-ctx.Done():
				return
			case ch <- []byte("dummy data from nacre\n"): // OK
			case <-time.After(time.Second * 10):
				log.Printf("Timeout while writing data in loggingStorage")
				return
			default: // OK
			}
		}
	}()
	return ch, nil
}

// GetAll returns a slice of dummy data.
func (s loggingStorage) GetAll(ctx context.Context, id string) ([][]byte, error) {
	return [][]byte{
		[]byte("dummy data from nacre\n"),
		[]byte("dummy data from nacre\n"),
		[]byte("dummy data from nacre\n"),
		[]byte("dummy data from nacre\n"),
	}, nil
}
