# Branch Test Map: `ReconcileDriver.RunOnce` (engine) — post-edit(v3 + V4c)

- Source: `internal/app/engine/reconcileloop.go` (430-492). 커버: `../../coverage/post-gstack-r1-internal-app-engine--reconciledriver.runonce.txt`(mode set, body=1 은 실행됨). RED/GREEN 원문 `../../tdd/red.txt` · `green.txt`, V4c 는 `../../tdd/red-gstack-r1.txt` · `green-gstack-r1.txt`.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 불안정 · 수집 오류 | 기존 `TestAnUnstableAccountDefersSilently` | 해당 없음(기존 분기) | yes — body=1 |
| B2 | 로컬 읽기 오류(표식 읽기 실패 흡수) | `TestA201AnUnreadableJournalFailsTheCycleAtTheLocalReadAsBefore` | 해당 없음(편집 전에도 통과하는 고정) | yes — body=1, PASS |
| B3 | 접기 오류 | 없음(선택 범위 — 편집 무관) | 해당 없음(기존 분기) | 미실행 — body=0(편집 전과 같음) |
| B4 | 수렴 오류(stale) | `TestA201W3aFillAfterConvergencePositionsIsCreditedLikeAnAdjustment` · `TestA201AFillJustBeforeTheAdjustmentTxIsCreditedLikeAnAdjustment` · `TestA201W3bWatermarkOnlyStaleThenTheFillIsCreditedLikeAnAdjustment` | yes — 다음 주기 해제 없음(`red.txt`) | yes — body=1, PASS |
| B5 | Refresh 오류 → V4c 반박 | `TestA201ARefreshFailureCycleStillRefutesADisputedFillCredit` · `TestA201ARefreshFailureCycleStillRefutesADisputedAdjustmentCredit` · `TestA201ARefreshFailureCycleKeepsACreditItsComparisonDoesNotDispute` · `TestA201ARefreshFailureDoesNotRefuteTheCreditItsOwnConvergenceGave` | yes — 체결 · 조정 판 FAIL(다음 깨끗한 주기 해제), 대조 둘 PASS(`red-gstack-r1.txt`) | yes — body=1, PASS |
| B6 | Refresh 오류 · 앞선 오류 없음 | 같은 시험(`c.Err` 단언) | 해당 없음(기존 분기) | yes — body=1 |
| B7 | Observe 오류 → 부여 없음 | `TestA201AnObserveErrorCycleGrantsNoFillCredit` · 기존 `TestTrackerReleaseFailureStopsBeforePriceAndAdoption` | 해당 없음(편집 전에도 통과하는 고정) | yes — body=1, PASS |
| B8 | Observe 오류 · 앞선 오류 없음 | 같은 시험 | 해당 없음(기존 분기) | yes — body=1 |
