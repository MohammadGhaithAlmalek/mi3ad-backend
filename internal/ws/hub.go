package ws

import "sync"

// Hub keeps track of connected clients and lets other packages
// (e.g. appointments) push a live event to a specific user.
// Feature packages depend on this via a small Notifier interface,
// so there's no import cycle between ws and other features.
type Hub struct {
	mu      sync.RWMutex
	clients map[string]*Client // userID -> client
}

func NewHub() *Hub {
	return &Hub{clients: make(map[string]*Client)}
}

func (h *Hub) register(userID string, c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[userID] = c
}

func (h *Hub) unregister(userID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, userID)
}

// Notify sends an event to a single connected user, if they're online.
// Silently does nothing if the user has no active connection - callers
// don't need to care whether the user is online.
func (h *Hub) Notify(userID string, event any) {
	h.mu.RLock()
	client, ok := h.clients[userID]
	h.mu.RUnlock()

	if !ok {
		return
	}
	client.send(event)
}
