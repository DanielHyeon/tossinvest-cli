# a198 주장 대조 (tasks 1.3)

- 기준: HEAD `e1eaf720c`(= base). AST 는 `analysis/cited-ast/<Receiver_Func>.ast.json`(tools/logic-map, 같은 HEAD). 분기 id 는 그 파일의 `branches[].id`.
- probe 재실행: `analysis/probe/probe-output.txt`(a197 원본 그대로, `zz_a197_probe_test.go.txt` 는 a197 브랜치 원본과 바이트 동일) · `probe-ext-output.txt`(같은 파일 끝에 덧붙인 (b) 확장 `zz_a198_probe_ext.go.txt`). 둘 다 버리는 worktree 의 미커밋 시험 파일로 돌렸고 worktree 는 지웠다.
- 판정: ✓ = AST/실측과 맞음 · △ = 맞지만 범위·조건이 더 좁음(정정) · ✗ = 틀림 · ? = 이 단계에서 잴 수 없음.

## proposal 「Why」 표

| # | 주장 | 판정 | 근거 |
|---|---|---|---|
| P1 | 길 치우기는 취소마다 `Issuer.IssueReduction` 으로 결정을 발행하고 그것을 `Submit.Cancel` 에 싣는다(exitloop.go:1683-1716) | ✓ (줄 ±1: 호출은 1684) | ExitObserver_clearTheSymbol B4(range L1665) 안 호출 L1684(`calls`) → B9(L1696 발행 실패) → L1706 `Submit.Cancel`(Decision: issued.Decision) |
| P2 | 운영의 `Issuer` 는 `RiskGuardian`(exitwiring.go:326·339) | ✓ | Context_ExitObserver B3(L327, 단언 실패) · 대입 L339 |
| P3 | `RiskGuardian.IssueReduction` 은 매도·`KindPlace` 로 고정, 매도 아니면 거절(:592·:614) | ✓ | RiskGuardian_IssueReduction B2(L592) · L614 `Kind: journal.KindPlace`. 주문 생성 결정에 키: DecisionRequest_build B22(L373-375) |
| P4 | 그 결정으로 취소하면 원장이 거절, 브로커 취소 0 | ✓ (재실측) | probe-output.txt: BUY·SELL 둘 다 `state= reason=invalid_mutation_request broker_cancels=0 … CANCEL carries no idempotency key`. 자리: prepareRequest L1081 `req.ClientOrderID = decision.ClientOrderID` → PrepareRequest_validateDecisionBinding B12(L267) |
| P5 | 키만 없애서는 통과하지 않는다 — 방향·종류가 맞아야. 매수 취소는 `checkPreimage` 방향 비교(:333), 모든 취소는 `checkIdempotencyKey`(:398)가 또 막는다 | **△ 정정** | (1) 결정 행에는 종류 열이 없다 — `journal.Decision` 에 Kind 필드 0, 종류는 키 유무로만 남는다(DecisionRequest_build B22). (2) 매수 주문: 키 없는 매도 결정도 checkPreimage B5(L333)에서 거절 — 실측 `BUY-order/decision-SELL/no-key → guardian_intent_hash_mismatch … authorises side SELL but the mutation carries BUY`(probe-ext). (3) 매도 주문: 키가 없으면 매도 방향 결정은 B5 를 통과한다 — 「키만 없애서는 통과하지 않는다」는 매수 주문에만 참. (4) checkIdempotencyKey B2(L398)는 키가 **있을 때만** 막는 둘째 벽(Prepare B12 뒤). → proposal 줄을 이 사실로 고침(결정 D1 불변) |
| P6 | flatten 모양(CANCEL·주문 방향) 결정은 접수(flatten.go:360·659-667 → issue.go:86-103) | ✓ (재실측) | probe-output.txt `issuer-like-flatten state=CONFIRMED broker_cancels=1`(BUY·SELL). Saga_settleStep L360 · Saga_decisionFor → Issuer_IssueReduction(분기 1: validate) |
| P7 | 실패한 자리에 기록 없음(1717-1720), `record` 가 noteDelay(:1402)·noteClearFailure(:1403)를 돌리지만 문장에 원인이 없다 | ✓ | clearTheSymbol B11(L1717) 본문 = `res.cleared=false; continue` 뿐. B9(L1696)·B10(L1702)도 같다. record B12(L1399) → L1402·L1403. noteDelay 본문 = `why`(exitClearWhyWorkingOrder 고정 문장) |
| P8 | 운영에서 발현 증거 없음(Manager 운영 장부 실측) | ? | 운영 장부·로그 접근 금지(이 작업 범위) — 대조하지 않음 |
| P9 | 기존 시험은 결합을 지나지 않는다(fakeSubmitter.Cancel exitloop_test.go:110-120, a094Submitter.Cancel a094_exit_clear_test.go:100-107) | ✓ + 보충 | 두 자리 그대로. 보충: 진짜 게이트웨이+진짜 RiskGuardian 을 쓰는 엔진 하네스가 **둘 있다** — `buildE2EStack`(exit_e2e_test.go:267-277) · `TestA087RealGatewayLowerLimitAndDormantFallback`(a087_real_gateway_e2e_test.go:27-31). 둘 다 브로커 취소에 닿는 시험이 0(grep `Cancel` 0). 3.1 의 하네스 후보 |

