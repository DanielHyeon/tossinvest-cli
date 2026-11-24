# a197 1.3 — 인용 AST 와 주장 대조

## 0. 재개 준비 재대조 (2026-11-25, HEAD `f18de89ef` = 병합 `e7e3f22b4` + base 재고정, base `b93ed133a`)

- 판정 방법: 각 `*.ast.json` 의 `source_sha256` 을 현재 파일 해시와 대조. 다르면 같은 함수를 현재 HEAD 로 다시 뽑고(`GOFLAGS=-trimpath` 로 빌드한 `tools/logic-map` 바이너리), 위치 · 해시를 뺀 AST 를 옛것과 비교해 「본문동일 / 본문변경」 을 가름. 아래 §1~§3 은 **2026-11-23 옛 base `064f59fd9` 기준 원문**이다 — 판정이 바뀐 것은 이 절의 표가 우선한다.
- 결과: 28 개 중 **무변 19 · 재추출 9**(본문변경 2 — `Context.ExitObserver`(a199: B2 `Verified` → `RunsExitSide()`) · `ExitObserver.clearTheSymbol`(a198: 취소 결정 발행 · D8 · 실패 줄); 본문동일 7 — 줄만 이동). 새로 뽑은 a198 계열 AST **16**(아래 0.3). 합계 44.

### 0.1 재추출 표

| AST | 함수 | 판정 | 옛 → 새 좌표 · 분기 수 |
|---|---|---|---|
| config--mergeEngine | `mergeEngine` | 무변 | L330-368 · 6 |
| engine--Context.ExitObserver | `Context.ExitObserver` | 재추출 · **본문변경** | L319-365 · 6(B2@323 조건만) |
| engine--ExitObserver.alertProposalRefused | | 재추출 · 본문동일 | L1901-1918 → L1912-1929 · 0 |
| engine--ExitObserver.clearTheSymbol | | 재추출 · **본문변경** | L1649-1743 → L1652-1754 · 17 |
| engine--ExitObserver.record | | 재추출 · 본문동일 | L1314-1455 → L1317-1458 · 19 |
| engine--ExitObserver.release | | 재추출 · 본문동일 | L1623-1628 → L1626-1631 · 1 |
| engine--ExitObserver.submit | | 재추출 · 본문동일 | L1495-1616 → L1498-1619 · 17 |
| engine--isFullExit | | 재추출 · 본문동일 | L1466-1469 → L1469-1472 · 0 |
| engine--isProtective | | 재추출 · 본문동일 | L1475-1477 → L1478-1480 · 0 |
| engine--recordGateSettings | | 재추출 · 본문동일 | L739-813 → L774-848 · 6 |
| execgw--checkIdempotencyKey · checkPreimage · classifyMutation · classifyRefusalCode · Gateway.Cancel · Gateway.checkSymbolFree · Gateway.prepareRequest · Gateway.submit · Gateway.unsettledFor · Issuer.IssueReduction · OfficialOrders.OrdersPageRaw · RiskGuardian.IssueReduction · ScanOrders | 13 개 | 무변 | 좌표 그대로 |
| flatten--parseTarget · Saga.enumerate | 2 개 | 무변 | 좌표 그대로 |
| journal--EngineBrokerOrderIDs · LiveOrdersForSymbol · PendingAttempts | 3 개 | 무변 | 좌표 그대로 |

pre-edit FLM 번들 3 개(`function-logic/`)도 같은 방법으로 `ast.json` 을 다시 뽑고 문서 좌표를 옮김: `ExitObserver.submit`(본문동일, +3) · `ExitObserver.alertProposalRefused`(본문동일, +11) · `Context.ExitObserver`(본문변경 — B2 행 갱신). `risk-pattern-report.md` 는 다시 생성(차이는 go-panic 행 좌표 `:2046` → `:2057` 하나).

### 0.2 주장 재판정 (§1 · §2 표의 현재 HEAD 판)

