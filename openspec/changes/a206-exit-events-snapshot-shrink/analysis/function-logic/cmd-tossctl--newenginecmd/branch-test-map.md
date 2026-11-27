# Branch Test Map: `newEngineCmd`

- Source: `cmd/tossctl/engine.go` (115-136) — 편집 뒤. AST branches 0 — 직선 함수라 B1 = 정상 경로 한 행.

| Branch | Anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | 직선(분기 0) | 묶음 명령이 새 `exit-evidence-compact` 를 가짐, 표지 mutating=true | `TestA206CompactCommandIsMutatingAndAttached` · `TestMutatingAnnotationOnTradeCommands`(기존) | 등록 전 「is not attached」 FAIL | PASS |
