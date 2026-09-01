//go:build tossos_testseams

package engine

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/risk"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyaccount"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyflow"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyproposal"
)

// 이 파일은 **오늘 실제로 주문을 막고 있는 줄**을 고정한다.
//
// 5.5 의 handoff 경계는 "dispatch 된 값이 Admit 을 거쳤다"까지만 증명한다.
// dispatchHandoff 는 패키지 사설 구조체의 메서드라서 엔진 안의 아무 함수나
// 자기가 고른 entries 로 수신자를 만들어 봉투를 주조할 수 있고, 철자를 세는
// 어떤 검사도 그것을 막지 못한다 — 4차 적대 리뷰가 그런 우회 넷을 컴파일해
// 보였다.
//
// 그 넷 중 주문을 낸 것은 하나도 없다. 이유는 경계가 아니라
// strategy_account_first_leg_authority.go 다. collectStrategyFirstLegAuthority
// 는 건너온 값을 믿지 않고 **자기 권한 쌍에서 제안을 다시 꺼내** identity 를
// 대조한다 — 개수 관문 `:217`, 재유도 `:221`–`:222`, 비교 `:223`–`:225`.
// 주조된 봉투가 나르는 결과는 그 쌍에 없으므로 여기서 거절된다.
//
// **그 줄들에는 시험이 없었다.** singleProposalAssumptionCensus 는 `entries` 의
// 모양을 세므로 비교만 지우면 5 가 유지돼 초록이다. 즉 5.5 가 "오늘의 안전은
// 여기서 온다"고 적은 줄은, 지워도 게이트가 조용한 줄이었다.
//
// L6 6.2 는 이 시험을 초록으로 유지한 채로만 그 다섯 줄을 지울 수 있다.

// firstLegIdentityFixture 는 두 시장이 모두 준비된 생산 1차 진입 권한 한 벌이다.
// 제안 쌍만 따로 들고 있는 것이 요점이다 — 아래 시험은 그 하나만 바꾼다.
type firstLegIdentityFixture struct {
	now       time.Time
	proposals strategyProposalAuthorityPair
	accounts  strategyAccountAuthorityPair
	risk      strategyRiskAuthorityPair
	fx        strategyFXAuthorityPair
	clk       *clock.Fake
	journal   *journal.Journal
	guardian  *execgw.RiskGuardian
}

