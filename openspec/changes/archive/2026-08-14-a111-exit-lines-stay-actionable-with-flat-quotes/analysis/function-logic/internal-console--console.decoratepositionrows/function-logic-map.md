# Function Logic Map: `Console.decoratePositionRows`

- Source: `internal/console/portfolio_pages.go`
- Post-edit AST evidence: `ast.json` (12 branches; revision `current`; source SHA-256 recorded by extractor)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `rows` | request-local broker/journal join for `/positions` or `/dashboard` | `joinPositions` / `Console.positions` | only display rows are mutated; no broker, journal, order, or config mutation |
| `asOf` | caller time used only for the engine-policy cache lookup | `enginePolicy.read` | it is never reused as the final exit-line freshness authority |
| policy/settings | optional runtime/list cache and one settings snapshot | engine-owned policy cache and settings reader | failed runtime/list stays unknown; failed settings load adds no desired claim |
| marker/time | exactly one marker read bounded by `c.now()` before and after it | `readProtectionMarker` + `protectionLivenessAt` | later time may downgrade running to stopped; stopped may never be resurrected by rollback |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `PositionPolicies != nil` at `internal/console/portfolio_pages.go:99` | records one cached engine reading; request-local only | unwired seam leaves runtime/effective state unknown | `TestPositionsShowRuntimeUnknownWhenCommanderUnavailableButDesiredIncludesUS` |
| B2 | policy state list succeeded at `internal/console/portfolio_pages.go:119` | allocates request-local `policyByID` | error leaves lifecycle lookup unavailable/fail-closed | `TestA111HoldingsRoutesRecheckFreshnessAfterPolicyCacheMiss` |
| B3 | range over returned states at `internal/console/portfolio_pages.go:121` | indexes states by trimmed position ID | duplicate IDs retain the last same-read state | `TestA111ActualPositionsAndPositionManagementRoutesRenderTheSameFreshLine` |
| B4 | settings seam wired at `internal/console/portfolio_pages.go:126` | may decorate desired include/exclude only | unwired settings contributes no desired claim | `TestA111ActualPositionsAndPositionManagementRoutesRenderTheSameFreshLine` |
| B5 | settings load succeeds at `internal/console/portfolio_pages.go:127` | permits one-snapshot desired decoration | load failure leaves existing request rows unchanged | `TestThePositionsScreenRendersWithEitherSourceMissing` |
| B6 | range over rows for settings at `internal/console/portfolio_pages.go:130` | stamps `Designated` and `Excluded` from one block | display-only mutation | `TestExcludingASymbolFromThePositionsScreen` |
| B7 | policy read was attempted at `internal/console/portfolio_pages.go:143` | enables effective management projection | unwired policy skips effective-state projection | `TestA111HoldingsRoutesRecheckFreshnessAfterPolicyCacheMiss` |
| B8 | range over rows for policy projection at `internal/console/portfolio_pages.go:144` | sets lifecycle/management display fields | each row shares the same runtime and response clock | `TestA111HoldingsRoutesRecheckFreshnessAfterPolicyCacheMiss` |
| B9 | row exists in journal at `internal/console/portfolio_pages.go:148` | requires lifecycle proof before actionable projection | broker-only row cannot invent lifecycle evidence | `TestA111ActualPositionsAndPositionManagementRoutesRenderTheSameFreshLine` |
| B10 | no matching engine state at `internal/console/portfolio_pages.go:151` | forces `journalKnown=false` | management remains unknown instead of guessed | `TestPositionsShowRuntimeUnknownWithoutDesiredFallback` |
| B11 | matching engine state at `internal/console/portfolio_pages.go:153` | records lifecycle status/generation and released proof | only same-read engine evidence can establish lifecycle | `TestA111ActualPositionsAndPositionManagementRoutesRenderTheSameFreshLine` |
| B12 | projected management has a reconciliation block at `internal/console/portfolio_pages.go:163` | builds block age from the post-marker `responseAt` | no pre-policy time can understate block age | `TestA111HoldingsRoutesRecheckFreshnessAfterPolicyCacheMiss` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `enginePolicy.read` | one cached runtime/list reading shared by the render | may block; failures are cached as failures and never revive older success | AST + policy-cache-miss RED |
| `Settings.Load` | stamp desired include/exclude from one coherent snapshot | error is ignored only by withholding desired decoration | AST + existing missing-source test |
| `readProtectionMarker` | read marker exactly once after policy/settings reads | called with a pre-read clock; returns read verdict without an upgrade path | AST + rollback RED |
| `protectionLivenessAt` | apply post-read response time to marker verdict | downgrade-only: running may become stopped, stopped remains stopped | `TestA111HoldingsRoutesNeverResurrectStoppedMarkerAfterClockRollback` |
| `attachPositionExitLines` | evaluate every persisted line at one response authority | stale/stopped/unknown evidence renders closed values | both named A111 holdings-route REDs |

## State mutations and fallbacks

- Mutates only request-local `positionRow` display fields.
- Policy/settings reads complete before `markerReadAt`; marker status is read exactly once; `responseAt` is captured afterwards and shared by management block age and all exit lines.
- A clock rollback cannot upgrade an already-stopped marker, while forward time crossing either the 30-second observation bound or engine stale bound closes the line.
- Runtime, lifecycle, settings, and marker failures withhold actionable claims rather than recomputing or reviving stored values.

## Safety conclusion

- Safe edit boundary: shared `/positions` and `/dashboard` read projection only.
- High-risk impact: yes—stale protection values are operator-facing; named post-policy and rollback REDs cover the changed timing seam.
