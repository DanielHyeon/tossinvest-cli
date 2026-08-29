# Function Logic Map: `lazyHoldings.Holdings`

- Source: `cmd/tossctl/console.go`
- AST evidence: `ast.json` (revision=current, L431–438, 분기 1개)
- Risk scan: `risk-pattern-report.md`
- 이 change의 base: `583772c4` — 본문 변경: 자기 lazy 캐시(`mu`/`l.read`)를 버리고 콘솔이 공유하는 클라이언트(`consoleBroker.resolve`)에서 method value를 매 호출 받는다 (revision=current)

대시보드/포지션 화면이 받는 **단 하나의 능력**.

- **넘어가는 것**: `holdingsFunc` 하나 — 즉 `broker.Holdings`의 **bound method value**. 반환은 `[]domain.Position`.
- **넘어가지 않는 것**: `verifylive.Broker`. `PlaceOrder`/`CancelOrder`/조건주문 변경을 가진 클라이언트는 이 타입의 어떤 필드에도 저장되지 않는다 — 필드는 `shared` 하나뿐이고, method value는 호출마다 지역 변수로 받았다가 버린다. internal/console 쪽에서 타입 단언으로도 도달할 수 없다.
- **계좌 해석**: 이 함수는 더 이상 자기 클라이언트를 만들지 않는다. `l.shared.resolve()`가 **콘솔 프로세스당 한 번** 해석하고 /orders 화면도 같은 것을 쓴다. 이전에는 seam마다 자기 클라이언트를 만들어서, 포지션 화면과 /orders를 모두 여는 세션이 `/api/v1/accounts`를 2회 읽었다 — 그 읽기가 2026-07-26에 429를 세 번 받아 실행 3스텝을 잃게 한 호출이다(measurements.md M4).
- **타임아웃**: 자기 데드라인 없음 — `ctx`는 요청의 것. 재시도 억제는 internal/console의 holdings 캐시(TTL)가 한다.
- **실패가 화면에 남기는 것**: 에러가 그대로 올라가 화면 위 문장이 된다. 구축 실패는 공유 resolver도 기억하지 않는다 — `tossctl openapi login` 이후 다음 TTL 만료에서 다시 시도된다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ctx`, `symbol` | 요청 context / 심볼 또는 빈 문자열 | internal/console 핸들러 | 취소는 브로커 콜 에러로 |
| `l.shared` | non-nil `*consoleBroker` | `runConsole`이 만든 공유 resolver 1개 | nil이면 첫 렌더에서 panic — 배선은 `runConsole` 한 곳뿐이고 `TestOpeningEveryConsoleReadScreenResolvesTheAccountOnce`가 그 배선을 읽는다 |
| 해석된 클라이언트 | `verifyBrokerFactory` 1회의 결과 | `consoleBroker.resolve` | 실패는 캐시되지 않고 그대로 반환 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `l.shared.resolve()` 실패(자격증명 없음 등) | 없음 — 이 타입은 아무것도 저장하지 않는다 | `nil, err` — 화면 위 문장 | 대시보드의 미측정 렌더 + `TestOpeningEveryConsoleReadScreenResolvesTheAccountOnce`의 factory 경로 |
| (else) | 해석 성공 | 없음 | `read(ctx, symbol)` — `broker.Holdings` 하나 | `TestTheConsoleIsHandedOneCapabilityAndNotABroker`, `TestOpeningEveryConsoleReadScreenResolvesTheAccountOnce` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `l.shared.resolve` | 콘솔이 공유하는 라이브 클라이언트 | 첫 호출만 `verifyBrokerFactory`(= `/api/v1/accounts` 1회), 이후는 캐시. 실패는 반환하고 기억하지 않는다 | `consoleBroker.resolve`의 map, verify.go `buildVerifyBroker` |
| `read` (= `broker.Holdings`) | 브로커의 Holdings만 | 요청 context를 그대로 전달 | ast.json calls, L437 |

## State mutations and fallbacks

- **상태 변이 없음**. 이 타입은 캐시를 갖지 않는다 — 캐시는 `consoleBroker` 한 곳으로 옮겼고, 렌더 억제는 internal/console의 TTL 캐시가 한다.
- 계좌 해석은 콘솔 프로세스당 1회다. 화면당도, 새로고침당도 아니다. 유일한 예외는 검증 실행(`consoleVerifyStarter`)이며 그 이유는 해당 함수의 주석과 map에 적혀 있다.

## Safety conclusion

- Safe edit boundary: `resolve()`에서 받은 값에서 무엇을 꺼내는가. `var read holdingsFunc = broker.Holdings` 대신 `broker`를 필드에 담는 **한 번의 편집**이면 콘솔이 주문 능력을 받는다.
- High-risk impact: yes (주문 경로) — 현재는 method value 하나만 넘어가고, 공유로 바뀐 뒤에도 넘어가는 값의 형상은 그대로다(`TestTheConsoleIsHandedOneCapabilityAndNotABroker`).
