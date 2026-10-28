# a165 design (3판) — 취소 409 `already-processing` 처리 · 401 재전송 근거 고정

> **상태**: 3판. 사용자 결정 3차(2026-10-23, F3 = 1)로 진입 재전송(R1·R2)을 a171 로 옮겼다. 1판·2판과 리뷰 처분은 `review.md`.
> 1.3 3차 리뷰(2026-10-28) **PASS** — 처분으로 D10 에 배선 시험·`definitiveRefusalCodes` 불변 두 항목을 더했다(본문 결정 변경 없음). 사람 결정 H1: D3 의 최악 추가 지연은 그대로 둔다.
> **3판 보완(1.3 2차 독립 리뷰 GPT-5.6 Terra FAIL 처분, 2026-10-28)**: 사람 결정 P2-2(작은 바깥 래퍼)를 D3·D5 에 일관되게 반영(`sendOnce` 추출 폐기),
> P2-3(R3 뒤 그 종목 진입은 계속 막음)을 **진입 전용 latch**(D11)로 정의, D10 배선 시험·변이 원장 보강, 8s 창 시작점을 dispatch 진입 직전으로(P3).
> **재리뷰 처분(1.3 재리뷰 GPT-5.6 Terra FAIL P1 1·P2 1, 2026-10-28)**: D11 해제 r2 에 원장 순번 경계(latch 정착 transition 뒤에 **기록이 시작된** 같은 대상 주문의 CONFIRMED 취소만)와
> 새 질의 함수(기존 `ConfirmedCancelOf` 불변), RED (f)~(h)·변이 ⑦~⑨, 운영 절 D12(조회·journalled 해제·경보). 새 evidence AST `journal--attempt.transition`·`journal--journal.prepare`(HEAD `d64620457`). schema 36 그대로.
> **3차 재리뷰 처분(1.3 3차 재리뷰 GPT-5.6 Terra FAIL P1 1·P2 1, 2026-10-28)**: D11 r2 의 비교 축에서 **거래일을 뺐다** — 대상 주문의 정규 범위(계좌·시장·종목·`target_order_id`)를 L 에서 읽어
> 비교하고, 번호 재사용은 거래일 대신 **주인 유일성**(그 범위에서 그 번호의 접수 확정 주인 intent 가 정확히 하나)으로 막는다. 순번 조건 `r_C > s_L` 은 그대로. RED (h) 갱신·(i)~(k) 추가, 변이 ⑩ 교체·⑪·⑫ 추가.
> D12 질의 2 를 r2 와 같은 조인으로. 새 evidence AST `execgw--gateway.preparerequest`(HEAD `5e1ae6260`). schema 36 그대로.
> 새 AST: `analysis/function-logic/internal-execgw--gateway.checksymbolfree/`(편집 대상 번들, 24번째) · `analysis/evidence/` 의 `execgw--gateway.checkentry` ·
> `journal--journal.confirmedcancelof` · `execgw--entrygate.blocksymbol` · `journal--journal.recoverpending`(HEAD `e4724910b`).
> base `34e12967e`. 함수 내부 주장은 `analysis/function-logic/*/ast.json` 의 AST 좌표로만 한다(23 번들).
> 실계좌 주문·`mutating: true` 명령·운영 토글 flip 은 없다. 시험은 httptest/fake 만.

## 0. 범위

| | 내용 | 근거 결정 |
|---|---|---|
| **R3** | 엔진 게이트웨이의 **취소**가 `409` + 본문 code `already-processing` → **FAILED_CONFIRMED(확정 미수락)**. 같은 dispatch 안 재시도 0, 대기 0. 대상 주문이 살아 있는 동안 그 종목 **진입만** 거절(D11 latch) | 2차 결정 1(D), 사람 결정 P2-3 |
| **R4** | CLI `tossctl flatten` 게이트웨이의 취소 409 `already-processing` → `retryAfterSeconds` 를 따라 누적 5s·최대 2회 재시도. **소진·중단 시 오늘 경로(AMBIGUOUS → IN_DOUBT → `resolveStep`)** | 2차 결정 2, 1.3 재리뷰 F1 |
| **R5** | 기존 401 재전송(`Client.send`)을 「확실히 미접수」로 — **코드 불변, 시험 고정** | 1차 결정 3 |

**범위 밖**: 주문 접수(place)의 모든 재전송 — 진입 결과 불명 재전송·정착 전 목록 대조는 **a171**(미해소 F2·F4·F6·A2·F8, 잔여 R-a165-1~4 이관),
429 갈래(a166 뒤), 청산 매도 접수(0회), 정정(Amend — 결정 6, 409 처리도 오늘 그대로), 조건주문, 재생 `Attested`, CLI `tossctl order …`, journal 스키마.

## D1. 지금의 결함 — 엔진에서 취소 409 는 그 종목 손절을 재시작까지 막는다 (1.3 재리뷰 F5)

- 분류(AST): `classifyRefusalCode`(번들 `internal-execgw--classifyrefusalcode`)의 확정 목록에 `already-processing` 이 없어 B6(74) 판정 없음 →
  `classifyMutation` B8(84) `statusOf` B1(127) → 409 → `ClassifyHTTPMutation` B7(337) default → **AMBIGUOUS** → `DispatchVerified` B16(204) MarkInDoubt → IN_DOUBT.
- 엔진에는 **세션 중 IN_DOUBT 해소기가 없다**. `Resolver.Resolve` 의 생산 호출자는 부팅 복구(`internal/reconcile/recovery.go:300`, `runtime.go:289–296` 의
  `Recover`)와 CLI flatten(`flatten.go:428`)뿐이다(grep). 
- 그래서 청산 전 미체결 정리의 취소(`exitloop.go:1686`)가 409 를 받으면 그 attempt 는 IN_DOUBT 로 남고, 다음 주기부터 `clearTheSymbol` B3(1639)
  `UnsettledOnSymbol` 에 걸려 그 종목은 **엔진 재시작이나 운영자 해소 전까지** 치워지지 않는다 — 그 종목의 손절이 나가지 못한다.
  already-processing 의 실측 의미는 「주문 접수를 아직 처리 중」(M16, `verifylive/mutate.go:370–373`)이고 1초 뒤 1회 재시도로 수락됐다(R9).

## D2. R3 — 엔진: 확정 미수락으로 기록하고 다음 주기에 다시 취소

- 판정 `alreadyProcessingRefusal(err) (wait time.Duration, ok bool)` — 새 함수, 순수:
  `errors.As(err, *official.APIError)` 이고 `Code == 409` 이고, `refusalCodes(Body)`(번들 `internal-execgw--refusalcodes`)로 읽은 최상위 `code`·`error.code` 가
  **두 자리 모두 있으면 서로 같아야** 하고(다르면 거짓 — 기존 「모순은 모호 강제」 규칙과 같은 방향), 있는 값이 대소문자 무시 전체 일치로 `already-processing`.
  status 만 보거나 문구를 읽지 않는다. wait 는 `error.data.retryAfterSeconds`(openapi 경로, 1.3 A9) — 숫자가 아니거나 없거나 ≤0 이면 1s.
- 적용: `plan.kind == journal.KindCancel` 이고 게이트웨이에 `CancelRefusal` 정책이 있을 때만. place·amend 의 409 는 오늘 그대로(AMBIGUOUS).
- 엔진 정책 `EngineCancelRefusal()`: `Retries 0`, `Definitive true`. 첫 결과가 409 AP 면 바로
  `DispatchOutcome{Class: REJECTED, ReasonCode: broker_rejected, Detail: "…already-processing… [a165 …]"}` → `DispatchVerified` B6(158) → FAILED_CONFIRMED.
  **대기·재전송 0** — 브로커 호출 1회, 시계 진행 0(시험으로 고정).
- 결과(AST `ExitObserver.clearTheSymbol`): B11(1697) `out.State != StateConfirmed` → `cleared=false; continue` — 같은 주기의 다음 주문으로 바로 간다(지연 0).
  attempt 는 종결이므로 다음 주기 B3(1639) `UnsettledOnSymbol` 에 걸리지 않고, 같은 working order 를 새 결정으로 **다시 취소**한다.
- 부작용(1.3 2차 리뷰 P1): 종결된 attempt 는 `PendingAttempts`(recovery.go:30–32, 상태 RECORDED·DISPATCH_STARTED·ACKED·IN_DOUBT 만)에서 빠지므로
  `checkSymbolFree` B2(804)의 SymbolInFlight 도 풀린다 — a165 이전(취소 IN_DOUBT)에는 그 종목 진입이 막혀 있었다. 사람 결정 P2-3 으로 **진입만** 다시 막는다: D11.
  공용 `unsettledFor`(evidence AST B1–B4, 32–41)를 넓히지 않는다 — `clearTheSymbol` B3(1639)이 같은 판정을 읽으므로 넓히면 exit 이 막힌다.
