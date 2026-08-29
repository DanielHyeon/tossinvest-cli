# Function Logic Map: `consoleOrdersSeam`

- Source: `cmd/tossctl/console.go`
- AST evidence: `ast.json` (revision=current, L500–502, 분기 0개)
- Risk scan: `risk-pattern-report.md`
- 이 change의 base: `137cc8d0` — base에는 이 함수가 없었다. 현재 본문을 diff hunk가 덮으므로 evidence가 요구된다 (revision=current)

/orders 화면이 받는 능력의 **생성 지점**이다. 경계를 넘는 것은 `*lazyOrders` 하나이고 그 값이 선언하는 메서드는 `Orders` 하나뿐이다 (`console.OrdersReader`).

- **넘어가는 것**: 계좌의 주문 목록을 *읽는* 능력 하나. `console.OrdersReading` 값(원문 문자열 필드)만 돌아온다.
- **넘어가지 않는 것**: `verifylive.Broker` 자체, 그 뒤의 `*official.Client`, 그리고 그 클라이언트가 가진 `PlaceOrder`/`CancelOrder`/`ModifyOrder`/조건주문 변경. `lazyOrders`는 브로커도 method value도 필드로 보관하지 않는다 — 필드는 공유 resolver(`shared`) 하나다. 따라서 internal/console 쪽에서 타입 단언으로도 주문 능력에 도달할 수 없다.
- **계좌 해석**: 여기서는 아무것도 해석하지 않는다. 인자가 `*rootOptions`가 아니라 `*consoleBroker`라는 것이 이 seam의 계좌 해석이 **포지션 화면과 같은 것**이라는 사실이다. 해석은 첫 `Orders` 호출 때 `consoleBroker.resolve` 안에서 콘솔 프로세스당 한 번 일어난다. `tossctl console` 기동 시점에 해석하지 않는 이유는 예전과 같다 — 운영자가 한 번도 열지 않을 화면이 콘솔 기동의 전제조건이 되면 안 된다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `shared *consoleBroker` | non-nil — `runConsole`이 만든 단 하나 | `newConsoleBroker(root)` | 여기서는 검증하지 않는다. 자격증명 실패는 첫 `Orders` 호출에서 에러로 나오고, 배선이 공유본인지는 `TestOpeningEveryConsoleReadScreenResolvesTheAccountOnce`가 소스에서 읽는다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| (무분기) | — | 없음 — `&lazyOrders{shared: shared}` 할당만 | `console.OrdersReader` (nil 아님) | `TestTheConsoleIsHandedOneCapabilityAndNotABroker` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (없음) | — | — | ast.json calls=null |

## State mutations and fallbacks

- 상태 변이 없음. 생성된 `*lazyOrders`는 캐시를 갖지 않는다 — 클라이언트 캐시는 `consoleBroker` 한 곳에 있다.
- 여기서 nil을 돌려주는 경로가 없다는 것이 의도다 — `consoleGateLimitsSeam`/`consoleSettingsSeam`과 달리 해석해야 할 파일 경로가 없으므로 "seam 미배선"이라는 상태가 존재하지 않는다. 실패는 화면 위 문장으로 나온다.

## Safety conclusion

- Safe edit boundary: 생성자 한 줄. `lazyOrders`에 브로커 자체를 담게 바꾸는 순간 콘솔이 주문 능력을 받는다.
- High-risk impact: yes (주문 경로) — 주문 가능 클라이언트에서 읽기 능력만 잘라내는 **바로 그 이음매**다. `&lazyOrders{shared: shared}`를 `broker`를 품는 구조체로 바꾸는 한 번의 편집이면 콘솔이 주문 능력을 갖는다. 현재는 갖고 있지 않다. 인자를 `*rootOptions`로 되돌리는 편집은 주문 능력이 아니라 rate 예산의 문제다 — 세션당 `/api/v1/accounts` 1회가 2회가 된다.
