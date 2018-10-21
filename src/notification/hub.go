package notification

import (
	"github.com/fankserver/torchapi-hive-system/src/hive"
	"github.com/sirupsen/logrus"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	system *hive.System
}

func NewHub(system *hive.System) *Hub {
	return &Hub{
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		system:     system,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			logrus.Info("client registered", client.hiveID.Hex(), client.sectorID.Hex())
			err := h.system.UpdateSectorState(client.hiveID, client.sectorID, hive.SectorStateBooting)
			if err != nil {
				logrus.Errorln(err)
			}
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			logrus.Info("client unregistered", client.hiveID.Hex(), client.sectorID.Hex())
			err := h.system.UpdateSectorState(client.hiveID, client.sectorID, hive.SectorStateOffline)
			if err != nil {
				logrus.Errorln(err)
			}
		}
	}
}
