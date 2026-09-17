package testapi

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	hcti "github.com/htmlcsstoimage/go-client"
	"github.com/htmlcsstoimage/go-client/management"
)

// serveOG runs under the fixture mutex. It models the public service's nullable
// fields, trim normalization, default readback, and downgrade disable-only path.
func (s *Server) serveOG(w http.ResponseWriter, r *http.Request) {
	create := r.Method == "POST" && r.URL.Path == "/v1/og-configs"
	if !create && r.URL.Path != "/v1/og-configs/og-management" {
		http.NotFound(w, r)
		return
	}
	if r.Method == "GET" && s.ReadStatus != 0 {
		w.WriteHeader(s.ReadStatus)
		return
	}
	if !create && s.OG == nil {
		http.NotFound(w, r)
		return
	}
	switch r.Method {
	case "GET":
		if s.OGReadBody != "" {
			w.Write([]byte(s.OGReadBody))
			return
		}
		json.NewEncoder(w).Encode(s.OG)
	case "POST":
		data, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(400)
			return
		}
		s.OGBody = nil
		if json.Unmarshal(data, &s.OGBody) != nil {
			w.WriteHeader(400)
			return
		}
		var common management.OGConfigOptions
		if json.Unmarshal(data, &common) != nil {
			w.WriteHeader(400)
			return
		}
		s.OGWrites++
		if !create && s.OGDowngraded && common.Disabled {
			s.OG.Enabled = false
			json.NewEncoder(w).Encode(s.OG)
			return
		}
		kind := ""
		json.Unmarshal(s.OGBody["config_type"], &kind)
		name := strings.TrimSpace(common.Name)
		description := common.Description
		if description != nil {
			v := strings.TrimSpace(*description)
			description = nil
			if v != "" {
				description = &v
			}
		}
		mode := management.PostProcess
		if common.OptimizationMode != nil {
			mode = *common.OptimizationMode
		}
		interval := uint32(86400)
		if common.RefreshIntervalSeconds != nil {
			interval = *common.RefreshIntervalSeconds
		}
		config := &management.OGConfig{ID: "og-management", DomainID: "domain-" + kind, ConfigType: management.OGConfigType(kind), Name: &name, Description: description, BaseURL: strings.TrimRight(strings.TrimSpace(common.BaseURL), "/"), Enabled: !common.Disabled, OptimizationMode: mode, RefreshIntervalSeconds: interval, CreatedAt: time.Date(2026, 9, 16, 12, 0, 0, 0, time.UTC), UpdatedAt: time.Date(2026, 9, 16, 12, 0, s.OGWrites, 0, time.UTC)}
		switch config.ConfigType {
		case management.OGConfigHTMLCSS:
			if _, ok := s.OGBody["template_id"]; ok {
				http.Error(w, "inactive variant", 400)
				return
			}
			var req management.HTMLCSSOGConfigRequest
			if json.Unmarshal(data, &req) != nil {
				w.WriteHeader(400)
				return
			}
			config.ExtractValues = req.ExtractValues
			config.DefaultOptions = req.DefaultOptions
			if config.DefaultOptions == nil {
				config.DefaultOptions = &management.OGDefaultImageOptions{}
			}
			o := config.DefaultOptions
			if o.IncludeHeadersOnSubrequests == nil {
				o.IncludeHeadersOnSubrequests = hcti.Ptr(false)
			}
			if o.CSS != nil && strings.TrimSpace(*o.CSS) == "" {
				o.CSS = nil
			}
			if len(o.Headers) == 0 {
				o.Headers = nil
			}
			if len(o.AdditionalHeaderOrigins) == 0 {
				o.AdditionalHeaderOrigins = nil
			}
		case management.OGConfigTemplated:
			if _, ok := s.OGBody["default_options"]; ok {
				http.Error(w, "inactive variant", 400)
				return
			}
			var req management.TemplatedOGConfigRequest
			if json.Unmarshal(data, &req) != nil {
				w.WriteHeader(400)
				return
			}
			config.TemplateID = hcti.Ptr(strings.TrimSpace(req.TemplateID))
			config.TemplateVersion = req.TemplateVersion
			config.TemplateValuesMapping = req.TemplateValuesMapping
			config.Headers = req.Headers
			config.AdditionalHeaderOrigins = req.AdditionalHeaderOrigins
			for i := range config.TemplateValuesMapping {
				m := &config.TemplateValuesMapping[i]
				m.TemplateKey = strings.TrimSpace(m.TemplateKey)
				if m.MetaKey != nil {
					m.MetaKey = hcti.Ptr(strings.TrimSpace(*m.MetaKey))
				}
			}
			if len(config.TemplateValuesMapping) == 0 {
				config.TemplateValuesMapping = nil
			}
			if len(config.Headers) == 0 {
				config.Headers = nil
			}
			if len(config.AdditionalHeaderOrigins) == 0 {
				config.AdditionalHeaderOrigins = nil
			}
		default:
			http.Error(w, "invalid discriminator", 400)
			return
		}
		s.OG = config
		json.NewEncoder(w).Encode(config)
	case "DELETE":
		s.OGWrites++
		s.OG = nil
		w.WriteHeader(204)
	default:
		w.WriteHeader(405)
	}
}
func (s *Server) OGSnapshot() (*management.OGConfig, int, map[string]json.RawMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Deep-copy so tests cannot race the HTTP fixture or mutate it accidentally.
	data, _ := json.Marshal(s.OG)
	var result *management.OGConfig
	json.Unmarshal(data, &result)
	data, _ = json.Marshal(s.OGBody)
	var body map[string]json.RawMessage
	json.Unmarshal(data, &body)
	return result, s.OGWrites, body
}
func (s *Server) DowngradeOG()            { s.mu.Lock(); defer s.mu.Unlock(); s.OGDowngraded = true }
func (s *Server) OGMalformed(body string) { s.mu.Lock(); defer s.mu.Unlock(); s.OGReadBody = body }
