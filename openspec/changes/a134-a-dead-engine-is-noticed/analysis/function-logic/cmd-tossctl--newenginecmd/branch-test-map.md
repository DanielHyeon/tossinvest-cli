# Branch Test Map: `newEngineCmd`

- Source: `cmd/tossctl/engine.go` (115-134) — 편집 뒤. AST branches 0 — 직선 함수라 B1 = 정상 경로 한 행.

| Branch | Anchor | Scenario | Test | a134 |
|---|---|---|---|---|
| B1 | 직선(분기 0) | 묶음 명령이 기존 하위 명령과 새 `heartbeat-check` 를 모두 가짐, 새 leaf 표지는 local · 비변이 | `TestA134HeartbeatCheckIsALocalReadOnlyLeafUnderEngine` · `TestTheAlertCommandsAreWiredUnderEngine`(기존) · `TestA094TheAttemptResolveCommandIsAttachedToTheEngineTree`(기존) · `TestMutatingAnnotationOnTradeCommands`(기존) | 편집 영향 |
