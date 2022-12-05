package nacre

import (
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/jarlopez/nacre/ws"
	"golang.org/x/sync/errgroup"
)

var (
	homeTemplate          = template.Must(template.ParseFiles("./templates/home.gohtml"))
	liveFeedTemplate      = template.Must(template.ParseFiles("./templates/liveFeed.gohtml"))
	plaintextFeedTemplate = template.Must(template.ParseFiles("./templates/plaintextFeed.gohtml"))
)

// HTTPServer handles nacre's HTTP requests and websocket upgrades.
type HTTPServer struct {
	quit        chan struct{}
	hub         Hub
	rateLimiter RateLimiter
	mux         *http.ServeMux
	wsUpgrader  websocket.Upgrader

	address string
	bufsize int
}

// NewHTTPServer allocates a HTTP server for serving nacre's HTTP traffic.
func NewHTTPServer(address string, hub Hub, rateLimiter RateLimiter) *HTTPServer {
	server := &HTTPServer{
		quit:        make(chan struct{}),
		hub:         hub,
		rateLimiter: rateLimiter,
		mux:         http.NewServeMux(),
		wsUpgrader: websocket.Upgrader{
			WriteBufferSize: 1024,
			ReadBufferSize:  1024,
		},

		address: address,
		bufsize: 1024,
	}

	middleware := func(next http.Handler) http.Handler {
		return withRecovery(withRequestID(next))
	}

	server.mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	server.mux.Handle("/feed/", middleware(http.HandlerFunc(server.handleFeed)))
	server.mux.Handle("/plaintext/", middleware(http.HandlerFunc(server.handlePlaintext)))
	server.mux.Handle("/websocket", middleware(http.HandlerFunc(server.handleWebsocket)))
	server.mux.Handle("/", middleware(http.HandlerFunc(handleHome)))
	return server
}

// Serve HTTP traffic on the configured address.
func (s *HTTPServer) Serve() error {
	return http.ListenAndServe(s.address, s.mux)
}

func handleHome(rw http.ResponseWriter, r *http.Request) {
	if err := homeTemplate.Execute(rw, nil); err != nil {
		http.Error(rw, "An error occurred on our end", http.StatusInternalServerError)
		return
	}
}

func (s HTTPServer) handleFeed(rw http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")[1:]
	if len(parts) != 2 {
		// ["feed", "${feedID}"]
		http.Error(rw, "Bad request", http.StatusBadRequest)
		return
	}
	feedID := parts[1]
	if parts[0] != "feed" || len(feedID) == 0 {
		http.Error(rw, "Bad request", http.StatusBadRequest)
		return
	}
	if exists, err := s.hub.Exists(r.Context(), feedID); err != nil {
		http.Error(rw, "An error occurred on our end", http.StatusInternalServerError)
		return
	} else if !exists {
		http.Error(rw, "Feed not found", http.StatusNotFound)
		return
	}
	data := struct {
		FeedID       string
		PlaintextURL template.URL
		HomeURL      template.URL
	}{
		FeedID:       feedID,
		PlaintextURL: template.URL(plaintextURL(s.address, parts[1])),
		HomeURL:      template.URL(homeURL(s.address)),
	}
	if err := liveFeedTemplate.Execute(rw, data); err != nil {
		http.Error(rw, "An error occurred on our end", http.StatusInternalServerError)
		return
	}
}

func (s HTTPServer) handlePlaintext(rw http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")[1:]
	if len(parts) != 2 {
		// ["plaintext", "${feedID}"]
		http.Error(rw, "Bad request", http.StatusBadRequest)
		return
	}
	if parts[0] != "plaintext" || len(parts[1]) == 0 {
		http.Error(rw, "Bad request", http.StatusBadRequest)
		return
	}
	id := parts[1]
	entries, err := s.hub.GetAll(r.Context(), id)
	if err != nil {
		http.Error(rw, "An error occurred on our end", http.StatusInternalServerError)
		return
	}
	data := struct {
		FeedID  string
		Entries []string
	}{
		FeedID:  id,
		Entries: make([]string, len(entries)),
	}
	for i, entry := range entries {
		data.Entries[i] = string(entry)
	}
	rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if err := plaintextFeedTemplate.Execute(rw, data); err != nil {
		http.Error(rw, "An error occurred on our end", http.StatusInternalServerError)
		return
	}
}

func (s HTTPServer) handleWebsocket(rw http.ResponseWriter, r *http.Request) {
	conn, err := s.wsUpgrader.Upgrade(rw, r, nil)
	if err != nil {
		http.Error(rw, "An error occurred on our end", http.StatusInternalServerError)
		return
	}
	msgType, msg, err := conn.ReadMessage()
	if err != nil {
		return
	}
	if msgType != websocket.TextMessage {
		return
	}
	ctx := r.Context()
	feedID := string(msg)
	if exists, err := s.hub.Exists(ctx, feedID); err != nil {
		_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "Internal error"))
		return
	} else if !exists {
		_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(ws.CloseNotFound, "Feed not found"))
		return
	}
	if canAdd := s.rateLimiter.TryAddPeer(ctx, feedID); !canAdd {
		_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(ws.CloseTooManyPeers, "Too many concurrent peers for this feed"))
		return
	}
	defer s.rateLimiter.RemovePeer(ctx, feedID)

	peer := &Peer{
		conn: conn,
		hub:  s.hub,
	}
	g := new(errgroup.Group)
	g.Go(func() error { return peer.readLoop(ctx) })
	g.Go(func() error { return peer.writeLoop(ctx, feedID) })
	if err := g.Wait(); err != nil {
		log.Printf("Internal error: %v", err)
	}
}

func withRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		defer func() {
			err := recover()
			if err == nil {
				return
			}
			log.Printf("panic: %v", err)
			http.Error(rw, "An error occurred on our end", http.StatusInternalServerError)
		}()
		next.ServeHTTP(rw, r)
	})
}

func withRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rid := NewUUID()
		header := rw.Header()
		header["X-Request-Id"] = []string{rid}
		// TODO Inject logger w/ request ID into context
		next.ServeHTTP(rw, r)
	})
}
