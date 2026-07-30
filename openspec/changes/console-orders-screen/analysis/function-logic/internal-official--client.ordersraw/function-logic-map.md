# Function Logic Map: `Client.OrdersRaw`

- Source: `internal/official/orders_reads.go`
- AST evidence: `ast.json` (revision `current`)
- Risk scan: `risk-pattern-report.md` (매칭 0건)

**신규 메서드**다(HEAD L235-296). 기존 `Client.Orders`를 고치지 않고 옆에 세운 두 번째
읽기이며, 같은 엔드포인트·같은 전송 경로를 쓴다.

## 계좌에 닿는 방식

- 엔드포인트: `GET /api/v1/orders` — `Client.Orders`가 부르는 바로 그것이다.
- rate-limit 그룹: `budgetKey("/api/v1/orders")` = `/api/v1/orders`. `Orders`와 **같은 그룹**이므로
  두 모양이 다 필요한 호출자는 이것 하나를 부르고 변환해야지, 둘을 부르면 §0.4 예산에서
  두 번을 쓴다(doc 주석에 명시).
- 계좌 해석: `getAcct` → `ensureAccountSeq`. 한 클라이언트 안에서 `Orders`와 **공유**되므로
  두 번째 `/api/v1/accounts` 호출이 생기지 않는다.

이 패키지의 모든 계좌 읽기가 공유하는 경로는 하나다:
`getAcct` → `ensureAccountSeq`(`c.mu` 아래 지연 1회 해석, 결과 캐시) →
`getWithHeaders` → `send`. `send`가 401을 만나면 `c.tm.refresh` 후 **정확히 한 번** 재요청하고,
2xx가 아니면 `classifyStatus`가 401/403→`ErrAuth`(본문에 `\bip\b`가 있으면 `ErrIPNotAllowed`),
429→`ErrRateLimited`, ≥500→`ErrServer`, 그 밖의 4xx→`*APIError`(passthrough, fallback 없음)로
사상한다. `ShouldFallback`은 sentinel 넷에만 true다.

## 빈 `Status`는 요청을 보내기 **전에** 거부한다

`strings.TrimSpace(filter.Status) == ""`이면 어떤 요청도 만들지 않고
`ErrOrderStatusRequired`를 감싼 문장을 돌려준다. 이유는 셋이다.

1. openapi가 `status`를 `required: true`로 표시한다.
2. `status`는 `symbol`·`limit` 같은 **필터가 아니라 질문의 모양을 고르는 스위치**다.
   `status=OPEN`은 미체결 전량을 돌려주며 `limit`과 `cursor`를 무시한다 — 이것이
   **잔여물을 구조적으로 놓칠 수 없는 유일한 형태**다. `status=CLOSED`는 페이지네이션하고
   `from`/`to`가 없으면 계좌 전 이력을 훑는다.
3. 이 규칙이 클라이언트에 있는 이유는 막으려는 결함의 모양이 **호출자가 잊었다**이기
   때문이다. `/orders`가 `?limit=100`에 status 없이 나갔고, 그것은 거부된 요청이거나 전 이력
   한 페이지다 — 후자면 101번째 살아 있는 주문이 화면과 건수에서 사라지고 "0건 이상"이라는
   아무도 해소할 수 없는 바닥으로 렌더된다. 호출 지점에만 사는 규칙은 다음 호출 지점에서
   다시 잊힌다.

`ErrOrderStatusRequired`는 retryable로 분류되지 않고 `ShouldFallback`이 false를 답한다 —
이 클라이언트가 보내지 않기로 한 요청은 웹 세션으로 갈 이유가 아니다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `filter.Status` | **필수**, 공백만도 불가. 값은 verbatim 전달 | 호출자 | `ErrOrderStatusRequired`(요청 미발생) |
| `filter.Symbol/From/To/Cursor` | 빈 값이면 파라미터 생략 | 호출자 | — |
| `filter.Limit` | `>0`일 때만 전송 | 호출자 | `status=OPEN`에서는 브로커가 무시 |
| 계좌 헤더 | `X-Tossinvest-Account` | `ensureAccountSeq`(지연 1회, 공유) | 해석 실패는 그대로 오류 |

