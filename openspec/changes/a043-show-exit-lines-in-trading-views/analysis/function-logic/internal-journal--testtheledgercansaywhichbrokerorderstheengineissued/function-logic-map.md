# Function Logic Map: `TestTheLedgerCanSayWhichBrokerOrdersTheEngineIssued`

- Source: `internal/journal/readonly_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| v10 fixture | bare broker ids | base test | removed with unsafe API |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B2 | query error/result mismatch | test-only | fail test | base evidence |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `BrokerOrderIDs` | formerly asserted global id list | test only | base AST |

## State mutations and fallbacks

- Removed and replaced by account/day collision and scoped origin tests.

## Safety conclusion

- Safe edit boundary: obsolete test removal.
- High-risk impact: no.
