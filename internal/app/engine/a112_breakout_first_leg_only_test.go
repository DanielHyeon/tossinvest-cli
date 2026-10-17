//go:build tossos_testseams

package engine

// a112 태스크 6.4 (Manager 판정 2026-10-01: A + CONSUMED (i)) — breakout 첫 레그 전용 생산 권한을 브로커 스파이로 잰다.
//
// **seam 주입 사유.** 생산 breakout 입력은 전부 `strategyproposal.ErrBreakoutEvidenceUnavailable` 로 거절된다(벽 — 핀은
// `strategyproposal` `TestEveryProductionBreakoutLaneInputIsRefusedAtTheWall`). 그래서 breakout 레인의 봉인된 제안은 생산 경로로 만들 수 없고,
// 여기서는 `strategyflow.AcceptedResultForAuthorityTest`(등록된 breakout descriptor 만 받음)로 같은 모양의 결과를 주입한다. 주입 뒤의 경로
// — 제안 권한 쌍 → 위험 · 계좌 적재기 → 1차 레그 권한 loader → dispatch 주기 → 생산 전달 몸통 → 실제 journal(원장) → Gateway 스파이 — 는 생산과 같다
// (`newA112TradingFixture`). 레인 쪽 멱등(같은 스냅숏 · prior → 같은 결정)은 `breakoutlane` `TestFinalRedTeamDuplicateSnapshotIsIdempotent` ·
// `TestAProposedSetupNeverReSizesIntoALowerLegOrARetreatedStop` 이 잰다 — 여기서는 그 결과가 몇 번 건너오든 브로커 요청이 하나임을 잰다.
//
// 층(각각 따로 잰다 — 앞 층이 뒤 층을 가리지 않게):
//   ① 전달 몸통 `dispatchStrategyMarketHandoffs` — 캠페인이 claim 됐거나 FLAT/CLOSED 가 아니면 dispatch 를 부르지 않는다.
//   ② 1차 레그 권한 loader — 같은 CAS 를 다시 읽어 `production position campaign CAS changed` 로 거절.
//   ③ journal admission — 활성 캠페인 유일 색인 · claim · 위치 CAS(`insertFirstLegCampaignTx`). journal 시험
//      `TestFirstLegAtomicAdmissionCompetingSameScopeHasOneWinner` · `TestFirstLegAtomicAdmissionExactReplayUsesOriginalJournalToken` 이 잰다.
//
// 잔여(면제 불가 선행 — tasks 6.4 · ROADMAP): 손절-종결(CLOSED · claim 해제) → 재시작(prior 상실) → 새 매니페스트 CampaignID → 같은 setup 둘째 첫 레그.
// 이 시험들이 다루는 것은 캠페인이 열려 있는 동안뿐이다.

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyflow"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyhandoff"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyproposal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

const a112BreakoutSymbol = "000660"

// newA112BreakoutFixture 는 둘째 범위(000660)를 breakout 레인으로 세우고 조정자 순서에서 앞에 둔다 — 첫 파도의 첫 레그가 breakout 이다.
func newA112BreakoutFixture(t *testing.T) (a112TradingFixture, strategyflow.Descriptor) {
	t.Helper()
	breakout := a112KRDescriptorOfFamily(t, strategyrouter.FamilyBreakoutRetest)
	fixture := newA112TradingFixture(t, a112TradingOptions{secondLane: breakout, coordinatorOrder: true,
		extraStrategies: []riskLoaderStrategy{{LaneID: breakout.LaneID, LaneVersion: breakout.LaneVersion, Horizon: riskbucket.HorizonShort,
			RiskID: "breakout", RiskVersion: "breakout-risk-v1", LimitMinor: "5000000"}}})
	if lane := fixture.second.Proposal().Lineage.LaneID; lane != breakout.LaneID || fixture.second.Proposal().Lineage.Symbol != a112BreakoutSymbol {
		t.Fatalf("arrangement: second scope lane=%s symbol=%s, want the breakout lane on %s", lane, fixture.second.Proposal().Lineage.Symbol, a112BreakoutSymbol)
	}
	return fixture, *breakout
}

func (fixture a112TradingFixture) placedFor(symbol string) int {
	fixture.spy.mu.Lock()
	defer fixture.spy.mu.Unlock()
	n := 0
	for _, call := range fixture.spy.calls {
		if call.Intent.Symbol == symbol {
			n++
		}
	}
	return n
}

// a112LedgerRows 는 그 종목의 첫 레그 흔적을 원장에서 센다: 캠페인 · 첫 레그 결속 · 캠페인 레그.
func a112LedgerRows(t *testing.T, fixture a112TradingFixture, symbol string) (campaigns, bindings, legs int) {
	t.Helper()
	db, err := sql.Open("sqlite", "file:"+fixture.journal.Path()+"?mode=ro")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for query, out := range map[string]*int{
		`SELECT count(*) FROM position_campaigns WHERE symbol=?`:                                                              &campaigns,
		`SELECT count(*) FROM strategy_first_leg_bindings b JOIN position_campaigns c ON c.id=b.campaign_id WHERE c.symbol=?`: &bindings,
		`SELECT count(*) FROM campaign_legs l JOIN position_campaigns c ON c.id=l.campaign_id WHERE c.symbol=?`:               &legs,
	} {
		if err := db.QueryRow(query, symbol).Scan(out); err != nil {
			t.Fatalf("%s: %v", query, err)
		}
	}
	return campaigns, bindings, legs
}

