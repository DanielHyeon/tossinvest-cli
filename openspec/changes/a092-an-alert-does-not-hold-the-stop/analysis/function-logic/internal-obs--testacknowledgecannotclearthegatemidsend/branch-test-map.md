# Branch Test Map: `TestAcknowledgeCannotClearTheGateMidSend`

- Source: `internal/obs/a096_one_send_per_condition_test.go`. 시험 코드 — 분기는 단언 갈래.

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | select at 385:2 | 단언 갈래 | 자기 자신 | 반증: 변이 L01(전송 위 잠금) · L14(셈을 잠금 밖) · L18(claim 을 잠금 밖) 각각 이 시험에서 FAIL | PASS(`fbc6df5f`) |
| B2 | select at 394:2 | 단언 갈래 | 자기 자신 | 반증: 변이 L01(전송 위 잠금) · L14(셈을 잠금 밖) · L18(claim 을 잠금 밖) 각각 이 시험에서 FAIL | PASS(`fbc6df5f`) |
| B3 | if at 396:3 | 단언 갈래 | 자기 자신 | 반증: 변이 L01(전송 위 잠금) · L14(셈을 잠금 밖) · L18(claim 을 잠금 밖) 각각 이 시험에서 FAIL | PASS(`fbc6df5f`) |
| B4 | if at 402:2 | 단언 갈래 | 자기 자신 | 반증: 변이 L01(전송 위 잠금) · L14(셈을 잠금 밖) · L18(claim 을 잠금 밖) 각각 이 시험에서 FAIL | PASS(`fbc6df5f`) |
| B5 | if at 407:2 | 단언 갈래 | 자기 자신 | 반증: 변이 L01(전송 위 잠금) · L14(셈을 잠금 밖) · L18(claim 을 잠금 밖) 각각 이 시험에서 FAIL | PASS(`fbc6df5f`) |
| B6 | if at 410:2 | 단언 갈래 | 자기 자신 | 반증: 변이 L01(전송 위 잠금) · L14(셈을 잠금 밖) · L18(claim 을 잠금 밖) 각각 이 시험에서 FAIL | PASS(`fbc6df5f`) |
| B7 | if at 413:2 | 단언 갈래 | 자기 자신 | 반증: 변이 L01(전송 위 잠금) · L14(셈을 잠금 밖) · L18(claim 을 잠금 밖) 각각 이 시험에서 FAIL | PASS(`fbc6df5f`) |
