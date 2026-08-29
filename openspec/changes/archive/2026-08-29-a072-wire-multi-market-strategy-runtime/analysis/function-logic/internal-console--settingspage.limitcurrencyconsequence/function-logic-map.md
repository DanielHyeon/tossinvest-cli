# Function Logic Map: `settingsPage.LimitCurrencyConsequence`

- Source: `internal/console/settings_card.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| settings page gate | configured account-base currency or absent | loaded automation gate config | absence explicitly says both markets remain unready; no default |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | trimmed currency is non-empty | none | delegate exact paired FX explanation | settings-card test |
| fallthrough | currency absent | none | account base unconfigured; no Guardian permission | unset test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `currencyConsequence` | render configured account-base/paired FX contract | pure; unknown stays unsupported | AST + sibling map |

## State mutations and fallbacks

- Read-only view projection. No setting, activation, approval, toggle or order mutation.

## Safety conclusion

- Safe edit boundary: update the absence explanation only; preserve current configured/absent branch.
- High-risk impact: yes — the always-visible notice is an operator safety statement.
