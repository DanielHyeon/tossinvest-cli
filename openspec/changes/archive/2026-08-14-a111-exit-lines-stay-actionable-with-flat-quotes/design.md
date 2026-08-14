# a111 — Design

## Context

A110 removed the false account-wide reconciliation block and production then adopted five holdings. The
first post-adoption exit cycle exposed a separate lifecycle defect. An adopted exit state opens as `SEED`
with `high_water == entry_price`. When the next valid official quote is unchanged,
`ExitObserver.judgeRatchet` and `judgeLadder` mark the pure snapshot `Changed=false` and return before
`RecordExitJudgementResult`; the state therefore remains `SEED` and the operator surfaces correctly hide
its actionable lines. The same early return also discards later successful flat-price observations, so an
already evaluated snapshot can exceed the 30-second operator freshness bound while fresh official quotes
continue to arrive.

The current judgement transaction is deliberately larger than a freshness write: it selects the monotone
effective snapshot, advances state, may arm one proposal, and appends one semantic `exit_events` row before
commit. Calling it every five seconds would repair freshness but would manufacture an unbounded event
stream. Updating only `last_observed_at` would be worse: that column is cross-checked against
`effective_snapshot_json`, and the immutable snapshot IDs are bound to the observation identity.

The affected high-risk functions and journal transaction must have pre-edit Go AST, Function Logic Map and
Branch Test Map evidence before implementation. The A110 deployment, audited reconciliation release and
managed lifecycles remain authoritative while A111 is developed.

## Goals / Non-Goals

**Goals:**

- Persist one complete canonical `EVALUATED` snapshot on the first valid observation after `SEED`, including
  when the official quote equals the adoption quote.
- Keep a semantically unchanged effective exit line fresh using durable observation evidence without
  creating an exit transition, event, proposal, intent, cancellation or broker order.
- Make console and HTTP positions use the same persisted freshness verdict and the same canonical line.
- Preserve monotone protection, recovery selection, proposal deduplication, safety gates and quote budget.

**Non-Goals:**

- Change the 30-second operator freshness bound or display raw/seed values as actionable lines.
- Change ladder or ratchet mathematics, adoption eligibility, default stop percentage or policy assignment.
- Add a price poll, schema migration, web mutation, LIVE action or automatic reconcile release.
- Backfill historical `SEED` rows without a new valid official observation.

## Decisions

### D1. Classify evaluation outcome separately from domain-state change

Both ratchet and ladder observers use one shared classification after pure snapshot evaluation:

1. `SEED` or no complete effective snapshot: send the complete judgement through the existing
   `RecordExitJudgementResult` transaction even when `Changed=false`.
2. Any state/action change, re-judgement, executable proposal, or semantic line difference: use the existing
   full judgement path unchanged.
3. `EVALUATED`, no re-judgement, no executable proposal and a candidate whose operational fields are exactly
   equal to the current effective line: use the observation-refresh path in D2.

Semantic equality includes policy identity, lifecycle and position generation, entry/stop,
current protection, high-water, rung/level, next target/protection, action, ratio, projected quantity,
orderability, state-only/suppression and cancel-first. It excludes only observation-bound provenance
(`ObservedPrice`, `ObservationID`, input/snapshot/decision IDs) and the derived `Changed` bit. The refresh
still replaces the complete candidate, so a changed in-band observed price is durably visible without being
mislabelled as an exit transition. A quantity change that alters line projection, or a non-scalar operational
change such as pending suppression changing action, ratio, projection or orderability, must therefore take the
full semantic path. The latter is the regression case the old high-water/protection/level predicate missed.

Alternative rejected: treat `SEED` as `Changed=true` in the pure evaluator. `SEED` is journal lifecycle
knowledge, not exit-policy arithmetic; injecting it into the pure snapshot API would make identical inputs
produce context-dependent results.

Alternative rejected: call the full judgement transaction for every flat tick. That appends a semantic
event every loop and makes audit history proportional to wall-clock time rather than decisions.

