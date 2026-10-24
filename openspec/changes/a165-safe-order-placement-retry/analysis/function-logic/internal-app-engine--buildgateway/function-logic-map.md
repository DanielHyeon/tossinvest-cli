# Function Logic Map: `buildGateway`

- Source: `internal/app/engine/gateway.go` (245-370)
- Qualified function: `buildGateway`
- Revision: `current` (a165 base `34e12967e`, `source_sha256` 497ec967c114b5e7…)
- AST evidence: `ast.json` — AST branches 6, 반환 8, 호출 22
- Risk scan: `risk-pattern-report.md`
- 실행 칸 측정: 편집 전 기준선 `go test -count=1 -coverprofile` (execgw·official·trading·flatten 패키지 전체, journal `-run 'Classify|SendTracker|Dispatch'`, app/engine `-run 'Clear|Cancel'` — journal·engine 은 부분 실행이라 count=0 이 「미실행」을 뜻하지 않을 수 있음, cmd/tossctl 은 미측정) — 분기 줄에서 시작하는 본문 블록의 실행 수. 시험 이름별 귀속은 GREEN 재추출에서 함.

## 역할 (a165)

엔진 게이트웨이 배선(execgw.New 호출 311). a165 는 `CancelRefusal: execgw.EngineCancelRefusal()` 한 필드를 더한다 — 취소 409 확정 미수락, 같은 주기 재시도 0.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| in | gatewayInputs | 엔진 기동 | B1–B6 오류 반환 |

## Branches and early returns

AST 열거 순서 그대로. Condition 칸은 AST 좌표 줄의 원문.

Exact AST return positions: 251:3, 273:3, 281:3, 285:3, 288:3, 309:3, 333:3, 355:2

| Branch | Coordinate | Condition | Mutation/side effect | Return/error | Executed (baseline) |
|---|---|---|---|---|---|
| B1 | `if at 250:2` | if err := checkProjectionWired(in.journal); err != nil { | 아래 「State mutations」 | 소스 줄 250 이하 | 미실행 — 패키지 스위트 count=0 |
| B2 | `if at 272:2` | if err := tracker.Restore(ctx); err != nil { | 아래 「State mutations」 | 소스 줄 272 이하 | 미실행 — 패키지 스위트 count=0 |
| B3 | `if at 280:2` | if err := restoreAlertEntryLatch(ctx, in.journal, entry); err != nil { | 아래 「State mutations」 | 소스 줄 280 이하 | 미실행 — 패키지 스위트 count=0 |
| B4 | `if at 284:2` | if err := bindOperatingModeProjection(ctx, in.journal, entry, in.accountRef, in.logger); err != nil { | 아래 「State mutations」 | 소스 줄 284 이하 | 미실행 — 패키지 스위트 count=0 |
| B5 | `if at 308:2` | if err != nil { | 아래 「State mutations」 | 소스 줄 308 이하 | 미실행 — 패키지 스위트 count=0 |
| B6 | `if at 332:2` | if err != nil { | 아래 「State mutations」 | 소스 줄 332 이하 | 미실행 — 패키지 스위트 count=0 |

## Calls and live bindings

`execgw.New` 등.

AST 호출 이름 전체: `bindOperatingModeProjection`, `checkProjectionWired`, `context.Background`, `entry.SetAuthorityRefresh`, `execgw.New`, `execgw.NewEntryGate`, `fmt.Errorf`, `in.official.BaseURL`, `newNotifier`, `newRetrier`, `productionProtectionAssemblies`, `productionProtectionDigests`, `protection.NewPairedReadinessAdapter`, `protectionreadiness.NewProductionProvider`, `provider.RuntimeContracts`, `restoreAlertEntryLatch`, `tracker.Refresh`, `tracker.Restore`.

## State mutations and fallbacks

- 엔진 배선.

## Safety conclusion

- 필드 1개 추가. 되돌리기 = 그 줄 삭제.
