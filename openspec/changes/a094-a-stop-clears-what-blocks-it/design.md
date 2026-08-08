# a094 · 설계

> 분기 인용은 전부 `analysis/function-logic/`의 AST 산출물에서 온다
> (함수 **15개** · 분기 **180개** — 3판이 6개 101분기를 더했다).

## D−1. 3판이 고친 것 — **잠금을 잘못 지목하고 있었다**

1판과 2판은 **미정산 attempt**를 동결의 원인으로 봤다. **틀렸다.**
2라운드 리뷰가 잡았고 매니저가 소스와 원장으로 재확인했다(`review.md` §2.1).

**진짜 잠금은 무장된 채 남은 발의(`exit_states.pending_action`)다.**

```text
submit B8  exitloop.go:1296-1300
  case out.State == StateInDoubt || StateUnresolvedInDoubt:
      return nil                      ← release() 를 부르지 않는다
                                        pending_action 이 무장된 채 남는다
        │
        ↓
EvaluateLadder  B25 :439  observed < baseline           ← 손절 조건은 성립한다
                B26 :441  PendingAction == ActionLadderStop
                          → Suppressed = SuppressedPending, return
                                        ← out.Proposal 을 채우지 않는다 = 빈 발의
        │  (RATCHET도 같다 — EvaluateRatchet B17 :423)
        ↓
record  exitloop.go:1082
  orderable := snapshot.Orderable && !proposal.Zero()   → false
        │
        ↓
record  exitloop.go:1117
  if orderable && (CancelPendingFirst || isFullExit)    ← 게이트가 열리지 않는다
        │
        ├─ clearTheSymbol 도달하지 않음   ← R2가 바꾸는 코드
        └─ submit 도달하지 않음           ← R1이 바꾸는 코드
```

**원장 실측 (2026-08-07)** — `pending_intent_id`가 얼어붙은 attempt를 가리킨다:

| 종목 | `pending_action` | `pending_level` | 가리키는 attempt |
| --- | --- | --- | --- |
| **475150** | `STOP_LOSS_LADDER` | `0` | `034e5b79…` `IN_DOUBT` `settled_at=NULL` |
| **080220** | `STOP_LOSS_LADDER` | `-1` | `8f68e7c3…` `IN_DOUBT` `settled_at=NULL` |
| 272210 · 066570 · TSLA | `None` | — | — |

**따라서 1·2판의 R1·R2는 이 두 종목에 대해 실행되지 않는 코드였다.**
272210은 `pending_action`이 비어 있어 매 주기 발의가 나므로 R1·R2가 실제로 돕는다 —
**동결에 두 모양이 있었고 앞선 두 판은 하나만 다뤘다.**

### 그래서 3판이 더하는 고리 하나

`pending_action`을 NULL로 만드는 non-test writer는 셋뿐이다:

| writer | 자리 | 호출자 |
| --- | --- | --- |
| `Journal.ResolveExitProposal` | `apply_hook.go:846-849` (**B9**) | `ExitObserver.release`(`exitloop.go:1317`) **하나** |
| `ApplyTx.ResolvePending` | `apply_hook.go:457-458` | 체결 적용 안 |
| `resetExitStateForReadoptTx` | `apply_hook.go:704-706` | 운영자 `ActionReadopt` |

**어느 것도 이 상태에서 불리지 않는다.** B8이 `release`를 건너뛰었고, 체결은 오지 않고,
운영자는 아직 행동하지 않았다.

**3판의 R3은 그 고리다** — *attempt가 비-CONFIRMED 종결에 이르면 그 intent를 가리키는
발의를 해제한다.* 연결은 이미 원장에 있다(`exit_states.pending_intent_id` →
`mutation_attempts.intent_id`). **이것이 B8 자신의 약속을 참으로 만든다** —
B8의 주석이 *"the proposal stays armed and **the resolver settles it**"*라고 쓰는데,
오늘 그 resolver가 발의까지 settle하지 않는다.

## D0. 하나의 원칙

**브로커가 이름을 준 code는 code가 판단한다. status도 message도 아니다.**

이번 사건이 그 근거를 셋 다 준다.

| 필드 | openapi 계약 | 프로덕션 실물 | 일치 |
| --- | --- | --- | --- |
| status | `422` | `409` | ✗ |
| message | `동일 종목에 반대 방향의 체결 대기 주문이 있습니다.` | `반대 포지션 미체결 주문이 존재합니다.` | ✗ |
| **code** | `opposite-pending-order-exists` | `opposite-pending-order-exists` | **✓** |

세 요청(`6GKYatiUehps5SQX`·`7d3we7ZD3dtxWTMO`·`7k5oRgmEHnoU5Vfi`)에서 같다.

**기존 코드는 이 원칙을 절반만 따른다.** `classifyRefusalBody`(`failclosed.go:223-238`)는
code 토큰(`trade_auth_required`·`fx_consent`·`funding_required`)과 **한국어 message
조각**(`"거래 인증"`·`"환전 동의"`·`"입금"`)을 **함께** 매칭한다. 위 표가 보이듯
message는 계약과 어긋날 수 있으므로 message 매칭은 같은 종류의 취약점이다.
**이 change는 새 항목을 code로만 건다.** 기존 세 항목의 message 매칭은 건드리지 않는다 —
그것을 지우면 지금 잡히던 것이 안 잡히는 방향이고, 이 change는 보수 방향만 취한다.
**그 취약점은 남는다는 사실을 여기 적는다**(침묵한 생략 아님).

## D1. R1 — 어디에 거는가

### 왜 `isDefinitiveRejection`이 아닌가

`journal.isDefinitiveRejection`(`dispatch.go:349-356`, 분기 3)에 409를 더하는 것은
**금지된 방향이다.** 409는 재생 경로의 최빈 응답이기도 하다:

> `409 request-in-progress`(openapi) → 원 요청 처리 중, 대기 후 재시도(상한 미소비)
> — `openspec/specs/order-execution/spec.md` IN_DOUBT 해소 §1

`request-in-progress`는 **원본이 실행됐을 수 있다.** 409 전체를 확정 거절로 바꾸면
그것을 "실행 안 됨"으로 확정하게 되고, 그 오류는 **살아 있는 주문을 은퇴시킨다** —
정확히 반대 방향의 위험이다. 상태 코드 표는 건드리지 않는다.

### 왜 `classifyRefusalBody`인가

`execgw.classifyMutation`(`classify.go:21-79`, 분기 7)의 순서가 이미 맞다.

