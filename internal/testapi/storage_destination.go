package testapi

import (
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"strings"
	"time"

	hcti "github.com/htmlcsstoimage/go-client"
	"github.com/htmlcsstoimage/go-client/management"
)

func (s *Server) serveStorage(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" && s.ReadStatus != 0 {
		w.WriteHeader(s.ReadStatus)
		return
	}
	if r.URL.Path == "/v1/storage-destinations/aws-external-id" && r.Method == "GET" {
		s.ExternalReads++
		json.NewEncoder(w).Encode(management.AWSExternalID{ExternalID: "org-external-test", WriterRoleARN: "arn:aws:iam::123456789012:role/hcti-writer"})
		return
	}
	create := r.Method == "POST" && r.URL.Path == "/v1/storage-destinations"
	if !create && r.URL.Path != "/v1/storage-destinations/storage-test" {
		http.NotFound(w, r)
		return
	}
	if !create && s.Storage == nil {
		http.NotFound(w, r)
		return
	}
	switch r.Method {
	case "GET":
		if s.StorageReadBody != "" {
			w.Write([]byte(s.StorageReadBody))
			return
		}
		json.NewEncoder(w).Encode(s.Storage)
	case "POST":
		data, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(400)
			return
		}
		var req struct {
			Name                string          `json:"name"`
			Disabled            bool            `json:"disabled"`
			HCTIStorageDisabled bool            `json:"hcti_storage_disabled"`
			Connection          json.RawMessage `json:"connection_info"`
		}
		if json.Unmarshal(data, &req) != nil {
			w.WriteHeader(400)
			return
		}
		s.StorageBody = nil
		json.Unmarshal(req.Connection, &s.StorageBody)
		var c management.StorageConnectionInfo
		var creds management.StorageCredentials
		if json.Unmarshal(req.Connection, &c) != nil || json.Unmarshal(req.Connection, &creds) != nil {
			w.WriteHeader(400)
			return
		}
		if !create && s.StorageDowngraded && req.Disabled {
			s.StorageWrites++
			s.Storage.Enabled = false
			json.NewEncoder(w).Encode(s.Storage)
			return
		}
		c.Bucket = strings.TrimSpace(c.Bucket)
		norm := func(v **string, lower bool) {
			if *v != nil {
				x := strings.TrimSpace(**v)
				if lower {
					x = strings.ToLower(x)
				}
				*v = nil
				if x != "" {
					*v = &x
				}
			}
		}
		norm(&c.KeyPrefix, false)
		if c.KeyPrefix != nil {
			x := strings.Trim(*c.KeyPrefix, "/")
			c.KeyPrefix = nil
			if x != "" {
				c.KeyPrefix = &x
			}
		}
		norm(&c.Region, true)
		norm(&c.AccessKeyID, false)
		norm(&c.RoleARN, false)
		norm(&c.CloudflareAccountID, true)
		norm(&c.CloudflareJurisdiction, true)
		norm(&c.Endpoint, false)
		if c.Endpoint != nil {
			x := strings.TrimRight(*c.Endpoint, "/")
			c.Endpoint = &x
		}
		retained := creds.RetainSecretAccessKey != nil && *creds.RetainSecretAccessKey
		if c.Provider != management.AWSS3 {
			if retained {
				if create || s.StorageSecret == nil || c.Provider != s.Storage.ConnectionInfo.Provider || c.AccessKeyID == nil || s.Storage.ConnectionInfo.AccessKeyID == nil || *c.AccessKeyID != *s.Storage.ConnectionInfo.AccessKeyID || creds.SecretAccessKey != nil {
					http.Error(w, "invalid retention", 400)
					return
				}
			} else if creds.SecretAccessKey == nil || strings.TrimSpace(*creds.SecretAccessKey) == "" {
				http.Error(w, "secret required", 400)
				return
			}
		}
		test := create || !reflect.DeepEqual(c, s.Storage.ConnectionInfo) || (!req.Disabled && !s.Storage.Enabled) || (c.Provider != management.AWSS3 && !retained)
		now := time.Date(2026, 9, 16, 12, 0, s.StorageWrites+1, 0, time.UTC)
		result := &management.StorageDestination{ID: "storage-test", Name: strings.TrimSpace(req.Name), Enabled: !req.Disabled, HCTIStorageDisabled: req.HCTIStorageDisabled, Provider: "Human readable label", ConnectionInfo: c, CreatedAt: time.Date(2026, 9, 16, 12, 0, 0, 0, time.UTC), UpdatedAt: now}
		if !create {
			result.LastTestedAt = s.Storage.LastTestedAt
			result.LastTestSucceeded = s.Storage.LastTestSucceeded
			result.LastTestError = s.Storage.LastTestError
		}
		if test {
			s.StorageTests++
			result.LastTestedAt = &now
			result.LastTestSucceeded = hcti.Ptr(!s.StorageFailure)
			if s.StorageFailure {
				result.LastTestError = hcti.Ptr("Connection test failed")
			}
			if s.StorageFailure && !req.Disabled {
				http.Error(w, "connection failed", 400)
				return
			}
		}
		if c.Provider == management.AWSS3 {
			s.StorageSecret = nil
		} else if retained {
			s.StorageRetentions++
		} else {
			s.StorageSecret = hcti.Ptr(strings.TrimSpace(*creds.SecretAccessKey))
		}
		s.StorageWrites++
		s.Storage = result
		json.NewEncoder(w).Encode(result)
	case "DELETE":
		s.StorageWrites++
		s.Storage = nil
		s.StorageSecret = nil
		w.WriteHeader(204)
	default:
		w.WriteHeader(405)
	}
}
func (s *Server) StorageSnapshot() (*management.StorageDestination, int, int, int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, _ := json.Marshal(s.Storage)
	var v *management.StorageDestination
	json.Unmarshal(data, &v)
	return v, s.StorageWrites, s.StorageTests, s.StorageRetentions
}
func (s *Server) StorageMode(downgraded, fail bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.StorageDowngraded = downgraded
	s.StorageFailure = fail
}
func (s *Server) StorageMalformed(body string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.StorageReadBody = body
}
