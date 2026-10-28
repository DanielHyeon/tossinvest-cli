# Branch Test Map: `a165BoundaryBroker.CancelPendingOrder`

| 분기 | 시나리오 | 회귀 계약 |
|---|---|---|
| B1 | 첫 cancel | 실제 `firstDelay` 뒤 AP 409 반환 |
| B2 | firstDelay 존재 | `time.Sleep`은 Fake Clock을 전진시키지 않음 |
| 반환 2 | retry cancel | window context deadline을 관찰하고 ACK 반환 |