- 안전: 그 주문은 살아 있다(취소가 실행되지 않음). 엔진은 그 주문을 지우지 않는다 — `LiveOrdersForSymbol` 은 CONFIRMED place·amend 와 체결 종결 증거만 보고
  취소의 FAILED_CONFIRMED 로 주문을 닫지 않는다(1.3 재리뷰 A2 확인, `journal/fills.go:1854–1885` `LiveOrdersForSymbol`). 매도는 `cleared` 가 참일 때만 나간다.
  잘못 분류됐을 때(취소가 실은 실행됨)의 결과는 다음 주기 재취소가 「이미 취소됨」 응답을 받는 것뿐이다.
- 반복 비용: 브로커가 계속 already-processing 이면 5s 주기마다 그 종목 GET(사전 확인)+POST 1쌍 ≈ 0.4 req/s/종목. 경보는 유지된다(`cleared=false` →
  `noteDelay`·`noteClearFailure`, `exitloop.go:1387–1391`). a166 3 req/s 안에서 동시에 막힌 종목 7개까지.

## D3. R4 — CLI flatten: 재시도, 소진하면 오늘 경로

- flatten 정책 `FlattenCancelRefusal()`: `Retries 2`, `MaxWait 5s`(누적), `Window 8s`(**바깥 래퍼가 원 dispatch 클로저를 처음 부르기 직전**부터 — P3, D5), `Definitive false`.
- 루프(새 함수, D5 의 바깥 래퍼 안 — 원 dispatch 클로저는 고치지 않는다):
  1. 첫 결과가 409 AP(D2 판정 + `Class == AMBIGUOUS`)가 아니면 그대로.
  2. 매 재시도 전: `d = min(wait, MaxWait − 누적)`, `d ≤ 0` 이거나 `since(start) + d ≥ Window` 이면 중단. `g.clk.Sleep(winCtx, d)` 실패면 중단. 잠든 뒤 `since(start) ≥ Window` 면 중단.
  3. 재전송은 **같은 원 dispatch 클로저를 `winCtx` 로 다시 부르는 것**이다(D5) — 클로저가 마지막 경계 검사(결정 재읽기 B30·`checkDecision` B31·키 일치 B32·
     `checkProtection` B33·`checkEntry` B34·`checkReservation` B36)와 `call` 안의 B37(685) `dctx.Err()` 를 그대로 다시 지난다. 그 `dctx` 가 이제 `winCtx` 이므로
     B37 이 창 마감(실시간 deadline)·호출자 취소를 `plan.call` 바로 앞에서 잡는다. 거절·창 닫힘이면 클로저가 NOT_SENT 를 돌려주므로(보낸 것 없음) 중단.
  4. 재전송 결과가 409 AP 면 계속, 아니면 그 결과로 끝.
- 최종(F1 수리): **브로커가 ACK 하거나 409 AP 가 아닌 자기 답(확정 거절·결과 불명)을 줬을 때만 그 답**으로 정착한다. 소진·중단·재전송의 전송 전 실패
  (`NOT_SENT`)이면 **첫 409 의 오늘 분류(AMBIGUOUS)** 로 정착한다 — 그래서 `settleStep` B9·B10(395·400) → `resolveStep` 동기 해소, 종목은
  `UnsettledCancelSymbols`(Pending·InDoubt·Held)에 남아 `planLiquidation` B5(307)가 그 종목 매도를 보류한다(oversell 가드 유지).
  `NOT_SENT` 를 그대로 두면 `classifyCancel` B3(458)이 `FlattenStepFailed`(=「이미 없음」, `flatten_test.go:487–489`)로 읽으므로 금지한다.
- 앞선 409 AP 전송은 브로커가 처리를 거절한 것이라(실행 안 됨) 강등 대상이 아니다. 재전송의 확정 거절(예: 404 — 주문이 이미 사라짐)은 브로커의 답이다.
- 즉시성(AST `Saga.settleStep`·`Saga.resolveStep`): 오늘은 409 1건마다 `resolveStep` 이 그 자리에서 동기로 돈다(해소 최소 45s, 최대 5분 — `DefaultResolveConfig`).
  R4 는 흔한 경우(R9: 1s·1회)에 그것을 ~1s+왕복으로 줄이고, 최악(소진)에는 오늘 경로 앞에 ≤ `min(5s + 2×(GET+POST), 8s)` 를 더한다.
  **최악 추가 시간이 0 이 아니다** — 단 그 최악은 브로커가 5s 동안 계속 처리 중이라고 답한 경우이고, 그 뒤의 동기 해소(최소 45s)가 지배한다.

## D4. 즉시성 요약 (안전 불변식 3)

| 경로 | a165 뒤 |
|---|---|
| 엔진 청산 매도 접수 | 변화 없음 |
| 엔진 청산 전 취소 | 409 AP 1건당 같은 주기 지연 **0**(시험 고정). 다음 주기 재취소 — 오늘은 재시작까지 막힘(D1) |
| CLI flatten 취소 | 흔한 경우 단축(동기 해소 45s+ → ~1s). 최악 +≤8s 뒤 오늘 경로(D3) |
| CLI flatten 매도 접수 | 변화 없음 |
| 진입 접수 | 변화 없음(a171). 단 R3 latch(D11)가 선 종목은 진입 거절 — a165 이전(취소 IN_DOUBT 로 SymbolInFlight)과 같은 방향 |
| 엔진 exit·손절·flatten·reconcile | latch 를 읽지 않음(D11 — `checkSymbolFree` B3(810)에서 반환). 지연 0 |

## D5. 구현 구조 — 원 dispatch 클로저는 그대로, 작은 바깥 래퍼 (사람 결정 P2-2)

**3판의 `sendOnce` 추출(646–725 를 지역 클로저로 옮김)은 폐기한다.** 기존 dispatch 클로저(AST `gateway.submit` B29–B48, 636–726)의 본문·순서·반환은 한 글자도 바꾸지 않는다.

- `submit` 편집은 **한 줄**: B49(728) `if plan.strategy == nil` 갈래의 `attempt.DispatchVerified(ctx, dispatch, g.roundTripFor(plan))`(729)를
  `attempt.DispatchVerified(ctx, g.withCancelRefusal(plan, dispatch), g.roundTripFor(plan))` 로 바꾼다. B50/B51(730–731)·B52(732–733) strategy 갈래는 `dispatch` 그대로 —
  취소 계획은 `strategy` 를 갖지 않는다(`Gateway.Cancel` 의 plan 구성 404–417, 손으로 읽음; `plan.strategy` 는 PlaceClaimedStrategy 만).
- `withCancelRefusal(plan, dispatch) journal.DispatchFunc`(새 함수, 새 파일): `g.cancelRefusal == nil || plan.kind != journal.KindCancel` 이면 **받은 `dispatch` 를 그대로 돌려준다.**
- 래퍼 본문(취소 + 정책 있음일 때만):
  1. `start := g.clk.Now()`, `wallStart := time.Now()` — **원 클로저를 처음 부르기 전**(P3 — 창 시작점은 dispatch 진입 직전, 보수적. 첫 전송의 B30–B36 재검사 시간도 창에 든다).
  2. `first := dispatch(dctx, a)` — 원 클로저를 **원 ctx** 로 1회. 첫 전송은 오늘과 같은 클로저·같은 ctx 이므로 같은 순서·같은 거절·같은 분류.
  3. `alreadyProcessingRefusal(first.Err)`(D2)가 참이고 `first.Class == DispatchAmbiguous` 일 때만 계속(그 밖은 `first` 그대로). `first.Err` 는 `classifyMutation` 의 status 갈래가
     원 오류를 그대로 싣는다(`classifyMutation` 번들 B8(84) 갈래의 `outcome.Err = err`(86) — 줄은 손으로 읽음).
  4. 엔진 정책(`Definitive`): 재호출·대기 없이 REJECTED + 꼬리표(D2·D6).
  5. flatten 정책: `winCtx, cancel := context.WithDeadline(dctx, wallStart.Add(Window)); defer cancel()` 뒤 D3 루프 — **재전송은 `dispatch(winCtx, a)`, 즉 같은 클로저를 `winCtx` 로 다시 부름.**
- **B37(685) 마지막 취소 확인은 각 재호출 직전에도 적용된다**: 재호출은 클로저 전체를 다시 들어가므로 B30–B36 재검사 뒤 `call` 클로저의 첫 문장
  `if err := dctx.Err(); err != nil { return err }`(685)를 `plan.call`(708) 바로 앞에서 다시 지난다. 그 `dctx` 가 `winCtx` 이므로 창 마감(실시간)·`cancel()`·호출자 취소 셋 다
  여기서 걸린다 → `call()` 오류 → B47(717) `notSent(…)` → 래퍼가 「재전송의 전송 전 실패」로 받아 첫 409 의 AMBIGUOUS 로 정착(D3 최종). 주입 시계 판정(D3 2)과 실시간 deadline 의 두 겹.
