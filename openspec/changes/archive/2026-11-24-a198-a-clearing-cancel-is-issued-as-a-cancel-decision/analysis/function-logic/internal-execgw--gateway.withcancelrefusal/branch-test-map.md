# Branch Test Map: `Gateway.withCancelRefusal` (post-edit — 새 번호)

시험 이름은 grep 근거(internal/execgw/cancel_refusal_test.go · a165_*_test.go · a198_settled_cancel_refusal_test.go). GREEN 은 4b.8 선택 시험 원문(`analysis/verify/4b-*.txt`).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 정책 nil·비취소 | TestA165NilPolicyAndNonCancelKeepTheOriginalOutcome · TestA165PolicyDoesNotChangeNilOrNonCancelAmbiguousOutcomes · TestA198OtherAnswersKeepTheirOutcome(nil 정책·PLACE·AMEND) | n/a(기존) | PASS |
| B2 | 엔진 정책(Definitive) 갈래 | TestA198EngineSettledOrderRefusalsAreDefinitive · TestA198OtherAnswersKeepTheirOutcome(flatten 정책 = 거짓) | — | PASS |
| B3 | 세 code 모순 없이 → 확정 거절 / 그 밖 → 통과 | TestA198EngineSettledOrderRefusalsAreDefinitive(6 본문) · TestA198OtherAnswersKeepTheirOutcome · TestA198ANonAmbiguousOutcomeIsNotReclassified | FAIL(red-4b.txt) | PASS |
| B4 | already-processing 아님 | TestA165EnginePolicyLeavesOther409Ambiguous · TestA165CancelRefusalDoesNotOverrideNonAmbiguousDispatch | n/a | PASS |
| B5 | 엔진 정책 확정 미수락 | TestA165EngineAlreadyProcessingIsOneConfirmedRefusalWithoutDelay · TestA198AlreadyProcessingKeepsTheA165Path | n/a | PASS |
| B6 | 정책 값 0 | 미확인 | n/a | — |
| B7 | 재시도 루프 | TestA165FlattenRetriesWithinCumulativeBudget | n/a | PASS |
| B8 | 대기 상한 자르기 | TestA165FlattenRetriesWithinCumulativeBudget | n/a | PASS |
| B9 | 대기 0 이하 | TestA165FlattenExhaustionKeepsTheFirstAmbiguousOutcome | n/a | PASS |
| B10 | 창 초과 예상 | TestA165WindowStartsBeforeFirstDispatch | n/a | PASS |
| B11 | Sleep 취소 | TestA165CallerCancellationStopsRetry | n/a | PASS |
| B12 | 대기 뒤 창 초과 | TestA165WindowStartsBeforeFirstDispatch | n/a | PASS |
| B13 | 창 context 종료 + 재전송 모호/미전송 | TestA165RetryDeadlineKeepsFirstAmbiguousAsCancelled | n/a | PASS |
| B14 | 재전송이 already-processing 모호 아님 | TestA165FlattenRetriesWithinCumulativeBudget | n/a | PASS |
| B15 | 재전송 미전송 | TestA165FlattenRetryNotSentKeepsTheFirstAmbiguousOutcome | n/a | PASS |
