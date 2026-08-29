# Function Logic Map: `NewStrategyEntrySupervisor`

- Source: `internal/app/engine/strategy_entry_supervisor.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `Workers` | exactly one `KR` and one `US` descriptor | a072 engine-safety + constructor | refuse the whole assembly |
| queue/cycle bounds | queue 1..1024; cycle 0..30s | local constants | refuse the whole assembly |
| effective descriptor | cycle + unexpired generation/evidence/latch authority | descriptor + injected clock | refuse that assembly; no loop starts |
| dormant descriptor | zero authority and no effective entry | release default | refuse authority-bearing dormant input |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | invalid queue/cycle bound | none | validation error | invalid assembly table |
| B2 | worker count/market duplicate/missing | none | validation error | invalid assembly table |
| B3 | effective worker lacks cycle/current authority | none | validation error | invalid assembly table |
| B4 | dormant worker carries authority | none | validation error | invalid assembly table |
| B5 | valid paired workers | allocate independent queues/runtime state | supervisor | dormant + paired active tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `clk.Now` | validate authority at construction | zero clock fails closed | AST B3 |
| `validStrategyDigest` | canonical evidence validation | false refuses assembly | AST B3/B4 |

## State mutations and fallbacks

- The edit may add restart metadata validation only; it must not add activation, broker, Gateway, journal or approval dependencies.
- KR and US remain separately allocated and production defaults remain two dormant descriptors.

## Safety conclusion

- Safe edit boundary: constructor-only validation and initialization of market-local restart/refusal observation state.
- High-risk impact: yes — engine entry supervision is safety-adjacent, while the callback remains evaluation-only.
