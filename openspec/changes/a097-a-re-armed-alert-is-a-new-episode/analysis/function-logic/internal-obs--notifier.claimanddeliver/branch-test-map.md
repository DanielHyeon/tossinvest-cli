# Branch Test Map: `Notifier.claimAndDeliver`

측정: `go test -covermode=set -coverprofile ./internal/obs/`.
**RED 85.4% → GREEN 86.7%.** `-covermode=set`은 **횟수를 세지 않고** 실행 여부만 0/1로
기록하므로 이 표에서 읽을 수 있는 것은 "그 경로가 발생했다/발생하지 않았다"뿐이다.

**RED의 핵심 관찰**: `B1`(claim 실패)은 `227-229 count=0` — **미진입**. 이 저장소의 어떤
테스트도 claim 실패 경로를 지난 적이 없었다.

**분기가 2개에서 4개로 늘었다.** a097이 그 분기 안에 `n.Log != nil`·`n.Gate != nil`
두 nil 가드를 넣었기 때문이다. 아래 줄번호는 구현 후 재생성한 `ast.json` 기준이고,
RED 칸의 줄번호는 구현 전 기준이다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:245` claim 트랜잭션이 실패한다 | `TestAClaimThatFailsBlocksNewEntries`, `TestAFailedClaimStillReturnsItsError` (a097 2.4) | **미진입** (`227-229 count=0`) | **진입** (`245-258 count=1`) |
| B2 | `:258` 실패를 적을 logger가 배선돼 있다 | 위와 같음 | 분기 없었음 | 진입 (`258-260 count=1`) |
| B3 | `:261` 잠글 gate가 배선돼 있다 | 위와 같음 | 분기 없었음 | 진입 (`261-263 count=1`) |
| B4 | `:266` 창 안이라 전송이 필요 없다 | `TestOneConditionIsOneSend` 등 a096 다수 | 진입 (`230-239 count=1`) | 진입 (`266-275 count=1`) |

정상 경로(`n.deliver`)는 a096의 전달 테스트 다수가 지난다.

## B1이 이 change의 측정점이다

RED에서 `count=0`이라는 것은 아무도 지나지 않은 분기였다는 뜻이고, 실제로 그 분기는
로그도 gate도 승격도 하지 않고 오류만 돌려주고 있었다. 그리고 호출자 하나는 그 오류를
버린다 — `internal/flatten/flatten.go:694`, 그쪽도 `count=0`이다.

**두 개의 0이 겹친 자리가 P2 ③이다.** GREEN에서 진입으로 바뀐 것이 R2의 직접 증거다.

RED에서 이 분기를 진입시키는 방법: `*journal.Journal`이 구체 타입이라 mock을 만들 수
없으므로 **열린 journal을 닫는다.** `BeginTx`가 실패해 `ClaimAlertForDelivery`의 B3으로
떨어지고 오류가 여기로 올라온다.

## nil 가드 둘은 분기를 늘리지만 위험을 늘리지 않는다

`n.Log`와 `n.Gate`는 둘 다 선택 배선이다(`Notifier` 구조체 주석). 가드 없이 부르면
미배선 조립에서 패닉하므로 가드가 필요하고, 그래서 분기가 둘 늘었다. 둘 다 GREEN에서
진입한다 — 2.4가 로그와 gate를 모두 배선하기 때문이다.

**미배선 쪽(`n.Log == nil`, `n.Gate == nil`)은 커버되지 않는다.** 조립 실수 경로이며 이
change의 완료 조건으로 삼지 않는다(`not-applicable`). 그 경우에도 오류가 그대로 반환되어
계약이 깨지지 않는다는 것은 `TestAFailedClaimStillReturnsItsError`가 따로 고정한다.
