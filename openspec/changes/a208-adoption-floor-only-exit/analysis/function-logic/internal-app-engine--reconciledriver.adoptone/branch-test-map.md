# Branch Test Map: `ReconcileDriver.adoptOne`

편집 뒤 좌표(a208 3.3 재추출 · AST 3 분기, `internal/app/engine/adoption.go` 373–432). 「미측정」 = 덮는 시험을 이름으로 특정하지 않았음(실행 커버리지 미측정).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 합성 손절 유도 실패 → 편입 연기 | 기존 편입 시험군(미측정) | n/a | 미측정(덮는 시험 미특정) |
| B2 | 편입 거절 → 연기 | 기존 편입 시험군(미측정) | n/a | 미측정(덮는 시험 미특정) |
| B3 | exit state 열기 실패 → 기록 유지, 다음 주기 완결 | 미측정(원장 수준 크래시 계약은 `TestCrashBetweenAdoptionAndExitStateIsCompleted`) | n/a | 미측정(덮는 시험 미특정) |

a208 새 동작(분기 밖 값): ON 정책 선택 `TestA208FloorOnlyAdoptionRecordsTheRunnerCopy` · `TestA208FloorOnlyAdoptionReplacesRatchet`, OFF 바이트 `TestA208OffAdoptionIsUnchanged`, ON 알림 `TestA208FloorOnlyAlertSaysNoPartialTakeProfit`, 스냅숏 유지 `TestA208TurningTheToggleOffKeepsTheAdoptedPolicy` — RED 관측 `analysis/tdd/red.txt`(T3 는 base 에서도 통과가 기대값).
