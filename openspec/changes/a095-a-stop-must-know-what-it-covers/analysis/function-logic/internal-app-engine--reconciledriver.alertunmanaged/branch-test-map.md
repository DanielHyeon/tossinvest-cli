# Branch Test Map: `ReconcileDriver.alertUnmanaged`

- Source: `internal/app/engine/adoption.go`

> **「진입 실측」은 측정값이다** — 패키지 시험 전체를 `-covermode=set`으로 돌린 프로파일에서 그 분기가 만든 블록의 count다. 어떤 **개별** 시험이 그 분기를 밟는지는 이 실행이 답하지 않는다. 따라서 「Test」 열은 **a095 3판이 요구하는 시험**이며 현존 증명이 아니다. `[비움 — Qn]`은 사용자 결정이 덮지 않아 Manager 에게 올린 질문이다.

| Branch | 조건 | 진입 실측 | Test (a095 요구) | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:441` `if d.opts.NotificationsEnabled && fact == factEnabledFailed {` | 예 | **a095 2.2 · 2.5 · 2.5a · 2.5b** `TestA095AFailedAdoptionIsCriticalAndDurable` · `TestA095NotificationsOffNeverRecordsACritical` · `TestA095OnWithoutATopicIsStillCritical` · `TestA095TheProductionAssemblyReadsTheLoadedSwitch` | yes | yes |
| B2 | `:445` `if d.unmanaged[p.ID][fact] {` | 예 | **a095 2.12 · 2.15 · 2.16 · 2.17** `TestA095ADeliveredFailureIsRemindedAfterTheWindow` · `TestA095ARecordingFailureIsRetriedWhenTheStoreRecovers` · `TestA095TheFactIdentity` | yes | yes |
| B3 | `:448` `if d.unmanaged[p.ID] == nil {` | 예 | **a095 2.17** `TestA095TheFactIdentity` — 다른 사실은 삼키지 않음 · 같은 사실은 억제 | yes | yes |

**미진입 분기 0개**: 없음
**자체 블록 없는 분기 0개**: 없음 — 컴파일러가 별도 블록을 만들지 않는 형태(빈 `switch {` 등)이며 미커버와 다르다.
