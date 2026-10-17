# Function Logic Map: `StartPositionPolicyCommandServer`

- Source: `internal/app/engine/position_policy_transport.go`
- AST evidence: `ast.json` (post-edit, 50–172, 17 branches). Pre-edit map (HEAD `09465697`, 50–167, 16 branches): `analysis/pre-edit/5.5-socket/internal-app-engine--startpositionpolicycommandserver/`.
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| engineDir | the locked engine directory (0700, no symlink traversal) | `runEngineRun` under the journal flock | error, nothing started |
| commands | the engine's `PositionPolicyCommandService` (production) or a fake | caller | nil → error |
| optional capabilities | `exitQuarantineCommands` (a079), `riskRelaxationCommands` (a066 5.5) | type assertion on `commands` | absent → routes not registered, old route set |

## Branches and early returns (post-edit, from `ast.json`)

| Branch | Line | Condition | Effect |
|---|---|---|---|
| B1 | 52 | `commands == nil` | error |
| B2 | 56 | engine directory blank | error |
| B3 | 59 | engine directory not private | error |
| B4 | 64 / B6 65 / B5 68 | control directory create (exists is fine) | error or `createdControlDir` |
| B7 | 71 / B8 72 | control directory validation fails | remove what this call created, error |
| B9 | 78 | cleanup helper | remove the directory it created |
| B10 | 107 | loopback listen fails | cleanup, error |
| B11 | 112 | token generation fails | close, cleanup, error |
| B12 | 123 | `/v1/health` non-GET | 405 |
| B13 | 130 | `/v1/positions` non-GET | 405 |
| B14 | 135 | `List` error | mapped RPC error |
| B15 | 148 | a079 quarantine capability present | three quarantine routes |
| B16 | 153 | **new (a066 5.5)**: `riskRelaxationCommands` capability present | two relaxation routes (`registerRiskRelaxationRoutes`) |
| B17 | 163 | descriptor publication fails (was B16) | close, cleanup, error |

## Calls and live bindings

| Call | Binding | Note |
|---|---|---|
| `registerRiskRelaxationRoutes(mux, server, token, relaxations)` | `risk_relaxation_transport.go` | same listener, same bearer `auth` wrapper, same descriptor as the policy routes |
| `writePositionPolicyDescriptor` | unchanged | still last, so a client can only read the address once every route exists |

Production binding: `cmd/tossctl/engine.go` `enginePositionPolicyCommandStart` → this function with `*PositionPolicyCommandService`, which implements both optional capabilities.

## State mutations and fallbacks

- The edit registers two handlers on the in-process mux. No file, no journal, no other side effect.
- If the capability is absent, nothing is registered. Clients then get a plain 404, which `positionpolicyrpc` reports as `riskrelaxation.ErrUnwired` (`TestA066AnEngineWithoutTheCapabilityOffersNoRelease`).

## Safety conclusion

- The new routes can only relax through the journal API. That API requires OPERATOR, an approval reference and an audit line before commit, and it is bound to the state the operator saw.
- The engine is not stopped, and no loop is touched. Stop-loss, emergency exit, reconcile and fill detection are not on this path.
- High-risk impact: yes (a relaxation path). Tests: `internal/app/engine/a066_risk_relaxation_test.go`, `cmd/tossctl/engine_risk_relaxation_test.go`.
