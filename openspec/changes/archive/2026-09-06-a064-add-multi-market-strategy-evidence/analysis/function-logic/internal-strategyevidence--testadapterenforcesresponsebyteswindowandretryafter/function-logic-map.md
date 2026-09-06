# Function Logic Map: `TestAdapterEnforcesResponseBytesWindowAndRetryAfter`

- Source: `internal/strategyevidence/source_test.go`
- AST evidence: `ast.json` (extracted at the re-baselined comparison base)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `policy.MaxResponseBytes = 4` | re-sealed after the edit | `validSourcePolicy()` | an oversized body must be refused |
| `policy.MaxCalls = 1` | re-sealed after the edit | `validSourcePolicy()` | the second fetch must be rate limited |
| `fakeWaiter` | records the durations it was asked to wait | test double | an unbounded wait fails the test |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | an 8-byte body against a 4-byte limit was not refused | none | `t.Fatalf` | `TestAdapterEnforcesResponseBytesWindowAndRetryAfter` |
| B2 | the first fetch under a one-call window failed | none | `t.Fatal` | `TestAdapterEnforcesResponseBytesWindowAndRetryAfter` |
| B3 | the second fetch was not rate limited | none | `t.Fatalf` | `TestAdapterEnforcesResponseBytesWindowAndRetryAfter` |
| B4 | the 429-then-200 sequence did not take exactly two attempts | none | `t.Fatalf` | `TestAdapterEnforcesResponseBytesWindowAndRetryAfter` |
| B5 | the recorded wait is not the single bounded Retry-After | none | `t.Fatalf` | `TestAdapterEnforcesResponseBytesWindowAndRetryAfter` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `newAdapter` | build the adapter under test | none | AST |
| `policy.officialSeal` | re-stamp the seal after mutating a bound, so the field check is what refuses | none | AST |
| `adapter.Fetch` | drive each enforcement path | typed refusals | AST |

## State mutations and fallbacks

- Test-local state only: a fake transport, a fake waiter and a minted policy value.
- No journal, broker, Guardian or toggle is constructed; the transport is a spy that records calls.

## Safety conclusion

- Not production code. It is mapped because the constructor rename touched its body and the gate compares the frozen base against the working tree.
- The completion pass changed one token in this test: the constructor it calls is now the unexported `newAdapter`. `NewAdapter` was exported, validated nothing and left the shared rate budget nil, so N adapters on one official contract gave N × the policy rate; `NewOfficialAdapter`, which requires a budget, is now the only door in from outside the package (issues.md I17). No condition, assertion or branch in this test changed.
- Purpose: the byte limit, the absolute call window and a bounded Retry-After are each enforced.