// a112ReplaceBreakoutProposal 는 000660 의 봉인 제안을 다른 캠페인 · 다른 증거의 breakout 제안으로 바꾼다 — 같은 종목의 둘째 첫 레그 시도
// (config 재버전 · 같은 세션의 새 setup · 재시작 뒤 새 매니페스트 CampaignID 가 모두 이 모양: 새 캠페인 ID 의 같은 레인 · 같은 종목 제안).
func a112ReplaceBreakoutProposal(t *testing.T, fixture *a112TradingFixture, breakout strategyflow.Descriptor, campaign string) {
	t.Helper()
	result, err := strategyflow.AcceptedResultForAuthorityTest(breakout, "acct-risk-loader", a112BreakoutSymbol, campaign, 8, "100", "95", "120",
		fixture.now.Add(-time.Second), fixture.now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if result.Lineage.CampaignID == fixture.second.Proposal().Lineage.CampaignID {
		t.Fatal("arrangement: the replacement must carry a new campaign ID")
	}
	sealed, ok := strategyproposal.ProductionBatchAuthorityForTest("sha256:breakout-second-first-leg", map[string]strategyflow.Result{a112BreakoutSymbol: result}).For(a112BreakoutSymbol)
	if !ok {
		t.Fatal("arrangement: the replacement proposal could not be sealed")
	}
	kr := fixture.proposals.kr
	entries := make([]strategyProposalEntryAuthority, 0, len(kr.entries))
	for _, entry := range kr.entries {
		if entry.authority.Proposal().Lineage.Symbol == a112BreakoutSymbol {
			entry = strategyProposalEntryAuthority{authority: sealed}
		}
		entries = append(entries, entry)
	}
	kr.entries = entries
	kr.snapshot.ProposalSetDigest = strategyProposalSetDigest(kr.entries)
	fixture.proposals.kr = kr
	fixture.accounts = a112AccountLoaderWith(t, fixture.now, "", nil, "").collect(context.Background(), fixture.proposals)
	fixture.wave(t)
}

// a112BodyProbe 는 전달 몸통이 그 종목에서 CAS 를 읽었는지(몸통 도달)와 dispatch 를 불렀는지를 센다 — 층 ① 의 건너뛰기를 직접 잰다.
type a112BodyProbe struct {
	campaigns  strategyCampaignCASReader
	dispatcher strategyHandoffDispatcher
	casReads   map[string]int
	dispatches map[string]int
}

func (probe *a112BodyProbe) CurrentPositionCampaignCAS(ctx context.Context, accountRef, market, symbol string) (journal.PositionCampaignCASRead, error) {
	probe.casReads[symbol]++
	return probe.campaigns.CurrentPositionCampaignCAS(ctx, accountRef, market, symbol)
}

func (probe *a112BodyProbe) dispatch(ctx context.Context, delivered strategyhandoff.Delivered) (execgw.Outcome, error) {
	probe.dispatches[delivered.Result().Lineage.Symbol]++
	return probe.dispatcher.dispatch(ctx, delivered)
}

// 중복 평가 멱등: 같은 breakout 제안이 파도마다 다시 건너와도(재시도 · 재시작 뒤 같은 결정 재전달) 브로커 요청은 하나다.
func TestARedeliveredBreakoutFirstLegReachesTheBrokerOnce(t *testing.T) {
	fixture, _ := newA112BreakoutFixture(t)
	_ = fixture.deliverKR(t) // 첫 파도: breakout 000660 이 첫 레그(같은 파도의 005930 은 공유 버킷 stale 로 원장 거절 — 기존 시험과 같은 설계)
	if got := fixture.placedFor(a112BreakoutSymbol); got != 1 {
		t.Fatalf("first wave placed the breakout scope %d times, want 1", got)
	}
	for wave := 0; wave < 3; wave++ {
		fixture.wave(t)
		if err := fixture.deliverKR(t); err != nil && !strings.Contains(err.Error(), "BUCKET_USAGE_STALE") {
			t.Fatalf("redelivery wave %d: %v", wave, err)
		}
	}
	campaigns, bindings, legs := a112LedgerRows(t, fixture, a112BreakoutSymbol)
	if got := fixture.placedFor(a112BreakoutSymbol); got != 1 || campaigns != 1 || bindings != 1 || legs != 1 {
		t.Fatalf("after three redeliveries: broker requests=%d campaigns=%d first-leg bindings=%d legs=%d, want exactly one first leg",
			got, campaigns, bindings, legs)
	}
}

// 활성 claim 중 거절(층 ①): 첫 레그가 열린 동안 같은 종목의 다른 캠페인 breakout 제안은 전달 몸통에서 dispatch 되지 않는다.
func TestASecondBreakoutFirstLegWhileTheClaimIsActiveNeverReachesTheBroker(t *testing.T) {
	fixture, breakout := newA112BreakoutFixture(t)
	_ = fixture.deliverKR(t)
	if fixture.placedFor(a112BreakoutSymbol) != 1 {
		t.Fatal("arrangement: the first breakout leg was not placed")
	}
	a112ReplaceBreakoutProposal(t, &fixture, breakout, "campaign-breakout-second-first-leg-000660")
	probe := &a112BodyProbe{campaigns: fixture.journal, dispatcher: fixture.cycle, casReads: map[string]int{}, dispatches: map[string]int{}}
	if err := dispatchStrategyMarketHandoffs(context.Background(), probe, probe, fixture.proposals.kr.dispatchHandoffs()); err != nil &&
		!strings.Contains(err.Error(), "BUCKET_USAGE_STALE") {
		t.Fatalf("delivery of the second first-leg proposal: %v", err)
	}
	// 몸통에 도달했고(CAS 를 읽음) 그 자리에서 건너뛰었다(dispatch 0) — 거절이 앞 단계(조정자 · handoff)의 것이 아님.
	if probe.casReads[a112BreakoutSymbol] != 1 || probe.dispatches[a112BreakoutSymbol] != 0 {
		t.Fatalf("body probe for %s: CAS reads=%d dispatches=%d, want the body reached once and the dispatch skipped",
			a112BreakoutSymbol, probe.casReads[a112BreakoutSymbol], probe.dispatches[a112BreakoutSymbol])
	}
	campaigns, bindings, _ := a112LedgerRows(t, fixture, a112BreakoutSymbol)
	if got := fixture.placedFor(a112BreakoutSymbol); got != 1 || campaigns != 1 || bindings != 1 {
		t.Fatalf("second first-leg proposal on an open claim: broker requests=%d campaigns=%d bindings=%d, want the first leg alone", got, campaigns, bindings)
	}
}

// 활성 claim 중 거절(층 ②): 전달 몸통의 CAS 건너뛰기를 우회해 dispatch 를 직접 불러도 1차 레그 권한 loader 가 같은 CAS 로 거절한다 — 브로커 0.
func TestTheFirstLegAuthorityRefusesASecondBreakoutLegEvenPastTheDeliverySkip(t *testing.T) {
	fixture, breakout := newA112BreakoutFixture(t)
	_ = fixture.deliverKR(t)
	if fixture.placedFor(a112BreakoutSymbol) != 1 {
		t.Fatal("arrangement: the first breakout leg was not placed")
	}
	a112ReplaceBreakoutProposal(t, &fixture, breakout, "campaign-breakout-second-first-leg-000660")
	var refused error
	reached := false
	_ = deliverEachStrategyHandoff(fixture.proposals.kr.dispatchHandoffs(), func(delivered strategyhandoff.Delivered) error {
		if delivered.Result().Lineage.Symbol != a112BreakoutSymbol {
			return nil
		}
		reached = true
		_, refused = fixture.cycle.dispatch(context.Background(), delivered)
		return nil
	})
	if !reached {
		t.Fatal("arrangement: the replacement breakout proposal was never delivered to the body")
	}
	if refused == nil || !strings.Contains(refused.Error(), "production position campaign CAS changed") {
		t.Fatalf("direct dispatch of a second breakout first leg: err=%v, want the loader's campaign CAS refusal", refused)
	}
	campaigns, bindings, _ := a112LedgerRows(t, fixture, a112BreakoutSymbol)
	if got := fixture.placedFor(a112BreakoutSymbol); got != 1 || campaigns != 1 || bindings != 1 {
		t.Fatalf("broker requests=%d campaigns=%d bindings=%d, want the first leg alone", got, campaigns, bindings)
	}
}

// 개입 금지: 캠페인이 열린 동안 몇 파도가 돌아도(같은 제안 · 다른 캠페인 제안이 섞여도) breakout 종목에 추가 노출 레그 · scale-in 이 생기지 않는다.
func TestNoScaleInOrExtraBreakoutLegWhileTheCampaignIsOpen(t *testing.T) {
	fixture, breakout := newA112BreakoutFixture(t)
	_ = fixture.deliverKR(t)
	_, _, legsAfterFirst := a112LedgerRows(t, fixture, a112BreakoutSymbol)
	for wave := 0; wave < 4; wave++ {
		if wave == 2 {
			a112ReplaceBreakoutProposal(t, &fixture, breakout, "campaign-breakout-intervention-000660")
		} else {
			fixture.wave(t)
		}
		_ = fixture.deliverKR(t)
	}
	campaigns, bindings, legs := a112LedgerRows(t, fixture, a112BreakoutSymbol)
	if got := fixture.placedFor(a112BreakoutSymbol); got != 1 || campaigns != 1 || bindings != 1 || legs != legsAfterFirst {
		t.Fatalf("after four waves on an open campaign: broker requests=%d campaigns=%d bindings=%d legs=%d (first wave %d), want no extra exposure leg",
			got, campaigns, bindings, legs, legsAfterFirst)
	}
}
