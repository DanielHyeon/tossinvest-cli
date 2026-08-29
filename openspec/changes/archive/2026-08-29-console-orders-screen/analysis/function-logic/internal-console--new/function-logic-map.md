# Function Logic Map: `New`

- Source: `internal/console/console.go`
- Change: `console-orders-screen`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

이 change의 diff가 이 함수의 본문을 바꿨다. 아래 분석은 현재 HEAD 본문에 대한 것이다.

콘솔 생성자. 세션·CSRF 토큰을 프로세스당 1회 생성하고, 브로커/주문 캐시를 만들고, `routes()`가 조립한 게이트 체인을 `c.handler`에 고정한다. 이 change들이 바꾼 것은 `c.ordersCache = newOrdersCache(o.Orders, ordersTTL)` 한 줄이며, B1~B4 네 분기는 선행 change에서 온 그대로다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `o.StartVerify` | non-nil | cmd/tossctl 콘솔 배선 | nil이면 `ErrNoVerifyWiring`으로 생성 거부 — 검증 경로 없는 콘솔은 만들어지지 않는다 |
| `o.Now` | nil 허용 | 테스트는 fake clock, 운영은 nil | nil이면 `time.Now().UTC()` |
| `o.Out` | nil 허용 | cmd/tossctl가 stdout 전달 | nil이면 `io.Discard` — 배너를 못 쓰는 것이 panic보다 낫다 |
| `o.Binary` | nil 허용 | `binstamp.Self` | 실패해도 zero stamp를 유지하고 `Stamp.Same`이 '변경 없음'을 답한다 — 답할 수 없는 질문을 경고로 바꾸지 않는다 |
| `o.Holdings` / `o.Orders` | nil 허용 | cmd/tossctl seam | nil이면 캐시가 `Wired=false`를 답하고 화면이 `seam 미배선`으로 렌더한다 |
| session / csrf | `newToken(32)` / `newToken(16)` | crypto/rand | 외부에서 주입할 방법이 없다 — `TestNothingCanPresetTheSessionOrCSRFToken`이 소스로 고정 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `o.StartVerify == nil` | 없음 | `nil, ErrNoVerifyWiring` | `TestNewRefusesAConsoleWithNoWayToRunAVerification` |
| B2 | `c.now == nil` | `c.now = time.Now().UTC` | 계속 | `newHarness`가 Now를 비운 모든 케이스(예: `TestListenBindsTheLoopbackInterface`) |
| B3 | `c.out == nil` | `c.out = io.Discard` | 계속 | 이 패키지 테스트는 전부 `Out`을 설정한다 — 커버 없는 nil 방어 분기이며 이 change가 건드리지 않았다 |
| B4 | `c.opts.Binary == nil` | `c.opts.Binary = binstamp.Self` | 계속 | `TestTheStaleEngineBinaryIsWarnedAbout`(Binary 주입 쪽)과 나머지 전 케이스(기본값 쪽) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `newToken(32)` / `newToken(16)` | 세션·CSRF 토큰 발급 | `crypto/rand` 실패 시 panic — risk 스캔의 `go-panic` 발견 지점(console.go:665). 토큰 없이 계속 가는 것이 더 나쁘므로 의도된 정지다 | ast.json calls + risk-pattern-report.md |
| `newHoldingsCache` / `newOrdersCache` | TTL 캐시 조립 | 브로커 호출 없음 — lazy | ast.json calls |
| `c.opts.Binary()` | 설치 바이너리 지문 1회 측정 | 에러 무시하고 zero stamp 유지 | ast.json calls |
| `c.routes()` | 게이트 체인 조립 | 에러 없음 | ast.json calls |
| (금지 바인딩) | `internal/official`·`client`·`hybrid`·`execgw`·`flatten`·`app/engine`·`trading`·`orderintent` 중 무엇도 이 패키지 비테스트 소스가 import하지 않는다 | `TestTheConsoleHoldsNoBrokerOfItsOwn`이 소스 문자열로 고정 | `go list -f '{{.Imports}}' ./internal/console` |

## State mutations and fallbacks

- `*Console` 필드 초기화만 한다. 파일·네트워크·계좌 부작용 없음.
- `c.handler`가 여기서 고정되므로 이후 라우트 추가는 불가능하다 — 라우트 표는 생성 시점에 닫힌다.
- 실패 대체값 셋(now·out·Binary)은 모두 '더 조용한 쪽'이며 어느 것도 게이트를 완화하지 않는다.

## Safety conclusion

- Safe edit boundary: B1의 거부와 토큰 생성 두 줄. 캐시 조립 라인 추가는 그 앞뒤 어느 것도 바꾸지 않는다.
- High-risk impact: yes (인증 경로 — 세션·CSRF 토큰이 여기서 만들어지고, 이 두 토큰이 콘솔의 유일한 인증이다)
