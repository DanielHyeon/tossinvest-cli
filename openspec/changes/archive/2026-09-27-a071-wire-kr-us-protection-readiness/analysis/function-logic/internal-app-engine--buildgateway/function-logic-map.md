# Function Logic Map: `buildGateway`

- Source: `internal/app/engine/gateway.go` (234-355)
- Revision: current — HEAD `648df8ef`; source_sha256 `cf4833845140eba8f7ca16de823d821985c3826c04b2e4f74c8c068ca61ff29d`
- AST evidence: `ast.json` (`tools/logic-map` 추출기 출력)
- Risk scan: `risk-pattern-report.md` (ast-grep, `tools/logic-map/sgconfig.yml`)
- Extractor counts: AST branches 5 · returns 7 · calls 20
- Exact AST return positions: 240:3, 262:3, 270:3, 273:3, 294:3, 318:3, 340:2
- 2026-09-25 task 5.1 갱신: 분기 표는 `analysis/harness/51_refresh.py` 가 ast.json 에서 생성함. 아래 산문 절은 a071 저자 기록(2026-08-04) 그대로이며 base 뒤 다른 change 가 이 함수에 더한 동작은 기술하지 않음.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| journal/trading/official/account/clock/config/pin | exact existing engine dependencies | `gatewayInputs` | typed startup failure, owned journal closed by caller |
| protection assemblies | exact KR and US, both `Wired=false` until committed-fill lifecycle exists | `productionProtectionAssemblies` | readiness remains UNWIRED |

## Branches and early returns

| Branch | AST anchor | Source text at anchor | Coverage at HEAD (untagged package suite, per-test runs) |
|---|---|---|---|
| B1 | if at 239:2 | `if err := checkProjectionWired(in.journal); err != nil {` | NOT entered 0/439 |
| B2 | if at 261:2 | `if err := tracker.Restore(ctx); err != nil {` | NOT entered 0/439 |
| B3 | if at 269:2 | `if err := restoreAlertEntryLatch(ctx, in.journal, entry); err != nil {` | NOT entered 0/439 |
| B4 | if at 293:2 | `if err != nil {` | NOT entered 0/439 |
| B5 | if at 317:2 | `if err != nil {` | NOT entered 0/439 |

## Author concept rows (2026-08-04 — N 번호는 AST 분기 번호가 아님)

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| N1 | projection or reconciliation restore fails | no protection/broker mutation | typed startup error | engine regression |
| N2 | manifest missing/invalid or lifecycle unwired | read-only provider/adapter assembled | Gateway starts, exposure remains refused | paired UNWIRED test |
| N3 | read-only adapter/Gateway construction fails | no official protection adapter exists | typed startup error | engine regression |
| N4 | normal construction | safety loops and execution Gateway assembled | `engineWiring` without protection mutation authority | storage-failure assembly test |

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
