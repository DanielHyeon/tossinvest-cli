# Branch Test Map: `ReadOnly.AccountExitEvents`

본문 무변경이므로 RED는 없다(이 change가 깨뜨린 동작이 없다). GREEN은
`go test ./internal/journal/...` 전체 통과로 확인한다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `limit<=0`이면 질의 없이 빈 결과 | account_views_test.go 한도 케이스 | — (동작 무변경) | yes |
| B2 | 질의 실패가 문장으로 보고된다 | 스키마 거절은 `TestALedgerWithoutTheAttemptTableIsRefusedAtOpenRatherThanPerQuery`가 open 단계에서 선차단함을 증명 | — | yes |
| B3 | 행 순회 — 최신 창을 오름차순으로 되돌림 | account_views_test.go exit history 케이스 | — | yes |
| B4 | Scan 실패가 삼켜지지 않는다 | 컬럼 계약 회귀 가드 | — | yes |
| B5 | `rows.Err()`가 무시되지 않는다 | 동상 | — | yes |