## design

| # | 주장 | 판정 | 근거 |
|---|---|---|---|
| D1 | 취소 결정 = CANCEL·위험 감소·주문 방향·최대 수량 = 주문 수량 | ✓ (실측) | probe-ext: `BUY qty7/decmax10 CONFIRMED`, `SELL qty10/decmax10 CONFIRMED`, `BUY qty10/decmax7 → at most quantity 7 but the mutation carries 10`. 같은 값을 둘에 넣으면 checkPreimage B16(L375) 통과 |
| D2 전제 | 취소 결정에 운영 모드·킬 스위치 검사가 걸리지 않는다 | ✓ | 아래 (a) |
| D3 | 발행기 누락은 기동에서 드러난다 | — (설계 선택) | 현재 Context_ExitObserver B3 이 같은 방식. (가-B) 안에서는 대상이 사라짐 — edit-surface.md |
| D4 | 실패를 기존 에피소드 경고(`warnEpisode`)로 남긴다 | **✗ 운영에서 기록 0 — Manager 결정 필요** | `warnEpisode`(exit_held_proposal.go:336) → `o.log`(exitloop.go:1962) 첫 줄 `if o.opts.Log == nil { return }`. 운영 조립은 `opts.Log` 를 넣지 않는다(engine.go:671-683 · exitwiring.go 전수, 주석 exit_order_log.go:8 「관측자 전체 Log 는 생산에서 nil」). 기존 줄 1672(B6)도 운영에서 0 줄이다. 그대로 구현하면 spec 「원인과 함께 기록」이 운영에서 침묵한다 — 문서는 고치지 않음 |
| §2-1 | 매수 미체결: 취소 접수 → 그 주기에 치움 완료 → 손절 제출 | ✓ (코드상) + 위험 | clearTheSymbol B11 거짓 · B12 거짓(buy) → res.cleared 유지 → B13 거짓 → 반환 cleared. 위험: 브로커가 취소를 비동기로 처리하면 같은 주기의 손절 PLACE 가 `opposite-pending-order-exists`(refusal_code.go:29)로 거절될 수 있고, 그 매수는 종결 체결 기록 전까지 LiveOrdersForSymbol 에 남아(fills.go:1854 `NOT confirmedOrderTerminalEvidence`) 다음 주기에 **같은 주문을 다시 취소**한다(매수에는 ConfirmedCancelOf 대기가 없음 — B5 거짓) — 미측정 |
| §2-2 | 매도 미체결: 취소 접수 → 종결 증거 대기 → 종결 기록 뒤 손절 | ✓ | 아래 (c) |
| §2-3 | 모호 취소 → 미종결 가드가 보류, 기존 경보가 알림 | ✓ | 아래 (c) |

## (a) 취소에 걸리는 가드 — 「청산을 위한 취소는 어느 운영 모드에서도 나간다」

**참** — 운영 모드·킬 스위치·진입 게이트·위험 체인은 취소에 걸리지 않는다. 걸리는 것은 아래 열둘이고 어느 것도 운영 모드를 읽지 않는다.

