package httpapi

// a112 L5 — 결정 54가 인용한 부재의 정정.
//
// 그 결정이 "이 숫자는 어떤 운영 표면으로도 나가지 않는다"의 근거로 든 명령이
// `grep -rn "ConfigDigest" internal/httpapi/ cmd/tossctl/` → 0 hits 였다.
// 이 파일은 그 0을 REST 응답에서 실제로 읽어 확인하는 쪽으로 바꾼다.

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyprojection"
)

func TestStrategyRuntimeRESTCarriesTheDigestsTheOperatorMustWriteDown(t *testing.T) {
	config, build := "sha256:"+strings.Repeat("c", 64), "sha256:"+strings.Repeat("b", 64)
	// dormant 쌍이다 — 아무것도 활성화되지 않은 상태가 이 숫자를 찾는 상태다.
	snapshot := strategyprojection.WithRuntimeIdentity(
		strategyprojection.DormantSnapshot(time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC)), config, build)
	handler, err := NewRouter(Options{Reader: contractReader{}, StrategyRuntime: apiStrategyRuntimeStub{snapshot: snapshot}})
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/strategy-runtime", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	// 구조체가 아니라 **발행된 JSON 이름**으로 읽는다. 운영자와 OpenAPI 문서가
	// 보는 것이 이 이름이고, Go 필드명만 맞고 태그가 틀려도 통과하면 안 된다.
	var envelope struct {
		Data struct {
			Runtime struct {
				ConfigDigest *string `json:"configDigest"`
				BuildDigest  *string `json:"buildDigest"`
			} `json:"runtime"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Data.Runtime.ConfigDigest == nil || *envelope.Data.Runtime.ConfigDigest != config ||
		envelope.Data.Runtime.BuildDigest == nil || *envelope.Data.Runtime.BuildDigest != build {
		t.Fatalf("REST 응답이 매니페스트용 digest 를 싣지 않았다: %s", recorder.Body.String())
	}
}

// TestStrategyRuntimeRESTWithoutAnEngineReportsNoDigest 는 REST 가 부재를 가리지
// 않음을 잡는다. reader 가 없으면 이 배포의 값이 아니라 **관측 없음**이다.
func TestStrategyRuntimeRESTWithoutAnEngineReportsNoDigest(t *testing.T) {
	handler, err := NewRouter(Options{Reader: contractReader{}})
	if err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/strategy-runtime", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	var envelope struct {
		Data struct {
			Runtime struct {
				ConfigDigest *string `json:"configDigest"`
				BuildDigest  *string `json:"buildDigest"`
			} `json:"runtime"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Data.Runtime.ConfigDigest != nil || envelope.Data.Runtime.BuildDigest != nil {
		t.Fatalf("엔진이 없는데 digest 를 지어냈다: %s", recorder.Body.String())
	}
}

// TestOpenAPIPublishesTheRuntimeIdentitySchema 는 발행된 계약과 실제 응답이 같은
// 이름·같은 모양임을 잡는다. 문서에 없는 필드를 응답이 실으면 `additionalProperties:
// false` 계약을 우리 스스로 깨는 것이다.
func TestOpenAPIPublishesTheRuntimeIdentitySchema(t *testing.T) {
	raw, err := os.ReadFile("../../docs/api/openapi-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var document struct {
		Components struct {
			Schemas map[string]struct {
				AdditionalProperties *bool                      `json:"additionalProperties"`
				Required             []string                   `json:"required"`
				Properties           map[string]json.RawMessage `json:"properties"`
			} `json:"schemas"`
		} `json:"components"`
	}
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatal(err)
	}

	identity, ok := document.Components.Schemas["StrategyRuntimeIdentity"]
	if !ok {
		t.Fatal("OpenAPI 에 StrategyRuntimeIdentity 가 없다")
	}
	if identity.AdditionalProperties == nil || *identity.AdditionalProperties {
		t.Error("StrategyRuntimeIdentity 가 모르는 필드를 허용한다")
	}
	// 두 값은 함께 있거나 함께 없다. 문서도 그렇게 말해야 한다.
	if got := strings.Join(identity.Required, ","); got != "configDigest,buildDigest" {
		t.Errorf("required=%q; want configDigest,buildDigest", got)
	}

	projection := document.Components.Schemas["StrategyRuntimeProjection"]
	if _, ok := projection.Properties["runtime"]; !ok {
		t.Error("projection 문서에 runtime 이 없다 — 응답은 실어 보내는데 문서는 모른다")
	}
	var runtimeRequired bool
	for _, field := range projection.Required {
		if field == "runtime" {
			runtimeRequired = true
		}
	}
	if !runtimeRequired {
		t.Error("runtime 이 required 가 아니다 — Go 는 이 필드를 항상 직렬화한다")
	}
}
