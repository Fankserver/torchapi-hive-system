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
	Type string          `json:"type"`
	Raw  json.RawMessage `json:"raw"`
}

func (s *System) ProcessSectorEvent(hiveID bson.ObjectId, sectorID bson.ObjectId, data []byte) (broadcast bool, err error) {
	var event EventSectorChange
	err = json.Unmarshal(data, &event)
	if err != nil {
		return
	}

	switch event.Type {
	case EventTypeFactionCreated:
		broadcast = true

		var factionCreated EventFactionCreated
		err = json.Unmarshal(event.Raw, &factionCreated)
		if err != nil {
			return
		}

		err = s.CreateFaction(hiveID, sectorID, factionCreated)
		if err != nil {
			return
		}
	case EventTypeFactionCreatedComplete:
		var factionCreatedComplete EventFactionCreatedComplete
		err = json.Unmarshal(event.Raw, &factionCreatedComplete)
		if err != nil {
			return
		}

		err = s.AddFactionSector(hiveID, sectorID, factionCreatedComplete)
		if err != nil {
			return
		}
	default:
		logrus.Warnln("received unknown event type", event.Type)
	}

	return
}