func newFirstLegIdentityFixture(t *testing.T) firstLegIdentityFixture {
	t.Helper()
	riskFixture := newStrategyRiskLoaderFixture(t)
	now := riskFixture.results.observedAt
	riskPair := riskFixture.loader.collect(context.Background(), riskFixture.results, riskFixture.fx)
	proposals := strategyProposalAuthorityPair{observedAt: now}
	accounts := strategyAccountAuthorityPair{observedAt: now}
	for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {
		result := riskFixture.results.forMarket(market).result
		batch := strategyproposal.ProductionBatchAuthorityForTest("sha256:proposal-"+string(market),
			map[string]strategyflow.Result{result.Lineage.Symbol: result})
		authority, ok := batch.For(result.Lineage.Symbol)
		if !ok {
			t.Fatal("missing proposal test authority")
		}
		proposalMarket := strategyProposalMarketAuthority{market: market,
			entries:  []strategyProposalEntryAuthority{{authority: authority}},
			snapshot: StrategyProposalMarketSnapshot{Market: market, Ready: true, Reason: StrategyProposalReady}}
		quote, cash := "KRW", "5000000"
		accountMarket := strategyaccount.MarketKR
		if market == StrategyMarketUS {
			quote, cash, accountMarket = "USD", "1000", strategyaccount.MarketUS
		}
		state := risk.AccountState{Mode: risk.ModeNormal, AllowedSymbols: []string{result.Lineage.Symbol}, HeldQuantity: "0",
			CashAvailable: riskcalc.Money{Amount: cash, Currency: quote}, OpenExposure: riskcalc.Money{Amount: "0", Currency: "KRW"},
			DailyRealizedLoss: riskcalc.Money{Amount: "0", Currency: "KRW"}, AccountEquity: riskcalc.Money{Amount: "10000000", Currency: "KRW"}}
		accountAuthority := strategyaccount.AuthorityForTest(accountMarket, quote, state, now.Add(-time.Second), now.Add(time.Minute), 1,
			"sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
		accountValue := strategyAccountMarketAuthority{market: market, authority: accountAuthority,
			snapshot: StrategyAccountMarketSnapshot{Market: market, Ready: true, Reason: StrategyAccountReady}}
		if market == StrategyMarketKR {
			proposals.kr, accounts.kr = proposalMarket, accountValue
		} else {
			proposals.us, accounts.us = proposalMarket, accountValue
		}
	}
	handle := openTestJournal(t)
	fakeClock := clock.NewFake(now)
	guardian, err := execgw.NewRiskGuardian(execgw.RiskGuardianOptions{Journal: handle, Clock: fakeClock,
		AccountRef: "acct-risk-loader", Policy: risk.DefaultPolicy(), Costs: costs.DefaultModel(),
		PolicyVersion: "engine.automation_gate/risk-policy-v1"})
	if err != nil {
		t.Fatal(err)
	}
	return firstLegIdentityFixture{now: now, proposals: proposals, accounts: accounts, risk: riskPair,
		fx: riskFixture.fx, clk: fakeClock, journal: handle, guardian: guardian}
}

func (fixture firstLegIdentityFixture) loaderWith(proposals strategyProposalAuthorityPair) *productionStrategyFirstLegAuthorityLoader {
	return newProductionStrategyFirstLegAuthorityLoader(fixture.clk, fixture.journal, fixture.guardian,
		routeReadySchedulePair(fixture.now), proposals, fixture.risk, fixture.fx, fixture.accounts)
}

// TestFirstLegAuthorityRefusesAProposalItDidNotAuthorize 는 주조된 봉투가 나르는
// 결과를 1차 진입 권한이 거절하는지 **행동으로** 본다.
//
// 시장도 함께 다르다. 그래서 이 시험은 identity 대조가 위험 범위 대조보다
// **먼저** 온다는 것에 기댄다. 그 순서가 진짜라는 근거는 말이 아니라 뮤테이션이다:
// `:223`–`:225` 를 지우면 오류가 "production risk authority scope changed" 로
// 바뀌어 이 시험이 빨개진다(sha256 원복 확인).
func TestFirstLegAuthorityRefusesAProposalItDidNotAuthorize(t *testing.T) {
	fixture := newFirstLegIdentityFixture(t)
	accepted, refusal := validateStrategyFirstLegResult(fixture.proposals.kr.entries[0].authority.Proposal())
	if refusal.Code != "" {
		t.Fatalf("KR result refused before the test could start: %+v", refusal)
	}
	forged := fixture.proposals
	forged.kr = strategyProposalMarketAuthority{market: StrategyMarketKR, entries: fixture.proposals.us.entries,
		snapshot: fixture.proposals.kr.snapshot}
	_, err := fixture.loaderWith(forged).collectStrategyFirstLegAuthority(context.Background(), accepted)
	const want = "production proposal identity changed"
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Fatalf("권한 쌍에 없는 제안을 실은 결과=%v, want %q 를 담은 오류", err, want)
	}
}