| 순서 | 자리(AST) | 취소에서 | 운영 모드·킬 스위치? |
|---|---|---|---|
| 1 | Gateway_prepareRequest — 종목·시장·거래일·방향·통화 모양 거절 | 모양만 | 아니오 |
| 2 | Gateway_submit B7(L523) `claimSymbol` — 같은 종목 프로세스 내 진행 중 | 걸림 | 아니오 |
| 3 | B8/B10(L529·531) `checkSymbolFree` → Gateway_checkSymbolFree B2(L808) 같은 종목 미종결; B3(L814) `!raisesExposure` 로 반환 — UnresolvedAttempts·cancelRefusalLatch(B8) 는 노출 증가만 | B2 만 | 아니오 |
| 4 | B11/B13(L539·541) `checkDecisionUnspent` — 키 없음, nonce 소비 여부 | 걸림 | 아니오 |
| 5 | B14(L549) `journal.Prepare` → PrepareRequest_validate · validateDecisionBinding(B12 = 오늘의 결함) | 걸림 | 아니오 |
| 6 | B19(L587) `checkProtection` → Gateway_checkProtection B1(L90) `!raisesExposure` 즉시 통과 | 안 걸림 | — |
| 7 | B20(L592) `checkEntry` → Gateway_checkEntry B1(L860) `!raisesExposure` 즉시 통과 — **EntryGate(운영 모드·차단)가 사는 곳** | 안 걸림 | (여기 있으나 취소는 건너뜀) |
| 8 | B21(L598) preflight — Cancel 의 plan 에 없음 | 안 걸림 | — |
| 9 | B23·B24(L608·611) loadDecision·checkDecision(무결성·세대·계좌·등급⇔모양 B9/B10·preimage·만료·키·한도 checkLimits 위험 감소 = 한도 없음) | 걸림 | 아니오 |
| 10 | B28(L629) checkReservation → B1(L899) 위험 감소면 통과 | 안 걸림 | — |
| 11 | 전송 직전 재검(B30-B36 — loadDecision·checkDecision·키·checkProtection·checkEntry·checkReservation) | 9 와 같음 | 아니오 |
| 12 | 전송: `withCancelRefusal`(Gateway_withCancelRefusal B1: 엔진 정책 = 409 재전송 0) → `trading.Service.Cancel` → Service_guard B1(`policy.Cancel`)·B3(`AllowLiveOrderActions`)·B4(확인 토큰, 게이트웨이가 계산) → Service_Cancel B3 사전 조회 `GetOrderAvailableActions`(엔진 구현은 오류를 내지 않음, broker.go:57 → precheck.go:98 `OrderRawByID` 2초) | 걸림 | 아니오 — 정책 둘은 기동 interlock clause 3(interlock.go:597)이 켜짐을 강제, 재적재 없음 |

결정 발행 쪽: 종전 `RiskGuardian.IssueReduction` 은 `risk.Evaluate` 를 부르지만(B3 L604) Evaluate B2(L135) 매도면 `evaluateReduction`(보유 초과만; 정수 수량만 — chain.go:228 `parseWholeNumber`)으로 가고 킬 스위치는 진입 사슬(chain.go:251)에만 있다. `execgw.Issuer.IssueReduction` 은 검사 0(B1 validate 뿐). 두 안 모두 운영 모드를 읽지 않는다.
그 밖에 운영 모드와 무관하게 늦출 수 있는 것: a166 계좌 예산 admission — clearCtx 가 보호 청산이면 `accountbudget.Critical`(exitloop.go:1392-1394) 우선순위로 나가지만 retry matrix §0.6 「보호 무지연을 보장하지 않는다」.

## (b) checkPreimage·checkIdempotencyKey 가 취소에서 맞춰 보는 항목, 부분 체결

- checkPreimage(ReductionIntent 갈래 B12): 등급 = 위험 감소(B13), 계좌·시장(소문자)·종목(대문자)·**방향**(B2-B5 venue), **수량은 상한**(B16 `decimalAtMost(plan.quantity, MaxQuantity)`). 가격은 보지 않는다.
- checkIdempotencyKey: 취소면 B1 → 결정에 키가 있으면 거절(B2), 없으면 통과. checkLimits: 위험 감소는 한도 스냅숏이 **없어야** 함.
- 길 치우기가 넘기는 수량: `LiveOrder.Quantity` = 그 주문 intent 의 `i.quantity`(fills.go:1864 SELECT) — **원 주문(또는 정정 후) 수량이지 미체결 잔량이 아니다.** 원 10·체결 3 이면 결정 최대 수량에도, 취소 요청 수량에도 10 이 들어간다(clearTheSymbol L1690 `Quantity: order.Quantity` · L1700 `floatOf(order.Quantity)`) → 10 ≤ 10 통과. 브로커 취소 요청에는 수량이 없다(`CancelIntent{OrderID, Symbol}` — 잔량 전체 취소). 그 수량은 결정 preimage · 취소 intent 행 · fingerprint 에만 남는다.
- flatten: 브로커 미체결 목록 페이로드의 `quantity`(flatten.go:521·534 → `target.Quantity`)를 결정 최대 수량과 요청에 같은 값으로 넣는다(flatten.go:351·365). 그 필드가 원 수량인지 잔량인지는 이 단계에서 확인하지 않았다(fixture 미대조) — 어느 쪽이든 두 자리가 같은 값이라 비교는 통과한다.
- 실측(probe-ext): 결정 최대 7 < 요청 10 → `guardian_intent_hash_mismatch … at most quantity 7 but the mutation carries 10`, 결정 10 ≥ 요청 7 → CONFIRMED. 즉 D1 을 「잔량」으로 바꾸고 요청에 원 수량을 남기면 거절된다 — 두 자리는 같은 값이어야 한다.

