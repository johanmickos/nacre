package nacre

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v9"
)

const (
	maxRedisStreamLen = 100
	blockTimeout      = time.Second * 5
)

type redisStorage struct {
	client *redis.Client
}

var _ Storage = (*redisStorage)(nil)

// NewRedisStorage allocates a new Redis-backed storage implementation.
func NewRedisStorage(client *redis.Client) Storage {
	return &redisStorage{
		client: client,
	}
}

func (s *redisStorage) Push(ctx context.Context, id string, data []byte) error {
	args := &redis.XAddArgs{
		Stream: streamName(id),
		MaxLen: maxRedisStreamLen,
		Approx: true,
		// TODO Add relevant metadata to entries
		Values: map[string]any{
			"data": data,
		},
		ID: "*",
	}
	return s.client.XAdd(ctx, args).Err()
}

func (s *redisStorage) Listen(ctx context.Context, id string) (<-chan []byte, error) {
	ch := make(chan []byte)

	go func() {
		defer close(ch)

		stream := streamName(id)
		args := &redis.XReadArgs{
			Streams: []string{stream, "0"},
			Block:   -1,
		}
		streamData, err := s.client.XRead(ctx, args).Result()
		if err != nil {
			return
		}
		messages := streamData[0].Messages
		for _, msg := range messages {
			data := []byte(msg.Values["data"].(string))
			select {
			case ch <- data: // OK
			case <-ctx.Done():
				return
			}
		}
		lastSeenID := messages[len(messages)-1].ID
		for {
			args := &redis.XReadArgs{
				Streams: []string{stream, lastSeenID},
				Block:   blockTimeout,
			}
			streamData, err := s.client.XRead(ctx, args).Result()
			// TODO Query client connection state and return early when closed
			if err != nil {
				if err == redis.Nil {
					continue
				}
				return
			}
			messages := streamData[0].Messages
			lastSeenID = messages[len(messages)-1].ID
			for _, msg := range messages {
				data := []byte(msg.Values["data"].(string))
				select {
				case ch <- data: // OK
				case <-ctx.Done():
					return
				}
			}

		}
	}()

	return ch, nil
}

func (s *redisStorage) GetAll(ctx context.Context, id string) ([][]byte, error) {
	stream := streamName(id)
	args := &redis.XReadArgs{
		Streams: []string{stream, "0"},
		Block:   -1,
	}
	streamData, err := s.client.XRead(ctx, args).Result()
	if err != nil {
		return nil, err
	}
	results := make([][]byte, len(streamData[0].Messages))
	for i, msg := range streamData[0].Messages {
		data := msg.Values["data"]
		results[i] = []byte(data.(string))
	}
	return results, nil
}

func streamName(id string) string { return fmt.Sprintf("nacre:feed:%s", id) }
