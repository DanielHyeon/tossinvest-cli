You are an independent ADVERSARIAL senior engineer doing the SEVENTH proposal-freeze review of an OpenSpec change in a Go
repository that runs a real-money automated trading engine. You are read-only: do NOT edit, create, or delete any file,
do NOT run git commands that write, do NOT run network calls, do NOT run the engine or any command that could place an
order. Reading files and grep/rg are fine. Do not run Go tests.

Repository root: the current directory. It is an export of commit 9fa0bb90 (all code and documents are exactly that
commit). Every Go file the change cites is byte-identical to the change's base `3937e341` except four, all additive and outside this
change: `internal/journal/schema.go` (a066 schema v34/v35 entries, +7 lines at :165-171 — schema.go lines the change cites after :165 are
+7 here), `internal/journal/outbox.go` (a124 `PendingAlertsForDelivery`, +22 lines at :512-533 — lines after :512 are +22; `EnqueueAlert`
at :131 is unchanged), `cmd/tossctl/engine.go` (+2 lines at :121-122 — lines after :121 are +2), and `internal/execgw/retry.go` (a124
`clearEpochs`, `ClearEpoch`, `BlockUnlessClearedSince`, +39 lines from :475 — the 7th/8th draft cite these new symbols at their
positions in this tree).

Change under review — `openspec/changes/a094-a-stop-clears-what-blocks-it/`. The documents are the EIGHTH draft (8판, 7002270f).
**design.md §D−6 (8판) overrides §D−5 (7판), which overrides §D−4 (6판), §D−3, §D−2, and the rest.** Read first:
- design.md §D−6 (D−6.1 … D−6.3), §D−5 (D−5.1 … D−5.7; D−5.3 carries an 8판 banner), then §D−4 for what they did not change
- specs/order-execution/spec.md, specs/exit-policy/spec.md
- tasks.md (§0.5l–0.5o, §3 incl. 3.R7d, §4 incl. 4.N4–4.N4g, §7.2/7.3, §8.2, 선후 관계, safety table)
- proposal.md "8판" and "7판" boxes
- analysis/freeze-review/codex-r6-output.md (round 6: REJECT, R6-1 … R6-7)
- review.md sections 「6라운드」, 「7판」, 「Manager 판정 — Q7-1」, 「8판」
- cross-change contract: openspec/changes/a092-an-alert-does-not-hold-the-stop/design.md §D0.3h item 4 「K5 · K6 · K7」 (critical-record single
  entrance) and canonical openspec/specs/engine-safety/spec.md 「늦은 적용은 제때 적용과 같아야 한다」 (around :1455-1475)
Treat every claim, including the disposition tables and Manager decisions, as a claim to verify, not as truth.

Relevant code: internal/execgw/indoubt.go (matcher :638-650, resolvePlace, :307), internal/execgw/roundtrip.go
(confirmCreatedOrder, roundTripFor, roundTripTimeout), internal/execgw/gateway.go (checkSymbolFree, attemptTargets),
internal/execgw/replay.go (EnqueueAlert caller, parkAlert pre-latch), internal/execgw/symbolgate.go (BlockSymbol), internal/reconcile/mismatch.go, internal/journal/outbox.go (EnqueueAlert), internal/journal/resolution.go
(ResolveConfirmed, OperatorResolve), internal/journal/dispatch.go, internal/journal/fills.go (LiveOrdersForSymbol, terminal
snapshots), internal/journal/recovery.go, internal/brokerstate/derive.go (IsTerminal), internal/filldetect/*.go (cadence, SLO),
internal/reconcile/recovery.go (Recovery.Run), internal/official/client.go (401 refresh loop), internal/app/engine/exitloop.go
(record, submit, clearTheSymbol, release, noteDelay, clearDelay), internal/app/engine/gateway.go, cmd/tossctl/engineready.go,
cmd/tossctl/engine_reconcile.go (precedent for the operator command), internal/exitpolicy/ladder.go, internal/exitpolicy/ratchet.go,
openspec/specs/order-execution/spec.md, openspec/specs/exit-policy/spec.md.

Safety invariants that override everything: no weakening/delaying of stop-loss or emergency flatten (invariant 4) — the
6th draft claims two bounded, alerted yields (D−3.2 park; D−4.3 one fill-detection cycle over a cancelled working sell);
toggle OFF equals prior behaviour; only conservative-direction changes; never cancel or place an order the engine cannot
attribute; never clear an armed exit proposal while its order may be live; operating-mode relaxation needs a human; no
secrets/account data in logs.

Produce (concise, evidence with file:line against this tree):
1. For each round-6 finding R6-1 … R6-7: RESOLVED / PARTIAL / NOT RESOLVED, one line why.
2. New findings table: id | severity (P0 = unsafe or unimplementable as written, P1 = must fix before freeze, P2 = should
   record, P3 = editorial) | finding | evidence | suggested fix. Attack in particular:
   - D−5.1: one shared confirmation predicate strengthened to reject a missing/null/blank response symbol for both immediate
     read-back and boot confirmation. Is the measurement (contract `Order.symbol` required, zero recorded order-detail bodies,
     fixture census) sufficient to justify it? Which existing tests break, and is 4.N4e's plan to fix only fixtures that actually
     reach the read-back correct?
   - D−5.2/D−6.3: episode keys (attempt id, streak start, intent) with remindAfter 0; "episode = identity, remind window =
     re-delivery, orthogonal". Is every key finite and fact-bound? Does anything re-send or silently swallow a new episode?
   - D−6.1: the new critical writers go through a092's single critical-record entrance (remindAfter 0); the direct
     `BlockUnlessClearedSince` of D−5.3 is withdrawn; outside-entrance form must pre-latch per a092. Is the spec satisfiable in
     both implementation orders? Does anything in a092 D0.3h contradict it?
   - D−6.2: the clear-epoch read timing (after the record error returns) vs the canonical engine-safety wording.
   - D−5.3/Q7-1: account-wide (not per-symbol) entry latch on enqueue failure, justified as "a ledger write failure is an
     account-scope fault". Correct scope? Anything that makes it over- or under-block?
   - D−5.4: the unbounded terminal-evidence wait (no time-based release) and its alerts — honest and safe against invariant 4?
   - D−5.5: complete request accounting including token POSTs.
   - Spec deltas vs design vs tasks consistency; the MODIFIED order-execution delta still reproduces the canonical text
     verbatim where it claims to.
3. Final line: `VERDICT: PASS` (no P0/P1 left; P2 may be recorded) or `VERDICT: REJECT`, with one sentence why.
