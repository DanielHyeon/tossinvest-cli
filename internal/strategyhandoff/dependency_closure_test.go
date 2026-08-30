package strategyhandoff

import (
	"encoding/json"
	"go/parser"
	"go/token"
	"io"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"testing"
)

const modulePath = "github.com/JungHoonGhae/tossinvest-cli/"

// 이 경계는 아무것도 바꾸지 않는다. 그 약속은 주석이 아니라 import 목록이
// 지킨다. 태스크 5.5 가 요구하는 "compile/dependency 수준의 증명"이 여기다.
//
// 직접 import 를 허용 목록으로 본다. 금지 목록이 아니라 허용 목록인 이유는,
// 금지 목록은 빠뜨린 패키지를 조용히 통과시키기 때문이다. 첫 판본이 정확히
// 그랬다 — 금지 목록을 dispatch cycle 의 import 에서 베껴 왔고, 그것은 그
// 파일의 영수증이지 이 모듈의 변경 표면이 아니어서 internal/official 이 통째로
// 빠져 있었다.
func TestTheHandoffSeamImportsNothingOutsideItsAllowedClosure(t *testing.T) {
	allowed := map[string]bool{
		modulePath + "internal/strategyflow": true,
	}
	fset := token.NewFileSet()
	packages, err := parser.ParseDir(fset, ".", productionOnly, parser.ImportsOnly)
	if err != nil {
		t.Fatalf("parse package directory: %v", err)
	}
	scanned := 0
	unexpected := make([]string, 0)
	for name, pkg := range packages {
		if name != "strategyhandoff" {
			t.Fatalf("this directory declares package %q", name)
		}
		for path, file := range pkg.Files {
			scanned++
			for _, spec := range file.Imports {
				value, err := strconv.Unquote(spec.Path.Value)
				if err != nil {
					t.Fatalf("%s: unquote %s: %v", path, spec.Path.Value, err)
				}
				if !allowed[value] {
					unexpected = append(unexpected, path+" -> "+value)
				}
			}
		}
	}
	if scanned == 0 {
		t.Fatal("no production file was scanned, so this closure proves nothing")
	}
	if len(unexpected) != 0 {
		sort.Strings(unexpected)
		t.Fatalf("the handoff seam imports something outside its allowed closure: %v", unexpected)
	}
}

// 직접 import 만 보면 한 다리 건너 들어오는 능력을 놓친다. 이 검사는 전이
// 의존 전체를 훑는다.
//
// 양성 대조가 함께 있다: 걸음이 비었거나 잘렸으면 위의 단언들이 **틀린
// 이유로** 통과한다. deps_test.go 의 엔진 검사가 같은 이유로 양성 대조를 달고
// 있다.
func TestTheHandoffSeamTransitivelyReachesNoMutationCapability(t *testing.T) {
	command := exec.Command("go", "list", "-json", "-deps", ".")
	raw, err := command.Output()
	if err != nil {
		t.Fatalf("go list: %v", err)
	}
	// 브로커·원장·주문 의도를 바꿀 수 있는 패키지 전부. 손으로 적었지만
	// dispatch cycle 의 import 가 아니라 이 모듈에서 주문·원장 쓰기를 실제로
	// 노출하는 패키지를 골라 적었다.
	forbidden := []string{
		"internal/official",           // PlaceOrder/CancelOrder/ModifyOrder + 조건부 주문 3종
		"internal/hybrid",             // 공식 클라이언트가 없으면 WTS 로 주문을 흘린다
		"internal/client",             // WTS 세션 주문 변경자
		"internal/trading",            // Broker 인터페이스와 그 구현
		"internal/orderintent",        // 주문 의도 생성
		"internal/execgw",             // 실행 게이트웨이
		"internal/journal",            // 원장 쓰기와 lease
		"internal/ops",                // 운영 토글 flip
		"internal/protectionofficial", // 보호 주문 배치
		"internal/verifylive",         // 실계좌 검증 경로
		"internal/config",             // 운영 설정 쓰기
		"internal/app",                // CLI 배선 — 위를 전부 끌고 온다
	}
	deps := make(map[string]bool)
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	for {
		var entry struct{ ImportPath string }
		if err := decoder.Decode(&entry); err != nil {
			if err == io.EOF {
				break
			}
			t.Fatal(err)
		}
		deps[entry.ImportPath] = true
		for _, name := range forbidden {
			if entry.ImportPath == modulePath+name || strings.HasPrefix(entry.ImportPath, modulePath+name+"/") {
				t.Errorf("the handoff seam transitively reaches a mutation capability: %s", entry.ImportPath)
			}
		}
	}

	// 양성 대조. 이 두 줄이 없으면 `go list` 가 아무것도 못 읽은 날에도 위가
	// 통과한다.
	if len(deps) == 0 {
		t.Fatal("go list returned no dependency packages, so the walk above read nothing")
	}
	for _, required := range []string{
		modulePath + "internal/strategyflow",
		modulePath + "internal/strategyhandoff",
	} {
		if !deps[required] {
			names := make([]string, 0, len(deps))
			for name := range deps {
				if strings.HasPrefix(name, modulePath) {
					names = append(names, strings.TrimPrefix(name, modulePath))
				}
			}
			sort.Strings(names)
			t.Fatalf("expected %s in the dependency walk; it found only %v", required, names)
		}
	}
}
