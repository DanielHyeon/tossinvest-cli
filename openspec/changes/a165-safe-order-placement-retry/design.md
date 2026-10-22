# a165 design — 주문 접수 안전 재시도 · 취소·정정 409 재시도

> **상태: freeze 거절(1.3 FAIL, blocker 3) — D7 사람 결정 + D13 수리 필요.** 1.3 리뷰 기록은 `review.md`. §2(구현)는 시작하지 않았다.
> base `34e12967e`(`base-commit.txt`). 함수 내부 주장은 전부 `analysis/function-logic/*/ast.json` 의 AST 좌표로만 한다.
> 실계좌 주문·`mutating: true` 명령·운영 토글 flip 은 이 change 의 어떤 단계도 하지 않는다. 시험은 httptest/fake 만.

## 0. 입력과 범위

- 사용자 결정: 1.3 확정표(`verify-execution-capability/analysis/retry-matrix-slo-2026-10-22.md` §2 ✱ 두 행), proposal 「사람 결정
  (2026-10-23)」 1~4.
- 대상: 엔진 게이트웨이(`execgw.Gateway`)를 지나는 `POST /api/v1/orders`(접수), `…/cancel`, `…/modify`.
- **바뀌지 않는 것**
  - 조건주문 생성·정정·취소: 재시도 0회(사람 결정 1). 조건주문은 `Gateway.submit` 을 지나지 않는다(plan.kind 는 Place/Cancel/Amend 뿐 —
    `gateway.go` 의 `journal.KindPlace|KindCancel|KindAmend` 세 계획).
  - 재생(`ReplayInDoubt`): `Attested: nil` 그대로(2.7). a165 는 그 스위치를 건드리지 않는다.
  - CLI `tossctl order place|cancel|amend`: execgw 를 지나지 않는다(upstream 동작 = 토글 OFF 불변식).
  - 검증 도구 `verifylive` 의 자기 409 재시도: 그대로.
  - journal 스키마: 바꾸지 않는다(D8).

## D1. 재시도는 한 attempt 안의 여러 전송이다 — `submit` 의 dispatch 클로저 안

**사실(AST `internal-execgw--gateway.submit`)**: 함수 전체에 for/range/select/go 가 0 이다. 브로커 호출은 클로저 `call` 안의 `plan.call`
한 자리(708)이고, `call` 은 B44(`withStrategyEntryGateAuthority` 의 fn, 712) 또는 B45(717) 중 한 갈래에서 한 번 불린다. 분류는 725 의
`classifyMutation(tracker.State(), result, callErr)` 한 자리다. 즉 지금은 attempt 당 브로커 호출이 정확히 1회다.

**결정**: 재전송 루프를 dispatch 클로저 안(636–726)에 넣는다. 새 attempt·새 결정·새 nonce 를 만들지 않는다 — journal 행은 그대로 하나이고
DISPATCH_STARTED 뒤 「결과가 정해질 때까지」 같은 attempt 가 여러 번 보낸다.

- 근거: 새 attempt 를 만들면 결정·nonce·심볼 claim 을 다시 얻어야 하고, 그 사이에 같은 심볼 다른 mutation 이 끼어들 수 있다. 같은 attempt
  안이면 심볼 claim(B7)과 journal 의 DISPATCH_STARTED 차단이 재시도 내내 유지된다.
- **전송마다 마지막 경계 검사를 다시 통과한다.** 지금 첫 전송 직전에 하는 검사 묶음(645–676: `loadDecision` 재읽기 · `checkDecision` ·
  키 일치 B32 · `checkProtection` · `checkEntry` · `checkReservation`, strategy 면 `finalAuthorityCheck` · `checkReservation` ·
  `requireTransport` · `withStrategyEntryGateAuthority`)를 함수로 묶어 **재전송 직전마다** 다시 부른다. 하나라도 거절하면 재시도를
  멈추고 **직전 전송의 분류**로 정착한다(D4 가 강등) — 거절을 `notSent` 로 바꾸지 않는다. 이미 보낸 전송이 있기 때문이다.
- **대기는 EntryGate 잠금 밖.** AST `internal-execgw--gateway.withstrategyentrygateauthority`: 63 에서 `defer g.entry.mu.Unlock()`, fn 은 80 에서 1회.
  대기를 fn 안에 두면 잠금이 대기만큼 잡힌다(EntryGate.Block 을 부르는 조회 경로가 멈춘다). 그래서 전송마다 `withStrategyEntryGateAuthority`
  를 새로 통과하고, 대기·조회는 그 바깥에서 한다. 게이트 revision 이 바뀌었으면(B4) 재전송하지 않는다.
