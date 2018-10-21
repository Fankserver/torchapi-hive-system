package notification

import (
	"bytes"
	"encoding/json"

	"github.com/sirupsen/logrus"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	event        chan *event
	eventHandler func(hiveHex string, sectorHex string, message []byte) (broadcast bool, sectorEvents map[string][]byte, err error)
}

type event struct {
	hiveHex   string
	sectorHex string
	message   []byte
}

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte, 512),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		event:      make(chan *event, 512),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) RegisterEventHandler(eventHandler func(hiveHex string, sectorHex string, message []byte) (broadcast bool, sectorEvents map[string][]byte, err error)) {
	h.eventHandler = eventHandler
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
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
		case event := <-h.event:
			broadcast, sectorEvents, err := h.eventHandler(event.hiveHex, event.sectorHex, event.message)
			if err != nil {
				logrus.Errorln(err)
			}

			if broadcast {
				data, err := json.Marshal(event)
				if err != nil {
					logrus.Errorln(err.Error())
					return
				}
				data = bytes.TrimSpace(bytes.Replace(data, newline, space, -1))
				logrus.Info("broadcast")
				for client := range h.clients {
					if client.hiveID != event.hiveHex || client.sectorID == event.sectorHex {
						logrus.Infoln("skip client", client.hiveID, client.sectorID)
						continue
					}
					logrus.Infoln("send client", client.hiveID, client.sectorID)
					select {
					case client.send <- data:
					default:
						close(client.send)
						delete(h.clients, client)
					}
					break
				}
			} else if sectorEvents != nil {
				logrus.Info("select events")
				for k, v := range sectorEvents {
					for client := range h.clients {
						if client.hiveID != event.hiveHex || client.sectorID != k {
							logrus.Infoln("skip client", client.hiveID, client.sectorID)
							continue
						}
						v = bytes.TrimSpace(bytes.Replace(v, newline, space, -1))
						logrus.Infoln("send client", client.hiveID, client.sectorID)
						select {
						case client.send <- v:
						default:
							close(client.send)
							delete(h.clients, client)
						}
						break
					}
				}
			}
		}
	}
}
