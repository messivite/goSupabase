package handlers

import (
	"encoding/json"
	"net/http"
)

func init() {
	Register("CreateTrack", CreateTrack)
}

func CreateTrack(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"handler": "CreateTrack",
		"status":  "ok",
	})
}