- 재시작: 재시도 상태는 메모리에만 있다. 재시도 도중 프로세스가 죽으면 attempt 는 DISPATCH_STARTED 로 남고, 재시작 복구는 지금처럼
  IN_DOUBT 해소(조회 대조)로 간다. **재시작 뒤 재전송은 없다**(재생 `Attested: nil`).

## D2. 갈래표 — `classifyMutation` 반환 지점 위에 얹는다

`classifyMutation` 자체는 바꾸지 않는다. 재시도 판정은 **오류 값**(`statusOf` · `errors.Is` · 본문 code)과 그 전송의 `SendState` 로 하는
순수 함수 `mutationRetryBranch(kind, send, err)` 하나다(문구를 읽지 않는다).

| 관측(AST 근거) | 현재 분류(반환) | a165 갈래 | 동작 |
|---|---|---|---|
| `err == nil` (B1 → 23) | ACKED | — | 끝 |
| `policyRefusal` (B2 → 33) | NOT_SENT | 확정(로컬) | 재시도 없음 — 결정적 로컬 거절 |
| 거절 code 목록 (B4 → 46) | REJECTED | (c) 확정 거절 | 재시도 없음 |
| code 모순 (B5 → 53) | AMBIGUOUS | 불명이나 (b) 아님 | 재시도 없음(타임아웃·5xx·전송 끊김이 아님) |
| 사람 요구 거절 (B6 → 74, B7 이면 AMBIGUOUS) | REJECTED/AMBIGUOUS | (c)/불명 | 재시도 없음 |
| status 429 (B8: `statusOf` B4) → `ClassifyHTTPMutation` B7 → 93 | AMBIGUOUS | **(a) 미접수 확정(사용자 전제)** | place 만 재시도 |
| status 5xx (B8: `statusOf` B5) → B7 → 93 | AMBIGUOUS | **(b) 결과 불명** | place 만, 조회 후(D6) |
| status 409 + 본문 code `already-processing` (B8 → B7 → 93) | AMBIGUOUS | **409 일시 거절** | cancel·amend 만 재시도(D5) |
| 그 밖 status(409 place·408 등) (B8 → 93) | AMBIGUOUS/REJECTED | — | 재시도 없음 |
| 상태 없는 전송 오류, `SendNotStarted` (98 → `ClassifyHTTPMutation` B2 → 304) | NOT_SENT | **(a) 전송 전 실패** | place 만, `errors.Is(err, official.ErrTransport)` 일 때만 |
| 상태 없는 전송 오류, Partial/Complete (98 → B1 → 311) | AMBIGUOUS | **(b) 전송 후 끊김·타임아웃** | place 만, 조회 후(D6) |
| `context.Canceled`/`DeadlineExceeded` | NOT_SENT/AMBIGUOUS | — | 재시도 없음(ctx 가 끝났다) |
| 401/403 최종(`Client.send` 루프 소진 뒤, B8 → 93) | REJECTED | — | 재시도 없음(아래 D9) |

- (a)·(b) 는 **place 에만** 적용한다. cancel·amend 에는 멱등키가 없다 — 409 갈래만 갖는다(사람 결정 4, 정본 요구 「취소의 일시적 거절 처리」).
- 「요청을 보내기 전의 연결 실패」는 `journal.SendNotStarted`(요청 바이트를 하나도 쓰지 않음)와 `official.ErrTransport` 가 **둘 다** 맞을 때만이다.
  ErrTransport 중 본문 marshal·요청 생성 실패도 NotStarted 로 잡히는데, 그것은 결정적이라 재시도해도 같은 결과다(무해, 최대 2회).

## D3. 같은 키·같은 본문 강제

- 키: 재전송마다 `fresh := loadDecision` 을 다시 읽고 B32(`fresh.ClientOrderID != prep.ClientOrderID` → 거절)를 다시 통과한다. 재전송은
  `plan.call(tctx, fresh.ClientOrderID)` 한 자리로만 한다.
