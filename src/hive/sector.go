package hive

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/globalsign/mgo/bson"
	"github.com/gorilla/mux"
)

type SectorState uint

const (
	SectorStateUnknown SectorState = iota
	SectorStateOffline
	SectorStateOnline
	SectorStateHighPopulation
	SectorStateFull
)

type Sector struct {
	ID       bson.ObjectId `json:"id" bson:"_id,omitempty"`
	HiveID   bson.ObjectId `json:"-" bson:"hive_id"`
	Name     string        `json:"name" bson:"name"`
	Address  string        `json:"address" bson:"address"`
	State    SectorState   `json:"state" bson:"state"`
	Position struct {
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

	err := conn.DB("torchhive").C("sector").Insert(hs)
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
	err := conn.DB("torchhive").C("sector").Find(bson.M{"hive_id": bson.ObjectIdHex(vars["hive_id"])}).All(&hs)
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

	count, err := conn.DB("torchhive").C("sector").Find(bson.M{
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

	return conn.DB("torchhive").C("sector").Update(
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

func (s *System) DeleteSector(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	conn := s.db.Copy()
	defer conn.Close()

	err := conn.DB("torchhive").C("sector").Remove(bson.M{
		"_id":     bson.ObjectIdHex(vars["sector_id"]),
		"hive_id": bson.ObjectIdHex(vars["hive_id"]),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
