# a165 — 주문 접수 안전 재시도와 취소·정정 409 재시도

> **정의만 — 착수 전.** 이 문서는 proposal 수준이다. design freeze·base commit·FLM/BTM·코드는 아직 없다.
> 실계좌 주문·운영 토글 flip 은 이 change 의 어떤 단계도 자동으로 하지 않는다.

## Why

`verify-execution-capability` task 1.3 에서 사용자가 retry matrix 수치를 확정했다(2026-10-22,
`openspec/changes/verify-execution-capability/issues.md` 「사용자 결정 (2026-10-22, 1.3 확정)」, 확정 표
`openspec/changes/verify-execution-capability/analysis/retry-matrix-slo-2026-10-22.md` §2). 그중 두 결정은 코드와 정본이
지금과 다르다.

1. **주문 접수 자동 재시도 — 최대 2회, 안전 방식.**
   (a) 브로커가 확실히 접수하지 않은 경우(HTTP 429, 요청을 보내기 **전**의 연결 실패)만 즉시 재시도 — 같은 `clientOrderId`,
   8s 예산, 최대 2회. (b) 결과 불명(타임아웃·5xx·전송 **후** 연결 끊김)은 맹목 재전송 금지 — `clientOrderId`/미체결 목록으로
   접수 여부를 먼저 조회하고 없을 때만 재전송. (c) 확정 거절(4xx 검증 오류)은 재시도 없음.
2. **취소·정정 `409 already-processing` 재시도 허용** — `retryAfter` 를 따르고 총 5s 이내, 최대 2회.

지금은 둘 다 0회다. 정본 retry matrix(`openspec/changes/archive/2026-07-26-harden-execution-base/retry-matrix.md`) §0.1 이
「주문 mutation 은 어떤 오류에도 자동 재시도하지 않는다」이고, §1 표가 POST place 의 429·5xx·전송중단을 IN_DOUBT 로 보낸다.
order-execution 요구 「Retry Matrix 산출물」과 「취소의 일시적 거절 처리」의 「접수·조건주문 생성은 재시도하지 않는다」 시나리오도
같은 말을 한다. 그래서 429 한 번이면 주문이 IN_DOUBT 로 가고, 검증 도구 밖(엔진 경로)의 취소는 409 한 번에 실패로 끝난다(M16 이
검증 도구에서 그 연쇄 차단을 보였다).

## What Changes

- 주문 접수(`POST /api/v1/orders`) 결과를 세 갈래로 나눠 처리한다: 미접수 확정 → 같은 키로 즉시 재시도(8s·최대 2회),
  결과 불명 → 접수 조회 후 부재일 때만 같은 키로 재전송, 확정 거절 → 재시도 없음.
- 엔진 경로의 취소·정정이 `409` + 본문 code `already-processing` 일 때만 `retryAfterSeconds` 를 따라 재시도한다(총 5s·최대 2회).
  판정은 HTTP 상태와 본문 code 가 둘 다 맞을 때만이다(현행 요구 「취소의 일시적 거절 처리」와 같은 판정).
- 재시도 사실·횟수·갈래를 attempt 기록에 남긴다 — 재시도로 성공한 요청을 첫 시도 성공으로 기록하지 않는다.
- 정본 retry matrix §0.1·§1 과 order-execution 두 요구를 **코드와 같은 커밋에서** 고친다(정본 표 머리말 규칙).
- **바뀌지 않는 것**: 재생(`execgw.Gateway.ReplayInDoubt`)은 task 2.7 결정대로 꺼진 채다 — 이 change 는 `Attested` 배선을
  켜지 않는다. 조건주문 생성·정정·취소의 자동 재시도는 0회 그대로다(결정 1 은 「주문 접수」만 말한다 — 범위 해석은 아래 열린 질문 1).
  손절·비상 청산의 즉시성을 늦추지 않는다(재시도는 원래 아무 주문도 없었을 자리에만 생긴다 — design 이 시간 상한으로 증명).

### 2.7(재생 비활성)과의 관계

결정 1 의 재전송은 최초 전송 직후 짧은 창 안에서만 같은 `clientOrderId` 로 일어난다. 안전 근거는 M14·M35 — 같은 키·같은 본문
재요청이 같은 `orderId` 를 돌려주고 두 번째 주문을 만들지 않았다(open delta 1), 본문이 다르면 `422 idempotency-key-conflict`.
그 측정은 첫 응답 직후의 재요청이었고 유효 창 경계(문서상 10분)는 사용자가 측정을 거부해 미검증이다. **8s 예산은 그 경계에서
멀다** — 10분의 1/75, TTL/2(5분)의 1/37.5, 현행 `DefaultReplayMargin` 60s(`internal/execgw/replay.go:78`)의 1/7.5.
「429 를 받은 POST 는 미접수」는 사용자 결정의 전제이지 측정이 아니다. 그 전제가 틀려도 같은 키 재전송은 두 번째 주문을 만들지
않는다는 것(M14·M35)이 안전망이다. 키의 계좌 스코프는 단일 계좌라 원리적 미검증이다.

