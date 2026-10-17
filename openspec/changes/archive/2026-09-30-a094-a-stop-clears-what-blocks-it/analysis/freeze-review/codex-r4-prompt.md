You are an independent ADVERSARIAL senior engineer doing the FOURTH proposal-freeze review of an OpenSpec change in a Go
repository that runs a real-money automated trading engine. You are read-only: do NOT edit, create, or delete any file,
do NOT run git commands that write, do NOT run network calls, do NOT run the engine or any command that could place an
order. Reading files and grep/rg are fine. Do not run Go tests.

Repository root: the current directory. It is an export of commit d3aa309a09ddcf8ec8fabdc3a0ceed532b261974 (all code and
documents are exactly that commit).

Change under review — `openspec/changes/a094-a-stop-clears-what-blocks-it/`. The documents are the FOURTH draft (4판).
**design.md section "D−2. 4판" is normative and overrides the rest of design.md** (the rest is the third draft, kept as
history; D2 carries a banner marking its broker-widening parts as user-decision-pending). Read first:
- design.md §D−2 (all subsections D−2.1 … D−2.11, including the Manager decisions recorded in D−2.11)
- specs/order-execution/spec.md, specs/exit-policy/spec.md
- tasks.md (§0.5d–0.5h, §2.5e, §3 incl. 3.X, §4 incl. 4.3a–4.3f, §4bis incl. 4b.2a–4b.8, §7, the safety table)
- proposal.md "4판" box
- analysis/freeze-review/codex-r3-output.md (round 3: REJECT, findings F1–F10)
- review.md sections 「3라운드」 and 「4판」 (the gstack document review of the 4th draft: P0 1, P1 10, P2 8, and how each
  was handled) and 「Manager 판정」 with the read-only ledger re-measurement 0.5h
- analysis/third-round-errata.md, analysis/function-logic/*/ (AST bundles; branch IDs refer to these ast.json files)
Treat every claim, including the dispositions tables, as a claim to verify, not as truth.

Relevant code: internal/journal/durability.go, internal/journal/dispatch.go, internal/journal/lifecycle.go,
internal/journal/resolution.go (ResolveFailed, OperatorResolve), internal/journal/replay.go, internal/journal/schema.go,
internal/journal/execution_contract.go, internal/journal/apply_hook.go (armExitProposalTx, ResolveExitProposal),
internal/journal/fills.go (LiveOrdersForSymbol), internal/journal/recovery.go, internal/journal/strategy_dispatch_runtime.go,
internal/execgw/classify.go, internal/execgw/gateway.go, internal/execgw/replay.go, internal/execgw/roundtrip.go,
internal/official/errors.go, internal/app/engine/exitloop.go (record, submit, clearTheSymbol, noteDelay, floatOf),
internal/app/engine/gateway.go (restoreAlertEntryLatch), internal/app/engine/runtime.go, internal/obs/notifier.go,
internal/obs/event.go, internal/exitpolicy/ladder.go, internal/exitpolicy/ratchet.go, internal/reconcile/recovery.go,
cmd/tossctl/engine.go (engineRecoverySequence, recoverThenReady), openspec/specs/order-execution/spec.md,
openspec/specs/exit-policy/spec.md, openspec/changes/a089-an-unserved-stop-is-counted/specs/engine-safety/spec.md,
openspec/changes/a124-a-deliverer-that-keeps-failing-blocks-entry/design.md §D2.

Safety invariants that override everything: no weakening/delaying of stop-loss or emergency flatten (invariant 4);
toggle OFF equals prior behaviour; only conservative-direction changes to stop/take-profit/sizing; never cancel or place an
order the engine cannot attribute; never clear an armed exit proposal while its order may be live; operating-mode relaxation
needs a human; no secrets/account data in logs.

Produce (concise, evidence with file:line against this tree):
1. For each round-3 finding F1–F10: RESOLVED / PARTIAL / NOT RESOLVED, one line why, pointing at where the 4th draft
   answers it.
2. New findings table: id | severity (P0 = unsafe or unimplementable as written, P1 = must fix before freeze, P2 = should
   record, P3 = editorial) | finding | evidence | suggested fix. Attack in particular:
   - The release rule: only FAILED_CONFIRMED/NOT_DISPATCHED (never park), intent match, all-attempts / zero-attempt boot
     catch-up, the narrowed in-session `submit` release (State=="" with a recorded attempt keeps the proposal armed), and the
     Q4-1 decision (operator tool releases immediately through the same single judgement function). Is there ANY remaining path
     that clears `pending_action` while an order may be live? Is there any path that leaves a stop frozen that the draft does
     not name?
   - The retro-reclassification's seven conditions (transition provenance, `official: API error <n>: ` marker exactly once,
     code agreement, no replay). Can a readback-after-ack, replayed, strategy, or wrapped-error attempt still qualify? Given
     the 0.5h measurement (both incident attempts already UNRESOLVED_IN_DOUBT), is the reclassification still worth its risk?
   - Q4-4 (keep the `clearTheSymbol` withPending release; stop goes out over an invisible IN_DOUBT/parked take-profit) — is the
     trade-off honestly stated against invariant 4, and is anything claimed about broker oversell rejection?
   - Q4-5/Q4-6/Q4-7: boot-row failures → critical per position; no re-proposal interval and the named "429 → park" risk with
     the implementation-lot stop condition (is the claim that the ledger reason cannot distinguish 429 correct?); N =
     obs.DefaultCriticalAttempts borrowed from a124 — is the borrowing justified or a disguised magic number?
   - F8: the new trigger's distinct alert key, `delayAlerted` untouched, exclusion limited to RECORDED/DISPATCH_STARTED/ACKED
     cancels — implementable from the journal, and does it avoid weakening the existing 30-second alert?
   - R2 reduced to `LiveOrdersForSymbol`: any leftover requirement or task that still needs a broker read outside §3.X?
   - Spec deltas vs design vs tasks consistency; MODIFIED order-execution delta still reproduces the canonical requirement
     text verbatim; the a089 dependency (D−2.8).
3. Final line: `VERDICT: PASS` (no P0/P1 left; P2 may be recorded) or `VERDICT: REJECT`, with one sentence why.