| 분기 | 자리 | 순서상 의미 |
| --- | --- | --- |
| **B2** | `:32` | `policyRefusal` — 로컬 거부, 브로커 접촉 전 |
| **B3** | `:46` | `ClassifyBrokerRefusal` — **브로커가 요청을 서술하며 거절** |
| **B4** | `:49` | post-prepare 확인만 `DispatchAmbiguous`로 승격 |
| **B5** | `:64` | `statusOf` → `ClassifyHTTPMutation` — **status가 말하게 하는 마지막 수단** |

B3의 주석이 근거를 갖는다:

> *It is classified here, **before the transport tracker gets a say**, because the
> meaning comes from **the answer** and not from how far the bytes got.*

`opposite-pending-order-exists`는 정확히 그런 답이다. B3에 걸리면 B5는 **돌지 않고**,
`class = journal.DispatchRejected`가 되어 attempt가 **종결**한다.

### 무엇을 더하는가

1. `execgw`에 reason code 하나 — `ReasonOppositePendingOrder`.
   `AllReasonCodes()`(`failclosed.go:254`)에도 넣는다. 그 함수의 주석이 이유를 쓴다:
   *"these strings land in the journal and in operator alerts … a rename is a data
   migration and not a refactor."*
2. **본문의 `code` 필드를 파싱**해서 그 값이 `opposite-pending-order-exists`일 때만 건다.
   본문 통짜 substring 매칭이 **아니다**.

`refusalBody(err)`(`failclosed.go:196`)는 이미 브로커 자신의 본문만 읽는다
(*"It never falls back to err.Error()"*). 우리가 만든 문자열을 매칭할 위험이 없다.

### 왜 한 줄(`containsAny`)로는 안 되는가 — 이 delta 자신이 그것을 금지한다

`classifyRefusalBody`(`failclosed.go:223-238`)는 **본문 통짜에 대한 `strings.Contains`**다.
여기에 case 하나를 더하면 `message`에 그 문자열이 들어 있는 본문도 잡힌다.
그런데 이 change의 spec delta는 *"message 문구로 걸어서는 안 된다(SHALL NOT)"*를 쓴다.
**한 줄로는 자기 SHALL NOT을 만족할 수 없다.** tasks 2.5가 그것을 시험한다.

**기존 세 항목이 그 취약성의 증거다.** `"interactive"`는 code 토큰이 아니라 아무 본문에나
나올 수 있는 영어 단어인데 지금 `trade_auth_required`와 같은 case에 묶여 있다.
이 change는 그것을 고치지 않지만(§6 — 지우는 방향은 지금 잡히는 것을 놓친다),
**새 항목을 같은 방식으로 만들지도 않는다.**

### 본문 모양은 하나가 아니다 — 실측

| 출처 | 실물 | code 자리 | 표기 |
| --- | --- | --- | --- |
| `testdata/interactive_auth_challenge.json` | `{"code":"TRADE_AUTH_REQUIRED","message":…}` | **최상위** | UPPER_SNAKE |
| `testdata/fx_consent_required.json` | `{"code":"FX_CONSENT_REQUIRED",…}` | **최상위** | UPPER_SNAKE |
| **프로덕션 409 3건** (원장 `mutation_attempts.detail`) | `{"error":{"requestId":"7k5oRgmEHnoU5Vfi","code":"opposite-pending-order-exists","message":"반대 포지션 미체결 주문이 존재합니다."}}` | **`error` 아래** | lower-hyphen |

**두 가지가 동시에 다르다** — 자리(최상위 vs `error.`)와 표기(UPPER_SNAKE vs
lower-hyphen). 따라서 파서는 둘 다 읽고, 값 비교는 **대소문자 무시 + 전체 일치**다
(substring 아님). 이 표가 D0의 "code만이 안정 필드"를 한 겹 더 좁힌다 — **code의
*값*은 안정적이지만 그 *자리*와 *표기*는 아니다.**

### 그래서 R1은 함수 하나를 더한다

`classifyRefusalCode(body string) (ReasonCode, bool)`

- 본문을 JSON으로 읽어 `code`와 `error.code`만 본다. 다른 필드는 보지 않는다.
- JSON 파싱 실패·필드 부재·빈 값 → **분류하지 않음**(fail-closed: 모르면 종전 경로).
- `classifyRefusalBody`보다 **먼저** 부른다. 기존 세 항목의 substring 매칭은 그대로 둔다.

`internal/execgw/testdata/reason_codes.golden`는 새 code 하나만큼 늘어난다.
갱신은 `TOSSOS_UPDATE_GOLDEN=1 go test -run TestWriteReasonCodeGolden`이며 tasks 2.10이
그것을 명시한다 — 이 파일은 손으로 고치는 파일이 아니다.

### 재생 경계 — 지금 우연히 안전한 것을 계약으로 바꾼다

`classifyReplay`(`replay.go`)는 `classifyMutation`과 코드를 공유하지 않는다. 따라서
R1은 오늘 재생 분류를 오염시키지 않는다 — 그리고 그것은 **우연이 아니다**. `classifyReplay`의
default 분기가 정책을 의도적으로 적는다: *"A first dispatch would call several of these
definitive refusals; a replay may not"*(`replay.go:517-520`). **계약 공백은 여전히 실재한다** —
그 문장은 오늘의 동작을 설명할 뿐 새 code를 구속하지 않는다. 재생 attestation이 켜지는 날
이 code가 재생 응답에 붙으면 "원 요청이 접수 안 됨"으로 확정되어 **정확히 반대 방향**으로
작동한다(재생의 409는 `request-in-progress` 계열이다). spec에 SHALL NOT으로 못 박는다.

### R1이 바꾸는 관측 가능한 결과

| | 전 | 후 |
| --- | --- | --- |
| attempt 상태 | `IN_DOUBT`, `settled_at=NULL` | `FAILED_CONFIRMED`(종결) |
| `PendingAttempts` | 포함 → `checkSymbolFree` B4가 전 주문 차단 | 제외 → 차단 없음 |
| `submit`의 분기 | **B8** `:1296` — 제안 걸린 채 `return nil` | **B10** `:1304` — 알림 + `release(ProposalRefused)` → 레벨 재무장 |
| 운영자 | 아무 소리 없음 | `alertProposalRefused` |

**R1만으로는 팔리지 않는다.** 재무장된 레벨이 다음 주기에 같은 409를 받는다. R1은
*영구 동결*을 *보이고 세어지는 반복*으로 바꾼다. 파는 것은 R2다.

### R1의 소급 — 저장된 IN_DOUBT를 같은 증거로 다시 읽는다 (3판)

