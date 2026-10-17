package breakoutlane

// a112 breakout 덮개 B1(Manager 판정 2026-10-01): 골든 `allowed_transitions`(four-family-runtime-v1.json)는 **허용 가능한 변의 검증 집합**이지
// 평가기의 산출 의무가 아니다. 열넷 중 여섯 변은 v1 에 생산자가 없다 —
//   조기 INVALIDATED 다섯(DISCOVERED · RANGE_LOCKED · BREAKOUT_CLOSED · RECLAIMED · ARMED → INVALIDATED): 「v1 생산자 없음 · 예약」(review.md B1 표에 변마다 부재 근거),
//   PROPOSED → CONSUMED: v1 에서는 journal 첫-레그 결속이 곧 CONSUMED 기록(6.4 판정 (i) — design.md) — 레인 생산자 0 유지.
// 이 census 는 **지금 평가기가 낼 수 있는 변**을 고정한다. 누가 위 여섯 중 하나의 생산자를 더하면(또는 지금의 생산자를 빼면) 아래 둘 중 하나가
// 뒤집힌다 — 의도한 편집이면 census 와 review 의 B1 표를 같이 고친다(그 편집이 리뷰에 보인다). 골든 자체는 불변.
//   (1) 구조: 패키지 생산 파일 전체에서 phase 상수의 「생산 자리」 수(비교 피연산자가 아닌 모든 사용)와, 상수를 우회하는 철자
//       (phase 값과 같은 문자열 리터럴 · `phase(…)` 변환)의 수. 함수 하나가 아니라 패키지 전체를 세므로 헬퍼로 옮겨도 세어진다.
//   (2) 행동: 각 생산 자리를 지나는 픽스처 경로의 연속 쌍 = 관측 변 집합. 그 집합 ∪ 예약 여섯 = 골든 열넷, 교집합 0.

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// a112ReservedBreakoutEdges 는 골든이 허용하지만 v1 평가기에 생산자가 없는 변이다(B1 판정).
var a112ReservedBreakoutEdges = []string{
	"DISCOVERED>INVALIDATED", "RANGE_LOCKED>INVALIDATED", "BREAKOUT_CLOSED>INVALIDATED", "RECLAIMED>INVALIDATED", "ARMED>INVALIDATED", // v1 생산자 없음 · 예약
	"PROPOSED>CONSUMED", // v1: journal 첫-레그 결속이 CONSUMED 기록(6.4 판정 (i))
}

// a112ProducedBreakoutEdges 는 지금 평가기가 내는 변이다(아래 행동 census 가 실측으로 같음을 단언).
var a112ProducedBreakoutEdges = []string{
	"DISCOVERED>RANGE_LOCKED", "RANGE_LOCKED>BREAKOUT_CLOSED", "BREAKOUT_CLOSED>RETEST_WAIT", "RETEST_WAIT>RECLAIMED", "RECLAIMED>ARMED", "ARMED>PROPOSED",
	"RETEST_WAIT>INVALIDATED", "RETEST_WAIT>TIMED_OUT",
}

