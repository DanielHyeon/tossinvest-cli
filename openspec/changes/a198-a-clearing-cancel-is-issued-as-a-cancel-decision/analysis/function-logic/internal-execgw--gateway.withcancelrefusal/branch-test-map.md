# Branch Test Map: `Gateway.withCancelRefusal` (pre-edit)

시험 이름은 grep 근거(internal/execgw/cancel_refusal_test.go · a165_*_test.go), 실행은 4b.8 선택 시험에서 확인.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 정책 nil·비취소 | TestA165NilPolicyAndNonCancelKeepTheOriginalOutcome · TestA165PolicyDoesNotChangeNilOrNonCancelAmbiguousOutcomes | n/a(기존) | 4b.8 |
| B2 | already-processing 아님 | TestA165EnginePolicyLeavesOther409Ambiguous · TestA165CancelRefusalDoesNotOverrideNonAmbiguousDispatch | n/a | 4b.8 |
| B3 | 엔진 정책 확정 미수락 | TestA165EngineAlreadyProcessingIsOneConfirmedRefusalWithoutDelay | n/a | 4b.8 |
| B4 | 정책 값 0 | 미확인(grep 으로 특정 못 함) | n/a | — |
| B5~B13 | flatten 재시도 창 | TestA165FlattenRetriesWithinCumulativeBudget · TestA165WindowStartsBeforeFirstDispatch · TestA165CallerCancellationStopsRetry · TestA165FlattenRetryNotSentKeepsTheFirstAmbiguousOutcome · TestA165FlattenExhaustionKeepsTheFirstAmbiguousOutcome · TestA165RetryDeadlineKeepsFirstAmbiguousAsCancelled | n/a | 4b.8 |
| (예정) D7 | 엔진 정책 세 code | a198 분류 단위 시험(§4-13) · 결합 시험(§4-10~13b) | 4b.6 | 4b.7 |
