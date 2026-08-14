# A111 CodeGraph hard evidence

- Frozen worktree HEAD: `3355df0fe9c82c3bb8c522e2d79abf107dd5f2c3`
- Fingerprint refresh: `make sdd-sync` completed before collection.
- Commands: `codegraph context|callers|callees|impact <symbol>`.
- Note: unqualified `record` is ambiguous in CodeGraph. The definition and AST below pin the intended method to `internal/app/engine/exitloop.go:1077`; its direct callers are the two explicit calls in `judgeRatchet` and `judgeLadder`.

| Symbol | Definition | Direct callers | Principal callees | Impact evidence |
|---|---|---|---|---|
| `ExitObserver.judgeRatchet` | `internal/app/engine/exitloop.go:861` | `judge` | policy identity, snapshot context, `EvaluateRatchetSnapshot`, `ChangedFromState`, recovery policy, `record` | 5 symbols: method → `judge` → `ObserveOnce`/observer file |
| `ExitObserver.judgeLadder` | `internal/app/engine/exitloop.go:916` | `judge` | ladder selection/identity checks, snapshot context, `EvaluateLadderSnapshot`, `ChangedFromState`, recovery policy, `record` | 5 symbols: method → `judge` → `ObserveOnce`/observer file |
| `ExitObserver.record` | `internal/app/engine/exitloop.go:1077` | both judge methods | proposal projection, symbol clear, `RecordExitJudgementResult`, quarantine announce, submit | direct mutation fan-out to journal then submission; AST pins 14 branches/19 calls/5 returns |
| `Journal.RecordExitJudgementResult` | `internal/journal/exit_state.go:385` | observer `record`, compatibility wrapper, recovery test | `recordExitJudgementTx` | 47 affected symbols, including engine and journal persistence/crash/identity suites |
| `Journal.recordExitJudgementTx` | `internal/journal/exit_state.go:392` | public result wrapper | validation, recovery selection, quarantine CAS, snapshot encoder, state update, proposal arm, event append, commit | 6 affected symbols; single transaction has 51 AST branches/69 calls/31 returns |
| `Journal.RefreshExitObservation` | `internal/journal/exit_observation_refresh.go:39` | observer `refreshObservation` | validation, `scanExitProgress`, read-only refresh guard, lifecycle/temporal CAS, complete observation tuple update, commit | current source + regenerated AST pin 26 branches/51 calls/24 returns; no event/proposal/order path |
| `exitObservationRefreshGuardTx` | `internal/journal/apply_hook.go:682` | `RefreshExitObservation` only | one transactional SELECT returning snapshot status and a derived pending boolean | new current AST pins 1 branch/3 calls/2 returns and no mutation; graph sync intentionally deferred until Manager docs freeze |
| `armExitProposalTx` | `internal/journal/apply_hook.go:655` | judgement transaction | pending SELECT, guarded pending-tuple UPDATE | unchanged frozen-base AST pins 4 branches; placement proves the adjacent refresh helper does not enlarge writer authority |
| `ExitSnapshotView.WithFreshness` | `internal/journal/exit_snapshot.go:202` | console `exitFreshness`, HTTP `applyStoredPosition`, journal test | RFC3339Nano parse and age/future checks | 13 symbols across journal, console and HTTP API |
| `ExitSnapshotView.WithIntegrity` | `internal/journal/exit_snapshot_integrity.go:19` | console `exitFreshness` | parse and future checks only | 5 symbols through console positions projection |
| console `exitFreshness` | `internal/console/protection_liveness.go:71` | `attachPositionExitLines` | `WithFreshness` when unwired, fail-closed stopped marker, `WithIntegrity` when running | 5 symbols through positions decoration |
| `Console.handlePositions` | `internal/console/portfolio_pages.go:70` | console router registration (CodeGraph has no resolved direct caller) | `positions`, chrome/explain helpers, `decoratePositionRows`, render | `/positions` read route; current AST pins the branch-free shared-decoration call |
| `Console.accountPanelFrom` | `internal/console/overview.go:653` | `overview` | `holdings.peek`, `joinPositions`, `decoratePositionRows`, market aggregation | dashboard account panel; CodeGraph resolves one caller |
| `Console.decoratePositionRows` | `internal/console/portfolio_pages.go:93` | CodeGraph resolves `handlePositions`; current source/AST additionally pins `accountPanelFrom` | policy cache, settings, `readProtectionMarker`, `protectionLivenessAt`, `ProjectManagement`, `attachPositionExitLines` | 3 CodeGraph symbols plus the source-proven dashboard seam; 12 AST branches/16 calls |
| `attachPositionExitLines` | `internal/console/portfolio_pages.go:174` | `decoratePositionRows` | `exitFreshness`, `BuildExitLine`, `BuildExitLineReference` | one shared projection seam for both holdings routes; 11 AST branches/19 calls |
| HTTP `applyStoredPosition` | `cmd/tossctl/httpapi_reader.go:246` | `readPositions` | `WithFreshness(30s)`, `BuildExitLine`, `ExitLineFrom` | 4 symbols through `/api/v1/positions` reader |
| `operatorview.BuildExitLine` | `internal/operatorview/exit_line.go:66` | 10 callers across console, HTTP and tests | unknown/stale/fresh projection; no calculation | 25 symbols across both transports |

