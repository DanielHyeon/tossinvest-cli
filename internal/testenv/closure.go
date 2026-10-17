package testenv

// closure.go — 의존 폐포 가드의 공용 부품(a112 8.2, Manager 판정 2026-10-04).
//
// 레인 · 증거 · worker 패키지가 「주문 변경자 · 쓰기 원장 · Guardian 발급 · 활성화/토글 쓰기」에 닿지 않는다는 약속은 패키지마다 시험이
// 지킨다. 그 시험들이 각자 금지 목록을 베끼면 목록이 갈린다(8.2 census: strategyflow 의 목록은 client · hybrid · ops 를 빠뜨렸다). 그래서
// **금지 목록은 여기 한 곳**이고, 걸음(`go list`)도 여기 한 곳이다 — 태그 뒤 시험까지 걷는 모드를 빠뜨리지 않게.

import (
	"encoding/json"
	"errors"
	"io"
	"os/exec"
	"sort"
	"strings"
	"testing"
)

// ModulePath 는 이 저장소의 Go 모듈 경로다.
const ModulePath = "github.com/JungHoonGhae/tossinvest-cli/"

// MutationCapability 는 금지 대상 패키지 하나와 그것이 품은 능력이다.
type MutationCapability struct {
	Path string // 모듈 안 경로(접두어 일치 — 하위 패키지 포함)
	Why  string
}

// MutationCapabilities 는 이 모듈에서 스펙이 부른 네 능력(broker mutator · writable journal · Guardian issuer · activation/toggle writer)이
// 실제로 사는 패키지다. 새 능력 패키지가 생기면 여기 한 줄을 더한다 — 모든 가드가 함께 본다.
func MutationCapabilities() []MutationCapability {
	return []MutationCapability{
		{"internal/official", "공식 API 주문 변경자 — PlaceOrder · CancelOrder · ModifyOrder · 조건부 주문 3종"},
		{"internal/hybrid", "공식 클라이언트가 없으면 WTS 로 주문을 흘린다"},
		{"internal/client", "WTS 세션 주문 변경자"},
		{"internal/trading", "Broker 인터페이스와 구현"},
		{"internal/orderintent", "주문 의도 생성"},
		{"internal/execgw", "실행 게이트웨이 — Guardian 발급"},
		{"internal/strategydispatch", "전략 dispatch — Guardian 발급 경로"},
		{"internal/journal", "원장 쓰기 · lease"},
		{"internal/ops", "운영 토글 flip"},
		{"internal/protectionofficial", "보호 주문 배치"},
		{"internal/verifylive", "실계좌 검증 경로"},
		{"internal/config", "운영 설정 쓰기"},
		{"internal/app", "CLI · 엔진 배선 — 위를 전부 끌고 온다"},
	}
}

// Capability 는 import 경로가 금지 능력 패키지(또는 그 하위)면 그 항목을 돌려준다.
func Capability(importPath string) (MutationCapability, bool) {
	for _, capability := range MutationCapabilities() {
		full := ModulePath + capability.Path
		if importPath == full || strings.HasPrefix(importPath, full+"/") {
			return capability, true
		}
	}
	return MutationCapability{}, false
}

// WalkMode 는 한 번의 `go list -deps` 걸음이다.
type WalkMode struct {
	Name  string
	Tests bool   // -test: 그 패키지의 시험 이진 폐포까지
	Tags  string // 빌드 태그(빈 값이면 무태그)
}

// WalkModes 는 가드가 돌아야 하는 네 걸음이다: 생산 · 시험 이진 각각을 무태그와 `tossos_testseams` 로. 태그 뒤 시험 파일은 무태그 걸음에
// 보이지 않는다(8.2 census — 코디네이터 · 라우터의 태그 시험 폐포에 journal 이 있었는데 아무 걸음도 못 봤다).
func WalkModes() []WalkMode {
	return []WalkMode{
		{Name: "deps", Tests: false},
		{Name: "deps-test", Tests: true},
		{Name: "deps-tagged", Tests: false, Tags: "tossos_testseams"},
		{Name: "deps-test-tagged", Tests: true, Tags: "tossos_testseams"},
	}
}

