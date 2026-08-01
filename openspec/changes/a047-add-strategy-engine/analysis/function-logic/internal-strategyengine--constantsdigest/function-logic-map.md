# Function Logic Map: `constantsDigest`

- Source: `internal/strategyengine/lane.go`
- CodeGraph callers/callees: decision mint + activation binding
- AST: generated after implementation

## Inputs and invariants

| Input/state | Range | Source of truth | Failure behavior |
|---|---|---|---|
| constants | exact ordered StockOS conservative v1 key/value lines, including 30m/10m auctions, 400m after-hours offset and 45m cutoff buffer | frozen source commit | any change produces a different digest and invalidates manifest equality |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Test |
|---|---|---|---|---|
| Success | fixed payload | pure SHA-256 | digest | translated parity + decision binding tests |

## Calls and live bindings

| Callee | Contract | Failure path | Evidence |
|---|---|---|---|
| SHA-256 | deterministic exact bytes | none | tests |

## State mutations and fallbacks

- Pure local hashing; no mutable state or fallback.

## Safety conclusion

- Pure constant identity. Any change requires a new lane version/manifest.
