# Function Logic Map: `currencyConsequence`

- Source: `internal/console/settings_limits.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| configured `limit_currency` | trimmed account-base currency; currently KRW or USD | `engine.automation_gate.limit_currency` | unknown value is named but grants no market authority |
| paired FX readiness | KR and US are independent evidence scopes in one delivery wave | a072 account-base Guardian contract | text must not claim one market must stabilize before the other |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | switch on canonical currency | none | KRW, USD or unknown explanation | paired console test |
| B2 | KRW account base | none | KR identity + US official USD→KRW requirements | paired console test |
| B3 | USD account base | none | US identity + KR official KRW→USD requirements | paired console test |
| default | another configured currency | none | unsupported pair/readiness warning | unknown-currency test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.TrimSpace`/`ToUpper` | canonical display branch only | pure; no fallback authority | AST |

## State mutations and fallbacks

- Pure presentation helper. It writes no config, activation, toggle, account or order.
- Existing wording that one currency closes the peer market becomes false once account-base FX is wired;
  the replacement names per-market FX readiness and same-wave delivery.

## Safety conclusion

- Safe edit boundary: change explanatory text only; preserve the three-way branch and all save semantics.
- High-risk impact: yes — misleading currency text can cause an operator to believe KR-first/US-later is required.
