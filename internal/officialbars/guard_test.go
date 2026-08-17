package officialbars

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

// packageFiles는 이 꾸러미의 .go 파일 전부다.
func packageFiles(t *testing.T) []string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate this test file")
	}
	files, err := filepath.Glob(filepath.Join(filepath.Dir(thisFile), "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("no source files found")
	}
	return files
}

func productionFiles(t *testing.T) []string {
	t.Helper()
	var out []string
	for _, file := range packageFiles(t) {
		if !strings.HasSuffix(file, "_test.go") {
			out = append(out, file)
		}
	}
	if len(out) == 0 {
		t.Fatal("no production files found")
	}
	return out
}

func parseFile(t *testing.T, path string) *ast.File {
	t.Helper()
	parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		t.Fatalf("parsing %s: %v", path, err)
	}
	return parsed
}

// TestProductionImportsStayInsideTheAllowlist는 이 꾸러미가 자기 자리에서 벗어나지
// 못하게 막는다. 엔진·주문·저널·토글은 물론 net/http도 여기서 보이면 안 된다.
func TestProductionImportsStayInsideTheAllowlist(t *testing.T) {
	t.Parallel()
	allowed := map[string]bool{
		"github.com/JungHoonGhae/tossinvest-cli/internal/clock":            true,
		"github.com/JungHoonGhae/tossinvest-cli/internal/official":         true,
		"github.com/JungHoonGhae/tossinvest-cli/internal/scheduler":        true,
		"github.com/JungHoonGhae/tossinvest-cli/internal/strategyevidence": true,
	}
	for _, file := range productionFiles(t) {
		for _, spec := range parseFile(t, file).Imports {
			path, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				t.Fatalf("%s: unreadable import %s", file, spec.Path.Value)
			}
			if path == "net/http" {
				t.Fatalf("%s imports net/http; transport belongs to internal/official", filepath.Base(file))
			}
			// 표준 라이브러리는 첫 마디에 점이 없다. 그 밖의 것은 허용 목록에만 있어야 한다.
			if strings.Contains(strings.SplitN(path, "/", 2)[0], ".") && !allowed[path] {
				t.Fatalf("%s imports %s, which is outside the allowlist", filepath.Base(file), path)
			}
		}
	}
}

// TestProductionWritesEvidenceOnlyThroughTheCombinedConstructor는 L1a가 남긴 잔여
// 위험을 막는다. 손으로 만든 Header를 NewEnvelope에 넣으면 머리말과 본문이 서로 다른
// 이야기를 하는 줄을 쓸 수 있고, 그 줄은 세션 읽기를 영영 막는다.
func TestProductionWritesEvidenceOnlyThroughTheCombinedConstructor(t *testing.T) {
	t.Parallel()
	for _, file := range productionFiles(t) {
		parsed := parseFile(t, file)
		ast.Inspect(parsed, func(node ast.Node) bool {
			selector, ok := node.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			pkg, ok := selector.X.(*ast.Ident)
			if !ok || pkg.Name != "strategyevidence" {
				return true
			}
			if selector.Sel.Name == "NewEnvelope" || selector.Sel.Name == "Header" {
				t.Fatalf("%s reaches for strategyevidence.%s; only the combined constructors may build evidence",
					filepath.Base(file), selector.Sel.Name)
			}
			return true
		})
		source, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		for _, forbidden := range []string{"NewEnvelope", "Header{"} {
			if strings.Contains(string(source), forbidden) {
				t.Fatalf("%s spells %q", filepath.Base(file), forbidden)
			}
		}
	}
}

// TestPackageCarriesNoMeasurementSeamIdentifier는 M-B0 계측 이음매의 이름이 이
// 꾸러미 어디에도 나타나지 않게 한다(저장소 전역 가드와 같은 규칙).
func TestPackageCarriesNoMeasurementSeamIdentifier(t *testing.T) {
	t.Parallel()
	// needle을 통째로 적으면 이 파일 자신이 금지된 바이트열을 갖게 되므로 이어 붙인다.
	needle := "a112" + "MBUS"
	for _, file := range packageFiles(t) {
		source, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(source), needle) {
			t.Fatalf("%s references the measurement seam", filepath.Base(file))
		}
	}
}
