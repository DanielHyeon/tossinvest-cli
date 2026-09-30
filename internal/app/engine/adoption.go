package engine

// adoption.go is the last step of the reconciliation cycle: deciding which of
// the account's holdings the engine takes into exit management, and doing it
// (change adopt-external-positions task 2.2; exit-policy "외부 취득 포지션의
// 자동 편입", design A2/A4).
//
// # The order of the gates, and why each one is where it is
//
// A holding is judged against these in order, and the order is the point: every
// gate above a transition state is a *finding* the operator is told about, and
// every transition state is silent because it will resolve by itself.
//
//	already managed     an entry decision or an adoption already justifies it.
//	                    Nothing to do; an adopted one is additionally checked for
//	                    an external increase.
//	RECONCILE           the engine and the account disagree about this symbol.
//	                    Transition state — silent. Adopting into a disagreement
//	                    would freeze a t0 against a quantity nobody agrees on.
//	stale snapshot      the stable view is older than the freshness bound.
//	                    Transition state — silent; the next cycle re-reads.
//	adoption off        a finding: the engine is trading beside a position it
//	                    will not protect, and §0.2 keeps that alert regardless of
//	                    the toggle (design A4).
//	excluded            a finding: the operator asked for this one to be left
//	                    alone, and it is still unprotected.
//	otherwise           a candidate.
//
// # The observation, and the fifteen seconds
//
// Candidates are priced in one batched read taken immediately before the
// adoption transactions (design A6: 후보 전체를 한 번의 배치 호출로). The
// observation's age is re-checked per candidate against the price staleness
// bound, and a candidate whose observation has gone stale between the read and
// its own transaction is deferred rather than adopted (exit-policy SHALL:
// staleness 15초 초과 시 편입을 연기한다).
//
// The reason is the whole of manage-forward. `synthetic_stop` is frozen from
// that observation, and a stop frozen from a price the market has already left
// behind is a stop the first exit observation crosses — the adoption would read
// as an instant liquidation of somebody's long-held shares.
//
// # What this file does not do
//
// It issues no sell proposal. The adoption transaction records the t0 and opens
// the exit state, and that is all (SHALL NOT — design A2). The first exit
// judgement happens in the exit observation loop, on its own schedule, against
// its own observation — and a proposal from *that* is normal exit behaviour, not
// an artefact of adopting.

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reconcile"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

// candidate is one holding that passed every gate.
type candidate struct {
	position journal.Position
	holding  reconcile.Holding
}

// adoptResult 는 후보 하나가 이번 판정에서 어떻게 끝났는지임(a095 design D1 「후보별 결과」).
//
// 영값이 「연기」인 것은 의도임: adopt 가 B2(시세 읽기 오류)나 B7(관측 묵음)에서 중간에 반환하면 남은 후보는 결과 map 에
// 없고, 없는 후보는 시도되지 않은 것 — 연기 — 으로 읽혀야 함. 시도 실패는 adoptOne 이 실제로 거짓을 돌려준 자리에서만 섬.
type adoptResult int

const (
	adoptDeferred adoptResult = iota // 시도하지 못함 — 시세 읽기 오류 · 관측 없음 · 관측 묵음(adopt B2 · B6 · B7)
	adoptFailed                      // 시도했으나 실패 — adoptOne 거짓(편입 전 거절 · 영속 실패)
	adoptAdopted                     // 편입됨(보호 미개설 포함 — adoptOne B3 창, issues I7 의 이름 붙은 경계)
)

// adoptOutcome 은 후보 하나의 결과와, 시도 실패면 그 실패를 관측한 시각임. critical 문장이 말하는 시각은 보고를 만드는
// 순간이 아니라 이 시각이어야 함(Q8 — 앞 후보의 동기 전송이 전송 예산을 쓰는 동안 뒤 후보의 보고 시각이 밀려도, 그 문장은
// 실패가 일어난 때를 말해야 참임. codex 교차 리뷰 P2).
type adoptOutcome struct {
	result adoptResult
	at     time.Time
}