- **정착 ctx 는 `winCtx` 와 분리된다(불변식)**: `DispatchVerified` 는 받은 `ctx` 를 그대로 dispatch 에 넘기고(dispatch.go:144 `outcome := dispatch(ctx, a)`) 모든 정착을 같은 `ctx` 로 한다 —
  번들 `internal-journal--attempt.dispatchverified` B5(153 Settle NotDispatched)·B7(160 Settle FailedConfirmed)·B9(174)·B11(181 MarkAcked)·B12(185 verify)·B13(191)·B15(199 Settle Confirmed)·
  B17(206 MarkInDoubt)·B19(215). 래퍼는 `DispatchOutcome` **값**만 돌려주고 ctx 를 돌려주지 않으므로 `winCtx` 는 정착에 닿을 수 없다. 그래서 재시도 창이 끝나거나 `defer cancel()` 이 돌아도
  결과 기록은 끊기지 않는다. 호출자가 바깥 `ctx` 를 취소하면 정착도 실패하는 것은 오늘과 같다(그 attempt 는 DISPATCH_STARTED 로 남고 `RecoverPending` B6(103)·B7(105)이 재시작 때 IN_DOUBT 로 바꾼다 — evidence `journal--journal.recoverpending`).
- **정책 nil = 오늘과 바이트 동일(AST 근거)**: 정책 nil 이면 `withCancelRefusal` 이 받은 클로저 값을 그대로 돌려주므로 B49 갈래가 `DispatchVerified` 에 넘기는 함수 값이 오늘과 같다.
  클로저 B29–B48 은 편집 0, `DispatchVerified` 는 그것을 dispatch.go:144 에서 정확히 1회 부른다(편집 0). 취소가 아닌 계획(place·amend)도 같은 이유로 동일. strategy 갈래 B50–B52 는 편집 0.
  달라지는 것은 B49 의 인자 식 하나뿐이고, 그 식의 값은 nil 정책에서 원 `dispatch` 다. 회귀 시험: 정책 nil 취소 409 AP → IN_DOUBT, detail 바이트 동일, 브로커 호출 1.
- 클로저가 바깥 변수 `result` 를 덮으므로 `out.Result` 는 마지막 `plan.call` 의 값이다(정착 분류와 무관 — 2.1 Pre-Edit 에서 소비자 확인).
- `journal.DispatchFunc` 의 문서 주석 「It must not retry」(dispatch.go:48–49)는 GREEN 에서 「a165 취소 409 AP 예외 — 같은 attempt·같은 대상 주문, 재전송 횟수는 detail 꼬리표」로 고친다(주석만).

## D6. 기록

- 스키마 변경 없음. R3·R4 가 분류를 바꾸거나 재전송했으면 detail 끝에 `[a165 sends=<N> branch=cancel_already_processing stop=<사유> outcome=<definitive|retried|today>]`.
  ReasonCode 는 기존 어휘(`broker_rejected` / 원래 값)만 쓴다(`AllReasonCodes` 고정, `failclosed.go:254`).
- 409 가 아니거나 정책이 nil 이면 detail 은 오늘과 바이트 동일.
- 꼬리표는 D11 latch 의 식별자이기도 하다. 쓰는 함수(꼬리표 생성)와 읽는 함수(끝의 `[a165 …]` 한 덩어리만 파싱)는 같은 새 파일의 한 상수를 공유하고, 왕복 시험으로 묶는다.
- 401 재전송(R5)은 공식 클라이언트 안의 일이라 attempt 기록에 남지 않는다 — spec 에서 명시 면제(오늘 동작).

## D7. R5 — 401 재전송 근거 시험

AST `client.send` B4(344) `attempt < 2 && code == http.StatusUnauthorized`, 루프 안 `tm.refresh` 뒤 `makeReq(tok)` 재생성(354), B8(358) `!adopted` 면 break;
`tokenManager.refresh` B1(119) 유효한 다른 캐시 토큰이면 채택(121) 아니면 교환(125). `postWithHeaders` B1(431) 본문 1회 마샬.

시험(official, httptest, 실 엔드포인트 없음): ① 주문 POST 401 → 재전송 본문이 **바이트 동일**(같은 clientOrderId), 교환 경로 재전송 1회 ② 401 이 아닌
4xx·429·5xx 는 POST 1회 ③ 401 이 끝까지 오면 `ErrAuth`(→ 게이트웨이 REJECTED, 기존). 채택 경로(최대 2회)는 기존 `token_shared_cache_test.go` 가 다루는지 확인하고
없으면 추가. 실측 근거 없음(R-a165-1)은 잔여 — a171 로 이관 기록.

## D8. 켜는 법·되돌리기

- `execgw.Options.CancelRefusal *CancelRefusal`. nil = 오늘 동작.
- 엔진 `buildGateway`(internal/app/engine/gateway.go:311): `CancelRefusal: execgw.EngineCancelRefusal()`.
- CLI `buildFlattenWiring`(cmd/tossctl/flatten.go:233): `CancelRefusal: execgw.FlattenCancelRefusal()`.
- 되돌리기: 그 필드 삭제 → 재빌드(교체는 KST 05–09). 런타임 스위치 없음. 스키마·설정 변경 없음.
  필드를 지워도 latch 판정(D11)은 남는다 — 이미 쓴 R3 행이 해제될 때까지 진입을 막는다(보수 방향). 새 R3 행은 더 생기지 않는다.
  이전 바이너리로 되돌리면 latch 판정 자체가 없으므로 R3 행이 있어도 진입이 허용된다(그 바이너리는 R3 행을 만들지 않으므로 오늘 동작과 같음).
- 정본 retry matrix §0.1·§1·§4 와 이 기본값 상수를 GREEN 커밋에서 함께 바꾼다.

## D9. rate 예산(§0.4)

정상 상태 추가 0. 엔진: 409 AP 가 계속되는 종목마다 ≈0.4 req/s(D2). flatten: 취소 1건당 최대 (GET+POST)×2. a166 과 파일 표면이 겹치지 않는다.

## D10. 시험 전략

- RED(fake broker 에 `*official.APIError` 주입 + flatten 은 기존 Saga 시험 틀): 엔진 정책 409 AP → FAILED_CONFIRMED·호출 1회·시계 불변 / 같은 409 의
  code 불일치·두 자리 모순·status 만 409 → 오늘(IN_DOUBT) / place·amend 409 AP → 오늘 / 정책 nil → 오늘과 바이트 동일 / flatten 정책: 1회 409 뒤 ACK →
  CONFIRMED + detail 꼬리표, 대기 = retryAfter, 누적 5s 상한·횟수 상한·창 경계(가짜 시계) / 소진 → IN_DOUBT → `resolveStep` → 그 종목 Held(매도 보류) /
  재전송이 전송 전 실패 → IN_DOUBT(Failed 아님) / 재전송 직전 경계 거절 → 중단·IN_DOUBT / retryAfter 비숫자·없음 → 1s / 엔진 `clearTheSymbol` 다음 주기 재취소.
  R5 ①②③. D11 latch RED (a)~(h).
- **바깥 래퍼 RED(P2-2, 1.3 2차 처분)**:
  - (w1) **창 시작점(P3)**: fake broker 가 첫 호출 안에서 가짜 시계를 7s 진행시킨 뒤 409 AP(retryAfter 1)를 돌려주면 `since(start)+1s ≥ 8s` → 재전송 0 → IN_DOUBT.
    시작점을 첫 결과 뒤로 옮기면 재전송 1 이 되어 FAIL 이다.
  - (w2) **호출자 취소**: fake broker 가 첫 409 AP 를 돌려주기 직전에 시험의 바깥 ctx 를 취소 → `Sleep(winCtx)` 실패 → 브로커 호출 1(재전송 0), attempt 는 CONFIRMED·FAILED_CONFIRMED 가 아님
    (바깥 ctx 가 죽어 정착도 실패하므로 DISPATCH_STARTED 로 남고 `RecoverPending` 이 IN_DOUBT 로 바꿈 — 오늘의 호출자 취소와 같은 결과).
  - (w3) **창 마감 뒤 정착**: 시험용 `CancelRefusal{Retries 2, MaxWait 5s, Window 10ms(실시간), Definitive false}` + fake broker 가 첫 호출에서 실시간 50ms 늦게 409 AP →
    `winCtx` 는 이미 마감 → 재전송 0 → attempt 가 **IN_DOUBT 로 영속**(정착 성공, `res.Err` 는 409). 정착 ctx 가 `winCtx` 면 Settle 이 `context deadline exceeded` 로 실패해 FAIL.
  - (w4) **B37 재적용**: 재전송 직전 경계(B33 보호·B34 진입·B36 예약 중 하나를 fake 로 거절)를 둘째 호출에만 켬 → 브로커 호출 1, IN_DOUBT. 그리고 `winCtx` 를 `cancel()` 한 뒤의
    재호출이 `plan.call` 에 닿지 않음(B37 → B47 `notSent`)을 호출 수로 단언.
  - (w5) **nil 동일성**: 정책 nil 게이트웨이의 취소 409 AP → 브로커 호출 1, detail·reason·state 가 a165 이전 골든과 바이트 동일. 정책 있음 + place/amend 409 AP 도 같은 골든.
