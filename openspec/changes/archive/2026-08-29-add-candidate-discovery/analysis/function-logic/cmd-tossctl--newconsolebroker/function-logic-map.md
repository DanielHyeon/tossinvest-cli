# Function Logic Map: `newConsoleBroker`

- Source: `cmd/tossctl/console.go`
- AST evidence: `ast.json` (revision=current, L380–382, 분기 0개)
- Risk scan: `risk-pattern-report.md`
- 이 change의 base: `583772c4` — base에는 이 함수가 없었다. 계좌 해석을 콘솔 프로세스당 1회로 묶기 위해 이 change에서 추가되었다 (revision=current)

콘솔의 **모든 읽기 seam이 공유하는 라이브 클라이언트의 보관함**을 만든다. 아무것도 구축하지 않는다 — `root`를 들고 있을 뿐이다.

- **넘어가는 것**: 없다. 반환값은 `*consoleBroker`이고 internal/console은 이 타입을 보지 못한다. 경계를 넘는 것은 이 값을 받은 seam이 만드는 method value뿐이다.
- **왜 lazy인가**: 구축이 곧 계좌 해석(`/api/v1/accounts`)이다. `tossctl console` 기동 시점에 해석하면, 운영자가 한 번도 열지 않을 화면이 콘솔이 뜨는 것의 전제조건이 된다. 첫 렌더가 값을 치르고, 실패는 화면 위 문장이 된다.
- **왜 하나인가**: seam마다 하나씩 만들면 포지션 화면과 /orders를 모두 여는 세션이 `/api/v1/accounts`를 2회 읽는다. 그 읽기는 2026-07-26에 429를 세 번 받아 실행 3스텝을 잃게 한 호출이다(measurements.md M4).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `root *rootOptions` | nil 허용(`&rootOptions{}` 포함) | root 명령의 persistent flag | 여기서는 검증하지 않는다 — 자격증명 실패는 첫 `resolve()`에서 에러로 나온다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| (무분기) | — | 없음 — `&consoleBroker{root: root}` 할당만 | `*consoleBroker` (nil 아님) | `TestOpeningEveryConsoleReadScreenResolvesTheAccountOnce` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (없음) | — | — | ast.json calls=null |

## State mutations and fallbacks

- 상태 변이 없음. `client` 필드는 nil로 시작하고 첫 `resolve()`에서만 채워진다.
- 이 함수의 호출 횟수 자체가 불변식이다: `runConsole` 안에서 정확히 1회. 두 번 부르면 해석도 두 번이 되며, 그 검사는 `TestOpeningEveryConsoleReadScreenResolvesTheAccountOnce`의 소스 절반이 한다.

## Safety conclusion

- Safe edit boundary: 생성자 한 줄과 호출 지점의 개수.
- High-risk impact: yes (주문 경로 — rate 예산) — 여기서 만든 클라이언트가 주문 가능 클라이언트다. 이 값을 seam에 그대로 넘기는 대신 `console.Options`에 넣는 편집이면 콘솔이 주문 능력을 갖는다(`TestTheConsoleIsHandedOneCapabilityAndNotABroker`가 막는다). 실행 자체는 아무 호출도 하지 않으므로 실계좌 부작용이 없다.
