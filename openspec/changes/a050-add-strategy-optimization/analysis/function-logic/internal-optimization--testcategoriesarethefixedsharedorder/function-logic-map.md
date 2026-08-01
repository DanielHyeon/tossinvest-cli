# Function Logic Map: `TestCategoriesAreTheFixedSharedOrder`

- Source: `internal/optimization/registry_test.go`
- AST evidence: `ast.json` at the persisted base revision
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| category list | exact six server-defined categories in shared navigation order | `optimization.Categories` | test fails on count, order, or identity drift |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | list length differs | test failure only | fatal | this test |
| B2 | iterate expected positions | none | compare all category IDs | this test |
| B3 | category at index differs | test failure only | non-fatal error | this test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `optimization.Categories` | obtains server-owned navigation contract | pure/no retry | AST and assertion |

## State mutations and fallbacks

- Test-only assertions; no trading or persisted state and no fallback ordering.

## Safety conclusion

- Safe edit boundary: regression coverage for fixed category order.
- High-risk impact: no.
