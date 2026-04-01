package handlers

import (
	"encoding/json"
	"net/http"
)

func init() {
	Register("UpdateTrack", UpdateTrack)
}

func UpdateTrack(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"handler": "UpdateTrack",
		"status":  "ok",
	})
}
