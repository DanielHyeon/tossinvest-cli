# Branch Test Map: `Journal.MarkAlertDelivered`

Source: `internal/journal/outbox.go` (337-348). AST 기준 분기 1 / 이탈 2 /
defers 0 / go_statements 0.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:344` `ExecContext` 실패 | **없음** — 닫힌 DB 주입 테스트가 없다 | no | no |
| — `:347` | 정상 정산 → `PENDING` → `DELIVERED` | `outbox_test.go TestDeliveryAndAcknowledgementAreDistinctStates:108` | no | yes |
| — `:347` | **이미 `DELIVERED`인 행 → `ErrAlertNotFound`** | `outbox_test.go TestMarkingANonPendingAlertIsRefused:150` | no | yes |

분기 표의 행이 3개인데 AST 분기는 1개다. 아래 두 행은 `requireOneRow`가 만드는
**같은 이탈점의 두 결과**이고, 분기 ID를 새로 붙이지 않는다 — AST에 없는 분기를
표에 지어내지 않기 위해서다.

## a092가 이 함수에 대해 지는 것

편집하지 않으므로 새 RED가 없다. `TestMarkingANonPendingAlertIsRefused`가
**CAS가 실제로 CAS라는 것을 이미 관측하고 있고**, 17판 D0.3은 그 관측을 인용한다.

§6.0 R17-6(사이클당 1행 1시도)이 이 함수의 `attempts + 1`을 계수로 쓴다.
그 테스트는 이 함수의 분기가 아니라 **배달 루프가 이 함수를 몇 번 부르는지**를
관측하므로 위 표에 행을 만들지 않는다.

B1은 SQLite 실패 주입이 필요하고 a092의 범위가 아니다
(`not-applicable`: 이 change는 이 함수를 편집하지 않는다).
