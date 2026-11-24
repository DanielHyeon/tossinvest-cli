# Branch Test Map: `ValidReconcileReleaseCause` (journal) — post-edit(v3)

- Source: `internal/journal/reconcile_states.go` (111-119). 커버: `../../coverage/post-internal-journal--validreconcilereleasecause.txt`(mode set). RED/GREEN 원문 `../../tdd/red.txt` · `../../tdd/green.txt`.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 원인 분기 | 기존 `TestUnknownReleaseCauseIsRefused` | 해당 없음(기존 분기) | yes — body=1 |
| B2 | 허용 원인(`FILL_APPLIED` 추가) | `TestA201FillAppliedIsAReleaseCauseThisBuildWrites` | yes — FAIL(`FILL_APPLIED must be a valid release cause`) | yes — body=1, PASS |
| B3 | 그 밖 | 기존 `TestUnknownReleaseCauseIsRefused` | 해당 없음(기존 분기) | yes — body=1 |
