# Function Logic Map: `consoleBroker.resolve`

- Source: `cmd/tossctl/console.go`
- AST evidence: `ast.json` (revision=current, L395–407, 분기 2개)
- Risk scan: `risk-pattern-report.md`
- 이 change의 base: `583772c4` — base에는 이 함수가 없었다. seam마다 흩어져 있던 lazy 구축을 한 곳으로 모은 것이다 (revision=current)

콘솔의 **계좌 해석 1회**가 사는 곳이다. `verifyBrokerFactory`는 이 파일에서 여기와 `consoleVerifyStarter` 두 곳에서만 불린다.

- **넘어가는 것**: 호출자(같은 파일의 seam)에게 `verifylive.Broker`. internal/console에는 넘어가지 않는다 — seam이 여기서 method value만 꺼내 쓴다.
- **왜 락 안에서 구축하는가**: 두 화면이 같은 순간에 열리면 하나가 해석하고 다른 하나는 **기다린다**. 직렬화되는 것이 정확히 rate limit이 걸리는 호출이므로, 기다리는 편이 두 번 읽는 것보다 싸다.
- **실패를 기억하지 않는 이유**: 콘솔이 뜰 때 없던 자격증명이 `tossctl openapi login` 이후에는 있을 수 있다. 실패를 캐시하면 그 콘솔은 재시작 전까지 영구히 못 읽는다. 매 렌더 재시도를 막는 것은 internal/console의 TTL 캐시다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `c.root` | nil 허용 | `newConsoleBroker` | 자격증명 해석 실패는 에러로 반환 |
| `c.client` | nil 또는 해석 완료된 브로커 | `verifyBrokerFactory` 1회 | nil이면 구축 시도 |
| `c.mu` | 구축 전체를 덮는다 | 이 함수 | 동시 첫 호출 2건 → 해석 1회 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `c.client != nil` — 이미 해석됨 | 없음 | 캐시된 브로커 | `TestOpeningEveryConsoleReadScreenResolvesTheAccountOnce`, `TestTheOrdersSeamResolvesTheAccountOnceAndBuildsNoSecondClient` |
| B2 | `verifyBrokerFactory` 실패 | 없음 — 캐시하지 않는다 | `nil, err` | 동일(factory 실패 경로) + 대시보드 미측정 렌더 |
| (else) | 첫 성공 | `c.client` 설정 | 브로커 | 동일 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `c.mu.Lock` / `defer Unlock` | 구축을 한 번으로 만든다 | 구축(네트워크 포함)이 락 안에서 일어나므로 동시 첫 호출은 대기한다 | ast.json calls, L396–397 |
| `verifyBrokerFactory` | 라이브 클라이언트 | 내부에서 `/api/v1/accounts` 1회(429 지점, M4). 에러는 그대로 반환 | verify.go `buildVerifyBroker` |

## State mutations and fallbacks

- `c.client`만 변이한다. 성공만 저장하고 실패는 저장하지 않는다.
- 계좌 참조 문자열(두 번째 반환값)은 여기서 버린다. 읽기 화면은 계좌 번호를 표시하지 않으며, 기록에 계좌를 적는 것은 검증 실행의 일이다 — 그래서 검증은 자기 해석을 따로 한다(`consoleVerifyStarter`).

## Safety conclusion

- Safe edit boundary: 캐시 조건과 락 범위. 락 밖에서 구축하도록 바꾸면 동시 첫 호출이 다시 두 번 해석한다.
- High-risk impact: yes (주문 경로 — rate 예산) — 여기서 얻는 클라이언트가 주문 가능 클라이언트이고, 이 함수의 호출 결과가 콘솔의 모든 읽기 화면이 쓰는 값이다. 실패 캐싱을 도입하면 `openapi login` 이후에도 화면이 영구히 미측정으로 남는다.