**위 표는 앞으로 오는 409에 대해서만 참이다.** 이미 저장된 attempt는 분류가 끝났고,
D−1이 보였듯 그 종목에서는 다음 dispatch 자체가 일어나지 않으므로 **R1이 영영 닿지 않는다.**

**그런데 판단 근거가 원장에 그대로 있다.** `mutation_attempts.detail` 실물:

```text
HTTP 409 does not prove whether the mutation executed: official: API error 409:
{"error":{"requestId":"7k5oRgmEHnoU5Vfi",
          "code":"opposite-pending-order-exists",
          "message":"반대 포지션 미체결 주문이 존재합니다."}}
```

**따라서 기동 시 1회, 저장된 `IN_DOUBT` attempt 중 본문에 확정 거절 code가 있는 것을
`FAILED_CONFIRMED`로 재분류한다.**

**안전 논거는 R1 자신과 같다.** `opposite-pending-order-exists`는 브로커의 **주문 전
검증 거절**이며 그 상태에서 주문은 접수되지 않는다. 같은 증거를 나중에 읽는 것뿐이고,
살아 있는 주문을 은퇴시키는 위험이 없다. openapi가 이 code를 422(확정 거절군)에 둔 것이
같은 판단이다.

**경계 (SHALL NOT)**:

- **code로만 판단한다.** 저장된 `detail`의 **본문 부분**을 파싱하고, 우리가 앞에 붙인
  산문(`"HTTP 409 does not prove…"`)을 매칭 대상으로 삼지 않는다
- **확정 거절 code가 없는 IN_DOUBT는 건드리지 않는다.** `request-in-progress`를 포함해
  모호한 것은 모호한 채 둔다
- **재분류는 attempt 상태만 바꾼다.** 발의 해제는 R3의 고리가 한다 — 두 쓰기를 한
  함수에 섞지 않는다
- **기동 시 1회.** 주기적으로 돌지 않는다 — 새로 생기는 409는 R1이 dispatch 시점에 잡는다

**이것이 475150·080220을 실제로 녹이는 부분이다**: 재분류 → 종결 → R3의 고리가 발의 해제
→ 다음 주기에 발의 → R2가 막고 있던 매수를 치움 → 손절 제출.

## D2. R2 — 청소가 브로커를 본다

### 지금의 눈

`engine.clearTheSymbol`(`exitloop.go:1334-1392`, 분기 9)

| 분기 | 자리 | 역할 |
| --- | --- | --- |
| **B2** | `:1341` | `for _, order := range live` — 목록 순회 |
| **B3** | `:1343` | `if !buy && !withPending { continue }` — **매수(`buy`)는 항상 치우고, `withPending`이 필요한 것은 자기 방향(매도)이다** |
| **B6** | `:1379` | `if err != nil \|\| out.State != journal.StateConfirmed` — 취소 확정 실패는 `clear=false` |
| **B7** | `:1383` | `if !clear` — 하나라도 못 치우면 제출하지 않는다 |
| **B8** | `:1386` | `if withPending && m.state.Pending()` — 마지막에 제안 해제 |

`live`의 원천 `Journal.LiveOrdersForSymbol`(`fills.go:1849-1927`, 분기 7)은
`mutation_attempts JOIN intents`다. **엔진 밖 주문은 0행이다.**

### 넓히는 방식

`live`를 **저널 ∪ 브로커 미체결**로 만든다.

- **저널에 있는 주문**: 지금 그대로. lineage 해소(`fills.go`의 주석 —
  *"the official API answers a modify with a new order number"*)를 유지한다.
- **저널에 없는 주문**: 브로커가 보고한 `orderId`로 취소한다. lineage를 만들 수 없으므로
  **정정(AMEND) 대상이 아니고 취소만 한다.**
- **양쪽에 있는 주문**: `orderId`로 dedup. 이것은 새 규칙이 아니라 IN_DOUBT 조회 대조가
  이미 쓰는 규칙이다(`spec order-execution` §2 — *"유일성 판정 전에 orderId 기준 dedup"*).

### 배선 — 「새 API 표면 없음」은 브로커 API에 대해서만 참이다

1판 tasks 3.9가 *"새 API 표면 없음"*이라 썼다. **엔진 API에 대해서는 거짓이다.**
`OrderPager`(`indoubt.go:70-74`)는 `reconcile`·`filldetect`·`flatten`에만 배선돼 있고
`Context.ExitObserver`(`exitwiring.go:319-347`)가 채우는 필드에 **없다.** 필요한 자리 넷:

| # | 자리 | 무엇을 |
| --- | --- | --- |
| 1 | `ExitObserverOptions`(`exitloop.go:166-223`) | `Orders execgw.OrderPager` 필드 하나 |
| 2 | `NewExitObserver`의 거부 switch(`exitloop.go:263-282`) | **nil을 허용한다** — 필수로 하면 미배선 빌드가 기동조차 못 한다. nil이면 브로커를 보지 않고 **저널만으로 오늘과 똑같이** 동작한다 |
| 3 | `Context.ExitObserver`(`exitwiring.go:338-346`) | `opts.Orders == nil`이면 `execgw.OfficialOrders{Client: c.Official}` — `Floor`·`Names`·`Alerts`가 이미 쓰는 「nil이면 채운다」 형태 그대로 |
| 4 | `cmd/tossctl/engine.go:346` | 구성 리터럴은 **손대지 않는다**(3이 채우므로) |

**nil 허용은 이 change의 토글이 아니다.** 토글은 도입하지 않는다(tasks 7.4). nil 경로는
`ExitObserver`를 직접 만드는 **테스트**의 경로이고, 프로덕션 경로인 `Context.ExitObserver`는
항상 채운다. 그 차이를 tasks 3.11이 시험한다.

### ~~파싱 — 기존 함수 둘 다 모자란다~~ — **열거가 불완전했다 (3판)**

> **아래 표는 유효하지만 완전하지 않다.** 세 번째 후보 `filldetect.parseSnapshot`을
> 평가하지 않았고, 그것이 7필드 중 5개를 이미 준다. **결론은 다음 절(§0.4)이 대체한다** —
> 새 파서가 아니라 기존 것의 확장이다. 이 표는 「왜 `brokerstate`와 `official`로는
> 안 되는가」의 근거로만 남긴다.

#### (1·2판) 두 후보의 한계 — 유효

`clearTheSymbol`은 한 주문에서 **7개 필드**를 쓴다 — `Side`(B3 술어), `OrderID`,
`Symbol`, `Market`, `Quantity`, `Price`, `Currency`(`CancelRequest.Order`).

