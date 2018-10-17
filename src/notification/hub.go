package notification

import (
	"github.com/fankserver/torchapi-hive-system/src/hive"
	"github.com/sirupsen/logrus"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	system *hive.System

	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

func NewHub(system *hive.System) *Hub {
	return &Hub{
		system:     system,
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			err := h.system.UpdateSectorState(client.hiveID, client.sectorID, hive.SectorStateOnline)
			if err != nil {
				logrus.Errorln(err)
			}
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			err := h.system.UpdateSectorState(client.hiveID, client.sectorID, hive.SectorStateOffline)
			if err != nil {
				logrus.Errorln(err)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}
