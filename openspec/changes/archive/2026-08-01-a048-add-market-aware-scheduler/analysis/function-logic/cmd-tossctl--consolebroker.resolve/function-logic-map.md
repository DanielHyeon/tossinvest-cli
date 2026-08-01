# Function Logic Map: `consoleBroker.resolve`

- Source: `cmd/tossctl/console.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| cached broker | nil or one shared `verifylive.Broker` | console process | build errors return and are not cached |
| cached account reference | trimmed exact reference returned with the broker | `verifyBrokerFactory` | empty on build error; never synthesized |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | cached client exists | none | same client, cached exact account reference, nil | `TestConsoleBrokerTypedMarketCalendarReusesResolutionAndKeepsExactAccountRef` |
| B2 | factory fails | no cache mutation | nil broker, empty reference, original error | `TestConsoleBrokerTypedMarketCalendarFailsClosed/resolver_error` |
| success | first factory succeeds | trims and caches broker plus account reference under mutex | broker, exact reference, nil | shared resolution tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `verifyBrokerFactory` | construct one official client/account binding | error propagated without caching | CodeGraph + tests |
| `strings.TrimSpace` | canonicalize only surrounding whitespace while preserving exact broker identity | no fallback or masking | AST + exact-reference assertion |

## State mutations and fallbacks

- Mutex serializes first construction across positions, orders, and market-calendar reads.
- The exact trimmed account reference remains cached for a043 order evidence lineage even when a caller such as the calendar adapter discards its returned copy.

## Safety conclusion

- Safe edit boundary: shared console read client.
- High-risk impact: medium; it contains broader broker authority but only narrow method values cross into console.
