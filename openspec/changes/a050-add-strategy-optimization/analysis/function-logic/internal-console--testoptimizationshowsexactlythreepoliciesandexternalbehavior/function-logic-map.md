# Function Logic Map: `TestOptimizationShowsExactlyThreePoliciesAndExternalBehavior`

- Source: `internal/console/optimization_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| exit-protection page | three owner-defined a041 options | `settingmeta` registry | assertion failure on missing policy or extra control |

## Branches and early returns

| Branch group | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B3 | expected copy and exact option count | none | assertion failure | this test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| fake lifecycle commander | supplies finite owner registry and snapshot | local deterministic seam | AST |

## State mutations and fallbacks

- Rendering performs no save and exposes only owner-issued option IDs.

## Safety conclusion

- Safe edit boundary: legacy test adapted to the category-scoped lifecycle UI.
- High-risk impact: no.
