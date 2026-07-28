# Function Logic Map: `Client.ConditionalOrders`

- Source: `internal/official/conditional_reads.go`
- AST evidence: `ast.json` (revision `base` — base 쪽 hunk에만 걸린다)
- Risk scan: `risk-pattern-report.md` (매칭 0건)

기존 함수 — **본문 무변경**이다. 이 change는 이 함수 **뒤**(base L77 이후)에 원문 보존 읽기
블록을 삽입했고, import 2줄(`fmt`, `strings`) 추가로 함수가 2칸 아래로 밀렸다.
base(`137cc8d`) L48-76과 HEAD L50-78은 **바이트 동일**(함수 구간 sha256 `92dc36cb2ec3dc64…`
일치, 본 세션 확인). 파일 diff는 `+128 -0`으로 **삭제 0줄**이다.

## 계좌에 닿는 방식

- 엔드포인트: `GET /api/v1/conditional-orders`.
- rate-limit 그룹: `budgetKey("/api/v1/conditional-orders")` = 경로 그대로(식별자 세그먼트 없음).
- 계좌 해석: `getAcct` → `ensureAccountSeq`(지연 1회, 캐시). 새 원문 읽기와 **공유**한다.

이 패키지의 모든 계좌 읽기가 공유하는 경로는 하나다:
`getAcct` → `ensureAccountSeq`(`c.mu` 아래 지연 1회 해석, 결과 캐시) →
`getWithHeaders` → `send`. `send`가 401을 만나면 `c.tm.refresh` 후 **정확히 한 번** 재요청하고,
2xx가 아니면 `classifyStatus`가 401/403→`ErrAuth`(본문에 `\bip\b`가 있으면 `ErrIPNotAllowed`),
429→`ErrRateLimited`, ≥500→`ErrServer`, 그 밖의 4xx→`*APIError`(passthrough, fallback 없음)로
사상한다. `ShouldFallback`은 sentinel 넷에만 true다.

읽기 전용이다 — 조건주문을 만들거나 취소하지 않는다(그쪽은 `conditional_writes.go`).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `status`, `symbol`, `cursor` | 빈 값이면 파라미터 생략 | 호출자 | 브로커가 `status` 부재를 거부할 수 있다 |
| `limit` | `>0`일 때만 전송(기본 20, 최대 100) | 호출자 | — |
| 계좌 헤더 | `X-Tossinvest-Account` | `ensureAccountSeq` | 해석 실패는 그대로 오류 |

불변식: `adaptConditionalOrder`/`adaptCondition`의 `parseDecimal`이 빈 문자열을 0으로 만든다.
MARKET형 조건주문은 `orderPrice`가 null이고 STOP 다리는 `targetProfitRate`가 null이므로
이 읽기는 둘 다 "0"으로 렌더한다 — 그것이 옆에 원문 보존 읽기가 생긴 이유이며, 이 함수의
동작은 기존 호출자를 위해 **그대로 둔다**.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if, base L50) | `status != ""` | 쿼리에 `status` | — | `TestConditionalOrdersIntegration`(서버가 `OPEN` 단언) |
| B2 (if, base L53) | `symbol != ""` | 쿼리에 `symbol` | — | 동상(생략 시 부재) |
| B3 (if, base L56) | `cursor != ""` | 쿼리에 `cursor` | — | 동상 |
| B4 (if, base L59) | `limit > 0` | 쿼리에 `limit` | — | 동상 |
| B5 (if, base L63) | `getAcct` 실패 | 없음 | `domain.ConditionalOrderList{}, err` | `client_test.go`의 분류 테스트 |
| B6 (range, base L67) | 응답의 조건주문 순회 | 로컬 slice append | — | `TestConditionalOrdersIntegration` |

무분기 꼬리: `FetchedAt: time.Now().UTC()`를 붙여 반환.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `url.Values.Set` / `strconv.Itoa` | 쿼리 조립 | 순수 | ast.json calls |
| `c.getAcct` | 계좌 헤더 + 공유 send | 401 재시도 1회, `classifyStatus` | ast.json calls |
| `adaptConditionalOrder` | 응답 → domain | 오류 없음 | ast.json calls |
| `time.Now().UTC()` | `FetchedAt` | — | ast.json calls |

## State mutations and fallbacks

- 계좌 변경 없음(GET). 클라이언트 상태 변경은 `ensureAccountSeq` 캐시뿐이며 base와 동일.
- fallback 정책 무변경.

## Safety conclusion

- Safe edit boundary: **본문 0줄 변경**. 같은 파일 뒤쪽 가산과 import 삽입만.
- High-risk impact: **yes** — 계좌 게이트웨이의 조건주문 조회다. 이 제품에서 조건주문은
  드문 예외가 아니라 **지속되는 산물**이고(M18이 등록 프로세스보다 오래 사는 것을 측정했다),
  verifylive의 잔여물 판정은 조건주문과 일반 주문을 같은 것으로 센다.
  본문 무변경이지만 같은 파일에 새 코드가 붙었으므로 회귀 표면은 0이 아니다.
  additive인 근거: 파일 diff 삭제 0줄, 함수 구간 바이트 동일, 새 심볼
  (`RawConditionalOrder`, `RawConditionalOrderList`, `ConditionalOrdersRaw`)은 이 change가
  만든 것이라 선행 호출자가 없다.
