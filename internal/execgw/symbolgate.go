package execgw

// symbolgate.go gives EntryGate a symbol dimension (harden-execution-base task
// 4.2; issues.md 2026-07-26 "EntryGate에 심볼 차원이 없다").
//
// # What was wrong
//
// Two specs ask for symbol-scoped blocks — fill-detection says an
// UNKNOWN_BROKER_STATE blocks "해당 심볼", reconciliation defines a block scope of
// account/market/symbol — and the gate had only one scope: the whole account.
// Producers therefore had a choice between blocking too much (latch the account)
// and blocking nothing, and both tasks 3.2 and 3.6 correctly chose too much.
//
// Too much is safe but not free. An account-wide latch raised by a disagreement
// about one symbol also stops the engine from opening a position in an unrelated
// one, which over a trading day means the safety mechanism, not the market, is
// what decides whether the engine trades. The state table in
// internal/reconcile/mismatch.go already says which rows are symbol-scoped; this
// file is what lets the gate honour it.
//
// # What is deliberately unchanged
//
//   - Symbol blocks stop *entries* only, exactly like every other block here.
//     Cancels and liquidations are never gated (§0.3), and there is no method in
//     this file that could gate one.
//   - CheckEntry() keeps its old meaning — the account-wide question — so every
//     existing caller keeps its existing answer. The narrowing happens because the
//     gateway now asks the *narrower* question, not because the old one changed.
//   - Release semantics are the producer's business: a symbol block clears when
//     ClearSymbol is called, and whether that happens automatically (a clean
//     reconcile) or only by hand (UNKNOWN_BROKER_STATE) is decided by the state
//     table, not here.

import (
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

// SymbolBlock is one active per-symbol entry block.
type SymbolBlock struct {
	// Market is "kr"/"us". Empty means the block applies to the symbol in every
	// market, which is what a caller that does not know the market gets.
	Market string
	Symbol string
	Reason ReasonCode
	Detail string
	Since  time.Time
}

// symbolKey identifies a block for deduplication.
func symbolKey(market, symbol string, reason ReasonCode) string {
	return strings.ToLower(strings.TrimSpace(market)) + "|" +
		strings.ToUpper(strings.TrimSpace(symbol)) + "|" + string(reason)
}

// BlockSymbol latches new entries in one symbol.
//
// An empty market blocks the symbol everywhere it trades. Re-blocking an already
// blocked (market, symbol, reason) keeps the original detail and timestamp: the
// first observation is the one worth showing an operator, and a block whose
// "since" advances every poll cycle looks new forever.
func (g *EntryGate) BlockSymbol(market, symbol string, reason ReasonCode, detail string) {
	if strings.TrimSpace(symbol) == "" {
		// A symbol block with no symbol would silently block nothing. Latch the
		// account instead: whatever the caller observed, it observed something.
		g.Block(reason, detail)
		return
	}
	key := symbolKey(market, symbol, reason)

	g.mu.Lock()
	defer g.mu.Unlock()
	if g.symbolLatches == nil {
		g.symbolLatches = make(map[string]SymbolBlock)
	}
	if _, exists := g.symbolLatches[key]; exists {
		return
	}
	g.symbolLatches[key] = SymbolBlock{
		Market: strings.ToLower(strings.TrimSpace(market)),
		Symbol: strings.ToUpper(strings.TrimSpace(symbol)),
		Reason: reason,
		Detail: detail,
		Since:  g.clk.Now(),
	}
}

// ClearSymbol releases one symbol block.
func (g *EntryGate) ClearSymbol(market, symbol string, reason ReasonCode) {
	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.symbolLatches, symbolKey(market, symbol, reason))
}

// ClearSymbolReason releases every symbol block carrying a reason, whatever
// symbol it names. It is what a producer calls after a clean sweep — a
// reconciliation that agreed about everything cannot leave a mismatch block
// standing on a symbol it no longer disagrees about.
func (g *EntryGate) ClearSymbolReason(reason ReasonCode) {
	g.mu.Lock()
	defer g.mu.Unlock()
	for key, block := range g.symbolLatches {
		if block.Reason == reason {
			delete(g.symbolLatches, key)
		}
	}
}

// SymbolBlocks lists the active per-symbol blocks.
func (g *EntryGate) SymbolBlocks() []SymbolBlock {
	g.mu.Lock()
	defer g.mu.Unlock()
	out := make([]SymbolBlock, 0, len(g.symbolLatches))
	for _, block := range g.symbolLatches {
		out = append(out, block)
	}
	return out
}

// --- the journal's RECONCILE states, projected onto the gate (task 4.1) ------

