# Function Logic Map: `TestGatewayRefusesNewOrdersUntilRecoveryCompletes`

- Source: `internal/reconcile/recovery_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## 이 change가 한 일

동상. 복구 미완 사유로 거부되는지를 보려면 보호 사유가 앞서 걸리지 않아야 한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 테스트 픽스처 | 함수 본문이 세운다 | 테스트 자신 | 단언 실패 시 t.Error/t.Fatal |

## Branches and early returns

분기는 전부 단언 가드다. 각 가드는 실패 시 이 테스트 자신이 보고한다.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if` @ internal/reconcile/recovery_test.go:439 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B2 | `if` @ internal/reconcile/recovery_test.go:444 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B3 | `if` @ internal/reconcile/recovery_test.go:467 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B4 | `if` @ internal/reconcile/recovery_test.go:477 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B5 | `if` @ internal/reconcile/recovery_test.go:480 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B6 | `if` @ internal/reconcile/recovery_test.go:508 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B7 | `if` @ internal/reconcile/recovery_test.go:511 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B8 | `if` @ internal/reconcile/recovery_test.go:514 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B9 | `if` @ internal/reconcile/recovery_test.go:517 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B10 | `if` @ internal/reconcile/recovery_test.go:520 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B11 | `if` @ internal/reconcile/recovery_test.go:525 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B12 | `if` @ internal/reconcile/recovery_test.go:530 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B13 | `if` @ internal/reconcile/recovery_test.go:535 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| 테스트 대상 API | 이 테스트가 무엇을 보는지 | 단언으로 처리 | AST |

## State mutations and fallbacks

- 격리된 임시 디렉터리와 임시 journal 외에는 없다.

## Safety conclusion

- Safe edit boundary: 테스트 함수. 프로덕션 동작을 바꾸지 않는다.
- High-risk impact: no
