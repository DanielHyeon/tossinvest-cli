# Function Logic Map: `checkLimits`

- Source: `internal/execgw/guardian.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| persisted decision + mutation plan | exact class, complete supported limit schema, order shape matches decision | journal decision plus actual plan | typed tamper/unset/exceeded refusal |
| account-base q_final envelope | KR/US quote may differ from base; exact opaque evidence is supplied separately at Gateway boundary | persisted FX binding + private strategy plan capability | raw quote/base comparison MUST NOT authorize or reject before converted branch |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B2 | risk-reducing decision has no limits / forged limits | none | allow or tamper refusal | existing reduction tests |
| B3-B5 | missing/unreadable/incomplete exposure limit snapshot | none | limits-unset refusal | existing decision tests |
| B6 | quantity exceeds maximum | none | limit-exceeded refusal | existing decision tests |
| B7 | raw quote notional exceeds base maximum | none | current limit-exceeded refusal | legacy same-currency tests only |
| B8 | plan currency differs from base limit currency | none | current limit-exceeded refusal | legacy mismatch tests |
| added A1 | q_final account-base FX binding is present | none | defer notional/currency checks to opaque-FX Gateway validator; still enforce quantity | paired KR/US strategy Gateway tests |
| added A2 | q_final binding exists but strategy opaque capability is absent | none | typed FX/lease refusal, broker 0 | paired missing-authority tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `DecodeLimits`, `Limits.Validate` | strict schema and complete configured bounds | unknown fields/version fail closed | current source/tests |
| `mutationPlan.notional` | legacy quote-currency float notional | usable only when plan currency equals limits currency | current source/tests |
| future exact account-base validator | compare converted order notional under retained opaque evidence | exact decimal/ceil; no float FX | account-base Gateway matrix |

## State mutations and fallbacks

- Read-only. It must not reconstruct FX from the persisted binding.
- Target edit preserves the legacy path and only defers money checks when a supported account-base
  binding exists; `Gateway.submit` must then require the private opaque strategy capability.

## Safety conclusion

- Safe edit boundary: quantity remains checked here; base notional/currency move to an exact
  opaque-evidence branch, never to raw float conversion.
- High-risk impact: **yes** — final per-order limit check.
