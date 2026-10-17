package httpapi

// a112 7.3.1 SHADOW(브리프 v3.3 §9) — OpenAPI 의 레인 runtime · shadowOutcome enum 이 projection 어휘 함수와 **같다**.
//
// 앞 판(7.3)에는 runtime enum 을 고정하는 시험이 없었다(보이스 3 P1-5). 문서는 손으로 다시 적지 않고 projection 의 어휘 함수와 대조한다.
// shadowOutcome 은 required 라 언제나 직렬화되고 null 이 허용된다 — 핀 0(오늘) = runtime UNOBSERVED · shadowOutcome null.

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyprojection"
)

func TestOpenAPIRuntimeAndShadowOutcomeEnumsAreTheProjectionVocabulary(t *testing.T) {
	raw, err := os.ReadFile("../../docs/api/openapi-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var document struct {
		Components struct {
			Schemas map[string]struct {
				Required   []string `json:"required"`
				Properties map[string]struct {
					Type json.RawMessage `json:"type"`
					Enum []any           `json:"enum"`
				} `json:"properties"`
			} `json:"schemas"`
		} `json:"components"`
	}
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatal(err)
	}
	lane, ok := document.Components.Schemas["StrategyRuntimeLaneRuntime"]
	if !ok {
		t.Fatal("OpenAPI lacks StrategyRuntimeLaneRuntime")
	}
	strings := func(values []any) ([]string, bool) {
		out, null := []string{}, false
		for _, value := range values {
			if value == nil {
				null = true
				continue
			}
			out = append(out, value.(string))
		}
		return out, null
	}
	runtime, runtimeNull := strings(lane.Properties["runtime"].Enum)
	if !reflect.DeepEqual(runtime, strategyprojection.LaneRuntimes()) || runtimeNull {
		t.Errorf("runtime enum=%v (null=%v), want exactly %v", runtime, runtimeNull, strategyprojection.LaneRuntimes())
	}
	outcome, outcomeNull := strings(lane.Properties["shadowOutcome"].Enum)
	if !reflect.DeepEqual(outcome, strategyprojection.LaneShadowOutcomes()) || !outcomeNull {
		t.Errorf("shadowOutcome enum=%v (null=%v), want exactly %v plus null", outcome, outcomeNull, strategyprojection.LaneShadowOutcomes())
	}
	required := false
	for _, name := range lane.Required {
		required = required || name == "shadowOutcome"
	}
	if !required {
		t.Error("shadowOutcome must be required (always serialized; null when the lane is not in SHADOW)")
	}
}
