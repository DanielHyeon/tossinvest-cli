# Branch Test Map: `Notifier.Flush`

측정: `go test -covermode=set ./internal/obs/...` — RED 84.8%, GREEN 85.4%(3판, a096b 반영).
a096이 이 함수에 더한 것은 **`n.mu` 하나**다. 루프도 SQL도 반환값도 그대로다.

RED 열은 **측정하지 않았다(`미측정`)**. 이유는 `Acknowledge`와 같다 — base 시점의
분기 대응을 이 함수에 대해서는 뜨지 않았다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | journal 미배선 `:368` | 기존 | 미측정 | 미진입 |
| B2 | `PendingAlerts` 오류 `:378`(루프 진입은 `:378`) | 없음 | 미측정 | 미진입 |
| B3 | publisher 미배선 → break `:381` | 없음 | 미측정 | 진입 |
| B4 | 전송 실패 → 시도 실패 기록 후 continue `:382` | 기존 flush 테스트 | 미측정 | 미진입 |
| B5 | 전달 표시 실패 → 즉시 반환 `:391` | 없음 | 미측정 | 미진입 |
| B6 | `PendingAlerts` 조회 오류 `:395` | 없음 | 미측정 | 진입 |

## 왜 잠금을 더했나

`Flush`는 `PendingAlerts`가 돌려준 행을 직접 발행한다. 1판까지 이 함수는 `n.mu`를 **전혀**
잡지 않았고, 그래서 관측 경로의 전송과 flush가 같은 행을 같은 순간에 보낼 수 있었다.
독립 리뷰 1라운드가 blocker 1의 일부로 지적했고, 확인 결과 리뷰가 지적한 것보다 넓었다.

## 배선에 관한 사실 하나

**`Notifier.Flush`에는 non-test 호출자가 없다.** 확인했다. 따라서
`execgw.Gateway.parkAlert`가 outbox에 넣는 행("the notifier's Flush picks the row up and
delivers it" — `replay.go:101`)은 현재 배선에서 아무도 집어 가지 않는다.

a096이 만든 문제도 고치는 문제도 아니다. 이 표는 그 사실을 기록만 하며, 별도 change가
필요하다(tasks 7.4).
