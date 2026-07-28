# Function Logic Map: `registeredRoutes`

- Source: `internal/console/static_test.go`
- Change: `console-operator-overview`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

이 change의 diff가 이 함수의 본문을 바꿨다. 아래 분석은 현재 HEAD 본문에 대한 것이다.

라우트 표 추출기. console-operator-overview가 **한 파일에서 패키지 전체로** 넓혔고(task 1.1~1.3), console-orders-screen가 `readOnly` 인식을 더했다.

**이 가드가 잡는 것**: 등록 호출이 소스에 문자 그대로 적혀 있는 모든 라우트에 대해 {경로 리터럴, `session0` 적용 여부, `mutating` 적용 여부, `readOnly` 적용 여부}. 그리고 표를 빠져나가는 네 가지 모양을 **조용히 건너뛰지 않고 소리 내어 실패한다** — 등록자를 값으로 받는 형태(`register := mux.HandleFunc`), 서브트리 마운트(`mux.Handle("/x/", …)`), 리터럴이 아닌 경로, 들여다볼 수 없는 핸들러 인자.

**이 가드가 잡지 못하는 것(측정된 경계)**: ① `registrarNames`는 `{HandleFunc, Handle}` 둘뿐이다 — 다른 라우터 타입의 `.GET(...)` 같은 등록은 보이지 않는다. ② `_test.go` 파일의 등록은 대상이 아니다(`packageFiles`가 제외한다) — 의도된 범위다. ③ 핸들러가 자기 안에서 `r.URL.Path`로 다시 분기하는 형태는 등록 1건으로만 보인다. ④ 다른 패키지에서 등록해 mount하는 표는 원리적으로 보이지 않는다 — 서브트리 규칙이 그 경로를 막는 것이 이 가드가 그 구멍에 대해 하는 전부다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 패키지 비테스트 소스 | `packageFiles(t)` | 디스크 | 0개면 `t.Fatal` |
| `registrarNames` | `{HandleFunc, Handle}` | static_test.go:89 | 그 밖의 등록자는 보이지 않는다 — 위 경계 ① |
| `called` 집합 | call의 Fun으로 등장한 SelectorExpr | 같은 Inspect 순회 | `ast.Inspect`가 pre-order이므로 CallExpr가 자기 Fun보다 먼저 기록된다는 사실에 의존한다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `for name := range files` | 파일 이름 수집 | 없음 | 패키지 전 파일 파싱 — `TestEveryRouteGoesThroughTheSessionGate`의 하한 19가 카나리아 |
| B2 | `for _, name := range names`(정렬) | 파일별 파싱 | 없음 | 같은 위. 정렬은 실패 메시지의 재현성 |
| B3 | registrar가 call의 Fun이 아닌 SelectorExpr로 등장 | `t.Errorf` — 등록자를 값으로 받았다 | 없음(계속) | 리뷰가 변이로 시연한 형태(`register := mux.HandleFunc` 뒤 `register("POST /verify/order/cancel", h)`) — 그 상태에서 가드 다섯이 전부 통과했고 라우트는 **인증 없는 POST에 200**을 답했다 |
| B4 | `call, ok := n.(*ast.CallExpr); !ok` | 없음 | 순회 계속 | 구조 분기 |
| B5 | `sel, ok := call.Fun.(*ast.SelectorExpr); !ok || !registrarNames[...]` | 없음 | 순회 계속 | 구조 분기 — 경계 ①이 여기서 생긴다 |
| B6 | `len(call.Args) == 0` | `t.Errorf` — 경로 없는 등록 | 계속 | 인자 수로 등록을 건너뛰던 이전 동작의 대체(task 1.2) |
| B7 | 첫 인자가 STRING 리터럴이 아님 | `t.Errorf` — 리터럴 아닌 경로 | 계속 | 변수 경로는 추출 불가이므로 거부한다 |
| B8 | `strings.HasSuffix(r.Path, "/") && r.Path != "/"` | `t.Errorf` — 서브트리 패턴 | 계속 | Go 1.22+에서 후행 슬래시는 서브트리이므로 그 아래 전부가 이 스캔 밖이다 |
| B9 | `for _, arg := range call.Args[1:]` | 핸들러 인자 순회 | 없음 | `Handle`/`HandleFunc` 양쪽 형태 |
| B10 | `opaqueHandler(arg)` | `t.Errorf` — 들여다볼 수 없는 핸들러 | 계속 | `TestTheRouteTableJudgementStillCatchesAnActingRouteThatIsNotOnTheAllowlist`의 자매 경계 |
| B11 | 인자가 CallExpr | 없음 | 계속 | 게이트 체인 인식의 전제 |
| B12 | 가장 바깥 호출이 `session0` | `r.Session = true` | 없음 | `TestEveryRouteGoesThroughTheSessionGate` |
| B13 | 내부 CallExpr | 없음 | 계속 | 중첩 래퍼 인식 |
| B14 | 내부 CallExpr의 Fun이 SelectorExpr | 없음 | 계속 | 같은 위 |
| B15 | `switch fn.Sel.Name` | 없음 | 아래 둘 | 래퍼 이름 판정 |
| B16 | `case "mutating"` | `r.CSRFGated = true` | 없음 | `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate` |
| B17 | `case "readOnly"` | `r.ReadOnly = true` | 없음 | `TestTheOrdersRouteIsRegisteredReadOnlyAndNothingElseIs` — 개명(issues.md I-2) 뒤 래퍼를 떼는 변이로 실제로 무는지 재확인했다 |
| B18 | `len(routes) == 0` | `t.Fatal` | 종료 | positive control — 추출기가 파싱을 멈춘 상태를 잡는다 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `packageFiles(t)` | 패키지의 **비테스트** `.go` 파일 전부를 읽는다 | 0개면 `t.Fatal` — 가드가 패키지를 보고 있지 않다는 카나리아 | static_test.go:36 |
| `parseFile` | `go/parser`로 파싱 | 실패하면 `t.Fatalf` | static_test.go:60 |
| (런타임 무접촉) | 가드는 **선언**을 읽는다. 실행되지 않는 코드가 정확히 이 가드가 잡아야 하는 것이다 | reflect·실행 경로를 쓰지 않는다 | static_test.go 파일 주석 |
| `routePathLiteral` | 따옴표 두 형태에서 경로를 꺼낸다 | 순수 | static_test.go:219 |
| `opaqueHandler` | 핸들러 인자를 들여다볼 수 있는지 | 순수 | static_test.go:245 |

## State mutations and fallbacks

- `[]route`를 만들어 돌려준다. 파일 쓰기·네트워크 없음.
- 표에는 `Console.routes`에서 도달하지 않는 등록도 들어온다 — 과대 근사이며 안전한 방향이다.
- `r.ReadOnly`는 이 change가 더한 네 번째 사실이다. 그 전에는 '이 경로는 읽기다'가 'CSRF 게이트가 없다'로만 표현될 수 있었고, 그러면 예외가 **보호되지 않았다는 이유로** 부여된다.

## Safety conclusion

- Safe edit boundary: 파일 범위 확장 + `Handle` 인식 + 인자 수 무시 제거 + 서브트리·값-등록자·불투명 핸들러 거부 + `readOnly` 인식. 게이트 인식 로직 자체(`session0`/`mutating`)는 형태가 같다.
- High-risk impact: yes (인증·라우트 게이트 감시 — 이 추출기가 조용히 라우트를 놓치면 세션 게이트·CSRF 게이트·메서드 패턴·계좌 동사 검사 **넷이 동시에** 공허하게 통과한다)
