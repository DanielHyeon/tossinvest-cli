# a197 재개 — Manager 결정 2 조사 보충(2차): 선택지 (E) · SQL 전수 · 시한 구조

> 작성 2026-11-25, 팀메이트(Opus), worktree `~/tossos-work/a197-impl`, HEAD `224fa5ca0`(1차 커밋 위, production Go · 시험 편집 0).
> 근거: `analysis/cited-ast/*.ast.json`(이번에 17 개 더함 — `README.md` §0.4), probe 원문 `analysis/probe/resume2-engine-probe-output.txt`
> (엔진 조립 `buildE2EStack` 위 관측 주기 끝까지) · `analysis/probe/resume2-execgw-probe-output.txt`(게이트웨이 범위), 원본은 같은 폴더
> `resume2-zz_*.txt`. 버리는 worktree `~/tossos-work/a197-resume2-probe` 에서 1 회씩 돌리고 지웠다. 합성 번호 `EXT-1` · 네트워크 0.
> 결정 문장은 쓰지 않는다 — 선택지와 평가만.

## 결론 먼저

1. **(E) 는 성립한다(손절 · 청산 기준).** IN_DOUBT 외부 매수 취소를 기존 원장 전이 `Attempt.ResolveUnresolved` 로 park 하면 다음 관측 주기에 손절이 실제로 나간다(엔진 probe: park 전 `broker_sells=0` → park 뒤 `broker_sells=1`). 같은 종목 flatten 의 취소 · 매도도 통과한다. 대사 지역 상태 · 대사 추적기 · `reconcile-resolve` · 편입은 park 된 CANCEL 을 읽지 않는다. 다른 문으로 손절 · 청산을 막는 경로는 이번 범위에서 못 찾았다.
2. **대가는 진입이다.** park 는 그 종목의 신규 진입을 운영자 해동(`tossctl engine attempt-resolve`, mutating)까지 막고, 재시작으로 풀리지 않는다. 해소기의 park(`Resolver.park`)를 쓰면 **계정 전체 진입**도 그 프로세스가 끝날 때까지 막힌다(해동으로 안 풀림 — probe).
3. **권고를 (C) → (E) 로 바꾼다.** 둘 다 Q3 의 손절 · 청산 칸에서 오늘보다 늦어지는 칸이 0 이지만, (E) 는 주문 경로 판정 본체(`unsettledFor`)와 정본 `order-execution :46` SHALL 을 건드리지 않고 엔진 조립 위 probe 로 끝까지 확인됐다. (C) 가 (E) 보다 나은 점은 「진입 차단 없음」 하나뿐이다. 조건: 진입 차단 범위를 **종목**으로 둘지(래치 없는 park) Manager 가 정하고, 정본 order-execution 의 park 관련 문단을 MODIFIED 한다(아래 E5).
4. SQL 전수(31 파일 146 줄 = 질의 자리 104): **바꿈 1 · 결과 바뀜/판정 무관 1 · 무관 102 · 미확인 0**. 바꾸는 하나는 `attemptSelect` 를 쓰는 `PendingAttempts`(IN_DOUBT → 종목 전체 차단)와 `UnresolvedAttempts`(park → 종목 노출 증가 차단)이다 — 이미 아는 두 가드이고, 그 밖에 결과가 바뀌는 판독은 없다.
5. 시한(r1 #4 방향): 단계를 **주기 끝**에서 돌리는 것은 구조상 가능하다(`ObserveOnce` 의 `defer endExitOrderCycle@482` 는 모든 `judge@542` 뒤). 다만 `Run` 이 `ObserveOnce` 반환 뒤에야 `Sleep(Interval)` 하므로 단계 시간만큼 **다음 주기 시작이 늦어진다**. 유계 부분의 합은 21.05 초이고 토큰 잠금 대기는 상한이 없다. 같은 주기의 지연을 다음 주기의 지연으로 옮길 뿐이다.

## E1. park 로 넘기는 기존 전이

| 전이 | 자리 · AST | 상태 조건 | 사유 상수 | 예약 | 전이 행 | 부수 효과 |
|---|---|---|---|---|---|---|
| `Attempt.ResolveUnresolved` | `resolution.go:77-84`(분기 0, `a.transition` call `:78`) | `from: [IN_DOUBT]` 만 | 기본 `ReasonResolutionExhausted`(원장 값 `in_doubt_unresolved` — probe) | 해제 안 함(`releasesReservations` `reservation_release.go:119-126` 은 NOT_DISPATCHED · FAILED_CONFIRMED 만 참). CANCEL 결정은 원래 예약이 없음 | `transition` 이 `attempt_transitions` 에 넣음(`durability.go:650`) | 없음(게이트 래치 없음) |
| `Resolver.park`(비공개) | `indoubt.go:372-384`(B1 `if@376` 쓰기 실패, B2 `if@379 r.Gate != nil`) | 위 함수를 부름(call `:376`) | 위와 같음 + 해소 결과 사유 `ReasonUnresolvedInDoubt` | 같음 | 같음 | **`r.Gate.Block(ReasonUnresolvedInDoubt, …)`** call `:380` — 계정 래치 |
| 세션 해소기의 PLACE park | `inDoubtSettle.resolve`(`in_doubt_settle.go:210-237`) → `Resolver.Resolve`(`indoubt.go:229-357`) → `r.park`(call `:299` · `:325`) | IN_DOUBT ∧ PLACE 만(sweep B5 `case@182`) | — | — | — | park 면 `noteParkOutsideExit`(call `:234`) critical |
| 기동 `resolveCancel` 의 park | `amend_indoubt.go:44-158` → `r.park`(call `:54` · `:61` · `:74` · `:122` · `:141` · `:149` · `:154`) | 기동만(`Recovery.Run` call `:300`) | — | — | — | 래치 |

- **IN_DOUBT CANCEL 에 그대로 쓸 수 있다**: `ResolveUnresolved` 는 종류를 보지 않고 `from: IN_DOUBT` 만 요구한다. probe(E1)에서 `CANCEL|IN_DOUBT` → `UNRESOLVED_IN_DOUBT|in_doubt_unresolved`. **새 전이는 필요 없다** — 원장 · 스키마 무편집, 공개 API `Journal.Resume` + `Attempt.ResolveUnresolved` 로 닿는다.
- 세션 중 CANCEL 을 park 하는 기존 **경로**는 없다. sweep 의 CANCEL 갈래는 `closeCancel` 만 부르고(B6 `case@184` → call `:196`), `Resolver.park` 는 비공개라 exit 루프에서 부를 수 없다. 계정 래치까지 같게 하려면 새 코드가 `EntryGate.Block`(`retry.go:558-565`)을 직접 불러야 하고, `ExitObserverOptions` 에는 진입 게이트 칸이 없다(`exitloop.go:171-254` rg 0). 그러면 `Context.ExitObserver` 배선 한 줄이 늘어난다(`exitwiring.go`, census 밖).
- 정본: park 는 지금 「해소 불능」, 즉 해소 절차가 끝내 증명하지 못한 결과로만 정의돼 있다(`order-execution/spec.md:44` · `:70` · `:141`). (E) 는 절차 없이 park 하는 새 길이다. `:126`(「종료는 판정이 아니다 — 취소를 이유로 park 금지」)은 엔진 정지로 해소가 취소된 경우라 (E) 와 다른 사건이다.

## E2. park 가 막는 것 전부

| # | 무엇 | 막는가 | 근거 |
|---|---|---|---|
| (a) | `checkSymbolFree` 의 미해소 갈래 | **그 종목 · 그 시장의 노출 증가 요청만** | B3 `if@814 !plan.raisesExposure` → 반환이 `UnresolvedAttempts`(call `:817`)보다 앞. 미해소 대조는 `attemptTargets`(`gateway.go:842-849`, 분기 1 — intent 종목 · 시장)로 거름. probe E2a: 같은 종목 진입 BUY `reason=unresolved_in_doubt broker_places=0`, 다른 종목 진입 BUY `CONFIRMED broker_places=1` |
| (b) | 계정 전체 진입 래치 | **`Resolver.park` 를 쓸 때만** | `indoubt.go:380` `r.Gate.Block`, 생산 Gate = 엔진 진입 게이트(`internal/app/engine/gateway.go:298-304`). 래치는 프로세스 메모리(`EntryGate.latches`)이고 `Clear` 는 「nothing automatic calls it」(`retry.go:567-568`). `Clear(ReasonUnresolvedInDoubt)` 호출자는 0(rg — `Recovery.Run` 의 `Clear@338` 은 `ReasonRecoveryIncomplete`). probe E2b: 해소기 park 뒤 `latched_unresolved=true`, `OperatorResolve` 뒤에도 `still_latched=true`. 정본 `:131` 「세션 중 park 는 … 계정 전체의 신규 진입을 그 엔진 프로세스가 끝날 때까지 막으며(운영자 해동은 … 이 래치는 풀지 않는다), 청산은 막지 않는다」. `ResolveUnresolved` 를 직접 쓰면 래치 없음 — 엔진 probe E2b `Entry.Blocks()` = `map[]` |
| (c) | exit 루프 · 손절 | **막지 않는다(실측)** | `clearTheSymbol` B3 `if@1662` 의 입력 `UnsettledOnSymbol` → `PendingAttempts`(`recovery.go:30-33`, 상태 4 개 — UNRESOLVED 없음). 엔진 probe: 주기 1(IN_DOUBT) `broker_sells=0`, park 뒤 주기 2 `broker_sells=1`, attempt 행 `PLACE|CONFIRMED` 추가. park 알림 경로(`exit_held_proposal.go:218`)와 발의 해제 판정(`exit_proposal_release.go:212`)은 **발의 intent 의** attempt 만 본다 — 취소 intent 는 발의 intent 가 아님 |
| (d) | 수동 flatten | **막지 않는다** | 같은 B3(취소 · 매도는 노출 증가 아님). probe E2d: flatten 게이트웨이의 같은 종목 취소 `CONFIRMED` · 매도 `CONFIRMED`(broker places 1 · cancels 1) |
| (e) | 대사 · 편입 | **막지 않는다(이번 범위)** | `LocalStateFromJournal`(`compare.go:193-238`, 분기 5)은 `HeldBySymbol` + `TrackedFillOrders` 만 읽음. `TrackedFillOrders`(`fills.go:1543-1693`)는 `kind IN ('PLACE','AMEND')`(`:1559`). 엔진 probe E2e: 지역 OpenOrders 에 `EXT-1` 없음, tracked = `[BO-1]`(손절 매도)만. 대사 추적기의 영구 승격은 수량 · missing-order 두 종류뿐(`mismatch.go:420,434`)이고 둘 다 LocalState 에서 나옴. `Diff.BlocksEntry` = 수량 ∨ missing-order(`compare.go:312-314`) — 외부 주문은 진입을 막지 않음(외부 매수 자체는 오늘도 `ExternalOrd` 로 보고만 됨). `reconcile-resolve` 는 같은 LocalState 대조로 판정하고 외부 주문은 거절 사유가 아님(`engine_reconcile.go:82-89`). 편입(`adoption.go`)에는 attempt 상태 · 진입 게이트 소비가 rg 0. `reconcile_states` 에 쓰는 자리는 대사 추적기뿐이라 park 된 CANCEL 이 그 행을 만드는 경로는 없음. **미측정**: 대사 루프 전체(Tracker.Observe)를 park 와 함께 구동한 probe 는 하지 않았다 — AST · rg 근거 |

남는 예외(손절이 막히는 경우 — (E) 로도 못 닫음): 취소가 IN_DOUBT 로 끝나지 않고 원장 쓰기 실패로 **DISPATCH_STARTED** 나 **ACKED** 에 남는 경우. `ResolveUnresolved` 는 IN_DOUBT 에서만 받고, 정본 `:126` 은 진행 중 전송의 세션 종결을 금지한다. 재시작까지 막히며(a198 §6-1·6-2 와 같은 성질), (C) 도 IN_DOUBT 만 면제하면 같다.

## E3. park 의 해제

- **엔진이 도는 동안**: 닫는 길이 없다. 세션 스윕은 `PendingAttempts` 만 읽는다(엔진 probe E3: park 뒤 sweep 무변, `pending=0 unresolved=1`). D10 은 IN_DOUBT 만 받는다(`CloseInDoubtCancelOnTargetClose` B4 `if@88`).
- **재시작**: 기동 해소는 park 를 다시 보지 않는다(`Recovery.Run` B3 `range@270` 의 입력이 `PendingAttempts` call `:266`). 래치 없는 park 라면 재시작 뒤에도 그 종목 진입 차단이 남는다(`checkSymbolFree` 가 매 요청 원장 `UnresolvedAttempts` 를 읽음). 계정 래치는 재시작으로 사라진다(메모리).
- **`reconcile-resolve`**: 거절 조건에 걸리지 않는다(LocalState 에 없음). 대신 park 를 풀지도 않는다.
- **운영자가 할 일(순서)**:
  1. 앱에서 그 종목의 그 매수(`EXT-1`)가 남아 있는지, 취소됐는지 확인한다.
  2. `tossctl engine attempt-resolve --attempt <id> --target FAILED_CONFIRMED|CONFIRMED --operator … --approval … --note …`(CONFIRMED 면 `--broker-order-id` 필수)를 실행한다. mutating 이라 사람이 직접 친다. 엔진 lock 이 없어 엔진을 세울 필요는 없다(`engine_attempt_resolve.go:55-71`, `ResolveParkedAttempt` `attempt_thaw_command.go:47-100` — park 가 아니면 B6 `if@69` 에서 거절, audit `:78` 이 `OperatorResolve` `:82` 보다 먼저, 발의 해제는 그 intent 가 발의일 때만이라 취소 intent 에는 무동작).
  3. 계정 래치를 건 park(`Resolver.park` 계열)였다면 엔진을 재시작한다. 해동으로는 안 풀린다(probe E3).
  - 주의: 명령 도움말의 FAILED_CONFIRMED · CONFIRMED 문장(「주문이 브로커에 없다/있다」)은 **발주** 기준이다. 취소 attempt 에서는 「취소가 성립하지 않았다/성립했다」로 읽어야 하고, 그 번역은 문서에 없다.

## E4. Q3 표 다시 — (E) 로(IN_DOUBT 가 되는 순간 그 자리에서 park + critical + (포지션, 주문 번호) 재개 금지)

| 경우 | 오늘(토글 꺼짐) | (E) | 오늘보다 늦어지는가 |
|---|---|---|---|
| 1. 사람이 나중에 그 매수를 앱에서 치운다 | 치운 뒤 다음 주기에 손절 | park 뒤 다음 주기에 손절 제출 → 매수가 살아 있으면 반대 미체결 확정 거절(오늘과 같음), 단계는 그 주문에 다시 열리지 않음 → 사람이 치운 뒤 다음 주기에 손절 | **아니오**(엔진 probe 주기 2) |
| 2. 취소가 실제로는 성공했다(5xx 로 기록) | 사람이 치울 때까지 | park 뒤 다음 주기에 손절 접수 | 아니오(빠름) |
| 3. 취소가 실패했고 매수가 살아 있다 | 사람이 치울 때까지 | 1 과 같음(재취소 없음) | 아니오(같음) |
| 비상 청산(수동 flatten) | 막히지 않음 | 막히지 않음(probe E2d) | 아니오 |
| **진입**(새 칸) | 막히지 않음 | 그 종목의 신규 진입이 **운영자 해동까지**(재시작 무관) 막힘. `Resolver.park` 계열이면 **계정 전체 진입**도 엔진 재시작까지 | **예 — 진입 한정**(불변식 4 의 대상 아님) |
| 예외: 원장 쓰기 실패로 DISPATCH_STARTED/ACKED 에 남음 | — | 재시작까지 막힘(E2 끝) | 예(드묾, (C) 도 같음) |
| 참고: 단계 시간 | — | 아래 「시한」 — 같은 주기 또는 다음 주기 | 예(별도 r1 #4) |

손절 · 청산 칸에서 오늘보다 늦어지는 칸은 **0**(IN_DOUBT 경로 기준, 예외 행 제외)이다.

## E5. (C) 대 (E)

| 항목 | (C) 판정 본체에서 외부 매수 IN_DOUBT 취소 면제 | (E) 그 자리에서 park |
|---|---|---|
| 고칠 기존 함수(FLM) | `Gateway.unsettledFor` + `Gateway.UnsettledOnSymbol` = **2**(주문 경로 판정 본체 — 진입 · flatten · exit 모두가 지남) | **0 추가**. park 는 a197 의 새 파일이 기존 공개 API(`Journal.Resume` · `Attempt.ResolveUnresolved`)를 부름. 계정 래치를 고르면 `Context.ExitObserver` +1(exitwiring.go, census 밖, pre-edit 번들 있음). 단계를 주기 끝으로 옮기면 `endExitOrderCycle` +1(아래 「시한」) — 두 안 공통 |
| MODIFIED 정본 | order-execution `:46` 「심볼당 in-flight 1개 — **모든 safety class**」(SHALL, 주문 경로 핵심) | order-execution 「IN_DOUBT 해소」 의 park 정의(`:44` 해소 불능 결과로만 → a197 취소의 즉시 park 추가), 「세션 중 해소: 대상」(`:125` — 엔진 IN_DOUBT 취소는 D10 또는 기동이 닫음), 래치 없는 park 를 고르면 `:131`(세션 park = 계정 래치) |
| 구조 시험 · census | a094 핀(`a094_exit_clear_test.go:560`, 호출 관계) 유지 가능 · a166 무관 | a094 핀 무관 · a166 무관 · a198 핀(`a198_clear_cancel_log_order_test.go:140`)은 로그 쓰기 자리만 봄 |
| 거부하게 되는 정상 입력 | 없음 | **그 종목의 신규 진입**(운영자 해동까지, 재시작 무관) — 래치 없는 park. 래치를 쓰면 **계정 전체 신규 진입**(엔진 재시작까지) 추가 |
| 통과시키게 되는 입력 | 그 종목의 손절 매도 · 길 치우기 취소(· 범위를 열면 flatten) — IN_DOUBT 취소가 원장에 남은 채 | 그 종목의 모든 노출 비증가 요청(손절 · 길 치우기 · flatten) — 기존 park 규칙 그대로, 새 판정 없음 |
| 이중 매도 논증 | 새로 증명할 것(1차 메모) — SQL 전수로 보강됨(아래) | 기존 park 의 의미(정본 `:131` 「청산은 막지 않는다」) 그대로 — 새 논증 없음. 단 「모르는 주문 위에서 매도를 낸다」는 사실은 같고, SQL 전수 결과도 같다 |
| 재시작 뒤 | IN_DOUBT 는 기동 `resolveCancel` 이 단건 판독으로 닫음(사람 일 0) | 기동은 park 를 안 봄 — **항상 사람이 해동** |
| Q3 손절 · 청산 칸 | 0(코드 근거) | 0(**엔진 조립 위 probe** 근거) |
| a087 · a167 충돌 | 없음 | 없음 |
| 운영자 부담 | 없음(재시작 시 자동) | 발생 1 건마다 앱 확인 + `attempt-resolve` 1 회(+ 래치면 재시작) |

### 권고 (다시)

**(E), 래치 없는 park(종목 범위 진입 차단)** 를 권한다. (C) 를 유지하지 않는 근거:

- (C) 의 유일한 이점은 「진입 차단 없음」과 「재시작이 자동으로 닫음」 이다. 그 대가로 모든 변경 요청이 지나는 판정 본체를 고치고(FLM 2, 매 요청 원장 귀속 판독 1 추가 — 비용 미측정), 주문 경로의 핵심 SHALL 을 좁힌다.
- (E) 가 (C) 에 못 미치는 것은 진입뿐이다. 진입은 불변식 4(손절 · 비상 청산 즉시성)의 대상이 아니고, 정본은 이미 「모르는 주문이 있는 종목에는 노출을 더하지 않는다」를 park 로 표현한다(`:44`). 손절 · 청산은 엔진 조립 위 관측 주기로 통과가 실측됐다.
- 래치 없는 쪽을 권하는 이유: 계정 래치는 엔진 재시작까지 **모든 종목**의 진입을 막고 해동으로 안 풀린다(probe). 외부 매수 하나의 취소 불명으로 계정 전체를 멈출 근거는 코드에서 찾지 못했다 — 「심볼당 유일 매칭」 근거(`:46`)도 종목 범위다. 다만 `:131` 이 「세션 park 는 계정 래치」 라고 적고 있어 MODIFIED 가 필요하다. 래치를 고르면 MODIFIED 는 하나 줄고 진입 차단은 커진다 — Manager 결정.
- 「재개 금지」 의 저장 자리: park 된 CANCEL 행 자체(target_order_id = 그 번호, state = UNRESOLVED)가 재시작에도 남는 표시가 될 수 있다(새 상태 · 스키마 불필요 — 그 행을 찾는 판독은 새 코드). 설계 결정은 아니고 재료가 있다는 사실만 적는다.

## (C) 조건 3 — `mutation_attempts` 판독 전수

- 방법: 비시험 Go 중 문자열 `mutation_attempts` 가 있는 31 파일 146 줄을 감싼 문자열 리터럴 단위로 묶어 질의 자리 **104** 개로 만들었다(스크립트로 뽑고, 필터 없는 자리는 원문을 열어 판정). 1차의 「28 파일 139 줄」 은 `internal/journal` 만 센 값이다 — 여기에 console 2 · protection 1 파일(7 줄)이 더해진다. 판정 단위: 「CANCEL IN_DOUBT 또는 UNRESOLVED, 대상 = 원장에 없는 외부 번호」 인 행이 그 판독의 **결과**를 바꾸는가.
- 요약: **바꿈 1**(#23 `attemptSelect` → `PendingAttempts` · `UnresolvedAttempts` — 알려진 두 가드) · **결과 바뀜/판정 무관 1**(#12 `FailedConfirmedCancelsForSymbol` — a197 취소가 확정 거절일 때 행이 들어오나 래치 판정이 엔진 미체결 대상만 봄) · **무관 102**(주석 · 문자열 8, 다른 표 `protection_mutation_attempts` 8, DDL · 목록 18, 이관 1, kind 가 CANCEL 을 빼는 판독 37, kind 는 CANCEL 이나 외부 대상에 해당 없음 2(#11 · #37), id · intent · decision · 예약 한정 28) · **미확인 0**.
- 한계: kind 필터가 있는 자리는 SQL 문장만 봤다(그 함수의 다른 판독은 따라가지 않음). id 한정 자리 중 호출자를 따라간 것은 `proveZeroFill` · `authoritativeCampaignOrderInTx` 둘이고, 나머지(transition · lineage · replay · strategy · nonce)는 「자기 attempt 행만」 으로 판정했다.

| # | 자리 | 함수 · 상수 | 종류 | kind 필터 | state 필터 | 판정 | 근거 |
|---|---|---|---|---|---|---|---|
| 1 | `internal/console/orders.go:783` | originOf | READ | — | — | 무관 | 주석·문서 문자열(질의 아님) |
| 2 | `internal/console/templates_orders.go:190` | const ordersTemplates = ` | READ | — | — | 무관 | 주석·문서 문자열(질의 아님) |
| 3 | `internal/journal/a137_daily_loss_rows.go:177` | engineSellOverlaps | READ | kind = 'PLACE' | — | 무관 | kind = 'PLACE' ∧ side SELL |
| 4 | `internal/journal/a137_daily_loss_rows.go:193` | engineSellOverlaps | DDL | kind IN ('CREATE', 'REPLACE') | 있음 | 무관 | 다른 표 `protection_mutation_attempts`(부분 문자열 일치) |
| 5 | `internal/journal/a198_cancel_target_close.go:59` | const cancelTargetOwners = allFillSnapsh | READ | kind IN ('PLACE','AMEND') | 있음 | 무관 | D10 소유 후보 CTE, kind PLACE/AMEND — 외부 대상은 0 이 되어 「닫지 않음」(1차 Q2b) |
| 6 | `internal/journal/a198_cancel_target_close.go:79` | CloseInDoubtCancelOnTargetClose | READ | — | — | 무관 | D10 자기 attempt(id) 판독·전이 — state IN_DOUBT 만 받음(B4@88), park 는 무변 |
| 7 | `internal/journal/a198_cancel_target_close.go:152` | CloseInDoubtCancelOnTargetClose | WRITE | — | 있음 | 무관 | D10 자기 attempt(id) 판독·전이 — state IN_DOUBT 만 받음(B4@88), park 는 무변 |
| 8 | `internal/journal/a199_order_path_evidence.go:30` | const confirmedPlaceQuery = `SELECT 1 FR | READ | — | — | 무관 | kind = PLACE(인자) ∧ CONFIRMED — a199 청산 전용 판정 |
| 9 | `internal/journal/a200_adjustment_close_fill_recognition.go:191` | readRecognitionFacts | READ | — | — | 무관 | intent_id 한정 ∧ kind PLACE/AMEND |
| 10 | `internal/journal/account_views.go:421` | BrokerOrderExitLinks | READ | kind='AMEND';kind='PLACE' | 있음 | 무관 | kind IN ('PLACE','AMEND') 또는 kind = PLACE/AMEND 인자 — CANCEL 행 제외 |
| 11 | `internal/journal/cancel_refusal_latch.go:23` | ConfirmedCancelAfterLatch | READ | kind = 'CANCEL';kind IN ('PLACE','AMEND') | 있음 | 무관 | kind CANCEL ∧ CONFIRMED · 대상의 소유 intent 정확히 1(owners.n = 1) — 외부 대상은 0 이라 후보 없음 |
| 12 | `internal/journal/cancel_refusal_latch.go:58` | FailedConfirmedCancelsForSymbol | READ | kind = 'CANCEL' | 있음 | 결과 바뀜·판정 무관 | kind CANCEL ∧ FAILED_CONFIRMED 종목 목록 — a197 취소가 확정 거절이면 행이 들어오나, 소비자 cancelRefusalLatch 가 대상이 엔진 미체결(LiveOrdersForSymbol)일 때만 래치(B5@25) — 외부 대상은 건너뜀, 또 노출 증가 요청에서만 불림(checkSymbolFree B3@814 뒤) |
| 13 | `internal/journal/daily_buy.go:247` | proveZeroFill | READ | — | — | 무관 | proveZeroFill(attempt id) — 후보가 예약 행에서 옴(daily_buy.go:205), CANCEL 결정은 예약 0 |
| 14 | `internal/journal/daily_buy.go:268` | proveZeroFill | READ | kind IN (?, ?) | 있음 | 무관 | attempt id 한정 ∧ kind IN (PLACE, AMEND) 인자 |
| 15 | `internal/journal/daily_buy.go:315` | proveZeroFill | READ | kind = ? | — | 무관 | kind = AMEND 인자 ∧ target_order_id — CANCEL 제외 |
| 16 | `internal/journal/daily_buy.go:352` | EngineBrokerOrderIDs | READ | — | — | 무관 | EngineBrokerOrderIDs — kind IN (PLACE, AMEND) 인자(daily_buy.go:353) |
| 17 | `internal/journal/durability.go:421` | Prepare | WRITE | — | — | 무관 | Prepare 의 INSERT — a197 취소 자신의 기록 |
| 18 | `internal/journal/durability.go:435` | Prepare | WRITE | — | — | 무관 | Prepare 의 INSERT — a197 취소 자신의 기록 |
| 19 | `internal/journal/durability.go:561` | transition | READ | kind = | — | 무관 | transition — id 로 자기 행만(park 전이 자체가 여기) |
| 20 | `internal/journal/durability.go:579` | transition | READ | kind = | — | 무관 | transition — id 로 자기 행만(park 전이 자체가 여기) |
| 21 | `internal/journal/durability.go:634` | transition | WRITE | — | 있음 | 무관 | transition — id 로 자기 행만(park 전이 자체가 여기) |
| 22 | `internal/journal/durability.go:736` | IntentAttempted | READ | — | — | 무관 | IntentAttempted(intent id) — 취소 intent 는 발의 intent 와 다름 |
| 23 | `internal/journal/durability.go:756` | const attemptSelect = `SELECT id, intent | READ | — | — | 바꿈 | attemptSelect — 소비자 PendingAttempts(state IN RECORDED·DISPATCH_STARTED·ACKED·IN_DOUBT)와 UnresolvedAttempts(state = UNRESOLVED) 가 이 투영을 씀: IN_DOUBT 면 같은 종목 모든 변경 차단(1차 Q2a), park 면 같은 종목 노출 증가만 차단(2차 E2a). 그 밖의 소비자(LookupAttempt 등)는 id 한정 |
| 24 | `internal/journal/execution_contract.go:167` | const schemaV5 = ` | DDL | kind IN ('OPEN_EXPOSURE','DAILY_LOSS','CASH');kind IN ('RISK | 있음 | 무관 | 스키마 DDL · 외래 키 · 읽기 전용 표 목록(행 판독 아님) |
| 25 | `internal/journal/execution_contract.go:250` | const schemaV5 = ` | DDL | — | — | 무관 | 스키마 DDL · 외래 키 · 읽기 전용 표 목록(행 판독 아님) |
| 26 | `internal/journal/execution_contract.go:255` | const schemaV5 = ` | DDL | — | — | 무관 | 스키마 DDL · 외래 키 · 읽기 전용 표 목록(행 판독 아님) |
| 27 | `internal/journal/execution_contract.go:256` | const schemaV5 = ` | DDL | — | — | 무관 | 스키마 DDL · 외래 키 · 읽기 전용 표 목록(행 판독 아님) |
| 28 | `internal/journal/execution_contract.go:257` | const schemaV5 = ` | DDL | — | — | 무관 | 스키마 DDL · 외래 키 · 읽기 전용 표 목록(행 판독 아님) |
| 29 | `internal/journal/execution_contract.go:258` | const schemaV5 = ` | DDL | — | — | 무관 | 스키마 DDL · 외래 키 · 읽기 전용 표 목록(행 판독 아님) |
| 30 | `internal/journal/execution_contract.go:263` | const schemaV5 = ` | DDL | — | — | 무관 | 스키마 DDL · 외래 키 · 읽기 전용 표 목록(행 판독 아님) |
| 31 | `internal/journal/execution_contract.go:264` | const schemaV5 = ` | DDL | — | — | 무관 | 스키마 DDL · 외래 키 · 읽기 전용 표 목록(행 판독 아님) |
| 32 | `internal/journal/execution_contract.go:265` | const schemaV5 = ` | DDL | — | — | 무관 | 스키마 DDL · 외래 키 · 읽기 전용 표 목록(행 판독 아님) |
| 33 | `internal/journal/execution_contract.go:266` | const schemaV5 = ` | DDL | — | — | 무관 | 스키마 DDL · 외래 키 · 읽기 전용 표 목록(행 판독 아님) |
| 34 | `internal/journal/execution_contract.go:268` | const schemaV5 = ` | DDL | — | — | 무관 | 스키마 DDL · 외래 키 · 읽기 전용 표 목록(행 판독 아님) |
| 35 | `internal/journal/execution_contract.go:271` | const schemaV5 = ` | DDL | — | — | 무관 | 스키마 DDL · 외래 키 · 읽기 전용 표 목록(행 판독 아님) |
| 36 | `internal/journal/execution_contract.go:277` | const schemaV5 = ` | DDL | — | — | 무관 | 스키마 DDL · 외래 키 · 읽기 전용 표 목록(행 판독 아님) |
| 37 | `internal/journal/exit_proposal_release.go:156` | ConfirmedCancelOf | READ | kind = 'CANCEL' | 있음 | 무관 | ConfirmedCancelOf — kind CANCEL, 호출자가 엔진 미체결 주문 번호로만 부름(clearTheSymbol :1673) |
| 38 | `internal/journal/exit_proposal_release.go:253` | intentOrdersAwaitingCloseQ | READ | kind IN ('PLACE','AMEND') | 있음 | 무관 | kind IN ('PLACE','AMEND') 또는 kind = PLACE/AMEND 인자 — CANCEL 행 제외 |
| 39 | `internal/journal/fills.go:579` | confirmedFillOwners | READ | kind = 'AMEND';kind IN ('PLACE','AMEND') | 있음 | 무관 | kind IN ('PLACE','AMEND') 또는 kind = PLACE/AMEND 인자 — CANCEL 행 제외 |
| 40 | `internal/journal/fills.go:907` | const allFillSnapshotsCTE = `WITH all_fi | READ | kind = 'AMEND';kind IN ('PLACE','AMEND') | 있음 | 무관 | kind IN ('PLACE','AMEND') 또는 kind = PLACE/AMEND 인자 — CANCEL 행 제외 |
| 41 | `internal/journal/fills.go:949` | const allFillSnapshotsCTE = `WITH all_fi | READ | kind = 'AMEND';kind IN ('PLACE','AMEND') | 있음 | 무관 | kind IN ('PLACE','AMEND') 또는 kind = PLACE/AMEND 인자 — CANCEL 행 제외 |
| 42 | `internal/journal/fills.go:1131` | FillEventsScoped | READ | kind IN ('PLACE','AMEND') | 있음 | 무관 | kind IN ('PLACE','AMEND') 또는 kind = PLACE/AMEND 인자 — CANCEL 행 제외 |
| 43 | `internal/journal/fills.go:1259` | accountRefForOrder | READ | kind IN ('PLACE','AMEND') | 있음 | 무관 | kind IN ('PLACE','AMEND') 또는 kind = PLACE/AMEND 인자 — CANCEL 행 제외 |
| 44 | `internal/journal/fills.go:1349` | ExecutionCorrectionsScoped | READ | kind IN ('PLACE','AMEND') | 있음 | 무관 | kind IN ('PLACE','AMEND') 또는 kind = PLACE/AMEND 인자 — CANCEL 행 제외 |
| 45 | `internal/journal/fills.go:1405` | const orderIDCanonicalScopesSQL = ` | READ | kind IN ('PLACE','AMEND') | 있음 | 무관 | kind IN ('PLACE','AMEND') 또는 kind = PLACE/AMEND 인자 — CANCEL 행 제외 |
| 46 | `internal/journal/fills.go:1478` | NetPositions | READ | kind IN ('PLACE','AMEND') | 있음 | 무관 | kind IN ('PLACE','AMEND') 또는 kind = PLACE/AMEND 인자 — CANCEL 행 제외 |
| 47 | `internal/journal/fills.go:1557` | TrackedFillOrders | READ | kind = ?;kind IN ('PLACE','AMEND') | 있음 | 무관 | kind IN ('PLACE','AMEND') 또는 kind = PLACE/AMEND 인자 — CANCEL 행 제외 |
| 48 | `internal/journal/fills.go:1701` | guardTrackedFillIdentity | READ | kind = ?;kind IN ('PLACE','AMEND') | 있음 | 무관 | kind IN ('PLACE','AMEND') 또는 kind = PLACE/AMEND 인자 — CANCEL 행 제외 |
| 49 | `internal/journal/fills.go:1865` | LiveOrdersForSymbol | READ | kind IN ('PLACE','AMEND') | 있음 | 무관 | kind IN ('PLACE','AMEND') 또는 kind = PLACE/AMEND 인자 — CANCEL 행 제외 |
| 50 | `internal/journal/fills.go:2023` | const schemaV16 = ` | DDL | — | — | 무관 | 스키마 DDL · 외래 키 · 읽기 전용 표 목록(행 판독 아님) |
| 51 | `internal/journal/fills.go:2128` | const schemaV19 = ` | WRITE | kind IN ('PLACE','AMEND') | 있음 | 무관 | 이관(schemaV19) 일회성 백필, kind PLACE/AMEND |
| 52 | `internal/journal/lineage.go:140` | resolveWithLineage | WRITE | — | — | 무관 | resolveWithLineage — id 한정(정정 lineage) |
| 53 | `internal/journal/lineage.go:157` | resolveWithLineage | WRITE | — | — | 무관 | resolveWithLineage — id 한정(정정 lineage) |
| 54 | `internal/journal/lineage.go:207` | resolveWithLineage | WRITE | — | — | 무관 | resolveWithLineage — id 한정(정정 lineage) |
| 55 | `internal/journal/lineage.go:379` | scopedLineageChildrenV16 | READ | kind = ? | 있음 | 무관 | kind IN ('PLACE','AMEND') 또는 kind = PLACE/AMEND 인자 — CANCEL 행 제외 |
| 56 | `internal/journal/lineage.go:402` | legacyScopedLineageChildren | READ | kind = ? | 있음 | 무관 | kind IN ('PLACE','AMEND') 또는 kind = PLACE/AMEND 인자 — CANCEL 행 제외 |
| 57 | `internal/journal/nonce.go:53` | consumeDecisionNonce | READ | — | — | 무관 | consumeDecisionNonce — attempt id 한정 |
| 58 | `internal/journal/position_campaign.go:142` | const schemaV20 = ` | WRITE | — | 있음 | 무관 | 스키마 DDL · 외래 키 · 읽기 전용 표 목록(행 판독 아님) |
| 59 | `internal/journal/position_campaign.go:984` | ApplyPositionCampaignFill | READ | kind IN ('PLACE','AMEND') | 있음 | 무관 | kind IN ('PLACE','AMEND') 또는 kind = PLACE/AMEND 인자 — CANCEL 행 제외 |
| 60 | `internal/journal/position_campaign.go:1660` | campaignExposureBlockedInTx | READ | kind IN ('PLACE','AMEND') | 있음 | 무관 | kind IN ('PLACE','AMEND') 또는 kind = PLACE/AMEND 인자 — CANCEL 행 제외 |
| 61 | `internal/journal/position_campaign.go:1684` | authoritativeCampaignOrderInTx | READ | — | — | 무관 | authoritativeCampaignOrderInTx — id 한정, 호출자 2(LinkCampaignOrder · strategy dispatch) |
| 62 | `internal/journal/position_projection.go:142` | resolveFillOrigin | READ | kind IN ('PLACE','AMEND') | 있음 | 무관 | kind IN ('PLACE','AMEND') 또는 kind = PLACE/AMEND 인자 — CANCEL 행 제외 |
| 63 | `internal/journal/protection_orders.go:37` | const schemaV13 = ` | DDL | kind IN ('CREATE','REPLACE','CANCEL') | 있음 | 무관 | 다른 표 `protection_mutation_attempts`(부분 문자열 일치) |
| 64 | `internal/journal/provenance.go:22` | ? | READ | — | — | 무관 | 주석·문서 문자열(질의 아님) |
| 65 | `internal/journal/provenance.go:218` | const provenanceQuery = `, | READ | kind IN ('PLACE','AMEND') | 있음 | 무관 | kind IN ('PLACE','AMEND') 또는 kind = PLACE/AMEND 인자 — CANCEL 행 제외 |
| 66 | `internal/journal/readonly.go:71` | var ErrSchemaTooOld = errors.New("journa | READ | — | — | 무관 | 주석·문서 문자열(질의 아님) |
| 67 | `internal/journal/readonly.go:79` | var readOnlyTables = []string{ | READ | — | — | 무관 | 스키마 DDL · 외래 키 · 읽기 전용 표 목록(행 판독 아님) |
| 68 | `internal/journal/reconcile_order_terminal_release.go:212` | orderHasReplacementTrace | READ | kind = ? | — | 무관 | kind = AMEND 인자 ∧ target_order_id — CANCEL 제외 |
| 69 | `internal/journal/replay.go:103` | MarkReplayStarted | WRITE | — | — | 무관 | 재생(replay) — id 한정, PLACE 재생 전용 경로 |
| 70 | `internal/journal/replay.go:123` | MarkReplayStarted | WRITE | — | — | 무관 | 재생(replay) — id 한정, PLACE 재생 전용 경로 |
| 71 | `internal/journal/replay.go:151` | RefundReplay | WRITE | — | 있음 | 무관 | 재생(replay) — id 한정, PLACE 재생 전용 경로 |
| 72 | `internal/journal/reservation_release.go:152` | releaseReservationsForOrder | READ | — | 있음 | 무관 | 예약 해제(broker_order_id ∧ intent) — CANCEL 결정은 예약 0 |
| 73 | `internal/journal/reservation_release.go:229` | alertsForOrder | READ | kind IN ('PLACE','AMEND') | 있음 | 무관 | kind IN ('PLACE','AMEND') 또는 kind = PLACE/AMEND 인자 — CANCEL 행 제외 |
| 74 | `internal/journal/reservation_release.go:538` | sweepOrphanedTerminals | READ | — | 있음 | 무관 | sweepOrphanedTerminals — 예약 행에서 출발, CANCEL 결정은 예약 0 |
| 75 | `internal/journal/reservation_release.go:557` | sweepOrphanedTerminals | READ | kind IN ('PLACE','AMEND') | 있음 | 무관 | kind IN ('PLACE','AMEND') 또는 kind = PLACE/AMEND 인자 — CANCEL 행 제외 |
| 76 | `internal/journal/reservation_release.go:750` | ReservationsAwaitingOperator | READ | kind IN ('PLACE','AMEND') | 있음 | 무관 | kind IN ('PLACE','AMEND') 또는 kind = PLACE/AMEND 인자 — CANCEL 행 제외 |
| 77 | `internal/journal/risk_bucket_fill.go:84` | RegisterRiskBucketOrder | READ | kind IN ('PLACE','AMEND') | 있음 | 무관 | kind IN ('PLACE','AMEND') 또는 kind = PLACE/AMEND 인자 — CANCEL 행 제외 |
| 78 | `internal/journal/risk_bucket_owner.go:935` | releaseRiskBucketOwner | READ | — | 있음 | 무관 | 다른 표 `protection_mutation_attempts`(부분 문자열 일치) |
| 79 | `internal/journal/risk_bucket_owner.go:979` | releaseRiskBucketOwner | WRITE | — | — | 무관 | 다른 표 `protection_mutation_attempts`(부분 문자열 일치) |
| 80 | `internal/journal/risk_bucket_owner.go:1049` | unresolvedScopedMutation | READ | kind IN ('PLACE','AMEND') | 있음 | 무관 | kind IN ('PLACE','AMEND') 또는 kind = PLACE/AMEND 인자 — CANCEL 행 제외 |
| 81 | `internal/journal/schema.go:191` | var migrations = []migration{ | WRITE | — | — | 무관 | 주석·문서 문자열(질의 아님) |
| 82 | `internal/journal/schema.go:224` | const schemaV1 = ` | DDL | — | — | 무관 | 스키마 DDL · 외래 키 · 읽기 전용 표 목록(행 판독 아님) |
| 83 | `internal/journal/schema.go:269` | const schemaV1 = ` | DDL | — | — | 무관 | 스키마 DDL · 외래 키 · 읽기 전용 표 목록(행 판독 아님) |
| 84 | `internal/journal/strategy_dispatch_runtime.go:777` | beginStrategyDispatchTx | WRITE | — | 있음 | 무관 | strategy dispatch — id · decision · client_order_id · intent 한정(전략 경로) |
| 85 | `internal/journal/strategy_dispatch_runtime.go:787` | beginStrategyDispatchTx | WRITE | — | 있음 | 무관 | strategy dispatch — id · decision · client_order_id · intent 한정(전략 경로) |
| 86 | `internal/journal/strategy_dispatch_runtime.go:890` | transitionStrategyAttemptTx | READ | — | — | 무관 | strategy dispatch — id · decision · client_order_id · intent 한정(전략 경로) |
| 87 | `internal/journal/strategy_dispatch_runtime.go:901` | transitionStrategyAttemptTx | READ | — | — | 무관 | strategy dispatch — id · decision · client_order_id · intent 한정(전략 경로) |
| 88 | `internal/journal/strategy_dispatch_runtime.go:928` | transitionStrategyAttemptTx | WRITE | — | 있음 | 무관 | strategy dispatch — id · decision · client_order_id · intent 한정(전략 경로) |
| 89 | `internal/journal/strategy_dispatch_runtime.go:956` | loadStrategyDispatchAttemptAuthority | READ | — | — | 무관 | strategy dispatch — id · decision · client_order_id · intent 한정(전략 경로) |
| 90 | `internal/journal/strategy_dispatch_runtime.go:1507` | recoverablePreparedStrategyAttempt | READ | — | — | 무관 | strategy dispatch — id · decision · client_order_id · intent 한정(전략 경로) |
| 91 | `internal/journal/strategy_dispatch_runtime.go:1521` | recoverablePreparedStrategyAttempt | READ | — | — | 무관 | strategy dispatch — id · decision · client_order_id · intent 한정(전략 경로) |
| 92 | `internal/journal/strategy_lineage.go:583` | RecoverStrategyDispatch | READ | — | 있음 | 무관 | strategy dispatch — id · decision · client_order_id · intent 한정(전략 경로) |
| 93 | `internal/journal/trade_outcome_refusal.go:285` | priceLots | READ | — | — | 무관 | 주석·문서 문자열(질의 아님) |
| 94 | `internal/journal/trade_outcome_refusal.go:349` | instanceSellIntegrity | READ | kind IN ('PLACE','AMEND') | 있음 | 무관 | kind IN ('PLACE','AMEND') 또는 kind = PLACE/AMEND 인자 — CANCEL 행 제외 |
| 95 | `internal/journal/trade_outcomes.go:320` | roundTripLegs | READ | kind IN ('PLACE','AMEND') | 있음 | 무관 | kind IN ('PLACE','AMEND') 또는 kind = PLACE/AMEND 인자 — CANCEL 행 제외 |
| 96 | `internal/journal/trade_outcomes.go:362` | roundTripLegs | READ | kind IN ('PLACE','AMEND') | 있음 | 무관 | kind IN ('PLACE','AMEND') 또는 kind = PLACE/AMEND 인자 — CANCEL 행 제외 |
| 97 | `internal/journal/trade_outcomes.go:477` | roundTripLegs | READ | — | — | 무관 | 주석·문서 문자열(질의 아님) |
| 98 | `internal/journal/trade_outcomes.go:480` | roundTripLegs | READ | — | — | 무관 | 주석·문서 문자열(질의 아님) |
| 99 | `internal/journal/trade_outcomes.go:605` | engineSellOrders | READ | kind = ?;kind IN ('PLACE','AMEND') | 있음 | 무관 | kind IN ('PLACE','AMEND') 또는 kind = PLACE/AMEND 인자 — CANCEL 행 제외 |
| 100 | `internal/journal/trade_outcomes.go:688` | sumSellFills | READ | kind IN ('PLACE','AMEND') | 있음 | 무관 | kind IN ('PLACE','AMEND') 또는 kind = PLACE/AMEND 인자 — CANCEL 행 제외 |
| 101 | `internal/protection/repository.go:44` | const schemaDDL = ` | DDL | kind IN ('CREATE','REPLACE','CANCEL') | 있음 | 무관 | 다른 표 `protection_mutation_attempts`(부분 문자열 일치) |
| 102 | `internal/protection/repository.go:194` | recordAttempt | WRITE | — | — | 무관 | 다른 표 `protection_mutation_attempts`(부분 문자열 일치) |
| 103 | `internal/protection/repository.go:209` | markAttempt | WRITE | — | — | 무관 | 다른 표 `protection_mutation_attempts`(부분 문자열 일치) |
| 104 | `internal/protection/repository.go:225` | Attempts | READ | — | — | 무관 | 다른 표 `protection_mutation_attempts`(부분 문자열 일치) |

## 시한(r1 #4 방향) — 구조 확인

- **① 주기 끝 실행 — 구조상 가능.** `ObserveOnce`(`exitloop.go:480-552`)는 첫 문장에서 `defer o.endExitOrderCycle(ctx)`(`defer@482`)를 걸고 포지션 순회(B5 `range@524`) 안에서 `judge`(call `:542`)를 부른다. 따라서 `endExitOrderCycle`(`exit_order_log.go:235-255`, 분기 4) 안이나 `ObserveOnce` 꼬리(`reportSourceAge@548` 과 같은 자리 — a163 선례)에 둔 단계는 그 주기의 모든 포지션 판정 · 제출 뒤에 돈다.
  - 편집 표면: `endExitOrderCycle` 에 한 줄을 넣는 것이 a198 D4 의 선례다(그 함수의 a198 핀은 `flushClearCancelFailures` 호출 존재만 본다). `ObserveOnce` 를 고치면 큰 High-risk 함수의 FLM 이 필요하다(a111 핀은 `MaxExitObservationCycle` 호출 수만 봄). `exit_order_log.go` 는 a087 번들의 파일이지만, a198 종결 기록대로 그 해시는 이미 어긋나 있다.
  - 조건: 판정 경로(B15)에서는 「거절이 났다」만 대기열에 넣는다 — a198 의 `noteClearCancelFailure` → `flushClearCancelFailures` 모양. 이른 반환 경로(`Deferred` · outage · 빈 작업 집합)에서도 defer 는 돌지만, 대기열이 비어 있으면 무동작이다.
- **② 주기당 외부 취소 1 건 상한 — 가능**(새 코드의 대기열 길이 · 반복 상한, 기존 함수 편집 아님).
- **그때 한 주기에 더해지는 최악 시간**(취소 1 건):

  | 구성 | 상수 · 값 | 근거 |
  |---|---|---|
  | 미체결 목록 판독 | **코드 상수 없음** — a197 design D4 가 제안한 3 s(설계 값). 단계 ctx 시한이 없으면 HTTP 클라이언트 `defaultTimeout` 15 s × 쪽 수(`ScanOrders` maxPages ≤ 0 → 50) | `official/client.go:22`, `indoubt.go:730-765` |
  | 사전 단건 조회 | `precheckTimeout` = 2 s | `precheck.go:58·99` |
  | 취소 POST | `defaultTimeout` = 15 s | `official/client.go:22` |
  | 계좌 예산 간격 | `interval` = 350 ms × 요청 3 = 1.05 s(최소 간격만, 대기열 경합은 미측정) | `accountbudget/budget_linux.go:27` |
  | 401 재전송 | 최대 2 회 더(취소 POST 기준 +30 s) — a198 §2-4 인용, 이번에 코드 미확인 | a198 design §2-4 |
  | 토큰 잠금 대기 | **상한 없음**(`m.mu.Lock()`, ctx 로 끊기지 않음) | `official/token.go:63-65` |
  | **합(유계 부분)** | **3 + 2 + 15 + 1.05 = 21.05 s**(401 재전송 포함 51.05 s) + 토큰 잠금(무한) | |

- **다음 주기 시작을 얼마나 늦추나**: `Run`(`exitloop.go:417-428`)은 `ObserveOnce`(call `:422`, defer 포함)가 반환한 뒤 `cycleComplete`(`:423`) → `Sleep(Interval)`(`:424`, `DefaultExitObservationInterval` = 5 s `:99`)을 한다. 그래서 단계 시간 T 만큼 다음 주기 시작이 그대로 밀린다(주기 간격 5 s → 5 + T s, 유계 최악 26.05 s). ① 은 **같은 주기의 뒤 포지션 지연을 다음 주기 모든 포지션의 지연으로 옮긴다** — 손절 판정의 총 지연은 줄지 않는다.
- **대안(불가능하지 않으나 ① 로 해결되지 않는 부분)**: 단계를 exit 루프 밖의 **보조 실행자**로 돌리는 안이다. 선례는 세션 해소 실행자 `inDoubtSettle` — `AuxiliaryExecutor`(`in_doubt_settle.go:137`), 「손절 판정 goroutine 밖에서 돎」(같은 파일 머리 주석). exit 루프는 대기열에 넣기만 하므로 루프 지연이 0 이다. 대가로 취소가 전송 중인 동안 다음 주기의 손절은 `checkSymbolFree` B2 `symbol_mutation_in_flight` 로 그 종목만 한 주기 미뤄진다(엔진 취소가 전송 · 인수 단계면 연속 실패 계수 제외 — `onlyEngineCancelsInFlight` B5). design D5 가 「별도 실행자로 떼지 않는다」 로 기각한 안이라 재결정 대상이다. 참고로 오늘의 길 치우기 취소(`clearTheSymbol` call `:1715`)도 같은 비용을 **판정 경로 안에서** 치르고 있다(a198 §2-4).

## 실행하지 않은 범위 · 미측정

- 대사 루프(`Tracker.Observe`) · 편입 루프를 park 와 함께 구동한 probe 없음(AST · rg 근거만).
- 계정 래치 park 를 엔진 조립 위에서 구동하지 않음(게이트웨이 단위 probe 와 정본 `:131` 만).
- 401 재전송 횟수 · 예산 대기열 경합 · 토큰 잠금 실제 대기 — 코드 미확인 또는 미측정.
- 전체 시험 · make test/gate/lint · sdd-sync/sdd-check · 적대 리뷰 · CodeGraph 재질의 — 실행 안 함.
