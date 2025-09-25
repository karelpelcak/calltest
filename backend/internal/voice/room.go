package voice

import "sync"

type Room struct {
	ID    string
	Peers map[string]bool
	mu    sync.Mutex
}

func NewRoom(id string) *Room {
	return &Room{
		ID:    id,
		Peers: make(map[string]bool),
	}
}

func (r *Room) Join(userID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Peers[userID] = true
}

func (r *Room) Leave(userID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.Peers, userID)
}
