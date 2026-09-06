# Function Logic Map: `NewOfficialAdapter`

- Source: `internal/strategyevidence/official_source.go`
- AST evidence: `ast.json` (extracted at the re-baselined comparison base)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `policy` | a sealed policy that passes `validateOfficial` | `MintSourcePolicy` | the policy's typed error; no adapter is built |
| `transport`, `budget` | both non-nil | the caller | `ErrSourceDisabled` |
| `credentials` | non-nil when the policy requires a credential | the caller | `ErrSourceCredential` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | the sealed policy does not validate | none | the policy's typed error | `TestOfficialAdaptersDoNotCommitIncompleteOrUntrustedBatches` |
| B2 | the transport or the shared rate budget is absent | none | `ErrSourceDisabled` | `TestOfficialAdaptersDoNotCommitIncompleteOrUntrustedBatches` |
| B3 | a credential-requiring policy has no provider | none | `ErrSourceCredential` | `TestOfficialAdaptersDoNotCommitIncompleteOrUntrustedBatches` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `policy.validateOfficial` | refuse an unsealed or out-of-contract policy before an adapter exists | pure | AST + `TestOfficialAdaptersDoNotCommitIncompleteOrUntrustedBatches` |
| `newAdapter` | build the base adapter, then bind it to the caller's shared budget | allocation only | AST + `TestSharedRateBudgetCapsEveryAdapterOnOneContract` |

## State mutations and fallbacks

- Allocation only; the base adapter's `shared` budget and `budgetKey` are set here and nowhere else.
- The budget key is `authority + "/" + contractID`, so every adapter on one official contract shares one window.

## Safety conclusion

- This is now the only way to obtain an `*Adapter` from outside the package: the completion pass unexported `NewAdapter`, whose adapters carried no shared budget (issues.md I17).
- The change to this function is that rename and nothing else; its three refusals are unchanged.