- **배선 시험 4 건(1.3 3차 P1-1 + 2차 독립 리뷰 P2 — 정책 값과 행동을 함께)**: 취소의 FAILED_CONFIRMED 는 경로마다 뜻이 반대다 — 엔진(R3)에서는 「그 취소가 실행되지 않음, 주문은 살아 있음」이고
  flatten 의 `classifyCancel` B3(458)은 FAILED_CONFIRMED 를 「주문이 이미 없음」(`FlattenStepFailed` → `planLiquidation` B5(307) 보류 안 함)으로 읽는다. 그래서
  **`buildFlattenWiring` 의 필드 하나가 안전과 oversell 을 가른다**. Saga 시험은 자기 게이트웨이를 직접 만들므로 배선 변이를 못 보고, 포인터·non-nil 비교는 `Retries=2, Definitive=true` 같은 혼성 변이를 놓친다.
  - (W1) `cmd/tossctl` flatten 배선: `buildFlattenWiring` 이 만든 게이트웨이의 정책이 **정확히** `{Retries 2, MaxWait 5s, Window 8s, Definitive false}` 이고, **같은 시험에서** 그 실제 배선 +
    409 AP 를 계속 주는 fake broker 로 flatten 을 돌려 취소 소진 → IN_DOUBT → `resolveStep` → 그 종목 Held, 그 종목 매도 POST 0(`planLiquidation` B5(307) 보류)을 단언한다.
  - (W2) `internal/app/engine` 엔진 배선: `buildGateway` 의 정책이 **정확히** `{Retries 0, Definitive true}`(MaxWait·Window 는 0 — 쓰이지 않음)이고, 같은 시험에서 409 AP 1회 →
    FAILED_CONFIRMED·브로커 호출 1·가짜 시계 불변, `clearTheSymbol` 은 `cleared=false` 로 같은 주기 매도 POST 0.
  - (W3) flatten 모양 nil: `CancelRefusal` 없는 게이트웨이로 같은 flatten → 409 AP 1회에 브로커 호출 1, detail 바이트 동일(a165 이전 골든), IN_DOUBT → `resolveStep` → Held·매도 보류(오늘 경로).
  - (W4) 엔진 모양 nil: `CancelRefusal` 없는 게이트웨이로 같은 청산 전 취소 → 브로커 호출 1, detail 바이트 동일, IN_DOUBT, 다음 주기 `clearTheSymbol` B3(1639)이 막힘(오늘 경로).
  - 기존 `cmd/tossctl/flatten_test.go:115–144` 는 소스 문자열만 읽는다(손으로 읽음) — W1 이 행동 시험의 첫 자리다. 배선 seam 이 없으면 2.1 Pre-Edit 에서 정하되, 생산 `buildFlattenWiring`
    이 지도자인 실행이 하나는 있어야 한다.
- **변이 원장**(3.1 — 각 변이에 잡는 시험 이름을 적고, 양성 대조군 + 「변이가 대상 줄에 닿음」 표식):
  - 판정: code 판정(EqualFold→status 만), 두 자리 모순 검사 제거, 409 검사 제거, kind==Cancel 검사 제거, `first.Class == AMBIGUOUS` 검사 제거, Definitive 분기 뒤집기.
  - 루프: 누적 MaxWait 상한 제거, 창 검사 제거, 횟수 경계(2→3), retryAfter 기본값 1s 변경, NOT_SENT→첫 AMBIGUOUS 되돌림 제거, 소진 시 마지막 결과로 정착, 엔진 정책 Retries 0→2.
  - **ctx(2차 리뷰 P2)**: ① `winCtx` 부모를 `dctx` 에서 분리(`context.Background()`/`context.WithoutCancel(dctx)`) → (w2) 가 재전송 1 로 잡음 ② 정착 ctx 를 `winCtx` 로 교체(`winCtx` 를 submit 에서 만들어
    `DispatchVerified(winCtx, …)`) → (w3) 가 Settle 실패로 잡음 + AST 구조 시험(B49 갈래 `DispatchVerified` 첫 인자 = `submit` 의 `ctx` 매개변수) ③ 재호출에 `winCtx` 대신 원 `dctx` → (w3)/(w4) 의 실시간 마감 단언
    ④ 창 시작점을 첫 결과 뒤로 → (w1).
  - **배선**: flatten←엔진 정책 · 엔진←flatten 정책 · 배선 nil 둘 · **혼성 `{Retries 2, Definitive true}`**(W1 의 정확 값 + Held 단언이 잡음).
  - **latch(D11)**: ① latch 호출을 `checkSymbolFree` B3(810) 앞으로 옮김(exit 도 latch 를 봄) → RED-b ② latch 를 `unsettledFor` 에 넣음 → RED-b(`clearTheSymbol` 이 막힘) ③ 해제 r1(대상 주문 terminal) 누락 → RED-c1
    ④ 해제 r2(확인된 취소) 누락 → RED-c2 ⑤ 키에서 종목 비교 제거 → RED-e ⑥ latch 를 메모리(`EntryGate.BlockSymbol`)로만 → RED-d
    ⑦ **r2 순서 경계 제거**(`r_C > s_L` 삭제 또는 r2 를 `ConfirmedCancelOf` 호출로 되돌림) → RED-f ⑧ **r2 대상 id 비교 제거**(`c.target_order_id = L.target_order_id` 삭제) → RED-h ⑨ 순번 비교를 C 의 정착(CONFIRMED) transition 으로 → RED-g2
    ⑩ **r2 거래일 같음 조건 복원**(`TRIM(ci.trading_day) = TRIM(l.trading_day)` 추가 — 또는 r2 를 `ConfirmedCancelOf(…, L 의 거래일, …)` 로 되돌림) → RED-i1. 1.3 3차 재리뷰 R3-3 으로 옛 ⑩(「r2 거래일 비교 제거 → RED-h(다른 거래일)」)은 **폐기** — 거래일 비교 제거가 이제 설계다.
    ⑪ **r2 계좌·시장 비교 제거**(⑪a `account` 줄 삭제 → RED-j1, ⑪b `market` 줄 삭제 → RED-j2, ⑪c `symbol` 줄 삭제 → RED-j3 — 각각 따로 심는다)
    ⑫ **r2 주인 유일성 제거**(`owners.n = 1` 삭제 또는 `>= 1`) → RED-k1, `n = 1` 을 `n <= 1` 로 → RED-k2.
  - 공유 목록: `definitiveRefusalCodes` 에 already-processing 추가(아래).
- **`definitiveRefusalCodes` 불변(1.3 3차 P3-3)**: already-processing 판정은 새 함수에만 두고 공유 목록(`refusal_code.go:46`)에 넣지 않는다 — 넣으면 place·amend·flatten 취소의 409 AP 가
  `classifyMutation` B4(45)에서 확정 거절이 된다(flatten 은 F1 재발). 변이 「목록에 already-processing 추가」는 place·amend 409 AP → 오늘(IN_DOUBT) 시험과 flatten 소진 → Held 시험이 잡아야 한다.
- 영향 범위: `internal/execgw`(+`-race`), `internal/official`, `internal/flatten`, `internal/app/engine`(clear·배선), `cmd/tossctl`(flatten 배선), `internal/journal`(dispatch·latch 질의).

## D11. R3 진입 전용 latch (사람 결정 P2-3, 1.3 2차 리뷰 P1)

**목적**: R3 가 취소를 FAILED_CONFIRMED(종결)로 닫으면 그 attempt 가 `PendingAttempts` 에서 빠져 `checkSymbolFree` B2(804)의 SymbolInFlight 가 풀린다(D2 부작용). 대상 주문은 살아 있으므로
그 종목의 **새 진입만** 다시 막는다. exit·손절·flatten·reconcile·보호 경로는 막지 않는다(안전 불변식 4).