// 무관리 보고의 조건 칸(a095 design D1 「사실 식별자」). 사실 = (포지션, 조건)이고 durable key 와 normal 래치가 이 칸을 씀.
// 오류 문구 · 연기의 세 원인 같은 진단 원인은 칸이 아님 — 본문에만 실음.
const (
	factRejected        = "rejected"         // 편입 설정 거부
	factExcluded        = "excluded"         // 운영자 제외
	factEnabledFailed   = "enabled_failed"   // 편입 켜짐 — 시도 실패(알림 켜짐이면 critical)
	factEnabledDeferred = "enabled_deferred" // 편입 켜짐 — 연기
	factIncludeFailed   = "include_failed"   // include 지정(편입 꺼짐) — 시도 실패(알림 켜짐이면 critical)
	factIncludeDeferred = "include_deferred" // include 지정(편입 꺼짐) — 연기
	factOffUndesignated = "off_undesignated" // 편입 꺼짐 ∧ 미지정
)

// judgeHoldings runs the gates over the stable snapshot's holdings and adopts
// what passes.
func (d *ReconcileDriver) judgeHoldings(ctx context.Context, snapshot reconcile.Snapshot,
	cycle *ReconcileCycle) {
	stale := d.opts.SnapshotStaleness
	if stale <= 0 {
		stale = DefaultAdoptionStaleness
	}
	fresh := snapshot.Age(d.clk.Now()) <= stale

	var (
		candidates []candidate
		unmanaged  []journal.Position
	)
	for _, holding := range snapshot.Holdings {
		market := strings.ToLower(strings.TrimSpace(holding.Market))
		if market == "" {
			market = strings.ToLower(strings.TrimSpace(d.opts.DefaultMarket))
		}
		symbol := strings.ToUpper(strings.TrimSpace(holding.Symbol))
		if symbol == "" || market == "" || isZeroQuantity(holding.Quantity) {
			continue
		}

		p, err := d.opts.Journal.CurrentPosition(ctx, d.opts.AccountRef, market, symbol)
		if err != nil {
			// No instance yet: the fold that creates one runs earlier in this same
			// cycle, so this is a holding whose fold failed or was refused. It is
			// already reported through that path and is not a second finding.
			continue
		}
		if p.State == journal.PositionClosed || isZeroQuantity(p.Quantity) {
			continue
		}

		if p.ExitEligible() {
			// 보호 중인 포지션의 수량 증가 검사. 편입 포지션은 편입 수량이, 진입 결정으로 연 포지션은 체결이 설명하는 수량이
			// 기준임(a095 결정 (3)(i) — 전에는 편입 포지션만 봤고 엔진이 연 포지션은 아무도 보지 않았음).
			if p.Adopted() {
				d.checkExternalIncrease(ctx, p)
			} else {
				d.checkEngineOpenedIncrease(ctx, p)
			}
			continue
		}

		// --- transition states: silent -------------------------------------
		if d.blocked(market, symbol) {
			continue
		}
		if !fresh {
			continue
		}

		// --- findings ------------------------------------------------------
		// Exclusion is judged first: it wins over the global switch AND over a
		// per-symbol designation (exit-policy: exclude가 항상 우선 — 동시 등재는
		// 편입하지 않는다), so the alert can name the list the operator wrote.
		if d.opts.Adoption.Excludes(symbol) {
			unmanaged = append(unmanaged, p)
			continue
		}
		// A candidate is either globally admitted or individually designated
		// (change console-adoption-controls). The designated path runs through
		// the same gates above — nothing about RECONCILE, freshness or the
		// Stabiliser is relaxed for it.
		if !d.opts.Adoption.Enabled && !d.opts.Adoption.Included(symbol) {
			unmanaged = append(unmanaged, p)
			continue
		}
		candidates = append(candidates, candidate{position: p, holding: holding})
	}

	results := d.adopt(ctx, candidates, cycle)
	for _, c := range candidates {
		if results[c.position.ID].result != adoptAdopted {
			// A candidate the engine could not adopt is a position it is not
			// protecting, which is the same finding as an excluded one
			// (exit-policy: 제외 목록 심볼·편입 실패는 알림이 남는다).
			unmanaged = append(unmanaged, c.position)
		}
	}

	for _, p := range unmanaged {
		cycle.Unmanaged++
		// 후보가 아닌 보유(제외 · 꺼짐 · 설정 거부)는 결과 map 에 없어 영값(연기)을 받지만, 그 셋의 조건 칸은 결과를 읽지 않음.
		d.alertUnmanaged(ctx, p, results[p.ID])
	}
}

