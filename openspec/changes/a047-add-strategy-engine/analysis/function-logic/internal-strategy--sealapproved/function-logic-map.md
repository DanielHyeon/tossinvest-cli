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
| B1 | branchless happy-path sentinel (AST has no conditional); any approved or zero input | scalar copies only | snapshot mirrors validity, identity, thresholds and current-life proof exactly | candidate boundary and lane current-life tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| candidate scalar accessors | copy provenance without importing capabilities | pure, no error/retry | AST + pure-boundary typecheck |

## State mutations and fallbacks

- No mutation/fallback; no time, pointer, callback or mutable collection crosses the boundary.

## Safety conclusion

- Safe edit boundary: state/last-seen/exclusive validity are copied from the sealed approval, never caller asserted.
- High-risk impact: yes — candidate-to-strategy sanitizing handoff.
