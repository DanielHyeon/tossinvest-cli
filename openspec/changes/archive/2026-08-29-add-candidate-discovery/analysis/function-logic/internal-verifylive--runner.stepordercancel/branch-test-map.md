# Branch Test Map: `Runner.stepOrderCancel`

fakeBroker 하네스로 실행한다 — 실계좌·실브로커 무접촉.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 스펙 산출 실패 시 주문 없이 중단 | `runner_test.go`의 중단·거부 케이스 | — (무변경) | yes |
| B2 | 미승인·상한 초과는 요청 전 거부 | `plan_test.go`(`Plan.Authorises` 거부 표), `confirm_test.go` | — | yes |
| B3 | 배치 후 상태·TIF가 기록된다 | `TestTheRecordDoesNotCallAUSRequestAKROne`(US), `TestTheKRRecordStillSaysKR`(KR) — `order.place.ok` detail | yes (US 실행이 detail에 `KR`을 적었다) | yes |
| B4 | 취소 실패가 남긴 주문을 다음 실행이 정리한다 | `cleanup_test.go`(M16 재현: fail → redo → pass) | — | yes |
| B5 | 취소 후 상태·`canceledAt`·체결수량이 기록된다 | `runner_test.go`/`record_test.go` | — | yes |
