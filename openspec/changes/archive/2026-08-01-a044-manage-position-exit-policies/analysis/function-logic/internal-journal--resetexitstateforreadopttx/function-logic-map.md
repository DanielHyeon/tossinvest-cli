# Function Logic Map: `resetExitStateForReadoptTx`

- Source: `internal/journal/apply_hook.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| position/generation | exact released scope and next positive generation | lifecycle transaction | no partial commit |\n| observation | engine price, stop, time, registered policy | engine command service | invalid observation refused |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | ratchet seed invalid | none | invalid request | fresh-t0 test |\n| B2 | position source read fails | none | wrapped error | journal test |\n| B3 | policy identity invalid | none | error | registry test |\n| B4-B6 | reset/update/row-count fails | transaction rollback | error | atomicity tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `exitpolicy.OpenRatchetState` | derive fresh baseline/risk/watermark | deterministic validation | AST |\n| `appendExitEventTx` | record READOPT under new generation | same transaction | AST |

## State mutations and fallbacks

- Atomically clears old high-water, rung, taken ratio, pending proposal, and evaluated snapshot; writes the fresh lifecycle generation before its event.

## Safety conclusion

- Safe edit boundary: reset only the exact open exit row inside the lifecycle CAS transaction; never reuse old execution progress.
- High-risk impact: yes