| 항목 | 정의 |
|---|---|
| **기록 위치** | **새 표·새 열 없음 — schema 36 그대로.** R3 attempt 행 자체가 latch 다: `mutation_attempts`(schema.go:224–238)의 `kind='CANCEL'`·`state='FAILED_CONFIRMED'`·`reason_code='broker_rejected'`·`target_order_id`(230) 와 `detail` 끝의 D6 꼬리표(`branch=cancel_already_processing … outcome=definitive`). 종목·시장·계좌·거래일은 그 행의 `intents`(202–218). |
| **키** | (계좌 `intents.account_ref`, 시장 `intents.market`, 종목 `intents.symbol`, 원 주문 id `mutation_attempts.target_order_id`). 진입 판정은 (plan.market, plan.symbol) 로 그 키의 행을 찾는다(`attemptTargets` 와 같은 대소문자 무시 비교, gateway.go:838–839). |
| **세우는 조건** | R3: 엔진 정책 게이트웨이의 취소가 409 AP(D2 판정)이고 래퍼가 REJECTED + `outcome=definitive` 꼬리표로 정착(`DispatchVerified` B7(160) Settle FAILED_CONFIRMED). 따로 쓰는 단계가 없다 — 정착이 곧 세움(같은 트랜잭션). R4(flatten)는 FAILED_CONFIRMED 를 만들지 않으므로(D3: 재시도 성공 → CONFIRMED, 소진 → IN_DOUBT) latch 를 세우지 않는다. |
| **해제 조건** | 키의 행마다 아래 중 하나가 원장에 있으면 그 행은 풀림. 모든 행이 풀려야 진입 허용. **r1 원 주문 terminal**: `target_order_id` 가 그 계좌·시장·종목의 `Journal.LiveOrdersForSymbol` 결과에 없음 — 「working」 의 정의를 새로 만들지 않고 exit 청소가 쓰는 같은 판정(fills.go:1854–1918, `confirmedOrderTerminalEvidence` exit_proposal_release.go:281–294 = `terminal=1` fill snapshot)을 쓴다. terminal 은 `brokerstate.State.IsTerminal`(derive.go:94–102: FILLED·CANCELLED·CANCELLED_PARTIALLY_FILLED·REPLACED·REJECTED·REJECTED_PARTIALLY_FILLED). **r2 latch 뒤의 취소 성공(1.3 재리뷰 R3-1)**: 그 latch 행의 `FAILED_CONFIRMED` 정착 transition 보다 **뒤에 기록이 시작된**(RECORDED transition 이 더 큰 순번) 같은 대상 주문 id 의 CONFIRMED CANCEL 이 같은 대상 주문의 정규 범위(계좌·시장·종목·`target_order_id`, **거래일은 비교하지 않음** — 1.3 3차 재리뷰 R3-3)에 있고 그 번호의 주인이 하나뿐임 — 새 질의 `Journal.ConfirmedCancelAfterLatch`(아래 「r2 질의」). 엔진 다음 주기 재취소가 이 경로다. latch 보다 **앞선** CONFIRMED 취소는 해제 근거가 아니다 — 반례: C1 CONFIRMED(BUY 취소 접수, terminal snapshot 없음 → `LiveOrdersForSymbol` 에 남음) → 다음 주기 재취소 R3=409 AP → 시각을 안 보는 r2 는 C1 을 찾아 latch 를 즉시 풀고 살아 있을 수 있는 주문 위에 BUY 가 나간다. 순서는 벽시계(`settled_at`·`recorded_at`)가 아니라 원장 순번으로 정한다. 운영자가 원장 밖에서 낸 취소(`tossctl order cancel`·브로커 앱)는 r2 가 아니다 — D12. **r3 대사로 종결**: 별도 갈래가 아니다 — 원장에 terminal 증거를 쓰는 생산 호출자는 체결 감지·대사 경로의 `filldetect/ledger.go:80` `Journal.RecordFill` 하나뿐(grep, 2026-10-28)이므로 r3 은 r1 과 같은 술어로 판정된다. 대사의 불일치 상태(RECONCILE)는 차단이지 해제가 아니다. **증거 기준**: 원장 행만(브로커 즉석 조회 0, 메모리 상태 0). 읽기 오류는 해제가 아니라 오류(진입 전송 없음). |
| **재시작 내구성** | 판정을 진입마다 원장에서 다시 계산하므로 메모리 투영이 없다 — 재구성할 것이 없다. FAILED_CONFIRMED 는 terminal 이라 `RecoverPending` 이 건드리지 않는다(evidence AST B2(93) range 의 입력이 `PendingAttempts`(87) — 상태 목록 recovery.go:32 에 FAILED_CONFIRMED 없음). `internal/journal` 에 `DELETE FROM mutation_attempts` 0 건(grep, 2026-10-28) — 행이 지워지지 않는다. |
| **보는 경로** | `Gateway.checkSymbolFree` 의 B3(810) `!plan.raisesExposure` 반환 **뒤**, B7 뒤·828 앞(번들 `internal-execgw--gateway.checksymbolfree`). `raisesExposure` 가 참인 것은 place 매수(gateway.go:377)·amend 증량(447)뿐이고 취소는 거짓(416) — **AST 상 취소·청산 매도·flatten·보호 매도는 B3(811)에서 반환해 latch 질의에 닿지 않는다.** exit 청소 `clearTheSymbol` 은 `checkSymbolFree` 가 아니라 `UnsettledOnSymbol`(=`unsettledFor`, 편집 0)을 읽고, flatten 은 `UnsettledCancelSymbols`(편집 0)를 읽으며, reconcile 은 주문 경로가 아니다. |
| **사유 어휘** | 기존 `ReasonSymbolInFlight`(failclosed.go:266) — a165 이전 같은 상황(취소 IN_DOUBT)이 진입에 준 사유와 같다. detail: `cancel of order <id> on <symbol> was refused as already-processing (attempt <id>); the order is still working`. `AllReasonCodes` 변경 0. |

- **구현**: 새 파일의 `g.cancelRefusalLatch(ctx, plan) (*RejectedError, error)` + journal 새 파일의 읽기 함수(종목의 FAILED_CONFIRMED CANCEL 행 목록 — 일반 질의, a165 의미는 execgw 의 꼬리표 판정이 가짐).
  해제 r1 은 기존 `LiveOrdersForSymbol` 을 **호출**한다 — 같은 판정을 SQL 로 다시 쓰지 않는다(판정이 둘이면 서로의 시험을 통과시킨다). r2 는 기존 `ConfirmedCancelOf` 를 쓰지 않고 새 질의를 부른다(아래). 같은 latch 행은 한 번만 묻는다.
