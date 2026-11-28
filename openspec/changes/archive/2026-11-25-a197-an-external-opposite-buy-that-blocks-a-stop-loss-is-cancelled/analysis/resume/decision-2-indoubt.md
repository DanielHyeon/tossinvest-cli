# a197 재개 — Manager 결정 2(선행 2, r1 #3) 조사 메모

> 작성 2026-11-25, 팀메이트(Opus), worktree `~/tossos-work/a197-impl`, HEAD `f18de89ef`(병합 `e7e3f22b4` + base 재고정 `b93ed133a`).
> 근거는 `analysis/cited-ast/*.ast.json`(현재 HEAD 재추출 · 신규 16)과 일회성 probe 원문 `analysis/probe/resume-probe-output.txt`
> (원본 `analysis/probe/resume-zz_a197_resume_probe_test.go.txt`, 버리는 worktree 에서 1 회 실행 · 가짜 브로커 · 합성 번호 `EXT-1` · 네트워크 0).
> 「손으로 읽음」 표시가 없는 분기 주장은 AST 열거에서 왔다. 결정 문장은 쓰지 않는다 — 선택지와 평가만.

## 결론 먼저

1. **선행 2 는 a198 로 풀리지 않았다(실측).** a197 이 낸 외부 매수 취소가 IN_DOUBT 가 되면 같은 종목의 손절 매도와 수동 flatten 이 `symbol_mutation_in_flight` 로 막히고(probe Q2a · Q2e), a198 D10 세션 닫기는 소유 후보 0 이라 닫지 않으며(Q2b `closed=false`), 엔진 재시작의 기동 해소만 푼다(Q2d).
2. 그래서 토글을 켠 엔진은 「취소가 모호해진 뒤 사람이 앱에서 매수를 치운」 경우 손절이 **재시작까지** 늦어진다 — 오늘(토글 꺼짐)은 다음 주기에 나간다. 수동 flatten(비상 청산)도 같은 창에서 막힌다.
3. 권고: **(C)** — 단, 「모든 safety class 에 심볼당 in-flight 1개」 정본 SHALL 의 MODIFIED 와 원장 판독 전수(미측정)가 조건. (C) 를 받지 않으면 다음은 (A) 이고, (A) 뒤에도 오늘보다 늦어지는 창(≥ 45 초 ~ 5 분+)이 남는다.

## Q1. a198 뒤 exit 루프의 취소 결정 경로, 그리고 외부 BUY 번호를 받는가

| 단계 | 무엇 | 근거 |
|---|---|---|
| 발행기 | `(&execgw.Issuer{Journal: o.opts.Journal, Clock: o.clk, AccountRef: o.opts.AccountRef}).IssueReduction(ReductionRequest{Kind: KindCancel, Market, Symbol, Side: 주문 방향, MaxQuantity: 주문 수량, Reason})` | `clearTheSymbol` call `exitloop.go:1700`(AST `engine--ExitObserver.clearTheSymbol`) |
| Kind · 등급 | `CANCEL` · `RISK_REDUCING` | `Issuer.IssueReduction` `issue.go:86-103`(분기 1, 무변) — 결정 행만 쓰고 예약을 잡지 않음 |
| Side | 취소할 주문의 방향(BUY 가능) | `:1705` |
| 멱등 키 | 없음(CANCEL 결정에는 키가 붙지 않음 — 옛 결함의 원인이 PLACE 결정이었음) | probe Q1 `state=CONFIRMED`(Prepare 거절 없음) |
| 제출 | `o.opts.Submit.Cancel` → 생산 `c.Gateway`(`exitwiring.go:340`) = `*execgw.Gateway`, source `engine`, `CancelRefusal: EngineCancelRefusal()`(`internal/app/engine/gateway.go:327,329`). 브로커 어댑터가 취소 전에 사전 단건 조회 1(`precheck.go:58·99`, 2s, 결과는 advisory) | `:1715` |
| 게이트웨이 정책 | `withCancelRefusal`(AST `execgw--Gateway.withCancelRefusal` L44-112 · 15): 엔진 정책이면 a198 D7(`settledCancelRefusal`) → a165 `already-processing` 갈래 | AST |

