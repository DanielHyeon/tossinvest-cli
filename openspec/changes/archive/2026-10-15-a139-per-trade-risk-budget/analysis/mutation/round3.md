# a139 변이 원장

- 사본: 연결 워크트리 `e7302fdb` (시작 sha 단언 · 깨끗함 확인 · 무변이 대조군 19 시험 GREEN)

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
| N16 생성자에서 예산 되돌림(구현 리뷰 P1-1) | `internal/app/engine/guardian_wiring.go` | `TestA139TheProductionGuardianSizesWithAThirdOfTheDailyLoss` | CAUGHT |
| N17 유도 불가 audit 줄 건너뜀(구현 리뷰 P3-1) | `internal/app/engine/interlock.go` | `TestA139AGateOffStartWithoutADailyLossStillStarts` | CAUGHT |
| N18 account-base 경로가 일일 손실 전부 | `internal/risk/account_base.go` | `TestA139TheAccountBasePathSizesWithAThirdOfTheDailyLoss` | CAUGHT |
| N19 매도가 예산을 읽음 | `internal/risk/chain.go` | `TestA139ASellNeverReadsTheRiskBudget` | CAUGHT |

## 분류

- **19 / 19 CAUGHT**(CAUGHT(build) 0). round2(dadc386a) 15/15 에 구현 리뷰가 찾은 생존 자리 넷(N16 생성자 되돌림 · N17 audit 줄 건너뜀 · N18 account-base 일일 손실 전부 · N19 매도가 예산 읽음)을 더함 — 리뷰 사본에서 N16 · N17 은 SURVIVED 였고 새 시험으로 CAUGHT.
