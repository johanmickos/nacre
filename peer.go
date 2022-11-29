package nacre

import (
	"context"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	maxReadBytes = 256

	writeDeadline = time.Second * 10
	pongDeadline  = time.Second * 60
	pingPeriod    = time.Second * 45
)

// Peer represents a connected websocket peer and is responsible for
// driving the websocket read/write loops and peer connection maangement.
type Peer struct {
	conn *websocket.Conn
	hub  *Hub
}

func (peer *Peer) readLoop(ctx context.Context) error {
	defer func() {
		// TODO Unregister from peer management
		peer.conn.Close()
		log.Printf("Connection closed")
	}()
	peer.conn.SetReadLimit(maxReadBytes)
	peer.conn.SetReadDeadline(time.Now().Add(pongDeadline))
	peer.conn.SetPongHandler(func(string) error {
		peer.conn.SetReadDeadline(time.Now().Add(pongDeadline))
		return nil
	})
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// OK
		}
		_, _, err := peer.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			return err
		}
	}
}

// writeLoop pushes stream data to the connected peer.
func (peer *Peer) writeLoop(ctx context.Context, id string) error {
	data, err := peer.hub.PeerListen(ctx, id)
	if err != nil {
		return err
	}
	ticker := time.NewTicker(pingPeriod)
	for {
		select {
		case <-ctx.Done():
			_ = peer.conn.WriteMessage(websocket.CloseMessage, nil)
			return ctx.Err()
		case message, ok := <-data:
			peer.conn.SetWriteDeadline(time.Now().Add(writeDeadline))
			if !ok {
				_ = peer.conn.WriteMessage(websocket.CloseMessage, nil)
				return nil
			}
			if err := peer.conn.WriteMessage(websocket.BinaryMessage, message); err != nil {
				return nil
			}
		case <-ticker.C:
			peer.conn.SetWriteDeadline(time.Now().Add(writeDeadline))
			if err := peer.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return err
			}
		}
	}
}
