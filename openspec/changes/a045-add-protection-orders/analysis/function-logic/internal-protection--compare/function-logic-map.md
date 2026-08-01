# Function Logic Map: `Compare`

- Source: `internal/protection/reconcile.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Function inputs | One explicit typed scope plus local sagas and broker protections that must all belong to it. | Current HEAD + OpenSpec | Fail closed with typed error/decision |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1+ | Existing branches skip terminal, map by broker ID/symbol, and emit missing/orphan/duplicate/quantity/trigger findings. | No external mutation. Existing map overwrite lets duplicate broker IDs and mixed scopes corrupt ownership. | Typed refusal or validated result | See branch map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| Current callees | Pure map construction and comparison. | No implicit retry; errors propagate fail-closed | CodeGraph + AST |

## State mutations and fallbacks

- No external mutation. Existing map overwrite lets duplicate broker IDs and mixed scopes corrupt ownership.

## Safety conclusion

- Safe edit boundary: Prevalidate every item scope and broker ID uniqueness; return typed fail-closed error/discrepancy before ordinary comparison.
- High-risk impact: yes; dormant logic only, no broker mutation or WIRED binding.
