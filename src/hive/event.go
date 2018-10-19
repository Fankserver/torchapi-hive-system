package hive

const (
	EventTypeFactionCreated         = "factionCreated"
	EventTypeFactionCreatedComplete = "factionCreatedComplete"
)

type EventFactionCreated struct {
	FactionID      int64 `json:"FactionId"`
	Tag            string
	Name           string
	Description    string
	PrivateInfo    string
	AcceptHumans   bool
	FounderID      int64  `json:"FounderId"`
	FounderSteamID uint64 `json:"FounderSteamId"`
	FounderName    string
}

type EventFactionCreatedComplete struct {
	FactionID int64 `json:"FactionId"`
	Tag       string
}
