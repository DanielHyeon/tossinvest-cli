# Function Logic Map: `ExitObserver.ladderFor`

- Source: `internal/app/engine/exitloop.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| state policy ID | default compatibility ID or registered common ID | journal state | unknown/mismatched injected policy errors |
| adoption flag | affects RUNNER floor-only semantics | position projection | wrong variant is identity conflict |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | blank/default ID | return injected default only if exact ID | error on mismatch | custom ladder tests |
| B2 | common ID | resolve copied registry policy/variant | error on unknown | common policy tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `CommonLadderForPosition` | resolve immutable executable table | unknown policy fails closed | CodeGraph + AST |

## State mutations and fallbacks

- Runtime identity is compared by the evaluator; this function only resolves the executable table.

## Safety conclusion

- Safe edit boundary: preserve existing resolution branches and fixed registry semantics.
- High-risk impact: yes — policy table selection.
