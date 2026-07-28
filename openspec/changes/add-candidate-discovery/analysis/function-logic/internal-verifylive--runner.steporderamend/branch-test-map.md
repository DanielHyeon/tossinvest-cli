# Branch Test Map: `Runner.stepOrderAmend`

fakeBroker 하네스로 실행한다 — 실계좌·실브로커 무접촉.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 스펙 산출 실패 시 주문 없이 중단 | `runner_test.go` | — (무변경) | yes |
| B2 | 미승인·상한 초과는 요청 전 거부 | `plan_test.go`, `confirm_test.go` | — | yes |
| B3 | 정정 실패가 남긴 주문을 다음 실행이 정리한다 | `cleanup_test.go` | — | yes |
| B4 | US면 `price-only`, KR이면 `price+quantity`로 기록된다 | `TestTheRecordDoesNotCallAUSRequestAKROne`(US: `quantity` 문자열 부재 단언) + `TestTheKRRecordStillSaysKR`(KR: `quantity` 존재 단언) | yes (고정 문자열 시절 US 실행이 `KR price+quantity`를 적었다) | yes |
| B5 | 정정이 새 번호를 냈는지 기록한다 | `record_test.go`, `us_market_test.go` | — | yes |
| B6 | 원본 id의 상태를 읽어 기록한다 | 동상 | — | yes |
| B7 | 원본 id를 못 읽으면 `unreadable`로 기록한다 | 동상 | — | yes |
| B8 | 현재 id의 상태·가격을 기록한다 | 동상 | — | yes |
