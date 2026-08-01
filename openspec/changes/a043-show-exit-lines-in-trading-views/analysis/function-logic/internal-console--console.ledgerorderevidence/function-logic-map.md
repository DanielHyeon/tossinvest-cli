# Function Logic Map: `Console.ledgerOrderEvidence`

- Source: `internal/console/orders.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| filtered visible scopes | id+account+market+market-day tuples, possibly empty | rendered rows | empty skips evidence SQL; failure marks journal unreadable |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if branch at source line 660 | bounded test/read-model control flow only | Console.ledgerOrderEvidence coverage and focused package suite |
| B2 | if branch at source line 666 | bounded test/read-model control flow only | Console.ledgerOrderEvidence coverage and focused package suite |
| B3 | if branch at source line 671 | bounded test/read-model control flow only | Console.ledgerOrderEvidence coverage and focused package suite |
| B4 | range branch at source line 676 | bounded test/read-model control flow only | Console.ledgerOrderEvidence coverage and focused package suite |
| B5 | range branch at source line 680 | bounded test/read-model control flow only | Console.ledgerOrderEvidence coverage and focused package suite |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `BrokerOrderExitLinks` | get origin and exact lineage only for visible scopes | no retry; any error fails all evidence closed | current AST and focused tests |

## State mutations and fallbacks

- No unbounded `BrokerOrderIDs` query remains; empty pages do not issue evidence SQL.
- No live order call, operating-toggle write, or policy recomputation is introduced by this function change.

## Safety conclusion

- Safe edit boundary: read-model/view/test contract for a043 only.
- High-risk impact: no; account mutation and execution paths are unreachable.
