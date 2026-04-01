package handlers

import (
	"encoding/json"
	"net/http"
)

func init() {
	Register("ListPlaylists", ListPlaylists)
}

func ListPlaylists(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"handler": "ListPlaylists",
		"status":  "ok",
	})
}
