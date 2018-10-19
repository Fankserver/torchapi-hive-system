package hive

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/globalsign/mgo/bson"
	"github.com/gorilla/mux"
)

const CollectionSector = "sector"

type SectorState uint

const (
	SectorStateUnknown SectorState = iota
	SectorStateOffline
	SectorStateBooting
	SectorStateOnline
)

type Sector struct {
	ID          bson.ObjectId `json:"id" bson:"_id,omitempty"`
	HiveID      bson.ObjectId `json:"-" bson:"hive_id"`
	Name        string        `json:"name" bson:"name"`
	Address     string        `json:"address" bson:"address"`
	State       SectorState   `json:"state" bson:"state"`
	MaxPlayer   int           `json:"max_player" bson:"max_player"`
	PlayerCount int           `json:"player_count" bson:"player_count"`
	Position    struct {
		X int `json:"x" bson:"x"`
		Y int `json:"y" bson:"y"`
	} `json:"position" bson:"position"`
	LastFactionSync  *time.Time `json:"last_faction_sync" bson:"last_faction_sync"`
	LastCurrencySync *time.Time `json:"last_currency_sync" bson:"last_currency_sync"`
}

func (s *System) CreateSector(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var hs Sector
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&hs); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	hs.HiveID = bson.ObjectIdHex(vars["hive_id"])

	conn := s.db.Copy()
	defer conn.Close()

	err := conn.DB("torchhive").C(CollectionSector).Insert(hs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *System) GetSectors(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	conn := s.db.Copy()
	defer conn.Close()

	var hs []Sector
	err := conn.DB("torchhive").C(CollectionSector).Find(bson.M{
		"hive_id": bson.ObjectIdHex(vars["hive_id"]),
	}).All(&hs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(hs); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *System) IsSectorValid(hiveID bson.ObjectId, sectorID bson.ObjectId) (bool, error) {
	conn := s.db.Copy()
	defer conn.Close()

	count, err := conn.DB("torchhive").C(CollectionSector).Find(bson.M{
		"_id":     sectorID,
		"hive_id": hiveID,
	}).Count()
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (s *System) UpdateSectorState(hiveID bson.ObjectId, sectorID bson.ObjectId, state SectorState) error {
	conn := s.db.Copy()
	defer conn.Close()

	return conn.DB("torchhive").C(CollectionSector).Update(
		bson.M{
			"_id":     sectorID,
			"hive_id": hiveID,
		},
		bson.M{
			"$set": bson.M{
				"state": state,
			},
		},
	)
}
func (s *System) UpdateSectorPlayers(hiveID bson.ObjectId, sectorID bson.ObjectId, maxPlayers int, currentPlayers int) error {
	conn := s.db.Copy()
	defer conn.Close()

	return conn.DB("torchhive").C(CollectionSector).Update(
		bson.M{
			"_id":     sectorID,
			"hive_id": hiveID,
		},
		bson.M{
			"$set": bson.M{
				"max_player":   maxPlayers,
				"player_count": currentPlayers,
			},
		},
	)
}

func (s *System) DeleteSector(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	conn := s.db.Copy()
	defer conn.Close()

	err := conn.DB("torchhive").C(CollectionSector).Remove(bson.M{
		"_id":     bson.ObjectIdHex(vars["sector_id"]),
		"hive_id": bson.ObjectIdHex(vars["hive_id"]),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
