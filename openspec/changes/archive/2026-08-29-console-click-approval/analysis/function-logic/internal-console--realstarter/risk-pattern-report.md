# Risk Pattern Report: `realStarter`

- Source: `internal/console/console_test.go`

Run:

```bash
python3 tools/logic-map/risk_pattern_report.py internal/console/console_test.go
```

## Findings

| Rule | Location | Classification | Function Logic Map link |
|---|---|---|---|
| (없음) | `internal/console/console_test.go` | not-applicable | 테스트 함수 — 실계좌 경로 아님 |

판정: 이 change의 편집은 승인의 **형식**만 바꾼다. 주문 전송·취소·상한·계획 인가·프로세스 경계는
이 함수들 밖에 있고 무변경이다.
