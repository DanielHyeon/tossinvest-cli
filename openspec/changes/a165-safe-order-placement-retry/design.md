# a165 design (2판) — 진입 접수 결과 불명 재전송 · 취소 409 처리 · 401 근거 고정

> **상태**: 2판 — **1.3 재리뷰 FAIL(F1·F2·F3, `review.md` 「1.3 재리뷰(2판)」)**. F1(flatten 소진 의미)·F5(엔진 세션 중 해소기 없음 — 아래 「해소 최소 45s」 문장은 엔진에선 거짓) 등은 3판에서 고친다. 사용자 2차 결정(issues.md 「사용자 결정 (2026-10-23 2차)」)을 반영해 범위를 줄였다. 1판과 1차 리뷰 처분은 `review.md`.
> base `34e12967e`. 함수 내부 주장은 `analysis/function-logic/*/ast.json` 의 AST 좌표로만 한다(23 번들).
> 실계좌 주문·`mutating: true` 명령·운영 토글 flip 은 없다. 시험은 httptest/fake 만.

## 0. 범위

| 남는 것 | 근거 결정 |
|---|---|
| **R1 진입 접수의 (b) 결과 불명 → 8s 창 안 조회 후 재전송**(같은 attempt·같은 키·같은 본문, 최대 2회) | 1.3 확정 행 + 2차 결정 4·5 |
| **R2 재전송 ACK 의 정착 전 목록 대조**(일치 정확히 1건·같은 id 일 때만 CONFIRMED) | 2차 결정 7(A1 수리) |
| **R3 취소 409 `already-processing` → 확정 미수락(FAILED_CONFIRMED)**, 엔진은 같은 주기 재시도 없음 | 2차 결정 1(D) |
| **R4 CLI `tossctl flatten` 취소의 409 재시도**(retryAfter, 누적 5s, 최대 2회) | 2차 결정 2 |
| **R5 기존 401 재전송을 「확실히 미접수」로 — 코드 불변, 시험 고정** | 사람 결정 3 |

**범위 밖**: 429 갈래(a166 뒤로 — 결정 5), 전송 전 실패 갈래(결정 4), 청산·flatten 매도 접수 재시도(결정 3 — 0회), 정정(Amend — 결정 6),
조건주문(결정 1차-1), 재생 `Attested`(2.7), CLI `tossctl order …`(execgw 밖, upstream 동작), strategy first-leg 경로(D9), journal 스키마(D8).

## D1. 재전송은 한 attempt 안에서 — `submit` 의 dispatch 클로저

AST `gateway.submit`: for/range/select/go 0, 브로커 호출은 `plan.call` 한 자리(708), `call` 은 B44(712)/B45(717) 중 한 갈래에서 한 번.
분류는 725 `classifyMutation(...)` 한 자리.

- 첫 전송까지의 코드(645–725)를 클로저 `sendOnce` 로 묶어 **첫 전송은 지금과 같은 순서·같은 거절 반환**으로 돈다(바이트 동일 회귀 시험).
- 정책(`Options.MutationRetry`)이 nil 이거나 첫 결과가 재시도 대상이 아니면 첫 결과를 그대로 반환 — **오늘 동작과 같다**(OFF 동치 시험).
- 재전송은 `sendOnce` 를 다시 부른다 — 마지막 경계 검사(결정 재읽기 645 · `checkDecision` · 키 일치 B32(654) · `checkProtection` · `checkEntry` ·
  `checkReservation`)를 **재전송마다 다시** 통과한다. 재전송 차례에 경계 검사가 거절하거나 창이 닫혔으면 재전송하지 않고 **직전 전송의 결과**로 정착한다
  (D4 가 강등). 첫 전송이 아니므로 `notSent` 로 바꾸지 않는다(1.3 A-통과 「전략 경로 함정」과 같은 이유 — 이미 보낸 전송이 있다).
- 새 attempt·새 결정·새 nonce 없음. 심볼 claim(B7)과 DISPATCH_STARTED 차단이 재전송 내내 유지된다.
- 재시작: 재전송 상태는 메모리뿐. 재전송 도중 죽으면 attempt 는 DISPATCH_STARTED → 재시작 복구가 IN_DOUBT 해소(조회 대조)로 보낸다. 재시작 뒤 재전송 없음
  (`RecoverPending`, 재생 `Attested: nil`).