// ReconcileReasonFor maps a persisted RECONCILE cause onto the gate's reason
// code.
//
// The split is "can a later read disprove this":
//
//   - a snapshot that could not be read, was stale, or disagreed about a
//     quantity is a condition a fresh, agreeing comparison closes, so it lands
//     on ReasonReconcileMismatch (auto-released by a clean reconcile);
//   - an identifier in conflicting contexts, or a broker record nothing local
//     can be attributed to, is not. openapi documents CANCEL_REJECTED and
//     REPLACE_REJECTED as separate order records whose concrete shape is
//     [형태 미측정 — 2b 2.1], so "we saw it again and it still does not fit"
//     is not evidence that it has been resolved. Those land on
//     ReasonReconcilePermanent, which nothing automatic clears.
//
// An unknown cause maps to the operator-only reason. The journal refuses to
// write one, so reaching this branch means the row predates this build — and
// blocking wider than necessary is the safe direction.
func ReconcileReasonFor(cause string) ReasonCode {
	switch cause {
	case journal.ReconcileCauseSnapshotUnavailable,
		journal.ReconcileCauseSnapshotStale,
		journal.ReconcileCauseQuantityMismatch:
		return ReasonReconcileMismatch
	default:
		return ReasonReconcilePermanent
	}
}

// RebuildReconcileProjection replaces every reconcile-family block with the
// journal's active RECONCILE states.
//
// This is the gate's half of "journal이 권위, in-memory 차단 기계는 기동 시
// 재구성되는 투영이다"(SHALL). A restart calls it once, after opening the
// journal and before the first entry decision: without it a process that
// stopped mid-disagreement comes back with an empty latch map and trades.
//
// It replaces rather than merges, on purpose. A latch this process raised in a
// previous life but that the journal does not carry is a latch nothing can
// release — the row an operator would close does not exist.
//
// A state carries no market (reconcile_states has no market column — D9), so a
// symbol-scoped state blocks that symbol in every market it trades. That is the
// conservative reading of a scope the schema does not record.
func (g *EntryGate) RebuildReconcileProjection(states []journal.ReconcileState) {
	family := []ReasonCode{ReasonReconcileMismatch, ReasonReconcilePermanent}

	g.mu.Lock()
	for _, reason := range family {
		delete(g.latches, reason)
	}
	for key, block := range g.symbolLatches {
		for _, reason := range family {
			if block.Reason == reason {
				delete(g.symbolLatches, key)
			}
		}
	}
	g.mu.Unlock()

	for _, state := range states {
		if !state.Active() {
			continue
		}
		reason := ReconcileReasonFor(state.Cause)
		detail := state.Cause
		if state.Evidence != "" {
			detail += ": " + state.Evidence
		}
		if state.AccountWide() {
			g.Block(reason, detail)
			continue
		}
		g.BlockSymbol("", state.Symbol, reason, detail)
	}
}

// CheckEntryFor reports why a new entry in this market and symbol is refused, or
// nil when it is allowed.
//
// Account-wide blocks are evaluated first, so the reason an operator sees is the
// widest one that applies: "the credential was rejected" is more actionable than
// "005930 disagrees", and if both hold, fixing the second changes nothing.
//
// An empty symbol asks the account-wide question only — that is what CheckEntry
// does, and it is why an existing caller's answer is unchanged.
func (g *EntryGate) CheckEntryFor(market, symbol string) *RejectedError {
	if rejected := g.checkAccountEntry(); rejected != nil {
		return rejected
	}
	if strings.TrimSpace(symbol) == "" {
		return nil
	}

	g.mu.Lock()
	defer g.mu.Unlock()
	// latchOrder gives a deterministic reason when several symbol blocks apply,
	// for the same reason it does for account latches: an operator chasing a
	// reason code that changes between polls stops trusting it.
	for _, reason := range latchOrder {
		if block, blocked := g.symbolLatches[symbolKey(market, symbol, reason)]; blocked {
			return &RejectedError{Reason: block.Reason, Detail: block.Detail}
		}
		// A block registered without a market applies in every market.
		if block, blocked := g.symbolLatches[symbolKey("", symbol, reason)]; blocked {
			return &RejectedError{Reason: block.Reason, Detail: block.Detail}
		}
	}
	for _, block := range g.symbolLatches {
		if !strings.EqualFold(block.Symbol, strings.TrimSpace(symbol)) {
			continue
		}
		if block.Market != "" && !strings.EqualFold(block.Market, strings.TrimSpace(market)) {
			continue
		}
		return &RejectedError{Reason: block.Reason, Detail: block.Detail}
	}
	return nil
}
