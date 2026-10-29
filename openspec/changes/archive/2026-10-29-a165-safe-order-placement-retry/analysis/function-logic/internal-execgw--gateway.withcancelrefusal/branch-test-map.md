# Branch Test Map: `Gateway.withCancelRefusal`

- Source: `internal/execgw/cancel_refusal.go` (44-106), D5 4판 재전송 중 실시간 상한 GREEN AST
- AST 분기 13. 새/변경 대상은 window context 생성과 취소 갈래뿐이며, 기존 정책·latch·정착 경계는 유지함.

| 분기 | 시나리오 | RED | GREEN/회귀 |
|---|---|---|---|
| B1 | nil 정책·place·amend | 기존 RED | `TestA165NilPolicyAndNonCancelKeepTheOriginalOutcome` |
| B2 | 비 AP 또는 비 AMBIGUOUS | 기존 RED | `TestA165CancelRefusalDoesNotOverrideNonAmbiguousDispatch` |
| B3 | engine AP 409 | 기존 RED | 같은 주기 지연 0·호출 1 시험 |
| B4 | 무효 flatten 정책 | 기존 RED | 정책 정확값 시험 |
| B5 | `for at 70:3` retry loop | 기존 RED | `TestA165FlattenExhaustionKeepsTheFirstAmbiguousOutcome` |
| B6 | `if at 72:4` 남은 MaxWait로 대기 절삭 | 기존 RED | `TestA165FlattenRetriesWithinCumulativeBudget` |
| B7 | `if at 75:4` 대기 0 이하 | 기존 RED | `TestA165FlattenRetriesWithinCumulativeBudget` |
| B8 | `if at 78:4` sleep 전 Clock 창 | 기존 RED | `TestA165WindowStartsBeforeFirstDispatch` |
| B9 | `if at 81:4` sleep context 취소 | D5 RED | `TestA165RealTimeoutKeepsFirstAmbiguousAsCancelled` |
| 생성 식 | 직접 wall-clock 판정 금지 | D5 RED | `TestA165CancelRefusalUsesOnlyInjectedClockForJudgement` |
| B10 | `if at 85:4` sleep 뒤 Clock 창 | 기존 RED | `TestA165WindowStartsBeforeFirstDispatch` |
| B11 | `if at 92:4` retry dispatch 중 context 종료 | D5 재전송 상한 RED | `TestA165RetryDeadlineKeepsFirstAmbiguousAsCancelled` |
| B12 | `if at 96:4` retry 결과 | 기존 RED | `TestA165FlattenRetryNotSentKeepsTheFirstAmbiguousOutcome` |
| B13 | `if at 97:5` NOT_SENT | 기존 RED | `TestA165FlattenRetryNotSentKeepsTheFirstAmbiguousOutcome` |

변이 대조군: `WithTimeout` 부모를 `context.Background()`로 분리하면 caller timeout 동작/구조 핀이 실패해야 하며, `time.Now` 기반 deadline을 되살리면 구조 핀이 실패해야 함.
