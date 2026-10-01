package engine

// a112 5.2.2.2 — 범위 거절 타입(strategyScopeRefusal)의 언급 census(Manager 판정 J4 ①).
//
// 주문 경로는 이 타입으로만 「그 범위만 건너뛰고 다음 범위로 간다」를 고른다. 그러므로 이 타입을 **만드는** 자리가 늘면 원장 · Gateway ·
// 중앙 무결성 오류가 범위 거절로 오분류되어, 한 범위의 고장 뒤에도 같은 주기가 주문을 계속 낸다. 만드는 모양은 합성 리터럴만이 아니다
// (`new(T)` · 변환 · 별칭 · 지역 형 선언) — 그래서 모양을 인식하지 않고 **식별자 언급 전부**를 파일 · 함수 단위로 센다. 새 언급은
// 무엇이든 이 표를 바꿔야 하고, 표를 바꾸는 편집은 리뷰에 보인다.

import (
	"errors"
	"fmt"
	"go/ast"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyflow"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyhandoff"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

// strategyScopeRefusalMentions 는 생산 소스에서 그 타입 이름이 나오는 자리(파일:함수 — 함수 밖은 "(decl)")와 횟수.
var strategyScopeRefusalMentions = map[string]int{
	// 선언과 Error · Unwrap 메서드의 수신자.
	"strategy_owner_scope_authority.go:(decl)": 1,
	"strategy_owner_scope_authority.go:Error":  1,
	"strategy_owner_scope_authority.go:Unwrap": 1,
	// **만드는 자리는 여기 하나뿐**: 그 범위의 위험 권한이 범위 국소 원인(riskbucket.ErrProductionRiskScopeRefused)으로 준비되지 않음.
	// 2 → 1(2026-10-01, codex 재확인 P1 → Manager 판정 (A)): 계좌 쪽은 범위 국소 원인이 없어(시장 단위 매니페스트) 언제나 결함.
	"strategy_account_first_leg_authority.go:collectStrategyFirstLegAuthority": 1,
	// 둘째 만드는 자리(a112 6.2, Manager 판정 (A)): admit 이 Guardian precheck 의 버킷 고갈(QFinalRefusal 코드 BUCKET_CAP_EXHAUSTED — 타입 · 코드)을
	// 그 범위의 거절로 싣는다. 다른 precheck 거절 · 발급 단계 CAS(STALE)는 결함 그대로.
	"strategy_first_leg_admission.go:admit": 1,
	// 운반: 전달 몸통의 errors.As 대상(dispatch 는 수집 오류를 사슬째 나르므로 이 타입을 언급하지 않음 — 리뷰 수리).
	"strategy_market_handoff_delivery.go:deliverEachStrategyHandoff": 1,
}

func TestTheScopeRefusalTypeIsMadeOnlyWhereTheCensusSaysItIs(t *testing.T) {
	found := map[string]int{}
	for _, path := range engineProductionFiles(t) {
		file := parseEngineFile(t, path)
		base := filepath.Base(path)
		count := func(node ast.Node, label string) {
			ast.Inspect(node, func(inner ast.Node) bool {
				if ident, ok := inner.(*ast.Ident); ok && ident.Name == "strategyScopeRefusal" {
					found[base+":"+label]++
				}
				return true
			})
		}
		for _, decl := range file.Decls {
			if function, ok := decl.(*ast.FuncDecl); ok {
				count(function, function.Name.Name)
				continue
			}
			count(decl, "(decl)")
		}
	}
	var drift []string
	for site, got := range found {
		if strategyScopeRefusalMentions[site] != got {
			drift = append(drift, site)
		}
	}
	for site := range strategyScopeRefusalMentions {
		if _, ok := found[site]; !ok {
			drift = append(drift, site)
		}
	}
	if len(drift) != 0 {
		sort.Strings(drift)
		t.Fatalf("scope-refusal type mentions moved at %s: found %v, census %v — a new maker widens what the order path skips",
			strings.Join(drift, ", "), found, strategyScopeRefusalMentions)
	}
}

// 이름을 언급하지 않는 생산자(5.2.2.2 리뷰 A #2 — 사본 실측): `As(any) bool` 메서드를 가진 오류 감싸개는 errors.As 가 그 메서드로 대상 포인터를
// 채우게 해서 타입 이름 없이 범위 거절을 만들어 낸다. 엔진 생산 코드에는 `As` 메서드 선언이 하나도 없어야 한다(필요해지면 이 시험이 그 자리를
// 이름으로 받아들이는 편집이 리뷰에 보인다).
func TestNoEngineErrorTypeImplementsAs(t *testing.T) {
	var found []string
	for _, path := range engineProductionFiles(t) {
		for _, decl := range parseEngineFile(t, path).Decls {
			if function, ok := decl.(*ast.FuncDecl); ok && function.Recv != nil && function.Name.Name == "As" {
				found = append(found, filepath.Base(path))
			}
		}
	}
	if len(found) != 0 {
		t.Fatalf("engine production code declares As methods in %v — an As method can mint a scope refusal without naming its type", found)
	}
}

// 분류는 타입으로(J4 ① · ②): 전달 몸통은 `*strategyScopeRefusal` 만 건너뛰고(감싸도 — `%w`), 원장 · Gateway · 중앙 오류와
// **같은 문구를 흉내 낸 평문 오류**에서는 멈춘다. 멈추기 전에 건너뛴 범위 거절은 반환 오류에 남는다(J4 ③ — 각각 기록).
func TestOnlyATypedScopeRefusalIsSkippedEveryOtherFaultStops(t *testing.T) {
	selected := []strategyflow.Result{
		{Lineage: strategyflow.Lineage{Identity: "a", AccountRef: "acct", Market: strategyrouter.MarketKR, Symbol: "005930", PositionGeneration: 1}},
		{Lineage: strategyflow.Lineage{Identity: "b", AccountRef: "acct", Market: strategyrouter.MarketKR, Symbol: "000660", PositionGeneration: 1}},
		{Lineage: strategyflow.Lineage{Identity: "c", AccountRef: "acct", Market: strategyrouter.MarketKR, Symbol: "035420", PositionGeneration: 1}},
	}
	scope := &strategyScopeRefusal{scope: strategyrouter.OwnerKey{AccountRef: "acct", Market: strategyrouter.MarketKR, Symbol: "005930"},
		detail: "no ready risk authority for this owner scope"}
	for _, fault := range []struct {
		name string
		err  error
	}{
		{"journal", journal.ErrStrategyDispatchLeaseConsumed},
		{"gateway", execgw.ErrMalformedOrder},
		{"central", errors.New("production proposal identity changed")},
		{"mimic of the scope sentence", errors.New(scope.Error())},
	} {
		t.Run(fault.name, func(t *testing.T) {
			var seen []string
			err := deliverEachStrategyHandoff(strategyhandoff.AdmitEachOwnerScope(true, selected), func(delivered strategyhandoff.Delivered) error {
				identity := delivered.Result().Lineage.Identity
				seen = append(seen, identity)
				switch identity {
				case "a":
					return fmt.Errorf("engine: first-leg admission X: %w", scope) // 범위 거절 — 건너뜀
				case "b":
					return fault.err // 범위 거절 아님 — 멈춤
				}
				return nil
			})
			if strings.Join(seen, ",") != "a,b" {
				t.Fatalf("delivered=%v — want a skipped, b stopping the cycle before c", seen)
			}
			if !errors.Is(err, fault.err) {
				t.Fatalf("err=%v — want the stopping fault returned", err)
			}
			var skipped *strategyScopeRefusal
			if !errors.As(err, &skipped) || skipped != scope {
				t.Fatalf("err=%v — want the skipped scope refusal kept beside the stopping fault (recorded, not dropped)", err)
			}
		})
	}
	// 범위 거절만 있으면 끝까지 가고 거절을 돌려준다.
	var seen []string
	err := deliverEachStrategyHandoff(strategyhandoff.AdmitEachOwnerScope(true, selected), func(delivered strategyhandoff.Delivered) error {
		seen = append(seen, delivered.Result().Lineage.Identity)
		if delivered.Result().Lineage.Identity == "a" {
			return scope
		}
		return nil
	})
	var skipped *strategyScopeRefusal
	if strings.Join(seen, ",") != "a,b,c" || !errors.As(err, &skipped) {
		t.Fatalf("delivered=%v err=%v — want every scope reached and the refusal returned", seen, err)
	}
}
