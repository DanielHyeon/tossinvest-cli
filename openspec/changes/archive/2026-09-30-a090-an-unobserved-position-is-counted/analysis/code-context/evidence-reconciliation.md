# Evidence reconciliation — a090

- 기준: base `d3bd1843`, `exitloop.go` sha256 `522d5d81…`(AST·진입 실측과 동일 바이트).
- 확인한 파일·심볼: `internal/app/engine/exitloop.go`(`ObserveOnce` `observe` `checkOutage` `quoteUsable` `judge` `alert` 머리 주석) ·
  `internal/obs/event.go`(`EventExitObservationOutage`, 등급표) · `internal/obs/notifier.go`(`DefaultCriticalAttempts`, `n.mu`) ·
  `internal/official/market_reads.go`(`adaptPrices`) · `internal/journal/operating_mode.go`(`ModeTriggerExitObservationOutage`) ·
  `internal/app/engine/a111_flat_exit_observation_test.go`(:735 :759 :833 :1077) · `openspec/specs/exit-policy/spec.md` 「관측 경로와 fail-safe」.

| # | 불일치 | 현재 HEAD 로 내린 결론 |
|---|---|---|
| R1 | CodeGraph `callers ObserveOnce` 는 시험 6 만 보이고 하네스(`h.observe()`) 경유 시험 수십 개는 안 보인다 | 영향 시험 판단은 `affected --filter '*_test.go'`(898)와 커버리지 실측으로 한다 |
| R2 | a092 번들(7 분기)의 B6 `:455` · B7 `:462` ↔ 현재 AST 8 분기의 B6 `:453` · B8 `:465` | a111 `882a0b49` 가 `quoteUsable` 분기(현 B7)를 더했다. 번호는 이 change 의 `ast.json` 기준으로 쓴다 |
| R3 | a092 BTM: "B6 — 시험 없음" ↔ 실측 B6 진입 | a111 시험 :759 가 B6 에 닿는다. 다만 무음이 옳다고 단언한다 — 계수·보고 시험은 여전히 0 |
| R4 | `affected` 기본 실행 시험 1 | 알려진 판별식 결함(CLAUDE.md 하네스 주석). `--filter` 결과를 쓴다 |
