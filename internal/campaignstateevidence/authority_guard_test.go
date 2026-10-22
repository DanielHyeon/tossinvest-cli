package campaignstateevidence

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const evidenceImport = "github.com/JungHoonGhae/tossinvest-cli/internal/campaignstateevidence"

// offlineAssessment 는 a162 가 설계로 둔 검사 전용 소비자(design 「Input and authority boundary」) — 이 파일 하나만 이름으로 받음.
const (
	offlineAssessment       = "internal/offlinecontinuation/assessment.go"
	offlineAssessmentImport = "github.com/JungHoonGhae/tossinvest-cli/internal/offlinecontinuation"
)

func allowedImport(path string, source []byte) (bool, error) {
	f, err := parser.ParseFile(token.NewFileSet(), path, source, 0)
	if err != nil {
		return false, err
	}
	laneAliases := map[string]bool{}
	for _, spec := range f.Imports {
		p, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			return false, err
		}
		if strings.HasPrefix(p, "github.com/JungHoonGhae/tossinvest-cli/internal/") && (strings.HasSuffix(p, "/continuationlane") || strings.HasSuffix(p, "/reversallane") || strings.HasSuffix(p, "/weeklyvaluelane")) {
			alias := filepath.Base(p)
			if spec.Name != nil {
				alias = spec.Name.Name
			}
			if alias == "." {
				return false, nil
			}
			laneAliases[alias] = true
		}
		if p == evidenceImport && path != "internal/continuationlane/offline_evidence.go" && path != "internal/reversallane/offline_evidence.go" && path != "internal/weeklyvaluelane/offline_evidence.go" && path != offlineAssessment {
			return false, nil
		}
		// 허용한 오프라인 평가를 import 하는 비시험 파일은 캡처 경로를 런타임으로 옮기므로 거절함(a170).
		if p == offlineAssessmentImport {
			return false, nil
		}
	}
	// 오프라인 평가 파일의 lane CaptureOfflineEvidence 호출은 a162 설계가 정한 검사 전용 사용이라 받음.
	if path == offlineAssessment {
		return true, nil
	}
	allowed := true
	ast.Inspect(f, func(n ast.Node) bool {
		sel, ok := n.(*ast.SelectorExpr)
		if ok && sel.Sel.Name == "CaptureOfflineEvidence" {
			if id, ok := sel.X.(*ast.Ident); ok && laneAliases[id.Name] {
				allowed = false
			}
		}
		return true
	})
	return allowed, nil
}
func TestInspectionPackageHasNoRuntimeConsumers(t *testing.T) {
	root := filepath.Join("..", "..")
	for _, dir := range []string{"internal", "cmd", "tools"} {
		err := filepath.WalkDir(filepath.Join(root, dir), func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			source, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			allowed, err := allowedImport(filepath.ToSlash(rel), source)
			if err != nil {
				return err
			}
			if !allowed {
				t.Errorf("unexpected inspection import: %s", rel)
			}
			return nil
		})
		if err != nil && !os.IsNotExist(err) {
			t.Fatal(err)
		}
	}
	for _, src := range []string{"package p; import `" + evidenceImport + "`", "package p; import \"github.com/JungHoonGhae/tossinvest-cli/internal/campaignstate\\u0065vidence\""} {
		ok, err := allowedImport("internal/app/engine/forged.go", []byte(src))
		if err != nil || ok {
			t.Fatal("guard missed raw or escaped import", err)
		}
	}
}

func TestGuardRejectsRuntimeCaptureUse(t *testing.T) {
	for _, imp := range []string{"`github.com/JungHoonGhae/tossinvest-cli/internal/reversallane`", "\"github.com/JungHoonGhae/tossinvest-cli/internal/reversall\\u0061ne\""} {
		src := "package p; import lane " + imp + "; var x = lane.CaptureOfflineEvidence"
		if ok, err := allowedImport("internal/app/engine/forged.go", []byte(src)); err != nil || ok {
			t.Fatalf("guard missed runtime reference: %v", err)
		}
	}
	src := "package p; import . `github.com/JungHoonGhae/tossinvest-cli/internal/continuationlane`; var x = CaptureOfflineEvidence"
	if ok, err := allowedImport("internal/app/engine/forged.go", []byte(src)); err != nil || ok {
		t.Fatalf("guard missed dot import: %v", err)
	}
}

// a170 — a162 오프라인 평가는 설계상 검사 전용 소비자라 이름으로만 받음. 같은 원문이 다른 경로에 있거나
// 런타임 파일이 그 평가 패키지를 import 하면 캡처 경로가 런타임에 닿으므로 거절함.
func TestOfflineAssessmentIsTheOnlyNamedConsumer(t *testing.T) {
	source, err := os.ReadFile(filepath.Join("..", "offlinecontinuation", "assessment.go"))
	if err != nil {
		t.Fatal(err)
	}
	if ok, err := allowedImport(offlineAssessment, source); err != nil || !ok {
		t.Fatalf("named offline assessment refused: %v", err)
	}
	if ok, err := allowedImport("internal/offlinecontinuation/other.go", source); err != nil || ok {
		t.Fatalf("assessment source accepted at another path: %v", err)
	}
	// evidence import 없이 캡처만 부르는 형제 파일도 거절 — 이름 허용이 패키지 전체로 번지지 않는지 따로 잼(변이 M2).
	captureOnly := "package offlinecontinuation; import `github.com/JungHoonGhae/tossinvest-cli/internal/reversallane`; var x = reversallane.CaptureOfflineEvidence"
	if ok, err := allowedImport("internal/offlinecontinuation/other.go", []byte(captureOnly)); err != nil || ok {
		t.Fatalf("capture use accepted in a sibling file: %v", err)
	}
	// 이름 허용은 import 검사 **뒤**에만 서야 함 — 평가 파일이라도 lane 점 import 로 캡처를 부르면 거절(변이 M5).
	dotCapture := "package offlinecontinuation; import . `github.com/JungHoonGhae/tossinvest-cli/internal/continuationlane`; var x = CaptureOfflineEvidence"
	if ok, err := allowedImport(offlineAssessment, []byte(dotCapture)); err != nil || ok {
		t.Fatalf("dot-import capture accepted in the named assessment: %v", err)
	}
	for _, imp := range []string{"`" + offlineAssessmentImport + "`", "\"github.com/JungHoonGhae/tossinvest-cli/internal/offlinec\\u006fntinuation\""} {
		src := "package p; import " + imp
		if ok, err := allowedImport("internal/app/engine/forged.go", []byte(src)); err != nil || ok {
			t.Fatalf("guard missed runtime import of the offline assessment: %v", err)
		}
	}
}