- 본문: `Gateway.place` 의 `plan.call` 클로저(394–397)는 호출마다 `keyed(key)`(호출자 intent 의 사본 + 키)로 PlaceIntent 를 만들므로 같은 키면 같은
  intent → `buildOrderCreate` → `json.Marshal` 결과가 같다. 시험은 httptest 서버가 받은 **바이트**를 전송마다 비교하고, attempt 에 영속된
  wire body(`PlaceWireBody`)와도 같은지 본다.
- 이 강제가 깨지면 R11 의 안전망(같은 키·같은 본문은 두 번째 주문을 만들지 않음, 다른 본문이면 `422 idempotency-key-conflict`)이 사라진다.

## D4. 최종 분류 — 재시도가 판정을 낙관으로 바꾸지 못한다

전송 중 하나라도 **브로커에 닿았을 수 있으면**(그 전송의 SendState ≥ Partial, 또는 status 를 받음) 그 attempt 는 「확실히 안 보냄」이 아니다.

- 마지막 전송이 ACKED → ACKED(같은 키 → 같은 주문; 기존 round trip 이 id 를 확인).
- 마지막 전송이 NOT_SENT 또는 REJECTED 이고 앞선 전송 중 닿았을 수 있는 것이 있으면 → **AMBIGUOUS 로 강등**. 예: 429 → 재전송 → `422
  idempotency-key-conflict` 는 원 주문에 대해 아무것도 증명하지 않는다(정본 「IN_DOUBT 해소」 — FAILED_CONFIRMED 금지). 429 → 재전송 →
  전송 전 실패도 같다.
- 앞선 전송이 전부 SendNotStarted 였으면 마지막 전송의 분류 그대로.
- 결과: 재시도는 오늘의 첫 전송 분류보다 **낙관적인 종결**(NOT_DISPATCHED·FAILED_CONFIRMED)을 만들 수 없다(429 단독도 오늘 AMBIGUOUS).
  **단 비관 쪽 전이는 새로 생긴다**(1.3 S3): 첫 전송이 전송 전 실패(오늘 NOT_SENT, 심볼 즉시 풀림)인데 재전송이 전송 후 끊기면 IN_DOUBT 가 된다.
  또 재전송 ACK 를 그대로 CONFIRMED 로 닫는 것은 1.3 A1(BLOCKER) — 첫 전송이 따로 만든 주문이 남을 수 있다 — 에 걸려 있다(D13).
  사용자 전제 「429 는 미접수」는 **재시도 허가**에만 쓰고 **종결 근거**로는 쓰지 않는다.

## D5. 시간·횟수 상한

| 상수(execgw 기본값) | 값 | 근거 |
|---|---|---|
| `DefaultMutationRetry().Window` | **8s** — 최초 전송 시작부터. 재전송 **시작**은 `now < first+8s` 일 때만, 재전송·조회의 ctx 마감도 `first+8s` | 사용자 결정 1(8s 예산), 사람 결정 2(창 밖 재전송 금지) |
| `MaxRetries` | **2** — place 의 (a)+(b) 합산, cancel·amend 의 409 합산. 전송 총 ≤3 | 사용자 결정 1·2 |
| (a) 대기 | `DefaultRetryPolicy().BackoffFor(n, 400ms)` — 400ms·800ms ±25%, 남은 창으로 자름 | 정본 표 §2 의 조회 backoff 재사용. R1(300ms 간격 3회 전부 429)은 짧은 재시도의 회복 확률이 낮음을 보인다 — 대기를 늘리면 8s 창을 다 쓰므로 그대로 둔다 |
| 409 대기 | `data.retryAfterSeconds`(없거나 ≤0 이면 1s), **누적 대기 ≤5s**, 남은 창으로 자름 | 사용자 결정 2, R9(1s·1회로 수락), `verifylive.TransientRetryMaxWait` |
| (b) 대기 | 없음 — 조회(D6) 시간이 곧 간격 | — |

- **dispatch 전체 시간 상한**: `max(T1, 8s)` (T1 = 첫 전송 소요, `official` 클라이언트 타임아웃 15s 이하). 재시도가 늘리는 시간은 실패한
  경우에만, 그리고 `8s − T1` 이하다.
- 시각은 주입 시계(`g.clk`)로 잰다(정본 「시간 규율」). 시험은 가짜 시계로 창 경계 ±1ns 를 고정한다.

## D6. (b) 결과 불명 — 조회 후 부재일 때만 재전송

