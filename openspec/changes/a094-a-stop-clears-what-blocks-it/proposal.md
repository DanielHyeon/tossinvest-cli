# a094 · 손절은 자기를 막는 것을 치운다

- **Feature**: `FEAT-TOS-009` — Exit line truth and position policy lifecycle
- **Story**: `STORY-TOS-a094`
- **Spec**: `order-execution`, `exit-policy`
- **위험 등급**: **High-risk** (손절 주문의 분류와 충돌 해소. §0.3 손절 즉시성 적용.)

> **작성 순서.** 이 문서의 분기 주장은 전부 `analysis/function-logic/`의 AST 산출물에서
> 나왔다. 산출물 **15개(분기 180개)**가 이 문서보다 **먼저** 만들어졌다
> (`.claude/CLAUDE.md`「단계 건너뛰기 금지」). 손으로 읽은 분기 주장은 이 문서에 없다.
> 3판이 6개(101분기)를 더했고, **그 6개가 D−1의 진짜 잠금과 D3의 해동 경로를 담는다.**

## Why

**2026-08-07, 475150(SK이터닉스). 손절이 발동했고, 발동한 채로 30분 넘게 멈춰 있었다.
그 사이 포지션은 무보호였다.**

> **정정(1라운드).** 이 문단의 초판은 *"가격은 −5.2%까지 갔고"*라고 쓰면서 아래
> 원장을 출처로 달았다. **그 수치는 원장에 없다** — 475150의 `exit_events` 9건의
> `observed_price` 범위는 57,700~59,000(최저가 entry 대비 −0.35%)이고, 동결 시점
> 이후로는 **관측 자체가 없다**(그것이 동결이다). 가격 시계열 테이블도 없다.
> −5.2%는 운영자가 앱에서 본 값이며 이 저장소의 측정이 아니다.

원장(`mutation_attempts`, `exit_states`, `intents`):

```text
09:04:46  편입          entry 57,900 · 손절 56,163(−3%) · R = 1,737
09:15:53  1단 승격       +1.9%(59,000)에서 손절선 56,163 → 57,900(본전)
09:17:46  발동          57,700 → STOP_LOSS_LADDER → SELL LIMIT 3 @ 57,700
09:17:46  브로커 거절     HTTP 409 {"code":"opposite-pending-order-exists",
                                  "message":"반대 포지션 미체결 주문이 존재합니다."}
이후       동결          state=IN_DOUBT · attempt_no=1 · replay_count=0 · settled_at=NULL
                        fill_events 0건 · 재시도 0회
```

같은 상태가 셋이다. **셋 다 `safety_class = RISK_REDUCING`** — 위험을 줄이는 주문이다.

| 종목 | 멈춘 시각 | 경과 |
| --- | --- | --- |
| 272210 | 2026-08-06T02:51:01Z | **22시간 이상** |
| 080220 | 2026-08-07T00:12:17Z | 35분 이상 |
| 475150 | 2026-08-07T00:17:46Z | 30분 이상 |

**반대편 미체결 매수 주문은 엔진이 낸 것이 아니다.** 세 종목의 `intents` 전수에
BUY는 한 건도 없다. 앱에서 직접 넣었거나 남아 있던 주문이다.

## 동결의 사슬 — AST가 열거한 분기

네 자리가 이어져서 영구 동결이 된다. 각 자리는 그 자체로는 옳은 판단이다.

### ① 409는 상태 코드만으로 분류되고, 409는 확정 목록에 없다

`journal.isDefinitiveRejection`(`dispatch.go:349-356`, 분기 3):

| 분기 | 자리 | 내용 |
| --- | --- | --- |
| **B2** | `:351` | `case 400, 401, 403, 404, 405, 415, 422:` → `true` |
| **B3** | `:353` | `default:` → `false` |

**409는 B2에 없으므로 B3으로 간다.** 그래서 `journal.ClassifyHTTPMutation`
(`dispatch.go:301-345`, 분기 7)에서 **B5** `:323` `case isDefinitiveRejection(statusCode)`가
성립하지 않고 **B7** `:337` `default:`가 잡는다 — `DispatchAmbiguous`,
`ReasonDispatchAmbiguous`, 그리고 원장에 남은 그 문장:

