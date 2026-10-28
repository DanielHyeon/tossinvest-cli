# Branch Test Map: TestTheSummaryCostsNoBrokerCall

- Source: `internal/console/display_primitives_test.go` (359-371) — base 판. AST branches 2. 본문 무변경(function-logic-map.md). GREEN: 2026-10-28 `go test ./internal/console -run` 로 통과 확인.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 개요 3회 렌더 | `TestTheSummaryCostsNoBrokerCall` | no(사후 작성) | yes |
| B2 | 브로커 호출 증가 | `TestTheSummaryCostsNoBrokerCall`(실패 갈래 — 통과 시 미실행) | no | 미실행(설계상) |
