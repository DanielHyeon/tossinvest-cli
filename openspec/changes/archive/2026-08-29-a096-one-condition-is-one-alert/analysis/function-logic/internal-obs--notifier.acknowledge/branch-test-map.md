# Branch Test Map: `Notifier.Acknowledge`

측정: `go test -covermode=set ./internal/obs/...` — RED 84.8%, GREEN 85.4%(3판, a096b 반영).
a096이 더한 것은 `n.mu` 하나이며 분기는 바뀌지 않았다.

RED 열은 **측정하지 않았다(`미측정`)**. base `ec29dc72` 시점의 커버리지는 이 change가
그때 대상으로 삼았던 다섯 함수만 분기 대응을 떴고, 이 함수는 2판에서야 편집 대상이 됐다.
값을 추정해 적지 않는다 — 추정한 칸은 측정한 칸과 구별되지 않는다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 신원 없는 해제 `:417` | `TestAcknowledgeRequiresAnIdentity` | 미측정 | 진입 |
| B2 | journal 미배선 `:420` | 없음 | 미측정 | 진입 |
| B3 | journal 없이 gate만 푼다 `:421` | 없음 | 미측정 | 미진입 |
| B4 | id 미지정 → PENDING 전부 `:430` | 기존 acknowledge 테스트 | 미측정 | 진입 |
| B5 | `PendingAlerts` 오류 `:432` | 없음 | 미측정 | 미진입 |
| B6 | id 순회 `:435` | 기존 | 미측정 | 진입 |
| B7 | ack 오류가 not-found가 아님 `:439` | 없음 | 미측정 | 진입 |
| B8 | (같은 조건의 두 번째 항) `:440` | — | 미측정 | — |
| B9 | `UndeliveredCount` 오류 `:447` | 없음 | 미측정 | 미진입 |
| B10 | 남은 게 0이면 gate 해제 `:450` | `TestAcknowledgeWhileStillPendingKeepsTheBlock`, `TestRecoveredDeliveryDoesNotReleaseTheGateByItself` | 미측정 | 진입 |

## 잠금은 분기가 아니라 경계다

`TestAcknowledgeCannotClearTheGateMidSend`가 검증하는 것은 분기가 아니라 **상호 배제**다.
전송이 `Publish` 안에서 멈춰 있는 동안 `Acknowledge`가 완료되어서는 안 된다 — 완료된다면
그것은 세기 전의 세계를 근거로 gate를 푼 것이다.

이 위험은 a096이 키웠다. 이전에는 세기와 풀기 사이에 새 PENDING이 생기는 유일한 원천이
"새 조건의 첫 알림"이었는데, 재무장이 두 번째 원천을 만들었다.

## 미커버 분기에 대한 판단

B2·B3·B5·B7·B9는 조립 실수(journal 미배선)와 원장 실패 주입이다. a096은 이 분기들의
조건도 부작용도 바꾸지 않았다. `not-applicable`: 실패 주입 하니스 부재.