// blocked reports that the symbol is under RECONCILE.
func (d *ReconcileDriver) blocked(market, symbol string) bool {
	if d.opts.Tracker == nil {
		return false
	}
	return d.opts.Tracker.Permanent() || d.opts.Tracker.EntryAllowed(market, symbol) != nil
}

// adopt prices the candidates in one batched read and adopts what it can.
//
// It returns each candidate's result, so the caller can report the rest as
// unmanaged: a failed adoption is a position the engine is not protecting, and
// silence about it would be the alert regression design A4 exists to prevent.
// 결과는 편입됨 · 시도 실패 · 연기로 갈림(a095) — 연기와 시도 실패는 다른 사실이고, 시도 실패만 critical 이 될 수 있음.
// 중간 반환(B2 · B7)이 남긴 후보는 map 에 없고 영값 adoptDeferred 로 읽힘.
func (d *ReconcileDriver) adopt(ctx context.Context, candidates []candidate,
	cycle *ReconcileCycle) map[string]adoptOutcome {
	adopted := map[string]adoptOutcome{}
	if len(candidates) == 0 {
		return adopted
	}

	quotes, readAt, err := d.observeCandidates(ctx, candidates)
	if err != nil {
		cycle.Deferred += len(candidates)
		if cycle.Err == nil {
			cycle.Err = err
		}
		return adopted
	}

	bound := d.opts.PriceStaleness
	if bound <= 0 {
		bound = DefaultAdoptionPriceStaleness
	}
	for _, c := range candidates {
		key := adoptionQuoteKey(c.position.Market, c.position.Symbol)
		observed, ok := quotes[key]
		if !ok {
			// The read answered for other symbols and not this one. There is no
			// t0 to freeze, so the holding waits for the next cycle.
			cycle.Deferred++
			continue
		}
		if age := d.clk.Now().Sub(readAt); age > bound {
			// The observation went stale between the read and this transaction.
			// Freezing a synthetic stop from it would be freezing a stop the
			// market may already have passed.
			cycle.Deferred += len(candidates) - adoptedCount(adopted)
			d.logDeferred(c.position, fmt.Sprintf(
				"the price observation is %s old, past the %s bound", age, bound))
			return adopted
		}
		if d.adoptOne(ctx, c, observed) {
			adopted[c.position.ID] = adoptOutcome{result: adoptAdopted}
			cycle.Adopted++
			continue
		}
		adopted[c.position.ID] = adoptOutcome{result: adoptFailed, at: d.clk.Now()}
		cycle.Deferred++
	}
	return adopted
}

// adoptedCount 는 이번 판정에서 편입된 후보 수임 — B7 의 연기 셈이 시도 실패를 편입으로 세지 않게 함.
func adoptedCount(results map[string]adoptOutcome) int {
	n := 0
	for _, r := range results {
		if r.result == adoptAdopted {
			n++
		}
	}
	return n
}

