# Function Logic Map: `Gateway.submit`

- Source: `internal/execgw/gateway.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## 이 change가 한 일

노출을 늘리는 mutation에 대한 보호 검사 호출이 entry gate 앞에 추가됐다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| plan.raisesExposure | true/false — intent의 side에서 계산 | gateway.go:338/376/407 | true이고 보호 미배선이면 거부 |
| g.protection() | WIRED | UNWIRED | execgw.defaultProtection 상수 또는 Options override | UNWIRED가 fail-closed 방향 |

## Branches and early returns

추가된 분기는 checkProtection의 반환이 nil이 아닌 경우 하나다. 나머지는 기존 거부·전송 분기.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if` @ internal/execgw/gateway.go:454 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B2 | `if` @ internal/execgw/gateway.go:467 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B3 | `if` @ internal/execgw/gateway.go:473 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B4 | `else` @ internal/execgw/gateway.go:475 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B5 | `if` @ internal/execgw/gateway.go:475 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B6 | `if` @ internal/execgw/gateway.go:483 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B7 | `else` @ internal/execgw/gateway.go:485 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B8 | `if` @ internal/execgw/gateway.go:485 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B9 | `if` @ internal/execgw/gateway.go:493 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B10 | `if` @ internal/execgw/gateway.go:509 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B11 | `if` @ internal/execgw/gateway.go:514 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B12 | `if` @ internal/execgw/gateway.go:520 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B13 | `if` @ internal/execgw/gateway.go:521 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B14 | `if` @ internal/execgw/gateway.go:530 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B15 | `if` @ internal/execgw/gateway.go:533 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B16 | `if` @ internal/execgw/gateway.go:540 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B17 | `if` @ internal/execgw/gateway.go:558 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B18 | `if` @ internal/execgw/gateway.go:561 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B19 | `if` @ internal/execgw/gateway.go:564 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B20 | `if` @ internal/execgw/gateway.go:576 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B21 | `if` @ internal/execgw/gateway.go:580 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B22 | `if` @ internal/execgw/gateway.go:594 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B23 | `if` @ internal/execgw/gateway.go:597 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| g.checkProtection | 조항 6을 mutation 단위로 판정 | 오류 없음 — *RejectedError 또는 nil | protection.go |
| g.refuse | 거부를 attempt에 기록하고 Outcome을 만든다 | journal 오류는 그대로 전파 | 기존 경로 |

## State mutations and fallbacks

- 거부 시 attempt가 NOT_DISPATCHED로 기록된다. 브로커는 호출되지 않는다.

## Safety conclusion

- Safe edit boundary: 새 거부 하나를 기존 거부 사슬의 앞에 넣었다. 통과 경로는 무수정.
- High-risk impact: yes
