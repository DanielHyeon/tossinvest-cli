# a111 proposal-freeze review

## Frozen evidence

- Base commit: `3355df0fe9c82c3bb8c522e2d79abf107dd5f2c3`.
- Deployed A110 image: `sha256:a60c591c2ca36d1d2932632d0a34fa4a51cbc4624a7c2883deae5acc22626af6`.
- Runtime receipt: both services healthy on the exact A110 digest; journal schema 31 and integrity OK;
  active reconcile 0 after the audited `OPERATOR` release.
- Incident result: five holdings adopted; three stayed `SEED/not_evaluated_yet`; evaluated flat rows could
  age stale on the HTTP projection although official price batches continued.
- Pre-edit evidence: `analysis/function-logic/` contains CodeGraph hard evidence plus AST, Function Logic Map,
  Branch Test Map and risk reports for both judge functions, `record`, both journal judgement entry points,
  freshness and console/HTTP consumers. `check_analysis.py` passes.

## Independent adversarial rounds

The reviewer was read-only and separate from every implementation/test owner.

1. **P1 temporal regression** — an older flat refresh could overwrite a newer tuple. Closed with durable
   temporal CAS, exact replay no-op and ambiguous equal-time rejection.
2. **P1 quote-time fiction** — the current loop did not actually reject future/source-stale timestamps.
   Closed with explicit post-batch engine-clock authority, future tolerance 0, exact 15-second source bound
   and zero timestamp fallback contract.
3. **P1 Retrier false success** — HTTP success containing only invalid quote evidence could clear
   `QueryPrice` staleness before validation. Closed by making validation part of Retrier success accounting
   with typed non-retryable evidence failure and no second broker request.
4. **P1 use-time expiry** — a valid batch quote could expire while earlier positions were processed. Closed
   with per-position checks before judgement and record/refresh/first side effect, while preserving completion
   after a valid orderable path begins its first irreversible action.
5. **P1 invalid sibling leakage** — a running engine's a077 age bypass could keep one never-observed symbol
   actionable forever because another symbol remained valid. Closed by making the new persisted heartbeat
   authoritative per position: stopped is immediately stale; running/unavailable/unwired all apply integrity
   plus the exact 30-second bound.
6. **P1 fallback ordering** — process-local `cycle` sequence reset on restart and lacked cross-source total
   order. Closed with `cycle:N` persistence, working-set maximum recovery, official-over-cycle precedence,
   and lower/malformed conflict rules.
7. **P1 equal-cycle race** — two observers could issue the same `cycle:N` at the same timestamp with different
   identities. Closed with same N+same identity no-op and same N+different identity typed no-write conflict,
   plus named journal race/replay RED and mutation tasks.

P2 corrections also fixed the broker cadence language to one batched `Prices` call per cycle and froze the
freshness boundary: exactly 30 seconds is fresh; only greater age is stale.

## Freeze verdict

**ACCEPT — P0=0, P1=0, P2=0.**

No production or test implementation began before this verdict. Subsequent implementation review findings
are appended below; this section remains the immutable proposal-freeze receipt.

## T1 RED / A1 adversarial receipt

T1 added only `internal/app/engine/a111_flat_exit_observation_test.go`. Frozen A110 produces the intended
A111 failures, while breach durable-before-submit, pending dedupe, re-judgement, one-batch/fill priority and
malformed-transport preservation controls remain GREEN.

A1 initially rejected the suite with eight P1 gaps: ladder-only SEED special casing, non-official first
evidence, uncounted extra broker reads, same-row rather than working-set restart recovery, nonorderable-only
use-time coverage, missing outage reset assertion, incomplete re-judgement side-effect assertions and a
projection test that did not pin recovery/proposal safety. T1 closed all eight in its owned file. Final A1
verdict: **ACCEPT — P0=0, P1=0**. One non-blocking P2 remains: the test-only runtime-stack lease boundary
clock is refactor-sensitive; its focused comment and scope make that acceptable.

A final arithmetic audit rejected the first task 1.4 fixture even though it was GREEN: changing one share to
three at a 40% partial made the candidate orderable, and the legacy scalar already treats every orderable
candidate as changed. T1 replaced it with an evaluator-generated pending-partial transition. The second
same-price evaluation keeps high-water, protection and level equal (`Changed=false`) but changes action,
ratio, projection, orderability and suppression. Frozen `3355df0f` is RED; A111 is GREEN, records exactly one
semantic transition, preserves the pending intent and adds no issuer/place/cancel call. A1's final re-review
remains **ACCEPT — P0=0, P1=0**. Design, delta spec and task 1.4 now name this real non-scalar boundary while
retaining whole-share projection as a preservation rule.

## T2 RED / A2 adversarial receipt

T2 added only four A111 test files for journal, operatorview, console and HTTP API. Frozen A110 fails at the
planned refresh/freshness seams; clean-baseline snapshot integrity, rollback, quarantine, event and adapter
controls remain GREEN.