1. 창 안이고 재시도가 남았으면 `OrderLists`(새 `Options` 필드, `OrderPager`)로 **OPEN·CLOSED 를 심볼 필터로 한 번씩** 끝 페이지까지 걷고,
   IN_DOUBT 해소와 **같은 매처**(`newMatcher` — 심볼·방향·수량·가격, 제출 창은 첫 전송 시각 ±`SubmitWindow`)로 대조한다. 조회 ctx 마감 = 창 끝.
2. 조회 실패·시간 초과·읽을 수 없는 레코드 → 재전송하지 않는다(AMBIGUOUS → IN_DOUBT, 오늘과 같다).
3. 일치 ≥1 → 재전송하지 않는다(접수됐을 수 있다 — IN_DOUBT 해소가 그 주문을 찾아 확정한다).
4. 일치 0 → 같은 키·같은 본문으로 재전송(창·횟수 재확인 뒤).
5. `OrderLists == nil` → (b) 재전송 없음(fail-closed). (CLI flatten 배선 여부는 D7 ③ 사람 결정.)

**한 번의 부재 조회는 부재 증명이 아니다**(정본 「IN_DOUBT 해소」 3: 연속 N회·최소 관찰 기간·잔고 교차). 목록은 늦게 반영될 수 있다. 그래서 이 재전송의
안전은 조회가 아니라 **같은 키**(R11)가 진다 — 첫 전송이 실은 접수됐다면 같은 키 재전송은 같은 orderId 를 돌려주거나 `409 request-in-progress` 로 끝나고,
둘 다 두 번째 주문이 아니다(후자는 D2 표의 「그 밖 status」 → 재시도 없음 → IN_DOUBT). 조회는 사람 결정이 요구한 추가 신중함이다.
조회 응답은 멱등키를 싣지 않으므로(openapi `Order` 스키마에 `clientOrderId` 없음) 키로 매칭하지 않는다.

## D7. 손절·flatten 즉시성 — **사람 결정 필요(이 change 를 §2 앞에서 멈춘 이유)**

> 1.3 리뷰 S1 이 첫 판의 사실 전제(「flatten 도 엔진 경로」)를 반증해 이 절을 다시 썼다. 아래 경로별 주장 중 함수 내부 분기에 기대는 것
> (`ObserveOnce` 의 `quoteUsable` continue, `clearTheSymbol`, `Saga` 루프, `resolveStep`/`resolveCancel`)은 **아직 AST 번들이 없다** — 1.3 S8/A10,
> 사람 결정 뒤 freeze 전에 `tools/logic-map` 번들로 확인한다(아래 「미검증」 표지).

**경로별 사실**

| 경로 | 오늘 실패(429·409) 때 | a165 재시도를 붙이면 | 근거 |
|---|---|---|---|
| ① 엔진 청산 루프 — 청산 전 미체결 취소(`exitloop.go:1686`) | 409 → IN_DOUBT, 그 자리에서 다음 주문으로 넘어감 — 뒤 포지션 지연 0, 그 종목 손절은 백그라운드 해소(최소 45s)까지 막힘 | 그 종목: 보통 ~1s 뒤 손절(R9). 뒤 포지션: 실패 1건당 최대 `min(누적 5s + 2×(GET 사전 확인+POST), 8s−T1)` 지연, ~2건 쌓이면 시세 lease 15s(`retry.go:192`)를 넘어 다음 주기(5s, `exitloop.go:98`)로 밀림 | 소스(미검증 — AST 전) |
| ② 엔진 청산 루프 — 청산 매도 접수(`exitloop.go:1539`) | 429 → IN_DOUBT, 뒤 지연 0, 그 종목 막힘 | 그 종목: 429 재시도는 R1 상 회복 확률 낮음. 뒤 포지션: 실패 1건당 ≤`8s−T1` 지연 | 소스(미검증) |
| ③ CLI `tossctl flatten`(`cmd/tossctl/flatten.go:233–247`, 유일한 `flatten.Saga` 생성 자리) | 취소 409 → IN_DOUBT → 같은 순차 루프 안에서 `resolveStep` 이 **동기로** 해소(최소 45s, 최대 5분) — 뒤 종목이 그만큼 늦음 | D11 첫 판은 이 배선을 제외했다. 붙이면 409 취소는 흔한 경우 오히려 **빨라짐**(~1s), 최악 +≤5s. 매도 접수 429 재시도는 ②와 같음 | 소스(미검증) |
| ④ 진입 접수 | 오늘 NOT_SENT(전송 전 실패)면 심볼 즉시 풀림 | 전송 전 실패 → 재전송 → 전송 후 끊김이면 IN_DOUBT 로 **비관 전이**(D4 정정) → 같은 종목 보유가 있으면 청산 치움이 `UnsettledOnSymbol`(`exitloop.go:1636–1643`)에 걸려 45s+ 막힘. 재시도 중 심볼 claim 이 ≤8s 더 잡힘. 429 재전송·(b) 조회가 계좌 공유 한도를 더 써 패널티를 늘릴 수 있음([미측정], 확정표 §3) | 소스 |
| ⑤ 수동 tracer(`tracer.go:283→297`) | — | 진입 재시도가 같은 goroutine 의 청산 관측을 ≤8s 늦춤 | 소스 |

