# Function Logic Map: `TestForbiddenRemoteCapabilitiesHaveNoReachableRoute`

- Source: `internal/httpapi/contract_static_server_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| router | read-only API router with contract fixture | `NewRouter` current implementation | construction error fails the test immediately |
| forbidden paths | live/gate/kill/protection/activation/reconcile/rollback route names | a052 remote-capability prohibition | any reachable mutation surface fails the test |
| methods | GET, POST, PUT, PATCH, DELETE | HTTP mutation/read surface | only 404 or 405 is accepted |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | AST `if` line 127: router construction error | test failure only | aborts setup | `TestForbiddenRemoteCapabilitiesHaveNoReachableRoute` |
| B2 | AST `range` line 130: each forbidden path | no production mutation | iterates full denied route set | `TestForbiddenRemoteCapabilitiesHaveNoReachableRoute` |
| B3 | AST `range` line 140: each common HTTP method | in-memory request only | exercises read and mutation verbs | `TestForbiddenRemoteCapabilitiesHaveNoReachableRoute` |
| B4 | AST `if` line 143: status is neither 404 nor 405 | test error only | reports reachable forbidden capability | `TestForbiddenRemoteCapabilitiesHaveNoReachableRoute` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `NewRouter` | build the real static route table | setup error fails immediately; no retry | AST B1 |
| `httptest.NewRequest` / `NewRecorder` | exercise each path/method without a network listener | deterministic in-memory HTTP; no timeout/retry | AST B2-B3 |
| `router.ServeHTTP` | prove the production router has no matching capability | must return 404/405 | AST B4 |

## State mutations and fallbacks

- Assignments are limited to the router and per-request recorder.
- No command client, live order path, reconciliation resolver, or operating-toggle mutation is invoked.
- The fallback is fail-closed: a newly reachable forbidden route makes the regression test fail.

## Safety conclusion

- Safe edit boundary: static negative contract testing only; no production route or command mutation.
- High-risk impact: indirect safety coverage for remote live/reconciliation authority; no runtime side effect.
