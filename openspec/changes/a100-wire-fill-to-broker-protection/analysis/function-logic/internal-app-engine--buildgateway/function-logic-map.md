# Function Logic Map: `buildGateway`

> **a100 R0 재동결(2026-10-04) — 채택 번들.** a100 편집 전 번들(base 882a0b49 기준)은 분기 구조가 바뀌어 낡았다. 이 함수를 마지막으로 바꾼 a092(아카이브 2026-09-30)의 번들을 그대로 가져온다 — 현재 소스와 SHA-256 일치(3d6dabe9…, 234-359). 분기 · 시험 인용은 a092 편집 뒤 기준이다. a100 의 T 로트가 이 함수를 편집할 때 이 번들 위에 Branch Test Map 을 덧붙인다.

- Source: `internal/app/engine/gateway.go`
- AST evidence: `ast.json` — **편집 뒤**, :234–359, 분기 6 · 반환 8 · 호출 22, source_sha256 `3d6dabe90a54…`, 추출 커밋 `2714e393`. 편집 전 번들은 `analysis/pre-edit/unit4/`.
- Risk scan: `risk-pattern-report.md`
- 편집(착지 단위 ④ `2714e393`): 알림 래치 복원 뒤 `bindOperatingModeProjection`(새 파일) 호출 — 분기 하나 추가(투영기 묶기 실패 = 배선 결함 → 기동 거부). 복원 실패는 그 함수 안에서 래치 + 계속.
- 재번호: difflib 정렬(편집 전 `22db26e7` 소스 대비): 새 분기 B4(:273); B4~B5 → B5~B6(같은 분기, 줄 이동).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 | 편집 전 계약 + 위 편집 | 호출자 | 아래 분기 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if err := checkProjectionWired(in.journal); err != nil` (:239) | — | — | (미실행) |
| B2 | `if err := tracker.Restore(ctx); err != nil` (:261) | — | — | (미실행) |
| B3 | `if err := restoreAlertEntryLatch(ctx, in.journal, entry); err != nil` (:269) | — | — | (미실행) |
| B4 | `if err := bindOperatingModeProjection(ctx, in.journal, entry, in.accountRef, in.` (:273) | — | — | (미실행) |
| B5 | `if err != nil` (:297) | — | — | (미실행) |
| B6 | `if err != nil` (:321) | — | — | (미실행) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `bindOperatingModeProjection`, `checkProjectionWired`, `context.Background`, `entry.SetAuthorityRefresh`, `execgw.New`, `execgw.NewEntryGate`, `fmt.Errorf`, `in.official.BaseURL`, `newNotifier`, `newRetrier`, `productionProtectionAssemblies`, `productionProtectionDigests`, `protection.NewPairedReadinessAdapter`, `protectionreadiness.NewProductionProvider`, `provider.RuntimeContracts`, `restoreAlertEntryLatch` | 편집 뒤 호출 | 위 편집 참조 | AST |

## State mutations and fallbacks

- 위 편집 요약이 전부 — 나머지는 편집 전과 같음(편집 전 번들 대조).

## Safety conclusion

- Safe edit boundary: 위 편집만.
- High-risk impact: High-risk(엔진 조립).
