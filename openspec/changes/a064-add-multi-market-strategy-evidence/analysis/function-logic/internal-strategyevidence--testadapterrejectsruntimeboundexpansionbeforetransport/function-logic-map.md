# Function Logic Map: `TestAdapterRejectsRuntimeBoundExpansionBeforeTransport`

- Source: `internal/strategyevidence/source_test.go`
- AST evidence: `ast.json` (extracted at the re-baselined comparison base)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| five `FetchRequest` values | each widens or inverts one bound | the test table | the subtest fails |
| `fakeTransport` | records every call | test double | a non-zero count fails the subtest |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | walk the five out-of-bound requests | none | runs a subtest per request | `TestAdapterRejectsRuntimeBoundExpansionBeforeTransport` |
| B2 | the error is not `ErrSourceBoundExceeded`, or the transport was called | none | `t.Fatalf` | `TestAdapterRejectsRuntimeBoundExpansionBeforeTransport` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `newAdapter` | build the adapter under test | none | AST |
| `adapter.Fetch` | the call that must be refused before any request | `ErrSourceBoundExceeded` | AST |
| `transport.Calls` | assert zero outbound requests | none | AST |

## State mutations and fallbacks

- Test-local state only: a fake transport, a fake waiter and a minted policy value.
- No journal, broker, Guardian or toggle is constructed; the transport is a spy that records calls.

## Safety conclusion

- Not production code. It is mapped because the constructor rename touched its body and the gate compares the frozen base against the working tree.
- The completion pass changed one token in this test: the constructor it calls is now the unexported `newAdapter`. `NewAdapter` was exported, validated nothing and left the shared rate budget nil, so N adapters on one official contract gave N × the policy rate; `NewOfficialAdapter`, which requires a budget, is now the only door in from outside the package (issues.md I17). No condition, assertion or branch in this test changed.
- Purpose: a request may not widen a bound the policy set, and the refusal comes before the transport.
