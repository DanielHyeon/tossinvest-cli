# Function Logic Map: `mintSavedStopProvenance`

- Source: `internal/continuationlane/execution_terms.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| plan/envelope | validated campaign plan and source evidence | internal strategy-flow/test seam | unsupported quote scale returns zero authority |
| price | selected saved effective stop minor units | saved stop state | bound into provenance and seal |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | unsupported quote currency scale | none | zero authority | unsupported-currency tests |
| success | supported scale | constructs private authority and seal | sealed authority | saved-stop regression tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `currencyMinorScale` | canonical minor-unit scale | false; no retry | AST B1 |
| `savedStopAuthoritySeal` | bind provenance to plan/evidence | deterministic; no I/O | AST |

## State mutations and fallbacks

- Creates a value only; no shared-state mutation or I/O.
- No public constructor or exported field exposes the authority seal.

## Safety conclusion

- Safe edit boundary: internal saved-stop authority minting only.
- High-risk impact: yes; authority remains package-private and exact-preimage sealed.
