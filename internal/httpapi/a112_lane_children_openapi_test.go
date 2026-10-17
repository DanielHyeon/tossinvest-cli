package httpapi

// a112 태스크 7.3 — OpenAPI 문서가 투영의 additive 자식 `lanes[8]` · `coordinators[2]` 를 **정확히 같은 이름**으로 적는지 잰다.
// 투영 스키마는 additionalProperties=false 라 문서가 이름을 모르면 문서상 응답이 무효가 된다. 이름 목록은 문서에서 손으로 다시 적지 않고
// strategyprojection 의 JSON 이름 목록(그 패키지 시험이 실제 직렬화와 대조)에서 읽는다.

import (
	"encoding/json"
	"os"
	"reflect"
	"sort"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyprojection"
)

func TestOpenAPIDocumentsTheLaneAndCoordinatorChildrenByTheirExactNames(t *testing.T) {
	raw, err := os.ReadFile("../../docs/api/openapi-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	type schema struct {
		AdditionalProperties *bool                      `json:"additionalProperties"`
		Required             []string                   `json:"required"`
		Properties           map[string]json.RawMessage `json:"properties"`
		MinItems             *int                       `json:"minItems"`
		MaxItems             *int                       `json:"maxItems"`
	}
	var document struct {
		Components struct {
			Schemas map[string]schema `json:"schemas"`
		} `json:"components"`
	}
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatal(err)
	}
	projection := document.Components.Schemas["StrategyRuntimeProjection"]
	for _, field := range []string{"lanes", "coordinators"} {
		if _, ok := projection.Properties[field]; !ok {
			t.Errorf("projection schema lacks %s", field)
		}
	}
	for _, check := range []struct {
		schema string
		fields []string
	}{
		{"StrategyRuntimeLaneRuntime", strategyprojection.LaneJSONFields()},
		{"StrategyRuntimeCoordinator", strategyprojection.CoordinatorJSONFields()},
		{"StrategyRuntimeSelectedScope", strategyprojection.SelectedScopeJSONFields()},
	} {
		item, ok := document.Components.Schemas[check.schema]
		if !ok {
			t.Errorf("OpenAPI lacks %s", check.schema)
			continue
		}
		properties := make([]string, 0, len(item.Properties))
		for name := range item.Properties {
			properties = append(properties, name)
		}
		required := append([]string(nil), item.Required...)
		want := append([]string(nil), check.fields...)
		sort.Strings(properties)
		sort.Strings(required)
		sort.Strings(want)
		if item.AdditionalProperties == nil || *item.AdditionalProperties ||
			!reflect.DeepEqual(properties, want) || !reflect.DeepEqual(required, want) {
			t.Errorf("%s: properties=%v required=%v strict=%v, want exactly %v", check.schema, properties, required, item.AdditionalProperties, want)
		}
	}
	// 개수 고정: lanes 8 · coordinators 2.
	var counts struct {
		Lanes        schema `json:"lanes"`
		Coordinators schema `json:"coordinators"`
	}
	lanesRaw, coordinatorsRaw := projection.Properties["lanes"], projection.Properties["coordinators"]
	if err := json.Unmarshal(lanesRaw, &counts.Lanes); err != nil || counts.Lanes.MinItems == nil || *counts.Lanes.MinItems != 8 ||
		counts.Lanes.MaxItems == nil || *counts.Lanes.MaxItems != 8 {
		t.Errorf("lanes schema=%s, want exactly 8 items", lanesRaw)
	}
	if err := json.Unmarshal(coordinatorsRaw, &counts.Coordinators); err != nil || counts.Coordinators.MinItems == nil ||
		*counts.Coordinators.MinItems != 2 || counts.Coordinators.MaxItems == nil || *counts.Coordinators.MaxItems != 2 {
		t.Errorf("coordinators schema=%s, want exactly 2 items", coordinatorsRaw)
	}
}