## D2. (b) 결과 불명의 정확한 조건 — 이미 접수된 응답을 거른다(1.3 A2)

진입 재전송은 아래가 **모두** 참일 때만 고려한다.

1. 대상: `plan.kind == place` · `plan.raisesExposure`(매수) · `plan.strategy == nil` · 지정가(`LIMIT`) · 비분할 정수 수량. 감소(매도) place 는 0회(결정 3).
2. 오류가 둘 중 하나:
   - **실제로 받은 5xx**: `err == official.ErrServer`(동일성). AST `classifyStatus` B5(56)는 5xx 에 sentinel 그 자체(57)를 돌려주고,
     2xx envelope 해석 실패(`unwrapAndDecode` B2–B4 → 219·222·225)와 빈 토큰(token.go:164)은 **감싼** `ErrServer` 이다. 따라서 동일성은
     「5xx status 를 받았다」만 참이다. `trading.Service.withLineage` 는 오류를 감싸지 않는다(service.go:102–105). 누가 감싸게 되면 (b) 는 꺼지는 쪽(재전송 없음)으로 틀린다.
   - **응답 전에 끊긴 전송**: `errors.Is(err, official.ErrTransport)` 이고 그 전송의 `SendState ≥ Partial`(요청 바이트를 쓰기 시작) 이고
     **응답 첫 바이트를 받지 않았음**. 마지막은 a165 가 그 전송의 ctx 에 덧붙이는 httptrace `GotFirstResponseByte` 로 잰다 — `doRequest` B2(201)의
     본문 읽기 실패는 status 를 반환하지만 `send` 가 버리므로(`ErrTransport` 만 남음), 응답을 받은 뒤의 실패를 이것으로 가른다.
     401 재전송(`send` B4) 중 첫 응답 401 을 받았으면 「응답 받음」이므로 (b) 아님 → 오늘처럼 IN_DOUBT.
3. echo 불일치(`orders_write.go:200–204`, sentinel 없는 평문 오류)·`policyRefusal`·거절 code·ACK·그 밖의 status 는 (b) 아님.
4. `dctx.Err() != nil` 이면 아님(ctx 판정은 `dctx.Err()` 로 — `doRequest` 는 ctx 오류를 `%s` 로 평탄화한다, 1.3 A6).

## D3. 같은 키·같은 본문

- 재전송마다 `fresh := loadDecision` 재읽기와 B32 재검사. 호출은 `plan.call(tctx, fresh.ClientOrderID)` 한 자리.
- `Gateway.place` 의 `plan.call`(394–397)은 `keyed(key)` 사본으로 intent 를 만든다. 본문은 struct 마샬(map 없음), `formatDecimal` 결정적, 시각 필드 없음,
  confirm token 결정적(1.3 A-통과). 시험은 httptest 가 받은 바이트를 전송마다 비교하고 attempt 의 영속 wire body 와도 비교한다.

## D4. 최종 분류 — 낙관 종결 금지

- 진입 재전송이 1회 이상 있었으면(앞선 전송은 정의상 (b) — 닿았을 수 있음): 마지막 전송이 NOT_SENT·REJECTED 면 **AMBIGUOUS 로 강등**
  (예: 재전송 → `422 idempotency-key-conflict`·전송 전 실패·경계 거절). 마지막이 ACK 면 R2(D6).
- 재전송이 없었으면 첫 결과 그대로.
- 진입 (b) 의 첫 결과는 오늘도 AMBIGUOUS 이므로, 재전송은 **ACK(+R2 통과)로만 결과를 바꾼다**. 1판의 「전송 전 실패 → 재전송 → IN_DOUBT」
  비관 전이는 그 갈래를 빼서(결정 4) 없다.
- 취소(R3·R4): 앞선 전송이 `409 already-processing` 이면 그 전송은 브로커가 처리를 거절한 것이다(실행 안 됨). 강등 없음. 마지막 결과 그대로이되 마지막도
  409 already-processing 이면 R3 로 FAILED_CONFIRMED.

## D5. 창·횟수·대기

| 상수(`execgw`) | 값 | 쓰는 곳 |
|---|---|---|
| `DefaultMutationRetryWindow` | **8s** — 첫 전송의 `plan.call` 직전 시각부터 | 진입 재전송, 조회, R2 대조, 취소 재시도 |
| `DefaultEntryResends` | **2** | 진입 (b) |
| `DefaultCancelRetries` | **2** | flatten 취소 409 |
| `DefaultCancelMaxWait` | **5s 누적** | flatten 취소 409 |