| 후보 | 주는 것 | 모자란 것 |
| --- | --- | --- |
| `brokerstate.ParseOfficialOrder`(`derive.go:645`) | `OrderID`·`RawStatus`·`Canceled`·`CanceledAt`·`Quantity`·`FilledQuantity` | **`Side`·`Symbol`·`Market`·`Price`·`Currency` 전부 없다.** `officialOrderPayload`(`derive.go:623-631`)가 6필드짜리 의도적 부분 미러다 |
| `official.Client.Orders()` → `domain.Order` | `Side`·`Symbol`·`Market`·`Price`는 있다 | **첫 페이지만 남긴다**(`indoubt.go:726` — *"Upstream's Orders() keeps only the first page"*). 부분 목록은 목록이 아니다 |

따라서 **새 파서 하나를 만든다** — `execgw`에 `ParseWorkingOrder(json.RawMessage)`.
`ScanOrders`(`indoubt.go:730`)가 이미 raw JSON을 완주해서 주므로 페이지 문제는 없고,
새로 필요한 것은 그 raw에서 위 7필드를 읽는 일뿐이다. 이것이 **새 엔진 API 표면**이며
tasks 3.9의 1판 문장을 정정한다.

**status 그룹**: `OrderQuery.Status = "OPEN"` 하나만 쓴다.
openapi가 그것을 정의한다(`indoubt.go:394-397`): *"OPEN: 진행 중 주문 그룹 —
`orders[].status ∈ {PENDING, PARTIAL_FILLED, PENDING_CANCEL, PENDING_REPLACE}`"*.
CLOSED는 조회하지 않는다 — 치울 대상은 미체결뿐이고, CLOSED를 더하면 조회가 두 배가
되면서 얻는 것이 없다. `PENDING_CANCEL`이 목록에 있으면 취소가 이미 진행 중이므로
새 취소를 내지 않고 `clear=false`로 둔다(그 주문은 아직 책 위에 있다).

**종목 필터는 서버측이다** — `OfficialOrders.OrdersPageRaw`(`orders_source.go:47-53`)가
`OrderQuery.Symbol`을 `official.OrdersFilter.Symbol`로 넘기고 그것이 쿼리 파라미터
`symbol`이 된다(`orders_reads.go:93-95`). 계좌 전체를 끌어와 걸러내는 것이 아니다.

### §0.4 — **동기 조회를 넣지 않는다. 이미 있는 스냅샷을 읽는다** (3판)

2판은 손절 경로에 **2초 타임아웃의 동기 브로커 왕복**을 넣고 그것을 §0.3 비용으로
받아들였다. **2라운드가 그 비용이 불필요함을 보였다**(`review.md` §2.8).

**`filldetect.Detector`가 이미 3초마다 계정 전체 OPEN 목록을 완주한다:**

| 사실 | 자리 |
| --- | --- |
| 프로덕션에서 **무조건** 돈다 | `cmd/tossctl/engine.go:391-396` — `{Name: "filldetect", Run: detector.Run, …}` |
| OPEN 목록을 완주한다 | `Detector.collect`(`detect.go:363-…`, 분기 19) **B1** `:365` 직전 — `ScanOrders(ctx, d.Orders, OrderQuery{Status: statusOpen}, cfg.MaxPages)` |
| 주기 3초 | `detect.go:128` — `PollInterval: 3 * time.Second` |
| 원천이 **같다** | `engine.go:420` — `Orders: execgw.OfficialOrders{Client: ectx.Official}` |

**2판이 넣으려던 것과 같은 API, 같은 pager, 같은 호출이다.**

**빈도 실측이 그 낭비를 못 박는다.** `clearTheSymbol`은 모든 full exit 직전에 돌고
주기는 5초(`exitloop.go:97`)다. proposal 자신의 측정: 272210이
`STOP_LOSS_LADDER → PROPOSAL_CANCELLED`를 **2h54m 동안 1931회, 중앙값 5.0초**로 반복했다.
**2판대로면 그것이 전부 브로커 OPEN 조회다** — 5종목이 발의 중이면 ~1 req/s로,
detector의 ~0.33 req/s 위에 **약 4배**다.

**3판의 결정: `clearTheSymbol`은 detector의 마지막 OPEN 스냅샷을 읽는다.**

| | 2판 (동기 조회) | 3판 (스냅샷) |
| --- | --- | --- |
| §0.3 손절 지연 | **최대 2초** | **~0** (메모리 읽기) |
| §0.4 새 호출 | 5초마다 종목당 1회 | **0** |
| 데이터 신선도 | 즉시 | **최대 3초** |
| 종목 필터 | 서버측 | **클라이언트측** (스냅샷은 계정 전체) |
| 목록 못 얻을 때 | 조회 실패·타임아웃 | **스냅샷 부재·상한 초과 노후** |

**신선도 상한은 5초로 둔다** — detector 주기 3초의 여유 1회분이다. 그보다 오래된
스냅샷은 「목록을 얻지 못했다」로 취급한다(아래 절).

**파서도 다시 만들지 않는다.** 2판의 「기존 함수 둘 다 모자란다」 표가
`filldetect.parseSnapshot`을 평가하지 않았다. 그 `Snapshot`(`detect.go:194-224`)은
`OrderID`·`Symbol`·`Market`·`Side`·`Quantity`를 이미 준다 — `clearTheSymbol`이 쓰는
**7필드 중 5개**다. 없는 것은 주문 `Price`와 `Currency`뿐이고(있는 `AveragePrice`는
체결가다), **그것은 확장이 답이지 세 번째 파서가 답이 아니다.**

**상태 판정도 이미 있다** — `PENDING_CANCEL`은 `brokerstate.StateCancelPending`
(`derive.go:421`)으로 모델링돼 있고 `Snapshot.Derived`가 그것을 싣는다. 문자열 비교를
새로 쓰지 않는다.

**배선 선례도 바로잡는다.** 2판 표는 `Context.ExitObserver`가 nil-fill하고
`cmd/tossctl/engine.go`는 손대지 않는다고 정했다. **detector 파생 의존성의 기존 선례는
정확히 거기서 주입된다** — `engine.go:349` `SLO: detectorPressure{detector: detector}`이고
`Context.ExitObserver`는 `opts.SLO`를 건드리지 않는다. 3판은 그 선례를 따라
**`engine.go`에서 주입한다.**

`exitwiring.go:313-317`의 doc comment는 **stale이다** —
*"this build constructs no fill detector: there is no production polling loop to defer to
yet"*라고 쓰는데 빌드는 만들고 넘긴다. 3판이 그 주석도 고친다.


### 브로커 오류는 `clearTheSymbol` 안에서 흡수한다

