# Function Logic Map: `QFinalEntryIssuance.SetFXAuthorityForTest`

- Source: `internal/execgw/export_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| public arithmetic FX evidence | valid test fixture only | test helper | downstream precheck validates the entire reserve policy |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | happy path | stores a value copy and marks test authority present | no error | focused q_final tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| none | direct private-field assignment | test binary only | AST |

## State mutations and fallbacks

- The test seam cannot exist in production builds and copies non-aliasing FX facts.

## Safety conclusion

- Safe edit boundary: do not expose a production setter for raw FX evidence.
- High-risk impact: test coverage for q_final authority only.
