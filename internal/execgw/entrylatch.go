package execgw

// entrylatch.go is add-core-domain task 2.4: the entry gate's verdict, published
// in the shape the Guardian chain's input struct reads.
//
// # Why this is a projection and not a second judgement
//
// risk-management puts "게이트 상태(진입 차단 latch: 401/403·SLO 위반·RECONCILE·
// recovery 미완료)" in the chain, and the tempting way to implement that is to
// give internal/risk its own view of the four conditions. That would be two
// judgements of one fact, and the two would eventually disagree — the gate
// auto-clears a staleness block the moment a poll succeeds, latches an auth
// failure until an operator clears it, and narrows reconciliation blocks to a
// symbol. A copy of those rules in another package is a copy that will be wrong
// on the day it matters.
//
// So the chain does not re-derive anything. This file asks the gate the question
// the gate already answers — CheckEntryFor, the same call the sealed submission
// sequence makes — and hands back the two plain values the chain's AccountState
// carries. The chain's role is to stop *before* a decision is recorded and a
// reservation taken; the gateway's re-check at submission time is the
// enforcement point, and it is unchanged (engine-safety 봉인된 제출 시퀀스).
//
// # What that means for double blocking
//
// One condition produces one refusal in each place it is asked, never two
// refusals for one intent:
//
//   - Guardian refuses first, with ENTRY_GATE_BLOCKED, and no decision row is
//     written. Nothing reaches the gateway, so the gateway refuses nothing.
//   - If the latch is raised *after* Guardian allowed and before the submission
//     lands, the gateway refuses with the gate's own code and the decision is
//     recorded as refused. That is the intended asymmetry: it is the same fact
//     observed at two instants, not the same instant judged twice.
//
// # The operating mode is deliberately in here too
//
// From task 3.1 the account's operating mode is projected onto this gate as a
// latch (ReasonOperatingModeBlocked), so it surfaces through this surface as
// well. The chain still evaluates its own mode rung *before* the latch rung,
// which is what makes the operator see OPERATING_MODE_BLOCKED — the specific,
// actionable reason — rather than the generic latch code. The two agree because
// both read the same journal row; the ordering only decides which name the
// refusal gets.

// EntryLatch is the entry gate's account-wide (or symbol-aware) verdict as plain
// values.
//
// The zero value means "not blocked", so a caller that never wires a gate gets
// the same shape as one whose gate is clear. That is safe here and only here:
// the absence of a gate is not the absence of the checks the chain does itself,
// and the gateway's own re-check is what a missing gate would still not bypass.
type EntryLatch struct {
	// Blocked is the chain's AccountState.EntryBlockedLatch.
	Blocked bool
	// Reason is the gate's own code. It travels into the chain's
	// AccountState.EntryBlockedReason as a string, so a Guardian refusal names
	// the gateway vocabulary that produced it and the two records join.
	Reason ReasonCode
	// Detail is the gate's operator-facing explanation, unmodified.
	Detail string
}

// EntryLatchFor reports the gate's verdict for one market and symbol.
//
// It is CheckEntryFor's answer in another shape — including the precedence
// latchOrder fixes, so the reason a Guardian refusal carries is the same reason
// the status output and the gateway would show for the same account state.
func (g *EntryGate) EntryLatchFor(market, symbol string) EntryLatch {
	rejected := g.CheckEntryFor(market, symbol)
	if rejected == nil {
		return EntryLatch{}
	}
	return EntryLatch{Blocked: true, Reason: rejected.Reason, Detail: rejected.Detail}
}

// EntryLatch reports the account-wide verdict, ignoring per-symbol blocks.
//
// Callers that know the symbol should use EntryLatchFor: a symbol-scoped block is
// a real refusal for that symbol, and asking the wider question would let the
// chain allow an intent the gateway is about to refuse.
func (g *EntryGate) EntryLatch() EntryLatch { return g.EntryLatchFor("", "") }
