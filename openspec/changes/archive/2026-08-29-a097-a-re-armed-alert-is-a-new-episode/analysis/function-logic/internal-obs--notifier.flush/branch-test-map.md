# Branch Test Map: `Notifier.Flush`

측정: `go test -covermode=set -coverprofile ./internal/obs/` RED 프로파일에서 블록 단위로
직접 읽었다. `-covermode=set`은 횟수를 세지 않는다.

**a097은 이 함수의 본문을 바꾸지 않는다.** 표가 필요한 이유는 proposal R1이 이 함수의
메시지 조립(`:409-414`)을 근거로 삼고, R4가 이 함수의 뮤텍스(`n.mu.Lock@398`)를
테스트로 고정하기 때문이다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:392` journal 미배선 | 없음 — 조립 실수 (`not-applicable`) | 미진입 (`392-394 count=0`) | GREEN 후 기재 |
| B2 | `:402` `PendingAlerts` 실패 | 없음 — 장애 주입 (`not-applicable`) | 미진입 (`402-404 count=0`) | GREEN 후 기재 |
| B3 | `:405` backlog 순회 | `TestRecoveredDeliveryDoesNotReleaseTheGateByItself`, `TestCriticalAlertSurvivesAProcessRestart` | 진입 (`405-406 count=1`) | GREEN 후 기재 |
| B4 | `:406` publisher 미배선 | 없음 — 조립 실수 (`not-applicable`) | 미진입 (`406-407 count=0`) | GREEN 후 기재 |
| B5 | `:415` flush 중 `Publish` 실패 | 없음 (범위 밖 — `MarkAlertAttemptFailed` 오류 버림은 a096 C4) | 미진입 (`415-417 count=0`) | GREEN 후 기재 |
| B6 | `:419` flush 중 `MarkAlertDelivered` 실패 | 없음 — 장애 주입 (`not-applicable`) | 미진입 (`419-421 count=0`) | GREEN 후 기재 |

**잠금 자체는 분기가 아니므로 이 표에 행이 없다.** 그것이 P2 ④의 형태다 —
`n.mu.Lock@398`은 커버리지가 항상 1이지만 그 사실은 잠금이 **필요한지**를 말하지 않는다.
`-covermode=set`으로는 이 결함을 볼 수 없고, 뮤테이션으로만 볼 수 있다.

## a097이 여기서 하는 것

`TestFlushCannotPublishBesideASend`(2.6, 신규)를 추가한다. publisher가 자기 재진입을 세고,
전송이 in-flight인 동안 Flush가 두 번째 `Publish`를 시작하면 실패한다. 판정은 경과 시간이
아니라 **동시 진입이라는 사건**이다.

검증(5.8)은 뮤테이션이다: `n.mu.Lock@398`/`Unlock@399`를 지우고 `./internal/obs/`를 돌려
**2.6만** 실패하는지 확인한다. 다른 테스트가 같이 실패하면 2.6이 잠금을 고정한 것이
아니라 다른 것을 깨뜨린 것이다.

RED 표가 보여주는 두 번째 사실: 이 함수는 행복 경로(B3) 말고는 어떤 분기도 테스트되지
않는다. a097은 그 미커버 분기들을 **채우지 않는다** — 조립 실수와 장애 주입이며 이
change의 목적 밖이다(`not-applicable`). 채우는 것은 잠금 하나다.
