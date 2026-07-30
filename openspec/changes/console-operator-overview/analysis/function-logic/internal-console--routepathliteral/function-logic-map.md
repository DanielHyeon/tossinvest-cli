# Function Logic Map: `routePathLiteral`

- Source: `internal/console/static_test.go`
- Change: `console-operator-overview`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

이 함수는 이 change의 base commit 이후에 **신설**됐다(base에 같은 이름의 선언이 없다). gate가 evidence를 요구하는 것은 diff hunk가 현재 본문과 교차하기 때문이며, 아래 분석은 현재 HEAD 본문에 대한 것이다.

Go 문자열 리터럴 안의 경로를 꺼낸다. 따옴표 두 형태 모두. 이전에는 `"`만 트림했고, raw string 리터럴로 등록한 경로가 **백틱을 그대로 표에 실었다**. 그것을 알아챈 가드는 `TestNoRouteIsRegisteredWithAMethodPattern`이었고 '메서드 패턴'이라고 보고했다 — 결함도 파일도 진짜가 아니었다. 잘못된 곳으로 보내는 실패 메시지는 침묵보다 나쁘다: 가서 아무것도 못 찾고, 그 다음부터 가드를 불신하게 된다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `value` | Go 문자열 리터럴 원문(따옴표 포함) | `ast.BasicLit.Value` | 이스케이프는 해석하지 않는다 — 경로에 이스케이프를 쓰면 표의 문자열이 실제 경로와 다르다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| (무분기) | — | 본문에 분기가 없다 | 단일 경로 | `TestTheRoutePathIsReadOutOfEitherQuotationForm` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.Trim`(큰따옴표와 백틱 둘 다 트림 집합에 있다) | 양쪽 따옴표 제거 — 두 리터럴 형태를 같은 경로 문자열로 만든다 | 순수 | ast.json calls |

## State mutations and fallbacks

- 없음(순수 함수).

## Safety conclusion

- Safe edit boundary: 신설. 이전에는 호출 지점에 인라인된 `strings.Trim(lit.Value, "\"")`였다.
- High-risk impact: yes (라우트 게이트 감시의 입력 — 경로 문자열이 틀리면 경로 키 대조를 하는 가드 전부가 조용히 어긋난다)
