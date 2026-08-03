# Function Logic Map: `attachPositionExitLines`

- Source: `internal/console/portfolio_pages.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Base commit: `openspec/changes/a061-screens-show-what-they-already-know/base-commit.txt`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `rows []positionRow` | request-local display rows from `joinPositions` | `internal/console/portfolio.go` | mutated in place; never shared across requests |
| `row.HasExit` | true only when the journal returned an exit state | `journal.ReadOnly.LivePositionExits` | false -> reference-only branch, no `ExitLine` values |
| `row.Exit.Snapshot` | `journal.ExitSnapshotView`; nil `Snapshot` carries a typed `UnknownReason` | `scanExitStateResult` | nil -> `BuildExitLine` renders `unknown`, every value an em dash |
| `row.LifecycleKnown`, `row.LifecycleGeneration` | lifecycle proof from `PositionPolicies.List` | `decoratePositionRows` | unknown -> `lifecycle_generation_unverified`, values suppressed |
| `row.Exit.LifecycleGeneration` | generation the stored evidence belongs to | journal | mismatch -> every stored price suppressed |
| `row.Quarantined()` (a061) | active quarantine on the current position generation only | `journal.ReadOnly.LivePositionExits` | read failure -> false, i.e. the pre-a061 behaviour; cannot invent a quarantine |
| `asOf time.Time` | the render instant, `c.now()` | `Console.now` | zero -> both freshness helpers are no-ops by their own guards |
| `runtime positionpolicy.ManagementRuntime` | zero value means the seam was never wired | `PositionPolicies.Runtime` | `EffectiveKnown` false -> reference view stays explicitly unknown |
| `live protectionLiveness` (a061) | `{Wired, Running}` from the engine marker | `Console.protectionLiveness` | unwired -> the pre-a061 observation-age bound |

**Invariant this function exists to hold**: no actionable price is computed here
or in the template. Every value it writes is copied from an already-persisted
snapshot, or it is an em dash.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `for i := range rows` | iterates request-local rows | none | every test below |
| B2 | `referenceStatus == "" && row.PendingDesignation()` | `referenceStatus = ManagementStatusUnknown` | falls through | existing a053 tests |
| B3 | `!row.HasExit` | writes `ExitReference` only | `continue` | existing a053 tests |
| B4 | lifecycle generation mismatch | clears `StoredExit`, `ExitLine = unknown(lifecycle_generation_mismatch)` | `continue` | existing a053 mismatch test |
| B5 | `LifecycleStatus == StatusReleased` | `ExitLine = unknown(operator_released_lifecycle)` | `continue` | existing release test |
| B6 | `ExitReference.LegacyRaw()` inside B5 | `StoredExit = storedExitEvidence` | falls to `continue` | existing legacy-raw test |
| **B7 (a061)** | `row.Quarantined()` | forces `Stale`/`StaleReason = snapshot_quarantined` | falls through | a061 2.4, 2.5 |
| **B8 (a061)** | quarantined **and** no canonical snapshot | also sets `UnknownReason` so the quarantine survives the unknown path | falls through | a061 2.4 |
| B9 | `snapshot.Snapshot != nil` | copies `Line`, `ObservationSource`, `ObservedAt` into `Source` | falls through | a061 2.1, existing fixtures |
| B10 | `ExitReference.LifecycleUnknown()` | clears `StoredExit`, `ExitLine = unknown(lifecycle_generation_unverified)` | `continue` | existing lifecycle-unknown test |
| B11 | `ExitReference.LegacyRaw()` | `StoredExit = storedExitEvidence` | end of iteration | existing legacy-raw test |

The freshness call itself (line 192, `exitFreshness`) is a061's replacement for
`WithFreshness(asOf, holdingsTTL)`. It is a call rather than a branch here; its
own three-way decision is mapped in `internal/console/protection_liveness.go`,
which is a new file and therefore outside this function's branch set.

**Ordering is load-bearing**: B7 runs after the freshness call and overrides it,
because a quarantined position must be closed regardless of how live the engine
is or how fresh the evidence looks. B4/B5 stay ahead of B7 because a generation
mismatch already hides the identity a quarantine reason would refer to.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `operatorview.BuildExitLineReference` | provenance/reference row, never actionable prices | pure, no error | AST, `internal/operatorview/exit_line_reference.go` |
| `operatorview.BuildExitLine` | the fail-closed five values | pure, no error; all em dashes on stale/unknown | AST, `internal/operatorview/exit_line.go` |
| `exitFreshness` (a061) | replaces the age bound with a liveness question | pure; falls back to `WithFreshness` when the marker is unwired | `internal/console/protection_liveness.go` |
| `storedExitReference`, `storedExitEvidence` | legacy raw evidence for the detail popup | pure | AST |

No network call, no journal write, no order and no config save is reachable from
this function or anything it calls. a061 does not change that.

## State mutations and fallbacks

- Writes only to `rows[i]`: `ExitLine`, `ExitReference`, `StoredExit`. No package
  state, no shared map, no clock read of its own.
- Fallback direction is always toward the em dash. Both conditions a061 adds
  (`engine_not_running` via `exitFreshness`, `snapshot_quarantined` here)
  suppress values that would otherwise show; the one path that reveals values is
  gated on positive liveness plus the unchanged lifecycle and integrity checks.
- A quarantine read failure degrades to `Quarantined() == false` and cannot
  manufacture a quarantine that is not on disk.

## Safety conclusion

- Safe edit boundary: display only. The function does not evaluate an exit
  policy, does not touch `exit_states`, and cannot reach an order path.
- High-risk impact: no, by the WORKFLOW risk list (no order submit/cancel/modify,
  no stop/target/sizing logic, no Guardian or kill switch, no ledger write, no
  reconciliation, no retry matrix, no auth, no fill detection). It draws the
  protection line, which is adjacent to a High-risk subject, so the change
  carries an adversarial Eng voice in `review.md` anyway.
- Safety invariant 0.3 (stop immediacy) untouched: the engine's judgement loop
  never reads this function.