- **r2 질의(1.3 재리뷰 R3-1)** — journal 새 파일(latch 행 읽기와 같은 파일)의 `func (j *Journal) ConfirmedCancelAfterLatch(ctx, latchAttemptID string) (AttemptRecord, bool, error)`.
  - **순서 기준 = `attempt_transitions.id`**: `id INTEGER PRIMARY KEY AUTOINCREMENT`(schema.go:246) — AUTOINCREMENT 라 재사용·역행이 없고, 원장 주석이 이 표의 「순서가 기록 대상」이라 쓴다(core_domain.go:247–250 「its order is the thing being recorded」, schema.go:244 「Append-only transition log. Rows are never updated or deleted」).
    쓰기는 DSN `_txlock=immediate`(journal.go:216–218·225)로 직렬이라 순번 = 커밋 순서다. `attempt_transitions` 를 고치거나 지우는 문은 0 건(`UPDATE/DELETE … attempt_transitions` grep, 2026-10-28 — INSERT 만 durability.go:451·650, lineage.go:174, strategy_dispatch_runtime.go:795·935).
  - 상태 전이와 그 transition 행은 같은 트랜잭션이다(AST): `Attempt.transition`(evidence `journal--attempt.transition`, 548–674) — B17(635) UPDATE 실패·B19(642) 영향 행 ≠1·B20(649) transition INSERT 실패·B23(668) Commit 실패는 모두 반환하고 `defer tx.Rollback`(557)이 둘 다 되돌린다.
    RECORDED 도 같다: `Journal.Prepare`(evidence `journal--journal.prepare`, 373–479) B14(450) RECORDED transition INSERT 실패 → 반환, Commit B16(466). 그래서 취소 경로(Prepare → `DispatchVerified` 의 Settle/MarkInDoubt → `transition`)가 원장에 남긴 상태에는 반드시 그 순번 행이 있다. 다른 두 INSERT 자리(lineage.go:157→174 정정 확정, strategy_dispatch_runtime.go:787→795·928→935 전략 진입)도 같은 트랜잭션 안의 UPDATE→INSERT 쌍이다(손으로 읽음 — 취소 계획은 둘 다 지나지 않음).
  - **비교**: latch 행 L 의 `to_state='FAILED_CONFIRMED'` transition 순번 `s_L` 과, 후보 C 의 `from_state='' AND to_state='RECORDED'` transition 순번 `r_C` 에 대해 `r_C > s_L`(엄격). 정착 순번(C 의 CONFIRMED transition)이 아니라 **기록 시작 순번**을 비교한다 —
    「L 정착 뒤 정착」만 보면 L 보다 먼저 보내져 브로커가 먼저 받은 취소가 늦게 커밋되기만 해도 해제가 된다(인과가 아니라 커밋 지연에 기댐). `r_C > s_L` 이면 C 의 정착 순번도 `s_L` 보다 크다(같은 attempt 의 RECORDED 가 먼저) — 「latch 뒤에 정착」 조건을 포함한다.
    순번은 유일하므로 동률이 없다. `s_L` 이나 `r_C` 행이 없으면 SQL 비교가 NULL → 해제 아님(fail-closed).
  - **범위·대상 = 대상 주문의 정규 범위(1.3 3차 재리뷰 R3-3 으로 고침)**: L 의 행에서 계좌·시장·종목과 `target_order_id` 를 **질의 안에서** 읽는다 — 호출자가 범위를 넘기지 않으므로 다른 주문·다른 범위를 섞을 자리가 없다.
    축별 계약(L·C 양쪽에 같은 정규화 — `ConfirmedCancelOf` 의 `TRIM(account_ref)`·`LOWER(TRIM(market))`·`UPPER(TRIM(symbol))`, exit_proposal_release.go:156–158):

    | 축 | 원장 자리 | 규칙 |
    |---|---|---|
    | 계좌 | `intents.account_ref`(schema.go:206) — L·C 각자의 intent | `TRIM` 같음 |
    | 시장 | `intents.market`(schema.go:204) | `LOWER(TRIM)` 같음 |
    | 종목 | `intents.symbol`(schema.go:207) | `UPPER(TRIM)` 같음 |
    | 대상 주문 id | `mutation_attempts.target_order_id`(schema.go:230) | 바이트 같음, L 의 값이 `''` 면 found=false |
    | **거래일** | `intents.trading_day`(schema.go:205) | **비교하지 않음** — 아래 「거래일을 뺀 이유」 |
    | 순번 | `attempt_transitions.id`(schema.go:246) | `r_C > s_L`(그대로) |
    | **주인 유일성**(번호 재사용 방어) | 접수 확정 주인 = `mutation_attempts.state='CONFIRMED' AND kind IN ('PLACE','AMEND') AND broker_order_id`(schema.go:231) + 그 intent | L 의 계좌·시장·종목에서 `broker_order_id = L.target_order_id`(바이트) 인 주인의 `COUNT(DISTINCT intent_id) = 1`. 0 이거나 2 이상이면 found=false |

    C 는 `kind='CANCEL' AND state='CONFIRMED' AND c.id <> L.id`. L 이 `kind='CANCEL' AND state='FAILED_CONFIRMED'` 가 아니면 found=false(해제 아님). 읽기 오류는 오류(진입 전송 없음). `ORDER BY r_C LIMIT 1`.
  - **거래일을 뺀 이유(R3-3)**: latch 는 진입 판정에서 시장·종목으로만 찾으므로(`attemptTargets` 와 같은 비교, gateway.go:838–839) D0 의 latch 는 D1 에도 선다. 그런데 엔진 재취소 C 의 intent 거래일은 **대상 주문의 거래일이 아니라 C 를 준비한 날**이다 —
    `Gateway.prepareRequest`(evidence `execgw--gateway.preparerequest`, 1002–1086)의 1011 `day, err := market.TradingDay(g.clk.Now())` 대입, 실패는 B3(1012) 반환뿐이고 그 `day` 가 intent `TradingDay`(1039)로 들어간다.
    재취소의 대상 id 는 `clearTheSymbol`(번들 `internal-app-engine--exitobserver.clearthesymbol`)이 B4(1645) `range live` 의 `order.OrderID` 를 1686 `o.opts.Submit.Cancel` 에 넘긴 값이고 `live` 는 1630 `LiveOrdersForSymbol(계좌, 시장, 종목)` — 거래일 인자가 없다.
    그래서 거래일을 같게 요구하면 D0 latch → D1 재취소 CONFIRMED 는 `r_C > s_L` 이어도 r2 가 아니어서, 체결 감지 r1 이 늦거나 실패하면 진입 차단이 풀리지 않는다(재리뷰 반례). 「latch 거래일 이후」(`>=`) 로 바꾸는 안은 버렸다 — 순번 `r_C > s_L` 이 이미 시간 순서를 정하므로 날짜 조건은
    벽시계(intent 의 `g.clk.Now()`)로 같은 것을 다시 묻는 판정이 되고(판정이 둘), 어느 시험도 그 조건만 따로 죽이지 못한다.
  - **번호 재사용은 주인 유일성으로 막는다**: 주문 번호는 거래일을 건너 재사용된다(exit_proposal_release.go:151–152 주석, 시험 `TestLiveOrdersForSymbolKeepsAReusedCurrentDayOrderAfterThePriorDayTerminal` fills_test.go:985 ·
    `TestTrackedAndLiveOrdersKeepTwoNonterminalReusedTradingDays` fills_test.go:635). 원장은 주문의 정규 범위를 그 번호를 받은 **접수 확정 주인**(PLACE/AMEND attempt + intent)에서 읽는다 —
    `LiveOrdersForSymbol`(evidence `journal--journal.liveordersforsymbol`, 1854–1919; SQL 은 손으로 읽음)의 `confirmed_orders`(fills.go:1860–1867)와 주인 술어 `1 = COUNT(DISTINCT owner.intent_id)`(1871–1878, 같은 계좌·시장·거래일·종목·매매방향).
    r2 는 거래일 축이 없으므로 그 술어를 **거래일·매매방향 없이** L 의 계좌·시장·종목 전체에서 센다: 같은 번호의 주인이 둘이면(재사용) C 가 어느 주문을 취소했는지 원장이 말하지 못하므로 해제하지 않는다(fail-closed, r1 은 그대로).
    다른 계좌·시장·종목의 같은 번호는 계좌·시장·종목 축이 거른다. 매 진입마다 다시 세므로, 해제 뒤에 재사용 주인이 생기면 latch 가 다시 선다(보수 방향).
  - **r2 질의 모양**(구현 자리는 journal 새 파일, SQL 은 이 모양을 따른다 — D12 질의 2 가 같은 조건을 쓴다):

    ```sql
    WITH l AS (
      SELECT a.id, a.target_order_id AS target, TRIM(i.account_ref) AS account,
             LOWER(TRIM(i.market)) AS market, UPPER(TRIM(i.symbol)) AS symbol,
             (SELECT MAX(t.id) FROM attempt_transitions t
               WHERE t.attempt_id = a.id AND t.to_state = 'FAILED_CONFIRMED') AS latch_seq
        FROM mutation_attempts a JOIN intents i ON i.id = a.intent_id
       WHERE a.id = ? AND a.kind = 'CANCEL' AND a.state = 'FAILED_CONFIRMED' AND a.target_order_id <> ''
    ), owners AS (
      SELECT COUNT(DISTINCT o.intent_id) AS n
        FROM l, mutation_attempts o JOIN intents oi ON oi.id = o.intent_id
       WHERE o.state = 'CONFIRMED' AND o.kind IN ('PLACE','AMEND') AND o.broker_order_id = l.target
         AND TRIM(oi.account_ref) = l.account AND LOWER(TRIM(oi.market)) = l.market AND UPPER(TRIM(oi.symbol)) = l.symbol
    )
    SELECT c.id, (SELECT MIN(t.id) FROM attempt_transitions t
                   WHERE t.attempt_id = c.id AND t.from_state = '' AND t.to_state = 'RECORDED') AS recorded_seq
      FROM l, owners, mutation_attempts c JOIN intents ci ON ci.id = c.intent_id
     WHERE c.kind = 'CANCEL' AND c.state = 'CONFIRMED' AND c.id <> l.id AND c.target_order_id = l.target
       AND TRIM(ci.account_ref) = l.account AND LOWER(TRIM(ci.market)) = l.market AND UPPER(TRIM(ci.symbol)) = l.symbol
       AND owners.n = 1 AND recorded_seq > l.latch_seq
     ORDER BY recorded_seq LIMIT 1;
    ```

    `latch_seq` 는 `MAX`, `recorded_seq` 는 `MIN` — 행이 여럿이어도 보수 쪽(늦은 latch·이른 기록)이고, 없으면 NULL 이라 비교가 거짓(fail-closed).
  - **기존 `ConfirmedCancelOf` 는 바꾸지 않는다**: 질문이 다르다 — 그것은 「그 거래일에 그 주문의 엔진 취소가 접수 확정된 적이 있는가(가장 이른 것)」(exit 의 역사 조회, 「판정이 둘」이 아님)이고 r2 는 「이 latch 뒤에」다. 생산 호출자 둘(grep `ConfirmedCancelOf(`, 2026-10-28):
    `internal/app/engine/exitloop.go:1650`(`clearTheSymbol` — 엔진 취소가 확정된 매도는 종결 증거를 기다림), `internal/app/engine/exit_held_proposal.go:178`; 시험 `internal/journal/a094_exit_proposal_release_test.go:311·315·319`. 시간 경계를 넣으면 이 둘의 손절 보류 판정이 바뀐다 — 그래서 새 함수.
  - **schema 영향 0**: 새 표·열·색인 없음, schema 36 그대로. 질의는 `idx_transitions_attempt(attempt_id, id)`(schema.go:255)와 `idx_attempts_state` 로 걷는다.
- **비용**: 노출을 늘리는 제출마다 질의 1(종목 필터, `idx_attempts_state`) + R3 행이 있는 종목만 대상 id 별 2 질의. 엔진 진입 빈도에서 무시할 수준(측정 안 함 — 3.3 에서 잴 것).
- **오탐 방향**: 꼬리표는 우리 코드가 detail **끝**에 붙이고 읽는 쪽은 끝의 한 덩어리만 본다. 브로커 문구가 우연히 같은 꼬리표로 끝나는 비-R3 행은 진입을 더 막을 뿐이다(보수 방향).
  꼬리표 형식이 바뀌어 읽기가 놓치는 것(진입 허용 — 위험 방향)은 쓰기·읽기 상수 공유 + 왕복 시험 + RED-a 가 막는다.
- **고려하고 버린 안 — `EntryGate.BlockSymbol`**(symbolgate.go:64–89): 진입 전용 축은 같다(`Gateway.checkEntry` evidence AST B1(851) `g.entry == nil || !plan.raisesExposure` → nil — `checkSymbolFree` B3 와 같은 `raisesExposure`).
  그러나 evidence AST(`execgw--entrygate.blocksymbol`)의 호출은 `strings.*`·`g.Block`·`symbolKey`·`g.mu.*`·`make`·`g.clk.Now` 뿐이고 원장 쓰기가 0 — 메모리 map(81) 이다. 재시작 내구성에는
  `RebuildReconcileProjection`(symbolgate.go:183–218, reconcile 사유만 재구성) 같은 원장 투영을 새로 만들어야 하고, 해제 생산자 셋(체결 감지·취소 성공·…)에 `ClearSymbol` 호출을 심어야 하며, 새 사유 코드(`AllReasonCodes` 고정,
  failclosed.go:245–250 「rename 은 data migration」)가 필요하다 — High-risk 편집이 셋 이상 는다. 원장 유도는 편집 1(checkSymbolFree 한 문장) + 새 파일 둘이다. 대가: 콘솔 `SymbolBlocks` 에 보이지 않는다(거절 사유·detail 로만 보임 — 잔여 R-a165-6, 조회 수단은 D12 ①).
