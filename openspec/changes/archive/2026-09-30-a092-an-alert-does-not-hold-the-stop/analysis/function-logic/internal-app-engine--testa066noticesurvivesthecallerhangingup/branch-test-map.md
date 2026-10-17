# Branch Test Map: `TestA066NoticeSurvivesTheCallerHangingUp`

- Source: `internal/app/engine/a066_risk_relaxation_test.go`. 시험 코드 — 분기는 픽스처 · 단언 갈래. `go test ./internal/app/engine ./cmd/tossctl` 전부 PASS(커밋 `0e4f26af`).

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 324:2 | 픽스처 · 단언 갈래 | 자기 자신 | 해당 없음(시험 코드) | PASS |