```text
HTTP 409 does not prove whether the mutation executed
```

**일반론으로는 참이다.** 409는 "같은 주문 키가 이미 처리 중"일 수도 있고 그때는 원본이
실행됐을 수 있다. 그러나 이 code에 대해서는 참이 아니고, **그 사실은 이미 정본에 있다.**

### 이 code는 이미 확정 거절로 취급되고 있었다 — 422에서

`docs/migration/openapi.latest.json`에서 이 code가 사는 자리:

```text
paths / /api/v1/orders / post / responses / 422 / content / application/json
      / examples / oppositePendingOrderExists / value / error / code
```

**openapi는 이것을 422로 선언한다.** 그리고 422는 `isDefinitiveRejection` **B2** `:351`의
목록에 **있다.** 즉 브로커가 문서대로 422를 냈다면 분류는 **처음부터 옳았다** —
확정 거절, 종결, 동결 없음.

`openspec/specs/order-execution/spec.md:44`도 이 code를 이름으로 인용하며 같은 전제를 쓴다:

> 심볼당 in-flight mutation 1개 제한은 **모든 safety class에** 유지된다(SHALL —
> 동시 반대 방향 제출은 브로커도 **`422 opposite-pending-order-exists`**로 거부한다(openapi))

**프로덕션은 409를 받았다.** 원장의 세 건, requestId 셋 모두 다르다:

```text
272210  official: API error 409: {"error":{"requestId":"6GKYatiUehps5SQX",
        "code":"opposite-pending-order-exists","message":"반대 포지션 미체결 주문이 존재합니다."}}
080220  … "requestId":"7d3we7ZD3dtxWTMO" … 같은 code
475150  … "requestId":"7k5oRgmEHnoU5Vfi" … 같은 code
```

**메시지 문구도 openapi와 다르다** — openapi는 *"동일 종목에 반대 방향의 체결 대기
주문이 있습니다"*, 실물은 *"반대 포지션 미체결 주문이 존재합니다"*. status도 message도
계약과 어긋났고, **일치한 필드는 `code` 하나다.**

**이 change는 새 정책이 아니라 적합성 수정이다.** spec과 코드는 이미 이 code를 확정
거절로 다루려 했다. 결함은 그 판단을 **status에 걸어 둔 것**이고, 브로커가 계약을
벗어난 status를 내자 확정 거절이 조용히 무기한 동결로 뒤집혔다.

### ② IN_DOUBT면 제안을 건 채로 반환하고, 해소를 resolver에 맡긴다

`engine.submit`(`exitloop.go:1237-1312`, 분기 11 · return 9):

| 분기 | 자리 | 내용 |
| --- | --- | --- |
| **B7** | `:1288` | `case out.State == journal.StateConfirmed:` |
| **B8** | `:1296` | `case out.State == journal.StateInDoubt \|\| ...UnresolvedInDoubt:` → `return nil` |
| **B9** | `:1301` | `case out.Reason == execgw.ReasonSymbolInFlight:` → `release(ProposalCancelled)` |
| **B10** | `:1304` | `default:` → `alertProposalRefused` + `release(ProposalRefused)` |

B8의 주석이 전제를 명시한다:

> *The proposal stays armed and **the resolver settles it**; releasing here would let
> the next observation submit a second sell on top of one that may already be live.*

**제안을 건 채로 두는 판단은 옳다.** 팔렸을지 모르는 주문 위에 두 번째 매도를 올리는
것이 더 나쁘다. 문제는 전제다.

### ③ resolver는 엔진이 도는 동안 한 번도 돌지 않는다

`Resolver.Resolve`에 닿는 경로는 둘뿐이다 — `flatten.go:428`(CLI 전량청산)과
`reconcile/recovery.go:253`. 후자를 담은 `reconcile.Run`(`recovery.go:207-296`, 분기 12)은
**B3** `:230` `for _, rec := range pending`에서 미정산 attempt를 훑고 **B4** `:231`이
IN_DOUBT를 골라 `:253`에서 해소한다. 그 `Run`을 부르는 자리는 `cmd/tossctl/engine.go:374`
**하나이고, 엔진 시작 시 1회**다. **실행 중 주기 해소 루프는 존재하지 않는다.**

