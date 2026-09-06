# Function Logic Map: `TestIncompleteOrUnverifiedPolicyMakesZeroCalls`

- Source: `internal/strategyevidence/source_test.go`
- AST evidence: `ast.json` (extracted at the re-baselined comparison base)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| mutated `SourcePolicy` | one field broken per case | `validSourcePolicy()` | the subtest fails |
| `fakeTransport` | records every call | test double | a non-zero count fails the subtest |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | walk the four mutation cases | none | runs a subtest per case | `TestIncompleteOrUnverifiedPolicyMakesZeroCalls` |
| B2 | the returned error is not the expected sentinel | none | `t.Fatalf` | `TestIncompleteOrUnverifiedPolicyMakesZeroCalls` |
| B3 | the transport was called at all | none | `t.Fatalf` | `TestIncompleteOrUnverifiedPolicyMakesZeroCalls` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `newAdapter` | build the adapter under test | none | AST |
| `adapter.Fetch` | the call that must be refused before any request | returns the policy's typed error | AST |
| `transport.Calls` | assert zero outbound requests | none | AST |

## State mutations and fallbacks

- Test-local state only: a fake transport, a fake waiter and a minted policy value.
- No journal, broker, Guardian or toggle is constructed; the transport is a spy that records calls.

## Safety conclusion

- Not production code. It is mapped because the constructor rename touched its body and the gate compares the frozen base against the working tree.
- The completion pass changed one token in this test: the constructor it calls is now the unexported `newAdapter`. `NewAdapter` was exported, validated nothing and left the shared rate budget nil, so N adapters on one official contract gave N × the policy rate; `NewOfficialAdapter`, which requires a budget, is now the only door in from outside the package (issues.md I17). No condition, assertion or branch in this test changed.
- Purpose: an incomplete or unverified policy must refuse with zero transport calls.
