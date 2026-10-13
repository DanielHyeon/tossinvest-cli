# Branch Test Map: `ApplyExitFill`

편집 전 좌표(base `233d5d9c`). RED/GREEN 은 구현 뒤 갱신(아래 「a130 시나리오」).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:963` exitContextForFill 오류 | 기존 exit apply 시험 | no | no |
| B2 | `:966` 대상 없음 | 외부 주문 시험 | no | no |
| B3 | `:973` exit 상태 없음 | TestAnExternalPositionHasNoOutcome | no | no |
| B4 | `:978` PendingState 오류 | 없음 | no | no |
| B5 | `:982` SELL | 매도 시험 | no | no |
| B6 | `:985` TakenAfterFill 오류 | 없음 | no | no |
| B7 | `:989` CompareDecimal 오류 | 없음 | no | no |
| B8 | `:992` cmp > 0 | 부분 익절 시험 | no | no |
| B9 | `:992` cmp > 0 | 부분 익절 시험 | no | no |
| B10 | `:993` MoveTakenRatioTotal 오류 | 없음 | no | no |
| B11 | `:999` 종결 · 대기 제안 일치 | 제안 해소 시험 | no | no |
| B12 | `:1001` ResolvePending 오류 | 없음 | no | no |
| B13 | `:1004` exit 이벤트 오류 | 없음 | no | no |
| B14 | `:1013` CLOSED ∧ 미완료 | TestTheOutcomeIsFrozenInTheClosingTransaction · TestAnAnalyticsFailureDoesNotRollBackTheClose | no | no |
| B15 | `:1014` markExitCompletedTx 오류 | 없음 | no | no |

## a130 시나리오 (3판)

- B14 — CLOSED + 거절 → ctx 목록에 쌓임, 커밋 뒤 1회 보고 — `TestA130AGrownPositionIsRefusedInsideTheFillTransaction` · `TestA130ARolledBackCloseReportsNothing` · `TestA130APanickingRefusalSinkDoesNotRollBackTheClose`
- 그 밖 — 무편집
