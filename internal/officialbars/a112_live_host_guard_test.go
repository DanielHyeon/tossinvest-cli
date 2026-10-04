package officialbars

// a112 8.2(Manager 판정 2026-10-04) — 「시험은 실호스트에 POST 할 수 없다」 를 이 패키지에서 기계로 지킨다. 8.2 census 에서 시험 폐포가
// net/http 와 공식 클라이언트를 함께 들여오는 레인/증거 패키지는 officialbars 하나였다.
//
// 막는 길과 그 한계:
//   - TestMain 이 http.DefaultTransport 를 testenv.Guard 로 바꾼다 — http.Get · Transport 없는 http.Client 로 나가는 요청이 Toss 호스트면
//     거절되고, 한 건이라도 있으면 패키지 시험이 실패한다.
//   - **공식 클라이언트는 자기 Transport 를 만든다**(internal/official client.go New → newOfficialHTTPClient) — DefaultTransport 가드가
//     그 길은 못 본다. 그래서 시험 파일의 official.New 호출은 전부 WithBaseURL 과 WithHTTPClient 를 함께 넘겨야 한다(기본 실호스트 ·
//     기본 Transport 둘 다 대체) — 아래 AST census.
//   - 시험 파일의 문자열 리터럴에 실호스트 도메인이 없다 · 주문 변경자 이름을 부르지 않는다.

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/testenv"
)

var (
	a112BlockedMu sync.Mutex
	a112Blocked   []string
)

func TestMain(m *testing.M) {
	previous := http.DefaultTransport
	http.DefaultTransport = &testenv.Guard{Base: previous, OnBlock: func(err *testenv.ErrRealHost) {
		a112BlockedMu.Lock()
		a112Blocked = append(a112Blocked, err.Error())
		a112BlockedMu.Unlock()
	}}
	code := m.Run()
	http.DefaultTransport = previous
	if len(a112Blocked) != 0 {
		fmt.Fprintf(os.Stderr, "officialbars tests reached a real Toss host %d time(s):\n%s\n", len(a112Blocked), strings.Join(a112Blocked, "\n"))
		code = 1
	}
	os.Exit(code)
}

// TestMain 이 실제로 가드를 깔았는지(시험이 TestMain 없이 돌거나 누가 Transport 를 되돌리면 여기서 멈춤).
func TestTheDefaultTransportIsTheRealHostGuardDuringTheseTests(t *testing.T) {
	guard, ok := http.DefaultTransport.(*testenv.Guard)
	if !ok {
		t.Fatalf("http.DefaultTransport is %T — TestMain did not install the real-host guard", http.DefaultTransport)
	}
	request, err := http.NewRequest(http.MethodPost, "https://openapi.tossinvest.com/api/v1/orders", nil)
	if err != nil {
		t.Fatal(err)
	}
	// 양성 대조: 가드에 OnBlock 없이 같은 Base 로 하나를 더 세워 실호스트 POST 가 거절되는지 잰다(설치된 가드의 기록은 오염하지 않음).
	probe := &testenv.Guard{Base: guard.Base}
	if _, err := probe.RoundTrip(request); err == nil {
		t.Fatal("the guard let a POST to a real Toss host through")
	}
}

func TestOfficialBarsTestsNeverPointAnOfficialClientAtItsDefaults(t *testing.T) {
	files, err := filepath.Glob("*_test.go")
	if err != nil {
		t.Fatal(err)
	}
	constructions, scanned := 0, 0
	for _, path := range files {
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		scanned++
		local := ""
		for _, spec := range file.Imports {
			if value, _ := strconv.Unquote(spec.Path.Value); value == a112OfficialPath {
				local = "official"
				if spec.Name != nil {
					local = spec.Name.Name
				}
				if local == "." || local == "_" {
					t.Errorf("%s imports internal/official as %q — the census below cannot see its calls", path, local)
				}
			}
		}
		ast.Inspect(file, func(node ast.Node) bool {
			switch value := node.(type) {
			case *ast.SelectorExpr:
				if a112OfficialMutators[value.Sel.Name] {
					t.Errorf("%s: names the order mutator %s", fset.Position(value.Pos()), value.Sel.Name)
				}
			case *ast.BasicLit:
				if value.Kind == token.STRING {
					text, _ := strconv.Unquote(value.Value)
					for _, suffix := range testenv.RealHostSuffixes() {
						// 이 파일 자신의 양성 대조 URL 하나는 가드를 재려고 일부러 쓴다.
						if strings.Contains(text, suffix) && !(path == "a112_live_host_guard_test.go" && strings.HasPrefix(text, "https://openapi.tossinvest.com/")) {
							t.Errorf("%s: string literal names a real Toss host (%q)", fset.Position(value.Pos()), text)
						}
					}
				}
			case *ast.CallExpr:
				selector, ok := value.Fun.(*ast.SelectorExpr)
				if !ok || local == "" {
					return true
				}
				if pkg, ok := selector.X.(*ast.Ident); !ok || pkg.Name != local || selector.Sel.Name != "New" {
					return true
				}
				constructions++
				options := map[string]bool{}
				for _, arg := range value.Args {
					if call, ok := arg.(*ast.CallExpr); ok {
						if option, ok := call.Fun.(*ast.SelectorExpr); ok {
							if pkg, ok := option.X.(*ast.Ident); ok && pkg.Name == local {
								options[option.Sel.Name] = true
							}
						}
					}
				}
				if !options["WithBaseURL"] || !options["WithHTTPClient"] {
					t.Errorf("%s: official.New without both WithBaseURL and WithHTTPClient — the client would dial the real API with its own transport, past the DefaultTransport guard",
						fset.Position(value.Pos()))
				}
			}
			return true
		})
	}
	if scanned == 0 || constructions == 0 {
		t.Fatalf("scanned %d test files, %d official.New calls — the census read nothing (producer_test.go builds one)", scanned, constructions)
	}
}
