package httpapi

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestPositionExitLineReferenceUsesCamelCaseAndNoAuthorityFields(t *testing.T) {
	value := Position{Market: "US", Symbol: "IONQ", ExitLineReference: &ExitLineReference{
		Kind: "ADOPTION_PLAN", Label: "기준선 미생성", EffectiveKnown: false,
		Baseline: "—", InitialStop: "—", StopPercent: "3%", Basis: "가격은 편입 시 확정",
	}}
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	body := string(raw)
	for _, want := range []string{`"exitLineReference"`, `"effectiveKnown":false`, `"stopPercent":"3%"`} {
		if !strings.Contains(body, want) {
			t.Errorf("JSON lacks %s: %s", want, body)
		}
	}
	for _, forbidden := range []string{"ExitLineReference", "EffectiveKnown", "capability", "command", "accountRef"} {
		if strings.Contains(body, forbidden) {
			t.Errorf("JSON leaks %q: %s", forbidden, body)
		}
	}
}

func TestOpenAPIIncludesNonActionableExitLineReference(t *testing.T) {
	raw, err := os.ReadFile("../../docs/api/openapi-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var document struct {
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
	if _, ok := position.Properties["exitLineReference"]; !ok {
		t.Fatal("Position schema lacks exitLineReference")
	}
	reference, ok := document.Components.Schemas["ExitLineReference"]
	if !ok {
		t.Fatal("OpenAPI lacks ExitLineReference")
	}
	for _, forbidden := range []string{"accountRef", "capability", "token", "command"} {
		if _, leaked := reference.Properties[forbidden]; leaked {
			t.Errorf("ExitLineReference exposes forbidden property %s", forbidden)
		}
	}
}