// observeCandidates is the one batched price read.
func (d *ReconcileDriver) observeCandidates(ctx context.Context, candidates []candidate) (
	map[string]string, time.Time, error) {
	marketsBySymbol := map[string]map[string]struct{}{}
	symbols := make([]string, 0, len(candidates))
	for _, c := range candidates {
		symbol := strings.ToUpper(strings.TrimSpace(c.position.Symbol))
		market := strings.ToLower(strings.TrimSpace(c.position.Market))
		if symbol == "" || market == "" {
			continue
		}
		markets := marketsBySymbol[symbol]
		if markets == nil {
			markets = map[string]struct{}{}
			marketsBySymbol[symbol] = markets
			symbols = append(symbols, symbol)
		}
		markets[market] = struct{}{}
	}
	sort.Strings(symbols)

	// Stamped before the request, not after. The quote the broker returns is at
	// best as fresh as the instant it was asked for, and a retried read can take
	// the whole query budget — so measuring the observation's age from the moment
	// the read *started* is the conservative reading of "the observation
	// immediately before the transaction". Measuring from the reply would let a
	// slow read hand back a price the market has already left behind and call it
	// zero seconds old.
	readAt := d.clk.Now()

	var quotes []domain.Quote
	read := func(ctx context.Context) error {
		var err error
		quotes, err = d.opts.Prices.Prices(ctx, symbols)
		return err
	}
	var err error
	if d.opts.Retrier != nil {
		err = d.opts.Retrier.Query(ctx, execgw.QueryPrice, read)
	} else {
		err = read(ctx)
	}
	if err != nil {
		return nil, time.Time{}, fmt.Errorf(
			"engine: observing %d adoption candidate(s): %w", len(symbols), err)
	}

	out := make(map[string]string, len(quotes))
	for _, q := range quotes {
		if q.Last <= 0 {
			// A quote with no last trade is not an observation. Recording it as
			// zero would freeze a synthetic stop of zero.
			continue
		}
		symbol := strings.ToUpper(strings.TrimSpace(q.Symbol))
		markets := marketsBySymbol[symbol]
		if len(markets) != 1 {
			// A symbol-only quote cannot prove which candidate market it belongs
			// to when the same symbol is present in more than one market.
			continue
		}
		market := ""
		for candidateMarket := range markets {
			market = candidateMarket
		}
		expectedCurrency, ok := adoptionCurrency(market)
		if !ok || !strings.EqualFold(strings.TrimSpace(q.Currency), expectedCurrency) {
			continue
		}
		out[adoptionQuoteKey(market, symbol)] = decimalOf(q.Last)
	}
	return out, readAt, nil
}

func adoptionQuoteKey(market, symbol string) string {
	return strings.ToLower(strings.TrimSpace(market)) + "\x00" +
		strings.ToUpper(strings.TrimSpace(symbol))
}

func adoptionCurrency(market string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(market)) {
	case "kr":
		return "KRW", true
	case "us":
		return "USD", true
	default:
		return "", false
	}
}

// adoptOne persists one adoption and opens its exit state.
//
// The two are separate transactions on purpose, and the ordering is the crash
// contract (exit-policy: 편입 기록 영속 후 크래시). The adoption commits first;
// if the process dies before the exit state opens, the position is left carrying
// an adoption and no protection, which the *next* cycle — or the exit
// observation loop's own working set — completes from the record on disk. The
// other order would leave an exit state whose t0 nothing explains.
func (d *ReconcileDriver) adoptOne(ctx context.Context, c candidate, observed string) bool {
	stop, err := exitpolicy.SyntheticStop(observed, d.opts.Adoption.DefaultStopPct)
	if err != nil {
		d.logDeferred(c.position, "the synthetic stop could not be derived: "+err.Error())
		return false
	}

	adoption, err := d.opts.Journal.AdoptPosition(ctx, journal.AdoptionRequest{
		PositionID: c.position.ID,
		Symbol:     c.position.Symbol,
		Market:     c.position.Market,
		Quantity:   c.position.Quantity,
		// The broker's own string, never a re-rendering of the float beside it.
		// Empty when the snapshot could not preserve one, which the record stores
		// as ABSENT.
		CostBasis:     c.holding.CostBasisRaw,
		ObservedPrice: observed,
		SyntheticStop: stop,
		ObservedAt:    journal.RFC3339(d.clk.Now()),
		ExitPolicyID:  d.opts.CommonPolicy,
	})
	if err != nil {
		d.logDeferred(c.position, "the adoption was refused: "+err.Error())
		return false
	}

	if _, err := d.opts.Journal.OpenAdoptedExitState(ctx, c.position.ID); err != nil {
		// The record is committed, so the position *is* adopted and the next pass
		// completes the opening. It is reported as adopted rather than deferred
		// for exactly that reason: telling the operator it was not would be
		// wrong, and the alert says the protection is not open yet.
		d.logDeferred(c.position, "the adoption is recorded but the exit state is not open yet: "+
			err.Error())
	}

	d.alert(ctx, obs.Event{
		Type:  obs.EventExitPositionAdopted,
		Key:   string(obs.EventExitPositionAdopted) + "|" + c.position.ID,
		Title: d.label(c.position.Symbol) + " 엔진 관리로 편입됐다",
		Body: "계좌가 보유 중인데 이를 설명하는 진입 결정이 없어 편입 기록을 남겼다. 지금부터 보호선 승격," +
			"중간 익절, 손절이 엔진이 직접 연 포지션과 동일하게 적용된다.\n" +
			"기준선은 매입 원가가 아니라 **편입 시점에 관측된 가격**으로 잡는다 — 오래 들고 있던 " +
			"수익 종목은 지금부터 보호되고, 과거 수익은 R 척도에 포함되지 않는다.",
		Fields: map[string]any{
			obs.FieldAccount:  d.opts.AccountRef,
			obs.FieldSymbol:   c.position.Symbol,
			obs.FieldQuantity: adoption.Quantity,
			"market":          c.position.Market,
			"position_id":     c.position.ID,
			"adoption_id":     adoption.ID,
			"observed_price":  adoption.ObservedPrice,
			"synthetic_stop":  adoption.SyntheticStop,
			"cost_basis":      adoption.CostBasis,
			"cost_basis_src":  adoption.CostBasisSource,
			"observed_at":     adoption.ObservedAt,
			"exit_policy_id":  adoption.ExitPolicyID,
		},
	})
	// A position that has just been adopted is no longer unmanaged; drop the
	// latch so a *later* loss of eligibility would alert again.
	delete(d.unmanaged, c.position.ID)
	return true
}

