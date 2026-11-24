# Branch Test Map: `ReconcileDriver.RunOnce` (engine) — post-edit(v3)

- Source: `internal/app/engine/reconcileloop.go` (425-485). 커버: `../../coverage/post-internal-app-engine--reconciledriver.runonce.txt`(mode set, body=1 은 실행됨). 편집 전 측정 `../../coverage/base-runonce-counts.txt`.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 불안정 · 수집 오류 | 기존 `TestAnUnstableAccountDefersSilently` | 해당 없음(기존 분기) | yes — body=1 |
| B2 | 로컬 읽기 오류(표식 읽기 실패 흡수) | `TestA201AnUnreadableJournalFailsTheCycleAtTheLocalReadAsBefore` | 해당 없음(편집 전에도 통과하는 고정 — `../../tdd/red.txt`) | yes — body=1, PASS(`../../tdd/green.txt`) |
| B3 | 접기 오류 | 없음(선택 범위 — 편집 무관) | 해당 없음(기존 분기) | 미실행 — body=0(편집 전과 같음) |
| B4 | 수렴 오류(stale) | `TestA201W3aFillAfterConvergencePositionsIsCreditedLikeAnAdjustment` · `TestA201AFillJustBeforeTheAdjustmentTxIsCreditedLikeAnAdjustment` · `TestA201W3bWatermarkOnlyStaleThenTheFillIsCreditedLikeAnAdjustment` | yes — 다음 주기 해제 없음(`red.txt`) | yes — body=1, PASS |
| B5 | Refresh 오류 | 없음(선택 범위 — 편집 무관) | 해당 없음(기존 분기) | 미실행 — body=0(편집 전과 같음) |
| B6 | Refresh 오류 · 앞선 오류 없음 | 없음(선택 범위 — 편집 무관) | 해당 없음(기존 분기) | 미실행 — body=0(편집 전과 같음) |
| B7 | Observe 오류 → 부여 없음 | `TestA201AnObserveErrorCycleGrantsNoFillCredit` · 기존 `TestTrackerReleaseFailureStopsBeforePriceAndAdoption` | 해당 없음(편집 전에도 통과하는 고정) | yes — body=1, PASS |
| B8 | Observe 오류 · 앞선 오류 없음 | 같은 시험 | 해당 없음(기존 분기) | yes — body=1 |
