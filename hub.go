package nacre

import (
	"context"
)

// Hub is a central client/peer and data management layer.
type Hub struct {
	storage Storage
}

// NewHub allocates a new client/peer management hub.
func NewHub(storage Storage) *Hub {
	return &Hub{storage: storage}
}

// AddPeer is unused for now.
func (hub *Hub) AddPeer(ctx context.Context, id string) error {
	return nil
}

// RemovePeer is unused for now.
func (hub *Hub) RemovePeer(ctx context.Context, id string) error {
	return nil
}

// NumPeers is unused for now.
func (hub *Hub) NumPeers(ctx context.Context, id string) (int, error) {
	return 0, nil
}

// ClientConnected is unused for now.
func (hub *Hub) ClientConnected(ctx context.Context, id string) error {
	return nil
}

// ClientDisconnected is unused for now.
func (hub *Hub) ClientDisconnected(ctx context.Context, id string) error {
	return nil
}

// PeerListen returns a channel reader for data updates on the stream identified by 'id'.
func (hub *Hub) PeerListen(ctx context.Context, id string) (<-chan []byte, error) {
	return hub.storage.Listen(ctx, id)
}
