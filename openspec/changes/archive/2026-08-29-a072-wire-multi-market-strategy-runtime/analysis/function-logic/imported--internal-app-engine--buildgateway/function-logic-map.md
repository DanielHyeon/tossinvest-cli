# Function Logic Map: `buildGateway`

- Source: `internal/app/engine/gateway.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| journal/trading/official/account/clock/config/pin | exact existing engine dependencies | `gatewayInputs` | typed startup failure, owned journal closed by caller |
| protection assemblies | exact KR and US, both `Wired=false` until committed-fill lifecycle exists | `productionProtectionAssemblies` | readiness remains UNWIRED |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | projection or reconciliation restore fails | no protection/broker mutation | typed startup error | engine regression |
| B2 | manifest missing/invalid or lifecycle unwired | read-only provider/adapter assembled | Gateway starts, exposure remains refused | paired UNWIRED test |
| B3 | read-only adapter/Gateway construction fails | no official protection adapter exists | typed startup error | engine regression |
| B4 | normal construction | safety loops and execution Gateway assembled | `engineWiring` without protection mutation authority | storage-failure assembly test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `NewProductionProvider`, `NewPairedReadinessAdapter` | publish exact read-only KR/US refusals | fail closed; no controller/gateway factory | CodeGraph + AST |
| `execgw.New` | construct sole normal-order mutation path | caller handles error | CodeGraph + AST |

## State mutations and fallbacks

- No official protection gateway, controller, protection DB, or arbitrary factory is constructed.

## Safety conclusion

- Safe edit boundary: read-only readiness can refuse entries but cannot create protection authority.
- High-risk impact: yes; production authority deliberately reduced to UNWIRED.
