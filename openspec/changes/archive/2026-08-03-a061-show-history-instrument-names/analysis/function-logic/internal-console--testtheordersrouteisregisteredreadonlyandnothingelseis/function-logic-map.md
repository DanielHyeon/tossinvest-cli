# Function Logic Map: `TestTheOrdersRouteIsRegisteredReadOnlyAndNothingElseIs`

- Source: `internal/console/orders_static_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| registered route list | every literal route and wrapper chain | `Console.routes` AST extractor | extraction failure fails the test |
| required read-only paths | orders, positions, history, position-management | local exact allowlist | missing wrapper/session or added CSRF fails |
| all other routes | must not carry `readOnly` | extracted route record | unexpected wrapper fails to force explicit review |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | iterate registered routes | test-local map mutation only | continue | this test |
| B2 | route is a required read-only view | mark found and verify wrapper chain | continue after checks | this test |
| B3 | required view lacks readOnly | `t.Errorf` | continue collecting findings | route extractor positive controls |
| B4 | required view is CSRF-gated | `t.Errorf` | continue collecting findings | this test |
| B5 | required view lacks session | `t.Errorf` | continue collecting findings | this test |
| B6 | another route carries readOnly | `t.Errorf` | continue collecting findings | this test |
| B7 | iterate required path results | test-local read only | continue | this test |
| B8 | required route was not registered | `t.Errorf` | finish with failure | this test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `registeredRoutes` | extract actual mux registrations and wrapper chains | test fails if source cannot be interpreted | route extraction tests + AST |
| `t.Errorf` | report every contract drift in one run | no retry | testing package |

## State mutations and fallbacks

- Only the local `found` map mutates. Product state is never reached.
- Adding `/history` is an explicit allowlist expansion paired with the runtime 405 regression test.

## Safety conclusion

- Safe edit boundary: add exactly `/history` to the required read-only paths.
- High-risk impact: no; this test strengthens route method enforcement.
