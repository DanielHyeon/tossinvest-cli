# Function Logic Map: `consoleBroker.instrumentMetadata`

- Source: `cmd/tossctl/console.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| request context | live history request with outer 10-second deadline | `instrumentNameCache.get` | canceled/expired context reaches builder and `/stocks` |
| broker gate | zero-value or initialized one-token channel | `consoleBroker.lock` | request cancellation ends gate waiting before client access |
| cached broker | nil or account-resolved | `consoleBroker.client` | nil builds once; non-nil reuses the identical instance |
| account builder | production `buildConsoleAccountBroker` or test fake | `newConsoleBroker` | nil/error returns explicit failure without cache mutation |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | broker gate cannot be acquired before context expiry | none | return context error | contended resolver cancellation test |
| B2 | no cached client | none yet | enter account-resolved builder path | shared-client test |
| B3 | account builder is nil | none | explicit configuration error | capability failure test |
| B4 | account builder fails | none | return context/credential error | cold resolver cancellation test |
| B5 | cached/built broker lacks `Stocks` | retain broker | explicit type error | narrow capability tests |
| tail | broker exposes `Stocks` | none | return narrow metadata reader | adapter rendering/chunk tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `build` | construct and account-resolve the one official client | request context bounds construction; failure is retryable | AST B2-B4 |
| `lock` / `unlock` | serialize client generation and reuse | request context bounds gate waiting | contended cancellation and race tests |
| type assertion to `consoleInstrumentMetadata` | expose only `Stocks` to name adapter | fail closed if unavailable | AST B4 |

## State mutations and fallbacks

- Successful first construction caches one account-resolved broker and trimmed reference.
- The cancellable gate single-flights construction so later `resolve` returns that same instance.

## Safety conclusion

- Safe edit boundary: return only the one-method metadata surface; account discovery stays in the command adapter builder.
- High-risk impact: yes for auth/client identity; shared-client/OAuth exchange test proves no second token manager is created.
