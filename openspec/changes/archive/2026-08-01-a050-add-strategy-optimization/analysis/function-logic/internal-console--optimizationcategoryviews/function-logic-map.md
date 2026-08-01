# Function Logic Map: `optimizationCategoryViews`

- Source: `internal/console/optimization_view.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| selected category | one server-defined category ID | `optimization.Categories` | unknown selection is normalized by the caller before this projection |
| lifecycle writable | true only when canonical optimization command seam exists | `Console.Options.Optimization` | exit-protection is labelled read-only when false |
| category descriptors | fixed ordered six-category registry | `optimization.Categories` | no client-provided category or status is rendered |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B2 | iterate and dispatch every fixed category | projection only | returns six ordered views | category contract tests |
| B3 | overview | projection only | read-only summary status | category contract tests |
| B4-B6 | exit-protection writable/unwired | projection only | owner-connected or command-unwired status | optimization lifecycle tests |
| B7-B9 | position/candidate/runtime categories | projection only | their owner/read-only status | optimization UI contract tests |
| B10 | performance-history | projection only | a049 deterministic performance, read-only, with evidence status delegated to body | performance/evidence UI tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strategyopt.Categories` | supplies the fixed order and server-owned labels | no error path; immutable copy | CodeGraph + AST |

## State mutations and fallbacks

- Mutates only the local view slice.
- The performance category no longer claims a049 is unintegrated; availability of actual evidence is still rendered from `optimization.Evidence`, so missing/error states are not converted to success.

## Safety conclusion

- Safe edit boundary: category status text and read-only availability only.
- High-risk impact: no trading mutation; evidence-backed candidate gating remains in the optimization service/provider.
