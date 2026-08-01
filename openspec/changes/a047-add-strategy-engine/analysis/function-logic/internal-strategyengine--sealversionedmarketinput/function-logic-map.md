# Function Logic Map: `SealVersionedMarketInput`

- Source: `internal/strategyengine/contracts.go`
- CodeGraph callers/callees: lane parity fixtures only; no production caller in dormant a047
- AST: generated after implementation

## Inputs and invariants

| Input/state | Range | Source of truth | Failure behavior |
|---|---|---|---|
| provenance | exact market/calendar/frozen config/indicator versions | trusted adapters + frozen constants | refuse zero bundle |
| session | trading-day open/close with more than 45m duration | official calendar | refuse zero bundle |
| entry cutoff | derived internally as `sessionClose - 45m`; no caller field | StockOS frozen config | caller cannot select cutoff |
| indicators | canonical required/optional decimals at one evaluation instant | sealed indicator snapshot | refuse zero bundle |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Test |
|---|---|---|---|---|
| B1 | provenance/version/time mismatch | none | provenance error | laundering table |
| B2 | trading session missing or too short for frozen buffer | none | calendar error | regular/early/short-session table |
| B3 | required or optional decimal invalid | exact parsing only | decimal error | indicator table |
| Success | all evidence exact | derive cutoff, copy UTC values into opaque bundle | valid immutable bundle | golden lane test |

## Calls and live bindings

| Callee | Contract | Failure path | Evidence |
|---|---|---|---|
| decimal helpers | exact strings only | fail closed; no coercion | lane tests |
| `time.Time.Add(-45m)` | frozen entry-cutoff derivation | short session refused | cutoff tests |

## State mutations and fallbacks

- Returns a new opaque value only. Invalid data yields the zero value; no coercion, clock read, or source fallback occurs.

## Safety conclusion

- Pure sealer only. No I/O, runtime wiring, order path, or fallback.
