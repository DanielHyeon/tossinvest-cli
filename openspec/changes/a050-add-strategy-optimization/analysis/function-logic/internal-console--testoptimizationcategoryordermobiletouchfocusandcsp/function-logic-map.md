# Function Logic Map: `TestOptimizationCategoryOrderMobileTouchFocusAndCSP`

- Source: `internal/console/optimization_ui_contract_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| authenticated optimization response | HTTP 200 with fixed category order, responsive CSS, focus rules, and CSP | rendered console | test fails on any missing contract |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B6 | status, ordered category loop, responsive/focus marker loop, CSP equality | test assertions only | failures identify the exact missing contract | this test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| dashboard harness / string checks | exercise server-rendered HTML and headers | local httptest; no retry or external IO | AST |

## State mutations and fallbacks

- Mutates test-local harness state only; no LIVE/order authority.

## Safety conclusion

- Safe edit boundary: responsive/accessibility response contract.
- High-risk impact: no.