## Proven flow and current fault boundary

```text
ObserveOnce -> judge -> judgeRatchet/judgeLadder
  -> Evaluate*Snapshot -> ChangedFromState
  -> [current !Changed && !reJudge early return]
  -> record -> RecordExitJudgementResult -> recordExitJudgementTx
       -> one UPDATE of scalar + flattened + effective JSON
       -> optional proposal arm -> exit event -> COMMIT

LivePositionExits -> persisted ExitSnapshotView
  -> console exitFreshness -> BuildExitLine
  -> HTTP applyStoredPosition -> WithFreshness(30s) -> BuildExitLine -> ExitLineFrom
```

The current early return is before every persistence call. Consequently a SEED has no path to EVALUATED on an unchanged first quote, and an already-EVALUATED flat quote cannot advance persisted `ObservedAt`. A111 must distinguish those cases without moving invalid/stale observations or weakening proposal/stop ordering.

## JSON / flattened integrity boundary

`recordExitJudgementTx` encodes `StoredExitSnapshot` (including output digest), then uses one `UPDATE exit_states` statement to replace:

- scalar state: baseline, high-water, level/rung, update time;
- status and policy identity;
- snapshot/decision/observation IDs and position generation;
- next target/protection;
- observation source and `last_observed_at`;
- action/ratio/projected quantity/state-only/suppression;
- `effective_snapshot_json`.

The same transaction may arm a proposal and append `exit_events`, then commits once. Existing readers decode JSON and cross-check the flattened tuple; therefore a narrow refresh must preserve this all-or-nothing replacement and must append neither event nor proposal.

## Post-edit AST reconciliation (A111 implementation worktree)

The table above is the frozen-base CodeGraph collection. The post-edit source is
bound by the regenerated AST bundles in this directory. After the production
freeze, `make sdd-sync` completed successfully and refreshed the worktree graph
fingerprint for the final independent checks. The last shared-holdings refresh
synced 3 changed source files and 64 nodes. CodeGraph currently resolves
`handlePositions` as the direct decorator caller and `overview` as the direct
`accountPanelFrom` caller; current source plus both AST bundles provide the
missing transitive `accountPanelFrom -> decoratePositionRows` hard evidence.

```text
ObserveOnce
  -> preserve fill-detection priority -> one observe batch
  -> observe captures clock.LeaseAnchor after the broker read
       -> persisted FetchedAt remains UTC wall evidence
       -> process-local anchor retains monotonic system time
       -> injected clocks use deterministic Now/Since fallback
  -> validate source evidence with execgw.QueryPriceEvidenceDuration
       -> official fetched-at only: no MaxCycle journal read
       -> successful zero-time evidence only:
            MaxExitObservationCycle -> reserve one cycle:N lazily
  -> quoteUsable at loop/judge/refresh/record boundaries
       -> reject wall time before FetchedAt
       -> require LeaseElapsed(anchor) <= evidence duration
  -> judge -> judgeRatchet/judgeLadder
       -> unusable deadline: no write
       -> SEED/rejudge/operational difference: record (existing full path)
       -> exact EVALUATED nonorderable line: RefreshExitObservation
            -> explicit orderable/executable firewall before generic validation
            -> scan current state inside BEGIN IMMEDIATE
            -> exitObservationRefreshGuardTx in apply_hook.go
                 -> one status/pending SELECT; return only derived pending boolean
                 -> no UPDATE and no raw guarded-state authority escapes
            -> current managed lifecycle + temporal-CAS checks
            -> one complete non-guarded observation tuple UPDATE
            -> no pending write, proposal arm, exit_event, or order

positions cache read
  -> capture cacheAt -> readPositionBrokerSnapshot caches broker/account data only
  -> every request calls projectPositions with the cached broker snapshot
       -> finish blocking journal + policy + runtime reads
       -> exitResponseAuthority captures pre-marker clock
       -> read engine marker exactly once
       -> capture post-marker response clock
       -> only downgrade the read status; never resurrect stopped after rollback
       -> ApplyExitFreshness independently for every position
  -> console /positions: handlePositions -> decoratePositionRows
  -> console /dashboard: overview -> accountPanelFrom -> decoratePositionRows
       -> both holdings routes finish cached policy + settings reads
       -> capture markerReadAt -> readProtectionMarker exactly once -> capture responseAt
       -> protectionLivenessAt only downgrades; stopped cannot resurrect on rollback
       -> share responseAt across reconciliation age and attachPositionExitLines
       -> exitFreshness -> BuildExitLine closes every stale/stopped actionable value
       -> dashboard still uses holdings.peek only; policy-only delay causes no broker read
  -> console /position-management uses the parallel marker-bound freshness seam
       -> journal read failure remains explicit journal_unavailable truth
       -> safety-line rendering never promotes unavailable/SEED evidence
       -> finish livePositions + holdingNames + quarantine RPC
       -> capture markerReadAt -> readProtectionMarker exactly once -> capture asOf
       -> protectionLivenessAt only downgrades; stopped cannot resurrect on rollback
       -> share asOf across marker liveness and every rendered exit line
```