## (c) 취소가 실제로 접수·모호해진 뒤 다음 주기의 분기

| 상황 | 그 주기 | 다음 주기 |
|---|---|---|
| 매수 취소 CONFIRMED | clearTheSymbol B11 거짓 · B12 거짓 → cleared → record B13(L1406) clearDelay·endClearStreak → 손절 PLACE | 손절이 접수되면 clear 를 다시 타지 않음. 손절이 안 나갔으면(위 §2-1 위험) B3 거짓(CONFIRMED 는 PendingAttempts 밖 — recovery.go:32 `RECORDED·DISPATCH_STARTED·ACKED·IN_DOUBT`) → 그 매수가 아직 LiveOrders 에 있으면 B5 거짓 → **재취소 발행**(B9·B11) |
| 매도 취소 CONFIRMED | B11 거짓 → B12 참 → `cleared=false, awaitingClose=true` → record B12(L1399) → noteDelay(why=exitClearWhyAwaitingClose; noteDelay B1 첫 호출이면 시계만 시작) · noteClearFailure(B3 거짓 → 계수) | B3 거짓 → B5 참 → `ConfirmedCancelOf`(gateway 가 쓴 `kind=CANCEL, state=CONFIRMED, target_order_id=그 주문` 을 읽음 — exit_proposal_release.go:156-158) → **B7 참 waiting** → 재취소 없음, 계속 보류. 종결 체결 기록이 생기면 그 매도가 LiveOrders 에서 빠져 루프에 안 들어옴 → cleared → B14 해제 판정 |
| 취소 IN_DOUBT | B11 참(state ≠ CONFIRMED) → cleared=false(countable 유지) → noteDelay·noteClearFailure 계수 | B3 참: UnsettledOnSymbol 이 그 CANCEL(IN_DOUBT)을 돌려줌 → `onlyEngineCancelsInFlight` B6(default — IN_DOUBT) → false → countable=true → 계속 계수 → noteClearFailure B4 3회째 critical(key `…|streak:<id>`), noteDelay B2 한계(30s) 뒤 critical. 세션 중 해소 없음 → 재시작 복구까지(정본·design §2-3 과 같음) |
| 취소 RECORDED·DISPATCH_STARTED·ACKED(중단) | — | B3 참 → onlyEngineCancelsInFlight B5 → 대상 전부가 그 취소면 true → countable=false(계수 제외), noteDelay 는 계속(record B12 는 countable 과 무관) |
| 브로커 거절(FAILED_CONFIRMED) | B11 참 → cleared=false | 다음 주기 재발행·재전송(B9 → L1706). retry matrix 행 31: 409 already-processing 도 같은 주기 재전송 0 |
| 결정 발행 실패 | B9 참 → cleared=false, 기록 없음 | 다음 주기 재시도 |

지금까지 이 판독들이 읽은 CANCEL 행은 시험이 원장에 직접 쓴 것뿐이다(a094_exit_clear_test.go:105 `a094RecordCancelOn`). 게이트웨이가 쓰는 행은 intent 를 새로 만들고(prepareRequest: side=주문 방향, quantity=주문 수량, trading_day=**취소 시점** 거래일) 그 intent 의 시장·거래일·종목으로 ConfirmedCancelOf 가 대조한다(:158 `i.trading_day = order.TradingDay`) — 주문 거래일과 취소 거래일이 다르면(전일 주문의 당일 취소) 대기가 성립하지 않는다. 시험이 쓴 행과 같은 결과인지는 3.1 결합 시험이 처음 잰다.

## D2 선행 확인 (Manager 결정 2026-11-23 — (가-B))

