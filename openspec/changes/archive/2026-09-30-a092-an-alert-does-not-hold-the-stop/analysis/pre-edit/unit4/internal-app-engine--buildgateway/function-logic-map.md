# Function Logic Map: `buildGateway`

- Source: `internal/app/engine/gateway.go`
- AST evidence: `ast.json` — **편집 전**, :234–355, 분기 5 · 반환 7 · 호출 20, source_sha256 `cf4833845140…`, 추출 HEAD `22db26e7`.
- Risk scan: `risk-pattern-report.md`
- 편집 목적(착지 단위 ④): 게이트 생성 직후 `SetModeProjector(entry)` → `RestoreOperatingModeProjection`(첫 진입 점검보다 먼저 — 루프는 이 함수 반환 뒤에 뜬다). 복원 실패는 기동 거부가 아니라 모드 사유 래치 + 기동 계속(C15 (ㄴ) — 손절이 산다). `SetModeProjector` 오류(재묶기)는 배선 결함이라 기동 거부.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 | 편집 전 계약 그대로 | 호출자 | 아래 분기 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if err := checkProjectionWired(in.journal); err != nil` (:239) | — | — | (미실행) |
| B2 | `if err := tracker.Restore(ctx); err != nil` (:261) | — | — | (미실행) |
| B3 | `if err := restoreAlertEntryLatch(ctx, in.journal, entry); err != nil` (:269) | — | — | (미실행) |
| B4 | `if err != nil` (:293) | — | — | (미실행) |
| B5 | `if err != nil` (:317) | — | — | (미실행) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `checkProjectionWired`, `context.Background`, `entry.SetAuthorityRefresh`, `execgw.New`, `execgw.NewEntryGate`, `fmt.Errorf`, `in.official.BaseURL`, `newNotifier`, `newRetrier`, `productionProtectionAssemblies`, `productionProtectionDigests`, `protection.NewPairedReadinessAdapter`, `protectionreadiness.NewProductionProvider`, `provider.RuntimeContracts` | 편집 전 호출 | 편집 전 계약 | AST |

## State mutations and fallbacks

- 편집 전 동작 그대로 — 편집은 위 목적의 범위.

## Safety conclusion

- Safe edit boundary: 위 목적에 적힌 것만.
- High-risk impact: High-risk(엔진 조립 · 진입 게이트).
