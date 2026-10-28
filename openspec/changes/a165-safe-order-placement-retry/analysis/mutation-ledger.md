# a165 변이 원장

2026-10-29. 모든 변이는 작업 트리에서 한 번에 하나씩 적용하고 지정 시험이 실패하는 것을 확인한 뒤 원복함. 아래 `잡힘`은 변이 뒤의 실패를 뜻하며, 대조군은 원복 뒤 같은 표적 시험을 다시 통과한 경우임. `대상 도달`은 컴파일 오류가 아니라 해당 분기 또는 정책 값을 통과해 관찰된 실패임.

## 공통 대조군과 원복

- 무변이 대조군: `TestA165...` 표적 시험군 및 영향 패키지 시험은 GREEN에서 통과함.
- 원복 확인 SHA-256: `cancel_refusal.go` `acd4d7de2fb725dcfae60314d164a13116be4d47d0e07ea2010237228f4a68e9`, `cancel_refusal_latch.go` `27002543995267a9c961eae26432922eba7ae3b2a949c4509120bd9548a94a56`, `journal/cancel_refusal_latch.go` `6debab689bcc9650fe64aa19af89f001b43fb8ba457ca39e041615d05364d784`, `gateway.go` `c304e32a62e42759ad1ffb857ca90461068301707dadb5ca7bd8c46dba8af2de`, `dispatch.go` `f77e0bab6044ecb1fab0c55ab0bd3fab2411939687ffdcc68ea7d785b2b1d679`.
- 변이 원문 중 대표 실패: `TestA165WindowStartsBeforeFirstDispatch`는 경계 `>=`를 `>`로 바꾸자 8초 뒤 `stop=cancelled`가 되어 실패함. `TestA165FlattenExhaustionKeepsTheFirstAmbiguousOutcome`은 소진 결과를 `outcome=last`로 바꾸자 실패함. 이 둘은 대상 줄 도달을 직접 보인 추가 경계 변이임.

## 판정·정책·루프

| ID | 변이 | 대상 도달·잡은 시험 | 결과 |
|---|---|---|---|
| M01 | `already-processing`의 대소문자 비교 제거 | `TestA165AlreadyProcessingRefusalRequiresAnUnambiguousCode/top-level_code` | 잡힘 |
| M02 | 최상위·중첩 code 모순 거부 제거 | `TestA165AlreadyProcessingRefusalRequiresAnUnambiguousCode/contradictory_codes` | 잡힘 |
| M03 | HTTP 409 조건을 400으로 변경 | `TestA165AlreadyProcessingRequiresHTTP409` | 잡힘 |
| M04 | `KindCancel` 범위 제거 | `TestA165PolicyDoesNotChangeNilOrNonCancelAmbiguousOutcomes/place`, `amend` | 잡힘 |
| M05 | 첫 결과의 `DispatchAmbiguous` 가드 제거 | `TestA165CancelRefusalDoesNotOverrideNonAmbiguousDispatch` | 잡힘 |
| M06 | `Definitive` 분기 반전 | `TestA165EngineAlreadyProcessingIsOneConfirmedRefusalWithoutDelay` | 잡힘 |
| M07 | retry-after 기본 1초를 2초로 변경 | `TestA165AlreadyProcessingRefusalRequiresAnUnambiguousCode/non-numeric_retry` | 잡힘 |
| M08 | 누적 대기 상한 5초를 4초로 변경 | `TestA165CancelRefusalPoliciesAreExact` | 잡힘 |
| M09 | 추가 재전송 상한 2회를 3회로 변경 | `TestA165CancelRefusalPoliciesAreExact`, `TestA165FlattenWiringPolicyKeepsExhaustedCancelInDoubt` | 잡힘 |
| M10 | 창 두 경계를 `>=`에서 `>`로 완화 | `TestA165WindowStartsBeforeFirstDispatch` — 8초 경계가 sleep 뒤 `stop=cancelled`로 변함 | 잡힘 |
| M11 | 재전송 `NOT_SENT`를 첫 AMBIGUOUS로 되돌리지 않음 | `TestA165FlattenRetryNotSentKeepsTheFirstAmbiguousOutcome` | 잡힘 |
| M12 | 소진 결과를 오늘의 첫 AMBIGUOUS가 아닌 마지막 결과로 표기 | `TestA165FlattenExhaustionKeepsTheFirstAmbiguousOutcome` — `outcome=last` 관찰 | 잡힘 |
| M13 | 공유 확정 거절 목록에 `already-processing` 추가 | `TestA165PolicyDoesNotChangeNilOrNonCancelAmbiguousOutcomes`, `TestA165FlattenWiringPolicyKeepsExhaustedCancelInDoubt` | 잡힘 |

## 컨텍스트·배선

