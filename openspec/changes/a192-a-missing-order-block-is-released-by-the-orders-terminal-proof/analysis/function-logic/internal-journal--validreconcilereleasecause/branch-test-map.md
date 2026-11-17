# Branch Test Map: `ValidReconcileReleaseCause` (journal)

- Source: `internal/journal/reconcile_states.go` (111-119). 커버 측정: `../../coverage/after-branch-counts.txt`(cover mode set — 1 = 실행됨). 편집 전 측정 `../../coverage/base-branch-counts.txt`.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 기존 분기 | 기존 시험(패키지 GREEN) | 해당 없음(기존 분기) | yes — cover cond=1 body=0 |
| B2 | 허용 원인 | `TestA192OrderTerminalReleaseCauseIsOneThisBuildWrites` | yes — `../../red/red-2-journal.txt` | yes — cover cond=0 body=1 |
| B3 | 그 밖 | `TestUnknownReleaseCauseIsRefused` | 해당 없음(기존) | yes — cover cond=1 body=1 |
