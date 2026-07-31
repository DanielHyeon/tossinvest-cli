# Function Logic Map: `TestRemovingADesignationOnlyAffectsTheFuture`

- Source: `internal/console/settings_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| designated symbol | normalized symbol present in fake include list | test fixture | removal mismatch fails |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | POST remove designation and verify future-only notice | fake settings save | assertion failure | this test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `dashboardHarness.post` | exercise designation removal | local in-memory HTTP only | CodeGraph + AST |

## State mutations and fallbacks

- Only the fake block changes; insertion adjacency made this unchanged test part of the diff evidence.

## Safety conclusion

- Safe edit boundary: unchanged neighboring behavior remains pinned.
- High-risk impact: no; no production code or live state.
