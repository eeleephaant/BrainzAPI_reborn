package wsmodels

import (
	"log"
	"sync"
)

type Hub struct {
	mu       sync.RWMutex
	channels map[string]map[*Client]struct{}
}

var MainHub = NewHub()

func NewHub() *Hub {
	return &Hub{
		channels: make(map[string]map[*Client]struct{}),
	}
}

// Добавляем клиента в канал
func (h *Hub) AddClient(channel string, client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.channels[channel] == nil {
		h.channels[channel] = make(map[*Client]struct{})
	}
	h.channels[channel][client] = struct{}{}
}

// Убираем клиента
func (h *Hub) RemoveClient(channel string, client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.channels[channel] != nil {
		delete(h.channels[channel], client)
		if len(h.channels[channel]) == 0 {
			delete(h.channels, channel)
		}
	}
}

// Отправляем всем клиентам канала
func (h *Hub) Broadcast(channel string, msg []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for client := range h.channels[channel] {
		select {
		case client.Send <- msg:
		default:
			log.Println("drop message, slow client")
		}
	}
}