1판은 오류를 `record`로 반환했다. `record`(`exitloop.go:1141-1144`)는
`cleared, err := o.clearTheSymbol(...)` 다음 `if err != nil { return err }`다.
**브로커 두절 한 번이 `RecordExitJudgementResult` 전에 판정을 통째로 중단시킨다** —
워터마크도 기준선도 전진하지 않고 `noteDelay`도 울리지 않는다. `exitloop.go:1113-1116`이
명시한 계약(*"A clear that does not complete does not stop the judgement"*)보다
**엄격히 나쁘다.**

그래서 브로커 목록 조회의 실패·타임아웃·페이지 초과는 **`clearTheSymbol` 내부에서**
흡수한다:

1. 저널분 청소는 **그대로 수행한다** (오늘과 동일).
2. 브로커분을 보지 못했다는 사실을 `clear = false`로 만든다.
3. `record`의 기존 `!cleared` 가지(`exitloop.go:1145-1148`)가 받는다 —
   `noteDelay` + `ArmSuppressedWorkingOrder`. **알림이 울리고 판정은 계속된다.**

즉 새 실패 모드를 만들지 않고 **이미 있는 실패 모드로 떨어뜨린다.**

**spec의 SHALL을 그 형태로 고친다.** 1판 delta는 *"브로커 미체결을 포함해야 한다(SHALL)"*로
폴백을 허용하지 않았고 tasks 3.7은 *"조회 실패해도 저널분 청소는 진행"*을 요구했다 —
**둘은 동시에 참일 수 없었다.** 새 문장은 「목록을 보지 못하면 제출하지 않는다」다.

### 자기 방향 미체결의 부재 확인 — **3판에서 철회했다**

2판은 「`withPending`과 무관하게 자기 방향(SELL) 미체결이 있으면 제출하지 않는다」를
넣었다. **2라운드가 그것을 차단했고 근거가 맞다**(`review.md` §2.2).

**철회 사유 1 — 막으려던 초과 매도는 이미 막혀 있다.**
`armExitProposalTx`(`apply_hook.go:655-676`, 분기 4) **B3** `:666`:

```go
if strings.TrimSpace(action.String) != "" {
    return fmt.Errorf("%w: %s holds %s", ErrProposalPending, positionID, action.String)
}
```

주석이 논거를 쓴다 — *"A second proposal while one is outstanding is refused rather than
overwritten."* **엔진 자신의 두 번째 매도는 여기서 거부된다.**

**철회 사유 2 — 남는 효과가 「보호를 영구 withhold」뿐이다.**
이 검사가 추가로 잡는 유일한 경우는 **사용자가 앱에 넣은 매도**다. R1이 발의를 해제한
다음 주기부터 `CancelPendingFirst = false`이므로(`ladder.go:447`) 청소는
`withPending=false`로 불리고, R2는 **반대 방향만 취소한다**(B3 `:1343`).
따라서 그 매도는 취소되지 않은 채 남고, 검사는 **매 주기 제출을 막는다.**

**사용자의 지정가 매도 하나가 그 종목의 손절을 영구히 막는다.** 오늘은 그것을 건너뛰고
제출한다. 이 계정의 사건이 정확히 「앱에서 직접 넣은 주문」이었다.

**따라서 R2는 반대 방향 취소만 한다.** 자기 방향은 오늘과 같이 `withPending`일 때만 본다.

### 취소할 수 없는 주문이 손절을 영구 보류시키는 문제 (차단 8) — 결정

오늘 `clear=false`의 후보는 **엔진 자신의 확정 주문뿐**이다. R2 이후 후보는 브로커가
보고하는 무엇이든이고, 엔진이 영원히 취소할 수 없는 주문이 손절을 **영구 보류**시킬 수
있다. 구체 경로 둘:

- (a) 목록 조회와 취소 사이의 체결 → 취소 거절 → `clear=false`
- (b) `floatOf`(`exitloop.go:1673-1679`)가 `ParseFloat("")`을 거절하므로 **가격 없는
  주문**(시장가)이 목록에 오면 `clear=false`가 고정된다

**결정: 빈 가격은 실패가 아니다 — 그리고 3판은 이것을 저널분에도 적용한다.**
2판은 *"`floatOf` 거절 경로는 저널분에만 남긴다"*고 했다. **2라운드가 a087과의 충돌을
잡았다**(`review.md` §2.4): `LiveOrdersForSymbol`은 `coalesce(i.price,'') AS price`
(`fills.go:1859`)이고 **a087이 보호 청산을 시장가로 바꾸면 저널분에 빈 가격 행이 생긴다** —
2판이 실패를 남겨 둔 바로 그쪽이다.

따라서 빈 `price`는 **양쪽 모두** 0으로 읽는다. `execgw.OrderRef.Price`는 취소 요청의
식별에 쓰이지 않으므로 안전하다. **이것으로 a087 선후 관계 제약이 사라진다.**

**결정: (a)는 보류하되 무한 보류하지 않는다.** 같은 종목에서 청소가 **연속 3회**
`clear=false`로 끝나면 `EventExitLiquidationDelayed`를 **critical**로 올린다.
그 뒤에도 자동으로 제출하지는 않는다 — 「못 치우면 팔지 않는다」(B7)를 뒤집는 것은
초과 매도 방향이고 §6에 걸린다. **바꾸는 것은 침묵의 길이지 규칙이 아니다.**

**단, 엔진 자신이 방금 낸 취소가 정산 중인 것(`PENDING_CANCEL`)은 그 카운터에서
제외한다** — 그것은 「치우지 못하는 중」이 아니라 「치우는 중」이다. 5초 주기에서
15초 넘게 걸리는 정상 취소가 critical을 만드는 것은 거짓 경보이고, critical 전달 실패는
`ENTRY_BLOCKED`까지 간다(`notifier.go:216-218`). 상태 판정은 문자열 비교가 아니라
`brokerstate.StateCancelPending`(`derive.go:421`)을 쓴다.


### 경계 — 새 권한이 아니라 기존 권한의 눈

| 조건 | 이유 |
| --- | --- |
| `clearTheSymbol`이 불릴 때만 | `record` **B3** `:1117`이 이미 게이트다 — `orderable && (CancelPendingFirst \|\| isFullExit)` |
| 같은 계좌·시장·종목 | 충돌의 정의. 종목 필터는 서버측 |
| 반대 방향(`buy`)은 취소, 자기 방향(SELL)은 **부재 확인** | **B3** `:1343` + 위 절 |
| 취소만. 신규·정정 없음 | 청소는 노출을 늘리지 않는다 |
| 취소 확정 실패 → `clear=false` | **B6** `:1379`의 기존 규칙 그대로. **못 치우면 팔지 않는다** |
| 브로커 조회 실패 → `clear=false` | 새 실패 모드가 아니라 기존 모드로의 낙하 |