| 질문 | 답 | 근거 |
|---|---|---|
| 관측기 시계 = 게이트웨이가 결정 만료를 판정하는 시계(운영) | **같다** — 한 값 | cmd/tossctl/engine.go:217 `clk := clock.System()` → `engineAssemble(…, clk, …)`(:219) → cmd/tossctl/engine_assembly.go:33 `Clock: clk` → internal/app/engine/engine.go:463 `clk := opts.Clock` → :537 `clock: clk` → gateway.go:325 `execgw.New{Clock: in.clock}`. 같은 `clk` 를 `engineRuntimeFactory(ctx, ectx, clk, …)`(engine.go:277) → `ExitObserverOptions{Clock: clk}`(:672) → 관측기 `o.clk`(exitloop.go:366-375). 시험 하네스는 관측기·게이트웨이에 각자 시계를 넘기므로 같은 가짜 시계를 넘겨야 함 |
| 결정 TTL 을 flatten 과 같은 출처로 | **값은 같고 출처는 다르다** | flatten 은 `flatten.DecisionTTL = 60s`(flatten.go:63)를 넘김. Issuer 에 TTL 을 비우면 `execgw.DefaultDecisionTTL = 60s`(issue.go:30, record :139-141). 둘 다 리터럴 `60 * time.Second` — 같은 상수가 아님. engine 이 flatten 패키지를 import 하지 않으려면 execgw 기본값이 자연스러움 |
| ID 생성 | **다르다(접두만)** | flatten 은 `"flatten-" + flatten.randomID()`(flatten.go:665, 12바이트). Issuer 에 NewID 를 비우면 `execgw.randomID`(gateway.go:1117, 16바이트). 관측기 `opts.NewID`(운영 `randomExitID` 16바이트, 시험은 순차 문자열)를 넘길 수도 있으나 시험의 순차 ID 가 결정 ID·nonce 에 들어감 — 기본값(비움) 권고 |
| 결정 행 모양 | flatten 과 같음 | 같은 함수 `Issuer.IssueReduction`(issue.go:86) → `record`(:137): SafetyClass 위험 감소, ReductionIntent(account·market·symbol·side·max_quantity·reason), Generation 0, nonce, TTL |

## D4 경로 실현 가능성 (Manager 결정 2026-11-23)

- 경로: `noteExitOrder`(exit_order_log.go:198, 분기 9: B1 `ExitOrderLog == nil` → 반환, B2 refusal 래치, B7 follow 아님 → 반환) → 주기 끝 `endExitOrderCycle`(:235, ObserveOnce defer) → `writeExitOrder`(:258, recover) → `opts.ExitOrderLog`(운영 cmd/tossctl/engine.go:680 `ExitOrderLog: logger`).
- 줄 타입 `exitOrderLine` 은 종류 필드가 없고 `args()`(exit_order_log.go:138-168, 분기 1 B1 `l.fill`)가 `side "SELL"`(:143)·`order_type "LIMIT"`(:144)을 고정 출력 — 그대로는 매수 주문의 취소를 「매도 지정가」로 적는다. 취소를 표현하려면 `args` 본문 1 곳 편집(종류가 취소일 때만 종류·주문 방향을 내보내고 다른 줄은 바이트 동일)이 필요.
- 고칠 기존 함수: `exitOrderLine.args` **1**(+ 호출 자리 `clearTheSymbol`, 이미 편집 대상). `noteExitOrder`·`endExitOrderCycle`·`writeExitOrder`·`exitOrderToken` 무편집. 구조체 필드 추가는 타입 편집(함수 아님). 기준 「2 개 초과면 멈춤」 이내 → 채택.
- 필드 규율: 수량·가격 칸은 비움, 사유는 `exitOrderToken`(영문자·밑줄만) 통과 코드, 상태는 `journal.AttemptState`. a129 시험은 부분 키 대조(`a129Want`, a129_exit_order_log_test.go:204) — 새 키가 취소 줄에만 붙으면 기존 시험 무편집.
- 사건 이름: 발행 실패·브로커 전 거절 → `obs.EventOrderNotDispatched`, `FAILED_CONFIRMED` → `obs.EventOrderRejected`, IN_DOUBT → `obs.EventOrderInDoubt`(submit 의 기존 매핑 exitloop.go:1599·1608-1610 과 같은 규칙). B10(수량·가격 해석 실패)은 Manager 결정 문장 밖 — 기록하지 않음(판정 무변화).