func TestTheBreakoutTransitionProducersAreExactlyTheCensus(t *testing.T) {
	values, uses, literals, conversions := a112PhaseCensus(t)
	// 생산 자리 수(2026-10-01 실측). INVALIDATED 2 = machine.go 범위 하단 아래 종가 · 거래량 확장 실패 재돌파, TIMED_OUT 2 = 두 timeout 판정.
	// TIMED_OUT 의 두 자리 중 `since > timeout`(machine.go:88-90)은 도달 불가다 — since 는 1 씩 늘고 `since >= timeout`(:99-101)이 같은 반복의
	// 뒤에서 먼저 돌려준다(전체 스위트 커버리지 0 실측). 지우면 동등 변이 — 자리 수에는 남아 있다(지우는 편집도 census 를 뒤집는다).
	// CONSUMED 0 · 조기 INVALIDATED 는 INVALIDATED 의 자리 수가 2 를 넘는 순간 여기서 뒤집힌다.
	want := map[string]int{
		"phaseDiscovered": 2, "phaseRangeLocked": 2, "phaseBreakoutClosed": 1, "phaseRetestWait": 2, "phaseReclaimed": 1,
		"phaseArmed": 4, "phaseProposed": 2, "phaseInvalidated": 4, "phaseTimedOut": 4, "phaseConsumed": 0,
	}
	if len(values) != len(want) {
		t.Fatalf("phase constants=%d (%v), census lists %d — a new phase needs its producer count here", len(values), values, len(want))
	}
	for name, n := range want {
		if uses[name] != n {
			t.Errorf("%s producing uses=%d, census %d — a producer was added or removed; update the census and review B1 together", name, uses[name], n)
		}
	}
	if literals != 0 || conversions != 0 {
		t.Errorf("phase spelled around its constants: %d string literals equal to a phase value, %d phase(...) conversions — want 0 (the census above cannot see them)", literals, conversions)
	}
}

func TestTheObservedBreakoutEdgesPlusTheReservedSixAreTheGoldenSet(t *testing.T) {
	var golden struct {
		States struct {
			Allowed []struct{ From, To string } `json:"allowed_transitions"`
		} `json:"states"`
	}
	path := filepath.Join(repoRoot(t), "openspec/changes/archive/2026-10-04-a112-run-four-strategy-families-independently/analysis/goldens/four-family-runtime-v1.json")
	if err := json.Unmarshal(mustRead(t, path), &golden); err != nil {
		t.Fatal(err)
	}
	allowed := map[string]bool{}
	for _, e := range golden.States.Allowed {
		allowed[e.From+">"+e.To] = true
	}
	// 행동: 도달 가능한 생산 자리마다 그 자리를 지나는 경로 하나 — 제안 · 범위 하단 아래 · 거래량 확장 실패 · 시한 · 돌파 전.
	// 경로마다 끝 단계를 단언한다 — 안 하면 픽스처가 다른 자리로 새도 관측 변 집합이 우연히 같을 수 있다.
	observed := map[string]bool{}
	for name, c := range a112EdgeCorpus(t) {
		d := Evaluate(snapshot(t, c.input), nil)
		path := d.Provenance().Transitions
		if d.Phase() != c.phase || len(path) < 2 || path[len(path)-1] != c.phase {
			t.Fatalf("%s: phase %s path %v, want it to end in %s", name, d.Phase(), path, c.phase)
		}
		for i := 1; i < len(path); i++ {
			observed[path[i-1]+">"+path[i]] = true
		}
	}
	if got, want := a112Sorted(observed), a112SortedList(a112ProducedBreakoutEdges); strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("observed edges=%v, census %v", got, want)
	}
	union := map[string]bool{}
	for _, e := range a112ProducedBreakoutEdges {
		union[e] = true
	}
	for _, e := range a112ReservedBreakoutEdges {
		if union[e] {
			t.Errorf("%s is both produced and reserved", e)
		}
		union[e] = true
	}
	if got, want := a112Sorted(union), a112Sorted(allowed); len(allowed) != 14 || strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("produced ∪ reserved=%v, golden allowed (%d)=%v", got, len(allowed), want)
	}
}

type a112EdgeCase struct {
	input EvidenceInput
	phase string
}

