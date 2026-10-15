# a139 변이 원장

- 사본: 연결 워크트리 `dadc386a` (시작 sha 단언 · 깨끗함 확인 · 무변이 대조군 15 시험 GREEN)

| 변이 | 파일 | 시험 | 결과 |
|---|---|---|---|
| N1 나눗수 3 → 2 | `internal/risk/risk_budget.go` | `TestA139PerTradeRiskBudgetIsAThirdRoundedDownToTheMinorUnit` | CAUGHT |
| N2 내림 → 올림 | `internal/risk/risk_budget.go` | `TestA139PerTradeRiskBudgetIsAThirdRoundedDownToTheMinorUnit` | CAUGHT |
| N3 USD 최소 단위 2 → 3 | `internal/risk/risk_budget.go` | `TestA139PerTradeRiskBudgetIsAThirdRoundedDownToTheMinorUnit` | CAUGHT |
| N4 KRW 최소 단위 0 → 2 | `internal/risk/risk_budget.go` | `TestA139PerTradeRiskBudgetIsAThirdRoundedDownToTheMinorUnit` | CAUGHT |
| N5 0 으로 내려가도 허용 | `internal/risk/risk_budget.go` | `TestA139PerTradeRiskBudgetIsAThirdRoundedDownToTheMinorUnit` | CAUGHT |
| N6 모르는 통화를 KRW 처럼 | `internal/risk/risk_budget.go` | `TestA139PerTradeRiskBudgetIsAThirdRoundedDownToTheMinorUnit` | CAUGHT |
| N7 생산 유도 우회(= dailyLoss) | `internal/app/engine/guardian_wiring.go` | `TestA139ProductionPolicyDerivesAThirdOfTheDailyLoss` | CAUGHT |
| N8 유도 실패 시 dailyLoss 로 폴백 | `internal/app/engine/guardian_wiring.go` | `TestA139AnUnderivableBudgetRefusesTheProductionGuardianByName` | CAUGHT |
| N9 기본 정책 값 되돌림 | `internal/risk/input.go` | `TestA139TheDefaultPolicyFollowsTheSameRule` | CAUGHT |
| N10 audit 줄 삭제 | `internal/app/engine/interlock.go` | `TestA139TheInterlockAuditsTheDerivedRiskBudget` | CAUGHT |
| N11 audit 만 dailyLoss | `internal/app/engine/guardian_wiring.go` | `TestA139TheInterlockAuditsTheDerivedRiskBudget` | CAUGHT |
| N12 factory 사전 검사 삭제 | `internal/app/engine/guardian_wiring.go` | `TestA139AnUnderivableBudgetRefusesTheProductionGuardianByName` | CAUGHT |
| N13 버전 v1 되돌림 | `internal/app/engine/guardian_wiring.go` | `TestA139TheProductionPolicyVersionNamesTheNewRule` | CAUGHT |
| N14 유도 불가 사유 삭제 | `internal/app/engine/guardian_wiring.go` | `TestA139AGateOffStartWithoutADailyLossStillStarts` | CAUGHT |
| N15 acceptance 칸 삭제 | `internal/app/engine/interlock.go` | `TestA139TheInterlockAuditsTheDerivedRiskBudget` | CAUGHT |

## 분류

- **15 / 15 CAUGHT**(CAUGHT(build) 0). round1(6f4f7cd3) 은 14/15 — 생존 N12(factory 사전 검사 삭제)는 Validate 가 같은 sentinel 로 대신 거부해 시험이 못 갈랐음 → 사전 검사 문구(「cannot be derived」 · 원인)를 단언(dadc386a) 해 round2 에서 CAUGHT.
