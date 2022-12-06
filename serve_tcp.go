package nacre

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

// TODO Support these in external configuration file with defaults
const (
	clientConnectedHeartbeat    = time.Second * 2
	clientConnectionReadTimeout = time.Minute * 1
)

// TCPServer handles nacre's TCP clients and their data streams.
type TCPServer struct {
	listener    net.Listener
	quit        chan struct{}
	wg          sync.WaitGroup
	hub         Hub
	rateLimiter RateLimiter

	address     string
	httpAddress string
	bufsize     int
}

// NewTCPServer returns a stoppable TCP server listening on
// the provided adderss.
func NewTCPServer(address string, httpAddress string, hub Hub, rateLimiter RateLimiter) (*TCPServer, error) {
	server := &TCPServer{
		quit:        make(chan struct{}),
		hub:         hub,
		rateLimiter: rateLimiter,
		wg:          sync.WaitGroup{},
		address:     address,
		httpAddress: httpAddress,
		bufsize:     1024,
	}
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}
	server.listener = listener
	return server, nil
}

// Serve incoming TCP connections and handle them in new goroutines.
func (s *TCPServer) Serve(ctx context.Context) {
	s.wg.Add(1)
	defer s.wg.Done()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Printf("error: listener.Accept: %s\n", err.Error())
			continue
		}
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			s.handle(ctx, conn)
		}()
	}
}

// handle the connection by reading incoming bytes and pushing them to
// the Hub implementation.
func (s *TCPServer) handle(ctx context.Context, conn net.Conn) {
	defer conn.Close()

	clientIP, _, err := net.SplitHostPort(conn.RemoteAddr().String())
	if err != nil {
		conn.Write([]byte("nacre: internal error"))
		return
	}
	if canAdd := s.rateLimiter.TryAddClient(ctx, clientIP); !canAdd {
		conn.Write([]byte("nacre: too many concurrent feeds from your IP\n"))
		return
	}
	defer s.rateLimiter.RemoveClient(ctx, clientIP)

	sid := NewUUID()
	msg := fmt.Sprintf("Connected to nacre\n%s\n", liveFeedURL(s.httpAddress, sid))
	n, err := conn.Write([]byte(msg))
	if err != nil {
		log.Printf("error: conn.Write: %s\n", err.Error())
		return
	}
	if n != len(msg) {
		log.Printf("error: conn.Write: wrote %d/%d bytes\n", n, len(msg))
		return
	}

	heartbeatCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	defer s.hub.ClientDisconnected(ctx, sid)
	go func() {
		heartbeat := time.NewTicker(clientConnectedHeartbeat)
		_ = s.hub.ClientConnected(heartbeatCtx, sid)
		for {
			select {
			case <-heartbeatCtx.Done():
				return
			case <-heartbeat.C:
				_ = s.hub.ClientConnected(ctx, sid)
			}
		}
	}()

	buf := make([]byte, s.bufsize) // NOTE: Could consider buffer pool to limit memory usage
	for {
		select {
		case <-ctx.Done():
			return
		case <-s.quit:
			return
		default: // Continue serving client
		}
		// TODO Bandwidth quota per IP
		nbytes, err := conn.Read(buf)
		if err != nil {
			return
		}
		if nbytes == 0 {
			return
		}
		if err := s.hub.Push(ctx, sid, buf[0:nbytes]); err != nil {
			log.Printf("Failed to push data: %s", err)
			return
		}
	}
}

// TODO Move to domain name & HTTP/HTTPS-aware config struct
func liveFeedURL(baseURL string, id string) string {
	return fmt.Sprintf("http://%s/feed/%s", baseURL, id)
}

func plaintextURL(baseURL string, id string) string {
	return fmt.Sprintf("http://%s/plaintext/%s", baseURL, id)
}

func homeURL(baseURL string) string {
	return fmt.Sprintf("http://%s", baseURL)
}