- 창은 두 겹이다. ① 논리 판정은 주입 시계(`g.clk`)로 `since(first) < 8s`(정본 「시간 규율」, 시험 가능). ② 실행 상한은 첫 전송 직전에 만든
  `context.WithTimeout(dctx, 8s)`(`winCtx`) — Go 의 timer 는 단조 시계라 wall-clock 이 뛰어도 늘지 않는다(1.3 A6). 재전송의 경계 검사·`plan.call`·
  조회·R2 대조·취소 대기는 **전부 `winCtx` 아래**에서 돈다. 그래서 토큰 갱신을 포함한 재전송의 모든 HTTP 가 `first+8s` 에 끊긴다(1.3 A7).
  첫 전송은 `dctx` 그대로 — 오늘처럼 클라이언트 timeout 만 받는다.
- 창 검사 자리: `call` 클로저 안 `plan.call` **바로 앞**(1.3 A6) — 경계 검사가 오래 걸려도 창 밖 전송이 없다.
- 진입 (b) 는 대기 없음(조회 시간이 곧 간격). 취소 409 대기 = `error.data.retryAfterSeconds`(openapi 경로, 1.3 A9). 숫자가 아니거나 없거나 ≤0 이면 1s,
  `min(값, 5s − 누적, 창 남은 시간)`, 남은 것이 0 이하면 재시도하지 않는다. 대기는 `g.clk.Sleep(winCtx, d)`.
- **dispatch 시간 상한**: 첫 전송 소요 T1 과 `first+8s` 중 큰 값 — 재시도가 늘리는 시간은 실패한 경우에만, `8s − T1` 이하.

## D6. (b) 조회와 R2 대조

- 조회: `Options.OrderLists`(새 필드, `OrderPager`)로 **OPEN 1페이지 + CLOSED 1페이지**(심볼 필터, `Limit` 100, cursor 없음 — 1.3 확정 행, 결정 5).
  `HasNext` 는 보지 않는다(전체 walk 금지). 각 행은 `parseOrderFacts` 로 읽고, IN_DOUBT 해소의 `newMatcher`(stamp = 첫 전송 시각, 창 ±SubmitWindow 5분)로 대조.
  **수량 결손(0)은 「모름 = 일치」**(1.3 A8 — `matcher.matches` B3 는 결손을 불일치로 본다; a165 조회는 비교 전 결손을 매처 수량으로 채운다).
- 판정: 조회 실패·읽을 수 없는 행 → 재전송 없음(AMBIGUOUS). 일치 ≥1 → 재전송 없음(IN_DOUBT 해소가 그 주문을 찾는다). 일치 0 → 재전송.
  `OrderLists == nil` → 재전송 없음.
- **한 번의 부재 조회는 부재 증명이 아니다.** 목록 지연·1페이지 한계로 거짓 부재가 날 수 있다. 그때의 안전망은 같은 키(R11)와 R2 다.
- **R2 대조**(재전송 뒤 ACK 일 때만): 기존 round trip(`roundTripFor` B1 — `g.orders` 있으면 id 상세 확인)을 통과한 뒤, 같은 1페이지 조회로 일치 집합을
  orderId(바이트) 로 dedup 해 **정확히 1건이고 그 id 가 ACK id** 일 때만 통과. 0건(목록 지연)·2건 이상·다른 id·조회 실패 → verify 오류 →
  `DispatchVerified` B13(185) → MarkInDoubt(191) — IN_DOUBT. 자리: `DispatchVerified` 의 verify 훅을 감싼다(MarkAcked 181 뒤, Settle 199 앞).
  조회 ctx 는 `winCtx` — 늦게 온 ACK 는 대조할 시간이 없어 IN_DOUBT(보수).
- 잔여 R-a165-3(실패 응답 뒤 같은 키 멱등 동작 미측정)은 사용자가 수락했다(결정 7). R2 는 그 잔여가 현실화할 때 「추적 밖 주문」을 IN_DOUBT 로 드러낸다.

## D7. 손절·비상 청산 즉시성

