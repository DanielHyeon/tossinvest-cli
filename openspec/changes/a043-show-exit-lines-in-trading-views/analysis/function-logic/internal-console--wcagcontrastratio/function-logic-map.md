# Function Logic Map: `wcagContrastRatio`

- Source: `internal/console/trading_views_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| foreground/background | strict `#RRGGBB` CSS colours | caller's approved UI tokens | delegated validation failure terminates test |
| luminances | finite values in `[0,1]` | `relativeLuminance` | order is normalized before ratio |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | foreground luminance is lower than background | swap local values | continue with lighter value first |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `relativeLuminance` | linearize each CSS colour | test-fatal on malformed colour | AST calls |

## State mutations and fallbacks

- Only local floating-point test values are assigned; no money, quantity, or production state is involved.

## Safety conclusion

- Safe edit boundary: WCAG `(L1+0.05)/(L2+0.05)` test calculation.
- High-risk impact: no; the float risk scan is reviewed-safe UI colour math.
