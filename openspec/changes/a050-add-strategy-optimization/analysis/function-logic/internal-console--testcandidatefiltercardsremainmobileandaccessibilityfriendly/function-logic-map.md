# Function Logic Map: `TestCandidateFilterCardsRemainMobileAndAccessibilityFriendly`

- Source: `internal/console/optimization_candidate_filters_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| candidate-filter HTML | wrapping cards, anchors and read-only semantics | a050 UI contract | assertion failure |

## Branches and early returns

| Branch group | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B3 | wide table absent and accessibility markers present | none | assertion failure | this test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| HTML helpers | inspect mobile/accessibility contract | deterministic local test | AST |

## State mutations and fallbacks

- No mutation; the unavailable provider remains read-only.

## Safety conclusion

- Safe edit boundary: category deep-link expectation only.
- High-risk impact: no.