A2's first pass found four P1 classes: missing RELEASED/pending authority cases, candidates that were invalid
before semantic equality could be exercised, a frozen clock that hid `updated_at` churn, and helper-only tests
that did not traverse actual routes. T2 added real lifecycle/pending tuples, evaluator/recovery-valid donors
for 18 operational fields, advanced-clock no-write hooks and actual `/api/v1/positions`, `/positions` and
`/position-management` routes.

The second pass found two more P1s: HTTP's 30-second projection cache could hide a stopped marker, and some
CAS losers still lacked physical no-write proof. T2 added the same-reader running→stopped cache test with no
extra broker read and all actionable fields hidden, then applied +1s clock, write-hook 0, complete tuple,
`updated_at` and event invariance to every loser/replay path. Final A2 verdict:
**ACCEPT — P0=0, P1=0, P2=0**.

## T3 GREEN / A3 adversarial receipt

T3 implemented the mapped observer classifier, atomic journal refresh/CAS, durable cycle recovery, quote-time
and use-time validation, non-retryable all-invalid accounting, shared operator freshness and actual console/API
projection. A3 independently traced every production branch to the T1/T2 RED corpus and attacked lifecycle,
pending/orderable guards, transaction rollback, official/cycle ordering, restart/race behavior, irreversible
order sequencing, invalid siblings and cache/liveness transitions.

A3 found one P1: the stopped HTTP overlay could mutate a shallow copy of the cached positions slice and keep a
subsequent running response stale inside the TTL. T2 added a real-router `running → stopped → running` RED with
one broker read. T3 cloned the value-typed item slice before applying the per-request overlay; the RED passed
ten repetitions and its race form passed five. A3 initially questioned a nested pointer alias, then verified
that `Position.ExitLine` is a value and retracted that concern. It also confirmed that non-zero `FetchedAt`
always becomes `quote_fetched_at` and only zero-time evidence uses `cycle:N`. Final A3 verdict:
**ACCEPT — P0=0, P1=0**.

## Mutation and compatibility receipt

The full engine suite exposed three legacy tests whose official adapter stamped wall time while their observer
used a frozen fake clock. Production correctly rejected that future evidence. The test owner preserved the real
official HTTP value path but restamped returned quotes at the explicit harness-clock seam, and aligned one direct
concurrency quote to the same fake clock. The three regressions, full engine suite and A111 source-time boundaries
passed; A1 independently confirmed frozen A110's future/source-stale REDs were not weakened:
**ACCEPT — P0=0, P1=0, P2=0**.

An isolated mutation run at `/tmp/a111-mutation.zhoGnU/repo` killed M1–M19: SEED/full classification, flat
early return, eventful refresh, orderable eligibility, scalar-only equality, partial tuple/identity writes,
lifecycle and temporal CAS, stronger-state overwrite, invalid/all-invalid evidence, use deadlines, per-position
freshness, durable cycle recovery, console clock divergence and HTTP stopped/cache-copy boundaries. M4 first
survived because candidate validation hid the explicit orderable guard. T2 added an exact persisted orderable
replay RED; T3 moved the existing conflict guard before generic judgement validation; the deletion then failed
20/20 and the independent mutation owner verified the closure. One `next_target`-column omission mutant is
measurement-equivalent because refresh requires that operational value to remain identical; JSON and every
observation-bound tuple omission were killed. Isolated/main hashes and main status were identical after restore.

Focused A111 race passed for engine, journal, operatorview, console and HTTP; repetition/write-count checks prove
one price batch per cycle, no flat event growth and zero flat issuer/place/cancel calls. Full engine,
operatorview, console and HTTP packages passed. A full journal run remains deferred to the final 30-minute gate:
on this filesystem its pre-existing SQLite WAL fsync tests exceeded ten minutes, while all A111 journal focused,
repeat and race suites passed.

## gstack fix-first receipt

The final gstack pass reviewed the Go backend, journal transaction, operator projections, console templates,
HTTP adapter/cache, tests and generated evidence. It initially found no security or data-integrity P0, but did
find one API safety P1 and five P2 quality gaps. Every behavioral finding followed the same separate-owner path:
T2 added an executable RED, T3 made the minimal production change, and an independent reviewer re-ran the
boundary before closure.

1. **P2 duplicated quote lifetime** — the 15-second `QueryPrice` bound existed separately in retry and observer
   code. Closed by exporting `execgw.QueryPriceEvidenceDuration` and using it for the entry-gate staleness,
   source-age validation and use lease. A structural RED rejects negated, unconditional and widened fallback
   guards.
2. **P2 eager cycle recovery** — every official-quote cycle scanned the journal for the maximum fallback
   sequence. Closed by recovering `MaxExitObservationCycle` only after the one successful batch contains valid,
   managed zero-`FetchedAt` evidence. Official-only cycles perform no scan and journal failure cannot repeat the
   broker read.
3. **P2 journal test authority** — direct invalid refresh requests and mixed corrupt/out-of-scope cycle rows were
   under-specified. Added physical no-write assertions for invalid identity/source/time/provenance and a mixed
   working-set maximum test; no production defect was found.
4. **P2 console hierarchy and journal truth** — `/position-management` rendered safety values as small provenance
   text and collapsed an unavailable exit journal into `no_saved_evaluation`. Closed with a 1rem safety line,
   preserved typed journal state/detail, explicit fail-closed warning and dashes on unreadable evidence.