// ImportGraph 는 `go list -json -deps` 의 결과다: 걸음이 닿은 패키지 → 그 패키지의 import.
type ImportGraph map[string][]string

// ListDeps 는 dir(시험의 작업 디렉터리 기준 상대 경로, 보통 ".")의 패키지에서 mode 걸음을 돈다.
func ListDeps(t *testing.T, dir string, mode WalkMode) ImportGraph {
	t.Helper()
	args := []string{"list", "-json", "-deps"}
	if mode.Tests {
		args = append(args, "-test")
	}
	if mode.Tags != "" {
		args = append(args, "-tags", mode.Tags)
	}
	args = append(args, dir)
	out, err := exec.Command("go", args...).Output()
	if err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			t.Fatalf("go %s: %v\n%s", strings.Join(args, " "), err, exit.Stderr)
		}
		t.Fatalf("go %s: %v", strings.Join(args, " "), err)
	}
	graph := ImportGraph{}
	decoder := json.NewDecoder(strings.NewReader(string(out)))
	for {
		var entry struct {
			ImportPath string
			Imports    []string
		}
		if err := decoder.Decode(&entry); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			t.Fatalf("decode go list output: %v", err)
		}
		graph[entry.ImportPath] = append(graph[entry.ImportPath], entry.Imports...)
	}
	if len(graph) == 0 {
		t.Fatalf("go %s returned no packages — the walk read nothing", strings.Join(args, " "))
	}
	return graph
}

// Reachable 는 root 에서 출발해 닿는 패키지 집합이다. cut 에 든 패키지로는 들어가지 않는다(그 패키지 자신도 집합에 넣지 않음).
// root 는 걸음에 나온 이름 그대로(시험 변형 `p [p.test]` 포함) 준다.
func (graph ImportGraph) Reachable(roots []string, cut map[string]bool) map[string]bool {
	seen := map[string]bool{}
	stack := append([]string(nil), roots...)
	for len(stack) > 0 {
		next := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		// 잘라 낼 패키지는 시험 변형 이름(`p [q.test]`)으로 나와도 자른다.
		base := next
		if index := strings.Index(base, " ["); index > 0 {
			base = base[:index]
		}
		if seen[next] || cut[next] || cut[base] {
			continue
		}
		seen[next] = true
		stack = append(stack, graph[next]...)
	}
	return seen
}

// Roots 는 걸음에서 대상 패키지(와 그 시험 변형)의 이름이다.
func (graph ImportGraph) Roots(importPath string) []string {
	var roots []string
	for name := range graph {
		if name == importPath || strings.HasPrefix(name, importPath+" [") || name == importPath+".test" ||
			strings.HasPrefix(name, importPath+"_test") {
			roots = append(roots, name)
		}
	}
	sort.Strings(roots)
	return roots
}

// ForbiddenIn 는 집합 안의 금지 능력 패키지를 「경로 — 이유」 로 정렬해 돌려준다. allowed 에 든 능력 경로(모듈 상대)는 뺀다 — 쓰는 쪽이
// 왜 허용하는지 자기 시험에 적어야 한다.
func ForbiddenIn(packages map[string]bool, allowed ...string) []string {
	skip := map[string]bool{}
	for _, path := range allowed {
		skip[path] = true
	}
	var found []string
	for name := range packages {
		// 시험 변형 이름(`p [q.test]`)은 실제 import 경로로 줄인다.
		path := name
		if index := strings.Index(path, " ["); index > 0 {
			path = path[:index]
		}
		if capability, ok := Capability(path); ok && !skip[capability.Path] {
			found = append(found, path+" — "+capability.Why)
		}
	}
	sort.Strings(found)
	return dedupe(found)
}

func dedupe(sorted []string) []string {
	out := sorted[:0]
	for index, value := range sorted {
		if index == 0 || value != sorted[index-1] {
			out = append(out, value)
		}
	}
	return out
}
