# Function Logic Map: `TestTheRoutePathIsReadOutOfEitherQuotationForm`

- Source: `internal/console/static_test.go`
- Change: `console-operator-overview`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

이 함수는 이 change의 base commit 이후에 **신설**됐다(base에 같은 이름의 선언이 없다). gate가 evidence를 요구하는 것은 diff hunk가 현재 본문과 교차하기 때문이며, 아래 분석은 현재 HEAD 본문에 대한 것이다.

`routePathLiteral`의 표 테스트. 세 케이스: 보통 문자열, raw string, 메서드 패턴. 세 번째는 **트림 후에도 메서드 패턴으로 남아야 한다**는 확인 — 따옴표 제거가 `TestNoRouteIsRegisteredWithAMethodPattern`이 볼 것을 지우면 안 된다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 케이스 표 | 3건 — 보통 문자열, raw string, 메서드 패턴(`GET /dashboard`) | 테스트 본문 | 해당 없음 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `for _, tc := range …` | 없음 | 없음 | 3케이스 전부 |
| B2 | `got != tc.want` | `t.Errorf` | 없음 | 같은 위 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `routePathLiteral` | 피검 대상 | 순수 | static_test.go:219 |

## State mutations and fallbacks

- 없음(순수 표 테스트).

## Safety conclusion

- Safe edit boundary: 신설. 추출기 자신을 재는 유일한 자리다 — 디스크의 소스를 파싱하는 검사에는 가짜 리터럴을 등록할 수 없다.
- High-risk impact: yes (라우트 게이트 감시의 회귀 방지)
