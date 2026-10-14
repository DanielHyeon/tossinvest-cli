# Branch Test Map: `RiskGuardian.dailyLossFloor`

- Source: `internal/execgw/a137_daily_loss.go` (145-165, 편집 뒤). 편집 전 분기수 3 → 편집 뒤 4. RED 는 `analysis/red/`, GREEN 은 `analysis/verify/receipts.md`.

| Branch | Anchor | Scenario | Test | a137 |
|---|---|---|---|---|
| B1 | if 147 | 기존 동작 | `TestA137AMeasurementFailureNeverFallsBackToTheCaller` · `TestA137APrecheckMeasurementMustBeThisGuardiansFreshTodaysOne` | 무편집 분기 |
| B2 | if 152 | 기존 동작 | `TestA137AnInvalidCallerValueIsNotReplaced` | 무편집 분기 |
| B3 | if 156 | a137 편집 | 분기 몸체 도달 불가(측정값 · 호출자 값 모두 정규 decimal 로 이미 검증됨) — 방어 갈래, 변이 원장에 동등으로 기록 | 편집 분기 |
| B4 | if 161 | 기존 동작 | `TestA137TheLedgerLossRefusesAZeroCallerAndEscalates` · `TestA137ALargerValidCallerValueWins` | 무편집 분기 |
