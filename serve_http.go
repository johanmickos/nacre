package nacre

import (
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

var (
	liveFeedTemplate      = template.Must(template.ParseFiles("./templates/liveFeed.gohtml"))
	plaintextFeedTemplate = template.Must(template.ParseFiles("./templates/plaintextFeed.gohtml"))
)

type HttpServer struct {
	quit       chan struct{}
	storage    Storage
	mux        *http.ServeMux
	wsUpgrader websocket.Upgrader

	address string
	bufsize int
}

func NewHttpServer(address string, storage Storage) *HttpServer {
	server := &HttpServer{
		quit:    make(chan struct{}),
		storage: storage,
		mux:     http.NewServeMux(),
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
	server.mux.Handle("/", middleware(http.HandlerFunc(handleInfo)))
	return server
}

func (s *HttpServer) Serve() error {
	return http.ListenAndServe(s.address, s.mux)
}

func handleInfo(rw http.ResponseWriter, r *http.Request) {
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte("OK"))
}

func (s HttpServer) handleFeed(rw http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")[1:]
	if len(parts) != 2 {
		// ["feed", "${feedID}"]
		http.Error(rw, "Bad request", http.StatusBadRequest)
		return
	}
	if parts[0] != "feed" || len(parts[1]) == 0 {
		http.Error(rw, "Bad request", http.StatusBadRequest)
		return
	}
	data := struct {
		FeedID string
	}{
		FeedID: parts[1],
	}
	if err := liveFeedTemplate.Execute(rw, data); err != nil {
		http.Error(rw, "An error occurred on our end", http.StatusInternalServerError)
		log.Printf("Could not execute template: %v", err)
		return
	}
}

func (s HttpServer) handlePlaintext(rw http.ResponseWriter, r *http.Request) {
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
	entries, err := s.storage.GetAll(r.Context(), id)
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
	if err := plaintextFeedTemplate.Execute(rw, data); err != nil {
		http.Error(rw, "An error occurred on our end", http.StatusInternalServerError)
		log.Printf("Could not execute template: %v", err)
		return
	}
}

func (s HttpServer) handleWebsocket(rw http.ResponseWriter, r *http.Request) {
	conn, err := s.wsUpgrader.Upgrade(rw, r, nil)
	if err != nil {
		http.Error(rw, "An error occurred on our end", http.StatusInternalServerError)
		return
	}
	msgType, msg, err := conn.ReadMessage()
	if err != nil {
		log.Fatal(err) // FIXME
		return
	}
	if msgType != websocket.TextMessage {
		log.Printf("error: invalid msgType %v", msgType)
		return
	}

	log.Printf("Handling message %s with type %v", msg, msgType)

	data, err := s.storage.Listen(r.Context(), string(msg))
	if err != nil {
		log.Printf("error: storage.Listen %v", err)
		return
	}

	// Read loop
	go func() {
		defer func() {
			// TODO Unregister from peer management
			conn.Close()
			log.Printf("Connection closed")
		}()
		conn.SetReadLimit(256)
		conn.SetReadDeadline(time.Now().Add(time.Second * 60))
		conn.SetPongHandler(func(string) error {
			conn.SetReadDeadline(time.Now().Add(time.Second * 60))
			return nil
		})
		for {
			select {
			case <-r.Context().Done():
				return
			default:
				// OK
			}
			_, _, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("error: %v", err)
				}
				return
			}
		}
	}()

	// Write loop
	ticker := time.NewTicker(time.Second * 50)
	for {
		select {
		case <-r.Context().Done():
			_ = conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		case message, ok := <-data:
			conn.SetWriteDeadline(time.Now().Add(time.Second * 10))
			if !ok {
				// Data channel was closed
				_ = conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := conn.WriteMessage(websocket.BinaryMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(time.Second * 10))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
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
