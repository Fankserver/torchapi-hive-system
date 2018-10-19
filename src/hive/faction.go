package hive

import (
	"encoding/json"
	"net/http"

	"github.com/globalsign/mgo/bson"
	"github.com/gorilla/mux"
)

const CollectionFaction = "faction"

type FactionSector struct {
	SectorID  bson.ObjectId `json:"sector_id" bson:"sector_id"`
	FactionID int64         `json:"faction_id" bson:"faction_id"`
}

type Faction struct {
	ID             bson.ObjectId   `json:"id" bson:"_id,omitempty"`
	HiveID         bson.ObjectId   `json:"-" bson:"hive_id"`
	Tag            string          `json:"tag" bson:"tag"`
	Name           string          `json:"name" bson:"name"`
	Description    string          `json:"description" bson:"description"`
	PrivateInfo    string          `json:"private_info" bson:"private_info"`
	AcceptHumans   bool            `json:"accept_humans" bson:"accept_humans"`
	FounderSteamID uint64          `json:"founder_steam_id" bson:"founder_steam_id"`
	Sectors        []FactionSector `json:"sectors" bson:"sectors"`
}

func (s *System) GetFactions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	conn := s.db.Copy()
	defer conn.Close()

	var factions []Faction
	err := conn.DB("torchhive").C(CollectionFaction).Find(bson.M{
		"hive_id": bson.ObjectIdHex(vars["hive_id"]),
	}).All(&factions)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(factions); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *System) CreateFaction(hiveID bson.ObjectId, sectorID bson.ObjectId, event EventFactionCreated) error {
	conn := s.db.Copy()
	defer conn.Close()

	err := conn.DB("torchhive").C(CollectionFaction).Insert(Faction{
		HiveID:       hiveID,
		Name:         event.Name,
		Tag:          event.Tag,
		Description:  event.Description,
		PrivateInfo:  event.PrivateInfo,
		AcceptHumans: event.AcceptHumans,
		Sectors: []FactionSector{
			{
				SectorID:  sectorID,
				FactionID: event.FactionID,
			},
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *System) AddFactionSector(hiveID bson.ObjectId, sectorID bson.ObjectId, event EventFactionCreatedComplete) error {
	conn := s.db.Copy()
	defer conn.Close()

	err := conn.DB("torchhive").C(CollectionFaction).Update(
		bson.M{
			"hive_id": hiveID,
			"tag":     event.Tag,
		},
		bson.M{
			"$push": bson.M{
				"sectors": bson.M{
					"sector_id":  sectorID,
					"faction_id": event.FactionID,
				},
			},
		},
	)
	if err != nil {
		return err
	}

	return nil
}
