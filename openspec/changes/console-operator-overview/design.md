# Design: console-operator-overview

## 문맥

TossOS 콘솔은 JavaScript도 CDN도 빌드 단계도 없다. `html/template` 서버 렌더 + meta refresh다.
이 change는 그 제약을 유지한다 — 화면은 표와 문장이고, 그 이상이 필요하지 않다.

참조 대상인 StockOS는 React/Vite SPA다. 가져오는 것은 **정보 구조**(무엇을 한 화면에 모으는가)
이지 구현이 아니다.

### StockOS에서 가져오는 것과 가져오지 않는 것

| StockOS | TossOS | 왜 |
|---|---|---|
| SummaryBento(순손익·총이익·총손실·승률·미실현·실현·건수) | 계좌 패널 + 오늘 패널 | 가져온다 |
| KR/US 시장별 분리 표 | 오늘 패널·계좌 패널의 시장 분리 | 가져온다 — D6 |
| RecentFillsStrip(최근 체결) | 최근 exit 이벤트 패널 | 축소해서 가져온다 — 체결은 원장 읽기 표면이 없다(D7) |
| 리스크 한도 카드 | 안전 패널의 Guardian 한도 | 가져오되 **소진분은 한 축만** — D8 |
| SSE `/api/stream` + 30초 폴백 폴링 | meta refresh | 가져오지 않는다 — JS 없음 |
| **UnifiedOrderTicket**(`POST /api/orders/submit`) | **없음** | D9 |
| **AutoOrderLaunchPanel**(자동주문 시작) | **없음** | D9 |
| **PositionsPanel 새로고침**(`POST /api/positions/refresh-quotes`) | **없음** | D9 |

## D1. `/`는 그대로 두고 `/dashboard`를 새로 만든다

`/`는 검증 콘솔이다. 이름이 `handleDashboard`일 뿐 내용은 soak·attestation·바이너리 판본·
검증 run이고, 그것은 지금도 필요하다. `/`를 옮기거나 리다이렉트하면 검증 중 열려 있던 탭들이
전부 다른 화면으로 갈아타고, 그 탭들은 승인 창을 보고 있는 탭이다.

**결정**: `/`는 손대지 않는다. 두 화면은 다른 질문에 답한다 — `/`는 "이 빌드를 믿어도 되나",
`/dashboard`는 "계좌가 지금 어떤가".

