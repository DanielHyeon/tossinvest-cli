# Branch Test Map: TestTheCurrencyConsequenceIsRenderedAndNotFolded

- Source: `internal/console/settings_card_test.go` (267-283) — base 판. AST branches 4. 본문 무변경. GREEN: 2026-10-28 `go test ./internal/console -run` 로 통과 확인.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 고지 표식 없음 | `TestTheCurrencyConsequenceIsRenderedAndNotFolded`(실패 갈래 — 통과 시 미실행) | no(사후 작성) | 미실행(설계상) |
| B2 | 고지가 `.notice` 아님 | `TestTheCurrencyConsequenceIsRenderedAndNotFolded`(실패 갈래) | no | 미실행(설계상) |
| B3 | 요구 단어 5개 순회 | `TestTheCurrencyConsequenceIsRenderedAndNotFolded` | no | yes |
| B4 | 단어 빠짐 | `TestTheCurrencyConsequenceIsRenderedAndNotFolded`(실패 갈래) | no | 미실행(설계상) |
