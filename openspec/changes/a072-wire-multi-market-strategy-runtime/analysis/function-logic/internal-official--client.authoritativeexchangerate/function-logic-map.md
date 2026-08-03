# Function Logic Map: `Client.AuthoritativeExchangeRate`

- Source: `internal/official/market_reads.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| client/context/currency pair | non-nil sealed production client | official client configuration | ErrAuthorityOrigin before transport on mismatch |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | nil receiver | none | ErrAuthorityOrigin | configured/nil authority tests |
| B2 | locked sealed-origin predicate fails | none | ErrAuthorityOrigin | configured zero-hit test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `ExchangeRate` | execute exact official GET while retaining config read lock | normal context/transport errors propagate | boundary success test |

## State mutations and fallbacks

- Holds `configMu.RLock` continuously from origin validation through completion of token/data HTTP operations.

## Safety conclusion

- Safe edit boundary: do not split origin check from HTTP read or permit fallback transport.
- High-risk impact: yes, this is the only read allowed to mint official FX authority.
