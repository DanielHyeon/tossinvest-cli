# Function Logic Map: `Client.OrderByID`

- Source: `internal/official/orders_reads.go`
- AST evidence: `ast.json` (revision `base` — base 쪽 hunk에만 걸린다)
- Risk scan: `risk-pattern-report.md` (매칭 0건)

기존 함수 — **본문 무변경**이다. 이 change는 이 함수 **뒤**(base L123 이후)에 원문 보존 읽기
블록을 삽입했고, import 3줄 추가로 함수가 3칸 아래로 밀렸다. base(`137cc8d`) L115-121과
HEAD L118-124는 **바이트 동일**(함수 구간 sha256 `1a1ca1c9f72eb251…` 일치, 본 세션 확인).
`internal/official/orders_reads.go`의 diff는 `+191 -0`으로 **삭제가 0줄**이다.

## 계좌에 닿는 방식

- 엔드포인트: `GET /api/v1/orders/{orderId}` (`url.PathEscape`로 이스케이프).
- rate-limit 그룹: `budgetKey`가 id 세그먼트를 접어 `/api/v1/orders/{id}` 하나로 계량된다.
  이 change 이후에는 그 응답의 예산 헤더가 `doRequest`에서 기록되지만, 이 함수의 동작은
  그것을 읽지 않는다.

이 패키지의 모든 계좌 읽기가 공유하는 경로는 하나다:
`getAcct` → `ensureAccountSeq`(`c.mu` 아래 지연 1회 해석, 결과 캐시) →
`getWithHeaders` → `send`. `send`가 401을 만나면 `c.tm.refresh` 후 **정확히 한 번** 재요청하고,
2xx가 아니면 `classifyStatus`가 401/403→`ErrAuth`(본문에 `\bip\b`가 있으면 `ErrIPNotAllowed`),
429→`ErrRateLimited`, ≥500→`ErrServer`, 그 밖의 4xx→`*APIError`(passthrough, fallback 없음)로
사상한다. `ShouldFallback`은 sentinel 넷에만 true다.

읽기 전용이다 — 계좌의 상태를 바꾸지 않는다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `orderID` | 임의 문자열, 경로에 escape되어 들어감 | 호출자 | 없는 id는 브로커의 4xx → `*APIError` |
| 계좌 헤더 | `X-Tossinvest-Account` | `ensureAccountSeq`(지연 1회) | 해석 실패 시 `lazy account-seq resolution: …` |

불변식: 반환 타입은 `domain.Order`이며 `adaptOrder`의 `parseDecimal`이 빈 문자열과 파싱 실패를
**둘 다 0**으로 만든다. 이 change는 그 성질을 고치지 않았다 — 고칠 수 없어서가 아니라 CLI와
MCP 도구의 직렬화 계약이고 기존 호출자 전부가 그것에 의존하기 때문이다. 그래서 부재와 0을
구분해야 하는 화면은 옆에 새로 생긴 원문 보존 읽기를 쓴다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if, base L117) | `getAcct` 실패(전송·401 재시도 후 인증·429·5xx·4xx) | 없음 | `domain.Order{}, err` — 분류된 sentinel 또는 `*APIError` | `TestOrderByIDIntegration`(정상 경로) + `client_test.go`의 분류 테스트 |

무분기 꼬리: 성공 시 `adaptOrder(raw)`.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `url.PathEscape` | 주문 id를 경로에 안전하게 | 순수 | ast.json calls |
| `c.getAcct` | 계좌 헤더 + 공유 send 경로 | 401 재시도 1회, `classifyStatus` | ast.json calls |
| `adaptOrder` | 응답 → `domain.Order` | 오류 없음(무조건 사상) | ast.json calls |

## State mutations and fallbacks

- 계좌 변경 없음(GET). 클라이언트 상태 변경은 `ensureAccountSeq`의 첫 해석 캐시뿐이며 이는
  base와 동일하다.
- fallback은 `ShouldFallback`이 sentinel에 대해 판단하는 상위 정책이고 무변경이다.

## Safety conclusion

- Safe edit boundary: **본문 0줄 변경**. 같은 파일 뒤쪽 가산과 import 삽입만.
- High-risk impact: **yes** — 계좌 게이트웨이(주문 조회) 표면이다. 본문은 무변경이지만
  같은 파일·같은 타입(`OrdersFilter`)·같은 헬퍼를 공유하는 코드가 추가됐으므로 회귀 표면이
  0은 아니다. additive인 근거: 이 파일의 diff에 삭제가 0줄이고, `Orders`·`OrderByID`·
  `adaptOrder`·`adaptOrders`의 함수 구간이 base와 바이트 동일하며, 새 심볼
  (`RawOrder`, `RawOrderList`, `ErrOrderStatusRequired`, `OrdersRaw`, `marketFromCurrency`)은
  이 change가 만든 것이라 선행 호출자가 없다.