// krProposal 은 fixture 의 KR 제안과 계좌·시장·종목·수량·시각이 전부 같고
// campaignID 와 가격만 인자로 받는 KR 제안 권한을 만든다.
//
// 인자를 둘로 나눈 이유가 이 파일의 요점이다. `:223` 은 **논리합 둘**이고
// 그 둘은 서로 다른 것을 지킨다 — 왼쪽은 lineage, 오른쪽은 수량과 세 가격까지.
// campaignID 를 바꾸면 lineage 가 움직이고, lineage 는 terms 의 해시에 들어가므로
// (`strategyflow/types.go:315`) **두 항이 함께 참이 된다.** 그래서 campaignID 만
// 바꾸는 시험은 두 항 중 어느 쪽이 일하는지 구별하지 못한다. 가격만 바꾸면
// lineage 는 그대로이고 오른쪽 항만 참이 된다.
func (fixture firstLegIdentityFixture) krProposal(t *testing.T, campaignID, entry, stop, target string) strategyproposal.ProductionAuthority {
	t.Helper()
	result, err := strategyflow.AcceptedResultForAuthorityTest(riskLoaderDescriptor(t, StrategyMarketKR),
		"acct-risk-loader", "005930", campaignID, 8,
		entry, stop, target, fixture.now.Add(-time.Second), fixture.now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	batch := strategyproposal.ProductionBatchAuthorityForTest("sha256:proposal-"+campaignID+"-"+stop,
		map[string]strategyflow.Result{result.Lineage.Symbol: result})
	authority, ok := batch.For(result.Lineage.Symbol)
	if !ok {
		t.Fatal("missing proposal test authority")
	}
	return authority
}

// refuseWithForgedKR 은 KR 자리에 주어진 제안을 실은 권한 쌍으로 1차 진입을
// 시도하고, 오류에 identity 거절 문구가 들어 있기를 요구한다.
func (fixture firstLegIdentityFixture) refuseWithForgedKR(t *testing.T, accepted strategyFirstLegAccepted,
	forgedKR strategyproposal.ProductionAuthority, what string,
) {
	t.Helper()
	forged := fixture.proposals
	forged.kr = strategyProposalMarketAuthority{market: StrategyMarketKR,
		entries:  []strategyProposalEntryAuthority{{authority: forgedKR}},
		snapshot: fixture.proposals.kr.snapshot}
	_, err := fixture.loaderWith(forged).collectStrategyFirstLegAuthority(context.Background(), accepted)
	const want = "production proposal identity changed"
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Fatalf("%s 를 실은 결과=%v, want %q 를 담은 오류", what, err, want)
	}
}

// TestFirstLegAuthorityRefusesASiblingCampaignOnTheSameSymbol 은 **identity 축
// 하나만** 바꾼다. 계좌·시장·종목·가격은 전부 같고 campaignID 만 다른, 똑같이
// 봉인된 제안이다.
//
// 위 시험만으로는 부족하다는 것이 5차 적대 리뷰의 P1 이었다. 위 시험은 시장까지
// 함께 바꾸므로, `:223` 의 술어가 identity 가 아니라 market 만 비교하도록
// **약해져도** 초록으로 남는다 — 리뷰어가 그 술어로 바꿔 두 스위트를 다 돌려
// 초록을 보였고, 같은 종목의 다른 캠페인으로 1차 진입 주문이 나가는 것까지
// 재현했다. 결함이 사는 축은 "그 블록이 있느냐"가 아니라 "무엇을 비교하느냐"다.
//
// a112 의 목표 상태는 한 종목에 네 가족이므로 "같은 시장·같은 종목·다른 identity"
// 가 오히려 흔한 모양이다. 이 시험이 그 축을 고정한다.
func TestFirstLegAuthorityRefusesASiblingCampaignOnTheSameSymbol(t *testing.T) {
	fixture := newFirstLegIdentityFixture(t)
	accepted, refusal := validateStrategyFirstLegResult(fixture.proposals.kr.entries[0].authority.Proposal())
	if refusal.Code != "" {
		t.Fatalf("KR result refused before the test could start: %+v", refusal)
	}
	sibling := fixture.krProposal(t, "campaign-risk-loader-kr-sibling", "100", "95", "120")
	proposal := sibling.Proposal()
	// 축이 하나뿐임을 시험이 스스로 확인한다. 이 확인이 없으면 형제 제안을 만드는
	// 쪽이 조용히 다른 축까지 바꿔도 아무도 모른다.
	if proposal.Lineage.Market != accepted.result.Lineage.Market || proposal.Lineage.Symbol != accepted.result.Lineage.Symbol ||
		proposal.Lineage.AccountRef != accepted.result.Lineage.AccountRef || proposal.Quantity != accepted.result.Quantity {
		t.Fatalf("형제 제안이 identity 말고 다른 축까지 바꿨다 — 시험이 축을 잃었다: %+v", proposal.Lineage)
	}
	if proposal.Lineage.Identity == accepted.result.Lineage.Identity {
		t.Fatal("형제 제안의 identity 가 같다 — 바꿀 축이 없다")
	}
	fixture.refuseWithForgedKR(t, accepted, sibling, "같은 종목의 다른 캠페인")
}