**외부 BUY 번호를 받는다.** 원장에 그 번호의 귀속이 하나도 없는 `EXT-1`(BUY)을 같은 발행 · 같은 정책 게이트웨이로 내면 `state=CONFIRMED reason=broker_accepted broker_cancels=1`(probe Q1). 거절하는 분기 없음: `Gateway.Cancel`(L407-426, 분기 0) · `prepareRequest`(무변, 대상 주문 조회 0) · `checkSymbolFree`(대상 소유를 보지 않고 같은 종목 미종결만 봄, 아래 Q2a) — 모두 옛 base 와 같은 AST. intent 소유 검사 · 예약 · 정책 어디에도 「대상 번호가 엔진 귀속이어야 한다」 는 조건이 없다(Prepare 는 이번에 AST 를 뽑지 않았으므로 「Prepare 가 소유를 보지 않는다」 는 probe 결과로만 주장).

## Q2. 그 취소가 5xx · 시한 초과로 IN_DOUBT 가 되면

probe Q2: 브로커 `official.ErrServer`(5xx 를 공식 클라이언트가 바꾸는 값) → `state=IN_DOUBT reason=dispatch_outcome_unknown`.

### (a) 같은 종목 미종결 가드 — **종목만 본다. 손절 제출을 보류한다**

- 판정 함수는 하나: `Gateway.unsettledFor`(`a094_shared_judgements.go:30-46`, AST 분기 4) = `PendingAttempts`(RECORDED · DISPATCH_STARTED · ACKED · IN_DOUBT, `recovery.go:30-33`) 를 `attemptTargets`(`gateway.go:842-849` — intent 의 **종목 · 시장만** 비교, 손으로 읽음)로 거름. attempt 의 종류 · 방향 · 대상 소유는 보지 않는다.
- 길 치우기 쪽: `clearTheSymbol` call `:1658` `UnsettledOnSymbol` → B3 `if@1662 len(unsettled) > 0` → `return clearResult{countable: !onlyEngineCancelsInFlight(…)}`(`:1665`). IN_DOUBT 는 `onlyEngineCancelsInFlight` B4 switch 의 default(`:73` `return false`)라 **셈**(연속 실패 계수에 들어감). → `record` B12 `if@1402 !cleared.cleared` → `noteDelay@1405`(30초 지연 critical) · `noteClearFailure@1406` · `orderable=false` → **`submit`(call `:1456`) 에 닿지 않음** → a197 단계(B15 안)도 다시 열리지 않음.
- 주문 경로 쪽(가드를 우회해도): `checkSymbolFree` B2 `if@808` 가 `symbol_mutation_in_flight` 로 거절 — probe Q2a 손절 SELL `state= reason=symbol_mutation_in_flight broker_places=0`, 대조군(취소 없음) `CONFIRMED broker_places=1`.

### (b) a198 D10 세션 스윕 — **닫지 않는다(소유 후보 0 갈래)**

- 스윕: `inDoubtSettle.sweep` B6 `case@184`(IN_DOUBT ∧ CANCEL) → `closeCancel`(call `:196`). CANCEL 은 조회 대조(`resolve`, call `:202`)의 대상이 아님(B5 `case@182` 는 PLACE 만).
- 판정: `CloseInDoubtCancelOnTargetClose` B4 `if@88`(IN_DOUBT · CANCEL · 대상 번호 · source 일치 — a197 취소는 source `engine` 이라 통과) → 소유 수 질의(`:96`, `owners` CTE `a198_cancel_target_close.go:55-63` = 대상 번호를 `broker_order_id` 로 가진 CONFIRMED PLACE/AMEND) → B6 `if@100 owners != 1` → 무변 반환(`:101`).
- probe Q2b: `closed=false err=<nil>`, 뒤에도 `unsettled=[CANCEL/IN_DOUBT/target=EXT-1]`. 외부 주문은 원장에 소유 발주가 없으므로 종결 체결 기록이 있든 없든 이 갈래에서 끝난다(a198 design §3 「소유 후보 0(외부·사람 주문이 대상)」 과 일치).

