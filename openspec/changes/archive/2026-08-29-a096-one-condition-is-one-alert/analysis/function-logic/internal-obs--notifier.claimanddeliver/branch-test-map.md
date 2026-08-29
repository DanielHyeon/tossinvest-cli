# Branch Test Map: `Notifier.claimAndDeliver`

측정: `go test -covermode=set ./internal/obs/...` — GREEN 85.4%(3판, a096b 반영). a096 2판이 만든 함수이므로 RED에 없다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | claim이 오류 `:227` | 없음 — 원장 쓰기 실패 주입 하니스 없음 | 없음 | 미진입 |
| B2 | `!owed` — 보내지 않고 반환 `:230` | `TestOneConditionIsOneSend`, `TestTheSameConditionIsRemindedOncePerWindow`, `TestSuppressingTheSendKeepsTheRecord` | 없음 | 진입 |

## 이 함수의 존재 이유가 곧 그 테스트다

`TestConcurrentObservationsOfOneConditionSendOnce`가 이 함수의 **경계**를 검증한다.
분기가 아니라 `n.mu`가 claim과 send를 함께 감싸는지가 대상이므로 분기 표에 칸이 없다.
`-race`로 8개 goroutine이 같은 key를 동시에 관측해 전송이 1회여야 통과한다.

1판에서는 claim이 이 잠금 **밖**에 있었고, 그래서 두 관측이 같은 미전달 행을 읽고 둘 다
보낼 수 있었다. 독립 리뷰 1라운드 blocker 1이 그것이다.
