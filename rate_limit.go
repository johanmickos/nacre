package nacre

import (
	"context"
	"sync"
	"time"
)

// Nacre natively supports two in-memory rate limiting strategies:
// - # of concurrent TCP connections by IP
// - # of concurrent websocket sessions by feed ID
//
// If horiziontally scaled, these strategies are not enough to guarantee per-IP or per-feed
// limits, as the in-memory implementations are only aware of the clients/peers connected
// to the other instances
//
// Instead, we can either
// 1) deploy the Nacre instances behind a load balancer (like HAProxy)
//     that has built-in support for these strategies, or
// 2) extend in-application support for distributed rate limiting (e.g. via database-backed expiring semaphores)

type empty struct{}
type semaphore chan empty

// RateLimiter defines functions for keeping track of ongoing peer/client connections
// and limiting incoming request rates.
type RateLimiter interface {
	// Stop the rate limiter.
	Stop()
	// TryAddClient to the rate limiter, or return false if it can currently not be done.
	TryAddClient(ctx context.Context, ip string) bool
	// RemoveClient from the rate limiter.
	RemoveClient(ctx context.Context, ip string) bool
	// TryAddpeer to the rate limiter, or return false if it can currently not be done.
	TryAddPeer(ctx context.Context, id string) bool
	// RemovePeer from the rate limiter.
	RemovePeer(ctx context.Context, id string) bool
}

type inMemoryRateLimiter struct {
	mu      sync.Mutex
	clients map[string]semaphore
	peers   map[string]semaphore

	maxClientsPerIP   int
	maxPeersPerFeedID int

	numRemovedClients   int
	numRemovedPeers     int
	gcMaxRemovedClients int
	gcMaxRemovedPeers   int

	gcPeriod time.Duration
	quit     chan empty
}

var _ RateLimiter = (*inMemoryRateLimiter)(nil)

// NewInMemoryRateLimiter returns a memory-backed rate limiter for managing
// incoming peer/client requests.
//
// Note that this implementation does not manage distributed state.
// It only tracks local calls to the RateLimiter interface and can therefore
// not accurately rate limit in a horizontally-scaled deployment.
func NewInMemoryRateLimiter() RateLimiter {
	r := &inMemoryRateLimiter{
		mu:                  sync.Mutex{},
		clients:             make(map[string]semaphore),
		peers:               make(map[string]semaphore),
		maxClientsPerIP:     5,
		maxPeersPerFeedID:   3,
		numRemovedClients:   0,
		numRemovedPeers:     0,
		gcMaxRemovedClients: 1_000_000,
		gcMaxRemovedPeers:   1_000_000,
		gcPeriod:            time.Second * 30,
		quit:                make(chan empty),
	}
	go r.garbageCollectLoop(context.Background())
	return r
}

func (r *inMemoryRateLimiter) Stop() { r.quit <- empty{} }

func (r *inMemoryRateLimiter) TryAddClient(ctx context.Context, ip string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if ch, ok := r.clients[ip]; !ok {
		r.clients[ip] = make(semaphore, r.maxClientsPerIP)
		r.clients[ip] <- empty{}
		return true
	} else if len(ch) < r.maxClientsPerIP {
		ch <- empty{}
		return true
	}
	return false
}

func (r *inMemoryRateLimiter) RemoveClient(ctx context.Context, ip string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.clients[ip]; !ok {
		panic("ratelimit: tried to remove client without add")
	}
	r.numRemovedClients++
	select {
	case <-r.clients[ip]:
		return true
	case <-ctx.Done():
		return false
	case <-r.quit:
		return false
	}
}

func (r *inMemoryRateLimiter) TryAddPeer(ctx context.Context, id string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if ch, ok := r.peers[id]; !ok {
		r.peers[id] = make(semaphore, r.maxPeersPerFeedID)
		r.peers[id] <- empty{}
		return true
	} else if len(ch) < r.maxPeersPerFeedID {
		ch <- empty{}
		return true
	}
	return false
}

func (r *inMemoryRateLimiter) RemovePeer(ctx context.Context, ip string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.peers[ip]; !ok {
		panic("ratelimit: tried to remove peer without add")
	}
	r.numRemovedPeers++
	select {
	case <-r.peers[ip]:
		return true
	case <-ctx.Done():
		return false
	case <-r.quit:
		return false
	}
}

func (r *inMemoryRateLimiter) garbageCollectLoop(ctx context.Context) {
	ticker := time.NewTicker(r.gcPeriod)
	for {
		select {
		case <-r.quit:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.garbageCollect()
		}
	}
}

func (r *inMemoryRateLimiter) garbageCollect() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.numRemovedClients < r.gcMaxRemovedClients && r.numRemovedPeers < r.gcMaxRemovedPeers {
		// Nothing to do
		return
	}
	// Explicitly create new maps and copy over the keys/values
	// to new ones to reduce memory footprint of the underlying
	// dasta structures.
	// For details, see https://github.com/golang/go/issues/20135
	clientsCopy := make(map[string]semaphore)
	for k, v := range r.clients {
		clientsCopy[k] = v
	}
	r.clients = clientsCopy
	peersCopy := make(map[string]semaphore)
	for k, v := range r.peers {
		peersCopy[k] = v
	}
	r.clients = peersCopy
}
