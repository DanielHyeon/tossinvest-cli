# Branch Test Map: `ValidReconcileReleaseCause` (journal) — pre-edit(v3)

- Source: `internal/journal/reconcile_states.go` (111-119). 편집 전 지도 — RED/GREEN 은 구현(2.2 · 2.4).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 원인 분기 | 기존 `TestUnknownReleaseCauseIsRefused` | 해당 없음(편집 전) | 미측정(2.4) |
| B2 | 허용 원인(v3: `FILL_APPLIED` 추가) | 기존 + v3 「FILL_APPLIED 는 유효 · 해제 쓰기 성공」 | 해당 없음(편집 전) | 미측정(2.4) |
| B3 | 그 밖 | 기존 `TestUnknownReleaseCauseIsRefused` | 해당 없음(편집 전) | 미측정(2.4) |
