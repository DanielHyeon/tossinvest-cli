# Function Logic Map: `Console.handlePositions`

- Source: `internal/console/portfolio_pages.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Request context | authenticated GET `/positions` | route + request context | rendering remains read-only; no POST/mutation |
| Journal rows | readable/unknown, KR/US, lifecycle generation >=1 when known | journal read model | unavailable or mismatched evidence is unknown, never actionable |
| Runtime settings | effective-known or unavailable | engine-owned `ManagementRuntime` | desired/default never substitutes for effective |
| Position lifecycle | MANAGED/RELEASED plus generation/version | `PositionPolicies.List` | absent lifecycle remains legacy generation 1 semantics |

## Branches and early returns

| Branch | AST kind/location | Condition/control path | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|---|
| B1 | `if` line 49 | policy seam exists | local runtime attempt flag only | continues | runtime known/unavailable |
| B2 | `if` line 55 | lifecycle list succeeds | builds local map | list failure stays unknown | lifecycle tests |
| B3 | `range` line 57 | each lifecycle state | map assignment | continues | multiple positions |
| B4 | `if` line 62 | settings seam exists | reads desired display facts only | continues | settings present/absent |
| B5 | `if` line 63 | desired settings load succeeds | stamps include/exclude display | load error does not become effective | failure test |
| B6 | `range` line 66 | each row | display-only designation assignment | continues | KR/US rows |
| B7 | `if` line 72 | runtime call was attempted | projects management truth | unavailable remains typed unknown | runtime failure test |
| B8 | `range` line 73 | each row | local row projection | continues | mixed markets |
| B9 | `if` line 77 | row exists in journal | joins lifecycle by exact position ID | no symbol/time fuzzy join | join tests |
| B10 | `if` line 79 | exact lifecycle missing | marks journal truth unknown | continues | missing lifecycle test |
| B11 | `else` line 81 | lifecycle found | stamps status and generation | continues | managed/released tests |
| B12 | `if` line 90 | covering reconcile block exists | display-only sanitized block | continues | account/market/symbol block tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `c.positions` | read broker+journal projection | existing cache/journal unknown contract | CodeGraph + tests |
| `PositionPolicies.Runtime/List` | engine effective settings and lifecycle | errors remain unknown; no retry/mutation | CodeGraph + a052/a053 tests |
| `Settings.Load` | desired display only | errors ignored for effective truth | current AST |
| `positionpolicy.ProjectManagement` | stable candidate/block status | pure projector | package tests |
| `attachPositionExitLines` | attach canonical/reference views | pure row mutation only | CodeGraph impact + focused tests |
| `c.render` | HTML response | template auto-escaping | httptest contracts |

## State mutations and fallbacks

- Only local page/row projections are mutated; no journal/config/order/reconcile state is written.
- Missing runtime or lifecycle truth fails closed and does not expose a candidate percentage.
- Exact position ID and lifecycle generation are the only accepted joins for exit evidence.

## Safety conclusion

- Safe edit boundary: stamp read-only lifecycle generation/effective settings and call the shared reference projector.
- High-risk impact: no direct order side effect, but exit price evidence is safety-sensitive and must fail closed on mismatch.
