# Function Logic Map: `consoleBroker.resolve`

- Source: `cmd/tossctl/console.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| cached broker | nil or one shared `verifylive.Broker` | console process | build errors return and are not cached |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | cached client exists | none | same client | console shared-client tests |
| B2 | factory fails | no cache mutation | error | console broker tests |
| success | first factory succeeds | caches broker under mutex | broker | shared resolution tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `verifyBrokerFactory` | construct one official client/account binding | error propagated | CodeGraph + tests |

## State mutations and fallbacks

- Mutex serializes first construction; this existing behavior is reused by calendar provenance.

## Safety conclusion

- Safe edit boundary: shared console read client.
- High-risk impact: medium; it contains broader broker authority but only narrow method values cross into console.
