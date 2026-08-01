# Function Logic Map: `TestCandidateFilterCardsContainExactMarketMetricMatrix`

- Source: `internal/console/optimization_candidate_filters_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| KR/US metric cards | exactly three known metrics per market | a045 read-model UI contract | assertion failure |

## Branches and early returns

| Branch group | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B5 | market/card/metric counts | none | assertion failure | this test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| DOM query helpers | count exact accessible card structure | deterministic local test | AST |

## State mutations and fallbacks

- No mutation; only the category deep link changed.

## Safety conclusion

- Safe edit boundary: rendered read-only evidence structure.
- High-risk impact: no.