### D2. Add a narrow transactional refresh of an already evaluated effective snapshot

The journal gains an explicit observation-refresh command. Its request carries the complete recomputed
snapshot, recovery-policy evidence, observation source/time, position ID and lifecycle generation. The
transaction MUST:

- validate the candidate with the same finite-decimal, policy identity, generation and stored-snapshot
  integrity rules as a full judgement;
- require the current state to be `EVALUATED`, managed, not completed, on the same lifecycle generation;
- require no executable proposal, no arm-suppression claim and exact D1 semantic equality with the current
  effective snapshot;
- atomically replace the effective snapshot JSON and every matching flattened identity/line/observation
  column with the new observation-bound tuple;
- append no `exit_events` row, arm no proposal and call no order/cancel adapter.

The refresh uses the candidate's new observation/snapshot/decision IDs rather than lying by attaching a new
timestamp to an old immutable identity. It also performs a temporal compare-and-swap against the current
effective `ObservedAt`: an older candidate is a typed stale/no-write conflict; the same timestamp and same
observation identity is a true no-op that does not churn `updated_at`; and only a strictly newer official
observation may replace the tuple. The same official timestamp with a different identity is ambiguous and
rejected. The zero-`FetchedAt` compatibility source stores `ObservationSource=cycle:<sequence>` rather than
plain `cycle`. Before issuing the cycle sequence, the observer scans the current working set's valid effective
sources, atomically raises its process counter to at least the largest persisted `cycle:N`, then allocates
N+1; restart therefore resumes rather than resets the order. At an equal timestamp, a strictly greater cycle
sequence may replace a cycle tuple, an equal sequence+identity is the same no-op, an equal sequence with a
different identity is a typed ambiguous/no-write conflict, and lower or malformed evidence is rejected.

Equal-time cross-source ordering is total and conservative: `quote_fetched_at` (official non-zero source)
outranks `cycle:N`. Thus a genuine official reading may replace a fallback at the same timestamp; fallback
may not replace official evidence; two different official identities at one timestamp remain ambiguous and
are rejected. If the state
moved after the observer read it, semantic equality or lifecycle validation likewise fails with a typed
conflict. No conflict updates any column, and the next cycle re-reads authoritative state.

This is a state projection refresh, not a semantic decision record. The latest canonical state remains
durable across process restart, while `exit_events` remains the audit trail of transitions and arming
decisions. No schema change is required because the v10 state row already stores the complete flattened and
JSON tuples.

Alternative rejected: update only `last_observed_at`. It would contradict `effective_snapshot_json`, omit
the observation identity and break the existing flattened-vs-JSON integrity checks.

Alternative rejected: keep freshness only in engine runtime memory. The HTTP sidecar, console restart and
journal-only diagnostics would disagree, and a process restart would make a continuously observed line
unknown without any domain reason.

### D3. A first flat observation is a full judgement but never an order by itself

For `SEED`, the observer evaluates the official quote once and invokes the existing record path. An
unchanged t0 quote produces a non-orderable snapshot with zero executable proposal, so the transaction moves
the row to `EVALUATED`, persists current/next lines and appends the single first-evaluation event, but does not
clear orders, arm an intent, increment `Proposed` or call the submitter. A price that really crosses an exit
condition remains the existing normal first-tick exit behavior; A111 does not suppress it.

### D4. A real observation heartbeat replaces a077's running-engine age bypass

`operatorview` remains the common authority for console and HTTP line rendering. A111 changes the premise
that forced a077's console exception: after D1/D2, `last_observed_at` is the latest successful per-position
evaluation, not merely the last state transition. The shared transport-neutral helper therefore applies:

- liveness positively known stopped: stale immediately with `engine_not_running`;
- liveness running, unavailable or unwired: require integrity and apply the existing 30-second persisted-
  observation bound. Age exactly 30 seconds is fresh; only age greater than 30 seconds is stale.