// alertUnmanaged reports a holding confirmed to be outside exit management.
//
// It fires regardless of `adoption.enabled` (SHALL — design A4). That toggle
// decides whether the engine *acts*; it does not decide whether the operator is
// told that the engine is trading beside a position it will not protect.
//
// 등급은 사실이 정함(a095 — engine-safety 「무관리 보유 보고의 등급은 사실이 정한다」):
//
//	critical ⇔ 알림 켜짐(로드된 설정) ∧ 조건 ∈ {편입 켜짐 — 시도 실패, include 지정 — 시도 실패}
//	          (include 지정은 운영자가 그 종목의 보호를 고른 것 — 정본 exit-policy 「종목별 편입」: include 경유 편입은 알림 규칙
//	          전부에서 enabled 경유와 동일. 빈 include 목록의 엔진은 무변화)
//	그 밖(제외 · 꺼짐∧미지정 · 설정 거부 · 연기 · 알림 꺼짐) → normal, 오늘 등급 그대로
//
// 억제도 등급을 따름. normal 은 (포지션, 조건) 메모리 래치로 같은 사실의 반복만 한 번 알림 — 1분마다 반복되는 알림은
// 아무도 읽지 않음. 래치는 메모리라 재시작이 다시 올림(운영자가 가장 볼 때). critical 은 래치를 거치지 않음 — 매 관측이
// durable 기록을 시도하고 중복은 outbox key 와 정본 재알림 창이 맡음. 앞에 래치를 겹치면 기록 실패 · 창 뒤 재알림이 영구히
// 삼켜짐(a095 6판 원칙, 5라운드 R5-1).
func (d *ReconcileDriver) alertUnmanaged(ctx context.Context, p journal.Position, outcome adoptOutcome) {
	fact, why := d.unmanagedFact(p, outcome.result)
	if d.opts.NotificationsEnabled && (fact == factEnabledFailed || fact == factIncludeFailed) {
		d.alertAdoptionFailed(ctx, p, fact, outcome.at)
		return
	}
	if d.unmanaged[p.ID][fact] {
		return
	}
	if d.unmanaged[p.ID] == nil {
		d.unmanaged[p.ID] = map[string]bool{}
	}
	d.unmanaged[p.ID][fact] = true

	d.alert(ctx, obs.Event{
		Type:  obs.EventExitPositionUnmanaged,
		Key:   reconcileUnmanagedKey(obs.EventExitPositionUnmanaged, fact, p.ID),
		Title: d.label(p.Symbol) + " 보유 중이지만 엔진이 관리하지 않는다",
		Body: "이 포지션에는 진입 결정도 편입 기록도 없어 기준선을 만들 손절도, R을 잴 최초 위험도 없다. " +
			"손절·익절이 자동으로 걸려 있지 않다.\n사유: " + why,
		Fields: map[string]any{
			obs.FieldAccount:   d.opts.AccountRef,
			obs.FieldSymbol:    p.Symbol,
			obs.FieldQuantity:  p.Quantity,
			"market":           p.Market,
			"position_id":      p.ID,
			"adoption_enabled": d.opts.Adoption.Enabled,
			"fact":             fact,
		},
	})
}