### (c) a198 D7(`already-*` 확정 거절) — **적용된다(대상 소유와 무관)**

- `settledCancelRefusal`(L30-52 · 5)은 정책이 Definitive 인 CANCEL 의 응답 code 만 본다. probe Q2c: `409 already-canceled` → `FAILED_CONFIRMED reason=dispatch_rejected_by_broker`, 미종결 0(가드 안 걸림). `409 already-modified` → `IN_DOUBT`(D7 제외 대상 — 가드에 걸림).
- 5xx · 429 는 공식 클라이언트가 `APIError` 가 아닌 값으로 바꿔 D7 에 닿지 않는다(a198 design §6 「경보」 문단, `internal/official/errors.go:54` 인용). 전송 시한 초과도 모호다. 즉 IN_DOUBT 로 남는 입력 = 5xx · 429 · 전송 불명 · `already-modified` · code 없는 409 등.

### (d) 엔진 재시작의 기동 해소 — **단건 판독으로 끝낸다(소유 검사 없음)**

- `Recovery.Run` 2단계: IN_DOUBT 면 재생 시도 뒤 `Resolver.Resolve`(`recovery.go:300`) → CANCEL 갈래(`indoubt.go:274-276`, 손으로 읽음) → `resolveCancel`(AST L44-158 · 20): 대상 번호 단건 판독(call `:70` `deriveOrder`)만 쓴다.
  - 취소됨 → CONFIRMED(B7 `case@81`), 체결 → FAILED_CONFIRMED(B9 `case@92`), 거부 → FAILED_CONFIRMED(B11 `case@103`), 미체결이 3 회 · 45 초 이상(`DefaultResolveConfig` `indoubt.go:156-163`) 안정 → FAILED_CONFIRMED(B15 `case@127`), 취소 · 정정 대기 → 대기(B13), REPLACED · UNKNOWN · 판독 실패가 예산(5 분) 소진 → park(UNRESOLVED_IN_DOUBT).
- probe Q2d: 판독이 「취소됨」 → `CONFIRMED in_doubt_resolved_found`, 미종결 0, 손절 SELL `CONFIRMED broker_places=1`. 판독 불가 → `UNRESOLVED_IN_DOUBT`, 미종결 0(park 는 `PendingAttempts` 밖), 손절 SELL 통과(`checkSymbolFree` B3 `if@814 !raisesExposure` 가 park 검사 앞에서 반환), **진입 게이트가 `unresolved_in_doubt` 로 잠김**(생산 Resolver 의 `Gate` = 엔진 진입 게이트 `internal/app/engine/gateway.go:298-304`). 기존 시험 `TestCancelInDoubtResolvesFromTheOriginalOrder`(`amend_indoubt_test.go`)도 원장에 귀속 없는 `O-1` 취소를 같은 표로 해소한다.
- **주의(기억 · advisory, 이번 회차 미재측정)**: KR 에서 앱으로 취소된 주문의 단건 조회는 `PENDING`·`canceledAt` 없음을 계속 돌려준다(측정 M32, a191 proposal `:26` 인용). 엔진의 API 취소에도 같은지는 미측정. 그 경우 기동 해소는 「미체결 안정」 갈래로 FAILED_CONFIRMED(`cancel never landed`)를 쓴다 — 판정 문장이 틀릴 수 있으나 차단은 풀린다(안전 방향).

### (e) 수동 flatten — **그 종목에서 보류된다**

