# Branch Test Map: `Tracker.Restore` (reconcile)

- Source: `internal/reconcile/mismatch.go` (966-1012). 커버 측정: `../../coverage/after-branch-counts.txt`(cover mode set — 1 = 실행됨). 편집 전 측정 `../../coverage/base-branch-counts.txt`.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 기존 분기 | 기존 시험(패키지 GREEN) | 해당 없음(기존 분기) | yes — cover cond=1 body=0 |
| B2 | 기존 분기 | 기존 시험(패키지 GREEN) | 해당 없음(기존 분기) | yes — cover cond=1 body=0 |
| B3 | 기존 분기 | 기존 시험(패키지 GREEN) | 해당 없음(기존 분기) | yes — cover cond=1 body=1 |
| B4 | 기존 분기 | 기존 시험(패키지 GREEN) | 해당 없음(기존 분기) | yes — cover cond=1 body=1 |
| B5 | 영구 행 있음 | 기존 + `TestA192APermanentBlockStaysOperatorOnly` | 해당 없음(기존 분기) | yes — cover cond=1 body=1 |
| B6 | 기존 분기 | 기존 시험(패키지 GREEN) | 해당 없음(기존 분기) | yes — cover cond=1 body=1 |
