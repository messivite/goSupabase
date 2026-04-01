package handlers

import (
	"encoding/json"
	"net/http"
)

func init() {
	Register("ListTracks", ListTracks)
}

func ListTracks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"handler": "ListTracks",
		"status":  "ok",
	})
}