// alertAdoptionFailed 는 알림이 켜진 엔진에서 엔진이 보호하기로 한 보유의 편입 시도가 실패한 critical 보고임.
//
// 문장은 **그 관측 시각의 사건**임(a095 Q8 답 (b)). outbox 행은 전달 성공이나 운영자 승인으로만 정산되고 편입의 회복은 그 행을
// 갱신하지 않으므로, 「지금 무보호」라고 쓰면 해소 뒤 늦게 배달된 행이 거짓을 전함. 시각을 담은 과거형 문장은 언제 배달돼도 참임.
// PENDING 행의 본문은 재무장되지 않아 첫 실패의 시각이 남고, 창이 지난 정착 행의 재무장은 그 관측의 문장으로 바뀜.
func (d *ReconcileDriver) alertAdoptionFailed(ctx context.Context, p journal.Position, fact string, failedAt time.Time) {
	if failedAt.IsZero() {
		failedAt = d.clk.Now() // 시도 실패는 늘 시각을 싣지만, 빠졌으면 보고 순간으로 — 문장이 시각을 잃지 않게
	}
	moment := journal.RFC3339(failedAt)
	d.alert(ctx, obs.Event{
		Type:  obs.EventExitPositionAdoptionFailed,
		Key:   reconcileUnmanagedKey(obs.EventExitPositionAdoptionFailed, fact, p.ID),
		Title: d.label(p.Symbol) + " 편입 시도 실패 — 보호가 열리지 않았다",
		Body: moment + "(UTC)에 편입 시도가 실패했다. 편입 설정(전체 켜짐 또는 종목 지정)으로 엔진이 보호하기로 한 보유인데, 그 시각 기준 이 보유에는 " +
			"손절·익절이 걸려 있지 않았다.\n편입 설정이 그대로면 다음 대사 사이클이 다시 시도한다 — 이 알림이 늦게 도착했다면 그 사이 편입이 " +
			"끝났을 수 있으니 콘솔에서 현재 상태를 확인하라. 이 알림의 승인은 운영자가 한다.",
		Fields: map[string]any{
			obs.FieldAccount:  d.opts.AccountRef,
			obs.FieldSymbol:   p.Symbol,
			obs.FieldQuantity: p.Quantity,
			"market":          p.Market,
			"position_id":     p.ID,
			"fact":            fact,
			"observed_at":     moment,
		},
	})
}

