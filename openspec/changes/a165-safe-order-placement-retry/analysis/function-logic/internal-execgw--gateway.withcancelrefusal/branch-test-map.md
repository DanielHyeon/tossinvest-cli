# Branch Test Map: `Gateway.withCancelRefusal`

- Source: `internal/execgw/cancel_refusal.go` (44-102), pre-D5 4판 정정 AST
- AST 분기 12. 새/변경 대상은 window context 생성과 취소 갈래뿐이며, 기존 정책·latch·정착 경계는 유지함.

| 분기 | 시나리오 | RED | GREEN/회귀 |
|---|---|---|---|
| B1 | nil 정책·place·amend | 기존 RED | `TestA165NilPolicyAndNonCancelKeepTheOriginalOutcome` |
| B2 | 비 AP 또는 비 AMBIGUOUS | 기존 RED | `TestA165CancelRefusalDoesNotOverrideNonAmbiguousDispatch` |
| B3 | engine AP 409 | 기존 RED | 같은 주기 지연 0·호출 1 시험 |
| B4 | 무효 flatten 정책 | 기존 RED | 정책 정확값 시험 |
| B5-B9 | flatten 대기·누적 상한·Clock 창 | 기존 RED | cumulative/window 경계 시험 |
| B8 | 호출자 취소 | 기존 RED | `TestA165CallerCancellationStopsRetry` |
| B8 | Fake Clock 정지 중 실시간 상한 만료 | D5 RED | `TestA165RealTimeoutKeepsFirstAmbiguousAsCancelled` |
| 생성 식 | 직접 wall-clock 판정 금지 | D5 RED | `TestA165CancelRefusalUsesOnlyInjectedClockForJudgement` |
| B10-B12 | retry 결과 분기 | 기존 RED | NOT_SENT·ACK·exhaustion 시험 |

변이 대조군: `WithTimeout` 부모를 `context.Background()`로 분리하면 caller timeout 동작/구조 핀이 실패해야 하며, `time.Now` 기반 deadline을 되살리면 구조 핀이 실패해야 함.