// TestFirstLegAuthorityRefusesRewrittenExecutionTermsUnderTheSameLineage 는
// `:223` 의 **오른쪽 논리항**이 지키는 값을 **하나씩** 고정한다.
//
// 6차 적대 리뷰가 잰 것: `:223` 은 논리합 둘이고, terms identity 해시가 lineage
// identity 를 품으므로(`strategyflow/types.go:315`) **왼쪽은 오른쪽에 포섭된다.**
// 앞의 두 시험은 campaignID 를 바꿔 lineage 를 움직이므로 두 항이 함께 참이 되고,
// 오른쪽 항만 지워도 둘 다 초록으로 남는다 — 그 상태에서 손절가를 바꾼 제안이
// 1차 진입을 낸다.
//
// **7차가 그 고침도 반증했다.** 손절과 익절을 함께 바꾸면, 둘 중 하나만 보는
// 술어에도 걸린다. 그리고 6.2 가 실제로 할 편집은 논리항 삭제가 아니라
// **비교를 필드별로 펼치면서 하나를 빠뜨리는 것**이다. 리뷰어가 `Entry`·
// `EffectiveStop`·`Target` 을 각각 하나씩 뺀 확장 넷을 넣어 보니 세 시험이 모두
// 초록이었고, 각각 그 가격만 바꾼 제안으로 주문이 나갔다.
//
// 그래서 축을 하나씩 나눈다. 각 케이스는 **자기 축 말고는 아무것도 안 움직였다**는
// 것을 스스로 확인한다 — lineage identity 가 같고, 수량이 같고, 나머지 두 가격이
// 같아야 한다. 손절·사이징은 프로젝트가 High-risk 로 못 박은 경로다.
func TestFirstLegAuthorityRefusesRewrittenExecutionTermsUnderTheSameLineage(t *testing.T) {
	// fixture 의 KR 제안은 entry 100 · stop 95 · target 120 이다.
	// 가격 순서 제약(stop < entry < target)을 지키면서 한 축씩만 옮긴다.
	for _, testCase := range []struct {
		name              string
		entry, stop, goal string
	}{
		{"entry 만", "101", "95", "120"},
		{"stop 만", "100", "80", "120"},
		{"target 만", "100", "95", "130"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			fixture := newFirstLegIdentityFixture(t)
			accepted, refusal := validateStrategyFirstLegResult(fixture.proposals.kr.entries[0].authority.Proposal())
			if refusal.Code != "" {
				t.Fatalf("KR result refused before the test could start: %+v", refusal)
			}
			twin := fixture.krProposal(t, "campaign-risk-loader-kr", testCase.entry, testCase.stop, testCase.goal)
			assertOnlyOnePriceMoved(t, accepted.result, twin.Proposal())
			fixture.refuseWithForgedKR(t, accepted, twin, "같은 lineage 아래 "+testCase.name+" 바꾼 제안")
		})
	}
}