func a112EdgeCorpus(t *testing.T) map[string]a112EdgeCase {
	t.Helper()
	with := func(mutate func(*EvidenceInput)) EvidenceInput {
		i := fixtureInput(t)
		i.Bars = append([]ClosedBar(nil), i.Bars...)
		mutate(&i)
		return i
	}
	bar := func(i *EvidenceInput, n int, set func(*ClosedBarInput)) {
		b := i.Bars[n].value
		set(&b)
		i.Bars[n] = ClosedBar{value: b}
	}
	// 돌파 뒤 retest 없이 범위 안에 머무는 봉을 n 개 잇는다(시한 경로) — 종가 105 는 retest 허용폭(1) 밖 · 범위 하단 위.
	drift := func(i *EvidenceInput, n int) {
		i.Bars = i.Bars[:16]
		for s := uint64(17); s < uint64(17+n); s++ {
			i.Bars = append(i.Bars, fixtureBar(t, s, 106, 104, 105, 1_000_000, 100_000))
		}
	}
	return map[string]a112EdgeCase{
		"proposal": {fixtureInput(t), "PROPOSED"},
		"close below the range low": {with(func(i *EvidenceInput) {
			bar(i, 17, func(b *ClosedBarInput) { b.CloseMinor, b.HighMinor, b.LowMinor = 89, 100, 85 })
		}), "INVALIDATED"},
		"volume-expanded failed reclaim": {with(func(i *EvidenceInput) {
			bar(i, 17, func(b *ClosedBarInput) { b.CloseMinor, b.HighMinor, b.VolumeExpanded = 99, 100, true })
		}), "INVALIDATED"},
		"timeout at the KR limit":  {with(func(i *EvidenceInput) { drift(i, 8) }), "TIMED_OUT"},
		"no breakout (range only)": {with(func(i *EvidenceInput) { i.Bars = i.Bars[:15] }), "RANGE_LOCKED"},
	}
}

// a112PhaseCensus 는 패키지 생산 파일(_test 아닌 .go) 전체에서 phase 상수의 생산 사용(==/!= 피연산자가 아닌 모든 사용)을 센다.
func a112PhaseCensus(t *testing.T) (values map[string]string, uses map[string]int, literals, conversions int) {
	t.Helper()
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	var parsed []*ast.File
	values = map[string]string{}
	for _, name := range files {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, name, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		parsed = append(parsed, f)
		for _, decl := range f.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.CONST {
				continue
			}
			for _, spec := range gen.Specs {
				v := spec.(*ast.ValueSpec)
				if id, ok := v.Type.(*ast.Ident); ok && id.Name == "phase" {
					for k, n := range v.Names {
						lit, _ := strconv.Unquote(v.Values[k].(*ast.BasicLit).Value)
						values[n.Name] = lit
					}
				}
			}
		}
	}
	byValue := map[string]bool{}
	for _, v := range values {
		byValue[v] = true
	}
	uses = map[string]int{}
	for _, f := range parsed {
		var stack []ast.Node
		ast.Inspect(f, func(n ast.Node) bool {
			if n == nil {
				stack = stack[:len(stack)-1]
				return true
			}
			parent := ast.Node(nil)
			if len(stack) > 0 {
				parent = stack[len(stack)-1]
			}
			stack = append(stack, n)
			switch x := n.(type) {
			case *ast.ValueSpec:
				if id, ok := x.Type.(*ast.Ident); ok && id.Name == "phase" {
					stack = stack[:len(stack)-1] // 선언 자체는 세지 않는다
					return false
				}
			case *ast.Ident:
				if _, ok := values[x.Name]; ok {
					if b, ok := parent.(*ast.BinaryExpr); ok && (b.Op == token.EQL || b.Op == token.NEQ) {
						return true
					}
					uses[x.Name]++
				}
			case *ast.BasicLit:
				if s, err := strconv.Unquote(x.Value); err == nil && x.Kind == token.STRING && byValue[s] {
					literals++
				}
			case *ast.CallExpr:
				if id, ok := x.Fun.(*ast.Ident); ok && id.Name == "phase" {
					conversions++
				}
			}
			return true
		})
	}
	return values, uses, literals, conversions
}

func a112Sorted(set map[string]bool) []string {
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
func a112SortedList(list []string) []string {
	out := append([]string(nil), list...)
	sort.Strings(out)
	return out
}