따라서 ②의 전제 *"the resolver settles it"*은 **프로덕션 배선에서 거짓이다.**
세션 중 IN_DOUBT가 된 attempt는 다음 재시작까지 아무도 보지 않는다.

### ④ 미정산 attempt는 그 종목의 모든 주문을 막는다 — 취소까지

`execgw.checkSymbolFree`(`gateway.go:799-834`, 분기 9 · return 8):

| 분기 | 자리 | 내용 |
| --- | --- | --- |
| **B2** | `:804` | `for _, rec := range pending` — 미정산 attempt 루프 |
| **B4** | `:809` | `if same` → `reject(ReasonSymbolInFlight)` · **면제 없음** |
| **B5** | `:815` | `if !plan.raisesExposure { return nil, nil }` — **위험 비증가는 여기서 면제된다** |
| **B7·B9** | `:822`·`:827` | UNRESOLVED_IN_DOUBT 루프 — B5를 통과한 것만 온다 |

**설계는 이미 원칙을 갖고 있고, 그 원칙이 절반에만 적용돼 있다.** B5 위의 주석이 원칙을
쓴다: *"An UNRESOLVED_IN_DOUBT attempt blocks only new exposure … the operator must
still be able to cancel and liquidate through the engine (§0.3)."* 그런데 그 면제는
**B5 아래**에 있고, B2–B4의 미정산 루프에는 없다. 그래서 얼어붙은 attempt 하나가
같은 종목의 **손절·익절·취소까지 전부** 막는다.

이것이 두 가지 동결 모양을 만든다. **475150**은 자기 attempt가 B8에서 얼어 제안이 걸린
채 조용히 멈춘다. **272210**은 얼어붙은 옛 attempt 때문에 새 제안이 매번 B9로 가서
`STOP_LOSS_LADDER → PROPOSAL_CANCELLED`를 **5초 주기**로 반복한다 — **라이브락**이다.
실측: `PROPOSAL_CANCELLED` 1931건이 약 2시간 54분에 걸쳐 있고 그 주 action은
`STOP_LOSS_LADDER` 1606건(83%)·`LADDER_PARTIAL` 326건(17%)이다. inter-arrival 중앙값 5.0초.
**attempt가 22시간 미정산인 것과 루프가 22시간 돌았다는 것은 다르다** — 후자는 거짓이다
(1라운드 정정).
모양은 둘, 결과는 하나: **손절이 나가지 않는다.**

### ⑤ 그리고 무장된 발의가 그 상태를 영구화한다 (3판이 더한 고리)

**①~④는 attempt가 어떻게 얼어붙는지를 설명하고, 그것으로 충분하지 않다.**
2라운드가 잡았다 — `submit` **B8** `:1296`이 `release`를 부르지 않으므로
`exit_states.pending_action`이 무장된 채 남고, 그러면
`EvaluateLadder` **B26** `:441`(RATCHET은 `EvaluateRatchet` **B17** `:423`)이
손절 조건 성립에도 **빈 전이**를 돌려준다. 빈 발의는 `record` `:1082`의
`orderable`을 거짓으로 만들고 `:1117`의 게이트가 열리지 않는다.

**즉 ①~④를 전부 고쳐도 475150·080220은 녹지 않는다** — R1이 고치는 `submit`도
R2가 고치는 `clearTheSymbol`도 **도달하지 않는 코드**이기 때문이다.
`design.md` D−1이 소스 사슬과 원장 실측으로 그것을 적고, **3판의 R3과 R1 소급이
그 고리를 푼다.**

## 이미 있는 장치가 왜 작동하지 않았나

엔진에는 충돌 주문을 치우는 장치가 **이미 있다.**

`engine.record`(`exitloop.go:1077-1197`, 분기 14) **B3** `:1117`
`if orderable && (snapshot.CancelPendingFirst || isFullExit(proposal))`이 청소 경로를 열고,
`:1141`이 `clearTheSymbol(ctx, m, snapshot.CancelPendingFirst)`를 부른다.
`CancelPendingFirst`는 `exitpolicy/ladder.go:447`에서
`in.State.PendingAction != ActionNone`으로 정해진다 — 475150은 `pending_action`이
`STOP_LOSS_LADDER`이므로 **참이다. 경로는 도달 가능하다.**

