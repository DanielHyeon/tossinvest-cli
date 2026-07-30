# Function Logic Map: `TestGatewayKeepsExitsOpenUnderAMismatch`

- Source: `internal/reconcile/mismatch_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## 이 change가 한 일

게이트웨이 생성에 ProtectionOverrideForTest를 더했다. 이 테스트는 불일치 하에서 매수가 막히고 매도가 열리는지를 보며, 조항 6이 먼저 걸리면 그 구분을 관측할 수 없다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 테스트 픽스처 | 함수 본문이 세운다 | 테스트 자신 | 단언 실패 시 t.Error/t.Fatal |

## Branches and early returns

분기는 전부 단언 가드다. 각 가드는 실패 시 이 테스트 자신이 보고한다.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if` @ internal/reconcile/mismatch_test.go:512 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B2 | `if` @ internal/reconcile/mismatch_test.go:537 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B3 | `if` @ internal/reconcile/mismatch_test.go:544 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B4 | `if` @ internal/reconcile/mismatch_test.go:547 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B5 | `if` @ internal/reconcile/mismatch_test.go:550 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B6 | `if` @ internal/reconcile/mismatch_test.go:560 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B7 | `if` @ internal/reconcile/mismatch_test.go:563 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B8 | `if` @ internal/reconcile/mismatch_test.go:570 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B9 | `if` @ internal/reconcile/mismatch_test.go:583 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B10 | `if` @ internal/reconcile/mismatch_test.go:586 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B11 | `if` @ internal/reconcile/mismatch_test.go:592 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| 테스트 대상 API | 이 테스트가 무엇을 보는지 | 단언으로 처리 | AST |

## State mutations and fallbacks

- 격리된 임시 디렉터리와 임시 journal 외에는 없다.

## Safety conclusion

- Safe edit boundary: 테스트 함수. 프로덕션 동작을 바꾸지 않는다.
- High-risk impact: no
