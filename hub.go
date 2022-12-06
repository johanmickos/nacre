package nacre

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v9"
)

// Hub is a central client/peer and streamed data management layer.
type Hub interface {
	Exists(ctx context.Context, id string) (bool, error)
	Push(ctx context.Context, id string, data []byte) error
	Listen(ctx context.Context, id string) (<-chan []byte, error)
	GetAll(ctx context.Context, id string) ([][]byte, error)

	AddPeer(ctx context.Context, id string) error
	RemovePeer(ctx context.Context, id string) error
	ClientState(ctx context.Context, id string) (ClientState, error)
	ClientConnected(ctx context.Context, id string) error
	ClientDisconnected(ctx context.Context, id string) error
}

// TODO Support these in external configuration file with defaults
const (
	maxRedisStreamLen            = 1000
	maxStreamPersistenceDuration = time.Hour * 24
	blockTimeout                 = time.Second * 5
	clientConnectedDuration      = time.Second * 15
)

// ClientState indicates whether the data-streaming client is still connected.
type ClientState string

// Possible client states.
const (
	ClientStateConnected    ClientState = "CONNECTED"
	ClientStateDisconnected ClientState = "DISCONNECTED"
	ClientStateUnknown      ClientState = "UNKNOWN"
)

type redisHub struct {
	client *redis.Client
}

var _ Hub = (*redisHub)(nil)

// NewRedisHub allocates a new Redis-backed hub implementation.
func NewRedisHub(client *redis.Client) Hub {
	return &redisHub{
		client: client,
	}
}

func (hub *redisHub) Exists(ctx context.Context, id string) (bool, error) {
	exists, err := hub.client.Exists(ctx, streamName(id)).Result()
	return exists > 0, err
}

func (hub *redisHub) Push(ctx context.Context, id string, data []byte) error {
	pipe := hub.client.Pipeline()
	stream := streamName(id)
	addCmd := pipe.XAdd(ctx, &redis.XAddArgs{
		Stream: stream,
		MaxLen: maxRedisStreamLen,
		Approx: true,
		// TODO Add relevant metadata to entries
		Values: map[string]any{
			"data": data,
		},
		ID: "*",
	})
	// Refresh expiration for this stream
	// FIXME: Use ExpireGT if Redis v7 and higher
	pipe.Expire(ctx, stream, maxStreamPersistenceDuration)
	if _, err := pipe.Exec(ctx); err != nil {
		return err
	}
	return addCmd.Err()
}

func (hub *redisHub) Listen(ctx context.Context, id string) (<-chan []byte, error) {
	ch := make(chan []byte)

	go func() {
		defer close(ch)

		stream := streamName(id)
		args := &redis.XReadArgs{
			Streams: []string{stream, "0"},
			Block:   -1,
		}
		streamData, err := hub.client.XRead(ctx, args).Result()
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
			state, err := hub.ClientState(ctx, id)
			if err != nil {
				return
			}
			if state == ClientStateDisconnected {
				return
			}

			args := &redis.XReadArgs{
				Streams: []string{stream, lastSeenID},
				Block:   blockTimeout,
			}
			streamData, err := hub.client.XRead(ctx, args).Result()
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

func (hub *redisHub) GetAll(ctx context.Context, id string) ([][]byte, error) {
	stream := streamName(id)
	args := &redis.XReadArgs{
		Streams: []string{stream, "0"},
		Block:   -1,
	}
	streamData, err := hub.client.XRead(ctx, args).Result()
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

func (hub *redisHub) AddPeer(ctx context.Context, id string) error    { return nil }
func (hub *redisHub) RemovePeer(ctx context.Context, id string) error { return nil }

func (hub *redisHub) ClientState(ctx context.Context, id string) (ClientState, error) {
	state, err := hub.client.Get(ctx, clientKey(id)).Result()
	if err != nil {
		if err == redis.Nil {
			return ClientStateDisconnected, nil
		}
		return ClientStateUnknown, err
	}
	return ClientState(state), nil
}

func (hub *redisHub) ClientConnected(ctx context.Context, id string) error {
	return hub.client.Set(ctx, clientKey(id), string(ClientStateConnected), clientConnectedDuration).Err()
}

func (hub *redisHub) ClientDisconnected(ctx context.Context, id string) error {
	err := hub.client.Del(ctx, clientKey(id)).Err()
	if err != redis.Nil {
		return err
	}
	return nil
}

func streamName(id string) string { return fmt.Sprintf("nacre:feed:%s", id) }
func clientKey(id string) string  { return fmt.Sprintf("nacre:client:%s", id) }
