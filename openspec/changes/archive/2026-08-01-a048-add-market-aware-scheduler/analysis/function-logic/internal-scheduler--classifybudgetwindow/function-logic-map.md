# Function Logic Map: `classifyBudgetWindow`

- Source: `internal/scheduler/budget.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| trusted state | fixed reset anchor, conservative deadline, exact reset kind | coordinator endpoint generation | absent anchor is initial; conflicting evidence never advances |
| observation | already validated official-parser-equivalent reset evidence | `trustedBudgetWindow` caller | kind/window mismatch returns conflict |
| delta tolerance | inclusive fixed ±1 second around anchor | scheduler constant | comparisons use ordered bounds, never absolute duration negation |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | no trusted reset exists | none | initial | initial observation test |
| B2 | reset kind differs from trusted kind | none | conflict | provenance conflict test |
| B3 | epoch instant exactly equals anchor | none | same | epoch identity test |
| B4 | epoch response arrives at/after prior boundary and names a later reset | none | next | epoch generation test |
| B5 | delta response arrives at/after anchor+tolerance and names a later boundary | none | next | delta boundary test |
| B6 | delta response arrives before anchor+tolerance and reset lies inclusively within anchor±tolerance | none | same | drift/boundary/MinInt-safe tests |
| B7 | no relation is proven | none | conflict | reset conflict tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `time.Time.Before/After/Equal/Add` | compare fixed reset bounds without overflow-prone duration absolute value | pure; tolerance is one second | CodeGraph + AST + boundary tests |

## State mutations and fallbacks

- Pure classification only. Generation mutation remains in `observeLocked` and requires a valid causal cycle even when commitments are already empty.

## Safety conclusion

- Safe edit boundary: reset-window relation classification.
- High-risk impact: yes, because a false next-window result can clear issued authority.
