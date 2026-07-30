# Function Logic Map: `Console.readOnly`

- Source: `internal/console/console.go`
- Change: `console-operator-overview`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

이 함수는 이 change의 base commit 이후에 **신설**됐다(base에 같은 이름의 선언이 없다). gate가 evidence를 요구하는 것은 diff hunk가 현재 본문과 교차하기 때문이며, 아래 분석은 현재 HEAD 본문에 대한 것이다.

console-orders-screen가 신설한 래퍼. `mutating`의 거울이며 GET·HEAD 외의 메서드를 405로 거부한다. 존재 이유는 리뷰 P0-2다 — 라우트 레코드가 `{Path, Session, CSRFGated}`뿐이면 '이 경로는 GET이다'라는 문장이 '**CSRF 보호가 없다**'로 퇴화하고, `/orders`의 계좌 동사 예외가 **보호되지 않았다는 이유로** 부여된다. 그 상태에서 세션 쿠키만 가진 POST가 CSRF 없이 서빙된다. 래퍼는 그 사실을 등록에 실어 정적 검사가 읽을 수 있게 만들고, 405는 런타임 두 번째 자물쇠다.

이름은 design D3이 `reading`으로 못박았으나 이 패키지에 이미 `type reading struct`(overview.go의 (값, 측정됨, 사유) 삼중항)가 있어 Manager 판정으로 `readOnly`로 개명됐다(issues.md I-2). Go는 둘 다 컴파일하지만 한 단어가 한 패키지에서 두 뜻을 갖는다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `r.Method` | `GET` 또는 `HEAD` | HTTP 요청 | 그 외는 405 + `Allow: GET, HEAD` |
| `next` | 감쌀 핸들러 | `c.routes()`의 등록 | 해당 없음 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `r.Method != GET && r.Method != HEAD` | `Allow` 헤더 설정 | 405 + '읽기 전용 화면이다' 거부 페이지 | `TestAPostToTheOrdersScreenIsRefusedByTheReadOnlyWrapper` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `c.refuse` | 405 거부 페이지 | '주문을 내거나 정정·취소하는 수단은 이 콘솔에 없다'를 명시 | pages.go:376 |
| `next` | GET·HEAD를 통과 | `c.handleOrders` | orders_page.go |
| (금지 바인딩) | 계좌·원장 무접촉 | 메서드 판정만 한다 | ast.json calls |

## State mutations and fallbacks

- 무상태. 응답 헤더 하나만 쓴다.
- `/orders` 하나에만 적용된다. 모든 읽기 경로에 붙이면 load-bearing한 한 곳이 보이지 않게 되므로 의도적으로 좁다(`console.go` 주석, add-candidate-discovery issues 9-3이 `/signals`에서 같은 근거로 붙이지 않았다).
- 정적 인식은 `registeredRoutes`의 `case "readOnly"` 한 줄에 달려 있다 — 개명 후 변이(래퍼 제거)로 가드가 실제로 무는지 재확인했다(issues.md I-2).

## Safety conclusion

- Safe edit boundary: 신설. 라우트 등록 한 곳(`/orders`)에만 붙고 다른 라우트의 체인은 바꾸지 않는다.
- High-risk impact: yes (인증·라우트 게이트 경로 — 이 래퍼가 없으면 `/orders`의 계좌 동사 예외가 '보호되지 않았다'를 근거로 부여되고 POST가 세션 쿠키만으로 서빙된다)
