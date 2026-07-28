# Function Logic Map: `routeReadsTheAccountRecord`

- Source: `internal/console/static_test.go`
- Change: `console-orders-screen`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

이 함수는 이 change의 base commit 이후에 **신설**됐다(base에 같은 이름의 선언이 없다). gate가 evidence를 요구하는 것은 diff hunk가 현재 본문과 교차하기 때문이며, 아래 분석은 현재 HEAD 본문에 대한 것이다.

`/orders`에 부여된 계좌 동사 예외가 한 라우트에 적용되는지를 답하는 **순수 함수**.

**요구하는 사실 셋, 전부 load-bearing**: ① 경로가 `consoleAccountReads`와 **바이트 일치**, ② `readOnly` 래퍼가 실제로 체인에 있음, ③ CSRF 게이트 밖(이 파일의 정의로 CSRF 게이트 안은 이름이 무엇이든 상태변경이다).

**부드러운 비교를 쓰지 않는 이유는 각각 이름이 있다**: 접두 일치면 `/orders/cancel`이 통과하고, `ToLower`면 `/Orders`가 통과하며(Go mux는 대소문자를 구분하므로 그것은 두 번째 라우트다), `TrimSuffix("/")`면 `/orders/`가 통과하는데 그것은 중복보다 나쁘다 — Go 1.22+에서 후행 슬래시는 **서브트리 패턴**이라 `/orders/cancel`이 그 핸들러로 라우팅된다.

**잡지 못하는 것**: 이 함수는 등록의 모양을 잰다. 핸들러가 실제로 읽기만 하는지는 재지 않는다 — 그쪽은 런타임 405(`TestAPostToTheOrdersScreenIsRefusedByTheReadOnlyWrapper`)와 `OrdersReader`의 메서드 집합이 맡는다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `r.Path` | `consoleAccountReads` = `{"/orders": true}`와 바이트 일치 | static_test.go:669 | 불일치면 예외 없음 |
| `r.ReadOnly` | true | 추출기 B17 | false면 예외 없음 |
| `r.CSRFGated` | false | 추출기 B16 | true면 예외 없음 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| (무분기) | — | 본문에 분기가 없다 | 단일 경로 | `TestTheOrdersExceptionAppliesOnlyToTheExactPathThatReadsAndDoesNotAct`(3 negative), `TestTheOrdersExceptionDoesNotReachAnyPathBeneathOrBesideIt`(5 negative) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (없음) | 맵 조회와 불 대수뿐 | 순수 | ast.json calls=null |

## State mutations and fallbacks

- 없음(순수 함수).

## Safety conclusion

- Safe edit boundary: 신설. 테스트 밖으로 뽑아낸 것이 요점이다 — `registeredRoutes`는 디스크의 소스를 파싱하므로 가짜 `/orders/cancel`을 등록해 가드가 거부하는지 볼 수 없다. 뽑아내지 않으면 유일하게 가능한 단언이 '그 경로가 allowlist에 없다'인데, 그것은 접두 일치·ToLower·후행 슬래시 트림 **셋 다에서 참으로 남으므로 아무것도 재지 않는다**.
- High-risk impact: yes (계좌 동사 금지의 유일한 예외 판정 — 이 함수가 넓어지면 `/orders/cancel`이 정적 검사를 통과한다)
