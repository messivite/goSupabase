package handlers

import (
	"encoding/json"
	"net/http"
)

func init() {
	Register("Health", Health)
}

func Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"handler": "Health",
		"status":  "ok",
	})
}
