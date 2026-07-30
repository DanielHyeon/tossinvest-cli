# Function Logic Map: `Console.routes`

- Source: `internal/console/console.go`
- Change: `console-operator-overview`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

이 change의 diff가 이 함수의 본문을 바꿨다. 아래 분석은 현재 HEAD 본문에 대한 것이다.

라우트 표 그 자체. 세 change가 각각 한 줄씩 더했다 — `c.registerOverview(mux)`, `c.registerOrders(mux)`, `c.registerSignals(mux)`. 세 줄은 등록을 **다른 파일로** 옮겼고, 그것이 console-operator-overview 리뷰 P1-1이 지적한 결함의 발동 조건이다: 그때까지 `registeredRoutes`는 `console.go` 한 파일만 파싱했으므로 다른 파일에서 등록한 라우트는 라우트 표 검사 **넷 모두에서 침묵으로 통과**했을 것이다. 그래서 이 change는 등록을 옮기기 전에 추출기를 패키지 전체로 넓혔다(task 1.1~1.3).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `http.NewServeMux()` | Go 1.22 mux | 표준 라이브러리 | 해당 없음 |
| `c.session0` | 모든 라우트의 첫 래퍼 | console.go:545 | 빠지면 `TestEveryRouteGoesThroughTheSessionGate`가 실패 |
| `c.mutating` | 상태변경 9경로에만 | console.go:577 | 빠지면 `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate`가 실패 |
| `c.readOnly` | `/orders` 하나에만 | console.go:633 | 빠지면 `TestTheOrdersRouteIsRegisteredReadOnlyAndNothingElseIs`가 실패하고 `/orders`의 계좌 동사 예외가 성립하지 않는다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| (무분기) | — | 본문에 분기가 없다 | 단일 경로 | `TestEveryRouteGoesThroughTheSessionGate` + `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate` + `TestTheOrdersRouteIsRegisteredReadOnlyAndNothingElseIs` + `TestEveryRouteRefusesARequestWithoutTheSessionToken`(런타임) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `mux.HandleFunc` ×19(직접 16 + 하위 등록 3) | 라우트 등록 | 에러 없음 | ast.json calls |
| `c.registerOverview` / `registerOrders` / `registerSignals` | 화면별 라우트를 각자 파일에서 등록 | 각각 `/dashboard`·`/orders`·`/signals` 1건 | overview.go:1153, orders_page.go, signals.go |
| (금지 바인딩) | 주문·정정·취소·게이트·자격증명 라우트가 없다 | `routeTableFindings`가 표 전체를 판정한다 | static_test.go:754 |

## State mutations and fallbacks

- `http.Handler` 하나를 만들어 돌려준다. 계좌·원장·파일 부작용 없음.
- 상태변경 라우트는 정확히 9개이며(`consoleStateChanging`), 셋 다 계좌를 건드리지 않는다 — 검증 제어 3, 프로세스 제어 4, 편입 설정 편집 2.
- `/orders`는 계좌 동사(`order`)를 이름에 가진 유일한 경로이고, 예외는 `{바이트 일치 경로 + readOnly wrapper + CSRF 게이트 부재}` 셋을 모두 요구한다.

## Safety conclusion

- Safe edit boundary: 등록 3줄 추가. 기존 16개 등록의 게이트 체인은 byte 그대로다(base 대비 diff가 삽입만이다).
- High-risk impact: yes (인증 경로 — 세션 게이트와 CSRF 게이트가 라우트에 붙는 유일한 지점이며, 한 줄을 빠뜨리면 인증 없는 라우트가 생긴다)