// assertOnlyOnePriceMoved 는 쌍둥이 제안이 가격 **하나**만 옮겼는지 본다.
//
// 이 확인이 없으면 시험이 조용히 축을 잃는다. 7차 적대 리뷰가 반증한 것이
// 정확히 그것이다 — 두 가격을 함께 옮긴 시험은 둘 중 하나만 보는 술어를 못 잡는다.
func assertOnlyOnePriceMoved(t *testing.T, accepted, twin strategyflow.Result) {
	t.Helper()
	if twin.Lineage.Identity != accepted.Lineage.Identity {
		t.Fatalf("lineage 까지 움직였다 — 오른쪽 항만 움직여야 한다: %q vs %q",
			twin.Lineage.Identity, accepted.Lineage.Identity)
	}
	if twin.ExecutionTerms.Identity() == accepted.ExecutionTerms.Identity() {
		t.Fatal("terms identity 가 같다 — 바꿀 축이 없다")
	}
	if twin.Quantity != accepted.Quantity || twin.ExecutionTerms.Quantity() != accepted.ExecutionTerms.Quantity() {
		t.Fatalf("수량까지 움직였다: %d vs %d", twin.Quantity, accepted.Quantity)
	}
	moved := 0
	for _, pair := range [][2]strategyflow.PriceProvenance{
		{twin.ExecutionTerms.Entry(), accepted.ExecutionTerms.Entry()},
		{twin.ExecutionTerms.EffectiveStop(), accepted.ExecutionTerms.EffectiveStop()},
		{twin.ExecutionTerms.Target(), accepted.ExecutionTerms.Target()},
	} {
		if pair[0] != pair[1] {
			moved++
		}
	}
	if moved != 1 {
		t.Fatalf("가격이 %d 개 움직였다 — 축은 하나여야 한다", moved)
	}
}

// TestFirstLegAuthorityRefusesARestatedStopProvenanceAtTheSamePrice 는 손절
// **숫자는 그대로 두고 출처만** 바꾼 제안을 거절하는지 본다.
//
// 8차 적대 리뷰가 남긴 구멍이다. `:223` 을 세 가격의 `priceMinor` 비교로 좁혀도
// 위의 여섯 하위 시험이 전부 초록이었다 — 시험 seam 이 숫자 말고는 못 움직여서
// `PriceProvenance` 의 나머지 일곱 필드가 한 번도 안 밟혔기 때문이다.
//
// 같은 숫자에 다른 출처는 생산 경로에 실재한다:
// `continuationlane/execution_terms.go` 는 신선한 후보와 저장된 권한에서 같은
// 손절값에 서로 다른 출처를 붙인다. 그 패키지는 위조된 손절 출처를 이미 위협으로
// 다루는데(`execution_terms_test.go` 가 `Version: "forged"` 를 심는다),
// 1차 진입 backstop 에는 그 시험이 없었다.
func TestFirstLegAuthorityRefusesARestatedStopProvenanceAtTheSamePrice(t *testing.T) {
	fixture := newFirstLegIdentityFixture(t)
	accepted, refusal := validateStrategyFirstLegResult(fixture.proposals.kr.entries[0].authority.Proposal())
	if refusal.Code != "" {
		t.Fatalf("KR result refused before the test could start: %+v", refusal)
	}
	restated, err := strategyflow.ResultWithRestatedStopProvenanceForTest(accepted.result,
		"saved-effective-stop", "stop-state-v1", "sha256:"+strings.Repeat("f", 64))
	if err != nil {
		t.Fatal(err)
	}
	assertOnlyOnePriceMoved(t, accepted.result, restated)
	// 축이 **숫자가 아니라 출처**임을 시험이 스스로 확인한다. 이 확인이 없으면
	// 이 시험도 숫자 비교만으로 통과하는 시험이 된다.
	for _, pair := range [][2]strategyflow.PriceProvenance{
		{restated.ExecutionTerms.Entry(), accepted.result.ExecutionTerms.Entry()},
		{restated.ExecutionTerms.EffectiveStop(), accepted.result.ExecutionTerms.EffectiveStop()},
		{restated.ExecutionTerms.Target(), accepted.result.ExecutionTerms.Target()},
	} {
		if pair[0].PriceMinor() != pair[1].PriceMinor() {
			t.Fatalf("숫자가 움직였다 — 이 시험이 바꾸는 축은 출처뿐이다: %q vs %q",
				pair[0].PriceMinor(), pair[1].PriceMinor())
		}
	}
	batch := strategyproposal.ProductionBatchAuthorityForTest("sha256:proposal-kr-restated-stop",
		map[string]strategyflow.Result{restated.Lineage.Symbol: restated})
	authority, ok := batch.For(restated.Lineage.Symbol)
	if !ok {
		t.Fatal("missing restated proposal test authority")
	}
	fixture.refuseWithForgedKR(t, accepted, authority, "손절 숫자는 같고 출처만 바꾼 제안")
}
