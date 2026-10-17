# Branch Test Map: 삭제된 a124 핀 (TheModeProjectorHasNoProductionCaller — base)

- Source: `internal/app/engine/a124_the_enforcement_boundary_internal_test.go`. 시험 코드.

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 93:2 | 단언 · 픽스처 갈래 | `TestTheModeProjectorIsBoundOnlyByTheEngineAssembly`(대체 핀) | 배선 편집 뒤 옛 핀 FAIL(`the mode projector is now wired in production`) — 예고된 빨강 | 삭제됨 |
| B2 | if at 97:2 | 단언 · 픽스처 갈래 | `TestTheModeProjectorIsBoundOnlyByTheEngineAssembly`(대체 핀) | 배선 편집 뒤 옛 핀 FAIL(`the mode projector is now wired in production`) — 예고된 빨강 | 삭제됨 |
