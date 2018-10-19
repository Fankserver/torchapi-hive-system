package hive

const (
	EventTypeServerStateChange        = "serverStateChange"
	EventTypeFactionCreated           = "factionCreated"
	EventTypeFactionCreatedComplete   = "factionCreatedComplete"
	EventTypeFactionEdited            = "factionEdited"
	EventTypeFactionAutoAcceptChanged = "factionAutoAcceptChanged"
)

type ServerStateChanged struct {
	State string
}

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

type EventFactionEdited struct {
	FactionID   int64 `json:"FactionId"`
	Tag         string
	Name        string
	Description string
	PrivateInfo string
}

type EventFactionAutoAcceptChangeEvent struct {
	FactionID        int64 `json:"FactionId"`
	AutoAcceptMember bool
	AutoAcceptPeace  bool
}