This deliberately supersedes only the a077 running-age workaround because its stated premise is no longer
true. It does not treat the protection numbers as market data; it asks whether this exact position has been
successfully evaluated recently. A valid sibling cannot keep an invalid/missing symbol actionable: that
symbol's own timestamp does not advance and becomes stale after the bound. `SEED`, corrupt, future-dated or
generation-mismatched evidence remains unknown/stale and renders dashes.

The console and HTTP adapters take their response-time authority only after every blocking read that can
delay projection, including journal/policy/runtime/quarantine reads and the single engine-marker filesystem
read. The marker read is bounded by a pre-read sample, but one post-marker sample is authoritative for both
marker liveness and every per-position 30-second comparison. A marker already classified stopped by the
pre-read `enginelock.Read` may only remain stopped; a backward wall-clock step after that read MUST NOT
resurrect it merely because its persisted refresh time now appears younger.

Any legacy comment or caller that assumes `last_observed_at` means “last state transition” is updated. After
A111 it means “last official observation that durably revalidated this exact effective line.”

### D5. Observation time is validated before judgement and failure cannot manufacture freshness

The present loop validates only positive `Last`; A111 adds an explicit source-time contract before any
position judgement. After the one batched price request returns, the observer captures one engine-clock
`validatedAt` for the batch. A non-zero `FetchedAt` is accepted only when it is not after `validatedAt`
(zero future tolerance, because the official adapter stamps it from the same local clock after the response)
and `validatedAt - FetchedAt <= 15s`, matching `execgw.QueryPrice`'s existing staleness bound. Exactly 15
seconds is accepted; older is source-stale. A zero `FetchedAt` retains the existing compatibility fallback:
the successful just-returned batch is stamped with the engine `validatedAt`, source `cycle:<sequence>`, and
the observer sequence recovered from the maximum durable working-set sequence before allocation. D2 uses
that persisted sequence and source precedence to order equal-clock refreshes. It is not copied as a zero
timestamp.

Validation participates in the Retrier attempt, before it records `QueryPrice` success. If no managed symbol
has valid evidence, the attempt returns a typed evidence-invalid error that the retry matrix classifies as
non-retryable: the batch is not called again, the entry-gate freshness latch is not cleared, and the exit
outage clock is not reset. If valid siblings exist, the one batch is a valid query success but only those
symbols are judged; invalid siblings remain unobserved.

Every accepted quote carries two distinct time authorities: its persisted UTC `FetchedAt`, and a
process-local lease anchor captured immediately after the batch returns. The production system-clock anchor
retains Go's monotonic component; injected clocks use their deterministic elapsed-time implementation.
Before each position judgement, and immediately before either the full record path or refresh
transaction—therefore also before `clearTheSymbol` or any other externally visible side effect—the observer
requires both that current wall time is not before `FetchedAt` and that monotonic elapsed time is at most 15
seconds. Exactly 15 seconds is usable; later is not. A backward wall-clock step therefore rejects
future-looking source evidence and cannot extend an already expired local lease. Once an orderable path begins its first
irreversible action while the evidence is valid, the existing bounded clear → durable arm → submit sequence
completes as one decision rather than abandoning a half-cleared/armed proposal. A later portfolio candidate
does not inherit that lease and is skipped if its own use-time check has expired. No extra price read is made.

Missing, non-positive, non-finite, future or source-stale quotes cannot reach full judgement, proposal or
refresh and keep the prior effective tuple unchanged. Journal error or refresh conflict also leaves freshness
unchanged. No failure falls back from a non-zero bad source timestamp to host time, and none weakens
protection.

### D6. Cadence and safety side effects remain bounded

A111 adds no broker read. The existing one batched `Prices(ctx, symbols)` call for the whole managed symbol
set per cycle and its fill-detection priority remain unchanged. A flat refresh performs one bounded local
journal transaction. It does not call
`clearTheSymbol`, issuer, submitter, floor, retrier, outbox or LIVE mutation adapters. Pending proposal and
re-judgement paths remain on the full judgement branch.

### D7. RED coverage follows both policies and the production lifecycle

RED tests first prove, for both ratchet and common ladder:

