package hive

const (
	EventTypeServerStateChange         = "serverStateChange"
	EventTypeFactionCreated            = "factionCreated"
	EventTypeFactionCreatedComplete    = "factionCreatedComplete"
	EventTypeFactionEdited             = "factionEdited"
	EventTypeFactionAutoAcceptChanged  = "factionAutoAcceptChanged"
	EventTypeFactionMemberSendJoin     = "factionMemberSendJoin"
	EventTypeFactionMemberCancelJoin   = "factionMemberCancelJoin"
	EventTypeFactionMemberAcceptJoin   = "factionMemberAcceptJoin"
	EventTypeFactionMemberPromote      = "factionMemberPromote"
	EventTypeFactionMemberDemote       = "factionMemberDemote"
	EventTypeFactionMemberKick         = "factionMemberKick"
	EventTypeFactionMemberLeave        = "factionMemberLeave"
	EventTypeFactionSendPeaceRequest   = "factionSendPeaceRequest"
	EventTypeFactionCancelPeaceRequest = "factionCancelPeaceRequest"
	EventTypeFactionAcceptPeace        = "factionAcceptPeace"
	EventTypeFactionDeclareWar         = "factionDeclareWar"
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

type EventFactionMember struct {
	FactionID     int64  `json:"FactionId"`
	PlayerID      int64  `json:"PlayerId"`
	PlayerSteamID uint64 `json:"PlayerSteamId"`
	PlayerName    string
}

type EventFactionPeaceWar struct {
	FromFactionID int64 `json:"FromFactionId"`
	ToFactionID   int64 `json:"ToFactionId"`
}
