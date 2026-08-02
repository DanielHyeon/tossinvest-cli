package httpapi

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestOpenAPIPathsMatchExactRuntimeSurface(t *testing.T) {
	raw, err := os.ReadFile("../../docs/api/openapi-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var document struct {
		OpenAPI    string                    `json:"openapi"`
		Paths      map[string]map[string]any `json:"paths"`
		Components struct {
			Schemas map[string]json.RawMessage `json:"schemas"`
		} `json:"components"`
	}
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatal(err)
	}
	if document.OpenAPI != "3.1.0" {
		t.Fatalf("OpenAPI version=%q", document.OpenAPI)
	}
	got := make([]string, 0, len(document.Paths))
	for path := range document.Paths {
		got = append(got, path)
	}
	sort.Strings(got)
	want := []string{
		"/api/v1/candidates", "/api/v1/engine", "/api/v1/optimization", "/api/v1/optimization/applications",
		"/api/v1/optimization/previews", "/api/v1/orders", "/api/v1/performance", "/api/v1/positions",
		"/api/v1/settings", "/api/v1/stream",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("OpenAPI paths=%#v want=%#v", got, want)
	}
	for _, forbidden := range []string{"/api/v1/engine/live", "/api/v1/gate", "/api/v1/kill-switch", "/api/v1/protection", "/api/v1/activation-manifest", "/api/v1/optimization/rollback-previews"} {
		if _, exists := document.Paths[forbidden]; exists {
			t.Errorf("OpenAPI exposes forbidden route %q", forbidden)
		}
	}
}

func TestOpenAPIUsesTheSameStableSchemaVersions(t *testing.T) {
	raw, err := os.ReadFile("../../docs/api/openapi-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, version := range []string{SchemaVersion, ErrorSchemaVersion} {
		quoted, _ := json.Marshal(version)
		if !bytes.Contains(raw, quoted) {
			t.Errorf("OpenAPI lacks runtime schema version %q", version)
		}
	}
}

func TestEveryOpenAPILocalReferenceResolves(t *testing.T) {
	raw, err := os.ReadFile("../../docs/api/openapi-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var document map[string]any
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatal(err)
	}
	var walk func(any)
	walk = func(value any) {
		switch typed := value.(type) {
		case map[string]any:
			for key, child := range typed {
				if key == "$ref" {
					ref, ok := child.(string)
					if !ok || !localReferenceExists(document, ref) {
						t.Errorf("unresolved local OpenAPI reference %#v", child)
					}
				}
				walk(child)
			}
		case []any:
			for _, child := range typed {
				walk(child)
			}
		}
	}
	walk(document)
}

func TestOpenAPIDocumentsReconcileAwareReadModelsWithoutResolutionAuthority(t *testing.T) {
	raw, err := os.ReadFile("../../docs/api/openapi-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var document struct {
		Paths      map[string]json.RawMessage `json:"paths"`
		Components struct {
			Schemas map[string]struct {
				Required   []string                   `json:"required"`
				Properties map[string]json.RawMessage `json:"properties"`
			} `json:"schemas"`
		} `json:"components"`
	}
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatal(err)
	}
	position := document.Components.Schemas["Position"]
	for _, name := range []string{"adoptionStatus", "statusKnown", "adoptionLabel", "adoptionReason", "included",
		"excluded", "candidate", "designationKnown", "coveringBlock", "storedExitEvidence"} {
		if _, ok := position.Properties[name]; !ok {
			t.Errorf("Position schema lacks %q", name)
		}
	}
	management := document.Components.Schemas["PositionManagementDescriptor"]
	for _, name := range []string{"desired", "effective", "effectiveKnown", "blockSource"} {
		if _, ok := management.Properties[name]; !ok {
			t.Errorf("PositionManagementDescriptor schema lacks %q", name)
		}
	}
	if _, ok := document.Components.Schemas["AdoptionSettings"]; !ok {
		t.Error("OpenAPI lacks AdoptionSettings")
	}
	for path := range document.Paths {
		if strings.Contains(strings.ToLower(path), "reconcile") {
			t.Errorf("OpenAPI exposes reconcile authority at %q", path)
		}
	}
	for _, schema := range []string{"ReconcileBlock", "StoredExitEvidence"} {
		body, ok := document.Components.Schemas[schema]
		if !ok {
			t.Errorf("OpenAPI lacks %s", schema)
			continue
		}
		for _, forbidden := range []string{"accountRef", "capability", "token", "command"} {
			if _, leaked := body.Properties[forbidden]; leaked {
				t.Errorf("%s exposes forbidden property %s", schema, forbidden)
			}
		}
	}
}

func localReferenceExists(document map[string]any, ref string) bool {
	if !strings.HasPrefix(ref, "#/") {
		return false
	}
	var current any = document
	for _, raw := range strings.Split(strings.TrimPrefix(ref, "#/"), "/") {
		key := strings.ReplaceAll(strings.ReplaceAll(raw, "~1", "/"), "~0", "~")
		object, ok := current.(map[string]any)
		if !ok {
			return false
		}
		current, ok = object[key]
		if !ok {
			return false
		}
	}
	return true
}
