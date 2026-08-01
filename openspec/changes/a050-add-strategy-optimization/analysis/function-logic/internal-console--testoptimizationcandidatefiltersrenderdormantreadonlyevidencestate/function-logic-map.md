# Function Logic Map: `TestOptimizationCandidateFiltersRenderDormantReadOnlyEvidenceState`

- Source: `internal/console/optimization_candidate_filters_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| rendered candidate-filter category | server HTML | fixed a050 category IA | test fails on missing evidence markers or any mutation control |

## Branches and early returns

| Branch group | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B6 | expected text and DOM nodes | none | assertion failure | this test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| dashboard harness and HTML walker | verify read-only rendered contract | deterministic local test | AST |

## State mutations and fallbacks

- No runtime mutation; the test asserts the absent a045 adapter remains unavailable and input-free.

## Safety conclusion

- Safe edit boundary: category deep-link expectation only.
- High-risk impact: no.