`engine.clearTheSymbol`(`exitloop.go:1334-1392`, 분기 9) **B3** `:1343`
`if !buy && !withPending { continue }` — **매수 주문을 치우는 것이 이 함수의 명시적 일이다.**

그런데 그것이 훑는 `live`는 `Journal.LiveOrdersForSymbol`
(`fills.go:1849-1927`, 분기 7)이 돌려준다. 그 SQL은

```sql
FROM mutation_attempts a JOIN intents i ON i.id = a.intent_id
WHERE a.state = ? AND a.kind IN ('PLACE','AMEND') AND a.broker_order_id <> ''
```

— **저널이다. 브로커가 아니다.** 앱에서 직접 넣은 매수 주문은 intent도 attempt도 없으므로
이 질의에 **한 행도 나오지 않는다.** 엔진은 자기가 낸 주문만 치울 수 있다.

**이것이 사용자가 지목한 설계 결함이다.** 청소 장치는 있고, 경로도 열렸고, 매수를 치우는
분기까지 있는데, **치워야 할 그 주문이 보이지 않는다.**

## What Changes

**세 가지를 바꾼다.** (1판의 R4는 1라운드에서 철회했다 — 아래.)

> **3판이 잠금을 다시 지목했다.** 1·2판은 미정산 attempt를 동결의 원인으로 봤고, 2라운드가
> 그것이 틀렸음을 보였다 — **진짜 잠금은 무장된 채 남은 발의(`exit_states.pending_action`)**다.
> `design.md` D−1이 소스 사슬과 원장 실측으로 그것을 적는다. 아래 R1·R2·R3은 그 정정 뒤의
> 모습이다.

### R1 — 브로커가 요청 자체를 거절한 409는 확정 거절이다

`execgw.classifyMutation`(`classify.go:21-79`, 분기 7)의 **B3** `:46`
`if reason, refused := ClassifyBrokerRefusal(err); refused`는 **B5** `:64`의 상태 코드
분류보다 **먼저** 돈다. 그 안의 `classifyRefusalBody`(`failclosed.go:223-238`)는
브로커 자신의 어휘를 좁게 매칭하는 switch이고, `trade_auth_required`·`fx_consent`·
`funding_required`가 이미 그렇게 분류된다.

`opposite-pending-order-exists`를 **code로** 판단한다. 상태 코드 표
(`isDefinitiveRejection`)는 **건드리지 않는다** — 409 전체를 확정으로 바꾸면
`request-in-progress`(원본이 실행됐을 수 있다)까지 확정으로 만들어 정확히 반대 방향의
위험을 만든다. **code별로 좁게** 판단한다.

**단, 기존 switch에 case를 더하는 방식은 쓰지 않는다**(1라운드 정정). 그 switch는 본문
통짜에 대한 `strings.Contains`이고, 이 change의 spec은 *"message 문구로 걸어서도 안
된다(SHALL NOT)"*를 쓴다 — **한 줄로는 자기 SHALL NOT을 만족할 수 없다.** 그래서
본문의 `code` 필드를 읽는 함수를 하나 더한다. 실측된 본문은 자리가 두 가지다:
최상위 `{"code": "TRADE_AUTH_REQUIRED"}`(기존 fixture)와
`{"error":{"code":"opposite-pending-order-exists"}}`(프로덕션 409 3건). **둘 다 읽는다.**

효과: attempt가 `DispatchRejected`로 **종료**한다. `PendingAttempts`에서 빠지므로
④의 전면 차단이 풀리고, `submit`은 B10 `:1304`로 가서 알림을 올리고 레벨을 재무장한다.

### R2 — 청소는 브로커의 미체결 주문을 본다

`clearTheSymbol`이 훑는 목록을 **저널 ∪ 브로커 미체결**로 넓힌다. 엔진이 낸 주문은
지금처럼 lineage로 다루고, 저널에 없는 주문은 브로커가 보고한 그대로 취소한다.

