# Design: console-orders-screen

## 문맥

`console-operator-overview`의 제약을 그대로 상속한다 — JS 없음, CDN 없음, 빌드 없음,
`html/template` 서버 렌더 + meta refresh. 이 change가 더하는 것은 화면 하나와, **그 화면을
정직하게 만들기 위해 필요한 두 개의 가산 읽기 표면**이다.

StockOS의 `OrdersPage`에서 가져오는 정보 구조: 필터(시장·방향·상태) + `N/M건`, 그리고
`broker_order_id` 열(없으면 `—`), 상태 열, 제출 시각. 가져오지 않는 것: SPA·SSE·
`POST /api/orders/submit` 계열 전부(overview D9와 같은 이유).

## D1. `domain.Order`로는 이 화면을 만들 수 없다

이 change의 핵심 요구는 "브로커가 주지 않은 값은 0이 아니라 —"다. 그런데:

```go
// internal/official/market_reads.go:15
func parseDecimal(s string) float64 {
	if s == "" { return 0 }
	f, err := strconv.ParseFloat(s, 64)
	if err != nil { return 0 }
	return f
}
```

`Quantity`·`FilledQuantity`·`Price`·`AverageExecutionPrice`는 전부 이 함수를 지난 `float64`다.
**부재·해석불가·진짜 0이 콘솔이 보기 전에 하나가 된다.** API 스키마 자체가 `price`를 시장가에서
nullable로, `execution` 전체를 미체결 동안 null로 정의하므로, **모든 미체결 주문이 평균체결가
0으로 도착해 숫자로 렌더된다.**

같은 문제를 이 저장소가 이미 한 번 풀었다.

> `RawHolding` — "a value that has been through float64 and back has lost the evidence of what
> the broker actually said" ([asset_reads.go:75](../../../internal/official/asset_reads.go#L75))

**결정**: 같은 패턴을 주문에 적용한다. 브로커의 decimal 문자열을 보존하는 읽기를
`internal/official`에 **가산**하고, `Orders`/`adaptOrder`는 손대지 않는다. 기존 호출자의
동작·서명·에러 매핑이 하나도 바뀌지 않는 것이 이것을 계좌 게이트웨이에 넣어도 되는 이유다
— `RawHolding`이 같은 근거로 이미 거기 있다.

**`internal/brokerstate`를 쓰지 않는 이유는 절반뿐이다.** 그 패키지는 이미 주문에 대해
`State`/`FailClosed`/`Reason`/`Detail`을 모델링하고 있고 의존성이 `internal/domain` 하나뿐이라
콘솔이 import할 수 있다. 상태 판정은 거기서 가져올 수 있다. 그러나 그쪽 `parseDecimal(*string)`도
`nil`과 `""`를 둘 다 `0`으로 돌려주므로 **소수의 부재는 거기서도 무너진다.** 상태는
`brokerstate`, 소수의 부재는 원문 보존 읽기 — 둘 다 필요하다.

## D2. 조건주문을 세지 않는 미체결 건수는 측정이 아니다