| 경로 | a165 뒤 | 근거 |
|---|---|---|
| 엔진 청산 매도 접수(`exitloop.go:1539`) | **변화 없음** — 감소 place 는 재시도 대상 아님 | D2-1 |
| 엔진 청산 전 취소(`exitloop.go:1686`) | 409 already-processing 이 IN_DOUBT 대신 **FAILED_CONFIRMED**. 같은 주기 대기 0. AST `clearTheSymbol` B11(1697): 비확정이면 `cleared=false` 로 continue; 그 attempt 는 종결이라 다음 주기 B3(1639) `UnsettledOnSymbol` 에 걸리지 않고 다시 취소된다 — 오늘(IN_DOUBT → 해소 최소 45s)보다 **빨라짐** | 결정 1(D) |
| CLI `tossctl flatten` 취소(`flatten.go:382`) | 409 → 최대 2회·누적 5s 재시도, 소진 시 FAILED_CONFIRMED. AST `settleStep` B9·B10(395·400): 오늘은 IN_DOUBT → `resolveStep` **동기** 해소(최소 45s~최대 5분)가 뒤 종목을 막는다 → 실패 1건당 뒤 종목 지연이 「45s+」에서 「≤ min(5s + 2×(GET+POST), 8s−T1)」로 **줄어듦** | 결정 2 |
| CLI flatten 매도 접수 | 변화 없음 | D2-1 |
| 진입 접수 | 실패(5xx·응답 전 끊김) 때만 dispatch 가 ≤`8s−T1` 길어짐. 같은 종목 보유가 있으면 그동안 그 종목 청산이 `SymbolInFlight` 로 한 주기 밀릴 수 있음 — 오늘 같은 실패의 결과는 IN_DOUBT(해소 최소 45s 동안 같은 차단)이므로 차단 시간은 늘지 않는다(재전송이 성공·R2 통과면 줄어듦) | D4 |
| 수동 tracer(`tracer.go:283→297`) | 진입 실패 때 청산 관측 시작이 ≤8s 늦어짐 — 운영자가 실행하는 측정 도구, 엔진 상시 루프 아님 | 기록 |
| 계좌 공유 한도 | 진입 (b) 실패 1건당 최대 POST +2, 조회 GET +4(2회×OPEN·CLOSED), R2 GET +2. 429 갈래가 없어 패널티 창에서의 재전송은 없다 | D10 |

손절·비상 청산 경로의 지연을 **늘리는 변경은 없다** — 감소 place 는 그대로, 취소는 두 경로 모두 오늘보다 빨리 끝난다.

## D8. 기록

- 스키마 변경 없음. 재전송·재시도·분류 변경이 있었으면 정착 detail 끝에 꼬리표:
  `[a165 sends=<N> branch=<entry_ambiguous|cancel_already_processing> stop=<사유>]`. ReasonCode 는 바꾸지 않는다(어휘 고정 `AllReasonCodes`, `failclosed.go:254`) —
  R3 의 분류는 `broker_rejected` + 위 꼬리표.
- 전송 1회이고 분류 변경도 없으면 detail 은 오늘과 바이트 동일.
- 소비자: a129 청산 줄은 Detail 을 싣지 않는다(exitloop.go:1545 주석). flatten 은 detail 을 보고서 문장으로만 쓴다. 바이트 비교 소비자는 grep 0(3.3 에서 재확인).

## D9. strategy first-leg 경로는 범위 밖 — 보수

`PlaceClaimedStrategy`(plan.strategy ≠ nil)는 재전송하지 않는다. 이유: 그 경로는 EntryGate 잠금 안에서 브로커를 부르고(AST `withStrategyEntryGateAuthority`
63 defer Unlock, 80 fn 1회) 리스(SUBMITTING)·사전 전송 거절 정산(submit 473–497 defer)이 얽혀 있어 재전송 구조가 따로 필요하다. 오늘 생산에서 전략 레인은
DORMANT 이고, 재전송 없음은 오늘 동작이다. 전략 진입에 (b) 를 넓히는 것은 후속 change.

## D10. rate 예산과 a166

- 정상 상태 추가 0. 실패 1건당 최대: 진입 POST +2 · GET +6, flatten 취소 (GET 사전 확인 + POST) +2. 엔진 청산 전 취소의 FAILED_CONFIRMED 는 다음 주기 재취소를 낳는다 —
  브로커가 계속 already-processing 이면 5s 주기마다 GET+POST 1쌍(≈0.4 req/s, a166 3 req/s 상한 안).
- a166 과 파일 표면이 겹치지 않는다. 착지 순서 제약 없음(429 갈래는 a166 뒤 별도 결정).

## D11. 켜는 법·되돌리기

