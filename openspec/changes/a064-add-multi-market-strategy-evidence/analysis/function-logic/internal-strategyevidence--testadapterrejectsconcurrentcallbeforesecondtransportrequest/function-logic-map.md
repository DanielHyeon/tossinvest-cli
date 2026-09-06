# Function Logic Map: `TestAdapterRejectsConcurrentCallBeforeSecondTransportRequest`

- Source: `internal/strategyevidence/source_test.go`
- AST evidence: `ast.json` (extracted at the re-baselined comparison base)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `policy.MaxConcurrency = 1` | one in-flight request | `validSourcePolicy()` | a second concurrent call must be refused |
| `blockingTransport` | holds the first request open until released | test double | a timeout fails the test |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | wait for the first transport call to start, or time out after a second | none | `t.Fatal` on timeout | `TestAdapterRejectsConcurrentCallBeforeSecondTransportRequest` |
| B2 | the second concurrent fetch was not rate limited | none | `t.Fatalf` | `TestAdapterRejectsConcurrentCallBeforeSecondTransportRequest` |
| B3 | the transport saw more or fewer than one call | none | `t.Fatalf` | `TestAdapterRejectsConcurrentCallBeforeSecondTransportRequest` |
| B4 | the released first fetch returned an error | none | `t.Fatal` | `TestAdapterRejectsConcurrentCallBeforeSecondTransportRequest` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `newAdapter` | build the adapter under test | none | AST |
| `adapter.Fetch` | one held open, one refused | `ErrSourceRateLimited` for the second | AST |
| `transport.Calls` | assert the refused fetch never reached the transport | none | AST |

## State mutations and fallbacks

- Test-local state only: a fake transport, a fake waiter and a minted policy value.
- No journal, broker, Guardian or toggle is constructed; the transport is a spy that records calls.

## Safety conclusion

- Not production code. It is mapped because the constructor rename touched its body and the gate compares the frozen base against the working tree.
- The completion pass changed one token in this test: the constructor it calls is now the unexported `newAdapter`. `NewAdapter` was exported, validated nothing and left the shared rate budget nil, so N adapters on one official contract gave N × the policy rate; `NewOfficialAdapter`, which requires a budget, is now the only door in from outside the package (issues.md I17). No condition, assertion or branch in this test changed.
- Purpose: a second concurrent fetch is refused before it reaches the transport.
