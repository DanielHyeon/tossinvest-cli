# Function Logic Map: `SealApproved`

- Source: `internal/strategy/approved.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| approved candidate | exact immutable `candidate.ApprovedCandidate` | candidate approval boundary | invalid remains invalid |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | zero/refused value | none | invalid scalar snapshot | candidate boundary tests |
| B2 | approved value | scalar copies only | snapshot preserves identity, thresholds and current-life proof | lane fixture/current-life tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| candidate scalar accessors | copy provenance without importing capabilities | pure, no error/retry | AST + pure-boundary typecheck |

## State mutations and fallbacks

- No mutation/fallback; no time, pointer, callback or mutable collection crosses the boundary.

## Safety conclusion

- Safe edit boundary: state/last-seen/exclusive validity are copied from the sealed approval, never caller asserted.
- High-risk impact: yes — candidate-to-strategy sanitizing handoff.
