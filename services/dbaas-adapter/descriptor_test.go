package main

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

type fakeSource struct{}

func (fakeSource) GetSupportedRoles() []string { return []string{"admin", "rw", "ro"} }
func (fakeSource) GetFeatures() map[string]bool {
	return map[string]bool{"multiusers": true, "tls": false}
}
func (fakeSource) GetROHost() string { return "" }

func newApp(id string) *fiber.App {
	app := fiber.New()
	RegisterDescriptorHandler(app, "user", "pass",
		BuildDescriptor(id, "cassandra", map[string]string{"clusterName": "cassandra"}, fakeSource{}))
	return app
}

func TestDescriptorRequiresAuth(t *testing.T) {
	resp, err := newApp("ns").Test(httptest.NewRequest("GET", DescriptorPath, nil))
	if err != nil || resp.StatusCode != 401 {
		t.Fatalf("expected 401, got %v, err %v", resp, err)
	}
}

func TestDescriptorOK(t *testing.T) {
	req := httptest.NewRequest("GET", DescriptorPath, nil)
	req.SetBasicAuth("user", "pass")
	resp, err := newApp("ns").Test(req)
	if err != nil || resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %v, err %v", resp, err)
	}
	var d PhysicalDatabaseDescriptor
	if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
		t.Fatal(err)
	}
	if d.PhysicalDatabaseId != "ns" || d.Type != "cassandra" || len(d.SupportedRoles) != 3 ||
		len(d.ApiVersions.Specs) != 1 || !d.Features["multiusers"] {
		t.Fatalf("unexpected descriptor: %+v", d)
	}
}

func TestDescriptorUnavailableWithoutId(t *testing.T) {
	req := httptest.NewRequest("GET", DescriptorPath, nil)
	req.SetBasicAuth("user", "pass")
	resp, _ := newApp("").Test(req)
	if resp.StatusCode != 503 {
		t.Fatalf("expected 503, got %d", resp.StatusCode)
	}
}
