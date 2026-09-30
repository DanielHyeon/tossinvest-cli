You are an independent ADVERSARIAL senior engineer doing the THIRD codex proposal-freeze review of an OpenSpec change in a Go repository
that runs a real-money automated trading engine. You are read-only: do NOT edit, create, or delete any file, do NOT run git
commands that write, do NOT run network calls, do NOT run the engine or any command that could place an order. Reading files
and grep/rg are fine. Do not run Go tests.

Repository root: the current directory. It is an export of commit 9fa0bb90 (all code and documents are exactly that commit).
Every Go file this change cites is byte-identical to its base `d3bd1843` except `cmd/tossctl/engine.go`, which gained 2 lines at
:121-122 (a066 operator command registration, unrelated). The change's engine.go citations (:634-640, :639, :671) were taken from
this later layout and are already correct in this tree (at the base they are 2 lines lower).

Change under review — `openspec/changes/a090-an-unobserved-position-is-counted/` (FOURTH draft, 4판, 12dbff0f). Earlier rounds: Claude adversarial
voice 1 (F1–F16), codex round 1 (N1–N10) and codex round 2 (R2-1…R2-6, analysis/freeze-review/codex-r2-output.md); dispositions are in
review.md 「1라운드」, 「2판」, 「codex 1라운드」, 「3판」, 「codex 2라운드」, 「4판」 (incl. 「계좌 정보 사실 고정」). Read:
- proposal.md, design.md (4판 header, D1–D12 and 「Q — 결정 기록」), specs/exit-policy/spec.md, tasks.md, review.md
- analysis/function-logic/ (ObserveOnce, workingSet, Notifier.AnnounceOperatingMode), analysis/harness/, analysis/code-context/
- cross-change contracts it now depends on: openspec/changes/a092-an-alert-does-not-hold-the-stop/design.md §D0.3h item 4 「K5 · K6 · K7」
  (critical-record single entrance, per-writer remindAfter, pre-latch rule for writers outside it) and
  openspec/changes/a094-a-stop-clears-what-blocks-it/design.md §D−6 (same form); canonical openspec/specs/engine-safety/spec.md
  「늦은 적용은 제때 적용과 같아야 한다」 (around :1455-1475, when the clear epoch is read).
Treat every claim, including disposition tables and Manager decisions, as a claim to verify, not as truth.

Relevant code: internal/app/engine/exitloop.go (ObserveOnce :413-470, workingSet :493-612, openState :653-699, observe
:743-814, checkOutage :817-854, judge :857-893, quoteUsable :1068, alert :1706, delayedSince/noteDelay, file header),
internal/app/engine/tracer.go, internal/app/engine/exitwiring.go, internal/app/engine/gateway.go, cmd/tossctl/engine.go (engineRuntime),
internal/obs/mode.go (AnnounceOperatingMode), internal/obs/log.go, internal/app/engine/interlock.go (AccountRef, MaskedAccount), internal/clock/clock.go (LeaseAnchor, LeaseElapsed), internal/app/engine/a111_flat_exit_observation_test.go,
internal/app/engine/exitloop_test.go, internal/obs/event.go, internal/obs/notifier.go, internal/obs/alert_lease.go,
internal/journal/outbox.go (EnqueueAlert, claimOwed), internal/journal/operating_mode.go (EscalateOperatingMode),
internal/journal/exit_snapshot_integrity.go, internal/execgw/retry.go (EntryGate, ClearEpoch, BlockUnlessClearedSince,
Retrier.Gate), internal/official/market_reads.go (adaptPrices), internal/official/client.go.

Safety invariants that override everything: never weaken or delay stop-loss or emergency flatten; only conservative-direction
changes; operating-mode relaxation needs a human; no secrets/account data in logs/alerts; toggle OFF equals prior behaviour; no
automatic live-order side effects.

Produce (concise, evidence as file:line against this tree):
1. For each codex round-2 finding R2-1 … R2-6: RESOLVED / PARTIAL / NOT RESOLVED, one line why.
2. New findings table: id | severity (P0 unsafe/unimplementable as written, P1 must fix before freeze, P2 should record, P3
   editorial) | finding | evidence | suggested fix. Attack in particular:
   - R2-1: no account reference in the new alerts, mode announcement, keys or log lines; the dedicated `UnobservedLog` (new options
     field, wired in engineRuntime) instead of the observer's whole `Log` — does anything in the 4th draft still let the account
     ref, prices, quantities or raw error strings reach a new line or alert? Is the canary test (R17) sufficient? Is the claim that
     existing observer lines stay unchanged in production true?
   - R2-2: the pending-announcement queue for a committed tightening whose announcement enqueue failed (retry with the same
     transition id, survives streak end, lost on restart as a named residual). Consistent with journal.EscalateOperatingMode's
     return contract (ErrModeAnnouncementFailed with record + changed=true)? Can it double-announce or re-tighten?
   - R2-3: the go/parser exit census pin (tasks 2.15): coordinates = enclosing if-condition text + ordinal; B8 allowed only as its
     original `continue` node; five mutations. Well-defined and does it close the two holes you named last round?
   - The a092 single-entrance dependency (D4/D5): is "implement after a092's entrance lands, else outside-entrance form with the
     pre-latch rule" a complete spec for both orders? Does it conflict with anything a092 D0.3h actually says?
   - Clear-epoch read timing (after the record error returns) vs the canonical engine-safety wording.
   - Anything else still blocking: spec delta vs design vs tasks consistency; every RED implementable and sufficient.
3. Final line: `VERDICT: PASS` (no P0/P1 left; P2 may be recorded) or `VERDICT: REJECT`, with one sentence why.