## 지금 코드 — 접수가 재시도 없이 불리는 자리

근거: CodeGraph 1.6.0 `codegraph callers`(`analysis/code-context/codegraph-callers.txt`, 2026-10-22, HEAD `95160694e`) +
`grep -rn`(CodeGraph 가 놓친 자리 표시). 함수 내부 주장은 `tools/logic-map` AST(`analysis/ast/*.ast.json`, 같은 날 생성)의
구조 사실만 쓴다. 분기의 **의미**(어느 갈래가 어떤 오류를 받는지)는 아래에서 「미검증 — design 전 AST·FLM 필요」로 표시했다.

호출 사슬(생산):

| 자리 | 파일:줄 | 근거 |
|---|---|---|
| 엔진 보호 청산 제출 | `internal/app/engine/exitloop.go:1539` `o.opts.Submit.Place(...)`(메서드 `submit` :1483) | CodeGraph + grep |
| 엔진 진입 제출 | `internal/app/engine/tracer.go:442` `t.opts.Submit.Place(...)`(메서드 `submitEntry` :406) | CodeGraph + grep |
| 전략 진입 어댑터 | `internal/strategydispatch/adapters.go:112` `a.Gateway.Place(...)`(메서드 `PlaceStrategyEntry` :92) | CodeGraph + grep |
| flatten-all 매도 | `internal/flatten/liquidate.go:506` `s.Gateway.Place(...)`(메서드 `sell` :444) | CodeGraph + grep |
| 전략 게이트웨이 | `internal/execgw/strategy_gateway.go:64` `g.place(...)` | grep |
| 게이트웨이 진입 | `internal/execgw/gateway.go:358` `Gateway.Place` → :359 `g.place(ctx, req, nil)` | 소스 |
| 접수 클로저 | `internal/execgw/gateway.go:394–397` `plan.call` → :396 `g.trading.Place(...)` | AST `execgw--gateway.place` |
| 브로커 호출 1회 | `internal/execgw/gateway.go:708` `plan.call(tctx, fresh.ClientOrderID)` → :725 `classifyMutation(...)` | AST `execgw--gateway.submit` |
| 엔진 브로커 어댑터 | `internal/app/engine/broker.go:24` `b.off.PlaceOrder(ctx, intent)` | grep(CodeGraph 미검출) |
| 공식 클라이언트 | `internal/official/orders_write.go:182` `Client.PlaceOrder` → :192 `c.postWithHeaders` | AST `official--client.placeorder` |
| 전송 | `internal/official/client.go:427` `postWithHeaders` → :448 `c.send` → `send`(:320) → `doRequest`(:191) → :193 `c.hc.Do` | AST `official--client.{postwithheaders,send,dorequest}` |

취소 생산 호출자: `internal/app/engine/exitloop.go:1686` `o.opts.Submit.Cancel(...)`, `internal/flatten/flatten.go:382`
`s.Gateway.Cancel(...)`(grep). `Gateway.Amend` 의 생산 호출자는 grep 에서 찾지 못했다 — design 에서 CodeGraph 로 재확인.

범위 밖(바꾸지 않음): 검증 도구 `internal/verifylive/mutate.go:228`·`:270`·`:300` `r.broker.PlaceOrder`(이미 자기 409 재시도
`TransientRetryAttempts=2`·`TransientRetryMaxWait` 를 가진 운영자 도구), CLI `cmd/tossctl/order.go:229`
`app.tradingService.Place`(upstream 동작 — 토글 OFF = upstream 불변식).

AST 가 보여 주는 구조 사실:

- `Gateway.submit`(`gateway.go:466–762`) 은 분기 57(if 52·else 5)이고 **for/range/select 가 0, go 문 0** 이다. `plan.call` 은
  :708 한 자리다. 그 자리를 감싼 클로저 `call` 은 :712–713(`withStrategyEntryGateAuthority` 에 인자로 전달) 또는 :717 에서 쓰인다.
  `withStrategyEntryGateAuthority` 가 클로저를 몇 번 부르는지는 **미검증 — design 전 AST 필요**.
- `classifyMutation`(`internal/execgw/classify.go:21–99`) 은 분기 B1~B10(:22·:32·:44 switch·:45·:52·:66·:69·:84·:87·:89 else),
  반환 :23·:33·:46·:53·:74·:93·:98 이다. 429·5xx·전송 실패가 어느 반환으로 가는지는 **미검증 — design 전 FLM 필요**
  (정본 표 §1 은 IN_DOUBT 라고 적는다).
