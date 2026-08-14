# Function Logic Map: `Console.handlePositions`

- Source: `internal/console/portfolio_pages.go`
- Post-edit AST evidence: `ast.json` (0 branches; revision `current`; source SHA-256 recorded by extractor)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| request/context | authenticated read-only `GET /positions` request | console router and request context | downstream broker/journal failures stay explicit in the snapshot |
| holdings snapshot | request-local rows plus broker freshness | `Console.positions` | no broker retry or mutation is introduced by decoration |
| response time | initial `c.now()` is only the policy-cache lookup time | `Console.decoratePositionRows` owns the later marker-bound response authority | blocking policy or marker reads cannot leave actionable values fresh at an older handler time |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| none | branch-free orchestration at `internal/console/portfolio_pages.go:70-82` | mutates only request-local page/rows, then renders HTML | downstream read truth is rendered without order/config/journal writes | `TestA111HoldingsRoutesRecheckFreshnessAfterPolicyCacheMiss` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `c.positions` | obtain the broker/journal join | existing cache and read-failure contract; no extra broker call | AST + `TestA111HoldingsRoutesRecheckFreshnessAfterPolicyCacheMiss` |
| `c.decoratePositionRows` | share management, lifecycle, marker liveness and exit projection with dashboard | decorator samples marker-bound response time only after blocking policy/settings reads | AST + both named A111 holdings-route REDs |
| `c.render` | render the fully decorated request-local page | read-only response | AST |

## State mutations and fallbacks

- Sets request-local refresh/explain fields and decorates request-local rows only.
- The handler's early clock sample is not the safety verdict: `decoratePositionRows` captures pre/post marker clocks after policy/settings reads.
- `/positions` and `/dashboard` converge on the same decorator and therefore the same fail-closed exit-line contract.

## Safety conclusion

- Safe edit boundary: read-only console projection; no LIVE order, journal write, configuration save, or broker refresh authority added.
- High-risk impact: operator-facing exit lines, covered by the shared-route boundary REDs.
