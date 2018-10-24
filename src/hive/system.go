package hive

import (
	"encoding/json"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/sirupsen/logrus"
)

type System struct {
	db *mgo.Session
}

func NewSystem(dbConnectionString string) (*System, error) {
	dialInfo, err := mgo.ParseURL(dbConnectionString)
	if err != nil {
		return nil, err
	}

	logrus.Info("1")
	mongoSession, err := mgo.DialWithInfo(dialInfo)
	if err != nil {
		return nil, err
	}
	logrus.Info("2")

	return &System{
		db: mongoSession,
	}, nil
}

type EventSectorChange struct {
	Type string `json:"type"`
	Raw  string `json:"raw"`
}

func (s *System) ProcessSectorEvent(hiveHex string, sectorHex string, message []byte) (broadcast bool, sectorEvents map[string][]byte, err error) {
	hiveID := bson.ObjectIdHex(hiveHex)
	sectorID := bson.ObjectIdHex(sectorHex)
	var event EventSectorChange
	err = json.Unmarshal(message, &event)
	if err != nil {
		return
	}

	logrus.Info(event.Type)
	logrus.Info(event.Raw)

	promote := false

	switch event.Type {
	case EventTypeServerStateChange:
		var serverStateChanged ServerStateChanged
		err = json.Unmarshal([]byte(event.Raw), &serverStateChanged)
		if err != nil {
			return
		}

		switch serverStateChanged.State {
		case "Loaded":
			err = s.UpdateSectorState(hiveID, sectorID, SectorStateOnline)
		case "Unloading":
			err = s.UpdateSectorState(hiveID, sectorID, SectorStateOffline)
		}
	case EventTypeFactionCreated:
		broadcast = true

		var factionCreated EventFactionCreated
		err = json.Unmarshal([]byte(event.Raw), &factionCreated)
		if err != nil {
			return
		}

		err = s.CreateFaction(hiveID, sectorID, factionCreated)
		if err != nil {
			return
		}
	case EventTypeFactionCreatedComplete:
		var factionCreatedComplete EventFactionCreatedComplete
		err = json.Unmarshal([]byte(event.Raw), &factionCreatedComplete)
		if err != nil {
			return
		}

		err = s.AddFactionSector(hiveID, sectorID, factionCreatedComplete)
		if err != nil {
			return
		}
	case EventTypeFactionEdited:
		var factionEdited EventFactionEdited
		err = json.Unmarshal([]byte(event.Raw), &factionEdited)
		if err != nil {
			return
		}

		err = s.EditFaction(hiveID, sectorID, factionEdited)
		if err != nil {
			return
		}

		sectorEvents = make(map[string][]byte)

		var faction *Faction
		faction, err = s.GetFaction(hiveID, sectorID, factionEdited.FactionID)
		if err != nil {
			return
		}
		for _, v := range faction.Sectors {
			// ignore own sector
			if v.SectorID == sectorID {
				continue
			}

			factionEdited.FactionID = v.EntityID
			var data []byte
			data, err = json.Marshal(factionEdited)
			if err != nil {
				return
			}
			event.Raw = string(data)
			data, err = json.Marshal(event)
			if err != nil {
				return
			}
			sectorEvents[v.SectorID.Hex()] = data
		}
	case EventTypeFactionAutoAcceptChanged:
		var factionAutoAcceptChange EventFactionAutoAcceptChangeEvent
		err = json.Unmarshal([]byte(event.Raw), &factionAutoAcceptChange)
		if err != nil {
			return
		}

		err = s.ChangeAutoAccept(hiveID, sectorID, factionAutoAcceptChange)
		if err != nil {
			return
		}

		sectorEvents = make(map[string][]byte)

		var faction *Faction
		faction, err = s.GetFaction(hiveID, sectorID, factionAutoAcceptChange.FactionID)
		if err != nil {
			return
		}
		for _, v := range faction.Sectors {
			// ignore own sector
			if v.SectorID == sectorID {
				continue
			}

			factionAutoAcceptChange.FactionID = v.EntityID
			var data []byte
			data, err = json.Marshal(factionAutoAcceptChange)
			if err != nil {
				return
			}
			event.Raw = string(data)
			data, err = json.Marshal(event)
			if err != nil {
				return
			}
			sectorEvents[v.SectorID.Hex()] = data
		}
	case EventTypeFactionMemberSendJoin:
		var factionMember EventFactionMember
		err = json.Unmarshal([]byte(event.Raw), &factionMember)
		if err != nil {
			return
		}

		var faction *Faction
		faction, err = s.MemberSendJoin(hiveID, sectorID, factionMember)
		if err != nil {
			return
		}

		for _, v := range faction.Sectors {
			// ignore own sector
			if v.SectorID == sectorID {
				continue
			}

			factionMember.FactionID = v.EntityID
			var data []byte
			data, err = json.Marshal(factionMember)
			if err != nil {
				return
			}
			event.Raw = string(data)
			data, err = json.Marshal(event)
			if err != nil {
				return
			}
			sectorEvents[v.SectorID.Hex()] = data
		}
	case EventTypeFactionMemberCancelJoin:
		var factionMember EventFactionMember
		err = json.Unmarshal([]byte(event.Raw), &factionMember)
		if err != nil {
			return
		}

		var faction *Faction
		faction, err = s.MemberLeave(hiveID, sectorID, factionMember)
		if err != nil {
			return
		}

		for _, v := range faction.Sectors {
			// ignore own sector
			if v.SectorID == sectorID {
				continue
			}

			factionMember.FactionID = v.EntityID
			var data []byte
			data, err = json.Marshal(factionMember)
			if err != nil {
				return
			}
			event.Raw = string(data)
			data, err = json.Marshal(event)
			if err != nil {
				return
			}
			sectorEvents[v.SectorID.Hex()] = data
		}
	case EventTypeFactionMemberAcceptJoin:
		var factionMember EventFactionMember
		err = json.Unmarshal([]byte(event.Raw), &factionMember)
		if err != nil {
			return
		}

		var faction *Faction
		faction, err = s.MemberAcceptJoin(hiveID, sectorID, factionMember)
		if err != nil {
			return
		}

		for _, v := range faction.Sectors {
			// ignore own sector
			if v.SectorID == sectorID {
				continue
			}

			factionMember.FactionID = v.EntityID
			var data []byte
			data, err = json.Marshal(factionMember)
			if err != nil {
				return
			}
			event.Raw = string(data)
			data, err = json.Marshal(event)
			if err != nil {
				return
			}
			sectorEvents[v.SectorID.Hex()] = data
		}
	case EventTypeFactionMemberPromote:
		promote = true
		fallthrough
	case EventTypeFactionMemberDemote:
		var factionMember EventFactionMember
		err = json.Unmarshal([]byte(event.Raw), &factionMember)
		if err != nil {
			return
		}

		var faction *Faction
		faction, err = s.MemberPromoteDemote(hiveID, sectorID, factionMember, promote)
		if err != nil {
			return
		}

		for _, v := range faction.Sectors {
			// ignore own sector
			if v.SectorID == sectorID {
				continue
			}

			factionMember.FactionID = v.EntityID
			var data []byte
			data, err = json.Marshal(factionMember)
			if err != nil {
				return
			}
			event.Raw = string(data)
			data, err = json.Marshal(event)
			if err != nil {
				return
			}
			sectorEvents[v.SectorID.Hex()] = data
		}
	case EventTypeFactionMemberKick:
		fallthrough
	case EventTypeFactionMemberLeave:
		var factionMember EventFactionMember
		err = json.Unmarshal([]byte(event.Raw), &factionMember)
		if err != nil {
			return
		}

		var faction *Faction
		faction, err = s.MemberLeave(hiveID, sectorID, factionMember)
		if err != nil {
			return
		}

		for _, v := range faction.Sectors {
			// ignore own sector
			if v.SectorID == sectorID {
				continue
			}

			factionMember.FactionID = v.EntityID
			var data []byte
			data, err = json.Marshal(factionMember)
			if err != nil {
				return
			}
			event.Raw = string(data)
			data, err = json.Marshal(event)
			if err != nil {
				return
			}
			sectorEvents[v.SectorID.Hex()] = data
		}
	case EventTypeFactionSendPeaceRequest:
		var factionPeaceWar EventFactionPeaceWar
		err = json.Unmarshal([]byte(event.Raw), &factionPeaceWar)
		if err != nil {
			return
		}

		var fromFaction *Faction
		var toFaction *Faction
		fromFaction, toFaction, err = s.SendPeaceRequest(hiveID, sectorID, factionPeaceWar)
		if err != nil {
			return
		}
		for _, v := range fromFaction.Sectors {
			// ignore own sector
			if v.SectorID == sectorID {
				continue
			}

			// get to faction entity id
			for _, vv := range toFaction.Sectors {
				if vv.SectorID != v.SectorID {
					continue
				}

				factionPeaceWar.ToFactionID = vv.EntityID
				break
			}

			factionPeaceWar.FromFactionID = v.EntityID
			var data []byte
			data, err = json.Marshal(factionPeaceWar)
			if err != nil {
				return
			}
			event.Raw = string(data)
			data, err = json.Marshal(event)
			if err != nil {
				return
			}
			sectorEvents[v.SectorID.Hex()] = data
		}
	case EventTypeFactionCancelPeaceRequest:
		var factionPeaceWar EventFactionPeaceWar
		err = json.Unmarshal([]byte(event.Raw), &factionPeaceWar)
		if err != nil {
			return
		}

		var fromFaction *Faction
		var toFaction *Faction
		fromFaction, toFaction, err = s.CancelPeaceRequest(hiveID, sectorID, factionPeaceWar)
		if err != nil {
			return
		}
		for _, v := range fromFaction.Sectors {
			// ignore own sector
			if v.SectorID == sectorID {
				continue
			}

			// get to faction entity id
			for _, vv := range toFaction.Sectors {
				if vv.SectorID != v.SectorID {
					continue
				}

				factionPeaceWar.ToFactionID = vv.EntityID
				break
			}

			factionPeaceWar.FromFactionID = v.EntityID
			var data []byte
			data, err = json.Marshal(factionPeaceWar)
			if err != nil {
				return
			}
			event.Raw = string(data)
			data, err = json.Marshal(event)
			if err != nil {
				return
			}
			sectorEvents[v.SectorID.Hex()] = data
		}
	case EventTypeFactionAcceptPeace:
		var factionPeaceWar EventFactionPeaceWar
		err = json.Unmarshal([]byte(event.Raw), &factionPeaceWar)
		if err != nil {
			return
		}

		var fromFaction *Faction
		var toFaction *Faction
		fromFaction, toFaction, err = s.AcceptPeace(hiveID, sectorID, factionPeaceWar)
		if err != nil {
			return
		}
		for _, v := range fromFaction.Sectors {
			// ignore own sector
			if v.SectorID == sectorID {
				continue
			}

			// get to faction entity id
			for _, vv := range toFaction.Sectors {
				if vv.SectorID != v.SectorID {
					continue
				}

				factionPeaceWar.ToFactionID = vv.EntityID
				break
			}

			factionPeaceWar.FromFactionID = v.EntityID
			var data []byte
			data, err = json.Marshal(factionPeaceWar)
			if err != nil {
				return
			}
			event.Raw = string(data)
			data, err = json.Marshal(event)
			if err != nil {
				return
			}
			sectorEvents[v.SectorID.Hex()] = data
		}
	case EventTypeFactionDeclareWar:
		var factionPeaceWar EventFactionPeaceWar
		err = json.Unmarshal([]byte(event.Raw), &factionPeaceWar)
		if err != nil {
			return
		}

		var fromFaction *Faction
		var toFaction *Faction
		fromFaction, toFaction, err = s.DeclareWar(hiveID, sectorID, factionPeaceWar)
		if err != nil {
			return
		}
		for _, v := range fromFaction.Sectors {
			// ignore own sector
			if v.SectorID == sectorID {
				continue
			}

			// get to faction entity id
			for _, vv := range toFaction.Sectors {
				if vv.SectorID != v.SectorID {
					continue
				}

				factionPeaceWar.ToFactionID = vv.EntityID
				break
			}

			factionPeaceWar.FromFactionID = v.EntityID
			var data []byte
			data, err = json.Marshal(factionPeaceWar)
			if err != nil {
				return
			}
			event.Raw = string(data)
			data, err = json.Marshal(event)
			if err != nil {
				return
			}
			sectorEvents[v.SectorID.Hex()] = data
		}
	default:
		logrus.Warnln("received unknown event type", event.Type)
	}

	return
}
