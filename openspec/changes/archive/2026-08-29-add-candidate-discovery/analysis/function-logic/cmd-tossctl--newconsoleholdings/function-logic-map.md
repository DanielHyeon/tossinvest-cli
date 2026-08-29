# Function Logic Map: `newConsoleHoldings`

- Source: `cmd/tossctl/console.go`
- AST evidence: `ast.json` (revision=current, L414–416, 분기 0개)
- Risk scan: `risk-pattern-report.md`
- 이 change의 base: `583772c4` — 본문 변경: 인자가 `*rootOptions`에서 콘솔이 공유하는 `*consoleBroker`로 바뀌었다 (revision=current)

포지션 화면이 받는 능력의 **생성 지점**. 경계를 넘는 것은 `*lazyHoldings` 하나이고 그 값이 선언하는 메서드는 `Holdings` 하나뿐이다 (`console.HoldingsReader`).

- **넘어가는 것**: 보유 종목을 *읽는* 능력 하나.
- **넘어가지 않는 것**: `verifylive.Broker`와 그 뒤의 주문 메서드. `lazyHoldings`는 `shared` 필드 하나만 갖는다.
- **인자가 말하는 것**: `*consoleBroker`를 받는다는 것이 곧 "이 화면의 계좌 해석은 /orders와 같은 것"이라는 사실이다. `*rootOptions`를 받던 시절에는 seam마다 자기 클라이언트를 만들었고, 두 화면을 여는 세션은 `/api/v1/accounts`를 2회 읽었다(M4의 429 지점).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `shared *consoleBroker` | non-nil — `runConsole`이 만든 단 하나 | `newConsoleBroker(root)` | 여기서는 검증하지 않는다. 자격증명 실패는 첫 렌더에서 에러로 나온다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| (무분기) | — | 없음 — `&lazyHoldings{shared: shared}` 할당만 | `console.HoldingsReader` (nil 아님) | `TestTheConsoleIsHandedOneCapabilityAndNotABroker` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (없음) | — | — | ast.json calls=null |

## State mutations and fallbacks

- 상태 변이 없음. 반환값은 캐시를 갖지 않는다 — 클라이언트 캐시는 `consoleBroker` 한 곳이다.
- nil을 돌려주는 경로가 없다. 실패는 "seam 미배선"이 아니라 화면 위 문장으로 나온다.

## Safety conclusion

- Safe edit boundary: 생성자 한 줄과 인자의 타입.
- High-risk impact: yes (주문 경로) — `&lazyHoldings{shared: shared}`를 브로커를 품는 구조체로 바꾸는 한 번의 편집이면 콘솔이 주문 능력을 받는다. 인자를 `*rootOptions`로 되돌리는 편집은 세션당 계좌 해석을 1회에서 2회로 되돌린다.
