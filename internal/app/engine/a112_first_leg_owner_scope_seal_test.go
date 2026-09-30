//go:build tossos_testseams

package engine

// a112 6.2 봉인 로트(Manager 판정 2026-10-01 — 안 (C) 「의미 봉인」). 1차 레그 권한은 건너온 봉투를 믿지 않고, 조립이 새로 고침
// 때 중재한 자기 권한 쌍에서 **소유자 범위로** 항목 하나를 골라(0 또는 복수면 거절) 그 항목의 봉인된 identity 와 대조한다.
//
// 왜 이것이 봉인인가: 엔진은 조정자를 만들고 · 먹이고 · 돌리므로 조정자가 찍은 토큰은 「어떤 중재가 돌았다」만 증명한다(엔진이 새
// 조정자에 패자 가족의 진짜 봉투만 넣으면 유효한 토큰을 얻는다). 위조가 주문이 되는 유일한 자리는 여기이고, 여기서 **조립의 중재
// 결과**와 대조하면 봉투가 어떤 철자로 만들어졌든 의미로 잡힌다.
//
// 왜 범위로 고르고 identity 로 대조하는가(HANDOFF §4 의 자기 참조 함정): accepted 와 맞는 항목을 골라 놓고 그것을 accepted 와
// 비교하면 비교가 공허해진다. 범위(계좌 · 시장 · 종목 · 포지션 세대)는 accepted 에서 읽지만 **선택 기준일 뿐 대조 대상이 아니고**,
// 대조는 선택된 항목의 봉인된 identity 와 한다 — 같은 범위의 다른 가족 · 다른 캠페인 · 조건 재작성은 선택은 되지만 대조에서 걸린다.
//
// 위조 축 다섯 — 각 시험이 **거절 문구로** 어느 판정이 거절했는지 가린다:
//   ① 같은 범위의 패자(다른 캠페인)          → 선택됨 · identity 대조에서 거절
//   ② 게이트된 레인(같은 범위 · 다른 가족 레인) → 선택됨 · identity 대조에서 거절
//   ③ 미선택 범위(다른 종목)                  → 선택 실패로 거절
//   ④ 타 시장(US 결과를 KR 자리로)            → 선택 실패로 거절
//   ⑤ 같은 계보 · 조건 재작성                  → 선택됨 · identity(ExecutionTerms) 대조에서 거절
// 기제 시험: 같은 범위가 쌍에 둘이면 선택 실패(유일성), 범위 둘인 쌍은 범위마다 선택된다(5.2.2.2 — 개수 관문 제거 뒤 발급까지).

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/continuationlane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyflow"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyhandoff"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyproposal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

// 선택 실패 문구는 identity 거절 문구를 **머리로** 담는다 — 기존 backstop 시험(`strings.Contains`)이 그대로 초록이게.
const (
	a112ScopeRefusal    = "production proposal identity changed: owner scope is not uniquely authorized by the assembly"
	a112IdentityRefusal = "production proposal identity changed"
)

