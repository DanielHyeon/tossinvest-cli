# Function Logic Map: `ThresholdRegistry.LoadThresholdSet`

- Source: `internal/candidate/thresholdset.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| bound threshold inputs | accepted by package `LoadThresholdSet` | activation/evidence binding contract | zero set + error |
| registry version map | one canonical digest per version | a046 design decision 9 | same-version different-digest returns zero + error |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | binding loader fails | none | zero + propagated error | binding failure matrix |
| B2 | registry rejects alias | no conflicting write | zero + conflict error | same-version different-digest test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| package `LoadThresholdSet` | validate full binding | synchronous, no retry | CodeGraph + AST |
| `ThresholdRegistry.register` | serialize version/digest ownership | mutex protected, no I/O | registry race test |

## State mutations and fallbacks

- On success only, the registry records version→canonical digest under a mutex. No replacement or numeric fallback.

## Safety conclusion

- Safe edit boundary: immutable version registry; no order/RiskIntent dependency.
- High-risk impact: yes for approval integrity, covered by conflict and race tests.
