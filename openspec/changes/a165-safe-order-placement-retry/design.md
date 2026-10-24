# a165 design (3판) — 취소 409 `already-processing` 처리 · 401 재전송 근거 고정

> **상태**: 3판. 사용자 결정 3차(2026-10-23, F3 = 1)로 진입 재전송(R1·R2)을 a171 로 옮겼다. 1판·2판과 리뷰 처분은 `review.md`.
> base `34e12967e`. 함수 내부 주장은 `analysis/function-logic/*/ast.json` 의 AST 좌표로만 한다(23 번들).
> 실계좌 주문·`mutating: true` 명령·운영 토글 flip 은 없다. 시험은 httptest/fake 만.

## 0. 범위

| | 내용 | 근거 결정 |
|---|---|---|
| **R3** | 엔진 게이트웨이의 **취소**가 `409` + 본문 code `already-processing` → **FAILED_CONFIRMED(확정 미수락)**. 같은 dispatch 안 재시도 0, 대기 0 | 2차 결정 1(D) |
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
- 안전: 그 주문은 살아 있다(취소가 실행되지 않음). 엔진은 그 주문을 지우지 않는다 — `LiveOrdersForSymbol` 은 CONFIRMED place·amend 와 체결 종결 증거만 보고
  취소의 FAILED_CONFIRMED 로 주문을 닫지 않는다(1.3 재리뷰 A2 확인, `journal/fills.go:1854–1885` `LiveOrdersForSymbol`). 매도는 `cleared` 가 참일 때만 나간다.
  잘못 분류됐을 때(취소가 실은 실행됨)의 결과는 다음 주기 재취소가 「이미 취소됨」 응답을 받는 것뿐이다.
- 반복 비용: 브로커가 계속 already-processing 이면 5s 주기마다 그 종목 GET(사전 확인)+POST 1쌍 ≈ 0.4 req/s/종목. 경보는 유지된다(`cleared=false` →
  `noteDelay`·`noteClearFailure`, `exitloop.go:1387–1391`). a166 3 req/s 안에서 동시에 막힌 종목 7개까지.

## D3. R4 — CLI flatten: 재시도, 소진하면 오늘 경로

- flatten 정책 `FlattenCancelRefusal()`: `Retries 2`, `MaxWait 5s`(누적), `Window 8s`(첫 전송 시작부터), `Definitive false`.
- 루프(새 함수 `settleCancelRefusal`, dispatch 클로저에서 첫 결과 뒤 1회 호출):
  1. 첫 결과가 409 AP 가 아니면 그대로.
  2. 매 재시도 전: `d = min(wait, MaxWait − 누적)`, `d ≤ 0` 이거나 `since(첫 전송) + d ≥ Window` 이면 중단. `g.clk.Sleep(winCtx, d)` 실패면 중단.
  3. 재전송은 첫 전송과 같은 `sendOnce`(D5) — 마지막 경계 검사(결정 재읽기·`checkDecision`·키 일치·`checkProtection`·`checkEntry`·`checkReservation`)를
     다시 통과하고, `plan.call` 바로 앞에서 창을 다시 본다. 거절·창 닫힘이면 보낸 것이 없으므로 중단.
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
| 진입 접수 | 변화 없음(a171) |

## D5. 구현 구조 — `submit` dispatch 클로저

- 첫 전송까지의 본문(AST `gateway.submit` 646–725: 결정 재읽기 646 … `classifyMutation` 725)을 지역 클로저 `sendOnce(sctx, deadline)` 로 옮긴다.
  반환: (분류, 그 전송의 오류, 전송 여부). 첫 호출은 `deadline` 0 — **첫 전송은 오늘과 같은 순서·같은 거절 반환**(회귀 시험).
- `plan.call` 바로 앞(`call` 클로저 안)에 `deadline` 이 있으면 `g.clk.Now()` 가 그 앞인지 본다. 아니면 전송하지 않는다.
- `CancelRefusal == nil` 이거나 `plan.kind != KindCancel` 이거나 첫 전송이 없었으면 첫 분류를 그대로 반환 — **오늘 동작과 같다**.
- 재시도 ctx: 첫 전송 직전 시각부터 `Window` 인 `context.WithDeadline(dctx, …)`(단조 timer)와 주입 시계 판정 두 겹. 엔진 정책은 재시도가 없어 쓰이지 않는다.
- strategy 경로는 취소가 없으므로 해당 없음(`plan.strategy` 는 PlaceClaimedStrategy 만).

## D6. 기록

- 스키마 변경 없음. R3·R4 가 분류를 바꾸거나 재전송했으면 detail 끝에 `[a165 sends=<N> branch=cancel_already_processing stop=<사유> outcome=<definitive|retried|today>]`.
  ReasonCode 는 기존 어휘(`broker_rejected` / 원래 값)만 쓴다(`AllReasonCodes` 고정, `failclosed.go:254`).
- 409 가 아니거나 정책이 nil 이면 detail 은 오늘과 바이트 동일.
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
- 정본 retry matrix §0.1·§1·§4 와 이 기본값 상수를 GREEN 커밋에서 함께 바꾼다.

## D9. rate 예산(§0.4)

정상 상태 추가 0. 엔진: 409 AP 가 계속되는 종목마다 ≈0.4 req/s(D2). flatten: 취소 1건당 최대 (GET+POST)×2. a166 과 파일 표면이 겹치지 않는다.

## D10. 시험 전략

- RED(fake broker 에 `*official.APIError` 주입 + flatten 은 기존 Saga 시험 틀): 엔진 정책 409 AP → FAILED_CONFIRMED·호출 1회·시계 불변 / 같은 409 의
  code 불일치·두 자리 모순·status 만 409 → 오늘(IN_DOUBT) / place·amend 409 AP → 오늘 / 정책 nil → 오늘과 바이트 동일 / flatten 정책: 1회 409 뒤 ACK →
  CONFIRMED + detail 꼬리표, 대기 = retryAfter, 누적 5s 상한·횟수 상한·창 경계(가짜 시계) / 소진 → IN_DOUBT → `resolveStep` → 그 종목 Held(매도 보류) /
  재전송이 전송 전 실패 → IN_DOUBT(Failed 아님) / 재전송 직전 경계 거절 → 중단·IN_DOUBT / retryAfter 비숫자·없음 → 1s / 엔진 `clearTheSymbol` 다음 주기 재취소.
  R5 ①②③.
- 변이 원장: code 판정(EqualFold→status 만, 모순 검사 제거), 409 검사, kind 검사, Definitive 분기, 대기 상한(누적·창), 횟수 경계, NOT_SENT→오늘 되돌림 제거,
  재전송 경계 재검사 제거(deadline 검사 제거), 엔진 정책 Retries 0→2(지연 0 시험이 잡아야 함). 양성 대조군 + 「변이가 대상 줄에 닿음」 표식.
- 영향 범위: `internal/execgw`(+`-race`), `internal/official`, `internal/flatten`, `internal/app/engine`(clear·배선), `cmd/tossctl`(flatten 배선), `internal/journal`(dispatch).
