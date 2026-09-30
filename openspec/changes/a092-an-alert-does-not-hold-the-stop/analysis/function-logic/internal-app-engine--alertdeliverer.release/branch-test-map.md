# Branch Test Map: `alertDeliverer.release`

- Source: `internal/app/engine/alertdelivery.go`; **편집 전** 측정 — `analysis/harness/coverage-pre-r26-deliverer.json`(연결 워크트리 `e55102f0`, 시험 19개).

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 605:2 | `if _, err := d.led().ReleaseAlertClaim(relCtx, id, token); e` | (미실행) | 편집 전(기준선) | 측정 표본의 시험 0개 |
