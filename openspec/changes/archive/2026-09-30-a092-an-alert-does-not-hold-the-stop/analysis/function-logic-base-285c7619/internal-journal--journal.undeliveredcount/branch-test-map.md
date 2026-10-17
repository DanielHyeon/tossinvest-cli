# Branch Test Map: `Journal.UndeliveredCount`

Source: `internal/journal/outbox.go` (408-415). AST 기준 분기 1 / 이탈 2 /
defers 0 / go_statements 0.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:410` 조회/스캔 실패 | **없음** — 닫힌 DB 주입 테스트가 없다 | no | no |
| — `:414` | 배달·승인된 행은 세지 않는다 (0 방향) | `outbox_test.go TestDeliveryAndAcknowledgementAreDistinctStates:122`, `internal/obs/measurement_test.go:131`(`want 0`), `internal/obs/mode_test.go:74`(`want 0`) | no | yes |
| — `:414` | **미전달이 남으면 0이 아니다** | `internal/obs/a096_one_send_per_condition_test.go:277`(`want 1 — a failed send is preserved`), `a096b_round2_test.go:77`(`want 1 — the mark failed, so the row is still owed`) | no | yes |

AST 분기는 1개이고 아래 두 행은 같은 이탈점(`:414`)의 서로 다른 결과다.

## a092가 이 함수에 대해 지는 것

편집하지 않으므로 이 함수에 대한 새 RED는 없다. 다만 **§6.0 R17-10**(배달 루프가
감독 아래 있다)과 **R17-11**(`Acknowledge`가 게이트를 푼다)이 이 수를 판정에
쓰므로, 이 함수가 프로덕션 경로에 처음 들어가는 시점이 a092다.

B1은 SQLite 실패 주입이 필요하고 a092의 범위가 아니다
(`not-applicable`: 이 change는 이 함수를 편집하지 않는다).
