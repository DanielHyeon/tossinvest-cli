# Function Logic Map: `BudgetCoordinator.BeginObservation`

- Source: `internal/scheduler/budget.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| key | exact normalized endpoint budget key | future official request adapter | token digest and endpoint state are bound to this value; mismatch later fails closed |
| endpoint generation | current reset generation, or generation 1 for a first request | coordinator state | exhausted generation returns zero cycle |
| completion watermark | coordinator-wide monotonic sequence at request start | coordinator mutex | later completions have larger sequence and cannot be reconciled by this cycle |
| entropy / cycle memory | `crypto/rand.Reader`; max 1024 issued observation cycles per endpoint/generation | coordinator | unavailable entropy, cap, or collision returns zero cycle |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | nil coordinator | none | zero cycle | scope test |
| B2 | endpoint absent | initializes bounded generation-1 state | continue | initial/manual observation tests |
| B3 | entropy unavailable, generation exhausted, or cycle cap reached | none | zero cycle | entropy/cap fail-closed coverage |
| B4 | capability read fails | marks coordinator entropy unavailable | zero cycle | entropy failure coverage |
| B5 | capability collision in issued set | none | zero cycle | opaque capability collision coverage |
| success | capability is fresh | stores active record plus issued memory with generation and completion watermark | opaque cycle | held-response and binding tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `io.ReadFull` | mint fixed-width unpredictable cycle capability | short read/error fails closed | CodeGraph + AST |
| `sha256.Sum256` | bind exact endpoint key without exposing it | pure | cross-key test |
| coordinator mutex | serialize request-start watermark against completion | defer unlock | race test |

## State mutations and fallbacks

- Cycle issuance does not consume broker budget and never affects safety-class grants.
- Active cycles are one-shot; issued memory prevents capability reuse until a proven reset and is absolutely bounded.

## Safety conclusion

- Safe edit boundary: dormant low-priority budget chronology only.
- High-risk impact: yes, because fabricated chronology could spend reserved capacity.
