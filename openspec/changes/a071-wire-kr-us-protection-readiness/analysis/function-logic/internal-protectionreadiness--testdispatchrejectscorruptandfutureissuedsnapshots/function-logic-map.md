# Function Logic Map: `TestDispatchRejectsCorruptAndFutureIssuedSnapshots`

- Source: `internal/protectionreadiness/dispatch_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| valid sealed KR snapshot then mutated provenance | mutation invalidates market seal | test fixture | assert state-corrupt and denied |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | sealed expiry changed without resealing | test-only value mutation | `RefusalStateCorrupt` | named test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Assess`, `Dispatch` | create then validate sealed snapshot | fail assertion | CodeGraph + AST |

## State mutations and fallbacks

- No production mutation; corrupts a local snapshot copy only.

## Safety conclusion

- Safe edit boundary: corruption rejection assertion only
- High-risk impact: no (test)
