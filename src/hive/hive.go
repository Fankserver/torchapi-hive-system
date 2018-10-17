package hive

import (
	"encoding/json"
	"net/http"

	"github.com/globalsign/mgo/bson"
)

type Hive struct {
	ID   bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Name string        `json:"name" bson:"name"`
}

func (s *system) CreateHive(w http.ResponseWriter, r *http.Request) {
	var h Hive
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&h); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	conn := s.db.Copy()
	defer conn.Close()

	err := conn.DB("torchhive").C("hive").Insert(h)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *system) GetHives(w http.ResponseWriter, r *http.Request) {
	conn := s.db.Copy()
	defer conn.Close()

	var h []Hive
	err := conn.DB("torchhive").C("hive").Find(nil).All(&h)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(h); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
