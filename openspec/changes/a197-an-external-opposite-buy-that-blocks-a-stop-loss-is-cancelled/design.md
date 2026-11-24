# a197 Design

> **재개 개정 2026-11-25**(Manager 결정 2026-11-25 — `review.md` 「Manager 결정 (2026-11-25)」). 2026-11-23 판의 D1~D10 을 §7 의 결정과 선행 2 결정으로 고쳐 썼고, r1 처분 11 건을 어디에 반영했는지는 §7 표에 있다.
> 분기를 근거로 삼는 문장의 AST 는 `analysis/cited-ast/`(현재 HEAD 기준, `README.md` §0) · `analysis/function-logic/` 번호다. probe 원문은 `analysis/probe/resume*-`.

## 0. a094 D−2.4 의 선행 조건 넷을 어떻게 갖는가

| 선행 조건(a094 design D−2.4) | 이 change 의 답 |
|---|---|
| ① 외부 주문의 귀속 규칙 — 무엇이 「이 포지션의 보호를 막는 주문」인가, 불완전·미지 주문은 거절 | D2·D3 |
| ② 다른 포지션·사람의 보호 매도는 취소하지 않는다 | D2(매수만) |
| ③ 공표된 OPEN 스냅숏 계약 — 신선도, 「비어 있음」과 「읽지 못함」의 구분, 실패 처리, 비차단 | D3·D4·D5. 공표된 OPEN 스냅숏은 **없다**(체결 감지는 OPEN 행을 세기만 하고 보관하지 않음 — `filldetect/detect.go:376-380` `cycle.OpenOrders = len(raws)`). 그래서 이 단계가 그 자리에서 종목 필터로 읽고, 쓰고, 버린다 |
| ④ 취소 사실의 감사 기록 | D6 |

## 1. 결정