어느 범위도 안전 불변식 3 을 한 방향으로 증명하지 못한다 — 「실패한 그 종목」은 빨라지고 「뒤 종목·같은 종목 보유」는 늦어질 수 있다. 그래서 Manager 지시
(「늘리는 방향이면 멈추고 보고」)대로 §2 앞에서 멈춘다. 사람에게 물을 질문은 `review.md` 「사람 결정 요청」.

**첫 판 선택지(참고 — 질문은 review.md 가 정본)**: A 전부 / B 진입 place + 취소·정정 409(청산 place 0회) / C 진입 place 만. B 도 ①에서 뒤 포지션 지연을
늘리고, C 도 ④로 0 이 아니다.

## D8. 기록 — 재시도로 성공한 것을 첫 시도 성공으로 적지 않는다

- journal 스키마는 바꾸지 않는다. attempt 의 정착 `detail` 끝에 기계 판독용 꼬리표를 붙인다:
  `[a165 retry sends=<N> branches=<rate_limited|pre_send|presence_absent|already_processing,…> stop=<why>]`. ReasonCode 는 바꾸지 않는다
  (reason code 어휘는 `AllReasonCodes`(`failclosed.go:254`)가 시험으로 고정하고 Phase 2 원장이 읽는다 — 새 code 는 데이터 어휘 변경이다).
- 꼬리표는 전송이 2회 이상일 때만 붙는다. 1회면 detail 은 오늘과 바이트 동일(OFF 동치 시험).
- 근거: 스키마 변경(additive-nullable + rollback)보다 작고, 정본 요구가 요구하는 것(사실·횟수·갈래가 attempt 기록에 남음)을 채운다.

## D9. 기존 401 재전송(사람 결정 3) — 바꾸지 않고 시험으로 고정

AST `internal-official--client.send`: B4 `for attempt := 0; attempt < 2 && code == http.StatusUnauthorized; attempt++`(344), 루프 안에서
`c.tm.refresh` 뒤 `makeReq(tok)` 로 요청을 **다시 만들어** `doRequest`(354), B8 `if !adopted { break }`(358). postWithHeaders 는 본문을 루프 밖에서
한 번 `json.Marshal` 하고 makeReq 가 그 바이트로 reader 를 만든다(client.go 427–448).

고정할 시험(official 패키지, httptest):
1. 401 → 같은 본문 바이트(같은 clientOrderId)로 재전송, 재전송은 최대 2회(채택 토큰 경로) / 1회(교환 경로).
2. 401 이 아닌 4xx·429·5xx 는 루프를 돌지 않는다(서버가 받은 요청 수 = 1).
3. 게이트웨이 쪽: `Client.send` 안의 401 재전송은 D5 의 `MaxRetries` 를 소비하지 않는다(한 번의 `plan.call` 안의 일) — 그리고 그 401 재전송 뒤 최종
   429 면 D2 의 (a) 갈래로 간다.

**401 을 받은 POST 가 실제로 미접수라는 실측은 없다** — 잔여(issues.md). D4 가 그 전제를 종결 근거로 쓰지 않으므로, 전제가 틀려도 같은 키 재전송이
두 번째 주문을 만들지 않는다(R11)는 안전망만 남는다.

## D10. rate limit 예산(§0.4)과 a166

- 최악 추가 요청(attempt 1건): place = POST +2, (b) 조회 = OPEN 1페이지 + CLOSED ≤`MaxPages` 페이지(심볼 필터, 창 마감으로 잘림).
  cancel = (GET 사전 확인 + POST) × 2.
