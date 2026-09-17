package testapi

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/htmlcsstoimage/go-client/management"
)

// serveKey is called under the server mutex.
func (s *Server) serveKey(w http.ResponseWriter, r *http.Request) {
	create := r.Method == "POST" && r.URL.Path == "/v1/api-keys"
	if !create && r.URL.Path != "/v1/api-keys/key-management" {
		http.NotFound(w, r)
		return
	}
	if r.Method == "GET" && s.ReadStatus != 0 {
		w.WriteHeader(s.ReadStatus)
		w.Write([]byte(`{"error":"test failure"}`))
		return
	}
	if !create && s.Key == nil {
		http.NotFound(w, r)
		return
	}
	switch r.Method {
	case "GET":
		if s.KeyReadBody != "" {
			w.Write([]byte(s.KeyReadBody))
			return
		}
		json.NewEncoder(w).Encode(s.Key)
	case "POST":
		var req management.APIKeyRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Permissions == nil {
			http.Error(w, "invalid request", 400)
			return
		}
		if req.AllFuturePermissions && len(req.Permissions) > 0 {
			http.Error(w, "expected empty permissions", 400)
			return
		}
		name := "Key created 2026-09-16 12:00:00"
		if req.Name != nil && strings.TrimSpace(*req.Name) != "" {
			name = strings.TrimSpace(*req.Name)
		}
		var desc *string
		if req.Description != nil && strings.TrimSpace(*req.Description) != "" {
			v := strings.TrimSpace(*req.Description)
			desc = &v
		}
		grants := req.Permissions
		if req.AllFuturePermissions {
			grants = []management.Permission{"images:create", "templates:read"}
		}
		s.KeyWrites++
		s.Key = &management.APIKey{ID: "key-management", APIID: "key-auth-id", Name: name, Description: desc, Enabled: !req.Disabled, AllFuturePermissions: req.AllFuturePermissions, Permissions: grants, CreatedAt: time.Date(2026, 9, 16, 12, 0, 0, 0, time.UTC), UpdatedAt: time.Date(2026, 9, 16, 12, 0, s.KeyWrites, 0, time.UTC)}
		if create {
			json.NewEncoder(w).Encode(management.APIKeyWithSecret{APIKey: *s.Key, Secret: "create-only-secret"})
		} else {
			json.NewEncoder(w).Encode(s.Key)
		}
	case "DELETE":
		s.KeyWrites++
		s.Key.Enabled = false
		w.WriteHeader(204)
	default:
		w.WriteHeader(405)
	}
}
func (s *Server) KeySnapshot() (*management.APIKey, int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.Key == nil {
		return nil, s.KeyWrites
	}
	k := *s.Key
	return &k, s.KeyWrites
}
func (s *Server) ExpandKeyGrants() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Key.Permissions = append(s.Key.Permissions, "usage:read")
}
func (s *Server) KeyBody(body string) { s.mu.Lock(); defer s.mu.Unlock(); s.KeyReadBody = body }