| # | 결정 | 근거 |
|---|---|---|
| D1 | **여는 조건(전부)**: 토글 켜짐 ∧ 발의가 보호 손절(`isProtective`, AST `engine--isProtective` `exitloop.go:1478-1480`) ∧ 제출 결과 `FAILED_CONFIRMED` ∧ 사유 `ReasonOppositePendingOrder`(`submit` B15 `case@1605`) ∧ 그 종목에 **판정 불능으로 park 된 엔진 발주·정정(PLACE/AMEND · UNRESOLVED_IN_DOUBT)** attempt 가 없음(r1 #2). 판정 경로(B15)에서는 「열 조건이 섰다」를 관측기 메모리 대기열에 넣기만 하고 실행은 D5. 그 밖의 거절 사유 · 익절 · 사다리 부분 청산에서는 열리지 않는다. 청산 전용 기동(a199)에서도 열린다(r1 #9 공시 — `Context.ExitObserver` B2 `if@323 RunsExitSide()`) | 브로커가 「반대 미체결이 있다」고 말한 뒤에만 읽는다 — 손절의 첫 제출 앞에 브로커 왕복을 더하지 않는다. 같은 종목에 **미종결**(RECORDED · DISPATCH_STARTED · ACKED · IN_DOUBT) attempt 가 있으면 이 단계는 애초에 열릴 수 없다: `clearTheSymbol` B3 `if@1662`(입력 `UnsettledOnSymbol` → `PendingAttempts` `recovery.go:30-33`)가 치움 미완료를 돌려주고, `record` B12 `if@1402` 가 `orderable=false` 로 `submit`(call `:1456`)에 닿지 않게 한다. **park** 는 `PendingAttempts` 밖이라 이 가드에 안 걸리므로(r1 #2 의 남은 절반) 새 가드가 필요하다 — 번호 없는 park 엔진 매수가 브로커에 살아 있으면 D2 의 귀속 판독에도 없어 「외부」로 오분류된다. 새 가드의 대상은 PLACE/AMEND park 뿐이다(D11 이 만드는 이 단계 자신의 CANCEL park 는 엔진 주문이 아님) |
| D2 | **대상**: 같은 계좌·시장·종목의 미체결 행 가운데 방향이 매수이고, 주문 번호가 있고, 그 번호가 엔진 귀속 집합에 **없는** 것. 귀속 집합 = `EngineBrokerOrderIDs`(AST `journal--Journal.EngineBrokerOrderIDs`, `daily_buy.go:350-371` — 계좌 전체 · PLACE/AMEND · **상태 무관**: IN_DOUBT · FAILED 도 귀속으로 봄 = 넓게, 보수 방향). 행의 종목 · 방향은 **행 값으로 다시 대조**하고(조회의 종목 필터를 믿지 않음 — 브로커가 OPEN 에서 필터를 지키는지 미측정), 주문 번호는 행의 바이트 그대로 취소 요청에 싣는다(정규화 금지)(r1 #6). 매도는 대상이 아니다. 한 주기에 취소는 **1 건**(D4) — 대상이 여럿이면 행 순서의 첫 것 | 브로커 거절 사유는 「반대 방향 미체결의 존재」다. 가격 교차 · 수량 조건은 근거가 없으므로 매수 전부가 후보다. 번호 바이트 보존은 공식 번호가 불투명 식별자라는 저장소 규칙(`resolution.go:40-47` 주석)과 같다 |
| D3 | **하나라도 모르면 전부 안 한다**: 목록의 어느 행이든 방향 · 주문 번호 · 종목을 읽을 수 없으면, 목록을 끝까지 읽지 못했으면(오류 · 시한 초과 · 「다음 쪽 있음」인데 커서가 비어 있음 · 쪽 상한 초과) 그 주기에는 아무것도 취소하지 않는다. `ScanOrders` 는 「다음 쪽 있음 ∧ 커서 없음」을 성공으로 돌려주므로(AST `execgw--ScanOrders`, `indoubt.go:754` `!HasNext \|\| next == ""`) 그 갈래를 그대로 쓰지 않는다(r1 #5). 대상 0 건(매수가 없음)도 취소 0 이다. 「읽었고 대상 없음」과 「읽지 못함」은 다른 결과다 | 미지의 주문 위에서 일부만 치우면 무엇이 남았는지 모른다 |
| D4 | **조회 빈도 · 상한 · 시한**: (가) 조회는 D5 가 그 주기에 처리하는 포지션 하나에 대해 종목 필터 OPEN 1 회 — 공표된 스냅숏이 없어서(§0 ③) 더 작은 안이 없다. OPEN 목록은 cursor/limit 를 무시하고 전량 1 요청이다(`verify-execution-capability/analysis/retry-matrix-slo-2026-10-22.md:47`, M7) — 쪽은 1. 재시도 0. (나) **주기당 처리 포지션 1 · 취소 1 건**(r1 #4 ②). 같은 주기에 둘 이상의 포지션이 열 조건을 세우면 가장 오래 처리되지 않은 포지션 하나(메모리 시각)를 고르고, 나머지는 다음 거절 주기에 다시 선다. (다) 조회에 시한 3 초(새 상수, 넘기면 「읽지 못함」). 단계 전체에도 ctx 시한을 걸지만 **hard 상한은 걸 수 없다** — 토큰 잠금(`official/token.go:63` `m.mu.Lock()`)이 ctx 로 끊기지 않는다. (라) 한 주기에 더해지는 시간의 **조건부 추정**(취소 1 건, r2 #8 로 정정): 요청 셋의 시한 합 = 조회 3 s + 사전 단건 조회 `precheckTimeout` 2 s(`precheck.go:58`) + 취소 POST `defaultTimeout` 15 s(`official/client.go:22`) = 20 s, 여기에 계좌 예산의 **최소** 간격 350 ms × 3(`accountbudget/budget_linux.go:27` — 대기 상한이 아님)을 더한 **21.05 s 는 하한 쪽 추정**이다. 이 값 밖: 요청 전 토큰 교환(`client.go:336` `c.tm.token` → `token.go:132` `exchange`, 별도 HTTP), 401 응답 뒤 재교환 · 재전송(최대 2 회, `client.go` 의 401 루프 — 코드 확인), 계좌 예산 경합 대기, 토큰 잠금 대기(`token.go:63`, 상한 없음). 단계 전체 ctx 의 값은 구현 단계에서 정한다(미정). 따라서 **단계 시간에는 코드상 상한이 없고**, 유계 요소만의 합도 확정값이 아니다(§4-16 에서 인증 갱신 · 예산 경합을 넣어 잰다) | (가) 「거절 주기마다 1 GET」은 그 포지션의 손절이 반대 미체결로 계속 거절되는 동안 — 사람이 매수를 치울 때까지, D7 로 그 번호에는 다시 취소하지 않아도 — 주기마다 1 GET 이 계속된다는 뜻이다: 조회 몫만 계좌 예산(350 ms 간격 ≈ 2.86 요청/s)의 약 7%(사전 조회 · 취소 · 토큰 교환은 따로). 더 줄이려면 포지션별 「다음 조회 시각」 같은 새 상태가 필요하고 D10(새 간격 없음)과 같은 축의 결정이라 넣지 않았다 — r2 가 이 선택을 보도록 근거를 남긴다. (나) a198 §2-4 · r1 #4: 취소 건수가 늘면 그만큼 루프가 선다 |
| D5 | **실행 자리 = 관측 주기의 끝(인라인, 별도 실행자 없음)**: `endExitOrderCycle`(`exit_order_log.go:235-255`, FLM `function-logic/internal-app-engine--exitobserver.endexitordercycle`)의 **첫 문장**에서 이 단계를 부른다. `endExitOrderCycle` 은 `ObserveOnce` 의 첫 defer(AST `engine--ExitObserver.ObserveOnce` `defer@482`)라 그 주기의 모든 `judge`(call `:542`) 뒤에 돈다 — 같은 주기의 뒤 포지션이 시세 만료(`quoteUsable`, `ObserveOnce` B7 `if@534`)로 판정을 건너뛰는 일이 없다(r1 #4 ①). B1 `if@236 ExitOrderLog == nil` 의 조기 반환 **앞**이어야 한다(뒤면 전용 로거 없는 조립에서 단계가 조용히 안 돈다). 같은 ObserveOnce 의 다른 defer `flushStrategyProtectionAlerts`(`defer@485`)는 이 단계보다 **먼저** 돈다(defer 역순). **대기열의 낡음(r2 #3)**: 판정 경로에서 넣은 「열 조건」과 주기 끝 실행 사이에 그 포지션이 청산 · 세대 변경될 수 있는데(체결 감지는 다른 goroutine), 이 설계는 실행 직전에 포지션의 존재 · 세대 · 보호 필요성을 다시 확인하지 않는다 — 재검증 가드는 새 가드라 넣지 않았고 Manager 결정 M-3(§7). **비차단의 한계(공시)**: `Run`(AST `engine--ExitObserver.Run`, `exitloop.go:417-428`)은 `ObserveOnce`(call `:422`, defer 포함)가 반환한 뒤에야 `Sleep(Interval)`(`:424`, 5 s)을 하므로 단계 시간 T 만큼 **다음 주기 시작이 밀린다** — 5 s 간격이 5 s + T 가 되고, T 에는 코드상 상한이 없다(D4 라 — 요청 시한 합만으로 약 21 s). 같은 주기의 지연을 다음 주기 모든 포지션의 지연으로 옮기는 것이다 | 단계가 열리는 것은 손절이 이미 거절된 포지션이 있을 때뿐이고 주기당 1 건이다. **보조 실행자 안은 기각**: exit 루프 밖 goroutine(선례 `inDoubtSettle` `in_doubt_settle.go:137`)이면 루프 지연은 0 이지만, 그 취소가 전송 중인 동안 다음 주기의 손절이 같은 종목 미종결(`checkSymbolFree` B2 `if@808`)로 거절되므로 게이트웨이의 종목 단위 직렬화와 취소 → 제출 순서를 다시 설계해야 한다. 기존 길 치우기의 취소(`clearTheSymbol` call `:1715`)도 판정 경로 안에서 같은 비용을 치르는 성질이다(a198 §2-4) — KISS |
| D6 | **발행 · 감사**: 취소 결정은 a198 의 길 치우기와 같은 발행 — `execgw.Issuer{Journal, Clock: o.clk, AccountRef}.IssueReduction(Kind CANCEL, Side BUY, MaxQuantity = 행 수량 − 체결 수량)`(선례 `clearTheSymbol` call `:1700`) → `o.opts.Submit.Cancel`(생산 `c.Gateway` `exitwiring.go:340`, source `engine`, `EngineCancelRefusal` `internal/app/engine/gateway.go:327,329`). 원장에 결정 · intent · CANCEL attempt 가 남는다. 브로커 어댑터가 취소 전에 사전 단건 조회 1 을 한다(`precheck.go:58·99`, advisory). 외부 번호도 받는다(probe `resume-probe-output.txt` Q1 `CONFIRMED broker_cancels=1`; 대상 소유 판독 없음 — AST `journal--Journal.Prepare` 호출 목록, `execgw--Gateway.Cancel` 분기 0). **운영자 알림(r1 #7 · #11)**: 취소를 **내기 전에** critical 「손절을 막는 외부 매수 주문의 취소를 시도한다(포지션, 종목)」 1 건(알림기의 critical 단일 입구 `o.opts.Critical`, `exitwiring.go:353`; key = 타입 · 포지션 · 주문 번호, 재알림 창 0). 결과가 접수가 아니면(확정 거절 · park · 기록 뒤 실패) 두 번째 critical 「치워지지 않았다 — 사람이 할 일」 1 건(같은 key 단위, 다른 타입), 본문에 운영자 절차(§5). 알림 제목 · 본문 · 필드에 계좌 정보와 주문 번호를 싣지 않는다(종목 · 포지션 · attempt id 만). 알림이 꺼진 엔진은 일반 등급 기록. **critical 기록 실패(r2 #6, 기존 안전 장치)**: 알림기의 critical 기록이 실패하면 진입 게이트에 계정 전체 `ReasonAlertUndelivered` 래치가 걸리고 운영 모드 승격이 시도된다(`obs/record_only.go:137-157`) — 이 단계의 알림도 같다. 그 래치 · 승격은 attempt 해동으로 풀리지 않는다 | 사람의 주문을 엔진이 지웠다는 사실은 운영자가 반드시 알아야 한다. 취소 전에 기록하면 접수와 기록 사이의 충돌로 알림이 사라지지 않는다(r1 #7). 이 기록은 주기 끝(D5)에서 하므로 a198 핀 `a198_clear_cancel_log_order_test.go:140`(판정 경로 안 쓰기 금지)과 충돌하지 않는다 |
| D7 | **같은 (포지션, 주문 번호)에는 취소를 한 번만 낸다**(Manager 결정 2). 결과가 접수 · 확정 거절 · 모호(D11 park) 어느 것이든 그 번호에 다시 내지 않는다. 치워지지 않았으면 D6 의 두 번째 critical 이 사람에게 넘긴다. **「이미 냈다」 표시 = 원장의 CANCEL attempt 행**: kind CANCEL ∧ 취소 intent 의 source `engine` ∧ `target_order_id` = 그 번호(바이트) ∧ 같은 계좌 · 시장 · 종목 · 취소 intent 의 거래일, **상태 무관**. 그런 행이 하나라도 있으면 낸 것으로 본다. 판독은 새 함수(새 파일) 하나가 필요하다 — 기존 판독은 상태나 범위가 맞지 않는다: `ConfirmedCancelOf`(AST `journal--Journal.ConfirmedCancelOf`, CONFIRMED 만), `FailedConfirmedCancelsForSymbol`(FAILED_CONFIRMED 만, `cancel_refusal_latch.go:58`), `CloseInDoubtCancelOnTargetClose`(id 한정). 스키마 · 새 상태 무변(열 `kind` · `state` · `target_order_id` 는 schemaV1 `schema.go:224-`, `source` · `trading_day` 는 intents). 키를 포지션 대신 종목 범위로 잡는 것은 (포지션, 번호)보다 넓다(같은 번호에 더 적게 냄 = 보수). **거래일의 귀결(r2 #2, 사실 정정)**: 취소 intent 의 거래일은 대상 주문의 날이 아니라 취소 시각으로 게이트웨이가 정한다(`gateway.go:1020` `market.TradingDay(g.clk.Now())`) — 그래서 이 표시는 「주문 생애당 한 번」이 아니라 **「취소한 거래일당 한 번」**이다. 같은 OPEN 주문이 시장의 날짜 경계(KR · US 각 시장 달력)를 넘어 살아 있으면 다음 날 다시 대상이 되고, D6 알림 key 에는 거래일이 없어 그날의 「시도」 알림이 생략될 수 있다. 실제 브로커에서 그런 주문이 날짜를 넘는지는 미측정. 범위(생애당/거래일당)는 Manager 결정 M-1(§7). 엔진 귀속 번호는 D2 가 먼저 빼므로 길 치우기의 CANCEL 행이 이 표시로 읽힐 일이 없다. **경합 · 재시작**: 이 단계를 부르는 것은 exit 관측 goroutine 하나뿐이고(D5), 엔진은 프로세스 하나(엔진 lock)라 표시 판독과 `Prepare` 사이의 동시 발행이 없다. 표시는 원장에 있으므로 재시작 뒤에도 남는다. `Prepare` 뒤(RECORDED) 전송 전에 죽으면 재시작 규칙이 NOT_DISPATCHED 로 닫고 표시는 「냈음」으로 남아 다시 내지 않는다(보수 — D6 의 첫 critical 은 이미 기록됨). 게이트웨이가 **기록 전에** 거절하면(같은 종목 미종결 `symbol_mutation_in_flight` · 결정 기록 실패 — attempt 행 없음) 낸 것이 아니므로 다음 거절 주기에 다시 설 수 있다. flatten(source `flatten-all`)이 낸 취소는 표시가 아니다 | 2026-11-23 판의 「취소가 거절 · 모호면 그 주문은 남고, 다음 거절 주기가 다시 연다」는 **삭제**: 모호(IN_DOUBT)면 같은 종목 미종결 가드가 다음 주기의 손절 제출을 막아 거절 주기가 오지 않고(probe `resume-probe-output.txt` Q2a — 그래서 D11), 확정 거절이 반복되면 주기마다 조회 + 취소가 더해져 다른 포지션과 호출 예산을 깎는다. 매수 취소 접수는 치움이다(정본 exit-policy `spec.md:919` 의 같은 귀결) |
| D8 | **토글**: 설정 칸 `engine.cancel_external_opposite_buy`(bool, 기본 false — 키 없음 · 명시 false = 꺼짐). 자리는 a163 선례대로 `config.Engine` 칸 + raw `*bool` + `mergeEngine` 의 B4 `if@349` 뒤 · B5 `if@352` 조기 반환 앞 복사(FLM `function-logic/internal-config--mergeengine`). 엔진은 기동 때만 읽는다 — `Context.ExitObserver` 가 호출자 값과 무관하게 덮어씀(선례 `exitwiring.go:358`). **audit 자리(r1 #10)**: `NewContext`(a166 census 고정)를 고치지 않고 `Context.ExitObserver` 안에서 `c.Audit`(`engine.go:227`)의 `RecordChange`(값이 바뀐 기동만 기록, `audit.go:225`)로 새 동작 상수를 기록한다 — 편집 함수는 `Context.ExitObserver`(census 밖, FLM `function-logic/internal-app-engine--context.exitobserver`). audit 기록이 실패했을 때의 처리는 **열린 항목 O-1**(§7). 꺼짐 = 브로커 조회 0 · 취소 0 · park 0 · 단계 알림 0(D9 알림 문장만 토글과 무관하게 바뀜). 롤백 = 토글 끄고 엔진 재시작(r1 #9) | 안전 불변식 2·5·7. 운영 flip 은 사람이 한다. 기동 때 읽으므로 즉시 롤백은 재시작을 요한다 |
| D9 | **알림 문장(토글 무관)**: `alertProposalRefused`(AST `exitloop.go:1912-1929`, 본문 `Sprintf@1919`)가 결과 상태 · 사유를 받아, 사유가 `ReasonOppositePendingOrder` 이면 본문을 「이 종목에 매수 미체결 주문이 있어 브로커가 손절을 거절했다. 앱에서 그 주문을 취소해야 손절이 나간다」로, 브로커가 거절한 그 밖의 경우(`FAILED_CONFIRMED`)에는 「브로커가 거절했다」로 말한다. key · 등급 · 재알림 창은 그대로. 호출 자리 2(`:1533` B3 Guardian 거절 — 「브로커에 닿기 전」이 참이라 유지, `:1610` B15 default) 모두 편집 | 오늘 문장 「브로커에 닿기 전에 거부됐다」는 브로커 거절에 거짓이다. key 를 바꾸지 않으므로 알림 수는 그대로다 |
| D10 | **재제출 간격은 넣지 않는다** | a094 Q4-6 「새 간격 없음(즉시성)」 유지. 간격은 사람이 주문을 치운 뒤의 손절을 늦춘다 |
| D11 | **모호(IN_DOUBT)면 그 자리에서 park, 계정 래치 없음**(Manager 결정 1 — 선행 2). 이 단계의 `Submit.Cancel` 이 IN_DOUBT 를 돌려주면 같은 단계 안에서 `Journal.Resume(attempt)` → `Attempt.ResolveUnresolved`(AST `journal--Attempt.ResolveUnresolved`, `resolution.go:77-84` — `from: [IN_DOUBT]` 만, 기본 사유 `ReasonResolutionExhausted`, 예약 해제 없음 `reservation_release.go:119-126`, 전이 행 `durability.go:650`)로 UNRESOLVED_IN_DOUBT 로 넘긴다. `Resolver.park`(AST `execgw--Resolver.park` B2 `if@379` → `Gate.Block` call `:380`)의 계정 래치는 **걸지 않는다** — 새 전이 · 스키마 없음, 기존 공개 API. park 뒤 효과(probe `resume2-*`): 같은 종목의 손절 · 길 치우기 · flatten 취소 · 매도는 통과(`checkSymbolFree` B3 `if@814 !raisesExposure` 가 미해소 대조 call `:817` 앞에서 반환; 엔진 조립 위 관측 주기 — park 전 `broker_sells=0`, park 뒤 `broker_sells=1`). 대사 지역 상태 · 대사 추적기 · `reconcile-resolve` · 편입은 park 된 CANCEL 을 읽지 않는다(`LocalStateFromJournal` `compare.go:193-238` → `TrackedFillOrders` `fills.go:1559` kind PLACE/AMEND). **거부하게 되는 정상 입력: 운영자 해동 전까지 그 종목(같은 계좌 · 시장)의 신규 진입**(`checkSymbolFree` B4~B7 — `UnresolvedAttempts` 를 매 요청 원장에서 읽으므로 재시작으로 안 풀림; probe: 같은 종목 BUY `unresolved_in_doubt`, 다른 종목 BUY `CONFIRMED`). 기동 해소는 park 를 다시 보지 않는다(`Recovery.Run` 입력 `PendingAttempts` call `:266`). 해제는 운영자 해동(§5) 하나. park 와 함께 D6 의 두 번째 critical(절차 포함). **park 불가 잔여(공시, 새 가드 없음)**: 원장 쓰기 실패로 취소가 DISPATCH_STARTED 나 ACKED 에 남으면 park 할 수 없고(전이가 IN_DOUBT 에서만; 정본은 진행 중 전송의 세션 종결을 금지) 그 종목의 손절 · flatten 이 재시작까지 막힌다(재시작 규칙 → IN_DOUBT → 기동 `resolveCancel`, ACKED 는 a198 D9). `ResolveUnresolved` 자체가 원장 오류로 실패해도 IN_DOUBT 로 남아 같다. **중단 창(r2 #1, 공시)**: ① IN_DOUBT 기록 뒤 park 전에 프로세스가 끝나면 기동 복구가 그 attempt 를 일반 해소기로 보내(`recovery.go:266` → `:300`) 단건 판독으로 닫거나, 닫지 못하면 `Resolver.park` 가 **계정 전체 래치**를 건다(`indoubt.go:376-380`) — D11 의 종목 범위가 이 창에서는 보장되지 않는다. ② park 기록 뒤 두 번째 critical 전에 끝나면 park 는 `PendingAttempts` 밖이라 다시 다뤄지지 않고 결과 알림이 남지 않는다(첫 critical 「시도」만 남음). 이 두 창의 복구 정책은 Manager 결정 M-2(§7). 이 경우들 모두 지연 경보(30 초)와 연속 실패 경보(IN_DOUBT 는 계수에 듦 — `onlyEngineCancelsInFlight` default `:73`)가 나간다 | (가) 손절 · 청산 칸에서 오늘보다 늦어지는 칸 0(`analysis/resume/decision-2-addendum.md` E4). 주문 경로 판정 본체 `unsettledFor` 와 정본 order-execution 「심볼당 in-flight 1개」 SHALL 은 무변. (나) 계정 래치를 걸지 않는 이유: 오늘(토글 꺼짐)도 같은 불확실성(외부 매수가 살아 있는지 모름)이 래치 없이 존재한다. 계정 래치는 엔진 프로세스 종료로만 풀리고(`Clear(ReasonUnresolvedInDoubt)` 호출자 0, probe: 해동 뒤에도 `still_latched=true`), 엔진 정지는 손절 부재 창을 만든다(보호 UNWIRED). 종목 범위 차단은 정본의 park 기본 의미(「해당 심볼 신규 진입 차단, 운영자 해소만」)와 같다. (다) 기각: (C) 판정 본체 면제 · (A) 세션 중 단건 판독 해소 · (B) 가드 유지 + critical(`decision-2-indoubt.md` · `decision-2-addendum.md`) |

## 2. 순서 (한 주기)

```text
[판정 경로 — 포지션마다]
손절 제출 → 브로커 409/422 code=opposite-pending-order-exists → FAILED_CONFIRMED, Reason=opposite_pending_order_exists
  ├ (기존) alertProposalRefused — D9 문장
  ├ (신규) 토글 켜짐 ∧ 보호 손절 ∧ 사유 일치면 「열 조건」을 메모리 대기열에 넣기만 함(브로커 호출 0)
  └ (기존) 발의 해제(ProposalRefused) → 다음 관측 주기가 손절 재발의

[주기 끝 — endExitOrderCycle 첫 문장, 모든 judge 뒤]
대기열에서 포지션 1 개(가장 오래 처리 안 된 것) — 나머지는 버림(다음 거절 주기에 다시 섬)
  ├ 같은 종목 PLACE/AMEND park 있음 → 아무것도 안 함(D1)
  ├ 미체결 목록 OPEN 1 회(종목 필터, 시한 3s) ─ 읽지 못함/미지 행 → 취소 0, 로그
  │                                        └ 읽음 → 행 값 재대조 → 매수 ∧ 번호 있음 ∧ 엔진 귀속 아님 ∧ 「이미 냈다」 아님(D7) 의 첫 행
  │                                             ├ 없음 → 취소 0, 로그
  │                                             └ 있음 → critical 「취소를 시도한다」 → 취소 결정 발행 → Submit.Cancel(사전 단건 조회 포함)
  │                                                    ├ CONFIRMED → 끝(다음 주기 손절)
  │                                                    ├ FAILED_CONFIRMED · 기록 뒤 실패 → critical 「치워지지 않았다」(절차)
  │                                                    └ IN_DOUBT → ResolveUnresolved(park, 래치 없음) → critical 「치워지지 않았다」(절차)
  └ (기존) a129 줄 쓰기 · a198 실패 줄 · 래치 정리 · 체결 따라가기
→ Run: reportCycle → cycleComplete → Sleep(5s)   ← 단계 시간 T 만큼 다음 주기 시작이 밀림(D5)
```

치우기의 실패는 발의 해제를 막지 않는다(해제는 판정 경로의 기존 규칙). 단계 안의 오류는 삼키지 않고 로그(주기 끝이므로 a198 D4 의 쓰기 규칙과 같은 자리)와 D6 의 critical 로 남긴다.

## 3. 거절·침묵하는 정상 입력

- 토글 꺼짐: 조회 0 · 취소 0 · park 0 · 단계 알림 0. 달라지는 것은 D9 알림 문장뿐.
- 익절 · 사다리 부분 청산이 같은 사유로 거절: 단계 안 열림(D9 문장만).
- 미체결 목록에 식별 불완전 행 · 목록 미완: 취소 0. 그 포지션의 손절은 사람이 치울 때까지 계속 거절된다 — 오늘과 같고 알림이 원인을 말한다.
- 목록에 매수가 없음(브로커가 거절했는데 매수가 안 보임 — 조건주문 · 타 채널 가능성): 취소 0, 로그. 이름 붙인 잔여.
- **이미 취소를 낸 번호(D7)**: 그 매수가 확정 거절 · park 뒤에도 살아 있으면 다시 취소하지 않는다. 손절은 사람이 치울 때까지 반대 미체결 거절이 반복되고(오늘과 같음), 두 번째 critical 이 이미 사람에게 넘겼다. 그 사이 매 거절 주기마다 OPEN 조회 1 회는 계속된다(D4 가).
- **park 뒤 그 종목의 신규 진입(D11)**: 운영자 해동까지 거절(`unresolved_in_doubt`). 재시작으로 안 풀린다. 다른 종목은 무관.
- 같은 주기에 열 조건이 둘 이상: 하나만 처리(D4 나). 나머지 포지션의 처리는 다음 거절 주기로 밀린다(그 포지션 손절의 시점은 단계가 없을 때와 같은 반복).
- 같은 종목에 엔진 PLACE/AMEND park 가 있음: 단계 안 열림(D1).
- **귀속 판독의 범위(r2 #7)**: `EngineBrokerOrderIDs` 는 계좌 전체 · 기간 무제한이라(시장 · 종목 · 거래일 제한 없음, `daily_buy.go:350-357`) 오늘의 외부 매수 번호가 과거 어느 엔진 발주 번호와 같으면(번호 재사용) 엔진 주문으로 간주되어 취소되지 않는다 — 보수 방향(오늘 동작으로 남음). 브로커 번호의 영구 전역 유일성은 미측정.
- **park 의 실제 차단 범위(r2 확인)**: 「신규 진입」은 정확히는 그 시장 · 종목의 **노출 증가 요청 전부**다 — 같은 종목의 다른 포지션 · 전략 레인의 진입, 수량 · 가격을 높이는 정정 포함(`checkSymbolFree` 의 `raisesExposure`). 편입이 같은 게이트로 막힌다는 근거는 없다.
- **알림 기록 실패**(D6): 계정 전체 진입 래치 · 모드 승격 — 기존 안전 장치, 해동으로 안 풀림.

## 4. 시험 전략 (RED 먼저)

새 시험 파일 `internal/app/engine/a197_external_opposite_buy_test.go`. 결합 시험은 a198 의 진짜 게이트웨이 하네스(`newA198Fixture` — `buildE2EStack` + httptest 브로커)를 재사용하고, 미체결 판독은 가짜 판독기(쪽 · 오류 · 지연 주입)로 둔다. 가짜 제출자만으로 결합을 시험하지 않는다(정본 「길 치우기의 취소 결합은 실제 주문 경로를 지나는 시험으로 고정된다」와 같은 이유).

1. 토글 켜짐, 보호 손절이 반대 미체결로 거절, 목록에 외부 매수 1 → 그 주기 **끝**에 critical 「시도」 1 → 취소 제출 1(그 번호 바이트 그대로) → 원장에 CANCEL attempt CONFIRMED → 다음 주기 손절 제출 접수(실제 게이트웨이).
2. 토글 꺼짐 → 판독기 호출 0, 취소 0, park 0, 알림은 D9 문장 1, 재발의는 종전대로(places=2 — `a094_exit_clear_test.go:794` 의 기존 단언과 같은 모양).
3. 외부 매수 2 + 외부 매도 1 → 그 주기에 취소 1 건(첫 매수)만, 다음 거절 주기에 둘째 매수, 매도 무접촉.
4. 엔진 귀속 매수가 목록에 있음(귀속 집합에 번호 있음 — CONFIRMED · IN_DOUBT · FAILED 각각) → 이 단계는 그 주문을 취소하지 않는다.
5. 식별 불완전 행 1(번호 없음 / 방향 미지 / 행의 종목 불일치) + 외부 매수 1 → 취소 0.
6. 판독 오류 · 시한 초과(가짜 시계 · 막히는 판독기) · 「다음 쪽 있음 ∧ 커서 없음」 · 쪽 상한 초과 → 취소 0, 단계가 조회 시한 안에 끝남.
7. 목록에 매수 0 → 취소 0, 로그.
8. 익절 발의가 같은 사유로 거절 → 판독기 호출 0.
9. 다른 거절 사유(`dispatch_rejected_by_broker`) → 판독기 호출 0.
10. **한 번만(D7)**: 취소가 확정 거절(409 `already-canceled` · 400) → 두 번째 critical 1, 다음 거절 주기들에 같은 번호로 취소 요청 0(판독은 주기마다 1). 같은 원장으로 관측기를 새로 만든(재시작) 뒤에도 0. 게이트웨이가 기록 전에 거절한 경우(같은 종목 미종결) → attempt 행 0 → 다음 거절 주기에 다시 섬. RECORDED 에서 멈춘 취소(재시작 규칙으로 NOT_DISPATCHED) → 다시 내지 않음.
11. 알림 key 는 (포지션, 주문 번호) — 주기가 반복되어도 「시도」 critical 은 번호당 1. 알림 제목 · 본문 · 필드에 계좌 정보 · 주문 번호 0.
12. **모호 → park(D11)**: 취소 5xx → 같은 단계 안에서 UNRESOLVED_IN_DOUBT, 진입 게이트 래치 0, 다음 주기 손절이 park 때문에 막히지 않고 **제출**됨(실제 게이트웨이 — 2차 probe 모양). 두 갈래를 나눠 잰다(r2 #4): 「취소 실행 뒤 응답 유실」(매수 없음 → 손절 접수)과 「취소 미실행 5xx」(매수 남음 → 손절이 반대 미체결로 다시 거절, D7 로 재취소 0). 또 같은 종목 진입 BUY `unresolved_in_doubt`, 다른 종목 진입 통과, 두 번째 critical 의 본문에 운영자 절차, 세션 스윕 무변.
13. **주기 끝 실행(D5)**: 두 포지션(앞 포지션이 열 조건)에서 뒤 포지션의 판정 · 제출이 단계보다 먼저 일어남(막히는 판독기로 단계를 붙잡아도 뒤 포지션 손절이 그 주기에 나감). 전용 로거가 없는 조립에서도 단계가 돎(B1 앞).
14. 주기당 상한(D4 나): 같은 주기에 열 조건 셋 → 조회 1 · 취소 ≤ 1, 다음 주기들에 처리 대상이 돌아감.
15. 같은 종목에 엔진 PLACE park 있음 → 판독기 호출 0(D1).
16. 한 주기의 브로커 요청 수 계측(조회 쪽 수 + 사전 단건 조회 + 취소) — D4 의 실측.
17. `a087_submit_structure_test` · `a163_judgement_pin_test` · `a094_exit_clear_test` · `a111_flat_exit_observation_test` · `a198_clear_cancel_log_order_test` 무편집 통과(깨지면 Manager 에게).
18. 설정: 키 없음 · 명시 false → 꺼짐, 명시 true → 켜짐, `automation_gate` 키 없는 파일에서도 반영(`mergeEngine` B5 앞). audit 은 값이 바뀐 기동에만 1 줄(O-1 결정 뒤 실패 갈래 시험 추가).

변이(전부 CAUGHT 목표 — 닿음 확인 · 무변이 대조군 GREEN 뒤): 토글 검사 삭제 · 보호 손절 조건 삭제 · 사유 조건 삭제 · 매수 조건 삭제(매도도 취소) · 귀속 제외 삭제 · 미지 행 가드 삭제 · 「읽지 못함」을 「비어 있음」으로 · 커서 없는 다음 쪽 허용 · 조회 시한 삭제 · 「이미 냈다」 판독 삭제(또는 상태를 CONFIRMED 로 좁힘) · park 전이 삭제(IN_DOUBT 방치) · park 에 계정 래치 추가 · 주기 끝 호출을 판정 경로로 옮김 · B1 뒤로 옮김 · 주기당 상한 삭제 · PLACE park 가드 삭제 · 알림 key 에서 주문 번호 삭제 · 「시도」 알림을 취소 뒤로 옮김.

선택 시험: `./internal/app/engine -run 'A197|A094|A087|A129|A198|ExitClear|ProposalRefused'`, `./internal/execgw -run 'Refusal|Cancel'`, `./internal/config -run 'A163|A197'`. `-race` 는 `-run A197` 만.

## 5. 운영 (사람 승인)

1. 배포 뒤 토글은 꺼진 채다. 켜기 전에: 한 종목에 사람 매수 미체결을 두고 손절이 걸리는 상황을 만들 수는 없으므로(LIVE 주문 금지) 시험 증거와 design 을 사람이 읽고 flip 한다. 켜면 청산 전용 기동에서도 돈다(D1).
2. flip 뒤 첫 발생 때 원장의 취소 attempt · 알림 · 손절 접수 시각을 `review.md` 에 시각 · 상태만 기록한다.
3. **park 가 났을 때(D11) — 두 번째 critical 본문에 싣는 절차**:
   1. 앱에서 그 종목의 그 매수 주문이 남아 있는지, 취소됐는지, 체결됐는지 본다.
   2. `tossctl engine attempt-resolve --attempt <알림의 attempt id> --target … --operator … --approval … --note …` 를 사람이 실행한다(mutating, 엔진 lock 없음 — 엔진을 세우지 않는다).
   3. **취소 attempt 에서 `--target` 의 뜻(명령 도움말은 발주 기준이라 다르게 읽는다)**:
      - `CONFIRMED` = 「취소가 성립했다(그 매수가 취소됐다)」. `--broker-order-id` 에는 **그 매수의 주문 번호(취소 대상)** 를 넣는다 — 기동 해소와 a198 D10 이 취소를 닫을 때 대상 번호를 기록하는 것과 같은 관례(`resolveCancel` `amend_indoubt.go:81-87`).
      - `FAILED_CONFIRMED` = 「취소가 성립하지 않았다」. 도움말의 「브로커에 없다」는 이 경우 뜻이 다르다.
      - 앱에서 본 상태별 판단(기동 해소 `resolveCancel` 의 판정 표와 같게 — r2 #5 로 정정; `amend_indoubt.go` 갈래):

        | 앱에서 본 그 매수 | 해동 | 근거 갈래 |
        |---|---|---|
        | 취소됨(일부 체결 뒤 취소 포함) | `CONFIRMED`, `--broker-order-id` = 그 매수 번호 | B7 `case@81` |
        | 전량 체결 · 거부 | `FAILED_CONFIRMED` | B9 `case@92` · B11 `case@103` |
        | 취소 · 정정 처리 중 | **해동하지 않고 기다린다** — 확정 실패가 아니다 | B13 `case@116` |
        | 미체결로 남음 | 한 번 보고 닫지 않는다 — 수 분 간격으로 다시 봐도 미체결이면 `FAILED_CONFIRMED`(기동 해소는 3 회 · 45 초 이상 안정 관찰) | B15 `case@127` |
        | 정정되어 다른 번호로 대체됨 · 상태를 알 수 없음 | **해동하지 않는다** — 기동 해소도 park 하는 상태 | B19 `case@145` · default `case@153` |
   4. 해동은 그 종목의 신규 진입 차단만 푼다. 해동해도 D7 표시(원장의 CANCEL 행)는 남으므로, 매수가 아직 살아 있으면 엔진은 그 번호를 다시 취소하지 않는다 — 앱에서 사람이 취소해야 손절이 나간다.
   - 명령 코드(도움말) 수정은 이 change 범위 밖이다.
4. 「치워지지 않았다」 critical(확정 거절)의 절차: 앱에서 그 매수를 취소한다. 엔진은 그 번호에 다시 내지 않는다(D7).

## 6. 잔여·후속

- 브로커가 거절했는데 미체결 목록에 매수가 없는 경우(§3).
- 미체결 조회의 실제 지연 분포 · OPEN 이 종목 필터를 지키는지 — 운영 미측정. 시한 3 초는 기존 조회 SLO 에서 유도: R7 단건 GET p99 ≤ 0.6 s(`retry-matrix-slo-2026-10-22.md:27,46`)의 5 배, 관측 max 4.17 s 보다 짧음.
- 장 마감 뒤 재제출, 거절 사유가 Reason 에서 사라지는 `order-hours-closed` — 별도 후보.
- 취소한 매수 주문을 사람이 다시 넣으면 새 번호이므로 그 번호에 한 번 취소가 나간다 — 알림이 그것을 말한다.
- park 불가 잔여(D11 끝): DISPATCH_STARTED · ACKED 고착, `ResolveUnresolved` 실패 — 재시작까지.
- 단계 시간만큼의 다음 주기 지연(D5) · 토큰 잠금 무상한(D4) — 기존 길 치우기와 같은 성질. 후속 후보는 토큰 획득의 ctx 존중.
- 거절이 계속되는 동안의 주기당 OPEN 1 GET(D4 가).
- **a087 종결 쪽의 기존 비용**: a087 은 `landed-commit.txt` 가 없어, a197 이 `submit` · `alertProposalRefused` · `endExitOrderCycle` 을 먼저 착지시키면 a087 종결 때 그 함수들의 FLM 을 요구받는다(`exit_order_log.go` 쪽은 a198 로 이미 같은 비용이 생김).

## 7. 보류와 재개 — **재개 반영 완료(2026-11-25)**

> 2026-11-23 Manager 결정으로 보류됐다. 재개 조건 1(a198 착지 — 2026-11-24 main · 운영 반영), 2(선행 2 결정 — Manager 2026-11-25, (E)), 3(이 개정)이 섰다. 4(개정본 r2 적대 리뷰)는 `review.md`.

| r1 # | 2026-11-23 결정 | 반영 자리(이 개정) | Q5 의 전제 변경 적용 |
|---|---|---|---|
| #1 | 선행 a198 | D6 — a198 의 발행 경로 | a198 착지로 해소(probe Q1) |
| #2 | 미종결 · park attempt 있으면 열지 않음, 귀속은 상태 무관 판독 | D1 · D2 | 미종결 절반은 기존 코드가 보장(`clearTheSymbol` B3 → `record` B12 — 인용), 새 가드는 PLACE/AMEND park 쪽만 |
| #3 | 선행 2 | D11 | (E) 종목 범위 park — Manager 결정 1 |
| #4 | 단계 전체 시한 + 주기당 상한 | D4 · D5 | hard 시한 불가를 사실대로, 주기 끝 실행 · 1 건 상한(Manager 결정 3) |
| #5 | 「다음 쪽 있음 ∧ 커서 없음」 = 읽지 못함 | D3 | — |
| #6 | 번호 바이트 보존 · 행 값 재대조 | D2 | — |
| #7 | 알림은 취소 전에 | D6 | 주기 끝 자리라 a198 D4 핀과 충돌 없음 |
| #8 | ADDED + MODIFIED 로 경계 | spec delta | 인접 Requirement 3(`:885` · `:905` · `:917`) 포함, order-execution 「IN_DOUBT 해소」 MODIFIED 추가(Manager 결정 1) |
| #9 | 꺼짐 = 조회 0 · 취소 0, 롤백 = 끄고 재시작 | D8 · D1 | 청산 전용 기동에서도 돈다는 공시 |
| #10 | audit 은 `NewContext` 무편집 자리 먼저 | D8 | `Context.ExitObserver` 의 `c.Audit` — 실패 처리 O-1 |
| #11 | 알림 단위 (포지션, 주문 번호) | D6 · D7 | — |

**열린 항목(Manager 결정 필요)** — r2(`review.md` 처분표)에서 온 것 포함

- **M-1 — 「한 번만」의 범위(r2 #2).** 지금 글은 취소 intent 의 거래일을 키에 넣어 「취소한 거래일당 한 번」이다. 「주문 생애당 한 번」으로 하려면 거래일을 빼야 하고, 그러면 다른 날 번호가 재사용된 새 외부 매수는 영구히 대상에서 빠진다(보수 방향). 알림 key 의 범위도 같이 정한다.
- **M-2 — 중단 창의 복구(r2 #1).** ① IN_DOUBT 기록 뒤 park 전 중단 → 기동 해소기의 계정 래치 가능, ② park 뒤 결과 알림 전 중단 → 결과 알림 누락. 공시로 둘지, 기동 처리 · 누락 알림 복구 정책을 둘지.
- **M-3 — 대기열 항목의 실행 직전 재검증(r2 #3).** 주기 끝 실행 직전 포지션의 존재 · 세대 · 보호 필요성을 다시 볼 가드(새 가드).

- **O-1 — 토글 audit 기록 실패 시 처리.** `NewContext` 는 audit 실패면 기동을 거절한다(`engine.go:486-492`). 같은 규칙을 `Context.ExitObserver` 에 두면 관측기 자체가 없어져 손절이 멈추므로 그대로 옮길 수 없다. 후보: (가) audit 실패면 그 기동에서 토글을 꺼진 것으로 다루고 critical 1(오늘 동작으로 내려감), (나) audit 을 `NewContext` 의 `recordGateSettings` 에 넣고 a166 census 갱신을 승인(a199 선례). 작성자 기본안은 (가)이나 새 정책이라 결정 대기.
