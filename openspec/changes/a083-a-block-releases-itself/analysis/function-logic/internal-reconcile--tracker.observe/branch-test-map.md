# Branch Test Map: `Tracker.Observe`

AST의 모든 분기를 1행씩 덮는다. ★ 표시가 이 change가 편집하는 분기다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | (366) 첫 관측에서 block map이 초기화된다 | `TestAQuantityMismatchBlocksEntries` | n/a | 기존 |
| B2 | (373) 비교 전체가 일치하면 실패 카운터가 0으로 리셋된다 | `TestACleanReconcileResetsTheFailureCounter` | n/a | 기존 |
| B3 | (396) 비교가 불일치하면 실패 카운터가 증가한다 | `TestAQuantityMismatchBlocksEntries` | n/a | 기존 |
| B4 | (380) 활성 차단 전체를 해제 후보로 훑는다 | `TestAnAdjustmentAndAMatchingRecheckRelease` | n/a | 기존 |
| B5 | (381) 다른 producer(cause·release)의 상태는 해제하지 않는다 | `TestRestoreProjectsStatesThisTrackerDidNotEnter` | n/a | 기존 |
| B6 | ★(388) credit 사용 판정: 같은 비교는 쓰지도 버리지도 않고, 더 나중 as-of의 일치만 해제하며, as-of 부재·역순 credit은 쓰지 않는다 | `TestTheSameComparisonNeitherSpendsNorDiscardsACredit`, `TestAnAdjustmentAndAMatchingRecheckRelease`, `TestAnUndatedComparisonCannotSpendACredit`, `TestACreditFromALaterComparisonIsNotSpent`, `TestACoincidentalAgreementStillDoesNotRelease` | yes | yes |
| B7 | (398) diff의 불일치마다 심볼 차단이 만들어진다 | `TestAQuantityMismatchBlocksEntries` | n/a | 기존 |
| B8 | (399) 이미 있는 차단은 중복 추가되지 않고 `Since`가 보존된다 | `TestRefreshCannotOverwriteABlockPersistedByObserve` | n/a | 기존 |
| B9 | (405) 연속 3회 실패가 계좌 범위 영구 차단으로 승격된다 | `TestThreeConsecutiveFailuresArePermanent` | n/a | 기존 |
| B10 | (418) 영구 차단은 한 번만 만들어진다 | `TestThreeConsecutiveFailuresArePermanent` | n/a | 기존 |
| B11 | (440) 이번 관측이 추가한 차단 집합을 만든다 | `TestObservePublishesOnlyDurablePartialReleasesAndRetriesTheRemainder` | n/a | 기존 |
| B12 | (443) pending 차단을 다음 관측에서 재시도한다 | `TestObserveKeepsMemoryAndGateBlockedWhenDurableReleaseFails` | n/a | 기존 |
| B13 | (444) 이미 Added에 있는 pending은 두 번 넣지 않는다 | `TestObserveKeepsMemoryAndGateBlockedWhenDurableReleaseFails` | n/a | 기존 |
| B14 | (449) durable 확인된 차단의 pending이 풀린다 | `TestWriteThroughRecordsTheStateWithItsEvidence` | n/a | 기존 |
| B15 | (453) 다른 cause가 소유한 durable 상태가 authoritative로 교체된다 | `TestRestoreProjectsStatesThisTrackerDidNotEnter` | n/a | 기존 |
| B16 | (456) 커밋된 해제만 메모리에서 삭제된다 | `TestObservePublishesOnlyDurablePartialReleasesAndRetriesTheRemainder` | n/a | 기존 |
| B17 | ★(462) persist 실패 시 미커밋 해제 credit은 남고 미사용 credit은 보존된다 | `TestObservePublishesOnlyDurablePartialReleasesAndRetriesTheRemainder` | yes | yes |
| B18 | ★(475) 소멸은 심볼 단위다: 그 심볼이 여전히 불일치할 때만 소멸하고, 무관한 심볼의 불일치나 다른 심볼의 credit은 영향이 없다 | `TestAnAdjustmentIsSpentByTheRecheckItAnswers`, `TestAnUnrelatedMismatchDoesNotDiscardAnAnsweredCredit`, `TestAnAdjustmentOnAnotherSymbolDoesNotRelease` | yes | yes |
| B19 | ★(467) persist 실패 시 살아남은 차단만 credit 보존 대상으로 훑는다 | `TestObservePublishesOnlyDurablePartialReleasesAndRetriesTheRemainder` | yes | yes |
| B20 | ★(468) 영구 차단은 credit으로 풀리지 않는다 | `TestAnAdjustmentDoesNotReleaseAPermanentMismatch` | n/a | 기존 |

추가 통합 근거 (분기 단위가 아닌 계약):

- 드라이버 한 주기: 수렴 주기는 차단을 유지하고 다음 주기의 재조회가 해제한다 —
  `TestTheCycleAfterAConvergenceReleasesTheBlock` (RED yes / GREEN yes)
- 재시작은 credit을 복원하지 않는다 — `TestARestartKeepsTheReconcileBlock` (기존)
