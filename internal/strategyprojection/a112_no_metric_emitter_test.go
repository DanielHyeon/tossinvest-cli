package strategyprojection

// a112 태스크 7.4 — 메트릭 label cardinality 계약을 **지금 지킬 자리가 없다**는 사실을 금지 가드로 바꾼다(Manager 판정 (A), 2026-10-01).
//
// ── label 계약(이 가드를 뒤집는 사람이 지나치면 안 되는 문장) ─────────────────────────────────────────────────────────────
//   - 전략 런타임 메트릭의 label 은 **고정 다섯 차원**뿐이다: market · family · lane(lane_id) · version(lane_version) · reason.
//     값은 닫힌 열거형이어야 한다(골든 descriptors 의 여덟 열쇠, 골든 refusal_enums, 엔진 사유 상수 — 이 패키지의 어휘 함수들).
//   - symbol · setup · candidate 식별자는 label 로 쓰지 않는다 — 골든 four-family-runtime-v1 의
//     `operator_compatibility.unbounded_metric_labels_forbidden: ["setup","candidate","symbol"]`, 스펙 「Setup/candidate/symbol 처럼
//     unbounded 값은 metric label 로 사용해서는 안 되며 (MUST NOT)」. 그 식별자는 로그 · journal 질의와 읽기 전용 상태 payload
//     (예: coordinators[].selected[].symbol)에만 산다.
//
// ── 왜 가드인가 ─────────────────────────────────────────────────────────────────────────────────────────────────────
// 2026-10-01 실측: 생산 Go 에 메트릭 방출기가 0 이다(prometheus · expvar 0, otel/metric 은 go.mod indirect — releaseupdate → sigstore
// 경유, 우리 import 0, `/metrics` 경로 0). 묶을 label 이 없으니 「묶었다」고 적으면 거짓이고, 아무것도 안 하면 첫 방출기가 계약 없이
// 들어온다. 그래서 첫 방출기는 **이 시험을 뒤집어야** 들어올 수 있게 한다 — 허용 목록(이름 목록, 지금 비어 있음)에 자기 파일을
// 더하는 편집이 리뷰에 보이고, 그 change 가 위 label 계약을 그 자리에서 코드로 세워야 한다(이 머리말을 옮겨 적는 것이 아니라).

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// a112ForbiddenMetricImports 는 메트릭 방출 API 의 import 경로(정확히 같거나 이 접두사 + "/")다.
var a112ForbiddenMetricImports = []string{
	"expvar",
	"github.com/prometheus/client_golang",
	"github.com/prometheus/client_model",
	"go.opentelemetry.io/otel/metric",
	"go.opentelemetry.io/otel/sdk/metric",
	"go.opentelemetry.io/otel/exporters/prometheus",
	"github.com/VictoriaMetrics/metrics",
	"github.com/rcrowley/go-metrics",
	"github.com/armon/go-metrics",
	"github.com/hashicorp/go-metrics",
	"github.com/DataDog/datadog-go",
}

// a112AllowedMetricEmitters 는 label 계약을 세운 방출기 파일(저장소 상대 경로)이다. 지금 비어 있다 — 첫 항목을 더하는 편집이 계약을
// 세우는 change 의 일부여야 한다.
var a112AllowedMetricEmitters = map[string]bool{}

func TestNoProductionCodeEmitsMetricsWithoutTheFixedLabelContract(t *testing.T) {
	scanned, violations := 0, []string{}
	for _, root := range []string{"../../internal", "../../cmd"} {
		err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				if name := entry.Name(); name == "testdata" || name == "vendor" || strings.HasPrefix(name, ".") {
					return filepath.SkipDir
				}
				return nil
			}
			if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
			if err != nil {
				return err
			}
			scanned++
			rel := strings.TrimPrefix(filepath.ToSlash(path), "../../")
			for _, spec := range file.Imports {
				imported, err := strconv.Unquote(spec.Path.Value)
				if err != nil {
					return err
				}
				for _, forbidden := range a112ForbiddenMetricImports {
					if (imported == forbidden || strings.HasPrefix(imported, forbidden+"/")) && !a112AllowedMetricEmitters[rel] {
						violations = append(violations, rel+" imports "+imported)
					}
				}
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	// 표본이 0 이면 전칭 판정이 자동으로 참이 된다 — 실제로 훑었는지 하한으로 확인한다(2026-10-01 실측 838 파일 — internal · cmd 의 비시험 Go).
	if scanned < 500 {
		t.Fatalf("scanned only %d production Go files — the walk did not reach the tree", scanned)
	}
	sort.Strings(violations)
	if len(violations) != 0 {
		t.Fatalf("metric emitters without the fixed label contract (market/family/lane/version/reason; no symbol/setup/candidate — "+
			"see this file's header): %v", violations)
	}
}
