# Branch Test Map: `notifyRelaxation`

- Source: `internal/app/engine/risk_relaxation_command.go` (161-183); **편집 전** 측정 — `analysis/harness/coverage-pre-a066-relax.json`(연결 워크트리 `81934b46`, `./internal/app/engine` 시험 14개 `TestA066|PositionPolicyCommand` 를 하나씩).

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 177:2 | 적재 실패 → 완화됨 · 통지 실패 | `TestA066ReleaseStandsWhenTheNoticeFails` | 편집 전(기준선) | 블록 177.16-180.3을 시험 1개가 실행, PASS |