경계는 좁게 둔다. **보호 청산이 자기 길을 치울 때만**, **같은 종목**, **반대 방향**,
그리고 감사 흔적을 남긴다. 반대 방향 취소는 새 권한이 아니다 — `clearTheSymbol`
B3 `:1343`이 이미 하고 있고, 이 change는 그 눈을 완성한다.

**1판이 빠뜨린 것 넷을 여기서 채운다**(1라운드 차단 4~8):

- **배선은 공짜가 아니다.** `OrderPager`는 `ExitObserverOptions`에 없다. 필드·nil 정책·
  `Context.ExitObserver`의 채움이 새로 필요하고, **파서도 새로 필요하다** —
  `brokerstate.ParseOfficialOrder`는 `Side`·`Symbol`·`Price`를 주지 않고
  `official.Client.Orders()`는 첫 페이지만 준다.
- **브로커 오류는 판정을 멈추지 않는다.** 조회 실패·타임아웃은 `clearTheSymbol` 안에서
  흡수해 「치우지 못했다」로 떨어뜨린다. `record`로 반환하면 브로커 두절 한 번이
  워터마크와 기준선까지 멈춘다.
- **지연은 0이 아니다.** 이 경로는 손절 제출 직전에 돌므로 **2초 · 최대 3페이지**로
  유계로 만들고, 그 값과 근거를 적는다(`precheckTimeout` 선례).
- **자기 방향 매도의 부재도 확인한다.** R1이 제안을 해제한 다음 주기에는 청소가
  `withPending=false`로 불려 **살아 있는 매도를 보고도 지나쳤다.** 그 구멍을 막는다.

효과: **R1이 동결을 풀고, R2가 다음 주기의 손절을 실제로 내보낸다.**
**사용자가 매수 주문을 취소하지 않아도 손절·익절이 나간다.**

### R3 — IN_DOUBT는 엔진이 도는 동안 해소된다