**노출을 늘리는 주문은 이 경로로 나가지 않는다.** `raisesExposure`
(`gateway.go:377` — `side == "buy"`)는 취소(`gateway.go:416` — `false`)에 적용되지 않고,
청소는 취소만 낸다.

### 감사

엔진이 내지 않은 주문을 취소하는 것은 **기록되어야 한다.** 취소 intent에 그 사실을
남긴다 — 저널에 원본 intent가 없는 주문을 치웠다는 것, 그리고 그것이 어떤 보호 청산을
위한 것이었는지. `CLAUDE.md`「runtime config 변경은 audit로 추적 가능해야 한다」의
같은 이유가 여기에도 적용된다.

## D3. R3 — **종결된 attempt가 발의를 푼다** (3판에서 전면 교체)

### 무엇을 교체했는가

1·2판의 R3은 「세션 중 IN_DOUBT 해소」였다. **2라운드가 그것이 동결을 풀지 못함을
보였다**(`review.md` §2.1) — `Resolve`는 `mutation_attempts`에만 쓰고 `exit_states`는
건드리지 않는다. attempt를 park시켜도 `pending_action`은 무장된 채이고 사다리는 계속
억제한다.

**3판의 R3은 그 빠진 고리다.**

### 고리 하나

```text
attempt 가 비-CONFIRMED 종결에 이른다
  (FAILED_CONFIRMED · NOT_DISPATCHED · UNRESOLVED_IN_DOUBT)
        │
        ↓
그 attempt 의 intent_id 로 exit_states 를 찾는다
  exit_states.pending_intent_id = attempt.intent_id     ← 이미 있는 연결
        │
        ↓
Journal.ResolveExitProposal(positionID, ProposalRefused)
  B8  :842  이미 비었으면 무동작 (멱등)
  B9  :846  pending_action / pending_level / pending_intent_id → NULL
  B10 :852  LADDER 면 rung 되돌림
  B13 :859  exit_events 1행
        │
        ↓
다음 관측: EvaluateLadder B26 :441 이 억제하지 않는다 → 발의가 난다
        └─→ record :1117 게이트 열림 → clearTheSymbol(R2) → submit(R1)
```

### 안전 확인 — `ResolveExitProposal`이 무엇을 쓰는가 (분기 14, AST)

| 분기 | 자리 | 무엇을 | 안전 |
| --- | --- | --- | --- |
| **B8** | `:842` | `pending_action`이 비면 **무동작 반환** | **멱등** — 중복 호출이 무해하다 |
| **B9** | `:846` | 세 컬럼 → NULL | 해동의 실제 지점 |
| **B10·B11** | `:852`·`:853` | LADDER면 `RungIndex(pending_level)` | 음수 label은 거부된다(`ladder.go:536-538`) → 되돌림 없음 |
| **B12** | `:854` | `rollBackRungTx(rung-1)` | **`active_rung`만 쓴다**(`exit_state.go:980-987`). **손절 가격은 건드리지 않는다** |
| **B13** | `:859` | `exit_events` 1행 | 감사 흔적 |

**§6 확인**: 이 경로에서 `entry_price`·`initial_stop`·`baseline_price` 중 어느 것도
쓰이지 않는다. **손절 가격이 움직이지 않는다.**

원장 실측이 두 경우를 다 보인다 — 475150은 `pending_level="0"` → rung 0에서
되돌림, 080220은 `pending_level="-1"` → `RungIndex` 거부 → 되돌림 없이 해제.
**둘 다 손절 가격은 그대로다.**

### 왜 B8의 안전 논거를 깨지 않는가

`submit` B8의 주석:

> *The order may exist. The proposal stays armed and **the resolver settles it**;
> releasing here would let the next observation submit a second sell on top of one that
> may already be live.*

**그 논거는 「종결 전에 풀면 안 된다」이지 「영원히 풀지 말라」가 아니다.**
3판은 **종결된 뒤에만** 푼다:

| attempt 종결 상태 | 의미 | 발의 |
| --- | --- | --- |
| `CONFIRMED` | 주문이 **실제로 나갔다** | **해제하지 않는다** — 체결 경로가 처리한다 |
| `FAILED_CONFIRMED` | 접수되지 **않았음이 확정** | **해제한다** — 살아 있는 주문이 없다 |
| `NOT_DISPATCHED` | 전송조차 안 됨 | **해제한다** |
| `UNRESOLVED_IN_DOUBT` | **모른다** | **아래 참조** |

**`UNRESOLVED_IN_DOUBT`가 유일한 판단 지점이다.** B8의 논거대로면 모르는 상태에서
푸는 것은 초과 매도 위험이다. **그러나 오늘의 대안은 「영원히 무보호」다.**

**3판의 결정: park에서도 해제한다. 근거 셋.**

1. **초과 매도의 1차 방벽은 발의가 아니라 `armExitProposalTx`다**(`:666`) — 그리고
   그것은 발의를 해제해도 다음 발의가 무장될 때 다시 검사한다. 해제는 「두 번째 발의를
   허용」하는 것이 아니라 「첫 발의를 끝낸 것으로 기록」하는 것이다.
2. **park는 그 자체로 계정 전역 진입을 막는다**(`indoubt.go:379-382`). 즉 그 상태는
   이미 사람의 개입을 요구하는 상태이고, 그 위에 **손절만 막아 두는 것**은 방향이 거꾸로다.
3. **§0.3이 이것을 결정한다.** 「모르는 매도가 있을 수 있다」와 「확실히 보호가 없다」
   중 후자가 더 나쁘다. 안전 불변식 §4가 *"손절·비상 청산의 즉시성을 약화하거나 지연하지
   않는다"*이고, 영구 억제는 무한 지연이다.

**단, 이 결정은 R1의 소급과 짝일 때만 이 사건에 적용된다.** 475150·080220은
소급 재분류로 `FAILED_CONFIRMED`가 되므로 **위 표의 둘째 행**으로 들어간다 —
`UNRESOLVED_IN_DOUBT` 판단에 기대지 않는다. **park 해제는 일반 규칙이고, 이 사건의
해동은 그것에 의존하지 않는다.**

### 어디에 배선하는가

