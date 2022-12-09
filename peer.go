package nacre

import (
	"context"
	"time"

	"github.com/gorilla/websocket"
)

const (
	maxReadBytes = 256

	writeDeadline = time.Second * 10
	pongDeadline  = time.Second * 8
	pingPeriod    = time.Second * 5
)

// Peer represents a connected websocket peer and is responsible for
// driving the websocket read/write loops and peer connection maangement.
type Peer struct {
	conn *websocket.Conn
	hub  Hub
}

func (peer *Peer) readLoop(ctx context.Context) error {
	defer peer.conn.Close()
	peer.conn.SetReadLimit(maxReadBytes)
	peer.conn.SetReadDeadline(time.Now().Add(pongDeadline))
	peer.conn.SetPongHandler(func(string) error {
		peer.conn.SetReadDeadline(time.Now().Add(pongDeadline))
		return nil
	})
	for {
		select {
		case <-ctx.Done():
			return nil
		default: // OK
		}
		_, _, err := peer.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(
				err,
				websocket.CloseNormalClosure,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure,
			) {
				return err
			}
			return nil
		}
	}
}

// writeLoop pushes feed data to the connected peer.
func (peer *Peer) writeLoop(ctx context.Context, id string) error {
	data, err := peer.hub.Listen(ctx, id)
	if err != nil {
		return err
	}
	ticker := time.NewTicker(pingPeriod)
	for {
		select {
		case <-ctx.Done():
			return peer.conn.WriteMessage(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseGoingAway, "Request context completed"),
			)
		case message, ok := <-data:
			peer.conn.SetWriteDeadline(time.Now().Add(writeDeadline))
			if !ok {
				return peer.conn.WriteMessage(
					websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Data channel closed"),
				)
			}
			if err := peer.conn.WriteMessage(websocket.BinaryMessage, message); err != nil {
				return err
			}
		case <-ticker.C:
			peer.conn.SetWriteDeadline(time.Now().Add(writeDeadline))
			if err := peer.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return err
			}
		}
	}
}