- 정상 상태 추가 0 — 재시도는 실패한 경우에만 생긴다.
- a166(계좌 공유 3 req/s)이 먼저 착지하면 재전송도 그 한도를 지난다. 두 change 는 파일 표면이 겹치지 않는다(a166 은 공유 limiter, a165 는 submit).
  착지 순서 제약 없음 — 단 a165 단독 착지 동안 429 재전송이 패널티를 늘릴 수 있는지는 [미측정](D7 의 간접 영향 ②).

## D11. 켜는 방법과 되돌리기

- `execgw.Options.MutationRetry *MutationRetryPolicy` — nil 이면 재시도 0회(오늘 동작과 바이트 동일, OFF 동치 시험).
  엔진 배선(`internal/app/engine/gateway.go`)이 `execgw.DefaultMutationRetry()` 와 `OrderLists: orders` 를 넣는다. CLI flatten 배선은 넣지 않는다.
- 되돌리기: 배선 한 줄(nil)로 오늘 동작. 스키마·설정 파일 변경 없음.
- 정본 retry matrix §0.1·§1·§4 와 `execgw` 기본값은 GREEN 커밋에서 함께 바꾼다(정본 머리말 규칙).

## D12. 시험 전략(§2·§3 — D7 결정 뒤)

- RED(httptest, 실 엔드포인트 없음): 429→성공(같은 바이트 2회), 전송 전 실패→성공, 5xx·전송 후 끊김→조회 부재→재전송, 조회 일치→재전송 없음,
  조회 실패→재전송 없음, 창 경계(8s ±1ns)·횟수 경계(3번째 재시도 없음), 4xx 확정 거절→1회, 409 code 일치→대기·재시도·누적 5s 상한,
  409 code 불일치·모순→재시도 없음, place 409→재시도 없음, 재전송 직전 게이트 거절→멈추고 강등(D4), 429→422 key-conflict→AMBIGUOUS,
  strategy 경로 대기 중 EntryGate 잠금 미보유, `MutationRetry=nil` 이면 오늘과 바이트 동일, D7 선택지에 따른 class 구분.
- 변이 원장: 갈래 판정 각 줄, 같은 키(B32 재검사 제거·`fresh` 대신 빈 키), 8s 비교 연산자, `MaxRetries` 경계, 409 code 일치(status 만 보기),
  D4 강등 제거, 잠금 밖 대기. 양성 대조군(무변이 GREEN) 필수.
- 영향 범위: `internal/execgw`·`internal/official`·`internal/journal`(dispatch 시험)·`internal/app/engine`(배선)·`internal/flatten`, 동시성 시험이
  있는 것은 `-race`.

## D13. 1.3 리뷰가 연 항목(미해소 — 사람 결정 뒤 이 design 을 고쳐 재리뷰)

