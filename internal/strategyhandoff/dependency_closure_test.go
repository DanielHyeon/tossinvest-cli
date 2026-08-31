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
		"errors":                             true,
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
	// `-test` 를 붙이면 이 패키지의 시험 바이너리가 들여오는 것까지 들어온다.
	// 붙이지 않은 앞선 판본은 "어떤 자동 시험도 실계좌 주문 경로에 닿지
	// 않는다"는 저장소 규칙을 이 패키지에서 확인하지 않았다.
	for _, arg := range []string{"-deps", "-deps-test"} {
		t.Run(arg, func(t *testing.T) { assertNoMutationCapabilityInWalk(t, arg) }) //nolint:thelper
	}
}

func assertNoMutationCapabilityInWalk(t *testing.T, mode string) {
	args := []string{"list", "-json", "-deps", "."}
	if mode == "-deps-test" {
		args = []string{"list", "-json", "-deps", "-test", "."}
	}
	command := exec.Command("go", args...)
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
	//
	// 앞선 판본은 strategyflow(직접 import)와 자기 자신만 요구했다. 둘 다
	// 깊이 1 이하라서 걸음을 "자신 + 직접 import" 로 좁혀도 통과했다 —
	// 즉 전이성을 전혀 증명하지 못했다. 아래 둘은 strategyflow 의 직접
	// import 도 아니므로, 이 이름들이 보이면 걸음이 실제로 두 단계 넘게
	// 내려갔다는 뜻이다.
	for _, required := range []string{
		modulePath + "internal/strategyflow",
		modulePath + "internal/strategyhandoff",
		modulePath + "internal/candidate",
		modulePath + "internal/domain",
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

// TestOnlyTheEngineImportsThisSeam 은 이 경계를 들여오는 패키지를 고정한다.
//
// 왜 필요한가: `Handoff` 를 만드는 문은 `Admit` 하나뿐이고, `Delivered` 를
// 채우는 문은 `Deliver` 하나뿐이다. 그래서 "경계를 지나지 않은 제안이 주문
// 경로에 닿는" 편집은 반드시 어딘가에서 `Admit` 을 불러야 한다. 엔진
// 패키지 안의 그 자리는 engine 의 TestExactlyOneProductionSiteAdmitsIntoTheSeam
// 이 센다. 그러나 **다른 패키지**가 Admit 을 불러 만든 Handoff 를 엔진에
// 돌려주면 엔진 소스에는 그 토큰이 없고, 그 세기는 아무것도 못 본다.
//
// 여기서 그 나머지를 막는다. Admit 을 부르려면 이 패키지를 import 해야 하고,
// import 는 소스에 적히지 않으면 존재하지 않는다. 두 검사가 함께여야 사슬이
// 닫힌다 — 어느 한쪽만으로는 닫히지 않는다.
func TestOnlyTheEngineImportsThisSeam(t *testing.T) {
	command := exec.Command("go", "list", "-json", modulePath+"...")
	raw, err := command.Output()
	if err != nil {
		t.Fatalf("go list: %v", err)
	}
	seam := modulePath + "internal/strategyhandoff"
	// 이 경계를 들여와도 되는 패키지. 허용 목록이지 금지 목록이 아니다 —
	// 금지 목록은 새로 생긴 패키지를 조용히 통과시킨다.
	allowed := map[string]bool{
		modulePath + "internal/app/engine": true,
		seam:                               true,
	}
	scanned, importers := 0, make([]string, 0, 2)
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	for {
		var entry struct {
			ImportPath                         string
			Imports, TestImports, XTestImports []string
		}
		if err := decoder.Decode(&entry); err != nil {
			if err == io.EOF {
				break
			}
			t.Fatal(err)
		}
		scanned++
		for _, group := range [][]string{entry.Imports, entry.TestImports, entry.XTestImports} {
			for _, path := range group {
				if path != seam {
					continue
				}
				importers = append(importers, entry.ImportPath)
				if !allowed[entry.ImportPath] {
					t.Errorf("%s imports the handoff seam; only the engine may, "+
						"because importing it is the only way to call Admit", entry.ImportPath)
				}
			}
		}
	}
	// 양성 대조 둘. `go list` 가 아무것도 못 읽었거나 이 패키지를 아무도
	// 안 들여오면 위의 단언은 **틀린 이유로** 통과한다.
	if scanned == 0 {
		t.Fatal("go list returned no packages, so the scan above read nothing")
	}
	if len(importers) == 0 {
		t.Fatal("no package imports this seam, so the allow list above proves nothing")
	}
	sort.Strings(importers)
	t.Logf("importers of the handoff seam: %v (scanned %d packages)", importers, scanned)
}
