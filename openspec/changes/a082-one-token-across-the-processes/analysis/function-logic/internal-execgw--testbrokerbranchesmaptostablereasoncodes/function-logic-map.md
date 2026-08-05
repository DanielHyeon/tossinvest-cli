# Function Logic Map: `TestBrokerBranchesMapToStableReasonCodes`

- Source: `internal/execgw/failclosed_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

이 함수는 이 change가 **고치는 대상이 아니라 넓힌 것**이다. `isAuthRefusal`(failclosed.go:210)이 fail-closed 거부 분류의 입력을 판정하는데, 그 fixture가 전부 맨몸 sentinel이라 a082가 바꾼 실제 입력 모양을 아무도 검사하지 않았다. 독립 리뷰 P1-6.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 표의 각 행 | 오류 하나와 기대 판정 하나 | 같은 함수 | 행이 빠지면 그 모양이 검사되지 않는다 |
| 불변식 | **판정은 `errors.Is`로 한다** | 대상 코드가 그렇게 판정한다 | 감싼 오류를 못 받으면 a082 이후의 실제 입력을 놓친다 |

a082 이전에는 인증 sentinel이 맨몸으로 도착했고 이 표의 fixture도 맨몸이었다.
이제 상태 코드를 실어 감싸서 도착하므로, **감싼 모양이 표에 있어야** 대상이
unwrap을 멈춰도 알아차린다. 행을 더할 뿐 기존 행은 글자 그대로 보존한다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 표 순회 — 행이 늘었다 | 없음 | 없음 | 자기 자신 |
| B2~B7 | 기존 판정·보고 갈래 | 없음 | `t.Errorf`/`t.Fatalf` | 자기 자신 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| 대상 분류 함수 | 판정 | 순수 | AST calls |
| `fmt.Errorf` | **신규** — 감싼 fixture 생성 | 순수. `%w` | AST calls |

## State mutations and fallbacks

- 상태를 바꾸지 않는다. 표를 돌며 판정만 한다.
- fallback 없음.

## Safety conclusion

- Safe edit boundary: **표에 행을 더하는 것뿐.** 기존 행, 판정 구조, 하네스는
  글자 그대로 보존한다.
- 판정력을 약화하지 않는다 — 행이 늘었다.
- High-risk impact: **no.** 테스트 전용. 다만 이 표가 지키는 대상은 High-risk다
  (인증 거부 → 엔트리 게이트 차단).
