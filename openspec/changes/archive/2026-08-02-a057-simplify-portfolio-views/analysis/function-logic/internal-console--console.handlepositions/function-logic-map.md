# Function Logic Map: `Console.handlePositions`

- Source: `internal/console/portfolio_pages.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| HTTP request context | authenticated GET `/positions` | route/session middleware | request cancellation flows to read seams; no retry |
| Holdings/journal rows | KR, US, other; either source may be absent | `positions()` and read-only journal | remaining source renders; missing facts stay unknown |
| Decorated rows | canonical, raw-only, stale, mismatch, absent | shared `decoratePositionRows` projection | actionable values are suppressed unless canonical and valid |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | Branchless page assembly | local page and row projection only | render once | positions suite + a057 shared projection test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `c.positions` | lazy holdings refresh plus journal join | holdings cache contract bounds broker calls; journal is read-only | CodeGraph callees + AST |
| `decoratePositionRows` | shared desired/effective/lifecycle/exit projection | single request-local pass; read failures remain unknown | AST + a057 parity tests |
| `c.render` | render the input-free page | template error is handled by renderer | route caller + AST |

## State mutations and fallbacks

- Mutations are limited to request-local `positionRow` display fields and template rendering.
- All policy/settings/lifecycle branches live in `decoratePositionRows`; the
  handler deliberately has no independent interpretation branch.
- Unknown runtime/lifecycle and stale/generation-mismatched evidence remain non-actionable.

## Safety conclusion

- Safe edit boundary: extract request-local enrichment and preserve the holdings refresh/render boundary.
- High-risk impact: no; this is a read-only console projection with no order, config save, or toggle mutation.
