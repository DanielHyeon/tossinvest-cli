# Branch Test Map: `Journal.EnqueueAlert`

Source: `internal/journal/outbox.go` (111-151). AST 기준 분기 9 / 이탈 9 /
defers 1 / go_statements 0.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:113` `EventKey` 공백 → 오류 | `internal/journal` outbox 테스트 | no | yes |
| B2 | `:116` `Type` 공백 → 오류 | 같은 위 | no | yes |
| B3 | `:122` `BeginTx` 실패 | **없음** — 닫힌 DB를 주입하는 테스트가 없다 | no | no |
| B4 | `:129` switch 진입 | 아래 B5·B6이 대표 | — | — |
| B5 | `:130` **중복 키 → 기존 id 반환, 새 행 없음** | outbox 중복 테스트 | no | yes |
| B6 | `:132` 조회가 `ErrNoRows`가 아닌 오류 | **없음** | no | no |
| B7 | `:140` INSERT 실패 | **없음** | no | no |
| B8 | `:144` `LastInsertId` 실패 | **없음** | no | no |
| B9 | `:147` `Commit` 실패 | **없음** | no | no |

## a092가 이 함수에 대해 지는 것은 없다

편집하지 않으므로 새 RED가 없다. 이 표는 **`analysis/delivery-latency.md`의 해석이
B5에 근거한다**는 것을 남기기 위한 것이다.

미테스트 분기 5개(B3·B6·B7·B8·B9)는 전부 SQLite 실패 주입이 필요하고,
**a092의 범위가 아니다** — 기록해 두고 넘어간다(`not-applicable`: 이 change는
이 함수를 편집하지 않는다).
