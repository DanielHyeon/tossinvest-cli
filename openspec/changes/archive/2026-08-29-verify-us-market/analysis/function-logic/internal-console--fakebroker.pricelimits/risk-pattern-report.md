# Risk Pattern Report: `fakeBroker.PriceLimits`

- Source: `internal/console/fake_broker_test.go`

Run:

```bash
python3 tools/logic-map/risk_pattern_report.py internal/console/fake_broker_test.go
```

## Findings

| Rule | Location | Classification | Function Logic Map link |
|---|---|---|---|
| (없음) | `internal/console/fake_broker_test.go` | not-applicable | `function-logic-map.md` Safety conclusion |

판정: 이 change는 주문의 **대상 시장**을 매개변수로 만든다. 1주·체결 불가 지정가·단계 내 취소라는
노출 규칙과 승인 레일은 무변경이다.
