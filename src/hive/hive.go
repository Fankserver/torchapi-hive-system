package hive

import (
	"encoding/json"
	"net/http"

	"github.com/fankserver/torchapi-hive-system/src/model"
)

func CreateHive(w http.ResponseWriter, r *http.Request) {
	var h model.Hive
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&h); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

}
