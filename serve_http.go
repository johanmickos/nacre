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
		log.Printf("Could not execute template: %v\n", err)
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
	data := struct {
		FeedID  string
		Entries []string
	}{
		FeedID:  parts[1],
		Entries: []string{"dummy data from nacre"}, // TODO
	}
	if err := plaintextFeedTemplate.Execute(rw, data); err != nil {
		http.Error(rw, "An error occurred on our end", http.StatusInternalServerError)
		log.Printf("Could not execute template: %v\n", err)
		return
	}
}

func (s HttpServer) handleWebsocket(rw http.ResponseWriter, r *http.Request) {
	conn, err := s.wsUpgrader.Upgrade(rw, r, nil)
	if err != nil {
		http.Error(rw, "An error occurred on our end", http.StatusInternalServerError)
		return
	}
	defer conn.Close()
	msgType, msg, err := conn.ReadMessage()
	if err != nil {
		log.Fatal(err) // FIXME
		return
	}
	if msgType != websocket.TextMessage {
		log.Printf("error: invalid msgType %v\n", msgType)
		return
	}

	log.Printf("Handling message %s with type %v", msg, msgType)
	data := make(chan []byte) // TODO Get actual data based on feed ID in 'msg'
	go func() {
		for i := 0; i < 10; i++ {
			time.Sleep(time.Second * 1)
			data <- []byte("dummy data from nacre\n")
		}
		close(data)
	}()

	// TODO Request cancellation, signalling of closure to clients, etc.
	for update := range data {
		if err := conn.WriteMessage(websocket.BinaryMessage, update); err != nil {
			log.Fatal(err) // FIXME
			return
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
