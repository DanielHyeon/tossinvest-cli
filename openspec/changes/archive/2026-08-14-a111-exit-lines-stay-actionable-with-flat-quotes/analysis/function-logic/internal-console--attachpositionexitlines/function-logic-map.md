# Function Logic Map: `attachPositionExitLines`

- Source: `internal/console/portfolio_pages.go`
- Post-edit AST evidence: `ast.json` (11 branches; revision `current`; source SHA-256 recorded by extractor)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| rows/runtime | request-local joined rows and one effective runtime snapshot | shared decorator | missing runtime/lifecycle stays unknown; desired config never substitutes for engine truth |
| `asOf` | one post-policy, post-marker response time for the entire render | `Console.decoratePositionRows` | stale/future evidence is closed; no per-row time drift |
| `live` | unwired/running/stopped verdict from one marker read | `protectionLivenessAt` | stopped closes values and cannot be upgraded in this function |
| exit snapshots | persisted canonical/legacy evidence only | journal | never recalculates stop/take-profit values |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | range over request rows at `internal/console/portfolio_pages.go:176` | decorates display-only exit fields | no external side effect | `TestA111ActualPositionsAndPositionManagementRoutesRenderTheSameFreshLine` |
| B2 | blank management status plus pending designation at `internal/console/portfolio_pages.go:179` | substitutes explicit unknown status | never promotes desired defaults | `TestPositionsShowRuntimeUnknownWithoutDesiredFallback` |
| B3 | row has no persisted exit at `internal/console/portfolio_pages.go:185` | builds non-price reference and continues | no actionable line is invented | `TestUSPendingAndBlockedShowEffectiveStopPlanWithoutPrice` |
| B4 | lifecycle generation mismatch at `internal/console/portfolio_pages.go:194` | clears raw display, builds unknown line/reference, continues | cross-generation evidence is suppressed | `TestPositionsSuppressCrossLifecycleExitEvidence` |
| B5 | lifecycle is released at `internal/console/portfolio_pages.go:208` | builds released unknown line/reference and continues | released position is never shown protected | `TestReleasedDesignatedRowDoesNotShowPendingFallback` |
| B6 | released reference is legacy raw at `internal/console/portfolio_pages.go:219` | preserves raw evidence only as nonactionable stored context | actionable line remains closed | `TestReleasedDesignatedRowDoesNotShowPendingFallback` |
| B7 | row is quarantined at `internal/console/portfolio_pages.go:225` | overrides freshness with `snapshot_quarantined` | quarantine dominates otherwise fresh evidence | `TestAQuarantinedPositionIsNotDrawnAsProtected` |
| B8 | quarantined view has no canonical snapshot at `internal/console/portfolio_pages.go:233` | sets unknown reason to quarantine | closed unknown line | `TestExitLinesStayClosedWhenTheEvidenceCannotBeTrusted` |
| B9 | freshness view contains canonical snapshot at `internal/console/portfolio_pages.go:243` | copies persisted line/source/time into display source | still filtered by stale reason in `BuildExitLine` | `TestA111HoldingsRoutesRecheckFreshnessAfterPolicyCacheMiss` |
| B10 | exit reference reports lifecycle unknown at `internal/console/portfolio_pages.go:258` | clears raw display and replaces line with lifecycle-unverified unknown | continues fail-closed | `TestPositionsSuppressCorruptAndLifecycleUnverifiedRawEvidence` |
| B11 | admitted reference is legacy raw at `internal/console/portfolio_pages.go:265` | exposes raw evidence in its separate nonactionable view | primary exit line remains canonical/fail-closed | `TestPositionsLeadWithManagedLegacyReferenceAcrossKRAndUS` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `exitFreshness` | apply shared age/integrity/marker liveness verdict | stopped/stale/future/unavailable closes actionable values | both named A111 holdings-route REDs |
| `operatorview.BuildExitLine` | convert persisted source plus stale/unknown reasons to UI line | stale/unknown fields render as closed dashes | `TestA111HoldingsRoutesRecheckFreshnessAfterPolicyCacheMiss` |
| `operatorview.BuildExitLineReference` | provide lifecycle/effective reference context | reference never authorizes a primary actionable line | lifecycle/reference tests |

## State mutations and fallbacks

- Mutates request-local display rows only and performs no calculation from market prices.
- Every row receives the same post-marker `asOf` and downgrade-only liveness from the shared decorator.
- Stale observation after a policy delay and a stopped marker after wall rollback both preserve stored evidence for audit while closing all actionable UI values.

## Safety conclusion

- Safe edit boundary: projection of already-persisted exit evidence.
- High-risk impact: yes—primary stop/take-profit display; fail-closed branches and both route-level REDs cover the timing change.
