# Function Logic Map: `TestTheOrdersSeamResolvesTheAccountOnceAndBuildsNoSecondClient`

- Source: `cmd/tossctl/console_test.go`
- AST evidence: `ast.json` (revision=current, L693–731, 분기 5개)
- Risk scan: `risk-pattern-report.md`
- 이 change의 base: `137cc8d0` — base에는 이 함수가 없었다. 현재 본문을 diff hunk가 덮으므로 evidence가 요구된다 (revision=current)

주문 seam이 **계좌 해석을 반복하지 않는다**는 것을 두 방향에서 고정한다: 런타임(3회 새로고침에 브로커 구축 1회)과 소스(`official.New(`가 console.go에 없고 `l.shared.resolve()`가 있다).

주장의 단위는 **seam 하나**다. 이 테스트는 포지션 화면이 따로 해석하던 동안에도 통과했고, 그것이 콘솔 단위의 주장을 `TestOpeningEveryConsoleReadScreenResolvesTheAccountOnce`로 따로 세운 이유다.

근거는 측정값이다 — 두 번째 `*official.Client`는 계좌 시퀀스를 다시 해석하고, 그 `/api/v1/accounts` 읽기가 2026-07-26에 429를 세 번 받아 실행 3스텝을 잃게 했다(measurements.md M4).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `verifyBrokerFactory` (패키지 변수) | 테스트가 교체 | verify.go | 복원은 `t.Cleanup` |
| `newVerifyServer(t)` | httptest 서버 | 테스트 헬퍼 | — |
| `readSource(t, "console.go")` | 패키지 디렉터리의 소스 | 디스크 | 읽기 실패는 `t.Fatalf` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 3회 새로고침 루프 | `built` 증가는 factory 안에서 | — | 이 테스트 |
| B2 | `seam.Orders(...)` 에러 | — | `t.Fatalf` | 동일 |
| B3 | `built != 1` | — | `t.Errorf` — 매 호출 구축은 매번 계좌 해석 | 동일 |
| B4 | console.go가 `official.New(`를 포함 | — | `t.Error` — 두 번째 클라이언트 금지 | 동일 |
| B5 | console.go가 `l.shared.resolve()`를 포함하지 않음 — seam이 공유 클라이언트를 통하지 않는다 | — | `t.Error` | 동일 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `consoleOrdersSeam(newConsoleBroker(&rootOptions{}))` | 테스트 대상 seam | 구축은 `consoleBroker.resolve`가 락 안에서 1회 | console.go L500 |
| `strings.Contains(src, "official.New(")` | 소스 절반의 가드 | 런타임으로는 잡히지 않는 두 번째 구축 경로를 잡는다 | L722 |

## State mutations and fallbacks

- 테스트 — 실계좌 접촉 없음(httptest).

## Safety conclusion

- Safe edit boundary: 구축 횟수 기대값과 소스 문자열 두 개. 이 테스트의 범위는 seam 하나이며, 세션 전체의 해석 횟수는 여기서 측정되지 않는다.
- High-risk impact: yes (주문 경로 — rate 예산) — 이 테스트가 없으면 콘솔이 계좌 해석을 새로고침마다 반복하는 변경이 통과하고, 그 429는 라이브 검증의 실주문 스텝을 잃게 한다. 테스트 자체는 실계좌 부작용이 없다.
