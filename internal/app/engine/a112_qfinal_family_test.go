//go:build tossos_testseams

package engine

// a112 태스크 6.2 본문 — q_final 은 q_candidate 와 모든 Guardian · horizon · market · family(strategy) · sector · symbol cap 의 최솟값이고, 한 family 버킷의
// 고갈은 그 family 만 막는다(스펙 「한 family risk bucket 고갈」). 계산 자체는 a066 `riskbucket.CalculateAdmission` 이 하고 그 층은 a066 시험이 잰다 —
// 여기서는 a112 의 **생산 경로**(활성 두 범위 · family 결속 6.1) 위에서 다시 잰다.

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyflow"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyhandoff"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyproposal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

func a112KRDescriptorOfFamily(t *testing.T, family strategyrouter.Family) *strategyflow.Descriptor {
	t.Helper()
	for _, descriptor := range strategyflow.Descriptors() {
		if descriptor.Market != strategyrouter.MarketKR {
			continue
		}
		if got, ok := strategyrouter.ProductionLaneFamily(strategyrouter.MarketKR, descriptor.LaneID); ok && got == family {
			d := descriptor
			return &d
		}
	}
	t.Fatalf("no KR descriptor of family %s", family)
	return nil
}

func a112DecisionQuantities(t *testing.T, fixture a112TradingFixture, symbol string) (candidate, final int64) {
	t.Helper()
	db, err := sql.Open("sqlite", "file:"+fixture.journal.Path()+"?mode=ro")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.QueryRow(`SELECT q_candidate,q_final FROM risk_bucket_final_decisions WHERE symbol=?`, symbol).Scan(&candidate, &final); err != nil {
		t.Fatalf("decision for %s: %v", symbol, err)
	}
	return candidate, final
}

// min 결속: family(strategy) 버킷 한도를 줄여 그것이 가장 작은 cap 이 되게 하면, 발급 수량 = 그 cap(독립 계산 — MaximumQuantity) < q_candidate.
func TestQFinalIsBoundByTheFamilyBucketWhenItIsTheSmallestCap(t *testing.T) {
	probe := newA112TradingFixture(t, a112TradingOptions{})
	key, _ := strategyOwnerKeyOf(probe.winner.Proposal().Lineage)
	bundle, ok := probe.risk.kr.forScope(key)
	if !ok {
		t.Fatal("arrangement: no risk bundle for the first scope")
	}
	policy := bundle.Policy()
	limit, err := riskbucket.ReservationMinor(3, policy)
	if err != nil {
		t.Fatal(err)
	}
	fixture := newA112TradingFixture(t, a112TradingOptions{strategyLimit: limit})
	candidateQuantity := fixture.winner.Proposal().Quantity
	want, err := riskbucket.MaximumQuantity(limit, candidateQuantity, policy)
	if err != nil || want == 0 || want >= candidateQuantity {
		t.Fatalf("arrangement: family cap=%d (err=%v) must bind below q_candidate=%d", want, err, candidateQuantity)
	}
	_ = fixture.deliverKR(t)
	candidate, final := a112DecisionQuantities(t, fixture, "005930")
	if uint64(candidate) != candidateQuantity || uint64(final) != want {
		t.Fatalf("decision q_candidate=%d q_final=%d, want %d and the family cap %d", candidate, final, candidateQuantity, want)
	}
	fixture.spy.mu.Lock()
	defer fixture.spy.mu.Unlock()
	if len(fixture.spy.calls) == 0 || fixture.spy.calls[0].Intent.Quantity != float64(want) {
		t.Fatalf("placed=%+v, want the first leg sized to the family cap %d", fixture.spy.calls, want)
	}
}