| # | 옛 판정 | 현재 판정 | 근거(현재 좌표) |
|---|---|---|---|
| P1 | 참 | 참 | `clearTheSymbol` call `:1653` `LiveOrdersForSymbol`, 순회 B4 `:1668` |
| P2 | 참 | 참 | `classify.go` · `refusal_code.go` · `gateway.go:756` 무변 |
| P3 | 참 | 참 | `submit` B15 `case@1605` → `alertProposalRefused@1610` · `release@1615` |
| P4 | 참 | 참 | `record` B13 `else@1409` → `clearDelay@1410` · `endClearStreak@1411` |
| P5 | 참 | 참 | `alertProposalRefused` L1912-1929, 본문 `Sprintf@1919`; 호출 자리 `:1533`(B3) · `:1610`(B15) |
| P6 | 절반만 참 | **참** | a198 뒤 길 치우기가 `execgw.Issuer{…}.IssueReduction(Kind CANCEL, Side = 주문 방향)`(`clearTheSymbol` call `:1700`)으로 취소 결정을 내고 `o.opts.Submit.Cancel`(`:1715`, 생산 = `c.Gateway` `exitwiring.go:340`, source `engine` · `EngineCancelRefusal` `internal/app/engine/gateway.go:327,329`)로 보냄. 같은 발행으로 원장에 귀속이 없는 BUY 번호를 내면 접수 확정 — `analysis/probe/resume-probe-output.txt` Q1 `state=CONFIRMED broker_cancels=1` |
| D1 | 참 | 참 | `submit` B15 `case@1605` 에서 `out.State` · `out.Reason` 읽힘(무변) |
| D2 | 재료 일부 없음 | 그대로 | `EngineBrokerOrderIDs` 무변(L350-371, PLACE/AMEND · 상태 무관). 참고: a198 D10 · `resolveCancel` 이 CONFIRMED 로 닫은 CANCEL 행은 `broker_order_id` = 대상 번호를 갖지만 kind 필터(PLACE/AMEND) 때문에 귀속 집합에 들어오지 않음 |
| D3 | 재료 일부(r1 #5) | 그대로 | `ScanOrders` 무변(L730-765) |
| D6 | **거짓** | **해소(a198)** | P6 과 같음. 남은 차이: 엔진 게이트웨이 경유 취소에는 사전 단건 조회 1(`precheck.go:58` 2s)이 붙음 — `edit-surface.md` §0 |
| D7 | 참(위험 하나) | **위험이 실측됨** | IN_DOUBT 인 외부 매수 취소는 같은 종목의 손절 매도 · flatten 취소를 `symbol_mutation_in_flight` 로 막고(probe Q2a · Q2e), a198 D10 이 닫지 않으며(Q2b `closed=false`), 재시작 해소만 푼다(Q2d) — `analysis/resume/decision-2-indoubt.md` |
| D8 | census 편집 필요 | 그대로 | `NewContext` 는 여전히 `shapes.json` 고정 함수(렌더에 `recordGateSettings(… cfg.Engine.ExitSourceAgeReport)` 포함). a199 가 이 census 항목을 한 번 갱신한 선례가 생김 |
| D9 | 참(편집 표면) | 참 | 자리 2 좌표만 `:1533` · `:1610` |

### 0.3 새로 뽑은 AST (a198 이후 a197 이 기대게 될 경로, 16)

| AST | 파일 · 좌표 · 분기 | 왜 |
|---|---|---|
| engine--ExitObserver.noteClearCancelFailure | `a198_clear_cancel_log.go` L43-85 · 9 | 취소 실패 줄 대기열(a198 D4) |
| engine--ExitObserver.flushClearCancelFailures | 같은 파일 L89-105 · 4 | 주기 끝 쓰기 |
| engine--ExitObserver.noteClearCancelConfirmed | 같은 파일 L120-122 · 0 | 접수 확정 시 표시 지움 |
| engine--ExitObserver.endExitOrderCycle | `exit_order_log.go` L235-255 · 4 | flush 호출 자리 |
| engine--inDoubtSettle.sweep | `in_doubt_settle.go` L163-207 · 11 | 세션 스윕 — CANCEL 은 B6 `case@184` 로 `closeCancel` 만(조회 대조 대상 아님) |
| engine--inDoubtSettle.closeCancel | `a198_in_doubt_cancel_close.go` L17-41 · 4 | D10 실행자 쪽 |
| engine--onlyEngineCancelsInFlight | `exit_held_proposal.go` L61-87 · 9 | 연속 실패 계수 제외 — IN_DOUBT 는 B4 default `return false`(셈) |
| journal--Journal.ConfirmedCancelOf | `exit_proposal_release.go` L150-174 · 4 | D8(접수 확정 매수 재취소 안 함) |
| journal--Journal.CloseInDoubtCancelOnTargetClose | `a198_cancel_target_close.go` L68-179 · 22 | D10 판정 — 소유 후보 ≠ 1 이면 B6 `if@100` 에서 무변 반환 |
| journal--CancelVerdictForClosedTarget | 같은 파일 L42-51 · 4 | D10 판정 표 |
| execgw--EngineCancelRefusal | `cancel_refusal.go` L26-28 · 0 | 엔진 정책(Definitive) |
| execgw--Gateway.withCancelRefusal | `cancel_refusal.go` L44-112 · 15 | D7 갈래 |
| execgw--settledCancelRefusal | `a198_settled_cancel_refusal.go` L30-52 · 5 | already-* 409 → 확정 거절 |
| execgw--Resolver.resolveCancel | `amend_indoubt.go` L44-158 · 20 | 기동 해소(단건 판독만, 소유 검사 없음) |
| execgw--Gateway.UnsettledOnSymbol | `a094_shared_judgements.go` L21-26 · 0 | 길 치우기 쪽 입구 |
| reconcile--Recovery.confirmAcked | `acked_boot.go` L48-90 · 9 | D9(기동의 ACKED 취소 마침) |

`Journal.Prepare`(`durability.go:373`) · `Resolver.Resolve`(`indoubt.go:229`) · `Recovery.Run`(`recovery.go:247`)은 이번에 AST 를 뽑지 않았다 — 결정 메모는 이 셋의 분기를 근거로 삼지 않고 probe 원문과 위 AST 로만 주장한다(Resolve 의 CANCEL 갈래 `indoubt.go:274-276`, Run 의 해소 호출 `recovery.go:300` 은 손으로 읽음 표시).


- 추출: `go run ./tools/logic-map --file <f> --func <fn>`(바이너리로 한 번 빌드해 반복), HEAD `064f59fd9`, `GOFLAGS=-trimpath`, 2026-11-23.
- 파일 28개(`*.ast.json`). 각 파일의 `source_sha256` 이 그 시점 소스다. 조사 사본(`inv-B/`, HEAD `a9dd53400`)은 참고만 했다.
- probe 1건: `analysis/probe/` — 버리는 worktree(`064f59fd9` detached)에 시험 파일을 넣어 한 번 돌리고 worktree 를 지움. 이 브랜치의 Go 편집 0. 원문 `probe-output.txt`, 원본 `zz_a197_probe_test.go.txt`.

## 1. proposal 표 여섯 줄

| # | 주장 | 판정 | 근거 |
|---|---|---|---|
| P1 | 길 치우기는 원장의 엔진 귀속 주문만 본다(`LiveOrdersForSymbol`, `:1650`) | **참** | `engine--ExitObserver.clearTheSymbol` calls `:1650` `LiveOrdersForSymbol`, B4 range `:1665` 이 그 목록만 돈다. 목록 정의는 `journal--Journal.LiveOrdersForSymbol`(CONFIRMED PLACE/AMEND · 소유 intent 유일 · 종결 증거 없음) |
| P2 | 반대 미체결 거절은 확정 거절로 분류되고 사유가 호출자에게 온다 | **참** | `execgw--classifyMutation` B4 `case@45` → `DispatchRejected`, `ReasonCode = opposite_pending_order_exists`(`refusal_code.go:29`); `journal/dispatch.go:158-160` `Settle(StateFailedConfirmed, …)`; `gateway.go:756` `out.Reason = ReasonCode(res.ReasonCode)` |
| P3 | 확정 거절 뒤 알림 + 해제 → 다음 주기 재발의, 간격·상한 없음 | **참** | `engine--ExitObserver.submit` B15 `default@1602` → `alertProposalRefused@1607` · `release(ProposalRefused)@1612`. 좌표 정정: `1602-1614` → `1602-1615` |
| P4 | 치움 완료가 매 주기 지연 타이머를 지움 | **참** | `engine--ExitObserver.record` B13 `else@1406` → `clearDelay@1407` · `endClearStreak@1408`; B15 default 에는 `noteDelay` 없음 |
| P5 | 알림 1h 1회, 본문 「브로커에 닿기 전에 거부됐다」 | **참**(좌표 정정) | `engine--ExitObserver.alertProposalRefused` `1901-1918`, 본문 `Sprintf@1908`; 재알림 창 `obs.DefaultRemindAfter = time.Hour`(`notifier.go:60`), 유형 critical(`event.go:381`). 호출 자리 **2**(`:1530` Guardian 거절 — 이 자리에서는 문장이 참) |
| P6 | 엔진은 외부 주문 id 로 취소를 낼 수 있다 — flatten 이 이미 그렇게 한다 | **절반만 참 — Manager 결정 필요** | 게이트웨이는 소유를 보지 않는다(참: `execgw--Gateway.Cancel` 분기 0, `execgw--Gateway.prepareRequest` 에 대상 주문 조회 0, `execgw--Gateway.checkSymbolFree` 는 취소면 `:814` 에서 미해소·latch 검사 전에 반환). flatten 은 `execgw.Issuer.IssueReduction(Kind: CANCEL, Side: 주문 방향)` 으로 결정을 만든다. **exit 루프의 결정 발행기(`RiskGuardian.IssueReduction`)로는 취소를 낼 수 없다** — (b) |

## 2. design 주장

| # | 주장 | 판정 | 근거 |
|---|---|---|---|
| D1 | `submit` 에서 `out.State == FAILED_CONFIRMED ∧ out.Reason == ReasonOppositePendingOrder ∧ isProtective` 로 열 수 있다 | **참** | (a) |
| D2 | 대상 = 매수 · 주문 번호 있음 · 원장에 그 번호의 접수 확정 귀속 없음 | **재료 일부 없음 — Manager 결정 필요** | (c): 「주문 번호 → 엔진 귀속」 판독이 `EngineBrokerOrderIDs`(계좌 전체 · PLACE/AMEND · **상태 무관**) 하나뿐. 「접수 확정」으로 좁힌 판독은 없음 |
| D3 | 미지 행 하나·목록 미완 → 취소 0 | **재료 일부 — Manager 결정 필요**(r1 #5 로 정정) | `ScanOrders` 는 쪽 상한 초과·커서 반복·pager 오류면 nil 이지만, `HasNext=true ∧ NextCursor=""` 는 `indoubt.go:754` `!HasNext \|\| next == ""` 로 **성공 반환**(부분 목록 가능) — (d) |
| D4 | 시한 3s 근거 「조회 p99 0.3초 이하(R7)」 | **인용 정정**(값 무변경) | R7 은 OPEN 목록을 재지 않았다 — 단건 GET p99(accounts 569ms, max 4,171ms · holdings 282 · prices 252 …), 문서 결론 「R7 p99 ≤ 0.6s」(`retry-matrix-slo-2026-10-22.md:27,46`). OPEN 목록은 「cursor/limit 무시 · 전량 1요청」(`:47`, M7) |
| D5 | 같은 루프 안에서 동기 실행, 다른 포지션 판정이 그만큼 늦음 | 참(성질) | `record → submit` 이 `ObserveOnce` 의 포지션 순회 안(동기). 기존 `clearTheSymbol` 의 취소도 같은 성질 |
| D6 | 취소는 기존 경로(`Issuer.IssueReduction` → 게이트웨이 취소)로, 원장에 결정·취소 attempt | **거짓 — Manager 결정 필요** | (b) probe: 그 경로의 취소는 `Prepare` 에서 거절, attempt 0 · 브로커 호출 0 |
| D6' | critical 단일 입구가 관측자에 닿아 있다 | **참** | `ExitObserverOptions.Critical CriticalRecorder`(`exitloop.go:212`), 생산 `opts.Critical = c.Notifier`(`exitwiring.go:353`), 알림 꺼짐 판정 `NotificationsEnabled`(`:356`) |
| D7 | 매수 취소는 접수로 치움(정본 규칙), 해제·재발의는 기존대로 | 참(단 위험 하나) | B15 의 `release` 는 그대로 둘 수 있음. **위험**: 새 취소가 IN_DOUBT 로 끝나면 그 attempt 가 `PendingAttempts`(`recovery.go:30-33`, IN_DOUBT 포함)에 남아 다음 주기 `clearTheSymbol` B3 `:1659` 이 「치움 미완료」를 돌려줌 → 사람이 매수를 치운 뒤에도 손절이 IN_DOUBT 해소 때까지 막힘(오늘보다 늦어지는 경로) |
| D8 | 토글 자리는 기존 관례(a163)를 보고 정한다 | **관례를 그대로 따르면 a166 census 고정 함수 편집 — Manager 결정 필요** | (f) |
| D9 | 사유가 반대 미체결이면 본문을 바꾸고, 그 밖의 브로커 거절은 「브로커가 거절했다」 | 참(편집 표면 정정) | `alertProposalRefused` 는 상태·사유를 인자로 받지 않음 → 시그니처 변경, 자리 2 모두 편집. 자리 1(Guardian)과 B15 의 `NOT_DISPATCHED` 는 「브로커에 닿기 전」이 참이라 유지해야 함 |
| D10 | 재제출 간격 없음 | 참 | 무편집 |
| §2 | 치우기 실패는 해제를 막지 않는다 | 가능 | B15 안에서 `release` 앞/뒤 어디에 두든 반환값은 `releaseErr` 그대로 둘 수 있음. 「기존 지연 경보 재료」: B15 에는 `noteDelay` 가 없다 — 쓰려면 새 호출(D10 과 충돌 없음) |

## 3. 재료 질문 (a)~(f)

### (a) `submit` default 갈래의 재료 — **있음**

- `out` 은 `execgw.Outcome`(`gateway.go:301-314`): `State journal.AttemptState` · `Reason ReasonCode` · `Detail` · `AttemptID` · `BrokerOrderID`. default(B15 `:1602`) 안에서 그대로 읽힌다.
- 상수 `execgw.ReasonOppositePendingOrder ReasonCode = "opposite_pending_order_exists"`(`refusal_code.go:29`). 분류: `classifyRefusalCode`(`refusal_code.go:53-78`, 분기 6) 가 본문 JSON 의 `code`/`error.code` 둘만 읽어 `"opposite-pending-order-exists"`(대소문자 무시 전체 일치)면 `refusalCodeDefinitive`; 두 자리가 모순이면 모호 강제. `classifyMutation` B3 switch `:44` → B4 `case@45` 가 `DispatchRejected` + 그 사유.
- 갈래 도달: B12(`ReasonSymbolInFlight`)·B13(`State ∉ {NOT_DISPATCHED, FAILED_CONFIRMED}`) 뒤라 `FAILED_CONFIRMED` 는 B15 에만 온다.
- `isProtective(p) = p.Action == ActionBaselineBreach || p.Action == ActionLadderStop`(`exitloop.go:1475-1477`, 분기 0). 익절(`ActionLadderTakeProfit` 등)과 사다리 부분 청산은 거짓. `isFullExit` 는 `ActionLadderTakeProfit` 를 포함하므로 「전량」과 「보호」는 다른 술어다.

### (b) 외부 주문 번호로 취소를 내는 기존 경로 — **게이트웨이는 있음, exit 루프의 결정 발행기로는 불가(Manager 결정 필요)**

- `clearTheSymbol` 사슬: `o.opts.Issuer.IssueReduction(ReductionIssuance{Intent{Side: SELL, Quantity: order.Quantity}, Account{HeldQuantity: order.Quantity}})`(`:1684`) → `o.opts.Submit.Cancel(CancelRequest{Intent{OrderID, Symbol}, Order: OrderRef{Market, Side, Quantity, Price, Currency}, Decision})`(`:1706`) → `Gateway.Cancel`(`gateway.go:407`) → `Gateway.submit`.
- 생산 `Issuer` = `*execgw.RiskGuardian`(`exitwiring.go:326,339`; `guardian_wiring.go:100`). 그 `IssueReduction`(`riskguardian.go:587-634`)은 Side 가 SELL 이 아니면 거절(B2 `:592`)하고 결정을 **`Kind: journal.KindPlace`** 로 기록한다(`:614`) → `DecisionRequest.build` 가 PLACE 결정에 멱등 키를 붙임(`decision.go:373-375`).
- 게이트웨이 `prepareRequest` 는 결정의 키를 `req.ClientOrderID` 로 옮기고(`gateway.go:1081`), `journal.Prepare` 가 「CANCEL carries no idempotency key」로 거절(`durability.go:267-269`). (그 앞을 통과했다면 `checkPreimage` venue 의 side 비교 `guardian.go:333` 가 BUY 취소를, `checkIdempotencyKey` B2 `:398` 가 모든 취소를 거절했을 것.)
- **probe 실측**(`analysis/probe/probe-output.txt`): RiskGuardian 결정 → `Gateway.Cancel` = BUY·SELL 모두 `state="" reason=invalid_mutation_request broker_cancels=0`, 오류 `recording the intent (nothing was submitted): … CANCEL carries no idempotency key`. 대조군 `execgw.Issuer(Kind CANCEL, Side=주문 방향)` = 둘 다 `CONFIRMED broker_cancels=1`.
- **귀결(기존 결함, a197 밖)**: a094 의 길 치우기는 생산에서 어떤 주문도 취소하지 못한다 — 엔진 귀속 미체결이 있는 종목의 손절은 `cleared=false` 로 무기한 보류(지연 경보 · 연속 실패 경보만 남). 엔진 시험 하네스는 가짜 제출자(`fakeSubmitter`/`a094Submitter`)를 써서 이 결합을 잰 적이 없다(`exitloop_test.go:222`, `a094_exit_clear_test.go:73-110`).
- 요청 필드: `CancelIntent{OrderID, Symbol}` + `OrderRef{Market, Side(BUY|SELL 필수 — prepareRequest B4 `:1025`), Quantity, Price(0 이면 MARKET), Currency(필수 — B7 `:1063`)}` + `GuardianDecision`. 거래일은 게이트웨이가 시계로 계산(`:1020`), 대상 주문 조회 없음. 브로커 행만으로 채움: flatten `parseTarget`(`flatten.go:526-552`)이 `orderId · symbol · side · quantity · price · currency` 를 읽고 시장은 `MarketOf(symbol, currency)` 로 추론(페이로드에 market 없음). 결정은 `ReductionRequest{Kind: CANCEL, Market, Symbol, Side: 주문 방향, MaxQuantity: 주문 수량}`.
- 원장에 남는 것(대조군 경로): decision 행(RISK_REDUCING, ReductionIntent, 키 없음) · intent 행(그 주문 조건) · mutation_attempt(kind CANCEL, `target_order_id` = 브로커 번호, 상태 CONFIRMED/FAILED_CONFIRMED/IN_DOUBT). 엔진 취소의 409 already-processing 은 `EngineCancelRefusal`(확정 종결) 정책.
- preflight 의 「원장에 그 주문이 있어야 한다」 요구: **없음**(`Cancel` 은 `plan.preflight` 를 두지 않음, `checkSymbolFree` 는 같은 종목 미종결만 봄).
- 수동 flatten 사슬: `enumerate → ScanOrders(OPEN, 종목 필터 없음) → parseTarget` → `settleStep` → `decisionFor(execgw.Issuer{Journal, Clock, AccountRef, TTL, NewID})` → `Gateway.Cancel`. 차이는 **결정 발행기 하나**다.

### (c) 「엔진 귀속」 판독 — **접수 확정 한정 판독은 없음(Manager 결정 필요)**

- `LiveOrdersForSymbol`(`fills.go:1854-1919`): `state = CONFIRMED ∧ kind ∈ {PLACE, AMEND} ∧ broker_order_id <> ''` · 같은 계좌·시장·종목 · 소유 intent 정확히 1 · **종결 증거 없음** · lineage 로 현재 번호. 종결된 주문과 **소유가 모호한 주문은 빠진다** → 귀속 판정에 쓰면 그런 엔진 주문이 「외부」로 오판된다. 귀속 판정에 쓰면 안 됨.
- `EngineBrokerOrderIDs`(`daily_buy.go:350-371`, a136): 계좌의 `kind ∈ {PLACE, AMEND} ∧ broker_order_id <> ''` attempt 의 번호 집합, **상태 무관**(IN_DOUBT·FAILED 포함 — 귀속을 넓게 잡는 보수 방향). 시장·종목 범위 없음, lineage 해석 없음(AMEND 자식 번호는 attempt 에 있으므로 포함).
- 번호 단건 조회 함수(`mutation_attempts.broker_order_id = ?`)는 없음. `orderlineage`·fills 의 귀속 함수(`guardTrackedFillIdentity` 등)는 체결 귀속용이다.
- 번호가 아직 안 적힌 엔진 매수: RECORDED · DISPATCH_STARTED · ACKED · IN_DOUBT 는 `PendingAttempts`(`recovery.go:30-33`)에 있음 → `clearTheSymbol` B3 `:1659` 이 `clearResult{cleared:false}` 를 돌려줌 → `record` B12 `:1399` 에서 무장 안 함 → `submit` 에 안 닿음 → 새 단계도 안 열림. **단 UNRESOLVED_IN_DOUBT(park) 매수는 `PendingAttempts` 에 없다**(상태 목록 4개) — 번호가 없는 park 매수가 브로커에 살아 있으면 `EngineBrokerOrderIDs` 에도 없어서 「외부」로 분류된다. → Manager 결정 필요(같은 종목 park attempt 가 있으면 새 단계를 열지 않는 가드 = 새 가드).

### (d) 브로커 미체결 목록 판독 — **판독기는 있음, exit 관측 루프에는 주입 0**

| 판독기 | 시그니처 | 종목 필터 | 쪽 넘김 · 「다음 쪽」 | 실패 | 재시도 · 시한 |
|---|---|---|---|---|---|
| `execgw.OfficialOrders.OrdersPageRaw` | `(ctx, OrderQuery{Status, Symbol, From, To, Limit}, cursor) (OrderPage{Orders []json.RawMessage, NextCursor, HasNext}, error)` | `Symbol` → 쿼리 `symbol`(`official/orders_raw.go:67-68`) — 브로커가 OPEN 에서 지키는지는 **미측정** | 한 쪽 | 공식 클라이언트 오류 그대로 | 없음(「no retry logic」 `orders_source.go:8`), ctx 만 |
| `execgw.ScanOrders(ctx, pager, q, maxPages)` | `[]json.RawMessage, error` | q 그대로 | 끝까지 걸음, `HasNext ∧ NextCursor` 로 다음 쪽; 상한 초과 `ErrTooManyPages`, 커서 반복 `ErrCursorLoop`; **단 `HasNext=true` 인데 커서가 비면 거기까지를 성공으로 반환**(`:754`, r1 #5) | 첫 오류에서 nil | 없음, ctx 만(maxPages ≤ 0 → 50) |
| flatten `enumerate`/`parseTarget` | 비공개(패키지 `flatten`) | 없음(계좌 전체 OPEN) | `ScanOrders` | 해석 실패 행 하나 → 전체 중단 | — |
| `reconcile.Collector` | `Collect` 가 `ScanOrders` | 계좌 전체 | 같음 | — | 대사 루프 소유 |
| `filldetect.Detector.collect` | `ScanOrders(OPEN)` 3s 주기 | 계좌 전체 | 같음 | 주기 실패 | 스냅숏 공표 없음(a094 D−2.4 ③) |

- 행 필드(OPEN 페이로드): `orderId · symbol · side · quantity · price · currency · status · execution.filledQuantity · canceledAt`(flatten `openOrderPayload`, brokerstate `officialOrderPayload`, execgw `officialOrderFacts`). **시장 필드 없음**(통화·종목 모양으로 추론). 미체결 수량은 `quantity − execution.filledQuantity` 로만 얻음.
- exit 관측 루프의 브로커 판독 주입: **0** — `ExitObserverOptions`(`exitloop.go:171-254`)에 주문 판독 필드 없음. 주입 자리: `Context.ExitObserver`(`exitwiring.go:319-365`)에 `execgw.OfficialOrders{Client: c.Official}` 한 줄(선례 `cmd/tossctl/engine.go:770`, `runtime_wiring.go:133`). 이 함수·파일은 a166 census 고정 목록에 **없다**.
- a166 예산 · 우선순위: 판독기를 통한 공식 호출은 계좌 공통 예산(350ms 최소 간격, `accountbudget/budget_linux.go:27`)을 탄다. exit 쪽 선례는 보호 청산이면 `accountbudget.Critical`(`exitloop.go:872` 가격 · `:1393` 치우기 · `:1554` 제출). 읽기를 `Retrier.Query(QueryOpenOrders, …)` 로 하면 open_orders freshness 를 찍어 진입 게이트 입력이 바뀐다 — 가장 가까운 선례는 a191 종결 목록 조회(직접 읽기 · 재시도 0 · 2s 취소 시한 · freshness 영향 없음, `filldetect/closed_lookup.go:81,210`; retry-matrix §1 a191 행).

### (e) 알림 — **있음**

- `alertProposalRefused` 원문: Type `EventExitProposalRefused`(`"exit.proposal_refused"`, critical) · Key `exit.proposal_refused|<position_id>|<action>|<level>` · Title `<label> 청산 주문이 제출되지 않았다` · Body `%s (단계 %s)가 브로커에 닿기 전에 거부됐다.\n원인: %s` · Fields `symbol · position_id · action · level · detail`. 입구 `o.alert` → `Alerts.Notify`(생산 `obs.RecordOnly`, critical 이면 `recordCritical(e, remindAfter=1h)`). 호출 자리 2.
- 「취소 접수마다 critical 1」의 입구: `o.opts.Critical.RecordCritical(ctx, e, remindAfter)`(`CriticalRecorder`, `exit_held_proposal.go:27-29`) — a094 의 창 0 critical 이 이미 쓰는 단일 입구. 생산 배선됨(`exitwiring.go:353`). 알림 꺼짐 판정 재료 `NotificationsEnabled`(`:356`).

### (f) exit 운영 토글 선례(a163 `exit_source_age_report`)

- 자리: `config.Engine.ExitSourceAgeReport bool \`json:"exit_source_age_report,omitempty"\``(`config/engine.go:61-64`) · raw `*bool`(`:275`) · `mergeEngine` 에서 **automation_gate 조기 반환 앞**에 복사(`:349-351`). 기본 false(키 없음 · 명시 false). 검증 없음(bool).
- audit: `recordGateSettings(log, gate, adoption, notifications, attestationFile, sourceAgeReport bool)`(`interlock.go:739-813`)가 `audit.ActionExitSourceAgeReport` 행을 표 끝에 덧붙여 기동마다 `RecordChange`(값이 바뀐 기동만 새 기록). **호출자는 `NewContext`(`engine.go:486-487`) — a166 census `shapes.json` 의 고정 함수**(`internal/app/engine/engine.go||NewContext`, 렌더 본문에 `recordGateSettings(…, cfg.Engine.ExitSourceAgeReport)` 포함). 같은 방식으로 새 토글을 audit 하려면 `NewContext` 의 인자 줄을 고쳐야 하고 census 가 깨진다.
- 콘솔 노출: 없음(`internal/console` grep 0).
- 엔진이 읽는 시점: **기동 때만** — `Context.ExitObserver` 가 `c.Config.Engine.ExitSourceAgeReport` 로 호출자 값과 무관하게 덮음(`exitwiring.go:358`), 주기 재읽기 없음.
