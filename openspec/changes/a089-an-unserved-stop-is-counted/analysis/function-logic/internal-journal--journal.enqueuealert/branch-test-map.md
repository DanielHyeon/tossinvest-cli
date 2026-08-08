# Branch Test Map: `Journal.EnqueueAlert`

AST 기준 분기 9 / 이탈 9 / defer 1. 기존 테스트는 `internal/journal/outbox_test.go`.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:113` 이벤트 키 없음 | `TestEnqueueAlertRequiresAKeyAndAType` | no | yes |
| B2 | `:116` 이벤트 타입 없음 | 같은 테스트 | no | yes |
| B3 | `:122` `BeginTx` 오류 | **없음** | no | no |
| B4 | `:129` 기존 행 조회 분기 | 다수 | no | yes |
| B5 | `:130` **기존 행 발견 → 상태 무관하게 id 반환** | `outbox_test.go:45-61` (**PENDING 행만 검증**) | no | 부분 |
| B6 | `:132` `ErrNoRows` 아닌 조회 오류 | **없음** | no | no |
| B7 | `:140` `INSERT` 오류 | **없음** | no | no |
| B8 | `:144` `LastInsertId` 오류 | **없음** | no | no |
| B9 | `:147` `Commit` 오류 | **없음** | no | no |

## B5의 기존 테스트가 절반만 본다

`outbox_test.go:45-61`은 같은 키를 두 번 넣고 **행이 1개인지**와 **본문이 첫 관측인지**만
단언한다. 그 시나리오의 행은 계속 `PENDING`이므로 이후 `Mark*`가 성공한다.
**`DELIVERED`/`ACKNOWLEDGED` 행에 재발이 들어오는 경로는 테스트가 없다** — 8/5에 실제로
일어난 것이 그것이다.

그리고 `:56-61`은 갱신을 **금지하는 방향**으로 단언한다:

```go
// The first observation's text is kept: it is the one that describes what was
// seen when the condition arose.
if pending[0].Body != "first" { ... }
```

재발 반영 설계는 이 단언을 뒤집어야 하고, 주석이 제시하는 반대 근거("최초 관측 본문의
진단 가치")를 어떻게 보존할지 함께 정해야 한다.

## 필요한 RED

| # | Scenario | 기대 |
|---|---|---|
| R1 | `DELIVERED` 행에 같은 키 재발 | 행이 `PENDING`으로 복귀, `attempts` 누적 가능 |
| R2 | R1 이후 전달 실패 | `UndeliveredCount`에 잡히고 `PendingAlerts`에 뜬다 |
| R3 | `Acknowledge`가 재발 중인 행을 남긴 채 게이트를 풀지 않는다 | `remaining != 0` |
| R4 | `parkAlert` 경로(`replay.go:551`) | **재개방되지 않는다** — 전달자가 없으므로 |
| R5 | 첫 발생 | 행 개수·반환 id·본문 **무변화** (회귀) |
| R6 | `ACKNOWLEDGED` 행에 재발 | **미결정** — 범위에서 뺄지 결정 후 작성 |
