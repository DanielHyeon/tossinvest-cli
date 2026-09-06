# Function Logic Map: `TestAdapterKeepsCredentialOutOfRequestMetadata`

- Source: `internal/strategyevidence/source_test.go`
- AST evidence: `ast.json` (extracted at the re-baselined comparison base)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `validDARTSourcePolicy()` | a minted OpenDART policy | `MintSourcePolicy` | the test fails at Fetch |
| `StaticCredential("top-secret")` | the secret under observation | test double | a leak fails the test |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | the fetch itself failed | none | `t.Fatal` | `TestAdapterKeepsCredentialOutOfRequestMetadata` |
| B2 | the metadata could not be marshalled | none | `t.Fatal` | `TestAdapterKeepsCredentialOutOfRequestMetadata` |
| B3 | the secret or the request identity appears in the encoded metadata | none | `t.Fatalf` | `TestAdapterKeepsCredentialOutOfRequestMetadata` |
| B4 | the transport did not receive the secret | none | `t.Fatal` | `TestAdapterKeepsCredentialOutOfRequestMetadata` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `newAdapter` | build the adapter under test | none | AST |
| `adapter.Fetch` | run one successful fetch | none | AST |
| `json.Marshal` | inspect what the metadata would serialise into a log | none | AST |
| `transport.LastCredential` | prove the secret crossed the transport boundary and only there | none | AST |

## State mutations and fallbacks

- Test-local state only: a fake transport, a fake waiter and a minted policy value.
- No journal, broker, Guardian or toggle is constructed; the transport is a spy that records calls.

## Safety conclusion

- Not production code. It is mapped because the constructor rename touched its body and the gate compares the frozen base against the working tree.
- The completion pass changed one token in this test: the constructor it calls is now the unexported `newAdapter`. `NewAdapter` was exported, validated nothing and left the shared rate budget nil, so N adapters on one official contract gave N × the policy rate; `NewOfficialAdapter`, which requires a budget, is now the only door in from outside the package (issues.md I17). No condition, assertion or branch in this test changed.
- Purpose: the OpenDART secret must reach the transport and nothing else.
