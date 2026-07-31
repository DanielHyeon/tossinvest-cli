# Function Logic Map: `rowFor`

- Source: `internal/console/portfolio_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| rendered positions HTML and symbol | fixture page; symbol must have `data-symbol` row marker | template output | `t.Fatalf` when missing or malformed |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | symbol marker absent | test process records fatal | no return | rowFor callers with fixture symbols |
| B2 | marker has no enclosing `<tr` | test process records fatal | no return | malformed-row guard |
| B3 | enclosing row has `</tr>` | none | exact single-row fragment | all label/exclusion fixture tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| strings.Index/LastIndex | locate trusted fixture markup boundaries | no timeout/retry; pure string scan | AST B1-B3 + current template |
| testing.T.Fatalf | stop a test whose fixture contract is malformed | test-only failure; no production effect | current function body |

## State mutations and fallbacks

- No product state mutation. Fallback is fail-closed test termination rather than returning another row's markup.

## Safety conclusion

- Safe edit boundary: test helper only; updated for the new one-primary-row markup.
- High-risk impact: no. No runtime, order, policy, journal, or settings code is called.
