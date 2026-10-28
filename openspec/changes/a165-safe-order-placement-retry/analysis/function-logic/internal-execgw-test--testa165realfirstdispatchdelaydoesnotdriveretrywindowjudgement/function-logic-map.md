# Function Logic Map: `TestA165RealFirstDispatchDelayDoesNotDriveRetryWindowJudgement`

- Source: `internal/execgw/a165_cancel_refusal_boundary_test.go` (85-107)
- Revision: D5 4판 GREEN, `source_sha256` `e728c257f29ffaedbf547080c55c64c836b900c8859895bd4563a1d2d276883e`
- AST evidence: `ast.json` — AST 분기 4, 반환 0, 호출 11.

## 역할

첫 broker 호출의 실제 25ms 지연이 Fake Clock 10ms 창을 소진시키지 않음을 고정함. D5의 판정은 `g.clk` 하나이고, retry sleep 1ms만 Fake Clock을 움직이므로 재전송은 CONFIRMED로 정착해야 함.

| 분기 | 계약 |
|---|---|
| B1 | 결과는 오류 없이 `StateConfirmed` |
| B2 | broker 호출은 첫 409 + retry ACK, 정확히 2회 |
| B3 | journal record도 `StateConfirmed` |
| B4 | Fake Clock 경과는 retry-after 1ms뿐 |

실시간 전송 상한이 먼저 끝나는 별도 갈래는 `TestA165RealTimeoutKeepsFirstAmbiguousAsCancelled`가 첫 AMBIGUOUS/`stop=cancelled outcome=today`로 고정함.