5. **P1 HTTP cached projection time** — the 30-second positions cache initially cached liveness/freshness and then,
   after the first correction, still captured response time before blocking journal/policy/runtime reads. Actual-
   route REDs cover cache-hit age transitions, stopped→running recovery, a slow holdings miss, and a slow runtime
   projection read. The cache now stores only the immutable broker snapshot; every request re-reads local truth
   and captures one response-time clock after all blocking reads for marker liveness and every exit line.
6. **P1 console projection time** — `/position-management` captured age/liveness before the quarantine RPC, which
   may block for five seconds. A two-case actual-route RED independently crosses the 30-second observation bound
   and the engine-marker bound during that RPC. The handler now captures one clock after journal/name/quarantine
   reads and shares it across every row.

The final independent closure re-ran the new HTTP and console route tests, focused race suites, shared duration
and lazy-recovery guards, and inspected the response-time authority after every blocking read.

**Final gstack disposition: ACCEPT — P0=0, P1=0, P2=0.**

Post-fix AST, Function Logic Map, Branch Test Map, risk reports and CodeGraph hard-evidence flow were regenerated
again after both response-time fixes. `check_analysis.py` reports no stale AST, unmapped branch or P1.

## Fresh-context clock and surface closure

Two final Codex review subprocesses (adversarial and structured) reached their five-minute read-only timeout
without producing a finding; they are recorded as inconclusive rather than PASS. A separate fresh-context SOL
review then found one P1 and one P2 time-authority class, and a second pass found the same class in the shared
holdings decorator. Each finding was handled RED-first by a test owner, fixed by a different Terra production
owner, and re-reviewed without edits.

1. **P1 wall rollback extended the 15-second quote lease.** The original absolute `ValidUntil` comparison
   could make expired evidence usable again after wall-clock rollback. The observer now persists only UTC
   fetched-at provenance while keeping a process-local lease anchor. `clock.System()` captures that anchor from
   raw `time.Now()` so Go's monotonic component survives; injected clocks retain deterministic `Now/Since`.
   Use requires both wall time not before official fetched-at and monotonic elapsed `<= 15s`, before `Judged`,
   record, refresh or the first externally visible side effect. The rollback RED, exact-boundary preservation,
   system monotonic-metadata test and AST helper-use guard are GREEN.
2. **P2/P1 marker-read response time.** HTTP, `/position-management`, `/positions` and dashboard could cross a
   snapshot or engine-marker boundary during a blocking cache/journal/policy/runtime/quarantine/marker read and
   still project from an older clock. Each actual-route RED now crosses both boundaries without another broker
   read. The adapters finish all blocking reads, take one pre-marker sample, read the marker exactly once, then
   share one post-marker response clock across every line. Re-evaluation is downgrade-only: a marker already
   classified stopped cannot be resurrected if the post-read wall clock moves backward.
3. **Shared holdings caller gap.** The first console fix covered `/position-management`, but the final adversary
   showed that `/positions` and dashboard still passed a pre-policy-read `asOf` into their common decorator.
   The common helper now uses that input only for cache eligibility, samples response authority after policy and
   settings reads, and shares the post-marker time for reconcile blocks and exit lines. Both routes pass the
   observation-age, marker-age and rollback REDs with one Runtime/List pair and zero extra broker reads.

The final fresh-context adversarial verdict is **ACCEPT — P0=0, P1=0, P2=0**. It independently confirmed the
monotonic lease, exact inclusive boundaries, journal CAS/event/proposal/order safety, HTTP cache behavior and all
four operator surfaces. Final production-freeze maps include the clock helpers and all post-marker branches;
the only remaining completion authority is the repository gate and Manager receipt below.

## Repository gate fix-first receipt

The first complete repository test pass inside `make gate` found one structural authority violation that the
focused A111 suites did not expose: `exit_observation_refresh.go` directly named the guarded
`exit_states.pending_action` column. The behavioral refresh was read-only, but the package contract requires
all four fill-time guarded columns to be named only at the atomic apply boundary so a second writer cannot be
introduced outside review.

The existing `TestGuardedExitColumnsAreWrittenOnlyByTheApplyHook` was retained unchanged as the RED. T3 moved
the single status/pending SELECT behind an unexported `exitObservationRefreshGuardTx` in `apply_hook.go`.
That helper returns only snapshot status and a derived pending boolean, exposes neither the transaction nor raw
guarded state, and performs no write. `RefreshExitObservation` still holds the same `BEGIN IMMEDIATE` across
state scan, pending/lifecycle guard, temporal CAS, non-guarded observation tuple update and commit. It names no
guarded column, appends no event and arms no proposal.

The structural RED passed 20 repetitions; A111 released/pending/lifecycle/stronger-state/CAS/cycle/race tests
and targeted race passed. A separate A3 re-review inspected transaction reachability and the no-write boundary:
**ACCEPT — P0=0, P1=0**. The initial failing gate remains part of this receipt; only a subsequent full gate may
authorize completion.
