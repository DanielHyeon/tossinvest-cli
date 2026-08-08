# Branch Test Map: `Notifier.Notify`

측정: `go test -covermode=set ./internal/obs/...`
RED = base `ec29dc72` (84.8%), GREEN = a096 적용 후 (84.8%).
a096은 이 함수를 편집하지 않았다. 줄 번호도 그대로다(편집 지점이 뒤에 있다).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | normal 등급은 best-effort로 빠진다 `:128` | `TestOrdinaryAlertsAreBestEffort` | 진입 | 진입 |
| — | critical 경로(else) | `TestCriticalAlertIsDurableBeforeItIsSent` 등 | 진입 | 진입 |

전 분기 커버. 미커버 분기 없음.

## a096이 이 함수에서 의존하는 사실

`logEvent`(:126)가 B1(:128)보다 **앞**에 있다는 배치. 이 순서 때문에 전송을 억제해도
구조화 로그는 관측마다 남는다. 이 배치가 바뀌면 a096의 spec 문장("억제되는 것은 전송뿐이며
관측 사실의 기록이 아니다")이 거짓이 되므로, 이 표는 그 배치의 고정을 기록한다.

`TestSuppressingTheSendKeepsTheRecord`가 그것을 직접 센다: 관측 5회 → 전송 1회, 로그 5줄.
RED에서는 전송 5회, 로그 9줄이었다(관측 5 + `alert_undelivered` 4).
