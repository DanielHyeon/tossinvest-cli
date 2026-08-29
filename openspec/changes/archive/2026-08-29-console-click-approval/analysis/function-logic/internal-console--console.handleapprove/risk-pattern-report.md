# Risk Pattern Report: `Console.handleApprove`

- Source: `internal/console/pages.go`

Run:

```bash
python3 tools/logic-map/risk_pattern_report.py internal/console/pages.go
```

## Findings

| Rule | Location | Classification | Function Logic Map link |
|---|---|---|---|
| (없음) | `internal/console/pages.go` | reviewed-safe | `function-logic-map.md` Safety conclusion |

판정: 이 change의 편집은 승인의 **형식**만 바꾼다. 주문 전송·취소·상한·계획 인가·프로세스 경계는
이 함수들 밖에 있고 무변경이다.
