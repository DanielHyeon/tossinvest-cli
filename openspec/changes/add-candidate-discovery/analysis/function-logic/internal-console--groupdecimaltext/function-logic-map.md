# Function Logic Map: `groupDecimalText`

- Source: `internal/console/portfolio.go`
- Change: `add-candidate-discovery`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

이 함수는 이 change의 base commit 이후에 **신설**됐다(base에 같은 이름의 선언이 없다). gate가 evidence를 요구하는 것은 diff hunk가 현재 본문과 교차하기 때문이며, 아래 분석은 현재 HEAD 본문에 대한 것이다.

console-operator-overview가 `decimalText`에서 뽑아낸 순수 문자열 함수. 개요가 원장의 decimal 문자열을 `math/big`으로 합산한 뒤 같은 구분을 적용해야 해서 float64 경유 경로와 분리했다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `text` | 부호 + 정수부 + 선택적 소수부의 평문 10진 문자열 | `strconv.FormatFloat` 또는 `math/big.Rat` 렌더 | 지수 표기(`1e+06`)를 받으면 `e+06`을 소수부로 취급한다 — 두 호출자 모두 지수 표기를 만들지 않는다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `strings.HasPrefix(text, "-")` | 부호 분리 | 없음(계속) | 음수 실현손익 렌더 — `TestOnlyOneGuardianAxisIsComputedAndTheRestSayWhyNot`의 손실 축 |
| B2 | `for i, r := range whole` | 정수부 순회 | 없음 | 모든 숫자 렌더 경로 |
| B3 | `i > 0 && (len(whole)-i)%3 == 0` | `,` 삽입 | 없음 | 4자리 이상 값의 렌더(`TestThePositionsScreenShowsTheExitLineOfAManagedPosition`의 평가액) |
| B4 | `frac != ""` | `.` + 소수부 재부착 | 없음 | 소수 단가 렌더(US 종목 가격) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.HasPrefix` / `strings.Cut` / `strings.Builder` | 문자열 분해·재조립 | 순수 | ast.json calls |
| (금지 바인딩) | 계좌·원장·브로커 무접촉 | 순수 함수 | ast.json calls |

## State mutations and fallbacks

- 없음(순수 함수).
- 정수부만 구분한다 — 소수부는 원문 그대로 붙는다.

## Safety conclusion

- Safe edit boundary: 신설(추출). `decimalText`가 하던 것과 동일한 출력이며, 새 호출자는 원장의 decimal 문자열 경로다.
- High-risk impact: no (렌더 전용, 산술 없음). 다만 이 함수가 렌더하는 문자열 중 하나가 원장이 동결한 오늘 실현손익이고 그것은 일일 손실 한도 축의 분자다 — 오표시는 운영자가 남은 한도를 잘못 읽게 만든다.
