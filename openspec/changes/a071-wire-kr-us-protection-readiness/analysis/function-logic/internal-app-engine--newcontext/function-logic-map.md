# Function Logic Map: `NewContext`

- Source: `internal/app/engine/engine.go` (432-602)
- Revision: current — HEAD `648df8ef`; source_sha256 `d0cb8011c0186beeedface3cb4ba47dc81b3a79202b9ab9650dda99a44a347d7`
- AST evidence: `ast.json` (`tools/logic-map` 추출기 출력)
- Risk scan: `risk-pattern-report.md` (ast-grep, `tools/logic-map/sgconfig.yml`)
- Extractor counts: AST branches 15 · returns 11 · calls 33
- Exact AST return positions: 435:3, 442:3, 472:3, 483:3, 490:3, 500:3, 519:3, 542:4, 562:3, 571:3, 573:2
- 2026-09-25 task 5.1 갱신: 분기 표는 `analysis/harness/51_refresh.py` 가 ast.json 에서 생성함. 아래 산문 절은 a071 저자 기록(2026-08-04) 그대로이며 base 뒤 다른 change 가 이 함수에 더한 동작은 기술하지 않음.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| resolved account/config paths | exact official account and canonical config directory | existing startup assembly | fail closed and close already-owned resources |
| protection readiness | paired read-only provider from pinned manifest; lifecycle assemblies always UNWIRED | buildGateway | never blocks safety runtime for a protection DB |

## Branches and early returns

| Branch | AST anchor | Source text at anchor | Coverage at HEAD (untagged package suite, per-test runs) |
|---|---|---|---|
| B1 | if at 434:2 | `if err != nil {` | entered 3/439 |
| B2 | if at 441:2 | `if err != nil {` | NOT entered 0/439 |
| B3 | if at 445:2 | `if clk == nil {` | entered 26/439 |
| B4 | if at 459:2 | `if opts.Publisher != nil {` | entered 2/439 |
| B5 | if at 467:2 | `if err := recordGateSettings(auditLog, gate, cfg.Engine.Adoption, notifications,` | NOT entered 0/439 |
| B6 | if at 482:2 | `if err != nil {` | entered 1/439 |
| B7 | if at 489:2 | `if err != nil {` | entered 1/439 |
| B8 | if at 498:2 | `if err := bindApplyHooks(jrn); err != nil {` | NOT entered 0/439 |
| B9 | if at 517:2 | `if err != nil {` | NOT entered 0/439 |
| B10 | if at 532:2 | `if gate.Enabled && guardian == nil && !opts.disableProductionGuardian {` | entered 2/439 |
| B11 | if at 534:3 | `if factory == nil {` | entered 1/439 |
| B12 | if at 540:3 | `if err != nil {` | entered 1/439 |
| B13 | if at 560:2 | `if err != nil {` | entered 6/439 |
| B14 | if at 564:2 | `if !automation.Verified {` | entered 48/439 |
| B15 | if at 569:2 | `if err != nil {` | NOT entered 0/439 |

## Author concept rows (2026-08-04 — N 번호는 AST 분기 번호가 아님)

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| N1 | existing order-path/audit/account/journal/gateway/guardian/interlock error | bounded journal cleanup only | startup refusal | engine regressions |
| N2 | successful assembly | context owns journal and safety runtime only | context | production assembly test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `buildGateway` | assemble normal execution stack and read-only protection refusal | no protection mutation path | CodeGraph + AST |

## State mutations and fallbacks

- No lane, gate, autostart or LIVE setting is changed. No protection supervisor exists in Context.

## Safety conclusion

- Safe edit boundary: pass canonical path/pin into read-only provider and preserve journal cleanup order.
- High-risk impact: yes — production engine assembly.