- `SEED` plus an unchanged first valid quote becomes `EVALUATED`, persists canonical current/next lines,
  appends exactly one semantic event and produces zero proposal/order/cancel side effects;
- a second and later identical fresh observation advances persisted observation provenance/freshness without
  adding an event or side effect;
- restart reads the refreshed tuple as fresh, and console/API projections agree;
- invalid/stale observations do not refresh; when the engine is not positively stopped, an exactly
  30-second-old tuple stays fresh and only an older tuple becomes stale on both surfaces;
- an all-invalid HTTP-success batch neither refreshes exit evidence nor records `QueryPrice` success, while a
  valid sibling is the only symbol judged;
- a slow early position can push a later quote past its 15-second use deadline, in which case the later
  position produces no record, refresh, clear, proposal or order;
- backward wall-clock movement cannot extend that deadline: elapsed monotonic time beyond 15 seconds, or a
  wall time now before official fetched-at, skips the later position with no side effect;
- out-of-order refresh (newer commit followed by older) cannot regress the tuple; exact duplicate delivery
  is a no-op including `updated_at`;
- concurrent state movement rejects the refresh and cannot overwrite stronger protection;
- a real state/action change still takes the existing record/arm/submit ordering.

Mutation tests remove the SEED exception, restore the unchanged early return, turn refresh into a full event,
relax semantic equality, update only the timestamp, allow invalid quotes to refresh and make console/API use
different clocks. Each mutation must be killed by a named test before final verification.

## Risks / Trade-offs

- **[A refresh overwrites a concurrent stronger line]** → compare lifecycle and all operational fields inside
  the journal transaction; conflict is a no-write result.
- **[Observation rows create write load every five seconds]** → one small state-row transaction per flat
  managed position is accepted to satisfy the 30-second durable freshness contract; no extra network call or
  event row is created. Benchmark the focused loop and retain rate priority.
- **[Audit history loses evidence that the engine looked]** → the state row carries the latest immutable
  observation identity and timestamp; `exit_events` intentionally records decisions, not heartbeats.
- **[Saved-monotone recovery semantics are weakened]** → refresh is permitted only when the recomputed
  operational line exactly equals the current effective line; otherwise the full selector path runs.
- **[Wall-clock rollback extends a stale quote lease]** → persist UTC provenance separately from a raw
  monotonic process anchor; require both wall provenance validity and bounded monotonic elapsed time at use.
- **[A blocking adapter read crosses a freshness boundary]** → take one post-marker response clock after
  all other blocking reads, never upgrade a marker already read stopped, and project every row from that
  single authority.
- **[First observation unexpectedly submits]** → unchanged-t0 RED fixtures inject counting cancel/issuer/
  submitter adapters and require zero calls; genuine threshold crossings preserve existing behavior.
- **[Console and API diverge]** → both consume the same operatorview result. Running/unavailable/unwired
  boundary tests require fresh at 29.999s and exactly 30s, stale only over 30s; stopped is immediately stale.
- **[Frozen/coarse clock repeats a fallback timestamp]** → persist and compare the monotone `cycle:<sequence>`
  source, recover its maximum from the working set after restart, and define official-over-cycle precedence;
  official same-time/different identity remains rejected as ambiguous.

## Migration Plan

1. Land code and tests with no database or config migration; existing schema remains 31.
2. Deploy the exact-digest A111 image one service at a time under the existing dormant preflight and rollback
   contract. The A110 image remains the rollback target and its released reconcile audit row is retained.
3. Restart the engine and verify the first valid cycle converts any remaining `SEED` managed rows to
   `EVALUATED`; then hold a flat quote beyond 30 seconds and prove journal, console and HTTP lines stay fresh.
4. Verify no new order/proposal/cancel event was created solely by flat refresh.
5. Rollback may restore A110 without schema downgrade. Refreshed v10 state tuples remain readable; A110 may
   again let them age stale, but must not corrupt or reinterpret them.

## Open Questions

None for implementation. Exact write-rate measurements are a verification result, not a contract choice.
