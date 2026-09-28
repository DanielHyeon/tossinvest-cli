You are an independent ADVERSARIAL senior engineer doing the SIXTH proposal-freeze review of an OpenSpec change in a Go
repository that runs a real-money automated trading engine. You are read-only: do NOT edit, create, or delete any file,
do NOT run git commands that write, do NOT run network calls, do NOT run the engine or any command that could place an
order. Reading files and grep/rg are fine. Do not run Go tests.

Repository root: the current directory. It is an export of commit <TREE_COMMIT> (all code and documents are exactly that
commit). <GO_DRIFT_NOTE — filled at run time from `git diff --stat 3937e341 <TREE_COMMIT> -- <cited Go files>`>

Change under review — `openspec/changes/a094-a-stop-clears-what-blocks-it/`. The documents are the SIXTH draft (6판).
**design.md section "D−4. 6판" is normative and overrides D−3 (5판), which overrides D−2 (4판), which overrides the rest.**
Read first:
- design.md §D−4 (D−4.1 … D−4.9), then §D−3 and §D−2 for what D−4 did not change (D−3.2, D−3.3, D−3.7 carry 6판 banners)
- specs/order-execution/spec.md, specs/exit-policy/spec.md
- tasks.md (§0.5j–0.5l, §3 incl. 3.R2–3.R7c, 3.R5/3.R5a, 3.R4, §4 incl. 4.0/4.0b, 4.N4–4.N4x, 4.T/4.Ta, 4.3e, §6.1a/6.2, §7.2/7.3,
  §8.2, 선후 관계, the safety table)
- proposal.md "6판" box
- analysis/freeze-review/codex-r5-output.md (round 5: REJECT, findings R5-1 … R5-7)
- review.md sections 「5라운드」, 「6판」 and 「Manager 판정 (2026-09-29) — Q6-1 · Q6-2」
- analysis/function-logic/internal-app-engine--exitobserver.record/function-logic-map.md (Safety conclusion corrected in 6판)
Treat every claim, including the disposition tables and the Manager decisions, as a claim to verify, not as truth.

Relevant code: internal/execgw/indoubt.go (matcher :638-650, resolvePlace, :307), internal/execgw/roundtrip.go
(confirmCreatedOrder, roundTripFor, roundTripTimeout), internal/execgw/gateway.go (checkSymbolFree, attemptTargets),
internal/execgw/replay.go (EnqueueAlert caller), internal/journal/outbox.go (EnqueueAlert), internal/journal/resolution.go
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
1. For each round-5 finding R5-1 … R5-7: RESOLVED / PARTIAL / NOT RESOLVED, one line why, pointing at where the 6th draft
   answers it.
2. New findings table: id | severity (P0 = unsafe or unimplementable as written, P1 = must fix before freeze, P2 = should
   record, P3 = editorial) | finding | evidence | suggested fix. Attack in particular:
   - **D−4.2-2 deliberately deviates from Recovery.Run's neighbour convention**: a ledger-write failure of the boot
     ACKED→CONFIRMED transition (and any read failure) is absorbed into a named critical and the row stays ACKED, instead of
     returning ErrRecoveryIncomplete like replay/resolve do (recovery.go:276-289). The stated reason is that the neighbour
     convention starts no observation loop at all (engineready.go:70-75). Attack this: can the absorbed failure leave the
     ledger in a state that a later step (replay, resolve, catch-up, the gateway's per-symbol check, fill detection) misreads?
     Can a half-committed transition exist? Does absorbing here mask a real journal fault that should stop the engine? Is the
     asymmetry with the neighbours safe, and is it confined to exactly this one transition?
   - D−4.2-1 boot confirmation: is it truly byte-exact on the recorded broker order id with no other inference; does it reuse
     the same judgement as confirmCreatedOrder across the package boundary; is every other outcome (mismatch, read failure,
     unparseable, no recorded id, CANCEL/AMEND) left unchanged and alerted? Can it confirm the wrong order?
   - D−4.3 (R5-7 form B): is "a cancelled SELL is cleared only after a terminal fill snapshot" implementable from the journal;
     is the terminal evidence keyed by the proposal intent's order numbers (not by list absence); does "no re-cancel while
     waiting" avoid both a double cancel and a stuck state; what if fill detection never records a terminal snapshot for
     that order — is the stop then frozen silently, or do the existing timer and the D−2.7 counter fire? Is leaving BUY
     cancels on ACK correct?
   - D−4.4 (R5-2): does moving the park-cause judgement to the judgement path actually reach the incident shape (armed
     proposal = the stop itself, evaluation suppressed)? Is it one alert per streak, and does it change nothing else?
   - D−4.5 (R5-3): is the operator command specified completely enough to implement safely (mutating, approval reference,
     audit before the ledger transition, stale rejection, same release judgement, no engine lock, no console)?
   - D−4.6 (R5-4): do all the new criticals really go through EnqueueAlert only; who delivers them, and does undelivered
     still latch entry (a124's executor)? Does anything still call Notify synchronously from the observation loop?
   - D−4.7 (R5-5): is the traced third mechanism correct at this tree; does using the gateway's same function in cleanup
     close it without blocking a legitimate stop; is the 30-second delay alert now guaranteed to fire?
   - D−4.8: is the broker-call recount (one OrderRaw per ACKED PLACE row at boot, ≤3 requests with 401 refresh, within
     roundTripTimeout; zero in the loop and stop path) complete?
   - Spec deltas vs design vs tasks consistency; the MODIFIED order-execution delta still reproduces the canonical
     requirement text verbatim where it claims to.
3. Final line: `VERDICT: PASS` (no P0/P1 left; P2 may be recorded) or `VERDICT: REJECT`, with one sentence why.