**`reconcile.Run`·`Journal.RecoverPending`을 재사용하지 않는다(SHALL NOT).**
`RecoverPending`(`journal/recovery.go:86-125`, 분기 10)은 세션 중에 부르면 원장을
위조한다 — **B4** `:95` `StateRecorded` → *"found at startup with no dispatch recorded"*로
**종결**시키고, **B6** `:103` `StateDispatchStarted` → *"process stopped after dispatch
started"*를 쓴다. 둘 다 세션 중에는 거짓이다.

**3판이 필요로 하는 것은 그 함수가 아니다.** attempt를 종결시키는 자리는 이미 있다:

| 자리 | 언제 |
| --- | --- |
| `classifyMutation` → `submit` | dispatch 시점 (R1이 여기를 고친다) |
| `Resolver.Resolve` | 재시작 복구 · (별도 change의) 세션 중 해소 |
| R1의 소급 재분류 | 기동 시 1회 (3판이 더한다) |

**R3의 고리는 그 셋 뒤에 붙는 하나의 후처리다** — attempt가 종결되면 그 intent를 가리키는
발의를 찾아 해제한다. **새 루프도, 새 주기도, 새 브로커 조회도 없다.**

### 세션 중 해소는 **별도 change로 분리한다**

1·2판의 R3이 가진 나머지(주기 · `Context.Resolver` 배선 · `resolveCancel`의 `r.Order`
요구 · 계정 전역 park의 운영 결과 · §0.4 계측)는 **이 사건의 인과 경로에 없다.**
proposal 자신이 *"R3만: 이 사건의 원인이 남아 매 주기 다시 언다"*라고 인정했다.

`issues.md`가 선행 조건을 기록한다. **침묵한 생략이 아니다.**

## D4. R4 — baseline — **철회했다 (1라운드 차단 2·3)**

> **아래 절은 진단으로만 남긴다.** 「무엇을 채우는가」의 처방은 **폐기**했다.
> 철회 사유는 이 절 끝의 「왜 철회했는가」에 있고, 판정 원문은 `review.md` §1.3·§1.7,
> 후속 조건은 `tasks.md` §5와 `issues.md`에 있다.
>
> **진단은 유효하다** — baseline이 비어 있어 부재가 증명되지 않는다는 것은 사실이고
> D3의 「이득이 무엇인지 정확히」가 그 사실에 의존한다. 틀린 것은 **그것을 지금 채우자**는
> 처방이다.

### 지금 무슨 일이 일어나는가

`gateway.go:1044` `Notes: EncodeBaseline(plan.baseline)`,
`plan.baseline`은 `gateway.go:378` `baseline: req.Baseline`.

**`Baseline`을 넘기는 호출자가 없다.** `exitloop.go:1281-1286`에도
`tracer.go:442-448`에도 그 필드가 없다. 따라서 `notes`는 항상 빈 문자열이고,
`absenceCorroborated`(`indoubt.go:445-…`)의

```go
baseline, ok := DecodeBaseline(intent.Notes)
if !ok {
    return false, "absence cannot be corroborated: the mutation was submitted without a pre-dispatch account baseline"
}
```

가 **항상** 실패한다. 원장이 그것을 확인한다 — 얼어붙은 세 attempt의 `notes` 전부 0바이트.

`indoubt.go:93-94`가 대가를 명시한다:

> *its absence can never be proven, and it will park as UNRESOLVED_IN_DOUBT.
> **That is the documented price of omitting it.***

### 이것도 적합성 수정이다

spec은 부재 판정을 이미 요구한다:

> **3. 부재 판정**: 최소 관찰 기간에 걸친 연속 N회(기본 3회) 안정화 조회 +
> **매수가능금액·보유수량 delta 교차 확인** 후에만.

delta 교차 확인은 **사전 기준선이 있어야 성립한다.** baseline이 없으면 그 SHALL은
만족될 수 없다. 코드는 필드를 갖고 있고, 부르는 쪽이 안 채웠을 뿐이다.

### ~~무엇을 채우는가~~ — 왜 철회했는가

1판은 *"보호 청산의 `PlaceRequest`에 결정 시점의 `Baseline`을 싣는다 — 이미 관측 주기가
읽어 둔 값에서"*라고 썼다. **그 값이 없다.**

**차단 3 — 값의 원천이 소스에 없다.** `ExitObserverOptions`(`exitloop.go:166-223`)에
계정 상태를 읽는 필드는 **하나도 없다**. `Prices`는 시세이고 `Floor`는 대사 하한이다.
관측 주기는 매수가능금액도 보유수량도 읽지 않으므로 「이미 읽어 둔 값」이 존재하지 않는다.
채우려면 손절 경로에 **새 계정 호출**을 넣어야 하고, 그것은 같은 문서의 D6이
**§0.3을 근거로 금지한 바로 그것**이다.

**차단 2 — 억지로 채우면 살아 있는 매도를 은퇴시킨다.** `absenceCorroborated`
(`indoubt.go:486-500`)의 증거 모델은 **매수 예약 모델**이다:

```go
notional := intentNotional(intent)
bpDelta := buyingPowerNow - baseline.BuyingPower
if notional > 0 && bpDelta < 0 && math.Abs(bpDelta) >= notional*0.5 {
    return false, "... the buying power dropped ... consistent with this order having been accepted"
}
return true, "the holding and the buying power are unchanged from the pre-dispatch baseline"
```

**매도는 매수가능금액을 줄이지 않는다.** 접수된 매도의 매수가능금액 delta는 0이므로
위 `if`가 걸리지 않고 함수는 `true`("부재가 확증됨")를 **반환한다** — 그 매도가 브로커에
살아 있어도. 보호 청산은 전부 매도다. **부재 확증의 증거 모델은 매도에 대해 아무것도
증명하지 못하며, 그 상태에서 baseline을 공급하는 것은 「모름」을 「없음」으로 바꾸는
것이다.**

오늘의 「항상 park」는 그 무지의 **안전측 표현**이고, R4는 그것을 제거한다. §6(보수
방향만)에 걸린다.

**남기는 것**: 매도용 부재 증거 모델(예: 체결 이벤트 부재 + 목록 완주의 결합)이
별도 change의 **선행 조건**이다. baseline은 그 모델이 생긴 뒤에 공급한다.
`tasks.md` 5.3이 `issues.md` 기록을 요구한다.

## D5. 셋의 상호작용 (3판)

