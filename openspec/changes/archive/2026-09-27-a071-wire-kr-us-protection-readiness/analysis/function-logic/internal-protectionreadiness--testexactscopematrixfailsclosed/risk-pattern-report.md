# Risk Pattern Report: `internal/protectionreadiness/attestation_test.go`

| Rule | Location | Message |
|---|---|---|
| go-panic | `internal/protectionreadiness/attestation_test.go:321` | panic can bypass normal error and shutdown handling; map the recovery boundary. |
| go-panic | `internal/protectionreadiness/attestation_test.go:325` | panic can bypass normal error and shutdown handling; map the recovery boundary. |
| go-panic | `internal/protectionreadiness/attestation_test.go:331` | panic can bypass normal error and shutdown handling; map the recovery boundary. |

> Findings are review candidates, not automatic defect verdicts.

## 함수 범위 대조 (task 5.1, 42-74)

- 범위 안 결과 0건 — 위 결과는 모두 같은 파일의 다른 함수임.
