# Branch Test Map: `Notifier.deliver`

측정: `go test -covermode=set ./internal/obs/...` — RED 84.8%, GREEN 85.4%(3판, a096b 반영).
a096은 이 함수의 **본문을 바꾸지 않았다.** 바뀐 것은 잠금의 소유자다 — 1판까지 이 함수가
`n.mu`를 잡았고, 2판은 호출자(`claimAndDeliver`)가 claim부터 함께 잡는다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `Attempts` 미설정 → 기본 3 `:307` | 없음 (헬퍼가 항상 3을 넣는다) | 미진입 | 미진입 |
| B2 | 재시도 루프 진입 `:313` | 기존 다수 | 진입 | 진입 |
| B3 | publisher 미배선 `:314` | 없음 | 미진입 | 미진입 |
| B4 | `Publish` 성공 `:319` | `TestCriticalAlertIsDurableBeforeItIsSent`, `TestASendThatCannotBeRecordedLatchesTheGate` 등 | 진입 | 진입 |
| B5 | 전달 기록까지 성공 `:321` | `TestCriticalAlertIsDurableBeforeItIsSent` 등 | 진입 | 진입 |
| B6 | 전송은 성공했지만 기록 실패 로그 `:337` | `TestASendThatCannotBeRecordedLatchesTheGate` | 없음 | 진입 |
| B7 | 전송은 성공했지만 기록 실패 gate 차단 `:342` | `TestASendThatCannotBeRecordedLatchesTheGate` | 없음 | 진입 |
| B8 | 전송 실패 기록조차 실패하고 logger가 배선됨 `:348` | 없음 | 미진입 | 미진입 |
| B9 | 마지막이 아닌 시도 `:351` | `TestPersistentDeliveryFailureBlocksEntries` | 진입 | 진입 |
| B10 | ctx 종료로 대기 중단 `:352` | 없음 | 미진입 | 미진입 |
| B11 | 재시도 예산 소진 로그 `:362` | `TestPersistentDeliveryFailureBlocksEntries` | 진입 | 진입 |
| B12 | 재시도 예산 소진 gate 차단 `:367` | 위와 같음 + `TestADeadTransportIsStillFoundAfterASuccessfulDelivery` | 진입 | 진입 |

## B5가 이 change의 측정점이다

B5 직후의 실패 경로는 `MarkAlertDelivered`가 오류를 돌려줬을 때만 들어간다. 그 오류의 실질 경로는
`WHERE id = ? AND state = 'PENDING'`이 0행을 갱신하는 것 — 즉 **이미 전달된 행에 다시
전달 표시를 시도**하는 것이고, 그 시점에 `Publish`는 이미 성공한 뒤다(B4가 참).
운영 로그의 `journal: no such alert: 14 (or it is no longer pending)`가 같은 지점이다.

RED 측정값 `notifier.go:258.88,260.5 1 1`(진입). 진입 테스트는 격리 측정으로
`TestNotifierIsConcurrencySafe`로 특정했다.

`-covermode=set`은 블록이 실행됐는지를 0/1로 기록하며 **횟수를 세지 않는다.**
따라서 이 칸에서 읽을 수 있는 것은 "그 경로가 발생했다/발생하지 않았다"뿐이고,
몇 번 발생했는지가 아니다. 1판 문서가 이 칸을 횟수로 읽었고 독립 리뷰가 그것을 지적했다.

## 미커버 분기에 대한 판단

B1·B3·B8·B10은 조립 실수(Attempts 미설정, Publisher 미배선)와 원장/ctx 실패 주입이다.
a096의 핵심 변경은 claim과 send를 하나의 임계 구역으로 묶고, 전송 성공 후 기록
실패(B6·B7)를 불안전으로 처리하는 것이다. 나머지 미커버 조립/장애 주입 분기는
이 change의 완료 조건으로 삼지 않는다(`not-applicable`: 변경 목적 밖의 오류 주입).
