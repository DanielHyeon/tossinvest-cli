You are an independent ADVERSARIAL senior engineer doing the SECOND codex proposal-freeze review of an OpenSpec change in a Go repository
that runs a real-money automated trading engine. You are read-only: do NOT edit, create, or delete any file, do NOT run git
commands that write, do NOT run network calls, do NOT run the engine or any command that could place an order. Reading files
and grep/rg are fine. Do not run Go tests.

Repository root: the current directory. It is an export of commit f478ddc6 (all code and documents are exactly that commit).
Every Go file this change cites is byte-identical to its base `d3bd1843` except `cmd/tossctl/engine.go`, which gained 2 lines at
:121-122 (a066 operator command registration, unrelated). The change's engine.go citations (:634-640, :639, :671) were taken from
this later layout and are already correct in this tree (at the base they are 2 lines lower).

Change under review — `openspec/changes/a090-an-unobserved-position-is-counted/` (THIRD draft, 3판). Earlier rounds: Claude adversarial
voice 1 (F1–F16) and codex round 1 (N1–N10, analysis/freeze-review/codex-r1-output.md); their dispositions are in review.md 「1라운드」,
「2판」, 「codex 1라운드」, 「3판」. Read:
- proposal.md, design.md (D1–D11 and 「Q — 결정 기록」), specs/exit-policy/spec.md, tasks.md, review.md
- analysis/function-logic/ (ObserveOnce, workingSet, Notifier.AnnounceOperatingMode bundles), analysis/harness/, analysis/code-context/
- background: the sibling change openspec/changes/a094-a-stop-clears-what-blocks-it/design.md §D−4.6 and §D−5.2–5.3 (same alert form),
  canonical openspec/specs/exit-policy/spec.md 「관측 경로와 fail-safe」, openspec/changes/archive/2026-09-28-a089-an-unserved-stop-is-counted/review.md.
Treat every claim, including disposition tables and Manager decisions, as a claim to verify, not as truth.

Relevant code: internal/app/engine/exitloop.go (ObserveOnce :413-470, workingSet :493-612, openState :653-699, observe
:743-814, checkOutage :817-854, judge :857-893, quoteUsable :1068, alert :1706, delayedSince/noteDelay, file header),
internal/app/engine/tracer.go, internal/app/engine/exitwiring.go, internal/app/engine/gateway.go, cmd/tossctl/engine.go (engineRuntime),
internal/obs/mode.go (AnnounceOperatingMode), internal/obs/log.go, internal/clock/clock.go (LeaseAnchor, LeaseElapsed), internal/app/engine/a111_flat_exit_observation_test.go,
internal/app/engine/exitloop_test.go, internal/obs/event.go, internal/obs/notifier.go, internal/obs/alert_lease.go,
internal/journal/outbox.go (EnqueueAlert, claimOwed), internal/journal/operating_mode.go (EscalateOperatingMode),
internal/journal/exit_snapshot_integrity.go, internal/execgw/retry.go (EntryGate, ClearEpoch, BlockUnlessClearedSince,
Retrier.Gate), internal/official/market_reads.go (adaptPrices), internal/official/client.go.

Safety invariants that override everything: never weaken or delay stop-loss or emergency flatten; only conservative-direction
changes; operating-mode relaxation needs a human; no secrets/account data in logs/alerts; toggle OFF equals prior behaviour; no
automatic live-order side effects.

Produce (concise, evidence as file:line against this tree):
1. For each codex round-1 finding N1–N10: RESOLVED / PARTIAL / NOT RESOLVED, one line why.
2. New findings table: id | severity (P0 unsafe/unimplementable as written, P1 must fix before freeze, P2 should record, P3
   editorial) | finding | evidence | suggested fix. Attack in particular:
   - R15 structural pin (tasks 2.15): marking right after workingSet B6 (`:520`), unmarking as the FIRST statement of B10 (`:533`),
     no new `continue`/`return` between them except B8's `:531` allowed by name. Is the pin well defined against the AST, does it
     actually fail for the three named mutations, and does allowing `:531` by name leave a hole (e.g. a new drop path inside B8's
     block, or before `:520`)?
   - N2: extracting `OperatingModeEvent` from AnnounceOperatingMode with no behaviour change; the new enqueue-only announcer whose
     key includes the transition id. Is the extraction really behaviour-preserving (key, title, body, fields)? Does putting the
     transition id in the key break any consumer that dedupes or looks up mode alerts by key? Is the mode commit still synchronous
     and is every other announcer path unchanged?
   - N4: elapsed measured with clock.LeaseAnchor/LeaseElapsed while display uses wall UTC; episode key uses a per-streak id from
     opts.NewID. Do these helpers behave as claimed for both the system clock and injected test clocks? Any path that still uses
     wall time for the threshold?
   - D10 failure transitions (enqueued / enqueue_failed / tightened / tighten_failed; ErrModeAnnouncementFailed): complete and
     consistent with journal.EscalateOperatingMode's actual return contract? Can an operator relaxation race re-latch or be
     silently overridden? Can a failure loop repeat an entry latch after an operator clear (principle E)?
   - N3: the one-line logger wiring and the new normal event type; is `logger` in scope at that call site; does starting to emit
     the observer's existing o.log lines in production have any side effect (volume, severity labels, secrets)?
   - D11 conditional guarantee and the named holes (persistent B2, restart loops): honest and complete?
   - Spec delta vs design vs tasks consistency; every RED in tasks §2 implementable and sufficient.
3. Final line: `VERDICT: PASS` (no P0/P1 left; P2 may be recorded) or `VERDICT: REJECT`, with one sentence why.
