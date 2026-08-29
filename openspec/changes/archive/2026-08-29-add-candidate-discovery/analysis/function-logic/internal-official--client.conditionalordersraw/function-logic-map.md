# Function Logic Map: `Client.ConditionalOrdersRaw`

- Source: `internal/official/conditional_reads.go`
- AST evidence: `ast.json` (revision `current`)
- Risk scan: `risk-pattern-report.md` (매칭 0건)

**신규 메서드**다(HEAD L150-204). 파라미터가 `Client.ConditionalOrders`와 정확히 같아서
호출 지점에서 서로 바꿔 끼울 수 있고, 그래서 호출자가 필요한 모양을 고르지 둘을 다 불러
요청을 두 번 쓰지 않는다.

## 계좌에 닿는 방식

- 엔드포인트: `GET /api/v1/conditional-orders` — `ConditionalOrders`가 부르는 그것.
- rate-limit 그룹: `/api/v1/conditional-orders`(식별자 세그먼트 없음) — 기존 읽기와 같은 그룹.
- 계좌 해석: `getAcct` → `ensureAccountSeq`. 기존 읽기와 **공유**하므로 두 번째
  `/api/v1/accounts` 요청이 생기지 않는다.

이 패키지의 모든 계좌 읽기가 공유하는 경로는 하나다:
`getAcct` → `ensureAccountSeq`(`c.mu` 아래 지연 1회 해석, 결과 캐시) →
`getWithHeaders` → `send`. `send`가 401을 만나면 `c.tm.refresh` 후 **정확히 한 번** 재요청하고,
2xx가 아니면 `classifyStatus`가 401/403→`ErrAuth`(본문에 `\bip\b`가 있으면 `ErrIPNotAllowed`),
429→`ErrRateLimited`, ≥500→`ErrServer`, 그 밖의 4xx→`*APIError`(passthrough, fallback 없음)로
사상한다. `ShouldFallback`은 sentinel 넷에만 true다.

## 빈 `status`는 요청을 보내기 **전에** 거부한다

`strings.TrimSpace(status) == ""`이면 요청을 만들지 않고 `ErrOrderStatusRequired`를 감싼
문장을 돌려준다. 문구는 일반 주문 쪽과 **다르게** 썼고, 그 차이가 요점이다:

- 일반 `/orders`: `status=OPEN`이 **전량 반환**이라 생략은 "거부 아니면 조용한 절단"이다.
- 조건주문: **두 그룹 모두 페이지네이션한다**(`limit` 기본 20, 최대 100). 그래서 여기서
  그룹을 대는 것은 전체 답을 사는 것이 아니라 **정의된 질문**을 사는 것이고, 호출자는
  여전히 `HasNext`를 읽어야 한다. 이쪽에 "전량 반환"이라고 적었으면 이 change가 막으려는
  바로 그 종류의 사실 아닌 문장이 됐을 것이다.
- 그룹의 의미: OPEN = `{WATCHING, PAUSED, ORDERING, ORDERED}` — 노출 상한을 채우는 바로 그
  집합. CLOSED = `{COMPLETED, EXPIRED}`.

`ShouldFallback`은 이 오류에 false다 — 보내지 않기로 한 요청은 웹 세션으로 갈 이유가 아니다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `status` | **필수**, 공백만도 불가. verbatim 전달 | 호출자 | `ErrOrderStatusRequired`(요청 미발생) |
| `symbol`, `cursor` | 빈 값이면 생략 | 호출자 | — |
| `limit` | `>0`일 때만 전송 | 호출자 | 브로커 기본 20 / 상한 100 |
| 계좌 헤더 | `X-Tossinvest-Account` | `ensureAccountSeq`(공유) | 해석 실패는 그대로 오류 |

불변식: 십진수는 문자열 그대로. **첫 다리(first)의 값만** 옮긴다 — 화면은 조건주문 하나당
한 행이고 OCO의 둘째 다리는 자기 모양이 필요하다. 두 다리를 합친 집계를 지금 지어내면
아무도 보내지 않은 숫자가 된다. 상태 문자열은 브로커의 어휘 그대로 보고한다(유도하면
잔여물이 다음 검증을 막는지 결정하는 상태에 두 번째 정본이 생긴다).

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if, L152) | `TrimSpace(status) == ""` | **없음 — 요청을 만들지 않는다** | `RawConditionalOrderList{}, %w(ErrOrderStatusRequired)` | `TestTheRawReadsRefuseARequestWithNoStatusGroup`(empty/blank 두 케이스) |
| B2 (if, L162) | `status != ""` | 쿼리에 `status` | — | `TestTheRawConditionalReadKeepsAnAbsentValueApartFromAZeroOne`(`OPEN` 전달) |
| B3 (if, L165) | `symbol != ""` | 쿼리에 `symbol` | — | 동상(생략 시 부재) |
| B4 (if, L168) | `cursor != ""` | 쿼리에 `cursor` | — | 동상 |
| B5 (if, L171) | `limit > 0` | 쿼리에 `limit` | — | 동상(`0` 전달) |
| B6 (if, L175) | `getAcct` 실패 | 없음 | `RawConditionalOrderList{}, err` | `orders_raw_test.go:TestRawReadsClassifyErrorsLikeEveryOtherRead`가 같은 `send` 경로를 잰다 |
| B7 (range, L183) | 페이지의 조건주문 순회 | 로컬 slice append | — | `TestTheRawConditionalReadKeepsAnAbsentValueApartFromAZeroOne` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.TrimSpace` | 공백만인 그룹도 부재 | 순수 | ast.json calls |
| `fmt.Errorf`(`%w`) | `ErrOrderStatusRequired` 감싸기 | `errors.Is` 가능 | ast.json calls |
| `url.Values.Set` / `strconv.Itoa` | 쿼리 조립 — 기존 읽기와 같은 조건 | 순수 | ast.json calls |
| `c.getAcct` | 계좌 헤더 + 공유 send | 401 재시도 1회, `classifyStatus` | ast.json calls |
| `make`/`append` | 결과 조립 | — | ast.json calls |

`market`은 유도하지 않는다 — 이 엔드포인트는 payload가 실어 보내므로 `o.Market`을 그대로 옮긴다.

## State mutations and fallbacks

- 계좌 변경 없음(GET). 클라이언트 상태 변경 없음.
- fallback 없음. B1의 거부는 fallback 신호가 아니다.
- 페이지네이션 순회 없음 — `HasNext`/`NextCursor`를 호출자에게 넘긴다.

## Safety conclusion

- Safe edit boundary: 신규 메서드·신규 타입 가산. `ConditionalOrders`/`ConditionalOrder`/
  `adaptCondition`/`adaptConditionalOrder` 무변경(파일 diff `+128 -0`).
- High-risk impact: **yes** — 계좌 게이트웨이에서 실제 요청을 보내고, 그 답이 잔여
  조건주문 판정에 들어간다. 잔여 조건주문은 노출 상한을 채워 다음 조치를 막으므로,
  0으로 세는 실패는 "안전한 쪽"이 아니다.
  additive인 근거: 기존 심볼 무변경(삭제 0줄), 같은 엔드포인트를 같은 `getAcct`/`send`로
  읽어 401 재시도·계좌 헤더·오류 분류가 동일, 계좌 seq 해석 공유로 요청 수 불변,
  새 거부는 이 change가 만든 심볼 위에서만 일어난다.
