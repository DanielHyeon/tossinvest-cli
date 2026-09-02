package strategyworker

import (
	"encoding/json"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"testing"
)

const modulePath = "github.com/JungHoonGhae/tossinvest-cli/"

// productionOnly 는 시험 파일을 뺀다. 시험 바이너리가 무엇을 들여오는지는
// 아래 `-deps-test` 걸음이 따로 본다.
func productionOnly(info fs.FileInfo) bool {
	return !strings.HasSuffix(info.Name(), "_test.go")
}

// 이 worker 는 아무것도 바꾸지 않는다. 그 약속은 주석이 아니라 import 목록이
// 지킨다. 스펙이 요구하는 "Worker dependency closure 에 broker mutator,
// writable journal, Guardian issuer, activation/toggle writer 가 없다"가 여기다.
//
// 금지 목록이 아니라 **허용 목록**으로 본다. 금지 목록은 빠뜨린 패키지를
// 조용히 통과시킨다 — strategyhandoff 의 첫 판본이 정확히 그렇게 internal/official
// 을 통째로 빠뜨렸다.
func TestTheWorkerImportsNothingOutsideItsAllowedClosure(t *testing.T) {
	allowed := map[string]bool{
		modulePath + "internal/breakoutlane":        true,
		modulePath + "internal/continuationlane":    true,
		modulePath + "internal/reversallane":        true,
		modulePath + "internal/strategyarbiter":     true,
		modulePath + "internal/strategycoordinator": true,
		modulePath + "internal/strategyrouter":      true,
		modulePath + "internal/weeklyvaluelane":     true,
		// 표준 라이브러리는 이름으로 하나씩 연다. "표준이면 다 허용"으로 두면
		// os/exec 와 net/http 가 함께 열리고, 그러면 이 허용 목록이 지키는 것이
		// 없어진다. 아래 셋은 능력을 만들지 않는다 — 시간을 읽는 것도 아니고
		// (레인은 now 를 인자로 받는다) 문자열과 시간 **타입**만 쓴다.
		"strconv": true,
		"strings": true,
		"time":    true,
	}
	fset := token.NewFileSet()
	packages, err := parser.ParseDir(fset, ".", productionOnly, parser.ImportsOnly)
	if err != nil {
		t.Fatalf("parse package directory: %v", err)
	}
	scanned := 0
	unexpected := make([]string, 0)
	for name, pkg := range packages {
		if name != "strategyworker" {
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
		t.Fatalf("the worker imports something outside its allowed closure: %v", unexpected)
	}
}

// 직접 import 만 보면 한 다리 건너 들어오는 능력을 놓친다. 이 검사는 전이
// 의존 전체를 훑는다.
//
// `-deps-test` 걸음이 함께 있는 이유: 시험 바이너리가 들여오는 것까지 봐야
// "어떤 자동 시험도 실계좌 주문 경로에 닿지 않는다"는 저장소 규칙이 이
// 패키지에서 확인된다.
func TestTheWorkerTransitivelyReachesNoMutationCapability(t *testing.T) {
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
	// 브로커·원장·주문 의도를 바꿀 수 있는 패키지 전부. 스펙이 이름을 부른
	// 네 능력(broker mutator, writable journal, Guardian issuer,
	// activation/toggle writer)이 이 모듈에서 실제로 어디에 사는지를 적었다.
	forbidden := []string{
		"internal/official",           // PlaceOrder/CancelOrder/ModifyOrder + 조건부 주문 3종
		"internal/hybrid",             // 공식 클라이언트가 없으면 WTS 로 주문을 흘린다
		"internal/client",             // WTS 세션 주문 변경자
		"internal/trading",            // Broker 인터페이스와 그 구현
		"internal/orderintent",        // 주문 의도 생성
		"internal/execgw",             // 실행 게이트웨이 — Guardian 발급이 여기 붙어 있다
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
				t.Errorf("the worker transitively reaches a mutation capability: %s", entry.ImportPath)
			}
		}
	}

	// 양성 대조. 걸음이 비었거나 얕으면 위의 단언이 **틀린 이유로** 통과한다.
	if len(deps) == 0 {
		t.Fatal("go list returned no dependency packages, so the walk above read nothing")
	}
	// 아래 이름들은 이 패키지의 직접 import 가 아니다. 보이면 걸음이 실제로
	// 두 단계 넘게 내려갔다는 뜻이다 — candidate 와 domain 은 strategyflow 를
	// 거쳐야 나오고, strategyflow 자체도 여기서 직접 들여오지 않는다.
	for _, required := range []string{
		modulePath + "internal/strategyworker",
		modulePath + "internal/strategyflow",
		modulePath + "internal/candidate",
		modulePath + "internal/domain",
	} {
		if !deps[required] {
			t.Fatalf("the walk did not reach %s, so it was too shallow to prove anything", required)
		}
	}
}
