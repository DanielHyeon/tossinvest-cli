# Function Logic Map: `TestTheConsoleIsHandedOneCapabilityAndNotABroker`

- Source: `cmd/tossctl/console_test.go`
- AST evidence: `ast.json` (revision=current, L249–309, 분기 12개)
- Risk scan: `risk-pattern-report.md`
- 이 change의 base: `47672c6f` — 본문 변경: 주문 seam의 형상 검사 + `console.Options`의 `Orders` 필수 + 이유 있는 예외 map (revision=current)

콘솔이 받는 능력의 **형상**을 두 방향에서 고정한다: 런타임 reflect(메서드 수)와 소스 파싱(`console.Options` 리터럴의 필드 이름).

이 branch range에서 추가된 것: ① 주문 seam에 대한 같은 주장(`Orders` 메서드 하나), ② `console.Options`가 `Orders` 필드를 반드시 받는다는 검사, ③ `consoleFieldExemptions` — 금지 단어(`order`)를 목록에서 지우는 대신 **이유를 적은 예외**를 두는 방식. 예외에 이유가 비어 있으면 그것도 실패다.

두 seam은 이제 `runConsole`처럼 **공유 resolver 하나**로 만들어진다. 계좌 해석을 공유해도 경계를 넘는 값의 형상은 변하지 않는다는 것이 이 테스트가 계속 하는 말이고, 공유 자체의 주장은 `TestOpeningEveryConsoleReadScreenResolvesTheAccountOnce`가 소유한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `newConsoleHoldings(shared)` | non-nil | console.go | 메서드 수 != 1이면 FAIL |
| `consoleOrdersSeam(shared)` | non-nil | console.go | 메서드 수 != 1이면 FAIL |
| `shared := newConsoleBroker(&rootOptions{})` | non-nil, 두 seam이 같은 값을 받는다 | console.go | 공유는 해석에만 적용된다 — 형상 주장은 이 값과 무관하게 그대로여야 한다 |
| `consoleOptionFields(t)` | 비어 있지 않아야 함 | console.go 소스 파싱 | 비면 가드가 아무것도 읽지 않은 것 — FAIL |
| `consoleFieldExemptions` | 이유 문자열이 비지 않은 map | 이 파일 | 빈 이유 = 금지가 조용히 짧아진 것 — FAIL |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | holdings seam의 메서드가 `[Holdings]` 하나가 아님 | — | `t.Fatalf` | 이 테스트 자신 (변이: method value 대신 broker를 필드로 보관) |
| B2 | 메서드 이름 수집 루프 | — | — | 동일 |
| B3 | orders seam의 메서드가 `[Orders]` 하나가 아님 | — | `t.Fatalf` | 동일 (변이: `lazyOrders`에 broker 필드 추가) |
| B4 | 메서드 이름 수집 루프 | — | — | 동일 |
| B5 | `console.Options` 리터럴을 못 찾음 | — | `t.Fatal` — 가드가 배선을 읽지 못함 | 동일 |
| B6 | `Holdings` 필드 부재 | — | `t.Error` | 동일 |
| B7 | `Orders` 필드 부재 | — | `t.Error` | 동일 (이 change가 추가) |
| B8 | 필드 순회 | — | — | 동일 |
| B9 | 예외 목록에 있는 필드 | `continue` | — | 동일 |
| B10 | 예외의 이유가 공백 | — | `t.Errorf` | 동일 (이유 없는 예외 금지) |
| B11 | 금지 단어 순회 (`broker/client/order/place/cancel`) | — | — | 동일 |
| B12 | 필드 이름에 금지 단어 포함 | — | `t.Errorf` | 동일 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `reflect.TypeOf` | 경계를 넘는 값의 메서드 집합을 런타임에서 센다 | 타입 단언으로도 넓어지지 않음을 보장 | L257, L270 |
| `consoleOptionFields` | `console.Options` 리터럴의 키를 소스에서 읽는다 | 런타임 테스트가 절대 실행하지 않을 실패(필드 추가)를 잡는다 | L283 |
| `strings.Contains` / `ToLower` | 금지 동사 검사 | 예외는 map으로만 | L301–304 |

## State mutations and fallbacks

- 테스트 — 계좌·원장·설정 어디에도 쓰지 않는다.

## Safety conclusion

- Safe edit boundary: 금지 단어 목록과 예외 map. 목록에서 `order`를 지우는 것이 정확히 D2가 경고한 실패다.
- High-risk impact: yes (주문 경로의 집행자) — 콘솔이 주문 가능 클라이언트를 받지 않는다는 주장의 유일한 자동 검사다. 이 테스트를 약화하면 `console.Options`에 브로커를 넣는 변경이 무음으로 통과한다. 테스트 자체는 실계좌 부작용이 없다.
