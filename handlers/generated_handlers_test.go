package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGeneratedHandlersRespondJSON(t *testing.T) {
	tests := []struct {
		name    string
		handler http.HandlerFunc
	}{
		{"Health", Health},
		{"ListTracks", ListTracks},
		{"CreateTrack", CreateTrack},
		{"UpdateTrack", UpdateTrack},
		{"DeleteTrack", DeleteTrack},
		{"ListPlaylists", ListPlaylists},
		{"CreatePlaylist", CreatePlaylist},
	}

	for _, tt := range tests {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()
		tt.handler(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("%s status = %d, want 200", tt.name, rr.Code)
		}
		if rr.Header().Get("Content-Type") != "application/json" {
			t.Fatalf("%s content-type = %q", tt.name, rr.Header().Get("Content-Type"))
		}
		var body map[string]string
		if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
			t.Fatalf("%s invalid json: %v", tt.name, err)
		}
		if body["handler"] != tt.name {
			t.Fatalf("%s body handler = %q", tt.name, body["handler"])
		}
		if body["status"] != "ok" {
			t.Fatalf("%s body status = %q", tt.name, body["status"])
		}
	}
}