불변식: 그룹 문자열을 **정규화하지 않는다**. 어떤 그룹이 존재하는지는 브로커가 정하고,
이 읽기는 답의 부재를 거부하지 처음 보는 답을 거부하지 않는다.
십진수는 **문자열 그대로** 옮긴다 — 여기서 `parseDecimal`을 부르면 이 타입의 존재 이유가
그대로 없어진다(빈 문자열이 0이 되어 미체결 주문이 "0에 체결됨"으로 렌더된다).
페이지 경계(`NextCursor`, `HasNext`)를 버리지 않는다 — 잘린 페이지에서 센 건수는
자신 있게 짧은 숫자다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if, L236) | `TrimSpace(Status) == ""` | **없음 — 요청을 만들지 않는다** | `RawOrderList{}, %w(ErrOrderStatusRequired)` | `TestTheRawReadsRefuseARequestWithNoStatusGroup`, `TestOrdersFilterEmptyOmitsEveryParameterIncludingTheRequiredOne` |
| B2 (if, L246) | `Status != ""` | 쿼리에 `status` | — | `TestBothOrderReadsSendTheGroupTheyWereGiven` |
| B3 (if, L249) | `Symbol != ""` | 쿼리에 `symbol` | — | 동상(생략 시 부재) |
| B4 (if, L252) | `From != ""` | 쿼리에 `from` | — | 동상 |
| B5 (if, L255) | `To != ""` | 쿼리에 `to` | — | 동상 |
| B6 (if, L258) | `Cursor != ""` | 쿼리에 `cursor` | — | 동상 |
| B7 (if, L261) | `Limit > 0` | 쿼리에 `limit` | — | 동상 |
| B8 (if, L266) | `getAcct` 실패 | 없음 | `RawOrderList{}, err`(분류된 sentinel/`*APIError`) | `orders_raw_test.go:TestRawReadsClassifyErrorsLikeEveryOtherRead`가 같은 `send` 경로를 잰다 |
| B9 (range, L275) | 페이지의 주문 순회 | 로컬 slice append | — | `TestTheRawOrderReadKeepsAnAbsentValueApartFromAZeroOne` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.TrimSpace` | 공백만인 그룹도 부재로 본다 | 순수 | ast.json calls |
| `fmt.Errorf`(`%w`) | `ErrOrderStatusRequired` 감싸기 | `errors.Is` 가능 | ast.json calls |
| `url.Values.Set` / `strconv.Itoa` | 쿼리 조립 — `Orders`와 같은 순서·같은 조건 | 순수 | ast.json calls |
| `c.getAcct` | 계좌 헤더 + 공유 send | 401 재시도 1회, `classifyStatus` | ast.json calls |
| `marketFromCurrency` | 시장 열 유도(응답에 market 필드가 없다) | 모르는 통화는 빈 문자열 | ast.json calls |
| `make`/`append` | 결과 조립 | — | ast.json calls |

## State mutations and fallbacks

- 계좌 변경 없음(GET). 클라이언트 상태 변경 없음 — `ensureAccountSeq`의 캐시는 `Orders`와
  공유하는 기존 것이고 이 함수가 새로 만드는 것이 아니다.
- fallback 없음. 특히 B1의 거부는 fallback 신호가 아니다(`ShouldFallback` false).
- 페이지네이션 **순회를 하지 않는다**. 한 페이지 뒤에 루프를 두면 "갱신당 N콜" 계약이
  무한대가 되면서도 테스트는 전부 통과한다 — 대신 `HasNext`를 호출자에게 넘긴다.

## Safety conclusion

- Safe edit boundary: 신규 메서드·신규 타입·신규 sentinel 가산. `Orders`/`OrderByID`/
  `adaptOrder`/`adaptOrders` 무변경(파일 diff `+191 -0`).
- High-risk impact: **yes** — 계좌 게이트웨이에서 실제 요청을 보내는 코드이고, 그 답이
  "지금 살아 있는 주문이 무엇인가"를 화면과 잔여물 판정에 공급한다. 잘못 세면 잔여물이
  숨고, 숨은 잔여물은 노출 상한을 채워 다음 조치를 막는다.
  additive인 근거: 기존 심볼 무변경(삭제 0줄), 같은 엔드포인트를 같은 `getAcct`/`send`로
  읽어 401 재시도·계좌 헤더·오류 분류가 동일, 계좌 seq 해석을 공유해 요청 수가 늘지 않음,
  그리고 새 거부(B1)는 **이 change가 만든 심볼 위에서만** 일어나므로 깨질 선행 호출자가 없다.