- flatten 게이트웨이(source `flatten-all`, `FlattenCancelRefusal`)의 같은 종목 취소: probe Q2e `state= reason=symbol_mutation_in_flight broker_cancels=0`. 같은 `checkSymbolFree` B2 이므로 flatten 의 매도도 같다(그 매도는 probe 하지 않음 — 같은 분기라는 AST 근거).
- flatten 은 기동 복구를 돌리지 않으므로 엔진이 멈춘 동안 이 보류는 풀리지 않고, 엔진 재기동으로 풀린 뒤에는 flatten 을 다시 실행해야 그 종목이 청산된다(a198 design §2-7 ③ · §6 첫 문단과 같은 성질 — a197 은 그 원인에 「외부 매수 취소」 를 더한다).

## Q3. 오늘(토글 꺼짐)과의 비교 — 판정 기준: 안전 불변식 4(손절 · 비상 청산의 즉시성을 약화 · 지연하지 않는다)

오늘: 반대 미체결 거절 → 해제 → 5 초 뒤 재발의 반복. 사람이 앱에서 매수를 치우면 다음 주기에 손절이 나간다. 토글 켜짐에서 a197 취소의 결과를 셋으로 나눈다.

| 경우 | a197 취소 결과 | 손절이 나가는 시점(토글 켜짐) | 오늘보다 늦어지는가 |
|---|---|---|---|
| 1. 사람이 나중에 그 매수를 앱에서 치운다 | 접수 확정(CONFIRMED) | 취소 접수 뒤 다음 주기(브로커 반영 지연 동안은 반대 미체결 거절 반복 — a198 §2-5) | 아니오(빠름) |
| | 확정 거절(4xx · D7 `already-*`) | 미종결 0 → 사람이 치운 뒤 다음 주기 | 아니오(같음) |
| | **IN_DOUBT**(5xx · 시한 · `already-modified`) | 같은 종목 미종결 가드가 막음 → **엔진 재시작의 기동 해소 뒤**(Q2d) | **예 — 재시작까지**. 그 사이 a198 D10 은 닫지 않음(Q2b) |
| 2. 취소가 실제로는 성공했다 | CONFIRMED 로 기록 | 다음 주기 | 아니오(빠름) |
| | **IN_DOUBT 로 기록**(브로커는 취소했는데 응답이 5xx · 시한) | 매수는 이미 없는데 가드가 막음 → 재시작 뒤. 사람은 앱에서 치울 매수가 없어 할 수 있는 일이 재시작뿐 | **예 — 오늘은 사람이 치우는 즉시, 여기서는 재시작까지**(사람 행동 시점은 미측정) |
| 3. 취소가 실패했고 매수가 살아 있다 | 확정 거절 | 다음 거절 주기에 단계가 다시 열려 재취소(주문당 주기마다) — 손절 시점은 사람이 치울 때까지로 오늘과 같음 | 아니오(같음, 브로커 호출만 늚) |
| | **IN_DOUBT**(실제로는 미성립) | 가드가 막음 → 재시작 뒤 기동 해소가 「미체결 안정」 으로 FAILED(≥ 45 초) → 다시 거절 주기. 그 사이 사람이 치워도 손절은 재시작까지 | **예** |
| 비상 청산(수동 flatten), 위 IN_DOUBT 세 칸 모두 | — | 그 종목의 flatten 취소 · 매도가 `symbol_mutation_in_flight`(Q2e) — 엔진 재기동 뒤 flatten 재실행 | **예** |
| 참고: 단계가 도는 주기 자체 | 모든 결과 | 같은 루프의 뒤 포지션 판정이 단계 시간(목록 ≤ 3s + 대상당 사전 조회 ≤ 2s + 취소 ≤ 15s + 토큰 잠금 대기 — 상한 미확정)만큼 늦음(design D5, r1 #4) | 예(유계 여부 미확정) — 선행 2 와 별개로 r1 #4 가 다룸 |

경보: IN_DOUBT 로 막힌 동안 손절은 제출되지 않으므로 D9 문장(「매수 미체결이 막고 있다」)은 나가지 않는다. 나가는 것은 지연 경보(30 초, `noteDelay@1405`)와 연속 실패 경보(IN_DOUBT 는 계수에 들어감 — `onlyEngineCancelsInFlight` `:73`).

## Q4. 선택지 평가

공통 사실: IN_DOUBT 외부 매수 취소는 다음 셋 어디에도 닿지 않는다는 근거가 있다 — ① 예약: `execgw.Issuer.IssueReduction`(`issue.go:86-103`)은 결정 행만 기록하고, 예약은 `RiskGuardian` 의 노출 증가 결정에서만 잡힌다(`riskguardian.go:533,535` · `riskguardian_qfinal.go:266`, rg) ② 귀속 · 소유 판독: `LiveOrdersForSymbol` · `EngineBrokerOrderIDs` · D10 `owners` CTE 모두 `kind IN ('PLACE','AMEND')`(AST · 원문) ③ 해소 증거: 취소 해소는 대상 단건 판독만(`resolveCancel` call `:70`), 발주 해소의 matcher 는 방향으로 거른다(`indoubt.go:691`, 손으로 읽음 — 단 브로커 행의 side 가 비면 거르지 않음). **미측정**: 원장의 `mutation_attempts` 를 읽는 SQL 은 비시험 28 파일 139 줄 — 전수 대조는 하지 않았다(위 셋과 `daily_buy.go:265-280` 만 봄).

### (A) 「세션 중 외부 대상 IN_DOUBT 취소 해소」 를 선행 change 로

- 모양: 스윕이 D10 에서 「소유 후보 0」 으로 끝난 엔진 취소를 기동과 같은 `Resolver.Resolve`(→ `resolveCancel`, 단건 판독)로 넘긴다.
- 고칠 기존 함수(FLM): `inDoubtSettle.sweep`(B4~B6 갈래) · `Journal.CloseInDoubtCancelOnTargetClose` 또는 `inDoubtSettle.closeCancel`(지금은 「소유 0」 을 `Closed=false` 로 뭉뚱그림 — 갈래를 돌려줘야 함) → **최소 2**, 알림 규칙(정본은 실행자가 CANCEL 에 알림을 내지 않는다고 함)을 바꾸면 `inDoubtSettle.resolve`/`noteParkOutsideExit` 까지 3~4.
- 구조 시험 · census: `sweep` 을 굳히는 AST 시험 없음(rg). a198 행동 시험(`TestA198TheCancelIsClosedBeforeThePlaceResolution` 등) 유지 대상. a166 census 무관.
- 정본: order-execution 「IN_DOUBT 해소」 의 「세션 중 해소: 대상」 문단(`openspec/specs/order-execution/spec.md:125`)과 시나리오(`:250` 「대상 주문의 종결 기록이 없는 IN_DOUBT 취소 attempt 를 읽으면」) **MODIFIED**, 「세션은 IN_DOUBT 로 남은 취소를…」(`:1209`) 범위 문장 MODIFIED.
- 거부하게 되는 정상 입력: 새 거절 없음. 대신 새 부수 효과 — 세션 중 단건 판독(5 초 간격, attempt 당 최대 5 분, 발주 해소와 직렬 → 다른 종목의 해소가 그만큼 밀림), park 되면 **세션 park = 계정 전체 진입 래치**(probe Q2d · a198 이 D10 에서 일부러 피한 것, Manager Q4).
- 통과시키게 되는 입력: KR 앱 취소 대상의 단건 판독이 `PENDING` 이면(M32) 「미체결 안정」 으로 FAILED_CONFIRMED → 차단 해제(판정 문장 오류 가능, 방향은 안전).
- Q3 에서 남는 나쁜 칸: IN_DOUBT 세 칸이 「재시작까지」 → 「IN_DOUBT 뒤 다음 스윕 + 최소 45 초(3 관측) ~ 최대 5 분 + 앞선 발주 해소 대기」로 **줄지만 0 이 아니다**. flatten 칸도 같은 창으로 줄어듦.
- a087 · a167 충돌: 없음(`in_doubt_settle.go` · a198 파일은 두 change 의 표면 밖).
- 범위: 소유 0 인 엔진 취소는 오늘 생산에서 a197 만 만든다(길 치우기 대상은 `LiveOrdersForSymbol` 이라 소유 ≥ 1) — 사실상 a197 전용 경로.

### (B) 모호해지면 (포지션, 주문 번호)에서 재개 금지 + critical 만, 가드는 그대로

- 고칠 기존 함수: a197 계획 그대로(`submit` · `alertProposalRefused` + 새 파일) — 추가 FLM 0. 정본: a197 ADDED 에 「모호 → 재시작까지 그 종목 손절 · flatten 보류」 공시만.
- 실효: IN_DOUBT 인 동안은 `record` B12 가 `submit` 앞에서 끊어 단계가 어차피 다시 열리지 않는다(Q2a). 「재개 금지」 가 의미를 갖는 것은 재시작 뒤인데, 메모리 표시는 재시작에 사라지고 원장에 남기려면 새 상태(범위 밖)가 필요하다. 남는 것은 사실상 「IN_DOUBT 순간의 critical 1」.
- 거부 · 통과: 새 것 없음.
- Q3 에서 남는 나쁜 칸: **전부 남음**(IN_DOUBT 세 칸 + flatten), 상한은 운영자의 재시작 — 코드상 무한. 불변식 4 를 공시로 양보하는 안이다.
- a087 · a167 충돌: 계획된 `submit` 편집 그대로(a087 `landed-commit.txt` 부재 비용은 기존 — `edit-surface.md` §0).

### (C) 외부 **매수**가 대상인 IN_DOUBT 취소는 보호 매도 제출을 보류시키지 않음 + 재개 금지 + critical

- 모양: 같은 종목 미종결 판정에서 「IN_DOUBT ∧ CANCEL ∧ 취소 intent source = engine ∧ 취소 intent 방향 BUY ∧ 대상 번호가 엔진 귀속 집합(`EngineBrokerOrderIDs`)에 없음」 인 attempt 를, **노출을 늘리지 않는 요청**의 판정에서만 뺀다. 진입(노출 증가)에는 그대로 차단.
- 고칠 기존 함수(FLM): a094 핀(`a094_exit_clear_test.go:560` — 길 치우기와 게이트웨이가 한 판정 `unsettledFor` 를 쓰는지 본다)을 지키려면 판정 본체에서 고쳐야 한다 → `Gateway.unsettledFor`(+ 계획에 요청 성격을 싣는 `Gateway.UnsettledOnSymbol`) **최소 2**, 원장 귀속 판독을 그 안에 넣음(모든 변경 요청의 종목 검사에 원장 읽기 1 이 더해짐 — 비용 미측정). `checkSymbolFree`·`clearTheSymbol` 본문은 무편집 가능.
- 구조 시험 · census: a094 핀 유지 가능(호출 관계 무변). a166 census 무관(`a094_shared_judgements.go` 목록 밖).
- 정본: order-execution 「IN_DOUBT 해소」 의 「조회 대조의 유일 매칭을 위해 심볼당 in-flight mutation 1개 제한은 **모든 safety class에** 유지된다(SHALL)」(`spec.md:46`) **MODIFIED** — 주문 경로의 핵심 불변식을 좁히는 개정이라 적대 리뷰 부담이 가장 크다. exit-policy 「치우기가 하나라도 확정 취소에 실패하면 보호 청산을 제출해서는 안 된다」(`:554` 요구)는 a197 취소가 길 치우기가 아니므로 문언상 무관하나 경계 문장 필요(r1 #8 과 같은 MODIFIED 에 담을 수 있음).
- 이중 매도 논증(코드 확인분): 취소는 매도를 만들지 않는다. 남는 불확실성은 「그 외부 매수가 살아 있는가」 하나이고, 그 위에 나간 손절 매도의 귀결은 브로커의 반대 미체결 거절(확정 거절, 종목 잠금 없음) 또는 매수 체결 경합에 의한 **과소 보호**(보유 증가)다 — 초과 매도 방향이 아니다. 예약 · 귀속 · 해소 증거 셋은 위 「공통 사실」 대로 닿지 않음. **미측정**: `mutation_attempts` 판독 전수(139 줄), 대사 `LocalStateFromJournal` 이 CANCEL 행을 어떻게 세는지.
- 거부하게 되는 정상 입력: 없음(거절을 하나 없앰).
- 통과시키게 되는 입력: 그 종목의 손절 매도 · 엔진 길 치우기 취소 — 그리고 판정을 「노출을 늘리지 않는 요청」 으로 열면 수동 flatten 의 취소 · 매도도(비상 청산 즉시성에는 이득이나 flatten saga 의 순서 가정이 바뀜 — 엔진 요청으로 좁힐지 Manager 결정). 한 종목에 모호 attempt 둘(이 취소 + 손절 매도 IN_DOUBT)이 공존할 수 있음 — 기동 해소는 서로 다른 증거(단건 판독 / 방향으로 거른 목록)로 각각 판정.
- Q3 에서 남는 나쁜 칸: **없음**(코드 근거 기준). IN_DOUBT 세 칸이 모두 「다음 주기에 손절 제출 → 매수가 없으면 접수, 있으면 반대 미체결 거절(오늘과 같음)」 로 바뀜. 남는 것은 IN_DOUBT 취소 attempt 가 재시작까지 원장에 남는 것(진입 차단 그 종목만 — 오늘과 비교 대상 아님)과 「단계 시간」 행(r1 #4).
- a087 · a167 충돌: 없음(같은 함수 편집 아님).

### (D) 기각한 안 하나 — 단계 자신의 OPEN 목록으로 닫기

단계가 이미 읽는 OPEN 목록에 대상 번호가 없으면 그 취소를 닫는 안. 「목록에 없음」 은 취소됨과 체결을 가르지 못해 판정(CONFIRMED / FAILED)을 정할 근거가 없고(a198 D10 은 같은 이유로 원장 종결 기록의 state 로만 닫음), 막힌 동안 단계가 다시 돌지도 않는다(Q2a). 재료 부족으로 기각.

### 권고

**(C)**. 근거: 판정 기준(불변식 4)에서 오늘보다 늦어지는 칸이 0 인 안은 (C) 뿐이다. (A) 는 창을 재시작에서 「≥ 45 초 ~ 5 분+」 로 줄이지만 0 으로 만들지 못하고, 세션 park 로 계정 진입 래치라는 새 비용을 더한다. (B) 는 창을 운영자의 재시작에 맡긴다.
조건(권고의 일부): ① 면제는 위 다섯 조건의 곱으로 좁히고 노출 증가 요청에는 적용하지 않는다 ② 정본 order-execution `:46` 의 SHALL 을 MODIFIED 로 명시 ③ freeze 전에 `mutation_attempts` 판독 전수(139 줄)와 대사의 CANCEL 행 처리를 실측해 「이중 매도 불성립」 의 미측정분을 닫는다 ④ flatten 을 면제에 넣을지는 Manager 결정. 이 조건을 받지 않으면 (A) 를 권한다(남는 창을 a197 정본에 공시).

## Q5. design §7 「재개 시 설계 변경(결정됨)」 9 항목 — a198 ~ a201 착지 뒤

| r1 # | 항목 | 지금도 성립 | 바뀐 전제 · 비고 |
|---|---|---|---|
| #2 | 그 종목에 엔진 미종결 · park attempt 가 있으면 열지 않음, 귀속은 상태 무관 판독 | 성립 | 「미종결」 절반은 이미 코드가 보장 — 미종결이 있으면 `record` B12 가 `submit` 앞에서 끊어 단계에 닿지 않음(Q2a). 실제로 새 가드가 필요한 것은 **park(UNRESOLVED_IN_DOUBT)** 절반(`PendingAttempts` 밖, `checkSymbolFree` B3 가 매도에는 park 를 보지 않음). `EngineBrokerOrderIDs` 무변. a198 의 CONFIRMED CANCEL 행은 `broker_order_id` = 대상 번호지만 kind 필터로 귀속 집합에 안 들어옴 |
| #4 | 시한은 단계 전체 + 주기당 취소 건수 상한 | 성립 | 전제 보강: 대상당 비용 = 사전 단건 조회(2s) + 취소 POST(HTTP 15s) + **context 로 끊기지 않는 토큰 잠금 대기**(a198 §2-4 · r2 P1-2) — 「단계 전체 시한」 은 ctx 로만 걸 수 있어 hard 상한이 아님. rate limit 행에 사전 조회 1 을 더해야 함(`edit-surface.md` §0.1) |
| #5 | 「다음 쪽 있음 ∧ 커서 없음」 = 읽지 못함 | 성립 | `ScanOrders` 무변(L730-765) |
| #6 | 주문 번호 바이트 보존 + 종목 · 방향 행 값 재대조 | 성립 | `parseTarget` · `Gateway.Cancel` 무변. 선례: a198 은 `order.OrderID` 를 가공 없이 `CancelIntent` 에 넘김(`:1716`) |
| #7 | 알림은 취소 **전에** 「취소를 시도한다」 | 성립 | 입구 `o.opts.Critical`(`exitwiring.go:353`) 무변. 새 사실: a198 D4 재결정 — 같은 루프 안의 **로그** 쓰기가 막히면 뒤 포지션 손절이 멈춘 실측이 있어 그 줄은 주기 끝에 씀(구조 핀 `a198_clear_cancel_log_order_test.go:140`). critical 기록(`RecordCritical`)이 루프를 막을 수 있는지는 미측정 — a094 가 같은 입구를 루프 안에서 이미 씀 |
| #8 | ADDED + 「길 치우기는 엔진 귀속 주문만」 MODIFIED | 성립 | 그 Requirement(`exit-policy/spec.md:554`)는 무변. 인접 Requirement 3 개가 a198 로 생김(`:885` 취소 결정 발행 · `:905` 실패 기록 · `:917` 접수 확정 매수 재취소 안 함) — 경계 문장이 이들과도 맞아야 함. 선택지 (A)/(C) 를 고르면 order-execution MODIFIED 가 더해짐(Q4) |
| #9 | 꺼짐 = 조회 0 · 취소 0, 알림 문장은 토글 무관, 롤백 = 토글 끄고 재시작 | 성립 | 설정은 여전히 기동 때만 읽힘(`exitwiring.go:358` 선례 무변). 새 전제: a199 로 **청산 전용 기동에서도 관측기가 돎**(B2@323 `RunsExitSide()`) — 토글이 켜지면 그 기동에서도 단계가 열림(취소는 진입 래치에 안 걸림 `gateway.go:859-861`) |
| #10 | 토글 · audit 자리는 `NewContext` 무편집 쪽 먼저, 없으면 census 갱신 Manager 승인 | 성립 | `NewContext` 여전히 census 고정. a199 가 이 census 항목을 갱신한 선례(`3445403ee` · `fadd2fc4c`)가 생겨 「census 갱신 승인」 경로의 비용 근거가 있음. `recordGateSettings` 본문 무변(L774-848) |
| #11 | 알림 「한 번」 단위 = (포지션, 주문 번호) | 성립 | a198 실패 줄의 중복 억제 key(포지션 · 주문 → 상태 · 사유)와 같은 축 |

§7 밖이지만 재개 개정 때 함께 고쳐야 할 본문 주장(결정 아님 — 사실 정정 후보): design D6(발행 경로 — a198 로 성립), D7 의 「취소가 … 모호면 … 다음 거절 주기가 다시 연다」 와 §4 시험 10(모호 → 다음 거절 주기 재시도) — **현재 코드에서 거짓**(Q2a: 막힌 동안 거절 주기가 오지 않음). 어느 선택지를 고르느냐에 따라 문장이 달라지므로 이번 회차에 고치지 않았다.
