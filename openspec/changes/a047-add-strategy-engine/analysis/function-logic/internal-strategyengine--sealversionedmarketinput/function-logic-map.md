# Function Logic Map: `SealVersionedMarketInput`

- Source: `internal/strategyengine/contracts.go`
- CodeGraph callers/callees: lane parity fixtures only; no production caller in dormant a047
- AST: generated after implementation

## Inputs and invariants

| Input/state | Range | Source of truth | Failure behavior |
|---|---|---|---|
| provenance | exact market/calendar/frozen config/indicator versions | trusted adapters + frozen constants | refuse zero bundle |
| session | same-KST-day `09:00` open and authoritative close with more than 45m duration | official calendar | refuse zero bundle |
| entry cutoff | derived internally as `sessionClose - 45m`; no caller field | StockOS frozen config | caller cannot select cutoff |
| indicators | canonical required/optional decimals at one evaluation instant | sealed indicator snapshot | refuse zero bundle |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Test |
|---|---|---|---|---|
| B1 | provenance/version/time mismatch | none | provenance error | laundering table |
| B2 | trading session missing/too short, shifted from fixed 09:00 KST, or crosses KST day | none | calendar error | regular/early/short/shifted-day table |
| B3-B4 | required VWAP/EMA9 iteration or validation fails | exact parsing only | required-decimal error | both required malformed rows |
| B5 | slope decimal invalid | exact parsing only | slope-decimal error | malformed slope row |
| B6 | LVN decimal invalid | exact parsing only | LVN-decimal error | malformed LVN row |
| B7 | tangled score negative or invalid | exact parsing only | tangled-decimal error | negative tangled row |
| B8-B9 | optional current price is present and invalid | exact parsing only | live-price-decimal error | present/absent and malformed rows |
| B10-B12 | optional expansion/HVN iteration, presence, or validation fails | exact parsing only | optional-decimal error | both optional malformed plus absent rows |
| Success | all evidence exact | derive cutoff, copy UTC values into opaque bundle | valid immutable bundle | synthetic derivation plus translated StockOS final-bar/indicator parity tests |

## Calls and live bindings

| Callee | Contract | Failure path | Evidence |
|---|---|---|---|
| decimal helpers | exact strings only | fail closed; no coercion | lane tests |
| `time.Time.Add(-45m)` | frozen entry-cutoff derivation | short session refused | cutoff tests |

## State mutations and fallbacks

- Returns a new opaque value only. Invalid data yields the zero value; no coercion, clock read, or source fallback occurs.

## Safety conclusion

- Pure sealer only. No I/O, runtime wiring, order path, or fallback.