| id | 심각도 | 내용 | 필요한 설계 변경 |
|---|---|---|---|
| A1 | BLOCKER | (b)·429 뒤 재전송이 ACK 를 받으면 CONFIRMED 로 닫히는데, 첫 요청이 브로커 안에서 따로 주문을 만들었고 멱등 기록이 실패 경로에서 남지 않았다면 그 주문은 추적되지 않는다. 「실패 응답 뒤 같은 키의 멱등 동작」은 미측정(R11 은 2xx 직후 재요청만 쟀다) | 잔여로 이름 붙여 사람 수락. 앞선 전송이 닿았을 수 있으면 재전송 ACK 뒤 정착 전에 목록 대조 — 일치가 정확히 하나이고 그것이 acked id 일 때만 CONFIRMED, 아니면 IN_DOUBT |
| S1 | BLOCKER | D7 첫 판의 사실 전제 오류(flatten 은 CLI 뿐, 경로별 방향이 다름) | 이번 판 D7 에서 정정. 내부 분기 주장은 AST 번들로 확인 필요 |
| S2 | BLOCKER | D7 사람 결정 미결 | review.md 「사람 결정 요청」 |
| A2 | MAJOR | `ErrServer` 에는 2xx 의 envelope 디코드 실패가, 상태 없는 `ErrTransport` 에는 2xx 의 본문 읽기 실패가, 평문 오류에는 clientOrderId echo 불일치가 섞여 (b) 로 간다 — 이미 접수된 주문이 재전송된다 | (b) 를 「실제 받은 status ≥500」 또는 「응답 status 없는 전송 오류 + Partial/Complete」로만 한정. 필요하면 official 쪽에 status 를 실은 타입 오류 추가 |
| A3/S6 | MAJOR | 429 재전송·(b) 조회가 계좌 공유 패널티를 늘려 손절을 간접 지연할 수 있음. (b) CLOSED walk(MaxPages 50)는 1.3 확정 행 「CLOSED 1페이지, 전체 walk 금지」와 충돌 | 조회는 OPEN·CLOSED 각 1페이지. a166 을 먼저 착지시키거나 그 전까지 429 갈래 끄기 — 사람 결정 |
| A4/S4 | MAJOR | 남긴 시나리오 「재생과 재시도의 구분」(키 없는 mutation 의 모호 오류 → 재전송 없음)이 예외 2(409 취소 재전송)와 모순 | 시나리오에 예외 2 제외를 적음 |
| S5 | MAJOR | 기존 「IN_DOUBT 해소」(정체 모른 채 재주문 금지, 재생은 attestation 전제)와 (b) 재전송의 관계가 delta 에 없음 | 「IN_DOUBT 해소」도 MODIFIED — DISPATCH_STARTED 8s 창 안의 같은 키 재전송은 해소가 아니며 attestation 전제의 예외임을 명시 |
| S7 | MAJOR | 유지된 SHALL 「429 는 Retry-After 상한 준수」를 D5 가 안 다룸. 「즉시」 대신 backoff 는 결정 이탈(보수 방향) | 헤더가 있고 남은 창보다 길면 재전송 안 함, 헤더 존재 [미측정], 이탈 명시 |
| S8/A10 | MAJOR(절차) | 조회 경로(`scanBoth`·`ScanOrders`·`newMatcher`·`parseOrderFacts`), `unwrapAndDecode`, `tokenManager.refresh`, `refusalCodes`, `trading.Service.Cancel`, D7 의 청산·flatten 함수들에 FLM 번들이 없음 — A2 결함이 정확히 이 빈칸에서 나옴 | 번들 생성 후 재주장 |
| A5 | MINOR | `Gateway.Amend` 엔진 생산 호출자 0. 409 재전송은 체결 전 스냅숏 수량을 다시 보냄 | Amend 를 범위에서 빼거나 재전송 전 주문 재읽기 |
| A6 | MINOR | 창 검사 자리(재검사 묶음이 POST 를 8s 뒤로 밀 수 있음, EntryGate `refresh()` 에 ctx 없음), `g.clk` 은 wall-clock(`time.Now().UTC()`), ctx 판정은 `doRequest` 가 `%s` 로 평탄화 | 창 검사는 `plan.call` 바로 앞, 경과는 단조 시계, ctx 판정은 `dctx.Err()` |
| A7 | MINOR | 「T1 ≤ 15s」 거짓 — 401 루프는 `doRequest` 최대 3회, 토큰 교환 hc 는 timeout 없는 `http.DefaultClient` 가능(token.go:46–47) | 근거 문장 정정, 「재전송의 모든 HTTP 는 first+8s ctx」를 불변식·시험으로 |
| A8 | MINOR | 매처: quantity 결손이 불일치로 판정됨(주석의 「모르면 일치 쪽」과 반대) | a165 조회에서는 결손을 일치로, (b) 는 비분할 LIMIT 에만 |
| A9 | MINOR | 실제 경로는 `error.data.retryAfterSeconds`(openapi) | 경로 고정, 비숫자·과대값 상한, code 판정은 `refusalCodes` 한 곳 |
| S10 | MINOR | 허가형 「할 수 있다(SHALL)」는 시험 불가, 시나리오 누락(조회 부재→재전송, 전송 전 실패→재전송, nil 동치, 게이트 거절→강등) | 「한다」 또는 MAY 로, 시나리오 추가 |
| A11 | NOTE | HTTP/2 transport 가 REFUSED_STREAM·GOAWAY 에서 같은 바이트를 내부 재전송할 수 있음 — 「attempt 당 1회」는 plan.call 기준 | 잔여에 적음 |
| S12 | NOTE | 되돌리기는 코드 수정 + 이미지 재빌드(교체는 KST 05–09 창), 런타임 끄기 스위치 없음. 전략 레인 `cycleDeadline` 30s 대비 dispatch 상한 여유 명시 | D11 에 반영 |
