# Risk Pattern Report: `internal/execgw/gateway.go`

| Rule | Location | Message |
|---|---|---|
| go-panic | `internal/execgw/gateway.go:1108` | panic can bypass normal error and shutdown handling; map the recovery boundary. |

> Findings are review candidates, not automatic defect verdicts.

## 함수 범위 대조 (task 5.1, 148-189)

- 범위 안 결과 0건 — 위 결과는 모두 같은 파일의 다른 함수임.