내비게이션 라벨과 `Nav` 키가 충돌한다. 현행 `head`는 `/`를 `대시보드`·`Nav == "dashboard"`로
표시한다([templates.go:66](../../../internal/console/templates.go#L66)). **결정**: `/`는
`검증 콘솔`(`Nav: "verify-console"`), `/dashboard`는 `개요`(`Nav: "overview"`).

## D2. 라우트 정적 검사가 **파일 하나만** 읽고 있다

이 change는 새 파일 `overview.go`를 만든다. 그 파일에서 라우트를 등록하면 어떻게 되는지 먼저
확인해야 한다.

```go
// static_test.go:78
src := packageFiles(t)["console.go"]
```

`registeredRoutes`는 `console.go`만 파싱한다. **다른 파일에서 등록한 라우트는 라우트 표
검사의 어떤 항목도 통과하지 않는다 — 검사가 그것을 보지 못하기 때문이다.**
`TestNoRouteNamesAnAccountMutation`도, `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate`도,
`TestTheDashboardScreensAreReads`도 전부 그 라우트에 대해 침묵한다.

이것은 D3이 인터페이스 검사에서 지적하는 것과 **같은 결함이 라우트 쪽에 있는 것**이다.
`add-candidate-discovery` 5.5가 `/signals`를 추가할 예정이므로 곧 두 번째 사례가 생긴다.

**결정**: `registeredRoutes`를 패키지 전체 파일로 넓히고, `Handle`도 `HandleFunc`과 같이
인식한다(현행은 `sel.Sel.Name != "HandleFunc"`로 `Handle`을 건너뛰고, 인자가 2개가 아닌 호출도
건너뛴다). 넓힌 뒤 **변이로 검증한다** — `overview.go`에 `/verify/steal`을 등록해 FAIL 확인,
제거 후 PASS 확인.

라우트 수 하한(`>= 13`)도 함께 올린다. 현재 표에는 16개가 있고 이 change가 하나 더 만든다.
그 하한은 "가드가 파싱을 멈췄다"의 카나리아이므로 실제 수를 따라가야 한다.

**메서드 패턴은 쓰지 않는다.** Go 1.22+의 `HandleFunc("GET /dashboard", …)`는 합법이지만,
현행 추출기는 리터럴을 그대로 경로로 읽으므로(`strings.Trim(lit.Value, "\"")`) 경로가
`GET /dashboard`가 되어 **모든 경로 대조 가드가 조용히 어긋난다.** 메서드 제약이 필요해지는
것은 `/orders`를 추가하는 change이며, 거기서 추출기와 함께 다룬다.

## D3. 정적 검사의 단위는 "인터페이스"가 아니라 "`Options`가 받는 능력"이다

`read-only 불변식`은 "콘솔이 주입받는 브로커 인터페이스는 조회 메서드만 선언한다"이고,
이것을 지키던 것이 `TestTheConsoleBrokerInterfaceDeclaresNothingButReads`다(§1에서 대체됨). 그 검사에는 구멍이
셋 있고, 셋 다 이 change가 새 파일과 새 seam을 추가하는 순간 열린다.

**구멍 1 — 파일 하드코딩.** `packageFiles(t)["holdings.go"]` 한 파일만 읽는다
([static_test.go:523](../../../internal/console/static_test.go#L523)). 다른 파일에 광폭
인터페이스를 선언하면 아무것도 실패하지 않는다.

**구멍 2 — func 타입은 인터페이스가 아니다.** 주입 seam 일곱 중 **다섯이 func 타입**이다.

```
console.go:111    type StartVerify func(...)
engineproc.go:56  type StartEngine func() (string, error)
engineproc.go:62  type StopEngine  func() (string, error)
restart.go:90     type Relaunch    func(port int) error
restart.go:97     type RestartSoak func() (string, error)
```

`*ast.InterfaceType`을 찾는 검사는 이 다섯을 전부 못 본다. `type PlaceOrderFunc func(ctx
context.Context, o domain.Order) error`를 새 파일에 두고 `cmd/tossctl`이 채우면 — cmd/tossctl은
이미 `internal/official`을 import한다 — 인터페이스도 없고 금지 import도 없어서 통과한다.

**구멍 3 — 하나의 allowlist를 모든 인터페이스에 적용할 수 없다.** 금지 동사 목록에 `"order"`와
`"conditional"`이 있고([static_test.go:530](../../../internal/console/static_test.go#L530))
allowlist는 `{"Holdings": true}` 하나다. 그러므로 검사를 패키지 전체로 넓히면 **기존의 정당한
seam에서 실패한다** — `Handoff{Mint, Consume}`, `AdoptionSettings{Load, Save}`. 후자는 spec이
명시적으로 요구하는 것이다. 즉 "모든 인터페이스에 하나의 allowlist"는 그 자체로 모순이다.

셋을 각각 땜질하면 하드코딩이 파일 이름에서 타입 이름으로 옮겨갈 뿐이다 —
`type accountSeam interface { Holdings(...); PlaceOrder(...) }`가 통과한다.

**결정**: 검사의 단위를 바꾼다. **`Options` 구조체의 필드를 걷고, 각 필드의 선언 타입을 해결해,
그 seam의 자기 allowlist를 적용한다.**

```
Options 필드 → 선언 타입 → (인터페이스면 메서드 집합 / func 타입이면 시그니처)
             → 그 seam에 등록된 allowlist와 대조
```

- 파일 이름에도 타입 이름에도 고정되지 않는다.
- func 타입 seam이 처음으로 검사 범위에 들어온다.
- allowlist에 없는 **새 필드가 `Options`에 생기면 그 자체로 실패한다** — 능력이 열거되지
  않은 채 주입되는 것을 막는 것이 실제 불변식이고, 이것이 그 문장이다.
- 임베디드 인터페이스 거부는 현행 동작이므로 유지한다. 구조체 필드·파라미터의 인라인
  `interface{...}` 리터럴도 걷는다(현재 0건이므로 지금 고정하는 비용이 가장 싸다).
- `packageFiles`는 `_test.go`를 제외하므로 `fakeBroker.PlaceOrder` 때문에 실패하지 않는다.

**변이 검증 필수**: 새 파일에 mutation 메서드를 가진 seam을 `Options`에 추가 → FAIL,
제거 → PASS. func 타입으로도 같은 것을 확인.

### 이 결정은 두 군데가 부족했다 (2026-07-28 구현 리뷰, 변이로 확인)

첫 구현은 변이 셋(인터페이스 seam·func 타입·미열거 필드)을 전부 잡았다. 그런데 리뷰가
**한 홉만 더 가면 통과하는 모양 넷**을 찾았고, 원인은 구현이 아니라 이 결정문에 있다.

**부족 1 — "능력"을 메서드 집합이 아니라 타입 이름으로 읽었다.** 위 도식은 "선언 타입 →
메서드 집합"이라고 썼지만, 실제로 필요한 것은 **도달 가능한 모든 인터페이스의 메서드 집합**이다.
`HoldingsReader.Holdings`가 `PlaceOrder`·`CancelOrder`를 가진 핸들을 **반환**하면 allowlist를
손대지 않고도 통과한다. 그 타입 이름만 `OrderHandle`로 바꾸면 실패한다 — **검사가 철자를 보고
있다는 증거**다. 같은 구멍으로 `any` + 인라인 타입 단언, 제네릭(`Seam[OrderPlacer]`), 타입
별칭이 지나간다.

**부족 2 — 단위로 고른 `Options`가 좁다.** 패키지 전역 `var`와 `*Console`의 메서드로도 능력이
붙는다. 리뷰가 `SetDesk(d Desk)` + 패키지 전역 변수로 **패키지 전 테스트를 통과**시켰다.
"콘솔이 받는 능력이 전부 열거되어 있다"는 불변식이 **한 구조체를 통해 받은 능력에 대해서만**
강제되고 있었다.

**정정된 결정**: 검사 단위는 `Options` 필드 + **`*Console`의 내보낸 메서드** + **패키지 전역
인터페이스 변수**이고, 각 진입점에서 **타입을 고정점까지 해소해** 도달 가능한 모든 인터페이스의
메서드 집합에 allowlist를 적용한다. 이름 대조는 그 위의 보조 장치이지 그 자체가 검사가 아니다.

라우트 쪽에도 같은 종류가 하나 있었다 — 추출기가 `call.Fun`이 `SelectorExpr`일 때만 보므로
`register := mux.HandleFunc`으로 등록한 라우트가 **다섯 가드 전부에게 보이지 않는다.** 리뷰가
그 경로로 인증 없는 `POST /verify/order/cancel`을 200으로 서빙시켰다. **등록자를 값으로 가져가는
것 자체를 실패로 만든다** — 값을 따라가도록 가르치지 않는다.

## D4. 브로커 0콜과 `peek` — 그리고 그 대가

개요는 계좌 요약을 보이므로 holdings 캐시를 원한다. 자기 요청으로 갱신하게 두면 **가장 오래
열려 있는 화면이 브로커 호출을 만든다.**

| 화면 | 갱신 1회의 브로커 호출 | TTL | 재로드 |
|---|---|---|---|
| `/positions` | holdings 1콜 | 15초 이상(현행 `holdingsTTL`) | TTL 이상 |
| `/dashboard` | **0콜** | — | `holdingsTTL` 이상 |

**함정**: 현행 `holdingsCache.get()`은 `!hold && TTL 경과`면 **무조건 갱신한다**
([holdings.go:165](../../../internal/console/holdings.go#L165)). 갱신하지 않고 읽는 방법이
없다. 여기서 자연스러운 우회는 `hold=true`로 부르는 것이고, 그러면 호출 수는 맞지만
`HeldReason`이 **"검증 중 — 갱신 보류"**로 렌더된다. 검증이 돌고 있지 않은데도. **지어낸
사유**이고, D5가 막으려는 사유 혼동이 정확히 그 자리에서 일어난다.

**결정**: `holdingsCache.peek(now)` — 갱신하지 않는 읽기를 신설한다. 캐시가 빈 경우의 사유는
`never_fetched`로 따로 둔다.

**대가를 적어 둔다**: 운영자가 `/positions`를 한 번도 열지 않으면 계좌 패널은 계속
"아직 읽지 않음"이다. 30초마다 같은 없음을 다시 그린다. 그래서 spec이 **값을 채우는 화면으로
가는 링크**를 요구한다 — D4의 산문에만 있으면 링크 없는 구현도 계약을 만족한다.

`holdingsSnapshot.Stale()`의 문서 주석 "It is only ever true alongside Held: an unheld request
refreshes instead"는 `peek` 도입으로 **거짓이 된다**. 같은 change에서 함께 고친다.

`c.positions(ctx)`도 그대로는 못 쓴다 — 내부에서 `holdings.get`을 부른다
([portfolio.go:398](../../../internal/console/portfolio.go#L398)). 관리/관리 외 집계에 필요한
원장 조인 부분을 fetch에서 분리해야 한다.

## D5. "못 읽었다"는 자료형으로 구분한다. 불리언 하나로는 안 된다

add-candidate-discovery D10과 같은 규칙이다. 값을 얻지 못한 것과 값이 0인 것은 화면에서
똑같이 생겼고 뜻이 다르다.

**결정**: 각 숫자는 `(값, 측정됨, 사유)`로 전달한다. 렌더 시점에 `0`과 `—`가 갈린다.

사유는 **일곱**이고 각각 다른 대응을 요구한다. 초안은 셋이라고 썼는데, 코드는 이미 넷을
모델링하고 있었다([holdings.go:83](../../../internal/console/holdings.go#L83)의 `Wired`).

| 사유 | 언제 | 운영자가 할 일 |
|---|---|---|
| `verify_suspended` | 검증 run이 살아 있다 | 기다린다 |
| `broker_read_failed` | 429·네트워크·인증 만료 | 고친다 |
| `journal_unreadable` | 원장 파일 부재·스키마 불일치 | 엔진 기동 또는 콘솔 갱신 |
| `seam_unwired` | 이 빌드에 그 seam이 배선되지 않았다 | 배선한다 |
| `never_fetched` | 캐시가 한 번도 채워진 적 없다 | 해당 화면을 연다 |
| `config_unreadable` | seam은 배선됐는데 config를 파싱하지 못했다 | 설정 파일을 고친다 |
| `journal_value_unparsable` | 원장은 열렸는데 값이 숫자가 아니다 | 원장을 조사한다 — 엔진 기동이 아니다 |

사유 없는 "—"는 대응할 수 없는 표시다. 일곱을 하나로 뭉치면 운영자는 기다릴지 고칠지
배선할지 알 수 없다.

**여섯 번째는 구현 중에 나왔다(I-2, Manager 판정).** 초안은 다섯이었고, 구현자는 다섯을
유지한 채 이 경우만 코드 없는 자유 문장으로 렌더했다. 판정은 **코드를 추가하는 쪽**이다 —
이 열거의 일이 "운영자가 무엇을 고쳐야 하는가"를 남김없이 적는 것이기 때문이다. 자유 문장으로만
존재하는 사유는 셀 수도, **없음을 테스트할 수도** 없고, 그러면 다음 사람이 일곱 번째도 문장으로
쓴다. 그 시점에 열거는 표면을 더 이상 기술하지 않는다.

**일곱 번째는 그 판정이 옳았다는 증거다(I-4).** 구현자가 같은 근거를 제가 열거하지 않은
경우에 적용했다 — 원장은 **열렸는데** 값이 숫자가 아닌 경우다. `journal_unreadable`은 여기서
틀린 조언이다(엔진을 기동하라는 뜻이 되는데 원장은 이미 열렸다). 판정문이 예고한 "일곱 번째"가
판정 다음 날 실제로 나왔고, 문장이 아니라 코드로 나왔다.

## D6. "오늘"의 경계는 시장마다 다르다

`trade_outcomes.closed_at`은 RFC3339 **UTC**로 저장된다
([decision.go:648](../../../internal/journal/decision.go#L648)). 콘솔의 시계도 UTC다.

UTC 자정을 "오늘"로 쓰면 경계가 **KST 09:00 — KR 장 시작 한 시간 뒤**에 떨어진다. 그날 아침
체결이 어제로 간다. ET 기준으로는 저녁이라 US 세션을 반으로 자른다.

저장소는 두 시장의 날짜를 이미 정의한다 — `MarketKR → Asia/Seoul`, `MarketUS →
America/New_York`([clock/market.go:44](../../../internal/clock/market.go#L44)).

**결정**: "오늘"은 **시장별 현지 자정**이다. 그리고 화면이 **어느 경계를 썼는지 출력한다** —
경계를 고르는 것과 고른 것을 감추는 것은 다른 문제이고, 감추면 두 사람이 같은 화면을 보고
다른 날을 읽는다.

같은 이유로 **계좌 패널도 시장별로 나눈다.** `domain.Position`에는 통화 필드가 없고
`MarketType` "KR"/"US"만 있다. 시장별 값은 각자의 통화이고, 그것을 더한 한 숫자는 아무 뜻도
없다. **합계를 만들지 않는 것이 합계를 잘못 만드는 것보다 낫다.**

## D7. 대시보드가 모으는 것은 "지금 결정에 필요한 것"뿐이다

| 패널 | 답하는 질문 | 출처 | 상태 |
|---|---|---|---|
| 엔진 | 돌고 있나, 안 돌면 왜 | `engineView` | 있음 |
| 계좌 | 평가액·평가손익·보유 수(관리/관리 외), **시장별** | holdings 캐시 `peek` | 있음 |
| 오늘 | 실현손익·왕복 건수·승패, **시장별** | `AccountTradeTrips`(Market·ClosedAt·비용차감 실현손익) | 있음 |
| 미체결 | 살아 있는 주문 몇 건 | — | **미측정(`seam_unwired`)** |
| 안전 | 검증 상태·잔여물·Guardian 한도 | `verifyView.Outstanding` + 한도 seam | 부분 |
| 최근 | 최근 exit 이벤트 | `AccountExitEvents` | 있음 |

**미체결 패널이 이 change에서 미측정인 것은 결함이 아니라 규칙의 적용이다.** 주문 조회 seam은
`console-orders-screen`이 배선한다. 그때까지 이 칸은 0이 아니라 `seam_unwired`다 — 이 change가
세우는 규칙을 이 change 자신에게 먼저 적용하는 자리이고, 임시 0을 넣었다가 나중에 고치는
쪽이 정확히 D5가 금지하는 것이다.

**최근 활동에서 체결은 뺀다.** StockOS는 최근 체결 8건을 보이지만, `journal.ReadOnly`에는
fills 접근자가 없다(`AccountRefs`·`LivePositionExits`·`AccountExitEvents`·`AccountTradeTrips`가
전부다). 접근자를 새로 만드는 것은 원장 읽기 표면 변경이고 `console-orders-screen`의 일이다.

**엔진 기동 거부 사유는 프로세스 로컬이다.** `c.engineNote`는 **이 콘솔 프로세스에서** 시작/정지를
누른 경우에만 채워진다([engineproc.go:196](../../../internal/console/engineproc.go#L196)).
갓 연 `/dashboard`에는 사유가 없다. 지속 저장된 사유를 찾아 헤매지 않도록 여기 적어 둔다.

**안전 패널은 두 시장을 다 읽는다.** `c.snapshot()`은 KR만 읽고
([data.go:192](../../../internal/console/data.go#L192)) KR/US 증거 기록은 별개 파일이다.
한 시장만 보이는 잔여물 패널은 다른 시장의 잔여물을 감추는데, 그것이 이 패널이 존재하는
이유 그 자체다.

## D8. Guardian 한도 — 한 축만 산출하고 나머지는 구조적으로 미측정이다

한도는 config에 있다 — `max_order_quantity`·`max_order_notional`·`max_total_exposure`·
`max_daily_loss_amount`·`max_daily_loss_ratio`
([config/engine.go:157](../../../internal/config/engine.go#L157)).

**문제 1**: 콘솔의 유일한 config seam은 편입 블록만 돌려주고, `TestTheSettingsSeamDeclaresExactly
LoadAndSave`가 세 번째 메서드에서 실패한다. spec은 그 규칙을 명시적으로 요구한다. 즉 초안은
같은 파일 안에서 "seam은 Load·Save 둘뿐"과 "대시보드가 Guardian 한도를 보인다"를 동시에
요구하고 있었다.

**결정**: 세 번째 메서드가 아니라 **별개 seam**이다. 하나짜리 읽기 seam을 따로 주입하고,
spec 문장을 **두 seam**으로 고쳐 쓴다. 좁은 인터페이스 규율은 seam을 늘려서 지키는 것이지
한 seam을 넓혀서 지켜지는 것이 아니다.

**정정(구현 중 발견)**: 이 결정의 초안은 그 seam의 반환형을 `config.AutomationGate`로 못박았다.
그러면 **기존 가드가 실패한다** — `TestTheConsoleDecidesNothingAboutTheGate`가 이 패키지의
모든 파일에서 `AutomationGate`·`Interlock`·`ProtectionReady`·`automation_gate` 식별자를 금지한다.
D3이 지적한 것과 같은 부류의 실수를 D8이 저지른 것이다: 새 표면을 설계하면서 그 표면에 걸리는
가드를 확인하지 않았다.

그리고 그 가드가 옳다. 콘솔은 엔진 프로세스에게 "돌아도 되냐"를 묻고 답을 그대로 표시한다.
게이트의 **자기 타입을 쥔 콘솔은 스스로 "게이트가 괜찮아 보인다"고 판단하는 코드에서 한 편집
거리**에 있다.

**정정된 결정**: seam은 **콘솔 로컬 값 타입**을 돌려준다 — 다섯 개의 상한과 통화. config 타입을
명명하지 않는다. 설계가 제안한 것보다 콘솔이 **더 적게** 쥔다. 0은 "미설정"이고 그것은 파일이
말하는 바이지 콘솔이 내리는 결론이 아니다 — 엔진의 기동 인터록은 미설정 상한을 허가가 아니라
거부로 다루며, 화면은 그것을 `미설정`으로 렌더한다.

**문제 2 — "오늘 소진분"은 어디에도 없다.** `OpenExposure`와 `DailyRealizedLoss`는 엔진이
in-process로 계산해 Guardian에 넘기는 값이고([risk/input.go:159](../../../internal/risk/input.go#L159)),
원장의 어떤 테이블에도 소진분이 없다.

**결정**: 소진분은 **한 축만** 산출한다 — 오늘 실현손익(원장 동결 값) 대 `max_daily_loss_amount`.

**통화 충돌(I-3, 구현 중 발견).** `max_daily_loss_amount`는 `limit_currency` 한 통화의 한
숫자이고 오늘 실현손익은 시장별 숫자다. "시장을 가로지르는 합계 금지"(D6)와 "실현손익 대 일일
손실 한도 산출"을 함께 지키려면 **통화가 일치하는 시장에서만** 비율이 성립한다. 한도 통화와
시장 통화가 다르면 두 숫자를 나란히 보이되 **비율은 내지 않고 그 이유를 행에 적는다.**
환산해서 비율을 만드는 것은 D6이 금지하는 그 합계를 한 칸 옆에서 다시 만드는 것이다.
나머지 축(개방 노출·주문 수량·주문 금액)은 **구조적으로 미측정**이라고 화면이 말한다. 이것을
적어 두지 않으면 구현자가 holdings에서 무언가를 계산하고, 그 숫자는 측정된 한도 소진분처럼
렌더된다 — 있지도 않은 보호를 믿게 만드는 화면이 한도가 없는 화면보다 나쁘다.

## D9. StockOS 대시보드의 절반은 **가져오면 안 되는 것**이다

StockOS의 `DashboardPage`는 읽기 화면이 아니다. 상시 마운트된 것들:

- `UnifiedOrderTicket` → `POST /api/orders/dry-run` · `/submit` — **주문 발주**
- `AutoOrderLaunchPanel` → `autoOrderExecutionStart` — **자동주문 실행 시작**
- `PositionsPanel` 새로고침 → `POST /api/positions/refresh-quotes`

TossOS 콘솔에서 이것들은 가져올 수 없는 것이 아니라 **존재해서는 안 되는 것**이다. 두 제품의
신뢰 모델이 다르다 — StockOS 대시보드는 조작 콘솔이고, TossOS 콘솔은 **관측창**이며 조작은
엔진 프로세스가 한다.

**결정**: 이 화면에는 form도 POST도 없다. StockOS에서 가져오는 것은 **어떤 숫자를 한 화면에
모으는가**뿐이고 **그 숫자로 무엇을 누르게 하는가**는 가져오지 않는다.

**초안의 오류**: 초안 D9는 "경로 allowlist가 이것을 기계적으로 보증한다"고 썼다. 보증하지
않는다 — 주문 티켓은 금지 동사가 하나도 없는 경로(`/ticket`)로 POST할 수 있고, 그 경로는 CSRF
게이트에 걸릴 때만 실패하는데 그것은 원하는 것의 반대다. 기계적 보증은 **D3의 `Options` 능력
열거**가 한다: 티켓이 주문을 내려면 주문 능력이 `Options`에 주입되어야 하고, 열거되지 않은
필드는 그 자체로 실패한다.

## D10. StockOS의 3-상태 렌더링은 이미 우리가 필요로 하던 증거다

D5를 쓴 뒤에 StockOS를 읽어 보니 같은 규칙이 이미 그쪽에 있다. 우연이 아니라 같은 실패를
같이 겪은 결과다.

| StockOS | 구분하는 것 |
|---|---|
| `useFreshness`: `lastUpdateMs === null → '—' idle` | **한 번도 받은 적 없음** vs 오래된 것 |
| `ExitPolicyPill`: `null → '—'` / 모르는 값 → `UNKNOWN` + 원문 / 아는 값 → 라벨 | 미배정 / 측정했으나 이 빌드가 모름 / 측정·인지 |
| `signalStatus === 'watchlist_empty'` vs "표시할 신호가 없습니다" | **볼 대상이 없어서 0건** vs 봤는데 0건 |
| `formatPercent`: `value == null → '미측정'` | 미측정 vs 0% |

TossOS에도 같은 규율이 이미 있다 — `positionRow.Unknown()`, `journalView.Readable()`,
`heldText`의 NULL 처리. 이 change는 새 규칙을 만드는 것이 아니라 **있는 규칙을 새 화면에
적용하는 것**이다.

**가져오지 않는 것 하나**: StockOS `signalExplanation.ts`의 `addDetail()`은 값이 없으면 그 줄을
**조용히 뺀다**. 라벨 없이 사라지는 것은 미측정 표시가 아니라 미측정의 은폐다. 같은 저장소
안에서도 exit strip은 `—`로 남기고 explanation은 지운다 — 남기는 쪽만 가져온다.

## D11. 이 화면에 행위가 없다 — 그래서 확인 마찰도 없다

`/dashboard`는 GET뿐이고 폼이 없다. 누를 것이 없으므로 확인할 것도 없다.

명시해 두는 이유: 이 저장소에는 "확인 문구를 타이핑하게 만들지 말라"는 사용자 지시가 있고
(2026-07-27), 리뷰 제안으로 되살아난 전례가 있다. 이 화면에 **어떤 형태의 확인 마찰도 넣지
않는다** — 타이핑 확인, 2단계 클릭, 추가 승인 전부. 혼자 쓰는 프로그램이다.

## 결정된 계약값

```yaml
route:
  overview:  /dashboard       # GET, session0, CSRF 밖
  nav_key:   overview
  nav_label: 개요             # `/`는 "검증 콘솔"로 라벨 변경
guards:
  route_scan:  package-wide   # HandleFunc + Handle, 파일 이름 하드코딩 금지
  seam_scan:   options-fields # 인터페이스 + func 타입, seam별 allowlist
  route_floor: 17             # 현행 16 + 이 change 1. 파싱 정지 카나리아
  method_patterns: false      # "GET /x" 리터럴은 경로 대조를 어긋나게 한다
rate_budget:
  overview_calls_per_refresh:  0    # peek only
  positions_calls_per_refresh: 1
refresh:
  overview_seconds: holdingsTTL     # 상수에서 파생
unmeasured_reasons:
  [verify_suspended, broker_read_failed, journal_unreadable, seam_unwired, never_fetched,
   config_unreadable, journal_value_unparsable]
today_boundary: per-market-local-midnight   # clock.Market.Location(); 화면이 어느 경계인지 출력
account_panel_split_by_market: true         # 통화를 섞은 합계 금지
today_panel_split_by_market:   true
guardian_axes_measured:   [daily_realized_loss_vs_max_daily_loss_amount]
guardian_axes_unmeasured: [open_exposure, order_quantity, order_notional]  # 기록 자체가 없다
open_orders_panel: unmeasured(seam_unwired) # console-orders-screen이 배선
safety_panel_reads_markets: [KR, US]
forms_on_this_screen: 0
```