```text
[기동 시 1회] R1 소급 — 저장된 IN_DOUBT 중 확정 거절 code를 가진 것
      └─→ FAILED_CONFIRMED 로 재분류        ← 475150 · 080220 이 여기서 종결
                    │
                    ↓
[R3의 고리] attempt 가 비-CONFIRMED 종결 → pending_intent_id 로 exit_states 를 찾아
      Journal.ResolveExitProposal(ProposalRefused)
        B9 :846  pending_action / pending_level / pending_intent_id → NULL
        B12      LADDER면 active_rung 만 되돌림 (손절 가격 불변)
                    │
                    ↓
[다음 관측 5초 뒤] EvaluateLadder B26 :441 이 더는 억제하지 않는다
      → 발의 발생 → record :1117 게이트 열림
                    │
                    ↓
[R2] clearTheSymbol — detector 의 OPEN 스냅샷(≤3초)을 읽는다
        ├─ 반대 방향(BUY) → 취소          ← 막고 있던 앱 매수가 여기서 치워진다
        ├─ 자기 방향(SELL) → 오늘과 같이 withPending 일 때만
        ├─ PENDING_CANCEL → clear=false (연속 3회 카운터에서는 제외)
        └─ 스냅샷 부재·5초 초과 노후 → clear=false → record :1145 noteDelay
                    │
                    ↓
[R1] submit — 그래도 409 가 오면 code 로 확정 거절 → release(ProposalRefused)
      → 다음 주기 재무장 (영구 동결이 아니라 보이고 세어지는 반복)
                    │
                    ↓
                손절 제출 → 체결
```

**셋이 각각 다른 자리를 막는다:**

| | 없으면 |
| --- | --- |
| **R1 소급** | 얼어붙은 attempt가 종결되지 않아 R3의 고리가 걸릴 대상이 없다 |
| **R3의 고리** | attempt가 종결돼도 발의가 무장된 채라 사다리가 계속 억제한다 |
| **R2** | 발의가 나도 브로커의 매수가 그대로라 같은 409를 다시 받는다 |
| **R1 (전방)** | 그 409가 다시 IN_DOUBT가 되어 **같은 동결이 재발한다** |

## D6. 무엇을 하지 않는가

| | 결정 | 근거 |
| --- | --- | --- |
| `isDefinitiveRejection`에 409 추가 | **안 한다** | `request-in-progress`를 확정으로 만들어 살아 있는 주문을 은퇴시킨다 |
| `submit` B8이 즉시 `release`하게 고치기 | **안 한다** | B8의 논거는 옳다 — **종결 전에** 풀면 초과 매도다. 3판은 **종결 뒤에** 푼다(D3) |
| `checkSymbolFree` 미정산 루프에 위험 비증가 면제 | **안 한다** | spec의 SHALL이고 archive `2026-07-26-extend-execution-contract/design.md:63`이 이미 검토·폐기 |
| 재생 attestation 플래그 켜기 | **안 한다** | `[미측정 — 2b 전 비활성]` |
| `classifyRefusalBody`의 기존 message 매칭 제거 | **안 한다** | 지금 잡히는 것이 안 잡히는 방향(D0) |
| **손절 경로에 동기 브로커 조회 추가** | **안 한다 (3판에서 뒤집음)** | detector가 3초마다 같은 목록을 이미 읽는다(§0.4). 2판은 2초를 §0.3 비용으로 받아들였고 **그 비용이 불필요했다** |
| **자기 방향 미체결의 부재 확인** | **안 한다 (3판에서 철회)** | 초과 매도는 `armExitProposalTx` `:666`이 이미 막고, 남는 효과는 **보호의 영구 withhold**뿐 |
| `Journal.RecoverPending`을 세션 중에 호출 | **안 한다** | `RECORDED`를 종결시키고 `DISPATCH_STARTED`에 지어낸 사유를 쓴다(D3) |
| 브로커 조회 오류를 `record`로 반환 | **안 한다** | 브로커 두절 한 번이 판정 전체를 중단시킨다 |
| 「못 치우면 팔지 않는다」(B7)를 N회 후 뒤집기 | **안 한다** | 초과 매도 방향. 바꾸는 것은 침묵의 길이뿐 |
| **세션 중 IN_DOUBT 해소 루프** | **분리한다 (3판)** | 이 사건의 인과 경로에 없다. 주기·`Context.Resolver` 배선·§0.4 계측을 자기 change로 |
| `entry_price`·`initial_stop`·`baseline_price` 쓰기 | **안 한다** | 3판의 어느 경로도 손절 가격을 움직이지 않는다(§6) |

## D7. 실패 모드 재검토 (3판)

| 우려 | 답 |
| --- | --- |
| R1 소급이 실제로 실행된 주문을 "안 됐다"고 확정하면? | 이 code는 **주문 전 검증 거절**이다. 브로커가 반대 주문의 존재를 이유로 거절했고 그 상태에서 주문은 접수되지 않는다. openapi가 422(확정 거절군)에 둔 것이 같은 판단이다. **확정 거절 code가 없는 IN_DOUBT는 건드리지 않는다** |
| R1 소급이 우리가 쓴 산문을 매칭하면? | 저장된 `detail`의 **본문 부분**만 JSON으로 파싱한다. `"HTTP 409 does not prove…"`는 매칭 대상이 아니다 |
| 발의 해제가 초과 매도를 만들면? | 해제는 「두 번째 발의를 허용」이 아니라 「첫 발의를 끝냈다고 기록」이다. 다음 발의가 무장될 때 `armExitProposalTx` `:666`이 다시 검사한다 |
| 발의 해제가 손절 가격을 움직이면? | **움직이지 않는다.** `rollBackRungTx`는 `active_rung`만 쓴다(`exit_state.go:980-987`). AST 분기표로 고정한다(tasks 4.5) |
| 해제가 중복 호출되면? | **B8** `:842`가 멱등을 보장한다 — 이미 비었으면 무동작 반환 |
| 스냅샷이 오래됐는데 그것으로 판단하면? | 5초 상한을 넘으면 **「목록을 얻지 못했다」로 취급**해 `clear=false`다. detector 주기 3초의 여유 1회분이다 |
| R2가 사용자의 의도적 매수를 지우면? | 지운다. 그것이 이 change의 요구다 — 보호 청산이 우선한다. 감사에 남긴다 |
| 정상 취소가 정산 중인데 3회 카운터가 critical을 울리면? | `PENDING_CANCEL`은 카운터에서 제외한다. 판정은 `brokerstate.StateCancelPending`으로 한다 |
| a087이 먼저 오면? | **제약이 사라졌다.** 3판은 빈 가격을 **저널분에도** 0으로 읽는다 — 2판이 남겼던 충돌이 없다 |
| RATCHET 포지션도 같은 동결을 겪는가? | 그렇다. `EvaluateRatchet` **B17** `:423`이 LADDER B26과 같은 억제다. **해동 경로가 두 정책 모두에 적용된다** |