- `Options.MutationRetry *MutationRetry` — nil 이면 a165 동작 전부 없음(오늘과 바이트 동일). `Options.OrderLists OrderPager` — nil 이면 (b) 재전송 없음.
- 엔진 배선(`internal/app/engine/gateway.go`): `MutationRetry: execgw.EngineMutationRetry()`(진입 (b) 2회 + 취소 409 확정 미수락, 같은 주기 재시도 0), `OrderLists: orders`.
- CLI flatten 배선(`cmd/tossctl/flatten.go`): `MutationRetry: execgw.FlattenMutationRetry()`(취소 409 재시도 2회·누적 5s). 진입 place 가 없으므로 OrderLists 불필요.
- 되돌리기: 배선 nil → 코드 수정 + 이미지 재빌드(교체는 KST 05–09 창) — 런타임 끄기 스위치 없음(1.3 S12). 스키마·설정 파일 변경 없음.
- 전략 레인 `cycleDeadline` 30s 와의 여유: 전략 경로는 재전송이 없으므로(D9) 영향 없음.
- 정본 retry matrix §0.1·§1·§4 와 `execgw` 기본값은 GREEN 커밋에서 함께 바꾼다. 정본의 GET 429 규정(「Retry-After 상한 준수」)은 이 change 가 쓰지 않는다
  (429 갈래 없음, 1.3 S7 소멸).

## D12. 401 재전송(사람 결정 3) — 시험으로 고정, 코드 불변

AST `client.send` B4(344) `attempt < 2 && code == 401`, 루프 안 `tm.refresh` 뒤 `makeReq(tok)` 로 재생성(354), B8(358) `!adopted` 면 break;
`tokenManager.refresh` B1(119) 유효한 다른 캐시 토큰이면 채택(121) 아니면 교환(125). postWithHeaders 는 본문을 루프 밖에서 1회 마샬(427–448).

시험(official, httptest): ① 401 → 같은 본문 바이트로 재전송, 채택 경로 최대 2회 / 교환 경로 1회 ② 401 이 아닌 4xx·429·5xx 는 요청 1회
③ 게이트웨이: `send` 안의 401 재전송은 진입 재전송 횟수를 쓰지 않고, 401 응답을 받은 뒤의 끊김은 (b) 가 아니다(D2-2).
실측 근거 없음(R-a165-1)은 잔여로 남는다.

## D13. 시험 전략

- RED(httptest·fake, 실 엔드포인트 없음): 5xx → 조회 부재 → 같은 바이트 재전송 → ACK + R2 통과 → CONFIRMED / R2 0건·2건·다른 id → IN_DOUBT /
  조회 일치 → 재전송 없음 / 조회 실패 → 없음 / 응답 받은 뒤 끊김(본문 읽기 실패)·2xx 디코드 실패·echo 불일치 → 없음 / 창 경계(8s ±1ns, 가짜 시계) /
  횟수 경계(세 번째 재전송 없음) / 재전송 → 422 key-conflict·경계 거절 → AMBIGUOUS / 매도·strategy·MARKET·분할 → 없음 /
  취소 409 code 일치(엔진 정책) → FAILED_CONFIRMED 1회 전송 / flatten 정책 → 대기·재시도·누적 5s / code 불일치·두 자리 모순 → 오늘과 같음 /
  place 409 → 오늘과 같음 / amend 409 → 오늘과 같음 / `MutationRetry=nil` → 오늘과 바이트 동일 / 401 시험 ①②③.
- 변이 원장: D2 의 각 조건(동일성 → `errors.Is`, `!responded` 제거, `≥Partial` 제거, 대상 조건 각각), B32 재검사, 창 비교(`<`→`<=`, 시계 대신 ctx 만),
  횟수 경계, 조회 판정(일치 ≥1 무시, 실패 무시, CLOSED 생략), 수량 결손 처리, R2(정확히 1 → ≥1, id 비교 제거), D4 강등 제거, 409 code 판정(status 만),
  누적 대기 상한, 엔진/flatten 정책 구분. 양성 대조군(무변이 GREEN)과 「변이가 대상 줄에 닿았는가」 표식.
- 영향 범위: `internal/execgw`·`internal/official`·`internal/journal`(dispatch)·`internal/flatten`·`internal/app/engine`(배선·clear 경로)·`cmd/tossctl`(flatten 배선), 동시성 시험이 있는 execgw 는 `-race`.