③을 고친다. 세션 중 IN_DOUBT가 된 attempt를 유계 시간 안에 resolver에 넘긴다.
이것은 409에 국한되지 않는다 — *어떤* 모호한 결과든 지금은 다음 재시작까지 그 종목을
동결시킨다. **`Resolve`는 mutator를 갖지 않는다**(`indoubt.go:190` — "settles IN_DOUBT
attempts by observation alone"). 관측만 하므로 주문 side effect를 만들지 않는다.

**진입점은 새로 만든다**(1라운드 차단 4). 1판은 `Context.Recovery`/`reconcile.Run`
재사용을 썼는데, 그 경로의 첫 동작인 `Journal.RecoverPending`은 `RECORDED`를
*"found at startup with no dispatch recorded"*로 **종결**시키고 `DISPATCH_STARTED`에
*"process stopped after dispatch started"*라는 **지어낸 사유**를 쓴다. 세션 중에 그것은
원장 위조다. 새 진입점은 `Journal.PendingAttempts`(읽기)와 `Resolver.Resolve`(관측)만
쓰고 **`IN_DOUBT`인 것만** 고른다.

**이득이 무엇인지 정확히 적는다**(1라운드 정정): 얼어붙은 세 건의 산출은
`FAILED_CONFIRMED`가 **아니다.** baseline이 없어 부재가 증명되지 않으므로 결론은
park(`UNRESOLVED_IN_DOUBT`)다. **그것이 이득이다** — `IN_DOUBT`는 그 종목의 **모든**
mutation을 막지만(취소까지) `UNRESOLVED_IN_DOUBT`는 **신규 노출만** 막는다
(`gateway.go:790-798`). 즉 R3은 확정을 주는 게 아니라 **그 종목의 취소·매도를 푼다.**

대가도 적는다: park는 `Gate.Block`을 **종목 인자 없이** 부르므로 **계정 전역 신규 진입**이
차단된다. 오늘은 재시작 때만 생기던 상태가 세션 중에도 생긴다. 막히는 것은 매수뿐이고
손절·익절·취소는 오히려 풀린다.

### R4 — **철회했다** (1라운드 차단 2·3)

1판의 R4는 보호 청산의 `PlaceRequest`에 계정 기준선을 싣는 것이었다. **철회한다.**
근거 둘 다 소스에서 확인됐다.

**(1) 값 원천이 없다.** `ExitObserverOptions`(`internal/app/engine/exitloop.go:166-223`)
전 22개 필드에 보유수량·매수가능금액 원천이 **0개**다. `journal.Position`에는 통화도 없다.
`m.position.Quantity`는 **원장의 믿음**이고 `absenceCorroborated`는 그것을 브로커
`Holdings` 합계와 뺀다 — **두 수는 같은 것을 재지 않는다**(엔진 관리 3주 + 앱 보유 7주면
delta 7이 나와 거짓 사유로 park한다).

**(2) 억지로 채우면 살아 있는 매도를 은퇴시킨다.** `absenceCorroborated`
(`internal/execgw/indoubt.go:445-500`)의 증거는 둘뿐이고 주석이 모델을 자백한다 —
*"a reservation of roughly this order's notional"*. **매수 예약 모델이다.**
접수됐으나 미체결인 **매도**는 보유수량도 매수가능금액도 안 바꾸므로 두 검사가 모두
통과해 `FAILED_CONFIRMED`가 되고, 다음 주기가 **살아 있는 매도 위에 두 번째 매도**를 낸다.
`BuyingPower = 0`으로 채우면 `bpDelta > 0`이라 가드가 **영원히 발화하지 않아** spec §3의
교차 확인 절반이 침묵으로 만족된다.

**오늘의 「항상 park」가 안전측이었다.** 그것을 제거하는 변경은 §6(보수 방향만)에 걸린다.
1판은 R4를 **순수 이득**으로 서술했고 그것이 틀렸다.

**대신 spec에 SHALL NOT으로 남긴다** — 매도용 부재 증거 모델이 정의되기 전에는 기준선을
공급하지 않는다. 그 모델(체결 이벤트 부재 + 목록 완주의 결합 등)은 별도 change의 선행
조건이며, `issues.md`가 그것을 기록한다.

## Why all three, and why not fewer

- **R1만**: attempt는 종료하지만 다음 주기가 같은 409를 다시 받는다. 동결이 *보이는
  루프*로 바뀔 뿐 **여전히 팔리지 않는다.**
- **R2만**: 얼어붙은 attempt가 ④로 모든 주문을 막으므로 청소 주문 자체가 나가지 못한다.
  **R1 없이는 R2가 실행되지 않는다.**
- **R3만**: 해소가 돌아도 이 사건의 원인(409 오분류)이 남아 매 주기 다시 언다.

**R1·R2가 이 사건을 끝내고, R3이 다음 사건의 동결을 막는다.**
R3의 산출은 대개 `FAILED_CONFIRMED`가 아니라 **park**이며, 그 park가 곧
`PendingAttempts`에서 빠져 **그 심볼의 취소·매도를 푸는 것**이 실제 이득이다
(1라운드 정정 — 1판은 이 이득을 틀리게 주장했다).

## 이미 얼어붙은 것들은 어떻게 되나 — **3판에서 답이 바뀌었다**

**1·2판의 이 절은 틀렸다.** *"배포는 재시작을 포함하므로 현재의 동결 자체는 배포 시점에
풀린다"*고 썼는데, **재시작은 이것을 풀지 못한다.**

### 왜 풀리지 않는가

동결에 **두 모양**이 있고 1·2판은 하나만 봤다:

| 종목 | `pending_action` | 무엇이 잠갔나 | 1·2판이 |
| --- | --- | --- | --- |
| **475150** | `STOP_LOSS_LADDER` | **무장된 발의** | 못 본다 |
| **080220** | `STOP_LOSS_LADDER` | **무장된 발의** | 못 본다 |
| **272210** | `None` | 미정산 attempt | 본다 |

무장된 발의가 있으면 `EvaluateLadder`가 손절 조건 성립에도 **빈 전이**를 돌려주고
(`ladder.go:441-443`), `record`의 게이트가 열리지 않으므로(`exitloop.go:1082`·`:1117`)
**R1이 고치는 `submit`도 R2가 고치는 `clearTheSymbol`도 도달하지 않는다.**

그리고 재시작은 attempt만 만진다 — `RecoverPending`(`journal/recovery.go:86-125`)도
`reconcile.Run`(`internal/reconcile/recovery.go:207`)도 `exit_states`를 쓰지 않는다.

### 3판이 무엇을 하는가

```text
배포 + 재시작
   │
   ├─ [R1 소급] 저장된 IN_DOUBT 중 본문에 확정 거절 code가 있는 것
   │            → FAILED_CONFIRMED      ← 475150 · 080220 의 attempt 가 종결
   │                    │
   ├─ [R3의 고리] 종결된 attempt 의 intent 를 가리키는 발의를 해제
   │                    │                 pending_action → NULL
   │                    ↓
   ├─ 다음 관측(5초) — 사다리가 더는 억제하지 않는다 → 손절 발의
   │                    │
   ├─ [R2] detector 스냅샷에서 막고 있던 앱 매수를 보고 취소
   │                    │
   └─ [R1] 그래도 409 면 code 로 확정 거절 → 해제 → 재무장 (반복이 보인다)
                        │
                        ↓
                   손절 제출 → 체결
```

**따라서 3판은 475150·080220을 배포로 녹인다.** 1·2판은 그러지 못했다.

### 그래도 배포 전에는 사람이 처리한다

**배포 시점까지 475150(32주)·080220(12주)은 무보호다.** 그리고 010170(30주)은
`exit_states` 자체가 없어 이 change의 범위 밖이다(a095가 그것을 **보이게만** 한다).

`make gate`와 독립 리뷰가 끝나기 전에는 배포하지 않는다. 그동안의 처리는 사람이 한다.


## Out of scope — 침묵하지 않고 적는다

- **④의 미정산 루프에 위험 비증가 면제를 주는 것**은 검토했으나 넣지 않는다.
  `gateway.go:791-794`의 주석이 반대 근거를 갖는다 — 미정산 mutation이 둘이면 어느
  브로커 주문이 어느 attempt의 것인지 말할 수 없고, 엉뚱한 것을 겨눈 취소가 늦은 취소보다
  나쁘다. 그 논증은 유효하다. R1이 이 사건에서 attempt를 **종료**시키므로 면제가 필요 없다.
- **a087(보호 청산은 시장가)와 겹치지 않는다.** 시장가로 바꿔도 409
  `opposite-pending-order-exists`는 그대로 난다 — 가격 문제가 아니라 주문 충돌이다.
- **a089·a091·a092와 겹치지 않는다.** 각각 계측·알림 등급·알림 체류이고, 어느 것도
  409 분류나 충돌 해소를 다루지 않는다.

## Impact

| | 자리 | 성격 |
| --- | --- | --- |
| R1 | `internal/execgw/failclosed.go`(신설 `classifyRefusalCode`) · `reason.go` · `testdata/reason_codes.golden` | **새 함수 하나** + reason code 하나 + golden 재생성. 기존 switch는 그대로 |
| R2 | `internal/app/engine/exitloop.go` `clearTheSymbol`·`ExitObserverOptions` · `exitwiring.go` · `cmd/tossctl/engine.go` · `internal/filldetect`(`Snapshot`에 2필드) | 목록 원천을 **detector의 기존 스냅샷**으로 넓힌다. **새 브로커 호출 0건**, 새 파서 없음 (3판 정정 — 2판의 동기 조회·`ParseWorkingOrder`는 폐기) |
| R3 | attempt 종결 후처리 1곳 → `Journal.ResolveExitProposal` | **종결된 attempt가 무장한 발의를 푼다.** 새 루프도 주기도 브로커 호출도 없다 (3판 교체 — 「세션 중 해소」는 별도 change) |
| R1 소급 | 기동 경로 1곳 | 저장된 IN_DOUBT를 **같은 code 증거로** 재분류. **475150·080220을 녹이는 부분** |
| ~~R4~~ | — | **철회** |

spec: `order-execution`(code 필드 분류 · 재생 경계 · 세션 중 해소 진입점 · 매도 부재
증거 모델), `exit-policy`(치우기의 범위 · 조회 예산 · 자기 방향 부재 확인 · 무한 보류 금지)

**기존 함수 내부를 고치므로 Function Logic Map 면제는 없다.** 산출물 9개는 이미 있고,
구현 후 재생성한다.
