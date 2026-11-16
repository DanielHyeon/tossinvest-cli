# Branch Test Map: `JournalTracked.TrackedOrders` (internal/filldetect)

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 저널 없음 | 기존 시험(ledger_test.go) | 기존 | yes |
| B2 | 계좌 없음 | 기존 | 기존 | yes |
| B3 | 원장 오류 | 기존 | 기존 | yes |
| B4 | IntentID 를 옮김 | `TestJournalTrackedCarriesTheIntentID` | **관측 안 함** — 시험을 구현 뒤에 씀. 대신 변이 m27(옮김 줄 삭제) CAUGHT(`../../mutation/mutation-run-3-wiring.txt`) | yes |
