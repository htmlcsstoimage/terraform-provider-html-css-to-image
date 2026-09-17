package provider

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/htmlcsstoimage/go-client/management"
)

func TestAPIDiagnostics(t *testing.T) {
	api := &management.APIError{StatusCode: 400, Code: "Bad Request", Message: "Destination connection test failed.", ValidationErrors: []management.ValidationError{{Path: "disabled", Message: "Cannot enable destination."}, {Message: "Check bucket permissions."}}}
	got := apiErrorDetail(fmt.Errorf("wrapped: %w", api))
	for _, want := range []string{"HTTP 400", "Bad Request", "Destination connection test failed.", "disabled: Cannot enable destination.", "Check bucket permissions."} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in %q", want, got)
		}
	}
	if got := apiErrorDetail(errors.New("network unavailable")); got != "network unavailable" {
		t.Fatal(got)
	}
	if got := apiErrorDetail(&management.APIError{StatusCode: 503}); !strings.Contains(got, "HTTP 503") {
		t.Fatal(got)
	}
}

func TestStorageCreateValidationDiagnostic(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(400)
		fmt.Fprint(w, `{"error":"Bad Request","message":"Connection test failed.","validationErrors":[{"path":"connection_info.secret_access_key","message":"Credential secret rejected for GOOG-test."}],"extra":"DO-NOT-PRINT-RAW-BODY"}`)
	}))
	defer server.Close()
	ctx := context.Background()
	r := &StorageDestinationResource{client: management.NewClient("id", "key", management.WithBaseURL(server.URL))}
	var schema resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schema)
	m := storageBase()
	plan := tfsdk.Plan{Schema: schema.Schema}
	if d := plan.Set(ctx, &m); d.HasError() {
		t.Fatal(d)
	}
	response := resource.CreateResponse{State: tfsdk.State{Schema: schema.Schema}}
	r.Create(ctx, resource.CreateRequest{Plan: plan}, &response)
	if !response.Diagnostics.HasError() {
		t.Fatal("missing diagnostic")
	}
	got := response.Diagnostics[0].Detail()
	for _, want := range []string{"HTTP 400", "Connection test failed.", "connection_info.secret_access_key", "[REDACTED]"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q: %s", want, got)
		}
	}
	for _, secret := range []string{"Credential secret", "GOOG-test", "DO-NOT-PRINT-RAW-BODY"} {
		if strings.Contains(got, secret) {
			t.Fatalf("sensitive/raw content included: %s", got)
		}
	}
}
