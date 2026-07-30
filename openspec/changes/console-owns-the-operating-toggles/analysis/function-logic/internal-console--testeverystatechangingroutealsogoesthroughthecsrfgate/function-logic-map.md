# Function Logic Map: `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate`

- Source: `internal/console/static_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## 이 change가 한 일

상태변경 목록에 /settings/trading과 /settings/gate를 더했다. 스펙이 문장과 정적 검사 목록이 같은 커밋에서 움직일 것을 요구한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 테스트 픽스처 | 함수 본문이 세운다 | 테스트 자신 | 단언 실패 시 t.Error/t.Fatal |

## Branches and early returns

분기는 전부 단언 가드다. 각 가드는 실패 시 이 테스트 자신이 보고한다.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `range` @ internal/console/static_test.go:382 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B2 | `switch` @ internal/console/static_test.go:384 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B3 | `case` @ internal/console/static_test.go:385 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B4 | `case` @ internal/console/static_test.go:387 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B5 | `range` @ internal/console/static_test.go:391 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B6 | `if` @ internal/console/static_test.go:392 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| 테스트 대상 API | 이 테스트가 무엇을 보는지 | 단언으로 처리 | AST |

## State mutations and fallbacks

- 격리된 임시 디렉터리와 임시 journal 외에는 없다.

## Safety conclusion

- Safe edit boundary: 테스트 함수. 프로덕션 동작을 바꾸지 않는다.
- High-risk impact: no