`compareObservationEvidence` gives same-time official evidence precedence over
`cycle:N`, makes equal identity a no-op, and rejects ambiguous identity. The
observer treats `ErrExitObservationConflict` and `ErrExitObservationStale` as
no-write outcomes for the next cycle. `EvidenceInvalidError` is classified as
permanent, so an HTTP-success/all-invalid quote batch neither retries nor
records entry-gate success. Every branch in these seams is linked to a named
A111 RED in the accompanying Branch Test Maps.

The final A111 reorder places the explicit orderable/executable-proposal
conflict guard before `validateJudgementSnapshot`. Consequently an identical
replay of an executable snapshot is still rejected as a refresh-authority
violation before a transaction can begin. The exact no-write proof is
`TestA111RefreshExplicitlyRejectsAnIdenticalExecutableSnapshotWithoutWriting`.

The final cache split removes the base `httpAPIReader.readPositions` function;
its base-revision AST remains as deletion evidence. Rate-limited broker data is
the only cached component. `TestA111RunningCachedPositionsRecheckTheExactPerPositionAgeBoundary`
proves a running response at exactly 30 seconds remains fresh and a later
request becomes stale without another broker read.
`TestA111RealPositionsRouteUsesPostReadClockForFreshnessAndMarkerLiveness`
additionally proves that time spent on a cache miss cannot make response
freshness or engine-marker liveness inherit the older cache-expiry clock.
`TestA111RealPositionsRouteUsesPostProjectionReadClockForFreshnessAndMarkerLiveness`
closes the remaining local-read interval: time spent in journal, policy and
runtime reads is also included before the single clock shared by liveness and
all per-position projections. The lazy fallback and shared price bound are jointly proved by
`TestA111FallbackSequenceRecoveryIsLazyAndPriceEvidenceUsesTheGateDuration`.
`TestA111MaxExitObservationCycleIgnoresEveryOutOfScopeEvidenceShape` closes the
account/lifecycle/completed/corrupt/official-source scan exclusions, while
`TestA111PositionManagementKeepsJournalFailureTruth` closes the console's
journal failure and safety-line branch.
`TestA111PositionManagementRechecksFreshnessAfterQuarantineRead` closes both
post-quarantine timing subcases: an unwired marker whose observation crosses
the 30-second bound, and a previously running marker that crosses the engine
stale bound while the quarantine RPC is in flight.

The production-freeze lease contract is proved by
`TestA111QuoteUseLeaseCannotBeExtendedByWallClockRollback` and
`TestA111ObserverUsesClockLeaseHelpersForTheUseLease`: wall-clock rollback can
neither extend a quote-use lease nor bypass the `clock.LeaseAnchor` /
`clock.LeaseElapsed` seam. `TestSystemLeaseAnchorRetainsMonotonicReading`
confirms that the system path preserves Go's monotonic reading without putting
it into a persisted UTC timestamp.

The HTTP post-marker boundary is proved by
`TestA111RealPositionsRouteUsesPostMarkerResponseClock` and
`TestA111PostMarkerClockRollbackCannotResurrectAStoppedEngine`. The equivalent
console boundary is proved by
`TestA111PositionManagementSamplesResponseTimeAfterMarkerRead` and
`TestA111PositionManagementNeverResurrectsAStoppedMarkerAfterClockRollback`.
Both consumers read the marker once, use the later response clock, and permit
only a running-to-stopped downgrade.

The final shared holdings decorator is proved on both `/positions` and
`/dashboard` by `TestA111HoldingsRoutesRecheckFreshnessAfterPolicyCacheMiss`:
time spent on the cached policy miss is included before both observation-age
and engine-marker projection, without a second broker read.
`TestA111HoldingsRoutesNeverResurrectStoppedMarkerAfterClockRollback` proves
that both routes retain a stopped marker verdict when the post-read wall clock
rolls backward. The four current AST/FLM/BTM bundles for `handlePositions`,
`accountPanelFrom`, `decoratePositionRows`, and `attachPositionExitLines` pin
the complete shared call path and every projection branch.

The repository gate's structural RED,
`TestGuardedExitColumnsAreWrittenOnlyByTheApplyHook`, fixes the authority
boundary independently of behavioral no-write assertions. The refresh source
no longer names `pending_action`; its unexported helper lives in
`apply_hook.go`, reads status/pending under the caller's existing transaction,
and returns only a boolean. `TestA111RefreshRejectsARealCurrentPendingProposalAndPreservesItsEvidence`
and `TestA111RefreshRejectsARealReleasedLifecycleAndPreservesItsTuple` retain
the real pending and lifecycle no-write proofs across that indirection. The A3
fresh-context re-review traced transaction reachability and the guarded-column
boundary and returned **ACCEPT — P0=0, P1=0**. This evidence update deliberately
does not run `make sdd-sync` while Manager task/PM documents are still moving;
the regenerated source-bound AST/FLM/BTM bundles are ready for that final sync.