`Client.Orders`는 `/api/v1/orders`만 본다. 조건주문은 `ConditionalOrders`로 다른 엔드포인트에
있다([conditional_reads.go:48](../../../internal/official/conditional_reads.go#L48)).

그런데 `verifylive`의 정리 로직은 **둘 다 잔여물로 센다** —
[cleanup.go:141](../../../internal/verifylive/cleanup.go#L141)이 양쪽에 대해 "노출 상한을
채우고 있는 동안에는 아무것도 보낼 수 없다"고 말한다. 그리고 M18은 조건주문이 프로세스
종료를 넘어 존속한다는 **측정**이다 — 이 제품에서 조건주문은 예외가 아니라 지속되는 산출물이다.

**함정**: 일반 주문만 부르는 화면은 조건주문 잔여물이 살아서 다음 검증을 막고 있는데
**"미체결 0건"을 측정된 값으로 렌더한다.** 이 change가 막으려는 실패를 이 change가 저지른다.

**결정**: 화면당 브로커 호출 **2콜**(주문 + 조건주문), 한 TTL, 두 목록을 구분해 표시한다.
그리고 **부분 실패는 합산하지 않는다** — 조건주문 조회만 실패하면 총계는 "N건 + 조건주문
미측정"이지 N건이 아니다. 확신에 찬 0은 도달 불가능해야 한다.

## D3. 라우트 예외는 정확 경로 1건, 그리고 "읽기"를 실제로 검사한다

현행 `TestNoRouteNamesAnAccountMutation`은 경로에 `order`가 들어가면 실패시킨다. 느슨한
문자열 검사이고 느슨한 것이 요점이다 — `/orders/cancel`도 `/order-place`도 한 번에 잡는다.

`/orders`는 그 금지가 겨냥한 것이 아니다. 그러나 그 구분은 사람이 읽어야 보인다.

**결정 1 — 예외는 정확 경로 집합이다.**

```
consoleAccountReads = {"/orders"}   // 바이트 일치. 접두 아님, 대소문자 무시 아님, 후행 슬래시 아님
```

- 접두 일치를 쓰면 `/orders/cancel`이 통과한다.
- `strings.ToLower`를 쓰면 `/Orders`가 통과한다.
- `TrimSuffix(p, "/")`를 쓰면 `/orders/`가 통과하는데, Go 1.22+에서 후행 슬래시 패턴은
  **서브트리 패턴**이라 `/orders/cancel`이 그 핸들러로 라우팅된다.
- 예외는 **두 루프 모두**에서 참조해야 한다. `actVerbs` 루프는 `accountVerbs`를 포함하므로
  `/orders`가 양쪽에 걸린다. `consoleStateChanging`에 넣어 조용히 만드는 것은 **틀린 수리**다
  — 그러면 `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate`가 CSRF 게이트를 요구한다.

**결정 2 — "읽기"를 검사 가능한 사실로 만든다.**

초안 spec은 "예외 경로가 CSRF 게이트 밖 GET임을 검사가 확인한다"고 썼다. 검사는 HTTP 메서드를
볼 수 없다 — `route`가 나르는 사실은 `{Path, Session, CSRFGated}`뿐이고, 패키지 전체에 메서드
검사는 `mutating()` 안의 한 줄뿐이다([console.go:533](../../../internal/console/console.go#L533)).
그래서 "GET"은 **"CSRF 보호가 없다"**로 퇴화하고, 예외가 보호되지 않았다는 이유로 부여된다.
그 상태에서 `POST /orders`는 세션 쿠키만으로 CSRF 없이 통과한다.

**결정**: `mutating`의 거울인 `reading(next)` wrapper를 도입한다. GET/HEAD가 아니면 405이고,
같은 `ast.Inspect` 분기가 인식한다. 예외 조건은 그때 비로소 문자 그대로 검사 가능해진다 —
`route.Path == "/orders" && route.Reading && !route.CSRFGated`.

**메서드 패턴(`"GET /orders"`)은 쓰지 않는다.** 현행 추출기는 리터럴을 그대로 경로로 읽으므로
(`strings.Trim(lit.Value, "\"")`) 경로가 `GET /orders`가 되어 **모든 경로 대조가 어긋난다.**
wrapper 쪽이 추출기를 바꾸지 않고 같은 사실을 준다.

**결정 3 — 예외의 크기를 예외 자신이 잰다.** `/orders/cancel`·`/orders/new`·`/orders/amend`·
`/Orders`·`/orders/`가 등록되면 가드가 실패함을 **직접 확인**하는 테스트. 이를 위해 판정부를
순수 함수로 뽑아야 한다 — 현행 `registeredRoutes`는 디스크의 소스를 파싱하므로 테스트가
가짜 라우트를 "등록"할 수 없다. 뽑지 않으면 구현자는 "allowlist에 그 경로들이 없다"는 약한
테스트를 쓰게 되고, 그것은 아무것도 재지 않는다.

## D4. 발주 주체 — 원장에 접근자가 없다

`domain.Order.ID`가 원장 `mutation_attempts.broker_order_id`에 있으면 엔진이 낸 주문이다.
그런데 `journal.ReadOnly`의 메서드는 일곱뿐이고(`AccountRefs`·`LivePositionExits`·
`AccountExitEvents`·`AccountTradeTrips`·`Close`·`Path`·`SchemaVersion`) **그 테이블을 읽는 것이
없다.** `readOnlyTables`도 `positions`·`exit_states`·`exit_events`·`trade_outcomes` 넷뿐이다.

(초안 design은 이 테이블을 `attempts`라 부르고 `dispatch.go`를 인용했다. 둘 다 틀렸다 —
테이블은 `mutation_attempts`이고 DDL은 `internal/journal/schema.go`에 있다.)

**결정**: 읽기 전용 접근자를 **가산**하고 `readOnlyTables`에도 등록한다. 등록하지 않으면
`OpenReadOnly`가 성공한 뒤 질의가 **하나씩 실패하고**, 그 실패는 0행으로 돌아온다 — 그리고
0행은 "전부 수동 주문"으로 읽힌다. 그 목록이 존재하는 이유가 정확히 그것이다.

**3-상태**: `엔진 발주` / `그 밖` / **`원장 미판독 — 불명`**. 원장을 못 읽었을 때 "그 밖"으로
적으면 화면은 엔진이 아무 일도 안 한 것처럼 보인다. `/positions`의 `JournalReadable` 규율을
그대로 따르고, 미판독 사유는 페이지 수준 안내 1회다.

## D5. 페이지 경계를 숨기지 않는다

```go
// orders_reads.go:106
var raw apiOrderPage           // NextCursor, HasNext를 갖고 있다
...
return adaptOrders(raw.Orders), nil   // 둘 다 버린다
```

콘솔은 더 있는지조차 알 수 없다. 한 페이지가 잘리면 **"미체결 N건"이 조용히 짧아진다.**

**결정**: `HasNext`를 seam 밖으로 노출한다. 참이면 건수는 숫자가 아니라 **"N건 이상"**이다.
`RawOrderPage`는 이미 `NextCursor`·`HasNext`를 보존하고 있으므로
([orders_raw.go:33](../../../internal/official/orders_raw.go#L33)) 새로 만들 것이 없다.

## D6. 필터는 링크로만, 그리고 캐시를 쪼개지 않는다

시장·방향·상태 필터는 GET 쿼리 파라미터와 링크로 구현한다(JS 없음).

**결정**: 세 필터는 전부 **한 번 가져온 캐시 위에서 in-process로** 적용한다. 브로커 파라미터로
넘기면 `/orders?status=OPEN`과 `?status=CLOSED`가 별개 캐시 키가 되어 TTL당 2콜이 4콜이 된다.

필터가 걸린 화면은 **필터 후 건수와 전체 건수를 함께** 보인다(`N/M건`). 숨긴 행이 보이지
않으면 "주문이 이것뿐"으로 읽힌다. **목록이 미측정이면 필터는 작동하지 않는다** — `0/—건`은
"0건이 일치"로 읽히므로, 미측정일 때는 필터 UI를 비활성으로 렌더하고 건수를 숨긴다.

## D7. 종목명과 시장은 이 화면에 없다

`adaptOrder`가 채우지 않고, 원문에도 그 필드가 없다. `apiOrder.Currency`("KRW"/"USD")는
디코드되고 버려진다.

**결정**: **시장은 `currency`에서 유도해 원문 보존 읽기가 나른다**(D1의 가산 안에서 처리).
**종목명은 이 화면에 두지 않는다** — 어디에도 없는 값을 위해 열을 만들고 전 행에 `—`를
찍는 것은 정보가 아니라 소음이다. 심볼로 충분하고, 이름이 필요하면 `/positions`가 준다.

이것을 spec에 적어 둔다. 적지 않으면 다음 사람이 "이름 열이 왜 없지"를 다시 조사한다.

## D8. 이 화면에도 행위가 없다

`/orders`는 GET뿐이고 폼이 없다. 주문을 내거나 정정·취소하는 수단이 없다. 확인 문자열
타이핑·2단계 클릭 등 **어떤 확인 마찰도 넣지 않는다**(사용자 지시 2026-07-27).

취소가 필요하면 그것은 콘솔이 아니라 `tossctl`의 일이고, 그 경계는 `콘솔 안전 불변식`이
소유한다.

## 결정된 계약값

```yaml
route:
  orders: /orders             # GET, session0, reading wrapper, CSRF 밖
guards:
  account_read_exact_paths: ["/orders"]   # 바이트 일치. 접두·대소문자·후행슬래시 전부 아님
  consulted_in_both_loops: true           # accountVerbs 루프 + actVerbs 루프
  do_not_touch: consoleStateChanging      # 여기 넣으면 CSRF 게이트가 요구된다
  reading_wrapper: true                   # GET/HEAD 외 405, 같은 AST 검사가 인식
  method_patterns: false
reads:
  orders_raw: additive-in-internal-official     # RawHolding 선례. Orders/adaptOrder 무변경
  conditional_orders: true                      # 다른 엔드포인트. 잔여물은 양쪽 다다
  journal_origin: additive-in-journal-readonly  # mutation_attempts + readOnlyTables 등록
rate_budget:
  orders_calls_per_refresh: 2   # 주문 1 + 조건주문 1
  orders_cache_ttl_seconds: 15
  filters_applied: in-process   # 캐시를 쪼개지 않는다
columns: [submitted_at, symbol, market, side, status,
          quantity, filled_quantity, price, avg_fill_price, broker_order_id, origin]
columns_absent_by_construction: [name]        # 원문에 없다. 열을 만들지 않는다
origin_states: [engine_issued, other, journal_unreadable]
partial_failure: never-summed                 # 조건주문만 실패하면 "N건 + 미측정"
truncated_page: "N건 이상"                    # HasNext가 참이면 숫자가 아니다
forms_on_this_screen: 0
```
