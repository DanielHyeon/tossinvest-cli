# Function Logic Map: `Console.openOrdersPanelFrom`

- Source: `internal/console/overview.go`
- Change: `console-orders-screen`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

이 함수는 이 change의 base commit 이후에 **신설**됐다(base에 같은 이름의 선언이 없다). gate가 evidence를 요구하는 것은 diff hunk가 현재 본문과 교차하기 때문이며, 아래 분석은 현재 HEAD 본문에 대한 것이다.

console-orders-screen가 신설한 개요의 미체결 패널 조립. `peek`이고 `get`이 아니다 — 이 화면의 계약은 렌더당 브로커 0콜이고 주문 화면의 갱신은 3콜(미체결 1 + 종결 1 + 조건주문 1)이 든다.

건수는 **주문 화면 자신의 합산 판독**(`combinedLive`)이며, 한쪽 목록이 미측정이면 합산하지 않는다. 여기서 덧셈을 재현하면 같은 콘솔에 '미체결 N건'의 정본이 둘이 되고, 조건주문 읽기가 처음 실패하는 순간 둘이 어긋난다. 리뷰 P0-1이 고정한 규칙이 그것이다 — **확신에 찬 0은 도달 불가능해야 한다**. 조건주문은 프로세스 종료를 넘어 존속하고 노출 상한을 채우는 잔여물이며, 한쪽만 세고 '0건'이라 적으면 다음 검증을 막고 있는 잔여물이 화면에서 사라진다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `c.ordersCache` | nil 아님(`New`가 항상 만든다) | `New` | seam이 nil이면 `peek`이 `Wired=false`를 답한다 |
| `snap.Lists.Open` / `.Conditional` | 두 목록 | `OrdersReader.Orders` 1콜 뒤의 세 브로커 호출 | 각자 자기 오류를 나른다 — 한쪽 침묵을 다른 쪽의 0으로 보고할 수 없다 |
| `snap.Lists.OpenError` / `.ConditionalError` | 목록별 실패 | 같은 곳 | 해당 목록만 미측정 |
| `now` | 주입 시계 | `c.now()` | 나이·TTL 판정 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `open.Known()` | `openLive = len(snap.Lists.Open)` | 없음(계속) | `TestAPlainReadThatFailedIsNeverAddedIntoTheTotalEither` |
| B2 | `conditional.Known()` | `conditionalLive = len(snap.Lists.Conditional)` | 없음(계속) | `TestAConditionalReadThatFailedIsNeverAddedIntoTheTotal` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `c.ordersCache.peek(now)` | 갱신 없는 읽기 | 브로커 0콜 — 이 화면의 계약 | orders.go의 `ordersCache` |
| `snap.TakenAt()` / `AgeSeconds()` / `Stale()` / `TTLSeconds()` | 값과 함께 나르는 출처 | 굵은 숫자만으로 말하지 않는다(D9) | orders.go / holdings.go |
| `c.verifyHold(now)` | `/orders`의 갱신이 보류 중인가 | **이 판독의 사유가 아니라** 별도 사실로 실린다 — `/orders`에 대한 사실이지 이 값에 대한 사실이 아니다 | holdings.go:267 |
| `listUnmeasured` / `combinedLive` | 목록별 미측정 판정과 합산 거부 | 한쪽이 미측정이면 합계도 미측정 | orders.go |
| (금지 바인딩) | `OrdersReader`는 메서드 1개(`Orders`)뿐이고, `order`·`conditional` 동사는 목록에서 **지우지 않고** 필드별·철자별 예외로 통과시킨다 | `TestTheOrdersSeamIsTheOnlyCapabilityWithVerbExemptionsAndTheyAreEnumerated` | orders_static_test.go:177 |

## State mutations and fallbacks

- `openOrdersPanel` 값 하나를 만든다. 캐시를 읽기만 한다.
- `Where = "/orders"` — 이 화면은 아무것도 갱신하지 않으므로 채우는 방법을 가리키지 않으면 운영자는 오지 않을 것을 기다린다(D4의 대가).
- `truncated`는 **측정된 목록의** 잘림만 센다 — 미측정 목록의 잘림 플래그가 합계에 '이상'을 붙이지 않는다.
- **문서 drift(발견 → 수정 완료)**: 이 함수의 doc 주석이 '주문 화면의 갱신은 **둘**이 든다'고 적고 있었다. design D2 개정(issues.md I-6)으로 일반 주문이 `OPEN`·`CLOSED` 두 호출로 갈라지면서 실제 비용은 **셋**이 됐고, spec delta(`주문 화면의 갱신 1회는 … 모두 3콜`)와 `console.go:199`·`console.go:275`가 모두 셋이라고 적는다. 동작은 처음부터 셋이었고 `TestOneRefreshAsksTheOpenGroupAndTheClosedGroupSeparatelyAndTheLiveOneWhole`이 그것을 고정한다 — 계약 위반이 아니라 주석 하나의 낡음이었다. 지금 주석은 '셋 — 미체결 그룹 1 + 종결 그룹 1 + 조건주문 1'로 고쳐졌고, 갈라 세는 이유(한 엔드포인트의 침묵을 다른 쪽의 0으로 보고하지 않는다)까지 함께 적는다. 본문은 바뀌지 않았으므로 아래 분기표와 AST 분기 집합은 그대로다.

## Safety conclusion

- Safe edit boundary: 신설. `overview`의 한 줄이 이 함수로 바뀌었고 다른 패널은 무변경.
- High-risk impact: yes (잔여물 건수 — 이 값이 다음 실계좌 검증의 노출 상한을 결정하는 사실이고, 미측정을 0으로 접으면 화면이 막힌 검증을 '깨끗하다'고 말한다)
