# Function Logic Map: `TestClientConfigurationCannotBeMutatedByRetainedOptionPointer`

- Source: `internal/official/client_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| retained client pointer from public Option | exact pointer observed during construction | adversarial test fixture | test fails if replay changes any sealed configuration field |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | custom option did not retain the constructed client | none | `t.Fatal` | this test |
| B2 | base, HTTP client, or account sequence changed after replay | none | `t.Fatalf` | this test |
| B3 | official-origin capability was revoked/replaced | none | `t.Fatal` | this test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| WithBaseURL/WithHTTPClient/WithAccountSeq | replay public options on retained pointer | must be synchronized no-ops after seal | AST + RED/GREEN |

## State mutations and fallbacks

- Test-only adversarial mutations; no network or account-state mutation occurs.

## Safety conclusion

- Safe edit boundary: covers the exact post-New replay vector without exposing production state.
- High-risk impact: validates official FX origin immutability.
