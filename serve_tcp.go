package nacre

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
)

type TcpServer struct {
	listener net.Listener
	quit     chan struct{}
	wg       sync.WaitGroup
	storage  Storage

	address string
	bufsize int
}

// NewTCPServer returns a stoppable TCP server listening on
// the provided adderss.
func NewTcpServer(address string, storage Storage) (*TcpServer, error) {
	server := &TcpServer{
		quit:    make(chan struct{}),
		storage: storage,
		wg:      sync.WaitGroup{},
		address: address,
		bufsize: 1024,
	}
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}
	server.listener = listener
	return server, nil
}

// Serve incoming TCP connections and handle them in new goroutines.
func (s *TcpServer) Serve(ctx context.Context) {
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
// the Storage implementation.
func (s *TcpServer) handle(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	sid := NewUUID()
	msg := fmt.Sprintf("Connected to nacre\n%s/feed/%s\n", s.address, sid)
	n, err := conn.Write([]byte(msg))
	if err != nil {
		log.Printf("error: conn.Write: %s\n", err.Error())
		return
	}
	if n != len(msg) {
		log.Printf("error: conn.Write: wrote %d/%d bytes\n", n, len(msg))
		return
	}
	for {
		select {
		case <-ctx.Done():
			return
		case <-s.quit:
			return
		default:
			// Continue serving client
		}
		buf := make([]byte, s.bufsize) // NOTE: Could consider buffer pool to limit memory usage
		nbytes, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				// TODO Clean up, signal to connected HTTP clients
				return
			}
			return
		}
		if nbytes == 0 {
			return
		}
		if err := s.storage.Push(ctx, sid, buf[:nbytes]); err != nil {
			return
		}
	}
}
