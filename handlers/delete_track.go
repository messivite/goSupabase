package handlers

import (
	"encoding/json"
	"net/http"
)

func init() {
	Register("DeleteTrack", DeleteTrack)
}

func DeleteTrack(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"handler": "DeleteTrack",
		"status":  "ok",
	})
}