// unmanagedFact 는 무관리 보고의 조건 칸과 운영자에게 보일 사유를 고름(why-matrix, change console-adoption-controls
// design D2 + a095 조건 칸). 순서가 판정의 일부임 — 설정 거부 → 제외 → 편입 켜짐 → include 지정 → 기본. 편입이 켜진
// 엔진의 include 지정 종목은 「편입 켜짐」 칸에 듦(a095 9판 r8 N8). 시도한 지정 종목에게 "adoption is off"라고 말하면
// 운영자를 틀린 설정으로 보냄 — 이 switch 가 절대 해서는 안 되는 일.
func (d *ReconcileDriver) unmanagedFact(p journal.Position, result adoptResult) (fact, why string) {
	switch {
	case d.opts.Adoption.Rejected != "":
		return factRejected, "the adoption settings were refused, so adoption is off: " + d.opts.Adoption.Rejected
	case d.opts.Adoption.Excludes(p.Symbol):
		return factExcluded, "the symbol is on adoption.exclude_symbols, so it is deliberately left unprotected"
	case d.opts.Adoption.Enabled:
		if result == adoptFailed {
			return factEnabledFailed, "adoption is on but the attempt to adopt this holding failed this cycle; " +
				"it is unprotected until a cycle succeeds"
		}
		return factEnabledDeferred, "adoption is on but this holding could not be tried this cycle (no usable " +
			"price observation); it is unprotected until a cycle adopts it"
	case d.opts.Adoption.Included(p.Symbol):
		if result == adoptFailed {
			return factIncludeFailed, "the symbol is designated on adoption.include_symbols and the adoption " +
				"was tried, but this cycle failed; it is unprotected until a cycle succeeds"
		}
		return factIncludeDeferred, "the symbol is designated on adoption.include_symbols but could not be " +
			"tried this cycle (no usable price observation); it is unprotected until a cycle adopts it"
	}
	return factOffUndesignated, "adoption is off and the symbol is not designated, so the engine records the " +
		"holding and leaves it alone"
}

// reconcileUnmanagedKey 는 대사 자리의 무관리 보고 key 임 — 종류 | 발신 자리 | 조건 | 포지션.
// exit 관측 자리(exitloop.go alertUnmanaged)의 `exit.position_unmanaged|<posID>` 와 늘 다름 — 같은 key 는 outbox 한
// 행으로 합쳐져 먼저 온 자리의 사유만 남음(a095 결정 (3)(iii)).
func reconcileUnmanagedKey(t obs.EventType, fact, positionID string) string {
	return string(t) + "|reconcile|" + fact + "|" + positionID
}

// checkExternalIncrease reports an adopted position the owner has since bought
// more of (design A8).
//
// The t0 is deliberately not recomputed: `exit_states` freezes the entry, the
// initial risk and the initial quantity, and moving any of them would rewrite
// the denominator every R on that position has already been expressed in.
// Reporting it is what the operator can act on; re-sizing is a later change.
//
// 래치는 (포지션, 보고한 최대 수량)임(a095 Q4) — 한 번 보고하고 영구히 침묵하면 뒤의 큰 증가(475150: 3 → … → 32)가
// 지워짐. 등급은 normal: 손절은 발동 시점의 투영 수량 전량을 청산하므로 증가분도 보호받음(review §4.4).
func (d *ReconcileDriver) checkExternalIncrease(ctx context.Context, p journal.Position) {
	adoption, err := d.opts.Journal.AdoptionOf(ctx, p.ID)
	if err != nil {
		return
	}
	cmp, err := riskcalc.CompareDecimal(p.Quantity, adoption.Quantity)
	if err != nil || cmp <= 0 {
		return
	}
	if !d.newGrowthMaximum(p) {
		return
	}

	d.alert(ctx, obs.Event{
		Type:  obs.EventExitPositionUnmanaged,
		Key:   string(obs.EventExitPositionUnmanaged) + "|grown|" + p.ID,
		Title: d.label(p.Symbol) + " 편입 후 수량이 늘었고 고정된 t0가 증가분을 덮지 않는다",
		Body: "계좌의 이 종목 수량이 편입 기록보다 많다. exit state의 진입가·최초 위험·최초 수량은 " +
			"고정된 채로 둔다 — 다시 계산하면 이 포지션에 대해 이미 보고된 모든 R의 분모가 바뀐다.\n" +
			"따라서 늘어난 수량은 **원래 수량 기준으로 산정된 손절**의 보호를 받는다.",
		Fields: map[string]any{
			obs.FieldAccount:   d.opts.AccountRef,
			obs.FieldSymbol:    p.Symbol,
			obs.FieldQuantity:  p.Quantity,
			"market":           p.Market,
			"position_id":      p.ID,
			"adopted_quantity": adoption.Quantity,
			"adoption_id":      adoption.ID,
		},
	})
}

