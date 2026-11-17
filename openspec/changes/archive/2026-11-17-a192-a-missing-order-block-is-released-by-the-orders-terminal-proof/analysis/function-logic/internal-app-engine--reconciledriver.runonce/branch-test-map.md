# Branch Test Map: `ReconcileDriver.RunOnce` (engine)

- Source: `internal/app/engine/reconcileloop.go` (418-472). 커버 측정: `../../coverage/after-branch-counts.txt`(cover mode set — 1 = 실행됨). 편집 전 측정 `../../coverage/base-branch-counts.txt`.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 기존 분기 | 기존 시험(패키지 GREEN) | 해당 없음(기존 분기) | yes — cover cond=1 body=1 |
| B2 | 기존 분기 | 기존 시험(패키지 GREEN) | 해당 없음(기존 분기) | yes — cover cond=1 body=0 |
| B3 | 기존 분기 | 기존 시험(패키지 GREEN) | 해당 없음(기존 분기) | yes — cover cond=1 body=0 |
| B4 | 기존 분기 | 기존 시험(패키지 GREEN) | 해당 없음(기존 분기) | yes — cover cond=1 body=0 |
| B5 | 기존 분기 | 기존 시험(패키지 GREEN) | 해당 없음(기존 분기) | yes — cover cond=1 body=0 |
| B6 | 기존 분기 | 기존 시험(패키지 GREEN) | 해당 없음(기존 분기) | 미실행 — base 측정에서도 같은 블록이 0 이었거나(`base-branch-counts.txt`) 선택 시험(-run) 밖; 이 change 가 만든 갈래 아님 — cover cond=0 body=0 |
| B7 | Observe 오류 | 기존 `TestTrackerReleaseFailureStopsBeforePriceAndAdoption` | 해당 없음(기존) | yes — cover cond=1 body=1 |
| B8 | 기존 분기 | 기존 시험(패키지 GREEN) | 해당 없음(기존 분기) | yes — cover cond=1 body=1 |
