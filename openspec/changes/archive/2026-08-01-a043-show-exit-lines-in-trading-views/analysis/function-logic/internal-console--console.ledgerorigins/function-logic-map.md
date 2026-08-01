# Function Logic Map: `Console.ledgerOrigins`

- Source: `internal/console/orders.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs and deterministic fixture/read state | values accepted by the typed function signature | current source plus OpenSpec a043 | tests fail explicitly; production reads degrade to typed unknown/unlinked evidence |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if branch at source line 635 | bounded test/read-model control flow only | Console.ledgerOrigins coverage and focused package suite |
| B2 | if branch at source line 641 | bounded test/read-model control flow only | Console.ledgerOrigins coverage and focused package suite |
| B3 | range branch at source line 646 | bounded test/read-model control flow only | Console.ledgerOrigins coverage and focused package suite |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| typed callees listed in `ast.json` | Existing read-only broker order-id origin projection; body is unchanged and was re-mapped because adjacent lineage code changed the diff span. | no retry is introduced; read errors and assertions preserve their existing fail-closed behavior | current AST and focused tests |

## State mutations and fallbacks

- Existing read-only broker order-id origin projection; body is unchanged and was re-mapped because adjacent lineage code changed the diff span.
- No live order call, operating-toggle write, or policy recomputation is introduced by this function change.

## Safety conclusion

- Safe edit boundary: read-model/view/test contract for a043 only.
- High-risk impact: no; account mutation and execution paths are unreachable.
