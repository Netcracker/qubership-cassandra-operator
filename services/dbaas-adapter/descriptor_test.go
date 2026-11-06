// Copyright 2024-2025 NetCracker Technology Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
