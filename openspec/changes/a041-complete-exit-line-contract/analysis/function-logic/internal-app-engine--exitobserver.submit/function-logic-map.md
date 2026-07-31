# Function Logic Map: `ExitObserver.submit`

- Source: `internal/app/engine/exitloop.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| armed proposal + provenance | deterministic intent and complete exit decision provenance | journal arm result | never submit before attach |
| reduction floor | confirmed nonzero quantity | floor source | release when zero/refused |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | floor error/zero | release or return | no broker call | existing floor tests |
| B2 | Guardian refuses | alert + release | no broker call | existing Guardian tests |
| B3 | attach fails | no broker call | error | crash-ordering tests |
| B4 | confirmed/in-doubt/symbol-busy/refused | log, retain, or release as specified | terminal outcome | existing gateway tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `IssueReduction` | authorize risk-reducing intent | refusal releases proposal | CodeGraph + AST |
| `AttachExitIntent` | persist intent link before broker mutation | failure aborts submit | CodeGraph + AST |
| `Submit.Place` | official mutation gateway | outcome controls retain/release | CodeGraph + AST |

## State mutations and fallbacks

- Adds typed exit provenance to `PlaceRequest`; it authorizes nothing and claims no a042 storage column.

## Safety conclusion

- Safe edit boundary: thread already-derived provenance without changing the Guardian/order path.
- High-risk impact: yes — live sell submission path (tests only invoke fakes).