// 한 family 버킷 고갈(스펙 시나리오): reversal family 의 risk_id 한도 0 → reversal 범위만 q_final 0 거절, continuation 은 발급, 계좌 cap 은 복제되지 않음.
func TestAnExhaustedFamilyBucketRefusesOnlyThatFamily(t *testing.T) {
	for _, order := range []struct {
		name  string
		first bool
	}{{"fixture order (continuation first)", false}, {"coordinator order (reversal 000660 first)", true}} {
		t.Run(order.name, func(t *testing.T) {
			reversal := a112KRDescriptorOfFamily(t, strategyrouter.FamilyReversal)
			fixture := newA112TradingFixture(t, a112TradingOptions{secondLane: reversal, coordinatorOrder: order.first,
				extraStrategies: []riskLoaderStrategy{{LaneID: reversal.LaneID, LaneVersion: reversal.LaneVersion, Horizon: riskbucket.HorizonShort,
					RiskID: "reversal", RiskVersion: "reversal-risk-v1", LimitMinor: "0"}}})
			err := fixture.deliverKR(t)
			if got := strings.Join(fixture.placedSymbols(), ","); got != "005930" {
				t.Fatalf("placed=%s err=%v, want the continuation scope issued and the exhausted reversal family refused", got, err)
			}
			refusal := a112ScopeRefusalOf(err)
			if refusal == nil || refusal.scope.Symbol != "000660" || !strings.Contains(err.Error(), string(riskbucket.RefusalBucketCapExhausted)) {
				t.Fatalf("err=%v, want the reversal scope's bucket-cap exhaustion recorded as that scope's typed refusal", err)
			}
		})
	}
}

// 공유 차원 고갈(조건 ①): horizon 버킷이 0 이면 두 범위 모두 같은 고갈로 **각자** 범위 거절되고(타입 · 기록), 발급 0, 주기는 결함으로 멈추지
// 않는다 — 범위 거절로 건너뛰어도 공유 차원 고갈에서는 새로 발급되는 것이 없다(「공허-안전」의 시험).
func TestASharedDimensionExhaustionRefusesEveryScopeAndIssuesNothing(t *testing.T) {
	fixture := newA112TradingFixture(t, a112TradingOptions{horizonShort: "0"})
	err := fixture.deliverKR(t)
	if placed := fixture.placedSymbols(); len(placed) != 0 {
		t.Fatalf("placed=%v err=%v, want no first leg when the shared horizon bucket is exhausted", placed, err)
	}
	refused := map[string]bool{}
	for _, wrapped := range a112JoinedErrors(err) {
		if refusal := a112ScopeRefusalOf(wrapped); refusal != nil && strings.Contains(wrapped.Error(), string(riskbucket.RefusalBucketCapExhausted)) {
			refused[refusal.scope.Symbol] = true
		}
	}
	if !refused["005930"] || !refused["000660"] {
		t.Fatalf("err=%v — want both scopes refused by the shared horizon exhaustion, each recorded as its own typed refusal", err)
	}
}

// a112JoinedErrors 는 errors.Join 묶음을 펼친다(하나면 그 자체).
func a112JoinedErrors(err error) []error {
	if joined, ok := err.(interface{ Unwrap() []error }); ok {
		return joined.Unwrap()
	}
	if err == nil {
		return nil
	}
	return []error{err}
}