- `Client.send`(`client.go:320–366`) 에는 이미 **반복이 하나 있다**: B4 `for` :344, 그 안에서 `c.tm.refresh` :346 과 두 번째
  `c.doRequest` :354. 소스 주석(:309–319, :333–343)은 이것을 「401 이면 새 토큰으로 한 번(최대 두 번) 다시 보낸다」라고 설명한다 —
  즉 POST 접수도 401 에는 지금 재전송된다. 루프 조건의 의미는 **미검증 — design 전 FLM 필요**. a165 는 이 기존 재전송을 결정 1 의
  갈래표에 넣어 정합을 보여야 한다(401 은 「확실히 미접수」 갈래인가).
- 전송 단계 추적은 이미 있다: `journal.WithSendTracker`(`gateway.go:707`)와 `journal.SendState`(`internal/journal/dispatch.go:226`),
  주석상 `SendNotStarted` = 「요청 바이트를 하나도 쓰지 않음(connect/DNS/TLS 실패)」(:229). 결정 1(a)의 「전송 전 연결 실패」는
  이 상태와 같은 뜻으로 보이지만, `classifyMutation` 이 이 상태를 어떻게 쓰는지는 **미검증 — design 전 FLM 필요**.

## 위험도와 의무 (High-risk)

라이브 주문 제출·취소·정정과 retry matrix 는 WORKFLOW 「위험도 분류」의 High-risk 경로다. 비례 원칙의 면제 대상이 아니다.

- **FLM/BTM 필수**: 최소 `execgw.Gateway.submit`, `execgw.classifyMutation`, `execgw.Gateway.place`, `execgw.Gateway.Cancel`,
  `execgw.Gateway.Amend`, `official.Client.send`, 그리고 design 이 고르는 접수 조회 경로 함수. 분기 주장은 AST 열거로만 한다.
- **Pre-Edit 선언** 필수(High-risk 면제 없음).
- **적대적 리뷰**: 적대적 Eng 보이스 포함 다중 리뷰 — 특히 「중복 주문을 만드는 반례」(키 없는 재전송, 다른 본문 재전송,
  8s 창 밖 재전송, 조회-재전송 경합, 재시도 중 프로세스 재시작)를 레드팀으로 찾는다.
- **변이 원장**: 갈래 판정(미접수/불명/확정 거절)·같은 키 강제·8s 예산·최대 2회·409 본문 code 일치의 각 가드에 변이를 걸어
  CAUGHT 를 기록한다. 양성 대조군 포함.
- 실계좌 주문을 내는 자동 시험 금지 — httptest 만. 실계좌 확인이 필요하면 사람 승인 절차(verify 도구)로 따로 한다.

## 열린 질문 (design 전 사람 확인)

1. 결정 1 의 「주문 접수」에 조건주문 생성(`POST /api/v1/conditional-orders`)을 넣는가? 이 proposal 은 **넣지 않는다**고 읽었다.
2. (b) 「조회 후 없으면 재전송」을 최초 전송 8s 창 밖에서도 하는가? 이 proposal 은 **8s 창 안에서만** 하고, 창을 넘으면 지금처럼
   IN_DOUBT 해소 절차로 넘기는 안을 기본으로 둔다 — 창 밖 재전송은 2.7 이 끈 재생과 사실상 같은 동작이라서다.
3. 401 의 기존 재전송(`client.go:344`)을 결정 1 의 「확실히 미접수」 갈래로 인정하는가.

## Capabilities

### Modified Capabilities

- `order-execution`: 「Retry Matrix 산출물」(주문 mutation 재시도 금지의 예외 두 개)과 「취소의 일시적 거절 처리」(엔진 경로 적용,
  접수 재시도 시나리오 개정). spec delta 는 proposal 단계 초안(ADDED)만 있다 — design 단계에서 기존 요구의 MODIFIED 로 다시 쓴다.

## Impact

- 대상 코드(예상): `internal/execgw/`(gateway·classify·retry), `internal/official/`(전송 갈래 노출이 필요하면), 정본
  `openspec/changes/archive/2026-07-26-harden-execution-base/retry-matrix.md`.
- 설정·journal: attempt 기록에 재시도 횟수·갈래 필드가 필요할 수 있다 — 필요하면 additive-nullable 과 rollback 순서를 design 에 적는다.
- rate limit 예산(§0.4): 재시도는 요청을 늘린다 — a166(계좌 공유 3 req/s)과 같은 예산에 계상한다. 두 change 의 착지 순서는
  design 에서 정한다.
- 위험도: **High**. NOT READY.
