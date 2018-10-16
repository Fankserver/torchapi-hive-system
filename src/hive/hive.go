package hive

import (
	"encoding/json"
	"net/http"
)

type Hive struct {
	ID   int    `json:"id" bson:"id"`
	Name string `json:"name" bson:"name"`
}

func CreateHive(w http.ResponseWriter, r *http.Request) {
	var h Hive
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&h); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

}