// 같은 소유자 범위 · 다른 family 의 동시 admission(스펙 「두 family 의 동시 owner 획득」): 두 조립(continuation 쌍 · reversal 쌍)이 **한 원장 ·
// 한 Guardian · 한 dispatch owner** 위에서 같은 범위(005930)를 동시에 dispatch 하면 원장의 원자 owner/q_final 트랜잭션 하나만 성공하고 다른 쪽은
// 브로커 요청 전에 거절된다(스파이 주문 1 · 결정 1 · 예약 한 세트).
//
// 조정자를 우회하는 이유: 조정자는 한 조립 안에서 같은 범위 둘을 OverCapacity 로 막는다 — 그 경로의 거절은
// `TestTheSameOwnerScopeSealedTwiceRefusesTheActivatedMarket`(5.2.2.1 C#2 핀)이 따로 잰다. 여기서 재는 것은 그 그늘 밖, 서로 다른 조립(예: 새로
// 고침 경계에 걸친 두 파도)이 원장에서 만났을 때의 단일성이다 — 조정자 그늘에 가리면 두 판정이 서로를 가린다.
func TestTwoFamiliesRacingForOneOwnerScopeLeaveOneTransaction(t *testing.T) {
	reversal := a112KRDescriptorOfFamily(t, strategyrouter.FamilyReversal)
	fixture := newA112TradingFixture(t, a112TradingOptions{extraStrategies: []riskLoaderStrategy{{LaneID: reversal.LaneID,
		LaneVersion: reversal.LaneVersion, Horizon: riskbucket.HorizonShort, RiskID: "reversal", RiskVersion: "reversal-risk-v1", LimitMinor: "5000000"}}})
	winner := fixture.winner.Proposal()
	rival, err := strategyflow.AcceptedResultForAuthorityTest(*reversal, winner.Lineage.AccountRef, winner.Lineage.Symbol,
		"campaign-reversal-race", winner.Quantity, "100", "95", "120", fixture.now.Add(-time.Second), fixture.now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	batch := strategyproposal.ProductionBatchAuthorityForTest("sha256:reversal-race", map[string]strategyflow.Result{rival.Lineage.Symbol: rival})
	sealed, ok := batch.For(rival.Lineage.Symbol)
	if !ok {
		t.Fatal("arrangement: reversal proposal not sealed")
	}
	pairB := fixture.proposals
	pairB.kr = strategyProposalMarketAuthority{market: StrategyMarketKR, activation: fixture.proposals.kr.activation,
		entries:  []strategyProposalEntryAuthority{{authority: sealed}},
		snapshot: StrategyProposalMarketSnapshot{Market: StrategyMarketKR, Ready: true, Reason: StrategyProposalReady}}
	pairB.kr.snapshot.ProposalSetDigest = strategyProposalSetDigest(pairB.kr.entries)
	riskB := fixture.riskLoader.collect(context.Background(), pairB.ResultAuthority(), fixture.fx)
	if _, scoped := riskB.kr.forScope(a112KeyOf(t, sealed.Proposal())); !scoped {
		t.Fatalf("arrangement: reversal scope has no risk bundle: %+v", riskB.kr.snapshot)
	}
	schedule := fixture.loader.schedule
	loaderB := newProductionStrategyFirstLegAuthorityLoader(fixture.clk, fixture.journal, fixture.guardian, schedule, pairB, riskB, fixture.fx, fixture.accounts)
	cycleB := newStrategyDispatchCycle(fixture.journal, fixture.spy, newStrategyFirstLegAdmissionBridge(fixture.guardian, loaderB), schedule,
		fixture.fx, riskB, pairB, fixture.cycle.owner)
	cycleB.revalidateSchedule = fixture.cycle.revalidateSchedule
	cycleB.now = fixture.cycle.now

	start := make(chan struct{})
	results := make(chan error, 2)
	for _, run := range []struct {
		cycle    *strategyDispatchCycle
		proposal strategyflow.Result
	}{{fixture.cycle, winner}, {cycleB, sealed.Proposal()}} {
		run := run
		go func() {
			<-start
			results <- dispatchStrategyMarketHandoffs(context.Background(), fixture.journal, run.cycle,
				strategyhandoff.AdmitEachOwnerScope(true, []strategyflow.Result{run.proposal}))
		}()
	}
	close(start)
	errs := []error{<-results, <-results}
	if placed := fixture.placedSymbols(); len(placed) != 1 || placed[0] != "005930" {
		t.Fatalf("placed=%v errs=%v, want exactly one first leg for the contested owner scope", placed, errs)
	}
	if errs[0] == nil && errs[1] == nil {
		t.Fatalf("both families reported success for one owner scope: %v", errs)
	}
	loser := errs[0]
	if loser == nil {
		loser = errs[1]
	}
	if !strings.Contains(loser.Error(), "owner conflict") {
		t.Fatalf("loser err=%v — want the journal's atomic owner conflict (both families must have reached admission)", loser)
	}
	db, err := sql.Open("sqlite", "file:"+fixture.journal.Path()+"?mode=ro")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var decisions, reservationSets int
	if err := db.QueryRow(`SELECT count(*) FROM risk_bucket_final_decisions WHERE symbol='005930'`).Scan(&decisions); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT count(DISTINCT decision_id) FROM risk_bucket_reservations WHERE symbol='005930'`).Scan(&reservationSets); err != nil {
		t.Fatal(err)
	}
	if decisions != 1 || reservationSets != 1 {
		t.Fatalf("decisions=%d reservation sets=%d for the contested scope, want one transaction only (loser rolled back)", decisions, reservationSets)
	}
}

func a112KeyOf(t *testing.T, result strategyflow.Result) strategyrouter.OwnerKey {
	t.Helper()
	key, ok := strategyOwnerKeyOf(result.Lineage)
	if !ok {
		t.Fatal("arrangement: owner key")
	}
	return key
}

// 경계(조건 ②의 양성 쪽): 버킷 고갈이 **아닌** precheck 거절은 범위 거절로 넓히지 않는다 — 계좌가 그 종목을 허용하지 않는 범위가 앞이면 주기가
// 멈춰 뒤 범위도 그 주기에 나가지 않는다(결함 · 타입 없음).
func TestAPrecheckRefusalOtherThanBucketExhaustionStillStopsTheCycle(t *testing.T) {
	fixture := newA112TradingFixture(t, a112TradingOptions{disallowFor: "005930"})
	err := fixture.deliverKR(t)
	if placed := fixture.placedSymbols(); len(placed) != 0 {
		t.Fatalf("placed=%v err=%v, want the cycle stopped at the first scope's non-exhaustion precheck refusal", placed, err)
	}
	if err == nil || a112ScopeRefusalOf(err) != nil || strings.Contains(err.Error(), string(riskbucket.RefusalBucketCapExhausted)) ||
		!strings.Contains(err.Error(), string(StrategyFirstLegAuthorityMismatch)) {
		t.Fatalf("err=%v — want an untyped precheck (AUTHORITY_MISMATCH) refusal that is not bucket exhaustion", err)
	}
	t.Logf("non-exhaustion precheck refusal: %v", err)
}

// 경계(조건 ② — 같은 q_final 타입의 다른 코드): 버킷 고갈이 아닌 q_final 거절(거래당 위험 예산 1 KRW → q_existing_guardian 0)은 계좌 단위 Guardian 판정이다 —
// 범위 거절로 넓히지 않는다(주기 멈춤 · 타입 없음). 코드 분류를 「고갈」 밖으로 넓히는 변이가 여기서 잡힌다.
func TestAnExistingGuardianCapRefusalIsNotAScopeRefusal(t *testing.T) {
	fixture := newA112TradingFixture(t, a112TradingOptions{riskBudget: "1"})
	err := fixture.deliverKR(t)
	if placed := fixture.placedSymbols(); len(placed) != 0 {
		t.Fatalf("placed=%v err=%v, want the cycle stopped at the first scope's guardian-cap refusal", placed, err)
	}
	var qFinal *execgw.QFinalRefusal
	if err == nil || a112ScopeRefusalOf(err) != nil || !strings.Contains(err.Error(), string(riskbucket.RefusalExistingGuardianCap)) ||
		!errors.As(err, &qFinal) && !strings.Contains(err.Error(), "q_final refused") {
		t.Fatalf("err=%v — want the untyped q_existing_guardian refusal (a q_final code other than bucket exhaustion)", err)
	}
	_ = qFinal
	t.Logf("guardian-cap refusal: %v", err)
}
