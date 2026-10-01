package strategyproposal

// a112 태스크 6.4 조건 ③(Manager 판정 2026-10-01): 벽 자체의 핀. breakout 첫 레그 권한의 잔여(손절-종결 → 재시작 → 새 매니페스트 CampaignID →
// 같은 setup 둘째 첫 레그)는 오늘 이 벽 뒤에 있다 — 생산 breakout 입력은 전부 `ErrBreakoutEvidenceUnavailable` 로 거절된다. 벽을 여는(해제) 편집은
// 이 이름 있는 시험을 뒤집어야 한다. 해제 로트는 B(SetupID 계보 결속) 또는 C(consumed-setup 원장 기록) 착지 전 해제 불가(tasks 6.4 · ROADMAP).
//
// 두 축: (1) 거절 — 등록된 breakout 레인 **전부**(손 목록이 아니라 서술자 두 출처에서 열거)가 어떤 scope 로도 벽 사유로 거절된다.
// (2) 우회 생산자 0 — breakout LaneInput 을 만드는 길(`strategyflow.BreakoutKR/US` 호출 · `laneBreakout*` 종류 리터럴)이 저장소의 시험 아닌 Go 파일
// 어디에도 생산자를 갖지 않는다(생성자 선언 두 자리 자신만). 새 생산자가 벽을 돌아가면 (2)가 뒤집힌다.

import (
	"context"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/breakoutlane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/officialfx"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyevidence"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyflow"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

func TestEveryProductionBreakoutLaneInputIsRefusedAtTheWall(t *testing.T) {
	// 출처 1: breakoutlane 서술자. 출처 2: strategyflow 등록 서술자 중 생산 라우터가 breakout 가족으로 보는 것. 둘이 같아야 한다.
	fromLane := map[string]bool{}
	for _, d := range breakoutlane.Descriptors() {
		fromLane[d.LaneID] = true
	}
	fromFlow := map[string]bool{}
	for _, d := range strategyflow.Descriptors() {
		if family, ok := strategyrouter.ProductionLaneFamily(d.Market, d.LaneID); ok && family == strategyrouter.FamilyBreakoutRetest {
			fromFlow[d.LaneID] = true
		}
	}
	if len(fromLane) != 2 || strings.Join(a112Keys(fromLane), ",") != strings.Join(a112Keys(fromFlow), ",") {
		t.Fatalf("breakout lanes: breakoutlane %v vs strategyflow/router %v — want the same two", a112Keys(fromLane), a112Keys(fromFlow))
	}
	full := productionScope{Symbol: "005930", CampaignID: "campaign-wall", SnapshotID: "snapshot-wall", SnapshotDigest: "sha256:" + strings.Repeat("a", 64)}
	for lane := range fromLane {
		for name, scope := range map[string]productionScope{"empty": {}, "populated": full} {
			scope.LaneID = lane
			input, weekly, err := buildLaneInput(context.Background(), ProductionConfig{}, scope, strategyevidence.Snapshot{ID: scope.SnapshotID, Digest: scope.SnapshotDigest},
				officialfx.Evidence{}, nil)
			if !errors.Is(err, ErrBreakoutEvidenceUnavailable) || weekly != nil || !reflect.ValueOf(input).IsZero() {
				t.Errorf("lane %s (%s scope): err=%v weekly=%v input-zero=%v — want the breakout wall", lane, name, err, weekly != nil, reflect.ValueOf(input).IsZero())
			}
		}
	}
}

func TestNoNonTestCodeBuildsABreakoutLaneInputAroundTheWall(t *testing.T) {
	root := a112RepoRoot(t)
	calls, literals, files := 0, 0, 0
	for _, dir := range []string{"internal", "cmd"} {
		err := filepath.WalkDir(filepath.Join(root, dir), func(path string, entry fs.DirEntry, err error) error {
			if err != nil || entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return err
			}
			files++
			file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
			if err != nil {
				return err
			}
			inFlow := file.Name.Name == "strategyflow"
			ast.Inspect(file, func(n ast.Node) bool {
				switch x := n.(type) {
				case *ast.FuncDecl:
					// 생성자 선언 자신(BreakoutKR/US 의 몸통 안 리터럴)은 생산자가 아니다 — 몸통을 건너뛴다.
					if inFlow && x.Recv == nil && (x.Name.Name == "BreakoutKR" || x.Name.Name == "BreakoutUS") {
						return false
					}
				case *ast.CallExpr:
					switch fun := x.Fun.(type) {
					case *ast.SelectorExpr:
						if id, ok := fun.X.(*ast.Ident); ok && id.Name == "strategyflow" && (fun.Sel.Name == "BreakoutKR" || fun.Sel.Name == "BreakoutUS") {
							calls++
						}
					case *ast.Ident:
						if inFlow && (fun.Name == "BreakoutKR" || fun.Name == "BreakoutUS") {
							calls++
						}
					}
				case *ast.ValueSpec:
					// 종류 상수의 선언 자신은 사용이 아니다.
					for _, name := range x.Names {
						if name.Name == "laneBreakoutKR" || name.Name == "laneBreakoutUS" {
							return false
						}
					}
				case *ast.Ident:
					if inFlow && (x.Name == "laneBreakoutKR" || x.Name == "laneBreakoutUS") {
						literals++
					}
				}
				return true
			})
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	if files < 500 {
		t.Fatalf("scanned %d non-test Go files — the walk did not reach the repository", files)
	}
	// laneBreakout* 의 생산 사용은 등록표(registry) 의 종류 대응 두 자리뿐이다(2026-10-01 실측). 그 밖의 사용은 새 LaneInput 생산자다.
	if calls != 0 || literals != 2 {
		t.Fatalf("breakout LaneInput producers outside the constructors: calls=%d, laneBreakout* uses=%d (want 0 calls and the 2 registry uses)", calls, literals)
	}
}

func a112Keys(set map[string]bool) []string {
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func a112RepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	for d := dir; ; d = filepath.Dir(d) {
		if _, err := os.Stat(filepath.Join(d, "go.mod")); err == nil {
			return d
		}
		if filepath.Dir(d) == d {
			t.Fatal("go.mod not found")
		}
	}
}
