# Function Logic Map: `decimalText`

- Source: `internal/console/portfolio.go`
- Change: `console-operator-overview`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

이 change의 diff가 이 함수의 본문을 바꿨다. 아래 분석은 현재 HEAD 본문에 대한 것이다.

브로커 float64를 화면 문자열로. 이 change가 바꾼 것은 3자리 구분 로직을 `groupDecimalText`로 뽑아 위임한 것뿐이며, 출력은 동일하다. 뽑아낸 이유는 개요가 원장의 **decimal 문자열**을 math/big으로 합산한 결과에 같은 구분을 적용해야 했기 때문이다 — float64를 경유하면 원장이 동결한 R 옆에 그것과 어긋나는 숫자가 놓인다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `v` | float64 | `domain.Position` | 해당 없음 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| (무분기) | — | 본문에 분기가 없다 | 단일 경로 | `TestThePositionsScreenShowsTheExitLineOfAManagedPosition`의 값 대조 + `TestABrokerValueThatIsNotADecimalIsPrintedVerbatim`(원문 보존 쪽 경계) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strconv.FormatFloat(v, 'f', -1, 64)` | 지수 표기 없이, 후행 0 없이 | 순수 변환 | ast.json calls |
| `groupDecimalText` | 정수부 3자리 구분 | 순수 문자열 변환 | portfolio.go:732 |
| (금지 바인딩) | 계좌·원장 무접촉 | 순수 함수 | ast.json calls |

## State mutations and fallbacks

- 없음(순수 함수). 가격에 대한 산술은 하지 않는다 — 표시 전용이다.

## Safety conclusion

- Safe edit boundary: 구분 로직의 위임. 입력·출력 계약 무변경.
- High-risk impact: no (렌더 전용, 산술 없음). 원장 쪽 화면은 원장이 저장한 decimal 문자열을 그대로 쓰고 이 경로를 타지 않는다.