// checkEngineOpenedIncrease 는 진입 결정으로 엔진이 연 포지션의 수량이 체결로 설명되지 않게 늘었는지 봄(a095 결정 (3)(i),
// Q3 답).
//
// 기준은 원장 조정의 순증임: 대사 수렴은 계좌가 체결로 설명되지 않는 수량을 보이면 조정 행을 더하고 투영을 계좌 값으로 옮김
// (converge.go ConvergeQuantities → ApplyPositionAdjustment — 행의 prev_quantity 는 투영이 들던 값, new_quantity 는 계좌 값,
// 살아 있는 인스턴스는 제자리 조정). 그러니 이 인스턴스 조정들의 Σ(new − prev) 가 0 보다 크면 계좌가 체결보다 많이 들고 있음.
// 조회 오류는 조용히 반환함 — 다음 사이클이 다시 봄(checkExternalIncrease B2 와 같은 규칙). 브로커 호출 없음(원장 읽기뿐).
func (d *ReconcileDriver) checkEngineOpenedIncrease(ctx context.Context, p journal.Position) {
	adjustments, err := d.opts.Journal.PositionAdjustments(ctx, p.ID)
	if err != nil || len(adjustments) == 0 {
		return
	}
	net, err := netAdjustedQuantity(adjustments)
	if err != nil {
		return
	}
	cmp, err := riskcalc.CompareDecimal(net, "0")
	if err != nil || cmp <= 0 {
		return
	}
	if !d.newGrowthMaximum(p) {
		return
	}

	d.alert(ctx, obs.Event{
		Type:  obs.EventExitPositionUnmanaged,
		Key:   string(obs.EventExitPositionUnmanaged) + "|grown|" + p.ID,
		Title: d.label(p.Symbol) + " 엔진이 연 포지션의 수량이 체결로 설명되지 않게 늘었다",
		Body: "계좌의 이 종목 수량이 엔진의 체결이 설명하는 수량보다 " + net + "주 많다(원장 조정의 순증). exit state의 " +
			"진입가·최초 위험은 고정된 채로 둔다 — 다시 계산하면 이미 보고된 모든 R의 분모가 바뀐다.\n" +
			"손절은 발동 시점의 보유 전량을 청산하므로 늘어난 수량도 **기존 손절선**의 보호를 받는다.",
		Fields: map[string]any{
			obs.FieldAccount:       d.opts.AccountRef,
			obs.FieldSymbol:        p.Symbol,
			obs.FieldQuantity:      p.Quantity,
			"market":               p.Market,
			"position_id":          p.ID,
			"unexplained_by_fills": net,
		},
	})
}

// netAdjustedQuantity 는 조정 행들의 Σ(new_quantity − prev_quantity) 임 — 투영이 체결에서 벗어난 순량.
func netAdjustedQuantity(adjustments []journal.Adjustment) (string, error) {
	net := "0"
	for _, a := range adjustments {
		delta, err := riskcalc.SubDecimal(a.NewQuantity, a.PrevQuantity)
		if err != nil {
			return "", err
		}
		if net, err = riskcalc.AddDecimal(net, delta); err != nil {
			return "", err
		}
	}
	return net, nil
}

// newGrowthMaximum 은 현재 수량이 이 포지션에서 보고한 최대 수량보다 클 때만 참이고 그때 래치를 옮김(a095 Q4 — 같은 수량의
// 반복과 감소는 다시 보고하지 않음). 비교 실패는 보고하지 않는 쪽으로 틀림 — normal 보고이므로 진입 · 청산에 닿지 않음.
func (d *ReconcileDriver) newGrowthMaximum(p journal.Position) bool {
	if last, ok := d.grown[p.ID]; ok {
		cmp, err := riskcalc.CompareDecimal(p.Quantity, last)
		if err != nil || cmp <= 0 {
			return false
		}
	}
	d.grown[p.ID] = p.Quantity
	return true
}

func (d *ReconcileDriver) logDeferred(p journal.Position, why string) {
	if d.opts.Log == nil {
		return
	}
	d.opts.Log.Event(obs.EventEngineHeartbeat,
		obs.FieldAccount, d.opts.AccountRef,
		obs.FieldSymbol, p.Symbol,
		"position_id", p.ID,
		obs.FieldDetail, "the adoption was deferred: "+why)
}