func (fixture firstLegIdentityFixture) sealedKR(t *testing.T, descriptor strategyflow.Descriptor, symbol, campaignID, entry, stop,
	target string,
) strategyproposal.ProductionAuthority {
	t.Helper()
	result, err := strategyflow.AcceptedResultForAuthorityTest(descriptor, "acct-risk-loader", symbol, campaignID, 8,
		entry, stop, target, fixture.now.Add(-time.Second), fixture.now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	batch := strategyproposal.ProductionBatchAuthorityForTest("sha256:seal-"+campaignID+"-"+symbol,
		map[string]strategyflow.Result{result.Lineage.Symbol: result})
	authority, ok := batch.For(result.Lineage.Symbol)
	if !ok {
		t.Fatal("missing proposal test authority")
	}
	return authority
}

func a112KRDescriptor(t *testing.T, lane string) strategyflow.Descriptor {
	t.Helper()
	for _, descriptor := range strategyflow.Descriptors() {
		if descriptor.Market == strategyrouter.MarketKR && descriptor.LaneID == lane {
			return descriptor
		}
	}
	t.Fatalf("missing KR descriptor %s", lane)
	return strategyflow.Descriptor{}
}

// collectWithPair 는 KR 권한 쌍을 entries 로 바꾼 loader 로 accepted 를 1차 레그 권한에 넣는다.
func (fixture firstLegIdentityFixture) collectWithPair(t *testing.T, accepted strategyFirstLegAccepted,
	entries ...strategyproposal.ProductionAuthority,
) error {
	t.Helper()
	pair := fixture.proposals
	list := make([]strategyProposalEntryAuthority, 0, len(entries))
	for _, entry := range entries {
		list = append(list, strategyProposalEntryAuthority{authority: entry})
	}
	pair.kr = strategyProposalMarketAuthority{market: StrategyMarketKR, entries: list, snapshot: fixture.proposals.kr.snapshot}
	_, err := fixture.loaderWith(pair).collectStrategyFirstLegAuthority(context.Background(), accepted)
	return err
}

func a112Accepted(t *testing.T, authority strategyproposal.ProductionAuthority) strategyFirstLegAccepted {
	t.Helper()
	accepted, refusal := validateStrategyFirstLegResult(authority.Proposal())
	if refusal.Code != "" {
		t.Fatalf("arrangement: result refused before the test: %+v", refusal)
	}
	return accepted
}

// requireRefusal 은 오류가 정확히 그 판정의 문구인지 본다 — identity 거절은 선택 실패 문구를 담지 **않아야** 한다.
func requireRefusal(t *testing.T, err error, want string, what string) {
	t.Helper()
	if err == nil {
		t.Fatalf("%s: issued a first leg, want refusal %q", what, want)
	}
	got := err.Error()
	// 위조 의심(선택 실패 · identity 불일치)은 범위 거절 **타입이 아니다**(J4) — 타입이면 전달 몸통이 건너뛰고 같은 주기의 다음 범위를 낸다.
	if scope := (*strategyScopeRefusal)(nil); errors.As(err, &scope) {
		t.Fatalf("%s: err=%v is typed as a scope refusal — a forgery must stop the cycle, not be skipped", what, err)
	}
	switch want {
	case a112ScopeRefusal:
		if !strings.Contains(got, a112ScopeRefusal) {
			t.Fatalf("%s: err=%v, want the owner-scope selection refusal %q", what, err, want)
		}
	case a112IdentityRefusal:
		if !strings.HasSuffix(got, a112IdentityRefusal) || strings.Contains(got, "owner scope") {
			t.Fatalf("%s: err=%v, want the sealed-identity comparison refusal exactly (selected by scope, then refused)", what, err)
		}
	}
}

// 위조 다섯 축은 **범위 하나인 쌍**과 **범위 둘인 쌍(두 순서)** 에서 모두 거절된다(5.2.2.2 Done — 개수 관문을 걷은 뒤에도 봉인이
// 범위마다 선다). 둘째 범위(000660)는 조립이 중재한 진짜 범위라 선택 대상이 될 수 있으므로, 미선택 범위 축은 어느 쌍에도 없는
// 종목(035420)으로 잰다.
func TestTheFirstLegSealRefusesEveryForgeryAxis(t *testing.T) {
	fixture := newFirstLegIdentityFixture(t)
	winner := fixture.proposals.kr.entries[0].authority
	winnerResult := winner.Proposal()
	continuation := riskLoaderDescriptor(t, StrategyMarketKR)
	second := fixture.sealedKR(t, continuation, "000660", "campaign-seal-second-scope", "100", "95", "120")
	for _, pair := range []struct {
		name    string
		entries []strategyproposal.ProductionAuthority
	}{
		{"one-scope pair", []strategyproposal.ProductionAuthority{winner}},
		{"two-scope pair", []strategyproposal.ProductionAuthority{winner, second}},
		{"two-scope pair, coordinator order", []strategyproposal.ProductionAuthority{second, winner}},
	} {
		t.Run(pair.name, func(t *testing.T) {
			t.Run("same-scope loser (another campaign)", func(t *testing.T) {
				loser := fixture.sealedKR(t, continuation, winnerResult.Lineage.Symbol, "campaign-seal-loser", "100", "95", "120")
				requireRefusal(t, fixture.collectWithPair(t, a112Accepted(t, loser), pair.entries...), a112IdentityRefusal, "same-scope loser")
			})
			t.Run("gated lane (same scope, another family lane)", func(t *testing.T) {
				gated := fixture.sealedKR(t, a112KRDescriptor(t, reversalLaneID(t)), winnerResult.Lineage.Symbol,
					winnerResult.Lineage.CampaignID, "100", "95", "120")
				if gated.Proposal().Lineage.LaneID == winnerResult.Lineage.LaneID {
					t.Fatal("arrangement: the gated proposal is on the winner's lane")
				}
				requireRefusal(t, fixture.collectWithPair(t, a112Accepted(t, gated), pair.entries...), a112IdentityRefusal, "gated lane")
			})
			t.Run("unselected scope (another symbol)", func(t *testing.T) {
				other := fixture.sealedKR(t, continuation, "035420", "campaign-seal-other", "100", "95", "120")
				requireRefusal(t, fixture.collectWithPair(t, a112Accepted(t, other), pair.entries...), a112ScopeRefusal, "unselected scope")
			})
			t.Run("other market (a US result in the KR slot)", func(t *testing.T) {
				accepted := a112Accepted(t, fixture.proposals.us.entries[0].authority)
				accepted.market = strategyrouter.MarketKR // 위조: KR 권한 쌍으로 보내진 US 결과
				requireRefusal(t, fixture.collectWithPair(t, accepted, pair.entries...), a112ScopeRefusal, "other market")
			})
			t.Run("same lineage, rewritten execution terms", func(t *testing.T) {
				rewritten := fixture.sealedKR(t, continuation, winnerResult.Lineage.Symbol, winnerResult.Lineage.CampaignID, "100", "90", "130")
				if rewritten.Proposal().Lineage.Identity != winnerResult.Lineage.Identity {
					t.Fatal("arrangement: the rewrite moved the lineage")
				}
				requireRefusal(t, fixture.collectWithPair(t, a112Accepted(t, rewritten), pair.entries...), a112IdentityRefusal, "rewritten terms")
			})
			// 대조: 같은 쌍에서 진짜 winner 는 발급된다 — 위의 거절이 쌍 자체의 고장이 아님.
			if err := fixture.collectWithPair(t, a112Accepted(t, winner), pair.entries...); err != nil {
				t.Fatalf("control: the genuine winner was refused in the %s: %v", pair.name, err)
			}
		})
	}
}

// 기제: 같은 소유자 범위가 쌍에 둘이면 어느 쪽도 고르지 않는다(유일성) — 하나를 골라 조용히 버리지 않는다.
func TestTheFirstLegSealRefusesAnOwnerScopeTheAssemblyHoldsTwice(t *testing.T) {
	fixture := newFirstLegIdentityFixture(t)
	winner := fixture.proposals.kr.entries[0].authority
	twin := fixture.sealedKR(t, riskLoaderDescriptor(t, StrategyMarketKR), winner.Proposal().Lineage.Symbol, "campaign-seal-twin",
		"100", "95", "120")
	requireRefusal(t, fixture.collectWithPair(t, a112Accepted(t, winner), winner, twin), a112ScopeRefusal, "scope held twice")
}

// 기제: 범위 둘인 쌍에서도 선택은 범위로 된다. 5.2.2.2 가 시장 단위 개수 관문을 걷어 낸 뒤(이 시험의 앞 판은 그 관문의 거절을 단언했다)
// 범위 선택은 **발급까지** 이어진다 — 순서에 기대지 않고 양방향으로 잰다:
//
//	· winner(뒤에 둠) → 발급, 발급된 제안은 winner 의 identity(entries[0] 선택이면 other 를 골라 identity 거절)
//	· other(앞에 둠)  → 선택 · identity 대조는 통과하고 그 범위의 위험 권한이 없어 **범위 거절 타입**(단일 범위 fixture 의 위험 권한 —
//	                    identity 거절 문구가 아니므로 other 가 자기 범위로 골라졌음이 가려진다)
func TestTheFirstLegSealSelectsByScopeInATwoScopePair(t *testing.T) {
	fixture := newFirstLegIdentityFixture(t)
	winner := fixture.proposals.kr.entries[0].authority
	other := fixture.sealedKR(t, riskLoaderDescriptor(t, StrategyMarketKR), "000660", "campaign-seal-second-scope", "100", "95", "120")
	pair := fixture.proposals
	pair.kr = strategyProposalMarketAuthority{market: StrategyMarketKR, snapshot: fixture.proposals.kr.snapshot,
		entries: []strategyProposalEntryAuthority{{authority: other}, {authority: winner}}}
	loader := fixture.loaderWith(pair)
	issuance, err := loader.collectStrategyFirstLegAuthority(context.Background(), a112Accepted(t, winner))
	if err != nil || issuance.Result.Lineage.Identity != winner.Proposal().Lineage.Identity {
		t.Fatalf("two-scope pair, winner: err=%v issued=%q — want the in-scope entry issued", err, issuance.Result.Lineage.Identity)
	}
	_, err = loader.collectStrategyFirstLegAuthority(context.Background(), a112Accepted(t, other))
	var refusal *strategyScopeRefusal
	if !errors.As(err, &refusal) || refusal.scope.Symbol != "000660" {
		t.Fatalf("two-scope pair, other: err=%v — want the other scope selected and refused alone by its missing risk authority", err)
	}
}

func reversalLaneID(t *testing.T) string {
	t.Helper()
	for _, descriptor := range strategyflow.Descriptors() {
		if descriptor.Market == strategyrouter.MarketKR && descriptor.Horizon == strategyrouter.HorizonShort &&
			descriptor.LaneID != riskLoaderDescriptor(t, StrategyMarketKR).LaneID {
			return descriptor.LaneID
		}
	}
	t.Fatal("no second KR short lane")
	return ""
}

// A-lite(가독 계약 — 2차 방어, 봉인 아님): 활성화된 시장에서 건네는 목록이 조립의 제안 집합 digest 와 다르면 시장을 닫는다.
// 활성화 없는 시장(오늘 생산)은 이 계약을 보지 않는다 — 토글 OFF = upstream.
func TestAnActivatedMarketWhoseListIsNotTheArbitratedSetIsClosed(t *testing.T) {
	_, proposals, _, _ := pairedStrategyDispatchCycleFixture(t)
	two := a112TwoScopeKR(t, proposals.kr, proposals.observedAt)
	two.activation = strategyrouter.FamilyActivationForTest(strategyrouter.MarketKR, 1,
		strategyrouter.AllFourFamiliesForTest(strategyrouter.MarketKR))
	if got := two.dispatchHandoffs(); len(got) != 2 {
		t.Fatalf("arrangement: the arbitrated two-scope list gave %d handoffs", len(got))
	}
	forged := two
	forged.entries = forged.entries[:1] // 중재 결과에서 한 범위를 뺀 목록 — digest 는 그대로
	handoffs := forged.dispatchHandoffs()
	if len(handoffs) != 1 || handoffs[0].Refusal() != strategyhandoff.MarketClosed {
		t.Fatalf("a list that is not the arbitrated set: %d handoffs, first refusal %q — want one MarketClosed",
			len(handoffs), handoffs[0].Refusal())
	}
	// 활성화 없는 시장은 digest 를 보지 않는다(오늘의 시장 단위 handoff 그대로).
	unactivated := forged
	unactivated.activation = strategyrouter.FamilyActivation{}
	if got := unactivated.dispatchHandoffs(); len(got) != 1 || got[0].Refusal() != strategyhandoff.Admitted {
		t.Fatalf("toggle OFF: %d handoffs, first refusal %q — the contract must not touch the unactivated market", len(got), got[0].Refusal())
	}
}

// 계약의 식이 조립의 식과 같다 — 실제 중재 경로(collectArbitrated)가 만든 권한의 digest 로 잰다.
func TestTheProposalSetDigestMatchesWhatTheAssemblyRecords(t *testing.T) {
	now := time.Date(2026, 8, 29, 1, 2, 3, 0, time.UTC)
	pair := collectArbitrated(t, now, familyScoresForTest(strategyrouter.MarketKR), "005930", continuationlane.KRContinuationLaneID)
	if !pair.kr.snapshot.Ready || len(pair.kr.entries) == 0 {
		t.Fatalf("arrangement: KR=%+v", pair.kr.snapshot)
	}
	if got := strategyProposalSetDigest(pair.kr.entries); got != pair.kr.snapshot.ProposalSetDigest {
		t.Fatalf("contract digest %s != assembly digest %s — every activated market would close", got, pair.kr.snapshot.ProposalSetDigest)
	}
}

// 선택 함수를 직접 잰다(6.2 리뷰 codex #3): 두 범위 쌍에서 범위마다 **그 범위의** 항목을 돌려주고(순서 무관), 없는 범위는 거절한다.
func TestTheScopeSelectorReturnsTheInScopeEntry(t *testing.T) {
	fixture := newFirstLegIdentityFixture(t)
	winner := fixture.proposals.kr.entries[0].authority
	other := fixture.sealedKR(t, riskLoaderDescriptor(t, StrategyMarketKR), "000660", "campaign-seal-selector", "100", "95", "120")
	absent := fixture.sealedKR(t, riskLoaderDescriptor(t, StrategyMarketKR), "035420", "campaign-seal-absent", "100", "95", "120")
	pair := strategyProposalMarketAuthority{market: StrategyMarketKR,
		entries: []strategyProposalEntryAuthority{{authority: other}, {authority: winner}}}
	for _, want := range []strategyproposal.ProductionAuthority{winner, other} {
		got, ok := pair.authorityForOwnerScope(want.Proposal().Lineage)
		if !ok || got.Proposal().Lineage.Identity != want.Proposal().Lineage.Identity {
			t.Fatalf("scope %s: ok=%v got %s — want the in-scope entry %s", want.Proposal().Lineage.Symbol, ok,
				got.Proposal().Lineage.Identity, want.Proposal().Lineage.Identity)
		}
	}
	if _, ok := pair.authorityForOwnerScope(absent.Proposal().Lineage); ok {
		t.Fatal("a scope the pair does not hold was selected")
	}
}

// 공유 배열 경로(6.2 리뷰 codex #1 P0): dispatch 쪽 사본의 원소를 **제자리에서** 같은 계보 · 다른 손절의 봉인 제안으로 바꿔도, 1차 레그 권한은
// 구성 때 떼어 낸 조립 원본과 대조한다 — 교체된 제안은 발급되지 않는다. 편집 전(loader 가 slice 를 그대로 들던 때)에는 선택 · 대조가
// 교체된 값끼리 이루어져 발급됐다(err=nil — RED 로그).
func TestAnInPlaceSwapInTheDispatchCopyDoesNotReachTheSeal(t *testing.T) {
	for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {
		t.Run(string(market), func(t *testing.T) {
			fixture := newFirstLegIdentityFixture(t)
			loader := fixture.loaderWith(fixture.proposals)
			shared := fixture.proposals.forMarket(market)
			winner := shared.entries[0].authority.Proposal()
			entry, stop, target := "100", "80", "120"
			if market == StrategyMarketUS {
				entry, stop, target = "10000", "9000", "12000"
			}
			result, err := strategyflow.AcceptedResultForAuthorityTest(riskLoaderDescriptor(t, market), winner.Lineage.AccountRef,
				winner.Lineage.Symbol, winner.Lineage.CampaignID, winner.Quantity, entry, stop, target,
				fixture.now.Add(-time.Second), fixture.now.Add(time.Minute))
			if err != nil {
				t.Fatal(err)
			}
			batch := strategyproposal.ProductionBatchAuthorityForTest("sha256:twin-"+string(market),
				map[string]strategyflow.Result{result.Lineage.Symbol: result})
			twin, ok := batch.For(result.Lineage.Symbol)
			if !ok || twin.Proposal().Lineage.Identity != winner.Lineage.Identity ||
				twin.Proposal().ExecutionTerms.Identity() == winner.ExecutionTerms.Identity() {
				t.Fatal("arrangement: the twin must keep the lineage and change the terms")
			}
			shared.entries[0].authority = twin // 같은 배열을 공유하던 모든 사본이 바뀐다
			_, err = loader.collectStrategyFirstLegAuthority(context.Background(), a112Accepted(t, twin))
			requireRefusal(t, err, a112IdentityRefusal, "in-place swap in the "+string(market)+" dispatch copy")
		})
	}
}

// 범위 키의 축을 하나씩(6.2 리뷰 보이스 B #6): 종목이 같아도 **시장만** 또는 **계좌만** 다른 항목은 다른 범위다 — 선택이 실패해야 한다.
// (세대 축과 정규화 실패 항목은 이 시험 seam 으로 만들 수 없다 — `AcceptedResultForAuthorityTest` 는 세대 1 과 유효한 키만 봉인한다.
// 그 둘은 review 에 이름 붙여 둔다.)
func TestTheScopeSelectorKeysOnMarketAndAccountToo(t *testing.T) {
	fixture := newFirstLegIdentityFixture(t)
	winner := fixture.proposals.kr.entries[0].authority
	lineage := winner.Proposal().Lineage
	marketOnly, err := strategyflow.AcceptedResultForAuthorityTest(riskLoaderDescriptor(t, StrategyMarketUS), lineage.AccountRef,
		lineage.Symbol, "campaign-seal-market-only", 8, "10000", "9500", "12000", fixture.now.Add(-time.Second), fixture.now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	accountOnly, err := strategyflow.AcceptedResultForAuthorityTest(riskLoaderDescriptor(t, StrategyMarketKR), "acct-seal-other",
		lineage.Symbol, "campaign-seal-account-only", 8, "100", "95", "120", fixture.now.Add(-time.Second), fixture.now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	for name, result := range map[string]strategyflow.Result{"market only": marketOnly, "account only": accountOnly} {
		batch := strategyproposal.ProductionBatchAuthorityForTest("sha256:axis-"+name, map[string]strategyflow.Result{result.Lineage.Symbol: result})
		entry, ok := batch.For(result.Lineage.Symbol)
		if !ok {
			t.Fatal("missing proposal test authority")
		}
		pair := strategyProposalMarketAuthority{market: StrategyMarketKR, entries: []strategyProposalEntryAuthority{{authority: entry}}}
		if _, selected := pair.authorityForOwnerScope(lineage); selected {
			t.Fatalf("%s: an entry differing only in that axis was selected for the winner's scope", name)
		}
		requireRefusal(t, fixture.collectWithPair(t, a112Accepted(t, winner), entry), a112ScopeRefusal, name)
	}
}
