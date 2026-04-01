package handlers

import (
	"encoding/json"
	"net/http"
)

func init() {
	Register("CreatePlaylist", CreatePlaylist)
}

func CreatePlaylist(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"handler": "CreatePlaylist",
		"status":  "ok",
	})
}
