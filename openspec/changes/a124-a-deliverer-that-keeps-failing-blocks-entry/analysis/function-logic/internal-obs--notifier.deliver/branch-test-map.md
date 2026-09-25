# Branch Test Map: `Notifier.deliver`

a124 는 이 함수를 편집하지 않는다 — 대조 근거 번들. 시험 이름은 `internal/obs/*_test.go` 의 함수 이름을
grep 으로 얻었고 분기 도달은 재지 않았다(tasks 1.4 에서 `covermode=set`).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B7 | 한 조건 한 발송 | `TestOneConditionIsOneSend` | no(회귀 핀) | 미측정 |
| B8 | 발송 중 승인 → 임차 상실은 게이트 사건 아님 | `TestAcknowledgeCannotClearTheGateMidSend` | no | 미측정 |
| B11·B12 | 발행됐으나 기록 못 함 → 잠금 | `TestASendThatCannotBeRecordedLatchesTheGate` | no | 미측정 |
| B14·B18 | 시도 사이 행 소실 → 잠금 | a099 회귀 핀 (`a099_regression_pins_test.go`) | no | 미측정 |
| B19·B20 | 시도 간 대기 후 재시도 | `TestAnUndeliveredConditionIsStillRetried` | no | 미측정 |
| B26·B27 | 소진 → 잠금 | `TestADeadTransportIsStillFoundAfterASuccessfulDelivery` | no | 미측정 |
| 나머지 | B1·B3·B9·B10·B13·B22·B25 | 1.4 에서 도달 여부 확인 — a124 의 편집 대상이 아니므로 미도달이어도 이 change 는 채우지 않는다(기록만) | ? | ? |