- **RED 계획**(`internal/execgw` 실 journal + fake broker, 실계좌 0):
  - (a) 엔진 정책 게이트웨이로 종목 A 의 working BUY 주문 취소가 409 AP → R3 행 → 같은 종목 A 매수 place 가 `ReasonSymbolInFlight` 로 거절, 브로커 place 호출 0, journal 에 place attempt 없음.
  - (b) 같은 상태에서 종목 A 의 엔진 재취소(cancel)·청산 매도(place SELL)·flatten 취소/매도가 `checkSymbolFree` 를 통과하고 브로커 호출에 닿음, `clearTheSymbol` 은 다음 주기 재취소를 냄, reconcile 주기는 영향 없음. `g.journal` 은 구체 타입(`*journal.Journal`)이라 질의 횟수 계측 fake 가 없다 — 「exit 은 latch 질의에 닿지 않음」은 AST 구조 시험(`checkSymbolFree` 안 `cancelRefusalLatch` 호출이 B3 반환 뒤에 있음)으로 고정한다.
  - (c) 해제: (c1) 대상 주문에 terminal fill snapshot 기록 → 진입 허용 (c2) latch 정착 **뒤에 Prepare 한** 대상 주문의 CANCEL attempt CONFIRMED 기록 → 진입 허용 (c3) 둘 다 없고 R3 행이 둘(연속 주기) → 계속 거절 (c4) 다른 주문의 terminal·취소 성공은 해제하지 않음.
  - (d) 재시작: R3 행을 쓴 뒤 Journal·Gateway 를 닫고 같은 파일로 새로 열어(`RecoverPending` 포함) → 진입 거절 유지. 메모리 latch 변이(⑥)는 여기서 FAIL.
  - (f) **latch 앞의 확인된 취소(1.3 재리뷰 R3-1 반례)**: 대상 주문 X 의 CANCEL C1 을 Prepare·CONFIRMED 정착 → 같은 X 의 엔진 재취소 409 AP → R3 latch → 종목 A 매수 place 는 `ReasonSymbolInFlight` 로 거절, 브로커 place 호출 0. terminal snapshot 은 쓰지 않는다(r1 이 우연히 풀지 않게).
  - (g) **latch 뒤의 확인된 취소만**: (g1) (f) 상태에서 X 의 CANCEL C2 를 Prepare·CONFIRMED → 진입 허용. (g2) 기록 시작과 정착의 구분 — C3 를 latch **전에** Prepare 하고 latch 정착 **뒤에** CONFIRMED 로 정착(journal 직접, 게이트웨이 우회) → 계속 거절. 순번 비교 대상을 C 의 정착 transition 으로 바꾸면 (g2) 가 FAIL.
  - (h) **다른 대상**: latch 뒤 같은 계좌·시장·종목의 다른 주문 Y 의 CANCEL CONFIRMED → 계속 거절.
    **갱신 기록(1.3 3차 재리뷰 R3-3)**: 옛 (h) 의 둘째 문장 「같은 X 라도 다른 거래일의 CONFIRMED 취소 → 계속 거절」은 **뜻이 뒤집혀 지웠다** — 그 모양은 이제 (i1) 의 양성(해제)이다. 거래일이 막던 번호 재사용은 (k) 가 맡는다.
  - (i) **거래일을 넘긴 해제(R3-3 반례)**: (i1) D0 에 X 의 엔진 재취소 409 AP → R3 latch L(L 의 intent 거래일 D0) → 게이트웨이 fake 시계를 다음 거래일 D1 로 → 엔진 재취소 C 가 CONFIRMED(C 의 intent 거래일 D1, `r_C > s_L`), terminal snapshot 없음
    → 종목 A 매수 place 허용(브로커 place 호출 1). 거래일 같음 조건을 되살리면(⑩) FAIL. (i2) D0 latch → D1 에 X 의 terminal fill snapshot(X 의 주인 범위 = D0) 기록 → 허용(r1 도 거래일을 넘겨 푼다; `LiveOrdersForSymbol` 이 거래일 인자를 받지 않음 — evidence AST 1854).
  - (j) **다른 범위의 같은 번호**: latch L(계좌 acct-1·시장 kr·종목 A·대상 X) 뒤, `r_C > s_L` 인 CONFIRMED CANCEL 의 대상이 바이트 같은 X 이지만 (j1) 계좌가 다름 (j2) 시장이 다름(us) (j3) 종목이 다름 → 셋 다 계속 거절(found=false).
    게이트웨이는 계좌가 하나라 (j1) 은 journal 층에서만(원장 직접 기록), (j2)·(j3) 은 두 층.
  - (k) **번호 재사용(주인 유일성)**: (k1) X 의 접수 확정 주인이 L 의 계좌·시장·종목에 둘(D0 BUY 주인 + D1 에 같은 번호를 받은 다른 intent 의 주인 — `TestTrackedAndLiveOrdersKeepTwoNonterminalReusedTradingDays` 와 같은 모양)이면
    latch 뒤 X 의 CONFIRMED 취소가 있어도 계속 거절. (k2) X 의 주인이 0(주인 행 없이 L·C 만 원장에 직접 기록) → 계속 거절. (k3) 대조군: 주인 하나 + latch 뒤 CONFIRMED 취소 → 허용((g1)과 같은 모양, ⑫ 변이가 대상 줄에 닿는지의 양성 대조).
  - (f)~(k) 는 `ConfirmedCancelAfterLatch` 단위 시험(journal)과 진입 판정 시험(execgw) 두 층에 둔다 — journal 층은 found/AttemptRecord.ID, execgw 층은 거절 사유·브로커 호출 수.
  - (e) 다른 종목 B·같은 종목의 다른 시장 진입은 영향 0. latch 읽기 오류 → 진입 거절(전송 0)은 오류 주입 seam 이 있을 때만 시험하고, 없으면 2.1 Pre-Edit 에 `not-applicable` 사유를 남긴다(오류 경로는 submit B8(525) 의 기존 처리).

## D12. 운영 — latch 조회·journalled 해제·경보 (1.3 재리뷰 R3-2, R-a165-5·6)

> 이 절이 a165 의 운영 문서 자리다. GREEN(2.3) 커밋에서 `docs/operations.md` 에 「엔진 취소가 already-processing 으로 거절된 뒤 진입이 막힐 때」 절로 옮긴다(아래 ①~④ 그대로, 질의 문구는 그때의 schema 로 다시 확인).
> 새 코드·새 명령·새 알림 채널은 없다. 아래 명령 중 `mutating: true` 인 것은 사람이 직접 실행한다(대화형 에이전트 자동 실행 금지).

**① 조회 — 기존 표면 점검 결과와 정한 수단**

| 기존 표면 | latch 를 보여 주나 | 근거 |
|---|---|---|
| 콘솔 `SymbolBlocks` | 아니오 — `EntryGate` 메모리 투영만 그린다(D11 「버린 안」) | symbolgate.go:64–89 |
| 콘솔 주문 화면 | 아니오 — 브로커 주문과 원장 출처(엔진/그 밖)만 | `internal/console/orders.go:702–731`(`ledgerOrderEvidence`) |
| `tossctl engine alerts list` | 아니오 — 미배달 critical 행의 개수·나이만, 본문 없음 | cmd/tossctl/engine_alerts.go:49–63 |
| `tossctl engine attempt-resolve` | 해당 없음 — UNRESOLVED_IN_DOUBT park 만 다룬다 | cmd/tossctl/engine_attempt_resolve.go:3–10 |
| 진입 거절 사유 | 부분 — 게이트웨이가 `symbol_in_flight` + detail(`cancel of order <id> … (attempt <id>)`, D11 사유 어휘)을 호출자에게 돌려준다. 운영자 화면까지 가는 경로는 이 라운드에 추적하지 않음 | D11 |

→ 정한 수단: **읽기 전용 원장 조회 절차**(새 읽기 표면 0). 원장은 `sqlite3 'file:<journal.db>?mode=ro'` 로만 연다(경로는 엔진 profile 규칙 — `consoleJournalPath`, cmd/tossctl/console_journal.go:12–17. 원장을 읽기 전용으로 여는 선례: operations.md:264 의 `mode=ro`·`query_only`).
조회는 **후보 목록**이다 — 해제 판정은 코드(`cancelRefusalLatch`·`ConfirmedCancelAfterLatch`)의 것이고, 이 결과로 원장을 고치지 않는다.
질의 2 는 같은 범위(계좌·시장·종목·바이트 같은 대상 id)의 CANCEL 을 **상태와 무관하게** 나열하고 `r2` 열(1 = r2 후보)을 붙인다 — `r2=1` 인 행이 있어도 **판정이 아니라 후보**다(코드의 SQL 과 이 문구가 갈리면 코드가 정본; 2.3 이식 때 구현 SQL 과 다시 대조). `owner_count` 가 1 이 아니면 그 대상의 r2 는 없다(번호 재사용·주인 없음 — r1 만 남음).

