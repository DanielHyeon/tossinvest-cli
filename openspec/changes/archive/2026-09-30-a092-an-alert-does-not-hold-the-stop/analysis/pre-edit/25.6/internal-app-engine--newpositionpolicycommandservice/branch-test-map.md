# Branch Test Map: `NewPositionPolicyCommandService`

- Source: `internal/app/engine/position_policy_command.go` (90-104); **편집 전** 측정 — `analysis/harness/coverage-pre-a066-relax.json`(연결 워크트리 `81934b46`, `./internal/app/engine` 시험 14개 `TestA066|PositionPolicyCommand` 를 하나씩).

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 91:2 | 원장 없는 Context | `TestPositionPolicyCommandServiceRequiresEngineOwnedJournal` | 편집 전(기준선) | 블록 91.40-93.3, PASS |
| B2 | if at 94:2 | 시계 기본값 | (미실행) | 편집 전(기준선) | 측정 표본의 시험 0개 |
