# Function Logic Map: `Context.Close`

- Source: `internal/app/engine/engine.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| context journal | nil, open, or already closed | Context ownership | idempotent cleanup |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | nil/already closed | none | nil | close idempotence test |
| B2 | owned journal | clear field then close journal | journal close error | engine lifecycle test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Journal.Close` | release sole durable Context-owned store | no broker operation | CodeGraph + AST |

## State mutations and fallbacks

- Cleanup only; no mutation transport.

## Safety conclusion

- Safe edit boundary: retain nil-safe idempotent journal close only.
- High-risk impact: low, but resource leaks can affect restart durability.