```sql
-- 질의 1: latch 후보(R3 행)와 그 정착 순번 — 계좌 열 포함(1.3 3차 재리뷰 R3-4)
SELECT a.id AS latch_attempt, i.account_ref, i.market, i.trading_day, i.symbol, a.target_order_id, a.settled_at,
       (SELECT MAX(t.id) FROM attempt_transitions t WHERE t.attempt_id = a.id AND t.to_state = 'FAILED_CONFIRMED') AS latch_seq
  FROM mutation_attempts a JOIN intents i ON i.id = a.intent_id
 WHERE a.kind = 'CANCEL' AND a.state = 'FAILED_CONFIRMED'
   AND a.detail LIKE '%[a165 %branch=cancel_already_processing%outcome=definitive]'
 ORDER BY latch_seq DESC;

-- 질의 2: 질의 1 의 latch 하나에 대한 r2 후보 — 코드 r2(D11 「r2 질의 모양」)와 같은 범위 조인·주인 유일성·순번을 SQL 에서 직접 판정
--         (1.3 3차 재리뷰 R3-4). 거래일은 출력만 하고 비교하지 않는다(D11 「거래일을 뺀 이유」).
WITH l AS (
  SELECT a.id, a.target_order_id AS target, TRIM(i.account_ref) AS account,
         LOWER(TRIM(i.market)) AS market, UPPER(TRIM(i.symbol)) AS symbol, i.trading_day AS latch_day,
         (SELECT MAX(t.id) FROM attempt_transitions t
           WHERE t.attempt_id = a.id AND t.to_state = 'FAILED_CONFIRMED') AS latch_seq
    FROM mutation_attempts a JOIN intents i ON i.id = a.intent_id
   WHERE a.id = '<latch_attempt>' AND a.kind = 'CANCEL' AND a.state = 'FAILED_CONFIRMED' AND a.target_order_id <> ''
), owners AS (
  SELECT COUNT(DISTINCT o.intent_id) AS n
    FROM l, mutation_attempts o JOIN intents oi ON oi.id = o.intent_id
   WHERE o.state = 'CONFIRMED' AND o.kind IN ('PLACE','AMEND') AND o.broker_order_id = l.target
     AND TRIM(oi.account_ref) = l.account AND LOWER(TRIM(oi.market)) = l.market AND UPPER(TRIM(oi.symbol)) = l.symbol
)
SELECT c.id AS cancel_attempt, c.state, TRIM(ci.account_ref) AS account, LOWER(TRIM(ci.market)) AS market,
       UPPER(TRIM(ci.symbol)) AS symbol, c.target_order_id, l.latch_day, ci.trading_day AS cancel_day,
       l.latch_seq,
       (SELECT MIN(t.id) FROM attempt_transitions t
         WHERE t.attempt_id = c.id AND t.from_state = '' AND t.to_state = 'RECORDED') AS recorded_seq,
       owners.n AS owner_count,
       (c.state = 'CONFIRMED' AND owners.n = 1
        AND (SELECT MIN(t.id) FROM attempt_transitions t
              WHERE t.attempt_id = c.id AND t.from_state = '' AND t.to_state = 'RECORDED') > l.latch_seq) AS r2
  FROM l, owners, mutation_attempts c JOIN intents ci ON ci.id = c.intent_id
 WHERE c.kind = 'CANCEL' AND c.id <> l.id AND c.target_order_id = l.target
   AND TRIM(ci.account_ref) = l.account AND LOWER(TRIM(ci.market)) = l.market AND UPPER(TRIM(ci.symbol)) = l.symbol
 ORDER BY recorded_seq;
```

r1(대상 주문 terminal)의 원장 확인은 체결 감지가 쓰는 종결 기록이다 — 확인 순서는 operations.md 「체결 감지가 멈췄을 때」(docs/operations.md:565–)의 ①·③ 과 같고, 원 SQL 은 2.3 에서 `LiveOrdersForSymbol` 이 읽는 표(fills.go:1854–)로 확정해 옮긴다.

**② 지원되는 journalled 해제 경로** — 원장에 해제 근거가 남는 것만

1. **r2 엔진 재취소**(자동): 청산 루프가 그 주문을 계속 치우려 하면 다음 주기 `clearTheSymbol` 이 `Gateway.Cancel` 로 다시 취소한다. latch 정착 뒤에 기록이 시작된 그 CANCEL 이 CONFIRMED 로 정착하면 풀린다 — **거래일을 넘긴 재취소도 푼다**(D0 latch → D1 재취소, D11 「거래일을 뺀 이유」). 단 그 대상 번호의 접수 확정 주인이 그 계좌·시장·종목에 둘 이상이면(번호 재사용) r2 는 없고 r1 만 남는다. 근거 행 = 그 CANCEL attempt + RECORDED transition 순번 + `owner_count`(질의 2 의 `r2` 열).
2. **r1 대상 주문 종결**: 브로커에서 그 주문이 체결·취소·거부로 끝나면 체결 감지(`filldetect`)가 종결 fill snapshot 을 쓰고(`JournalLedger.Apply` → `Journal.RecordFill`, internal/filldetect/ledger.go:80) 풀린다. 대상 주문은 엔진이 CONFIRMED 로 낸 주문이라(`LiveOrdersForSymbol` 의 `confirmed_orders`) 체결 감지의 추적 목록에 든다(ledger.go:112–113 「the orders that are not terminal, plus the ones a confirmed attempt named」 — 손으로 읽음).
   - **운영자가 그 주문을 직접 취소하는 것(브로커 앱 · `tossctl order cancel`)은 그 자체로 원장 해제가 아니다.** `tossctl order cancel` 은 `app.tradingService.Cancel` 직행(cmd/tossctl/order.go:295)이라 journal attempt 를 남기지 않고, 그래서 r2 가 되지 않는다. 해제는 그 뒤 체결 감지가 그 주문의 종결을 읽어 원장에 쓸 때(r1) 일어난다 — 체결 감지가 멈췄으면 풀리지 않는다.
3. **해제가 아닌 것**: `tossctl engine attempt-resolve` — park 가 아니면 거절하므로 R3 행(FAILED_CONFIRMED)에 쓸 수 없다(engine_attempt_resolve.go:9–11 「명령 시점에 park 가 아니면 거절」). 원장 손 편집 — 금지(operations.md 「④ 하지 말 것」과 같은 이유). `tossctl flatten-all [--cancel-only]` — 같은 원장에 `Gateway.Cancel` 로 기록하므로 r2 를 만들 수 있으나 계좌 전체를 취소하고 진입 차단을 스스로 세운다(cmd/tossctl/flatten.go:69–80) — 해제 수단이 아니라 비상 청산이다.

**③ 경보 — 기존 경로**

- latch 가 서는 R3 는 청산 루프의 청산 전 취소에서만 난다. 그 주기는 `cleared=false` 라 기존 두 경보가 그대로 돈다(코드 변경 0): `noteDelay`(exitloop.go:1910–1934 — 허용 지연을 넘기면 `exit.liquidation_delayed`)와
  `noteClearFailure`(exit_held_proposal.go:266–296 — 연속 `obs.DefaultCriticalAttempts`=3(internal/obs/notifier.go:46) 회 실패면 critical `exit.liquidation_delayed`, 본문 「a working order on the symbol could not be taken off the book」(exit_held_proposal.go:45)). 배달·미배달·`engine alerts ack` 는 기존 critical 경로 그대로.
- **경보가 없는 상태**: 청산 루프가 그 주문을 더는 치우지 않는데(포지션 종결 등) 대상 주문이 원장에서 working 으로 남은 경우 — 진입만 막히고 경보는 없다. 게이트웨이에는 경보 sink 가 없고(`Gateway` 필드, gateway.go:118–145), 그것을 배선하는 것은 새 High-risk 배선이라 a165 는 하지 않는다. 잔여 R-a165-6(수락 — 진입만 막는 보수 방향, 위 ① 조회로 확인).

**④ 둘 다 오지 않을 때 — 누가·무엇으로**

- 담당: 운영자(사람). 증거: 브로커 앱의 그 주문 상태(**읽기만**) + 질의 1·2 + 체결 감지 상태.
- 브로커에서 주문이 이미 끝났는데 latch 가 남으면 → 체결 감지 문제다: operations.md 「체결 감지가 멈췄을 때」 / 「감지는 건강한데 종결 증거가 오지 않을 때」(docs/operations.md:610) 를 따른다.
- 브로커에서 주문이 아직 working 이면 → latch 는 맞게 서 있다. 그 종목 진입이 필요하면 사람이 그 주문을 취소하고(②-2 의 경로로 r1 을 기다림) — 그 전까지 진입 거절 유지. 수동 해제 명령은 없다(만들지 않는다).
- exit·손절·flatten 은 latch 와 무관하게 열려 있다(D11 「보는 경로」) — 이 절차 중 엔진을 멈추지 않는다(엔진 정지 = 손절 정지).