| ID | 변이 | 대상 도달·잡은 시험 | 결과 |
|---|---|---|---|
| M14 | `winCtx` 부모를 호출자 context와 분리 | `TestA165CallerCancellationStopsRetry` | 잡힘 |
| M15 | `DispatchVerified` 정착 context를 `winCtx`로 교체 | `TestA165RealFirstDispatchDelayDoesNotDriveRetryWindowJudgement`, `TestA165SubmitSettlesWithItsCallerContext` | 잡힘 |
| M16 | 재호출에 `winCtx` 대신 원 `dctx` 사용 | `TestA165RetryInvokesTheBrokerWithTheWindowContext` | 잡힘 |
| M17 | 창 시작을 첫 결과 뒤로 이동 | `TestA165WindowStartsBeforeFirstDispatch` | 잡힘 |
| M18 | flatten 배선에 엔진 정책 주입 | `TestA165FlattenWiringUsesTheBoundedNonDefinitivePolicy`, flatten Held 시험 | 잡힘 |
| M19 | 엔진 배선에 flatten 정책 주입 | `TestA165EngineGatewayWiresOnlyTheDefinitiveCancelRefusalPolicy` | 잡힘 |
| M20 | flatten 또는 엔진의 정책 배선 제거(nil) | 각 배선 정책 시험과 `TestA165NilPolicyAndNonCancelKeepTheOriginalOutcome`의 오늘 경로 대조 | 잡힘 |
| M21 | flatten 정책을 혼성 `{Retries: 2, Definitive: true}`로 변경 | `TestA165FlattenWiringUsesTheBoundedNonDefinitivePolicy` | 잡힘 |
| M36 | `WithTimeout` 부모를 `context.Background()`로 분리 | `TestA165RealTimeoutKeepsFirstAmbiguousAsCancelled`가 fake-clock sleeper가 부모 timeout에 끝나지 않음을 관측, 구조 핀도 부모 식 0을 관측 | 잡힘 — 대상 `cancel_refusal.go:66` 도달 |
| M37 | `time.Now` 기반 `WithDeadline` 판정을 복원 | `TestA165CancelRefusalUsesOnlyInjectedClockForJudgement`가 `time.Now(`를 관측 | 잡힘 — 대상 `cancel_refusal.go:66` 도달 |
| M38 | retry dispatch 뒤 `winCtx.Err()` 보존 분기 제거 | `TestA165RetryDeadlineKeepsFirstAmbiguousAsCancelled`가 둘째 timeout의 `stop=ack outcome=retried`를 관측 | 잡힘 — 대상 `cancel_refusal.go:92` 도달, 원복 SHA 일치 |

## D11 latch·해제 범위

| ID | 변이 | 대상 도달·잡은 시험 | 결과 |
|---|---|---|---|
| M22 | latch 호출을 `!raisesExposure` 반환 앞으로 이동 | `TestA165LatchDoesNotBlockACancel`, `TestA165LatchDoesNotBlockAReduceOnlySell` | 잡힘 |
| M23 | latch를 공용 `unsettledFor`에 넣음 | `TestA165LatchDoesNotBlockACancel` — 후속 취소가 거절됨 | 잡힘 |
| M24 | terminal r1(live order 부재) 해제 제거 | `TestA165TerminalEvidenceReleasesTheCancelLatch` | 잡힘 |
| M25 | r2(`ConfirmedCancelAfterLatch`) 해제 제거 | `TestA165EngineCancelLatchBlocksOnlyNewEntryUntilJournalledRelease` | 잡힘 |
| M26 | latch 꼬리표 파싱을 항상 거짓으로 변경 | `TestA165EngineCancelLatchBlocksOnlyNewEntryUntilJournalledRelease` — 새 진입이 broker에 도달함 | 잡힘 |
| M27 | r2 순서 `recorded_seq > latch_seq` 제거 | `TestA165ConfirmedCancelAfterLatchRequiresRecordedAfterTheLatch` | 잡힘 |
| M28 | r2 대상 주문 id 비교 제거 | `TestA165ConfirmedCancelAfterLatchRequiresRecordedAfterTheLatch`의 다른 대상 경우 | 잡힘 |
| M29 | r2 순번을 RECORDED 대신 CONFIRMED transition으로 변경 | `TestA165ConfirmedCancelAfterLatchRequiresRecordedAfterTheLatch`의 기록 선행·정착 후행 경우 | 잡힘 |
| M30 | 거래일 동일 조건을 다시 추가 | `TestA165ConfirmedCancelAfterLatchUsesTheLatchScopeAndOneOwner/same_target_after_latch_releases_across_trading_days` | 잡힘 |
| M31 | 계좌 비교 제거 | `TestA165ConfirmedCancelAfterLatchUsesTheLatchScopeAndOneOwner/different_account_does_not_release` | 잡힘 |
| M32 | 시장 비교 제거 | `TestA165ConfirmedCancelAfterLatchUsesTheLatchScopeAndOneOwner/different_market_does_not_release` | 잡힘 |
| M33 | 종목 비교 제거 | `TestA165ConfirmedCancelAfterLatchUsesTheLatchScopeAndOneOwner/different_symbol_does_not_release` | 잡힘 |
| M34 | 주인 유일성 `n = 1`을 제거 또는 `>= 1`로 완화 | `TestA165ConfirmedCancelAfterLatchUsesTheLatchScopeAndOneOwner/multiple_owners_do_not_release` | 잡힘 |
| M35 | 주인 유일성 `n = 1`을 `n <= 1`로 완화 | `TestA165ConfirmedCancelAfterLatchUsesTheLatchScopeAndOneOwner/no_owner_does_not_release` | 잡힘 |

살아남은 변이: 없음. 메모리 latch는 구현에 존재하지 않아 독립 변이 대상이 아니며, D11은 재시작 뒤에도 원장 질의를 다시 하는 `TestA165EngineCancelLatchBlocksOnlyNewEntryUntilJournalledRelease`로 대체 검증함.
