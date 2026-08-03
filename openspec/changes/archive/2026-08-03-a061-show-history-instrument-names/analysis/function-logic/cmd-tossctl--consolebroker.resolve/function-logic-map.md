# Function Logic Map: `consoleBroker.resolve`

- Source: `cmd/tossctl/console.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| receiver gate | free or held only by this call | `consoleBroker.lock` | serializes first account resolution and shared client reuse |
| cached client | nil or account-resolved live broker | `consoleBroker.client` | nil triggers existing factory; non-nil returns unchanged |
| account reference | masked/trimmed display reference | `verifyBrokerFactory` | factory error returns no client/ref |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | defensive broker gate acquisition failure | none | return error | background context makes branch unreachable; lock cancellation tested on metadata path |
| B2 | cached client exists | none | return cached client/ref | `TestOpeningEveryConsoleReadScreenResolvesTheAccountOnce` |
| B3 | no client and existing factory returns error | no cache mutation | return error | console seam failure tests |
| tail | no client and existing factory succeeds | cache client and trimmed reference | return cached pair | shared resolver tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `verifyBrokerFactory` | construct and resolve the console's account-scoped official broker once | called under gate; failure is not cached so later login/refresh can recover | AST B5 |
| `strings.TrimSpace` | canonicalize the display account reference | pure | AST tail |

## State mutations and fallbacks

- Mutates only process-local cached client/account reference after a successful factory call.
- A failed factory call leaves the cache empty and is retryable.

## Safety conclusion

- Safe edit boundary: preserve the existing cached account-resolved path and failure retry semantics.
- High-risk impact: yes. The function selects the client later used by account-scoped reads; branch tests must prove one `/accounts` resolution and token-manager identity preservation.
