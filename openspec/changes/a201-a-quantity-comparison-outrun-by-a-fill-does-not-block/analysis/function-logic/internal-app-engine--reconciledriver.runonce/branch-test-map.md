# Branch Test Map: `ReconcileDriver.RunOnce` (engine) — pre-edit

- Source: `internal/app/engine/reconcileloop.go` (425-479). 커버 측정(편집 전): `../../coverage/base-runonce-counts.txt`(cover mode set, 선택 시험 28 — reconcileloop_test · a083 · a192 로그 · a201 재현; 1 = 실행됨).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 불안정 · 수집 오류 | 기존 `TestAnUnstableAccountDefersSilently` | 해당 없음(기존 분기) | yes — cover cond=1 body=1 |
| B2 | 로컬 읽기 오류 | `TestA201AnUnreadableJournalFailsTheCycleAtTheLocalReadAsBefore`(2.2 에서 추가 — 워터마크 읽기가 이 앞에 들어감) | 해당 없음(편집 전에도 통과해야 하는 고정) | 2.4 에서 측정 |
| B3 | 접기 오류 | 없음(선택 범위) | 해당 없음(기존 분기) | 미실행 — cover body=0 |
| B4 | 수렴 오류(stale) | `TestA201W3aFillAfterConvergencePositionsIsCreditedLikeAnAdjustment` 외 3(1단계 재현에서 옮김 — 2.2) | 해당 없음(편집 전 기존 분기) | yes — cover cond=1 body=1(편집 전 1단계 재현으로 측정) |
| B5 | Refresh 오류 | 없음(선택 범위) | 해당 없음(기존 분기) | 미실행 — cover body=0 |
| B6 | Refresh 오류 · 앞선 오류 없음 | 없음(선택 범위) | 해당 없음(기존 분기) | 미실행 — cover body=0 |
| B7 | Observe 오류 | 기존 `TestTrackerReleaseFailureStopsBeforePriceAndAdoption` | 해당 없음(기존 분기) | yes — cover cond=1 body=1 |
| B8 | Observe 오류 · 앞선 오류 없음 | 같은 시험 | 해당 없음(기존 분기) | yes — cover cond=1 body=1 |
