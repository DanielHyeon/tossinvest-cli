# Function Logic Map: `TestOptimizationSaveRequiresSessionAndCSRF`

- Source: `internal/console/optimization_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| preview POST | authenticated session, CSRF, version/category/key/finite option | console route and lifecycle commander | refusal before command seam |

## Branches and early returns

| Branch group | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B3 | no-CSRF refusal, seam counters, valid preview | preview candidate only | assertion failure | this test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| authenticated harness | exercises actual middleware and route | deterministic local test | AST |

## State mutations and fallbacks

- The legacy config save count must remain zero; valid POST reaches preview only.

## Safety conclusion

- Safe edit boundary: verifies middleware and no legacy bypass.
- High-risk impact: yes; test pins fail-closed command admission.
