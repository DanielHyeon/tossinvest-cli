# Branch Test Map: `restoreAlertEntryLatch`

- Source: `internal/app/engine/gateway.go` (153-168); current (a124 는 편집하지 않음 — 대조 번들)
- 시험 칸은 측정값: `analysis/harness/branch_coverage.py` 가 `github.com/JungHoonGhae/tossinvest-cli/internal/app/engine` 의 시험 498 개를 하나씩 돌린 커버 프로필에서 그 분기 본문 블록을 실행한 시험.

| Branch | AST anchor | Scenario (source text) | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 155:2 | `if err != nil {` | 없음 | 대조(편집 없음) | 블록 155.16-157.3 을 어느 시험도 실행하지 않음 |
| B2 | if at 158:2 | `if undelivered <= 0 {` | `TestACrashBetweenArmingAndFillingResumesWithoutASecondOrder` · `TestAFailedSweepDoesNotPropagate` · `TestAGuardianSizedAgainstOtherNumbersIsRefused` (+72) | 대조(편집 없음) | 블록 158.22-160.3 을 시험 75 개가 실행, 전부 PASS |
