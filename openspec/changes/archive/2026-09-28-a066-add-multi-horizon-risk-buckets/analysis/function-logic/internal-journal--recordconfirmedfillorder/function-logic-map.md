# Function Logic Map: `recordConfirmedFillOrder`

- Source: `internal/journal/fills_test.go`
- AST evidence: `ast.json` (extracted Wave 2A 2026-09-25 at HEAD `648df8ef`; 105–109, 0 branches, 1 call)
- Risk scan: `risk-pattern-report.md`

## Why this bundle is new in Wave 2A

Test helper that existed at base `23794f86`. a066 `c60fee07` (Wave 1C) turned its body into a one-line delegation
to the new `recordConfirmedFillOrderScopeQuantity` so risk-fill tests can register orders with a quantity other
than 10. No bundle was written in Wave 1C; the Wave 2A attribution found the gap.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| test journal, intent/attempt/order ids | fresh per test | calling test | helper fails the test through `t.Fatal` inside the delegate |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | branchless: delegates with the default US/AAPL/BUY scope and quantity `"10"` | whatever the delegate writes (a confirmed place attempt for the order) | none | every caller in `internal/journal` (19 call sites at HEAD) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `recordConfirmedFillOrderScopeQuantity` | shared body since `c60fee07` | test failure on any journal error | the single entry in `ast.json` calls |

## State mutations and fallbacks

- Test-only; no production path.

## Safety conclusion

- Safe edit boundary: delegation only; default quantity stays `"10"`, so the pre-a066 callers see the same fixture.
- High-risk impact: no — test helper.
