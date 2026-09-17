// Package testapi supplies an in-memory management API for provider tests.
package testapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	hcti "github.com/htmlcsstoimage/go-client"
	"github.com/htmlcsstoimage/go-client/management"
)

type Server struct {
	Storage           *management.StorageDestination
	StorageSecret     *string
	StorageWrites     int
	StorageTests      int
	StorageRetentions int
	StorageDowngraded bool
	StorageFailure    bool
	StorageReadBody   string
	ExternalReads     int
	StorageBody       map[string]json.RawMessage
	OG                *management.OGConfig
	OGWrites          int
	OGDowngraded      bool
	OGReadBody        string
	OGBody            map[string]json.RawMessage
	Key               *management.APIKey
	KeyWrites         int
	KeyReadBody       string
	*httptest.Server
	mu         sync.Mutex
	Proxy      *management.Proxy
	Password   *string
	Writes     int
	Retentions int
	ReadStatus int
}

func New(t *testing.T) *Server {
	t.Helper()
	s := &Server{}
	s.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		defer s.mu.Unlock()
		if !strings.HasPrefix(r.UserAgent(), "HCTIGo/") || !strings.Contains(r.UserAgent(), " HCTITerraform/") {
			t.Error("missing SDK/provider user-agent identifiers")
		}
		id, key, ok := r.BasicAuth()
		if !ok || id != "test-id" || key != "test-key" {
			t.Error("incorrect API credentials")
			http.Error(w, "unauthorized", 401)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/v1/storage-destinations") {
			w.Header().Set("Content-Type", "application/json")
			s.serveStorage(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/v1/og-configs") {
			w.Header().Set("Content-Type", "application/json")
			s.serveOG(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/v1/api-keys") {
			w.Header().Set("Content-Type", "application/json")
			s.serveKey(w, r)
			return
		}
		if !strings.HasPrefix(r.URL.Path, "/v1/proxies") {
			t.Errorf("unexpected endpoint %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case "GET":
			if s.ReadStatus != 0 {
				w.WriteHeader(s.ReadStatus)
				w.Write([]byte(`{"error":"test failure"}`))
				return
			}
			if s.Proxy == nil {
				http.NotFound(w, r)
				return
			}
			json.NewEncoder(w).Encode(s.Proxy)
		case "POST":
			var body management.ProxyRequest
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Error(err)
				w.WriteHeader(400)
				return
			}
			create := r.URL.Path == "/v1/proxies"
			if !create && s.Proxy == nil {
				http.NotFound(w, r)
				return
			}
			if a := body.Authentication; a != nil {
				retain := a.RetainPassword != nil && *a.RetainPassword
				if retain {
					if create || a.Password != nil || s.Proxy.Username == nil || *s.Proxy.Username != a.Username {
						t.Error("invalid retention request")
						w.WriteHeader(400)
						return
					}
					s.Retentions++
				} else {
					if a.Password == nil {
						t.Error("missing replacement password")
						w.WriteHeader(400)
						return
					}
					s.Password = a.Password
				}
			} else {
				s.Password = nil
			}
			s.Writes++
			port := body.Port
			if port == nil {
				port = hcti.Ptr(uint16(80))
				if strings.HasPrefix(strings.TrimSpace(body.URL), "https:") {
					port = hcti.Ptr(uint16(443))
				}
			}
			var username *string
			if body.Authentication != nil {
				username = &body.Authentication.Username
			}
			hosts := body.BypassHosts
			if hosts == nil {
				hosts = []string{}
			}
			now := time.Date(2026, 9, 16, 12, 0, s.Writes, 0, time.UTC)
			s.Proxy = &management.Proxy{ID: "proxy-test", Name: strings.TrimSpace(body.Name), URL: strings.TrimSpace(body.URL), Port: port, Username: username, Enabled: !body.Disabled, BypassHosts: hosts, CreatedAt: time.Date(2026, 9, 16, 12, 0, 0, 0, time.UTC), UpdatedAt: now}
			json.NewEncoder(w).Encode(s.Proxy)
		case "DELETE":
			if s.Proxy == nil {
				http.NotFound(w, r)
				return
			}
			s.Proxy = nil
			s.Password = nil
			s.Writes++
			w.WriteHeader(204)
		default:
			t.Errorf("unexpected method %s", r.Method)
			w.WriteHeader(405)
		}
	}))
	t.Cleanup(s.Close)
	return s
}
func (s *Server) Counts() (int, int) { s.mu.Lock(); defer s.mu.Unlock(); return s.Writes, s.Retentions }
func (s *Server) Status(status int)  { s.mu.Lock(); defer s.mu.Unlock(); s.ReadStatus = status }
func (s *Server) Snapshot() (*management.Proxy, *string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Proxy, s.Password
}
