# Function Logic Map: `NewContext`

- Source: `internal/app/engine/engine.go:437-607`
- AST evidence: `ast.json` (Pre-Edit HEAD `a284a2b0a`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| options/path/config | validated startup dependencies | `NewOrderPath` and config | existing startup refusal |
| journal/account/gateway | owned startup resources | construction sequence | cleanup on every later failure |
| lower-limit board | exactly one production construction | a087 new leaf | admission literal nil, no GET capability |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B8 | path/audit/account/journal/gateway failures | ordered cleanup/refusal | existing error return | existing startup suite |
| B9-B12 | guardian factory and automation variants | guardian selection | existing error return | existing interlock suite |
| B13-B15 | projection creation/interlock outcome | context assembly | existing error return | a087 board creation structural test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `NewOrderPath` | obtain read-only official client | B1 | AST B1 |
| `openEngineJournal` | durable scope | B5 | AST B5 |
| `newExitLowerLimitBoard` | construct dormant board | no network and admission literal nil | a087 structural test |
| Context literal | preserve ownership boundaries | no error | AST final return |

## State mutations and fallbacks

- board is constructed only in final engine Context assembly, not shared order path or CLI.
- production admission argument is literal nil until a166 activation; no runtime flag or option can make it nonnil.

## Safety conclusion

- Safe edit boundary: additive private Context board field and construction.
- High-risk impact: yes — engine-off and human paths must have zero trace.
