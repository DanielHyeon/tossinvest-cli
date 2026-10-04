package strategyprojection

// a112 8.2(Manager 판정 2026-10-04): 레인 투영은 표준 라이브러리만으로 선다 — 그래서 어떤 능력에도 닿을 길이 없다. 그 사실을 두 층으로
// 못 박는다:
//   ① 직접 import 허용 목록(표준 라이브러리도 이름 하나씩 — 「표준이면 다 허용」은 net/http · os/exec 를 함께 연다).
//   ② 폐포 걸음 넷(생산 · 시험 이진 × 무태그 · 태그)에 능력 패키지가 없다 — 생산 걸음에는 이 모듈의 패키지가 자기 자신뿐이다.

import (
	"go/parser"
	"go/token"
	"io/fs"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/testenv"
)

const a112ProjectionPath = testenv.ModulePath + "internal/strategyprojection"

func TestTheProjectionImportsOnlyItsNamedStandardLibrary(t *testing.T) {
	allowed := map[string]bool{
		"context": true, "crypto/sha256": true, "encoding/hex": true, "errors": true, "fmt": true,
		"strings": true, "sync": true, "time": true, "unicode": true, "unicode/utf8": true,
	}
	fset := token.NewFileSet()
	packages, err := parser.ParseDir(fset, ".", func(info fs.FileInfo) bool { return !strings.HasSuffix(info.Name(), "_test.go") }, parser.ImportsOnly)
	if err != nil {
		t.Fatal(err)
	}
	scanned := 0
	var unexpected []string
	for _, pkg := range packages {
		for path, file := range pkg.Files {
			scanned++
			for _, spec := range file.Imports {
				value, err := strconv.Unquote(spec.Path.Value)
				if err != nil {
					t.Fatal(err)
				}
				if !allowed[value] {
					unexpected = append(unexpected, path+" -> "+value)
				}
			}
		}
	}
	if scanned == 0 {
		t.Fatal("no production file was scanned, so this allow-list proves nothing")
	}
	if len(unexpected) != 0 {
		sort.Strings(unexpected)
		t.Fatalf("the projection imports outside its allow-list: %v", unexpected)
	}
}

func TestTheProjectionClosureReachesNoMutationCapability(t *testing.T) {
	for _, mode := range testenv.WalkModes() {
		t.Run(mode.Name, func(t *testing.T) {
			graph := testenv.ListDeps(t, ".", mode)
			roots := graph.Roots(a112ProjectionPath)
			if len(roots) == 0 {
				t.Fatalf("the walk has no root for %s", a112ProjectionPath)
			}
			reached := graph.Reachable(roots, nil)
			if found := testenv.ForbiddenIn(reached); len(found) != 0 {
				t.Fatalf("the projection closure (%s) reaches a mutation capability:\n%v", mode.Name, found)
			}
			if mode.Tests {
				return // 시험 이진은 testenv 등 시험 부품을 들여온다 — 위 금지 대조로 충분
			}
			// 생산 걸음: 이 모듈에서 닿는 패키지는 자기 자신뿐이다(양성 대조 겸 — 걸음이 자기 자신을 읽었다).
			var module []string
			for name := range reached {
				if strings.HasPrefix(name, testenv.ModulePath) {
					module = append(module, name)
				}
			}
			sort.Strings(module)
			if len(module) != 1 || module[0] != a112ProjectionPath {
				t.Fatalf("%s: module packages in the production closure = %v, want only %s", mode.Name, module, a112ProjectionPath)
			}
		})
	}
}
