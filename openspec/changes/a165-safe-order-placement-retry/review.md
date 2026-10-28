# a165 review

## 1.3 proposal-freeze 리뷰 — 2026-10-23 — **FAIL (blocker 3) → §2 미착수**

- 대상: `design.md`(첫 판), `specs/order-execution/spec.md`(MODIFIED 첫 판), `analysis/`(0.2–0.4 산출물). base `34e12967e`, 리뷰 시점 HEAD `5975af2b9`(base 뒤 `.go` 변경 0).
- 보이스(작성 컨텍스트와 분리된 서브에이전트 2, 둘 다 읽기 전용 — 파일 편집·`make`·`tossctl` 0):
  - **A — 적대적 Eng / 레드팀**(중복 주문 반례 10종: 키·본문 변경, 8s 창 밖, 조회-재전송 경합, 재시작, tracker 의미, D4 강등, 409, strategy 경로, hybrid fallback). 판정 **FAIL (blocker 1)**.
  - **S — 안전·범위(CEO/eng-manager + §0 불변식)**: 사용자 결정 충실도, D7 즉시성, delta 의 MODIFIED 정합, YAGNI, WORKFLOW 단계. 판정 **FAIL (blocker 2)**.
- 처분: 발견 전부를 `design.md` D13 표에 옮겼다. 사실 오류 둘(S1 의 D7 전제, S11 의 Amend caller)과 D4 의 거짓 문장(S3)은 이번 판에서 **정정**했다. 나머지는 D7 사람 결정 뒤
  design·delta 를 다시 써서 재리뷰한다 — 결정에 따라 갈래 자체가 바뀌기 때문이다(예: 429 갈래를 a166 전까지 끄면 A3 가 사라진다).

### 발견 요약과 처분

| id | 보이스 | 심각도 | 요약 | 처분 |
|---|---|---|---|---|
| A1 | A | BLOCKER | 재전송 ACK → CONFIRMED 인데 첫 요청이 따로 만든 주문이 추적 밖에 남을 수 있음. 「실패 응답 뒤 같은 키 멱등 동작」 미측정(R11 은 2xx 직후만) | 수용 — D13, 잔여 이름 붙임(issues.md), 사람 수락 필요 + 정착 전 목록 대조 설계 |
| S1 | S | BLOCKER | D7 사실 전제 오류: `flatten.Saga` 생성은 CLI(`cmd/tossctl/flatten.go:247`) 하나뿐, CLI flatten 의 409 재시도는 오히려 빠르게 함(오늘 `resolveStep` 동기 해소 45s+), 엔진 청산 루프만 「그 종목↑ / 뒤 종목↓」 교환 | 수용 — D7 경로별 표로 재작성(이번 판). 내부 분기 주장은 AST 번들 전 「미검증」 표지 |
| S2 | S | BLOCKER | D7 사람 결정 미결 — 어느 범위도 불변식 3 을 한 방향으로 증명 못 함(B 도 ①에서 증가, C 도 ④로 0 아님) | 수용 — 아래 「사람 결정 요청」 |
| S3 | S | MAJOR | D4 「ACK 쪽으로만」 거짓 — 전송 전 실패(오늘 NOT_SENT) → 재전송 → 전송 후 끊김이면 IN_DOUBT(비관 전이), 같은 종목 보유 시 청산 치움 45s+ 막힘 | 수용 — D4 문장 정정(이번 판), 질문 4 |
| A2 | A | MAJOR | `ErrServer`·상태 없는 `ErrTransport`·echo 불일치 평문 오류에 **2xx 로 접수된 응답**이 섞여 (b) 재전송을 탐 | 수용 — D13, (b) 한정 조건 재설계 |
| A3/S6 | A,S | MAJOR | 429 재전송·(b) 조회가 계좌 공유 패널티를 늘릴 수 있음, (b) CLOSED walk 가 1.3 확정 행(1페이지)과 충돌 | 수용 — D13, 질문 5 |
| A4/S4 | A,S | MAJOR | delta 의 「재생과 재시도의 구분」 시나리오가 예외 2 와 모순 | 수용 — 재작성 때 수리 |
| S5 | S | MAJOR | 「IN_DOUBT 해소」(정체 모른 재주문 금지·재생 attestation 전제)와의 관계가 delta 에 없음 | 수용 — 「IN_DOUBT 해소」도 MODIFIED 에 넣음 |
| S7 | S | MAJOR | 「429 는 Retry-After 상한 준수」 미처리, 「즉시」→backoff 이탈 미기재 | 수용 |
| S8/A10 | S,A | MAJOR | FLM 누락: 조회 경로·`unwrapAndDecode`·`tokenManager.refresh`·`refusalCodes`·`trading.Service.Cancel`·D7 의 청산/flatten 함수 | 수용 — 0.4 를 `[ ]` 로 되돌리지 않고 추가 번들 task 로 남김(tasks.md 0.4.1) |
| A5 | A | MINOR | Amend 엔진 생산 호출자 0, 재전송이 체결 전 스냅숏 수량을 보냄 | 수용 — 범위 제외 권고(질문 6) |
| A6 | A | MINOR | 창 검사 자리·wall-clock·ctx 판정 | 수용 — D13 |
| A7 | A | MINOR | 「T1 ≤ 15s」 거짓(401 루프 3회 + timeout 없는 토큰 교환 hc). 별건: `strategyDispatchOwnerTakeoverGrace` 20s 전제가 이미 깨짐 | 수용 — D13, 별건은 issues.md |
| A8 | A | MINOR | 매처 quantity 결손이 불일치 쪽 | 수용 — D13 |
| A9 | A | MINOR | `error.data.retryAfterSeconds` 경로 | 수용 — D13 |
| S9 | S | MINOR | tasks 미체크·proposal 머리말 낡음·review.md 없음·issues 절 없음·전부 untracked | 수리 — 이번 커밋 |
| S10 | S | MINOR | 허가형 SHALL·시나리오 누락 | 수용 — 재작성 때 |
| S11 | S | MINOR | Amend caller `ops.modifyHandler` 는 이름 충돌 | 수리 — code-context 두 파일 정정 |
| A11·S12 | A,S | NOTE | HTTP/2 내부 재전송 가능성, 되돌리기 = 재빌드(KST 05–09), 전략 cycleDeadline 30s 여유 | 기록 |
| A-통과 | A | — | 깨지지 않은 반례: 키·본문 결정적(struct 마샬, `formatDecimal`, 결정적 confirm token, B32 재검사), 창 밖 타임아웃 재전송 불가(T1 ≥ 8s), 재시작 뒤 재전송 없음(`RecoverPending` → IN_DOUBT, `Attested: nil`), D4 강등은 ACK 경로 외 낙관 종결 없음, `strategyCrossedTransport` 는 콜백 시작에서 true, 엔진은 hybrid fallback 없음 | — |

### 사람 결정 요청 (S·A 질문 통합 — D7)

1. **엔진 청산 루프의 청산 전 취소가 `409 already-processing` 을 받으면 같은 주기 안에서 재시도할까?**
   이득: 그 종목 손절이 IN_DOUBT 해소(최소 45s) 대신 보통 ~1s 뒤(R9). 대가: 같은 주기 뒤 포지션이 실패 1건당 최대 `min(누적 5s + 2×(GET+POST), 8s−T1)` 늦어지고,
   ~2건이 쌓이면 시세 lease 15s 를 넘어 다음 주기로 밀린다. 대안 D: 그 자리에서 기다리지 않고 409 `already-processing` 취소를 「확정 미수락」으로 분류해 다음 주기(5s)에
   다시 취소 — 뒤 지연 0, 단 사용자 결정 2 의 「retryAfter 를 따른다」 문언에서 벗어난다. 축소안 B′: 1회·retryAfter ≤1s.
2. **CLI `tossctl flatten` 의 취소에도 409 재시도를 배선할까?** 오늘은 막힌 취소 1건마다 45s~5분을 그 자리에서 동기 해소한다 — 붙이면 비상 청산이 빨라진다.
3. **청산 매도 접수(엔진 청산·CLI flatten)의 재시도는 0회로 유지할까?** 근거: R1(짧은 간격 429 재시도는 회복 못 함).
4. **진입 접수 재시도에 「전송 전 실패」 갈래를 남길까?** 오늘 NOT_SENT(심볼 즉시 풀림)로 끝날 결과를 IN_DOUBT 로 바꿀 수 있고, 같은 종목 보유 시 그 손절이 45s+ 막힌다.
5. **a166(계좌 공유 3 req/s)을 a165 보다 먼저 착지시킬까, 아니면 a166 착지 전까지 429 갈래를 끌까?** 그리고 (b) 조회를 OPEN·CLOSED 각 1페이지로 고정할까(1.3 확정 행)?
6. **정정(Amend)을 a165 범위에서 뺄까?** 엔진 생산 호출자 0(YAGNI).
7. **「실패 응답(5xx·429·전송 후 끊김) 뒤 같은 키 재요청의 멱등 동작은 미측정」 잔여를 수락하는가?** 수락해도 A1 수리(재전송 ACK 뒤 정착 전 목록 대조, 일치 정확히 1·같은 id 일 때만 CONFIRMED)는 넣는다.

### 1차 발견의 2판 처분 (사용자 2차 결정 뒤)

| id | 2판 처분 |
|---|---|
| A1 | 수리 — design D6 R2(재전송 ACK 의 정착 전 목록 대조), delta 「IN_DOUBT 해소」 MODIFIED. 잔여 R-a165-3 사용자 수락 |
| S1·S2 | 해소 — 사용자 결정 1~3 으로 감소 경로 place 재시도 없음, 취소는 두 경로 모두 오늘보다 빨리 끝남(design D7) |
| S3 | 소멸 — 전송 전 실패 갈래 제외(결정 4) |
| A2 | 수리 — D2: 5xx 는 `err == official.ErrServer` 동일성, 전송 끊김은 `GotFirstResponseByte` 전, echo 불일치 제외 |
| A3/S6 | 429 부분 소멸(결정 5). 조회 OPEN·CLOSED 각 1페이지 고정(D6) |
| A4/S4 | 수리 — 「재생과 재시도의 구분」 시나리오에 취소 409 를 확정 거절로 명시 |
| S5 | 수리 — 「IN_DOUBT 해소」 MODIFIED |
| S7 | 소멸 — 429 갈래 없음 |
| S8/A10 | 수리 — 0.4.1 번들 14(합계 23) |
| A5 | 소멸 — Amend 제외(결정 6) |
| A6·A7 | 수리 — D5 두 겹 창(주입 시계 판정 + 단조 `winCtx` 상한), 창 검사는 `plan.call` 바로 앞, ctx 판정 `dctx.Err()` |
| A8 | 수리 — D6 수량 결손 = 일치, (b) 는 비분할 LIMIT 만 |
| A9 | 수리 — `error.data.retryAfterSeconds`, code 판정은 `refusalCodes` |
| S10 | 수리 — 「한다(SHALL)」 형, 시나리오 추가 |
| A11·S12 | 기록 — 잔여 R-a165-4, D11 |
| 신규 | strategy first-leg 는 재전송 범위 밖(D9, 보수 — 오늘 동작) |

## 1.3 재리뷰(2판) — 2026-10-23 — **FAIL (blocker 3, 고유 3) → §2 미착수**

- 대상: design 2판 · delta 2판 · 23 번들. HEAD `6baa60a78`. 보이스 둘(새 컨텍스트, 읽기 전용): **A2 — 적대 Eng 레드팀**(FAIL, blocker 2), **S2 — 안전·범위**(FAIL, blocker 2). 파일 편집·`make`·`tossctl` 0.
  A2 가 실행한 기존 시험 `go test ./internal/flatten -run 'TestARefusedCancelIsNotAFailure|TestUnresolvedCancelHoldsItsSymbol'` PASS.

| id | 보이스 | 심각도 | 요약 | 처분(3판 계획) |
|---|---|---|---|---|
| F1 | A2·S2 | **BLOCKER** | flatten 취소 409 소진 → FAILED_CONFIRMED → `classifyCancel` B3(flatten.go:458) `FlattenStepFailed` → `UnsettledCancelSymbols`(journal/flatten.go:326–328)가 안 셈 → `planLiquidation`(liquidate.go:281–319)이 **살아 있는 주문 옆에서 전량 매도** — oversell 가드(`TestUnresolvedCancelHoldsItsSymbol`) 우회. 기존 시험 flatten_test.go:487–489 가 「FAILED_CONFIRMED = 주문이 살아 있지 않음」을 고정한다. already-processing 의 실측 의미는 「주문이 아직 처리 중 = 살아 있음」(M16) | 수용 — flatten 정책은 소진 시 **오늘 경로(AMBIGUOUS → `resolveStep`)** 로 둔다. 확정 미수락 분류(R3)는 엔진 정책에만. 시험 「소진 → 그 종목 Held」 추가. 엔진 쪽은 두 보이스 모두 안전 확인(`clearTheSymbol` B11 1697, `LiveOrdersForSymbol` 은 취소 FAILED_CONFIRMED 로 주문을 지우지 않음 — fills.go:1859–1882) |
| F2 | A2 | **BLOCKER** | (b) 조회·R2 를 1페이지로 판정하는데 1페이지 완전성 근거 없음 — openapi 는 CLOSED 정렬 방향을 정하지 않고, `From` 미지정이면 전 기간. 첫 주문이 즉시 체결돼 CLOSED 깊이에 묻히면 사전 부재 → 재전송 → R2 「1건 = ACK id」 → CONFIRMED, 첫 주문 추적 밖 | 수용 — `OrderQuery.From` 에 첫 전송 KST 거래일(−1일 여유), 페이지가 `HasNext=true` 이고 일치가 없으면 「모름」(재전송 없음·R2 실패). CLOSED 정렬은 [미측정] — 사람 측정 1회 권고 |
| F3 | S2 | **BLOCKER(사람 결정)** | **R1·R2 는 생산 도달 0** — 생산 `Place` 호출자 4곳 중 매수는 `tracer.go:442`(non-test 생성자 0)와 `strategydispatch/adapters.go:112`(`GatewayAdapter` 생산 배선 0)뿐이고, 생산 전략 진입은 `PlaceClaimedStrategy`(strategy_dispatch_cycle.go:206)인데 D9 가 뺐다 | **사람 결정 필요** — 아래 |
| F4 | A2·S2 | MAJOR | `err == official.ErrServer` 는 「주문 POST 가 5xx」가 아니다 — 토큰 교환 non-200 이 bare `ErrServer`(token.go:152–156), SendTracker·`GotFirstResponseByte` 는 ctx 전체라 토큰 교환·accounts GET 의 끊김도 (b) 로 셈 → 주문 바이트 0 인 실패가 「재전송」(결정 4 위반, 중복 위험은 0) | 수용 — 판정을 주문 POST 자체에 귀속(official 에 요청별 신호), 토큰 5xx·account-seq 끊김 시험 |
| F5 | A2·S2 | MAJOR(사실) | 엔진에는 세션 중 IN_DOUBT 해소기가 없다 — `Resolver.Resolve` 는 부팅 복구(runtime.go:289–296 → recovery.go:300)와 flatten 뿐. design 의 「해소 최소 45s」 거짓 — 오늘 엔진 409 취소는 **재시작·운영자 해소까지** 그 종목 손절을 막는다(R3 의 이득은 더 큼). R2 실패(목록 지연)도 올바른 ACK 뒤 재시작까지 차단 | 수용 — D6·D7 문장 정정 |
| F6 | S2 | MAJOR | R2 강도 과장(지연 사례에서 첫 주문을 못 봄 — R-a165-3 아래 잔여) | 수용 — 한계를 잔여·delta 에 명시, R2 조회를 창 끝 쪽으로 늦추는 완화 검토 |
| F7 | S2 | MAJOR(절차) | FLM 누락: `tokenManager.token/exchange`, `Resolver.Resolve`/`resolveStep`, `Saga.planLiquidation`, `journal.UnsettledCancelSymbols`, `Client.PlaceOrder`, `classifyRefusalCode`. 생성기 머리말의 coverprofile 범위 표기가 flatten·engine 번들과 안 맞음 | 수용 |
| F8 | S2·A2 | MINOR | D9 경계 `plan.strategy == nil` 이 `GatewayAdapter.PlaceStrategyEntry`(strategy nil 인 Gateway.Place)를 못 막음 — 배선되면 리스 안에서 8s 늘어남 | 수용 — 명시적 opt-in 으로 |
| F9 | A2·S2 | MINOR | POST 회계(401 루프 포함 최악 주문 POST 9), D10 산식 불일치, R3 재취소 N 종목 상한(≈0.4 req/s/종목, 경보 `noteDelay` 유지), delta 시나리오 공백(flatten 재시도 성공·소진, R2 통과, place·amend 409 불변, MARKET·분할·strategy 제외, 주문 전 5xx), 「비분할」 spec 누락, 401 재전송의 기록 면제 명시, proposal What Changes·tasks 2.2 의 1판 잔재, R2 의 읽을 수 없는 행, `roundTripFor` 가 바깥 ctx(상한 정정), D1 인용 645→646 | 수용 |
| F10 | S2 | NOTE | 재전송이 자기 claim·예약에 막히지 않는지 RED 양성 대조군 필요 | 수용 |

깨지지 않은 반례(A2): 키·본문 결정성, 응답 받은 뒤 실패 3종((b) 아님), 재전송 확정 거절 강등, 재시작, PARTIAL_FILLED dedup, HTTP/1.1 비재전송·HTTP/2 안전 재전송, `dctx.Err()`, 엔진 R3, flatten R4 무강등, 감소 place 제외, claim 유지, 예약 재검사, 오류 감싸기 없음.

### 사람 결정 요청 (F3 — 이것 하나)

R1·R2(진입 결과 불명 재전송 + 정착 전 목록 대조)는 지금 설계대로면 **생산에서 한 번도 돌지 않는다**. 생산 매수 진입은 전략 claimed 경로(`PlaceClaimedStrategy`)뿐인데 D9 가 리스·EntryGate 잠금 문제로 그 경로를 뺐다.

1. **R1·R2 를 a165 에서 뺀다(권고).** a165 = 엔진 취소 409 확정 미수락(R3) + flatten 취소 409 재시도(R4, 소진 시 오늘 경로) + 401 근거 시험(R5). 진입 재전송은 전략 claimed 경로와 함께 후속 change 로 설계한다(F2·F4·F6 도 그쪽으로).
2. **꺼진 채로 만든다** — 생산 도달 0 을 문서에 밝히고 배선은 후속. High-risk 검증 비용(변이·적대 리뷰)은 그대로 든다.
3. **D9 를 철회** — `PlaceClaimedStrategy`(EntryGate 잠금·SUBMITTING 리스 안)까지 재전송을 넓혀 새로 설계한다.

## 1.3 3차 proposal-freeze 리뷰(3판) — 2026-10-28 — **PASS** (P0 0 · P1 2 이번 커밋 수리 · P2 3 · P3 7)

- 대상: design 3판(`bc56522c7`) · delta 3판 · 23 번들. 리뷰 HEAD `0c9bfd272`(origin/main), worktree `TossOS.worktrees/docs-a165`, 브랜치 `docs/a165-review3`.
- 보이스: 독립 적대 리뷰어 1(새 컨텍스트, 작성 세션과 분리) — **적대 Eng 레드팀**(oversell·지연·오분류 반례)과 **안전·범위**(불변식 3·4·5·6, 사람 결정 충실도) 점검표를 한 보이스가 함.
  autoplan 의 CEO·design·DX 관점은 not-applicable — 3판에 UI·CLI 표면·제품 범위 변경이 없다(사람 결정으로 범위가 이미 정해짐).
- 소스 편집·시험 실행·운영 접근 0. 실행한 것: `go run ./tools/logic-map`(AST 8개), `check_analysis.py --change`(읽기 전용 판정), `openspec validate --strict`.
- 사람 결정(2026-10-28 「권장대로」): **H1** R4 최악 지연 유지 · **H2** gate ⑤ 창 — `--record-landing` 먼저, 안 되면 archive 번들 복사, base 재고정 없음(issues.md).

### 새 AST 증거 (`analysis/evidence/`, HEAD `0c9bfd272`)

번들 밖 함수의 분기를 근거로 쓰므로 먼저 뽑았다: `journal--journal.liveordersforsymbol` · `execgw--classifybrokerrefusal` · `execgw--classifyrefusalbody` ·
`execgw--gateway.checksymbolfree` · `execgw--gateway.unsettledfor` · `journal--releasesreservations` · `journal--exitintentattemptsq` · `engine--onlyenginecancelsinflight` (`*.ast.json`).

### (a) 2차 FAIL F1~F10 처분의 3판 반영

| id | 3판 | 판정 |
|---|---|---|
| F1 | D3 「최종」: 소진·중단·재전송 NOT_SENT → 첫 409 의 AMBIGUOUS → `DispatchVerified` B16(204) IN_DOUBT → `classifyCancel` B5(464) → `settleStep` B9(395) `resolveStep`. `resolveStep` 의 결과는 B4(434) Done · B5(437) **Pending**(취소 미실행 관측 = 보류) · B6(443)/B1/B2 InDoubt 뿐이고 Failed 가 없다 → `UnsettledCancelSymbols`(Pending·InDoubt·Held) → `planLiquidation` B5(307) 보류 | **반영** (단 배선 하나가 이것을 무너뜨릴 수 있음 → P1-1) |
| F2·F4·F6 | 진입 재전송과 함께 a171 로 이관(a171 proposal 표 24–28행 확인) | 반영(범위 밖) |
| F3 | 사용자 3차 결정 =1, design §0 | 반영 |
| F5 | D1: 세션 중 해소기 없음. `Resolver.Resolve` 생산 호출자 = `reconcile/recovery.go:300`·`flatten.go:428` 둘뿐(grep 재확인) | 반영 |
| F7 | 남는 범위가 기대는 함수 번들화(23). `tokenManager.token/exchange` 는 3판이 내부를 주장하지 않아 불필요 | 반영 |
| F8 | place 재전송이 없어 소멸 | 반영 |
| F9 | 401 기록 면제(delta)·D5 인용 646 정정·flatten 소진 시나리오. flatten 성공 시나리오는 이번 커밋에서 범위 명시로 보강(P1-2) | 반영 |
| F10 | D10 「1회 409 뒤 ACK → CONFIRMED」가 재전송이 자기 결정·예약 재검사(B31·B36)에 막히지 않음의 양성 대조군 | 반영 |

### (b) 갈래표 → FLM 반환 지점 (AST 좌표)

**R3 엔진**: 409 AP(코드 일치·모순 없음) + 엔진 정책 → 새 함수가 REJECTED → `DispatchVerified` B6(158)→B7(160) FAILED_CONFIRMED → `submit` B57(758) 반환 759 →
`clearTheSymbol` B11(1697) `cleared=false; continue` → B13(1706) 반환 1707. 다음 주기 B3(1639) 는 종결 attempt 를 안 셈(`unsettledFor` 는 원장의 미종결 목록만 돈다 — B2(36)) → 재취소.
아니면(코드 불일치 `classifyRefusalCode` B6(74)/반환 75 · 모순 B3(63)/64 → `classifyMutation` B5(52)/53 · 코드 없음 → `classifyMutation` B6(66) `ClassifyBrokerRefusal`
(본문 표지 없으면 B10(182) 거짓, 반환 188) → B8(84) `statusOf` B1(127)/128 → `ClassifyHTTPMutation` B7(337)/338) AMBIGUOUS → B16(204) IN_DOUBT — 오늘과 같음.
**매도는 B11 이 참일 때만 막히고 거짓일 때만 나간다** — R3 는 `cleared` 를 참으로 만드는 경로를 하나도 더하지 않는다(확인).
`LiveOrdersForSymbol`(1854) 는 SQL CTE 가 `kind IN ('PLACE','AMEND')`·CONFIRMED 만 보므로 취소 FAILED_CONFIRMED 는 주문을 닫지 않는다(분기 B1~B7 은 오류·스캔뿐).

**R4 flatten**: 첫 결과 비 AP → 그대로. AP → 루프(누적 5s·창 8s·횟수 2·Sleep 실패에서 중단). 재전송은 `submit` 의 결정 재읽기 B30(648)·B31(651)·키 B32(654)·보호 B33(658)·
진입 B34(665)·예약 B36(673)·ctx B37(685) 를 다시 지남(거절은 `notSent` → D3 규칙으로 AMBIGUOUS). 재전송 ACK → B8(165)…B15(199) CONFIRMED → `classifyCancel` B2(456) Done.
재전송 확정 거절(404·422 등) → B6(158) → `classifyCancel` B3(458) Failed → 매도 진행 — **오늘 첫 전송 4xx 와 같은 해석**(새 가정 아님). 재전송 모호 → IN_DOUBT → `resolveStep`.
**R5**: `send` B4(344) 루프 조건 `attempt<2 && code==401` · B5(347)/348 refresh 오류 · B7(355)/356 전송 오류 · B8(358) `!adopted` break · B9(362)/363 비 2xx · 365 정상.
`refresh` B1(119) 채택 121 / 교환 125. `postWithHeaders` B1(431) 본문 1회 마샬(435 클로저가 같은 `encoded` 재사용) → 바이트 동일.

### (c) 불변식 판정

| 불변식 | 판정 | 근거 |
|---|---|---|
| 3 토글 OFF = upstream | **PASS** | 런타임 토글 없음. 바뀌는 경로는 TossOS 전용(엔진 게이트웨이, `flatten-all` — 2026-07-26 `2910ddfbd`·`4734c5dfe` 추가). upstream `tossctl order cancel`(`order.go:295` → `tradingService` 직행)·`ops/write_operations.go:214` 와 공식 클라이언트(R5 는 시험만)는 무변. `Options.CancelRefusal == nil` = 오늘 동작(D5·D8) |
| 4 손절·비상 청산 즉시성 | **엔진: PASS** / **CLI flatten: 사람 수락 PASS(H1)** | 엔진 청산 전 취소 409 AP 는 같은 주기 추가 지연 0(대기·재전송 0, 시험 고정)이고, 오늘의 「재시작까지 그 종목 손절 막힘」(D1)을 없앤다. flatten 은 흔한 경우 단축(동기 해소 ≥45s → ~1s)이나 **최악은 막힌 취소 1건마다 오늘 경로 앞에 ≤8s, N건이면 N×8s 를 더한다 — 무조건 PASS 가 아니다.** 그 최악은 그 뒤의 `resolveStep`(≥45s)이 지배하고, 사람이 H1 로 수락했다 |
| 5 High-risk 의무 | **PASS(조건: P1-1 반영)** | 23 번들 유효((f)), Pre-Edit(2.1)·RED/GREEN·변이 원장(3.1)·다중 적대 리뷰(3.2) 계획 있음 |
| 6 손절·익절·사이징 보수 방향 | **해당 없음 / PASS** | 수준·수량·사이징 변경 0. 매도는 여전히 확정 취소(`clearTheSymbol` B11) 또는 oversell 가드(`planLiquidation` B5) 뒤에서만 |
| WORKFLOW §0.4 rate 예산 | PASS | D9 계상(엔진 ≈0.4 req/s/막힌 종목, flatten 취소당 ≤(GET+POST)×2). GREEN 에서 정본 §0.4 갱신 |

### (d)(e)(f)(g)

- (d) **공유 `definitiveRefusalCodes`(refusal_code.go:46)를 넓히지 않는다** — D2 는 새 순수 함수, place·amend 409 AP 는 오늘 그대로. 단 문서에 명시 문장이 없었음 → P3-3 로 D10 에 추가.
- (e) 정책 nil = 오늘 동작: D5 세 조건(nil · 비취소 · 전송 없음) → 첫 분류 그대로, D6 detail 바이트 동일, D10 시험. 통과. 편집 경계는 P2-2.
- (f) **23 번들 유효**: base `34e12967e`→HEAD 116 커밋 중 번들 소스 파일을 만진 것은 0(`git log --full-history` 대상: `internal/{execgw,official,flatten,trading}` 전체와 journal `flatten.go`·`dispatch.go`·`fills.go`·`classify.go`, engine `exitloop.go`·`gateway.go`, `cmd/tossctl/flatten.go` —
  걸린 것은 a169 `064f67297` 하나로 `internal/official/stock_reads.go`·a169 시험 파일뿐, 번들 소스 아님). 23 개 `source_sha256` 전부 HEAD 소스와 일치(실측). 다만 gate ⑤ 창 자체는 29건으로 불었다 → P2-1.
- (g) **a087 과의 `exitloop.go`**: a165 3판은 `exitloop.go` 를 **편집하지 않는다**(R3 은 게이트웨이 분류 + `engine/gateway.go:311` 배선 한 줄). 텍스트 충돌 0. a087 Phase 2 는 `ExitObserver.sellIntent` 를 고친다(a087 tasks 2.0~2.2, 미착수).
  계획: ① a087 이 먼저 착지하면 a165 의 `clearTheSymbol` 번들(`revision: current`)은 줄 이동으로 sha 가 낡는다 → a165 GREEN 전에 HEAD 에서 재추출하고 좌표(B3 1639·B11 1697)를 재대조.
  ② a087 의 `sellIntent` 편집은 a165 gate ⑤ 창의 형제 함수가 된다(H2 와 같은 처리). ③ 의미 충돌 없음 — a087 은 매도 주문 유형만 바꾸고 취소 분류에 기대지 않는다.
  ④ a167(잔량 재가격: 취소 → 재제출)이 엔진 취소를 쓰면 R3 의 「취소 FAILED_CONFIRMED = 주문 살아 있음」을 따라야 한다 — a167 착수 때 재대조.

### 발견

| id | 심각도 | 요약 | 처분 |
|---|---|---|---|
| P1-1 | **P1** | 취소 FAILED_CONFIRMED 의 뜻이 경로마다 반대다 — 엔진(R3) 「주문 살아 있음」, flatten `classifyCancel` B3(458) 「주문 없음」. 그래서 `buildFlattenWiring` 의 정책 필드 하나가 oversell(F1)을 가른다. Saga 시험은 자기 게이트웨이를 만들어 이 배선 변이(flatten ← `EngineCancelRefusal`)를 못 잡고, D10 변이 원장에도 없었다 | **수리(이번 커밋)** — design D10 「배선 시험」, tasks 3.1 에 배선 교차 둘·nil 둘 추가 |
| P1-2 | **P1** | delta 의 시나리오 「재시도 상한 소진 → 실패로 기록」·「일시적 거절 후 성공(정정 포함)」이 적용 대상을 말하지 않아 게이트웨이 flatten(소진 = IN_DOUBT)·엔진(정정 미적용)과 모순 — 그대로 구현하면 F1 이 돌아올 수 있다 | **수리(이번 커밋)** — 두 시나리오에 「검증 도구」/「CLI flatten 게이트웨이」 명시, strict validate 통과 |
| P2-1 | P2 | H2 의 전제가 낡았다 — gate ⑤ 창의 형제 함수는 1건이 아니라 **29건**(console 23·campaignstateevidence 2·a112 도구 3·adaptStocks 1). `--record-landing` 은 창의 끝만 좁히므로 base~착지 사이 형제는 남을 것으로 본다(예상) | issues.md 「H2 실측」. 2.1 에 처리 계획, GREEN 뒤 도구로 확인. 빨리 닫을수록 싸다 |
| P2-2 | P2 | D5 의 `sendOnce` 추출(646–725)은 strategy 갈래(B38–B46)까지 옮겨 High-risk 진입 경로를 리팩터한다. 더 작은 대안: 기존 `dispatch` 클로저를 **그대로** 두고 바깥 래퍼가 취소·정책·첫 결과 409 AP 일 때만 같은 클로저를 `winCtx` 로 다시 부른다 — B37(685) `dctx.Err()` 가 이미 `plan.call` 직전 마지막 검사이고, 결정·보호·진입·예약 재검사도 그대로 다시 돈다. 클로저 본문 편집 0 | 2.1 Pre-Edit 에서 선택. 어느 쪽이든 `winCtx` 는 재전송·대기에만 쓰고 정착(`DispatchVerified` 의 Settle/MarkInDoubt)은 바깥 ctx(P3-7) |
| P2-3 | P2 | R3 의 FAILED_CONFIRMED 는 그 종목의 `SymbolInFlight`(`checkSymbolFree` B2(804), `unsettledFor`)도 푼다 — 노출을 늘리는 진입도 그 종목에 다시 가능해진다. 오늘은 IN_DOUBT 가 재시작까지 진입까지 우연히 막았다. 나머지 FAILED_CONFIRMED 소비자(`releasesReservations` B2(121) — 취소 결정은 예약 없음 · `exitIntentAttemptsQ` B8(211) — 발의 intent 만 봄 · `onlyEngineCancelsInFlight` B6(72) · `ConfirmedCancelOf` CONFIRMED 만)는 취소 FAILED_CONFIRMED 가 오늘도 4xx 로 생기므로 새 상태 조합이 아니다 | 2.1 에서 「청산 중 종목의 진입은 어느 게이트가 막는가」를 적고 수락하거나 시험으로 고정 |
| P3-1 | P3 | D1 이 `classifyMutation` B6(66) `ClassifyBrokerRefusal`(본문 표지 `interactive`·`입금` 등)을 건너뛴다. AP 본문에 표지가 있으면 오늘은 REJECTED(flatten 에서는 Failed = 매도). AP 판정을 원 오류에서 분류보다 **먼저** 하도록 순서를 적을 것 | 2.1 |
| P3-2 | P3 | D3 「종목은 UnsettledCancelSymbols 에 남아」는 부정확 — `resolveStep` B4(434) Done(취소 실행 관측)이면 매도된다(올바름) | 기록 |
| P3-3 | P3 | `definitiveRefusalCodes` 불변이 문서에 없었다 | 수리(이번 커밋, D10·3.1) |
| P3-4 | P3 | R3 재취소는 막힌 종목마다 5s 에 결정·intent·attempt 1벌(시간당 ≈720행)과 경보 연속 계수를 만든다. 안전 문제는 아니고 운영 비용 | 기록 |
| P3-5 | P3 | R-a165-1 이 a165 잔여(issues)와 a171 이관(D7) 양쪽에 적힘 — 둘 다 맞게 「a165 잔여, a171 에도 기록」으로 읽는다 | 기록 |
| P3-6 | P3 | (g) a087 계획이 design 에 없었다 | 위 (g) 로 기록 |
| P3-7 | P3 | `winCtx` 가 만료된 뒤 정착을 같은 ctx 로 하면 기록이 실패한다 — 재전송 ctx 와 정착 ctx 분리를 Pre-Edit 에 적을 것 | 2.1 |
| P3-8 | P3 | tasks 0.4.1·1.1~1.3 의 낡은 문구(조회 경로 번들·「2판」·「F3 사람 결정 대기」) | 수리(이번 커밋) |

깨지지 않은 반례: R3 가 매도를 내보내는 경로(없음 — B11) · R4 소진·중단·전송 전 실패가 Failed 로 닫힘(없음 — D3 + `resolveStep` 에 Failed 반환 없음) · 재전송 중 프로세스 재시작(attempt DISPATCH_STARTED → 복구 IN_DOUBT, 취소라 중복 생성 없음) ·
place·amend 409 AP 의 분류 변경(없음 — kind 검사) · 엔진 지연(대기 0) · 재전송 사전 확인 GET 오류(엔진 `officialBroker` 의 `GetOrderAvailableActions` 는 오류를 내지 않음 — broker.go:57, flatten 도 같은 `engine.NewOrderPath`).

### 구현 착수 시 편집될 기존 함수(FLM 대상 — 모두 번들 있음)

`execgw.Gateway.submit`(dispatch 경계 — P2-2 선택에 따라 범위가 다름) · `execgw.New`(정책 필드 저장) · `engine.buildGateway`(필드 1) · `cmd/tossctl.buildFlattenWiring`(필드 1).
새 함수(`alreadyProcessingRefusal`·재시도 루프·`EngineCancelRefusal`·`FlattenCancelRefusal`)는 새 파일에 둔다. 기존 시험 파일의 fake·헬퍼를 고치면 그것도 gate ⑤ 대상이다.
`classifyMutation`·`classifyRefusalCode`·`Gateway.Cancel`·`Client.send` 는 편집 대상이 아니다(편집하면 번들 갱신 필요).

### 변이 원장 항목(3.1)

code 판정 EqualFold→status 만 · 두 자리 모순 검사 제거 · 409 검사 제거 · kind==Cancel 검사 제거(place·amend 에 적용) · Definitive 분기 뒤집기 · 누적 MaxWait 상한 제거 · 8s 창 검사 제거 ·
횟수 경계(2→3) · retryAfter 기본값 1s 변경 · NOT_SENT→첫 AMBIGUOUS 되돌림 제거 · 소진 시 마지막 결과로 정착 · 재전송 경계 재검사(ctx/deadline) 제거 · 엔진 정책 Retries 0→2 ·
**배선 교차 flatten←엔진 정책 / 엔진←flatten 정책 / 배선 nil 둘**(P1-1) · **`definitiveRefusalCodes` 에 already-processing 추가**(P3-3) · R5: `attempt<2`→`<3` · `code==401`→`code>=400` · `makeReq` 가 본문을 새로 마샬.
양성 대조군과 「변이가 대상 줄에 닿음」 표식 필수.

### 실행하지 않은 것

시험 전부(지시상 금지), `make sdd-sync`·`make gate`, 실계좌·운영 접근, CodeGraphContext·GBrain(advisory). `--record-landing` 효과는 GREEN 전이라 추정(P2-1).

## 1.3 2차 독립 리뷰 (GPT-5.6 Terra, 2026-10-28)

### 결론: **FAIL (구현 전 blocker 2)**

R3의 매도 방지·R4의 소진 시 IN_DOUBT 복귀·R5의 동일 본문은 3판 근거와 맞는다. 그러나 2026-10-28 사람 결정 P2-2·P2-3은 현재 design/FLM에 아직 반영되지 않았고, 특히 P2-3은 현재 공용 판정을 재사용하면 exit도 막히거나 차단 자체가 사라진다. blocker를 설계·FLM·RED에 반영하기 전 코드 착수 금지.

- 범위·증거: 23 FLM/BTM/위험 보고서와 `analysis/evidence` AST 8개를 읽었다. 분기 좌표는 아래 AST를 우선했고, `cmd/tossctl/flatten_test.go`의 현 시험 형식은 **손으로 읽음**이다.
- 실행 0: 소스·시험 편집 0, 시험·gate·운영 접근 0.
- skill 적용: `ai-memory-retrieval`은 a165 직접 결정 페이지를 찾지 못해 현재 HEAD/OpenSpec을 정본으로 삼았다. `search-first`는 새 코드·의존성·통합이 없는 독립 리뷰라 N/A. `grill-me`/`grilling`은 사람 결정이 이미 확정돼 재질문 권한이 없는 리뷰라 N/A. `impeccable: N/A — backend/documentation-only`.

### 확인된 안전 성질

- **중복 주문/oversell:** R4 대상은 `plan.kind == Cancel`뿐이고, `Gateway.Cancel`은 같은 `intent.OrderID`를 닫는다(`internal/execgw/gateway.go:402-421`, 손으로 읽음). `Gateway.submit` AST B30-B37은 재호출 때도 결정 재읽기·키 일치·보호·진입·예약·`dctx.Err()`를 `plan.call` 직전에 수행한다. 재전송 소진/중단/NOT_SENT를 첫 409의 AMBIGUOUS로 되돌리면 `DispatchVerified` B16(204)→IN_DOUBT, `classifyCancel` B5(464), `resolveStep` B4-B6(434-443), `UnsettledCancelSymbols` B2(335)·`planLiquidation` B5(307) 순으로 매도를 보류한다. 이 갈래는 AST 확인.
- **R3 청산 지연/매도:** 엔진 정책의 409 AP를 FAILED_CONFIRMED로 정착하면 `clearTheSymbol` B11(1697)이 `cleared=false`로 남겨 같은 주기 매도를 내지 않는다. 다음 주기에는 `unsettledFor`가 `PendingAttempts`만 읽으므로(B1(32), B2(36), evidence AST), 재취소가 가능하다. 재시작 중 DISPATCH_STARTED는 `RecoverPending`이 IN_DOUBT로 바꾸므로 재전송 중복 생성 대신 fail-closed 차단이다(`internal/journal/recovery.go:74-124`, 손으로 읽음).
- **토글 OFF/CLI upstream:** 새 배선은 엔진 `buildGateway`와 `flatten-all`에 한정되고, `Gateway.Options.CancelRefusal == nil`은 첫 dispatch를 그대로 한 번 돌려야 한다. `tossctl order` 경로·공식 클라이언트는 R5 시험 외 변경 대상이 아니다. 이 판정은 현 design D8과 호출 소스를 대조했다.
- **R5 동일 바이트:** `Client.send` AST B4(344)만 401에 재요청하고, `postWithHeaders`는 JSON을 한 번 마샬한 `encoded`를 두 `makeReq`에 재사용한다(`internal/official/client.go:320-365, 427-449`, AST 확인). 401 외 응답의 새 재시도는 없다.

### 발견

| 심각도 | 발견 | 근거 | 구현 전 처분 |
|---|---|---|---|
| **P1** | **P2-3은 현 구조로 아직 구현 불능이다.** `FAILED_CONFIRMED`는 terminal이라 `PendingAttempts`에서 빠지고, `checkSymbolFree` B2(804)는 미종결만 막는다. 반대로 공용 `unsettledFor`/`UnsettledOnSymbol`을 넓히면 `clearTheSymbol`도 같은 판정을 써 exit를 막는다. 현재 문서에는 entry-only 차단의 원장 질의·해제 조건이 없다. | evidence `gateway.checkSymbolFree` B1-B7(801-822), `gateway.unsettledFor` B1-B4(32-41), `clearTheSymbol` B3(1639); `releasesReservations` B2(121)는 FAILED_CONFIRMED에서 예약도 푼다. AST 확인. | P2-3 전용의 **entry-only** 판정과 해제 증거를 Pre-Edit에 명시한다. 공용 `unsettledFor`는 바꾸지 않는다. RED는 (1) R3 409 AP 뒤 새 진입 거절, (2) 같은 종목 재취소·exit 청소 진행, (3) flatten `resolveStep`·reconcile 진행, (4) 확인된 취소/주문 종결 뒤 차단 해제를 각각 고정한다. |
| **P1** | **P2-2와 design D3/D5·submit FLM이 모순된다.** 현 문서는 646-725를 `sendOnce`로 추출한다고 하지만 사람 결정은 원 dispatch를 보존한 바깥 래퍼다. 둘을 동시에 구현하면 고위험 dispatch 경계가 달라져 nil byte-identical 주장과 B37 최종 검사 범위가 불명확하다. | `gateway.submit` AST B29-B57(636-761), 특히 B30-B37(648-685); design D3/D5; `internal-execgw--gateway.submit/function-logic-map.md:100`. AST 및 문서 확인. | 코드 전 D3/D5/D10과 해당 FLM/BTM을 래퍼 모델로 고친다. 래퍼는 `first.Err`의 HTTP 409+code 일치만 보고 원 dispatch를 `winCtx`로 재호출한다. `winCtx`는 `dctx`의 자식이고, deadline·취소·NOT_SENT면 첫 409의 AMBIGUOUS를 반환한다. `DispatchVerified` 정착에는 바깥 ctx만 준다. |
| **P2** | 배선 교차·nil 네 시험은 **정확한 정책 값과 행동**을 함께 확인할 때만 F1을 잡는다. 현재 `cmd/tossctl`에는 `buildFlattenWiring` 행동 시험이 없고 기존 시험은 소스 문자열만 읽는다. 포인터/비nil 비교만으로는 `Retries=2, Definitive=true` 혼성 변이를 놓칠 수 있다. | `cmd/tossctl/flatten_test.go:115-144` 손으로 읽음; `classifyCancel` AST B3(458), `planLiquidation` AST B5(307). | flatten 배선은 `Retries=2, MaxWait=5s, Window=8s, Definitive=false`, 엔진은 `Retries=0, Definitive=true`를 정확히 단언한다. 실제 배선+409 AP로 flatten 소진→Held/매도 보류 행동까지 한 시험에서 확인한다. nil 둘은 한 번 호출·detail 바이트 동일까지 확인한다. |
| **P2** | 변이 원장에 P2-2의 취소 전파 축이 명시돼 있지 않다. `winCtx`를 `context.Background`/`WithoutCancel`에서 만들면 호출자 취소 뒤에도 재전송할 수 있고, 정착에 `winCtx`를 쓰면 창 만료 뒤 기록이 실패한다. | `gateway.submit` AST B37(685), `Attempt.DispatchVerified` AST B4-B19(151-215), `clock.Sleep`의 ctx 취소 계약(`internal/clock/clock.go:35-42`, 손으로 읽음). | 변이 두 개를 추가한다: `winCtx` 부모를 `dctx`에서 분리, 정착 ctx를 `winCtx`로 교체. 둘 다 취소 시 재전송 0·첫 409 IN_DOUBT 영속으로 CAUGHT여야 한다. |
| **P3** | 바깥 래퍼가 첫 `plan.call` 내부 시각을 볼 수 없으면 8초 기준을 dispatch 진입 전으로 잡아야 한다. 이는 재시도를 보수적으로 줄일 뿐 8초를 넘기지 않지만, 성공률 저하를 숨기면 안 된다. | `gateway.submit` AST B30-B37(648-685), P2-2 결정. | Pre-Edit에 보수적 시작점과 fake-clock 경계 시험을 기록한다. 첫 전송 뒤에 시각을 잡아 창을 연장하면 FAIL이다. |

### 사람 결정 안전 판단

| 결정 | 판단 |
|---|---|
| H1 | 사람 수락 범위에서 유지 가능. R4는 취소당 최대 8초를 직렬로 더하므로 N×8초라는 운영 위험은 남는다. 소진·중단이 Held/IN_DOUBT로 끝나고 매도에 진행하지 않는 RED가 전제다. |
| P2-1 | 안전성 동작에는 영향 없음. 단 archive 복사는 landed commit의 대상 함수·SHA-256·생성 도구를 각각 대조한 증거일 때만 gate ⑤ 충족으로 취급한다. |
| P2-2 | 조건부 안전. 원 dispatch 재호출은 같은 취소 대상·결정 재검사·B37 최종 ctx 검사를 보존한다. 위 P1 문서/FLM 모순과 ctx 변이를 먼저 닫아야 한다. |
| P2-3 | 의도는 보수적이고 타당하나 현재 설계에는 entry-only 판정과 해제 조건이 없다. 위 P1을 닫기 전에는 안전하다고 판정할 수 없다. |

### 구현 착수 전 필수 조치

1. P2-2 래퍼 모델을 design D3/D5/D10·submit FLM/BTM에 반영하고 ctx/정착 불변식을 RED로 만든다.
2. P2-3 entry-only 차단의 원장 근거·해제 조건을 정하고, exit/flatten/reconcile 비차단 네 갈래를 RED로 고정한다.
3. 배선 교차·nil 네 시험을 실제 정책 값+flatten Held 행동 검증으로 구체화하고, `Definitive` 혼성 및 ctx 부모/정착 변이를 원장에 더한다.

### 1.3 2차 독립 리뷰 처분 (2026-10-28, 작성자 — 사람 결정 「1」 = 설계 보완 라운드)

P2-2(작은 바깥 래퍼)·P2-3(R3 뒤 그 종목 진입 계속 막음)은 사람 결정대로 유지하고 설계로 일관되게 명세했다. 소스·시험 편집 0, 시험 실행 0, 운영 접근 0.

| 발견 | 처분 | 반영 위치 |
|---|---|---|
| P1 P2-3 구현 불능 | **진입 전용 latch 정의.** 기록 = R3 attempt 행 자체(새 표·열 없음, schema 36 그대로) · 키 = (계좌·시장·종목·`target_order_id`) · 세움 = R3 정착(FAILED_CONFIRMED + `outcome=definitive` 꼬리표) · 해제 = r1 대상 주문이 `LiveOrdersForSymbol` 에 없음(terminal fill snapshot) / r2 `ConfirmedCancelOf` found / r3 대사 종결은 r1 과 같은 술어(terminal 증거 생산자 `filldetect/ledger.go:80` 하나) · 내구성 = 진입마다 원장 재계산(메모리 투영 0, `RecoverPending` 은 FAILED_CONFIRMED 를 안 건드림, 행 삭제 경로 0) · 보는 곳 = `checkSymbolFree` B3(810) `!raisesExposure` 반환 **뒤**(B7 뒤·828 앞). 공용 `unsettledFor` 편집 0. `EntryGate.BlockSymbol` 은 진입 전용 축은 같지만 메모리 map 이라 버림(사유 D11) | design D2 부작용·D4·D8·D11, spec delta(R3 bullet + 시나리오 「엔진 취소 확정 미수락 뒤 같은 종목 진입」), 새 번들 `internal-execgw--gateway.checksymbolfree`(24번째) |
| P1 P2-2 ↔ D3/D5·submit FLM 모순 | **`sendOnce` 추출 폐기.** submit 편집은 B49 갈래 729 의 인자 식 하나(`dispatch` → `g.withCancelRefusal(plan, dispatch)`). 래퍼는 첫 결과가 409 AP(code 일치)·AMBIGUOUS 일 때만 같은 클로저를 `winCtx`(dctx 의 자식, `defer cancel()`)로 다시 부른다. B37(685) 은 재호출마다 `plan.call` 직전에 다시 선다. 정착은 `DispatchVerified` 의 `ctx`(dispatch.go:144 및 B5·B7·B9·B11–B13·B15·B17·B19) — `winCtx` 는 닿지 않음. 정책 nil 은 원 `dispatch` 값 그대로(B29–B48·B50–B57 편집 0) | design D3·D5, submit FLM 「a165 편집 계획」·Safety, BTM B37·B47·B49·B50·B52, `measure/gen_flm.py` NOTES |
| P2 배선 4 시험 | W1 flatten 배선 정확 값 `{2, 5s, 8s, false}` + 같은 시험에서 409 AP 소진 → IN_DOUBT → Held·매도 POST 0 / W2 엔진 `{0, true}` + FAILED_CONFIRMED·호출 1·시계 불변·같은 주기 매도 0 / W3·W4 nil 은 호출 1·detail 바이트 동일·오늘 경로(Held / 다음 주기 막힘). 혼성 `{2, true}` 변이 추가 | design D10 |
| P2 ctx 변이 | 원장에 `winCtx` 부모 분리((w2) 호출자 취소 → 재전송 0)·정착 ctx 를 `winCtx` 로((w3) 실시간 창 10ms + 50ms 지연 → IN_DOUBT 영속 + AST 구조 시험)·재호출에 원 dctx·창 시작점 이동((w1)) 추가 | design D10 |
| P3 8s 시작점 | 래퍼가 원 클로저를 처음 부르기 **전**(`g.clk.Now()`·`time.Now()`). 첫 호출 안 7s 진행 fake 로 경계 시험 (w1) | design D3·D5, spec delta |

- 새 AST(HEAD `e4724910b`, `go run ./tools/logic-map`): `analysis/function-logic/internal-execgw--gateway.checksymbolfree/ast.json`(branches 7 — 기존 evidence 와 열거 일치) ·
  `analysis/evidence/execgw--gateway.checkentry.ast.json` · `journal--journal.confirmedcancelof.ast.json` · `execgw--entrygate.blocksymbol.ast.json` · `journal--journal.recoverpending.ast.json`.
  base `34e12967e` → HEAD 사이 a165 관련 Go 파일 변경은 `internal/official/stock_reads.go`(a169)뿐이라 다른 번들의 좌표는 그대로다(`git diff --stat`).
- 손으로 읽은 근거(AST 아님): `raisesExposure` 대입 377·416·447, `Gateway.Cancel` plan 404–417, `classifyMutation` 86 `outcome.Err = err`, `PendingAttempts` 상태 목록 recovery.go:32,
  `IsTerminal` derive.go:94–102, `DELETE FROM mutation_attempts` 0 건(grep), terminal 증거 생산자 1 곳(grep).
- 실행하지 않은 것: 시험 전부·coverprofile(새 번들의 실행 칸은 「미측정」), `check_analysis.py`, `make sdd-sync`·`make gate`.
- **다음**: GPT 보이스 재리뷰. 재리뷰가 꼭 볼 곳 — D11 해제 r2 가 시각 순서를 안 보는 것, r3 를 r1 로 접은 근거, latch 위치(B3 뒤) AST, D5 정착 ctx 불변식, D10 (w3) 의 실시간 시험 안정성.

## 1.3 재리뷰 (GPT-5.6 Terra, 2026-10-28)

### 결론: **FAIL (구현 전 blocker 1)**

P2-2 바깥 래퍼·정착 ctx 분리와 P2-3의 진입 전용 위치는 2차 처분대로 반영됐다. 다만 D11 r2가 **현재 latch보다 앞선** 같은 주문의 CONFIRMED 취소도 해제 증거로 받아, 새 R3 latch를 즉시 풀 수 있다. 살아 있는 주문에 새 진입이 나가는 반례라 2.1 착수 금지다.

- 범위: change 전부, D3·D5·D10·D11, delta, tasks 2.1/2.2/3.1, issues R-a165-5·6, 24 FLM/BTM 목록과 latch/dispatch evidence AST를 읽었다. `Gateway.submit`·`Gateway.checkSymbolFree` AST의 source SHA-256은 현재 `internal/execgw/gateway.go`와 일치한다.
- 실행 0: 소스·시험 편집 0, 시험·gate·운영 콘솔·docker·주문·mutating 명령 0.
- skill 적용: `ai-memory-retrieval` 지침은 읽었으나 현재 세션에 ai-memory MCP 조회 도구가 없어 로컬 memory 색인만 확인했고 a165 직접 기록은 없었다. `search-first`는 새 코드·의존성·통합을 만드는 라운드가 아니라 N/A. `grill-me`/`grilling`은 확정된 사람 결정을 재질문할 권한이 없는 FAIL 재리뷰라 N/A. `impeccable: N/A — backend/documentation-only`.

### 발견

| id | 심각도 | 발견 | 근거 | 구현 전 처분 |
|---|---|---|---|---|
| R3-1 | **P1** | **r2에 시간/인과 경계가 없어 오래된 취소 성공이 새 R3 latch를 푼다.** D11은 같은 `(계좌, 시장, 거래일, 종목, target_order_id)`의 CONFIRMED CANCEL이면 시각을 보지 않고 해제한다고 쓴다. 그러나 `ConfirmedCancelOf`도 그 범위의 모든 CONFIRMED CANCEL 중 가장 이른 행을 돌려줄 뿐이다. working BUY는 CONFIRMED 취소만으로 `LiveOrdersForSymbol`에서 빠지지 않아 다음 엔진 주기에 다시 취소될 수 있다. 따라서 `C1=CONFIRMED`(terminal snapshot 없음) → 다음 주기 `R3=409 AP/FAILED_CONFIRMED` → 새 진입 검사 순서에서 r2가 C1을 찾아 해제 → 살아 있을 수 있는 대상 주문 위에 BUY가 나가는 반례가 성립한다. | `design.md:195`; `specs/order-execution/spec.md:33,70`; `internal/journal/exit_proposal_release.go:155-161`; `internal/journal/fills.go:1849-1883`; `internal/app/engine/exitloop.go:1645-1700`. `ConfirmedCancelOf`의 현 테스트도 시간 경계를 고정하지 않는다(`internal/journal/a094_exit_proposal_release_test.go:275-321`). | D11·delta를 고쳐 r2를 **latch R3 정착 뒤의** 같은 대상 주문 CONFIRMED CANCEL로 한정한다. 기존 `ConfirmedCancelOf`는 exit의 역사 조회이므로 재사용하지 말고, latch attempt의 durable settle cursor 뒤만 찾는 별도 질의를 둔다. 동시/동일 시각은 해제 증거가 아니게 fail-closed 처리한다. RED에 `C1(과거 성공) → R3 → 여전히 진입 거절`, 그 뒤의 `C2(새 성공) → 해제`를 추가하고 D10·tasks 2.2/3.1 변이에 시간 경계 제거를 넣는다. |
| R3-2 | P2 | **R-a165-5·6의 운영 해제/가시성은 아직 충분히 명시되지 않았다.** R3 행은 console `SymbolBlocks`에 안 보이고, 일반 `tossctl order cancel`은 `tradingService.Cancel` 직행이라 journal CONFIRMED CANCEL을 남기지 않는다. 현재 production `Gateway.Cancel` 호출자는 engine과 flatten뿐이다. 따라서 문서의 “운영자 취소”는 어떤 journalled 경로·동일 journal·대상 order id가 필요한지 명확하지 않다. | `issues.md:35-37`; `design.md:195,205-208`; `cmd/tossctl/order.go:285-303`; `internal/flatten/flatten.go:382-386`; `internal/app/engine/exitloop.go:1686-1699`. | P1 수리 뒤에도 2.1 전에 R-a165-6에 운영자에게 보이는 latch 조회/경보와 지원되는 해제 경로를 적는다. 직접 cancel이 해제하지 않는다는 사실, terminal snapshot 또는 **시간 경계 뒤 journalled cancel**만 해제한다는 사실, 불가하면 진입 차단을 유지하고 누가 어떤 증거로 조치하는지를 runbook에 명시한다. |

### 재검증한 처분

- **P2-2/D5: 통과.** B49만 래핑하고 취소는 strategy nil이다(`gateway.go:403-421, 728-733`). `DispatchVerified`는 받은 바깥 `ctx`로 dispatch와 Settle/MarkInDoubt를 수행한다(`journal/dispatch.go:140-160, 199-207`). 래퍼가 `DispatchOutcome` 값만 반환한다는 D5 모델이면 `winCtx`가 정착에 닿지 않는다. nil/비취소면 원 `dispatch` 값을 그대로 반환하므로 첫 호출·detail·reason·state는 오늘과 동일하다는 주장도 맞다. RED w5 골든은 여전히 필수다.
- **D11 위치: 통과.** `checkSymbolFree` B3(`gateway.go:810-811`)가 `raisesExposure=false`를 즉시 반환하고, 계획 삽입점은 B7 뒤·828 앞이다. `submit`은 이 함수를 모든 mutation 전에 부르지만(`gateway.go:525-529`), cancel·청산 SELL·flatten은 latch 질의까지 오지 않는다. `unsettledFor`를 넓히지 않는 판단도 맞다. exit가 latch로 막히는 불변식 4 위반은 이 설계 위치에서는 없다.
- **r3=r1: 통과.** production `Journal.RecordFill` 호출은 `internal/filldetect/ledger.go:80` 한 곳이고, terminal fill snapshot DML은 `RecordFill` 내부다. 별도 reconciliation terminal 생산자는 찾지 못했다. 그러므로 r3를 r1과 같은 `LiveOrdersForSymbol` 술어로 접는 것은 현재 소스와 일치한다.
- **D10 w3: 통과.** 10ms 창보다 50ms가 긴 첫 broker 지연 뒤에야 `winCtx`를 만들므로 deadline은 이미 과거다. CI 부하는 지연을 더 길게 할 뿐, 이 시험은 정확한 소요 시간을 단언하지 않고 `winCtx`가 정착 ctx가 아님만 단언한다. 실시간 50ms 비용은 남지만 timing-flake 근거는 아니다.

### 구현 착수 전 필수 조치

1. R3-1의 durable 시간/인과 경계를 D11·delta·RED·변이 원장에 반영한다. 이 재리뷰의 P1이 닫히기 전 2.1을 시작하지 않는다.
2. R3-2의 운영자 가시성·지원 해제 경로를 R-a165-5·6과 runbook에 정확히 기록한다.

### 미실행

지시상 소스·시험 실행 0. 따라서 RED/coverage/`check_analysis.py`/`make sdd-sync`/`make gate`/strict validate는 이번 재리뷰에서 실행하지 않았다.

### 1.3 재리뷰 처분 (2026-10-28, 작성자)

소스·시험 편집 0, 시험 실행 0, 운영 접근 0. `openspec validate --strict` 통과.

| 발견 | 처분 | 반영 위치 |
|---|---|---|
| R3-1 P1 r2 시간·인과 경계 없음 | **원장 순번 경계.** r2 = latch 행의 `FAILED_CONFIRMED` transition 순번 `s_L` 보다 RECORDED transition 순번 `r_C` 가 큰(엄격) 같은 대상 주문 id·같은 (계좌·시장·거래일·종목)의 CONFIRMED CANCEL. 순번 = `attempt_transitions.id`(`INTEGER PRIMARY KEY AUTOINCREMENT`, schema.go:246; 쓰기 직렬 `_txlock=immediate`, journal.go:216–218·225; UPDATE/DELETE 0 건). 정착이 아니라 기록 시작을 비교해 「latch 뒤 정착」을 포함하고 커밋 지연에 기대지 않는다. 벽시계 0. 새 질의 `ConfirmedCancelAfterLatch(ctx, latchAttemptID)`(journal 새 파일, 범위·대상은 latch 행에서 질의 안에서 읽음). 기존 `ConfirmedCancelOf` 불변 — 생산 호출자 exitloop.go:1650·exit_held_proposal.go:178. RED (f) latch 앞 확인 취소 → 거절 유지 · (g1) 뒤 → 해제 · (g2) 앞에 기록·뒤에 정착 → 거절 유지 · (h) 다른 대상 id·다른 거래일 → 거절 유지. 변이 ⑦ 순서 경계 제거 · ⑧ 대상 id 비교 제거 · ⑨ 순번을 정착 transition 으로 · ⑩ 거래일 비교 제거 | design 머리말·D11(해제 조건 r2·「r2 질의」·RED·변이), delta R3 bullet + 시나리오 「진입 거절보다 앞선 취소 확정은 거절을 풀지 않는다」, tasks 1.3·2.2·3.1 |
| R3-2 P2 운영 해제·가시성 | **design D12 운영 절.** 조회 = 읽기 전용 원장 질의 1·2(기존 콘솔 `SymbolBlocks`·주문 화면·`engine alerts list`·`attempt-resolve` 어느 것도 latch 를 안 보여 줌을 확인). journalled 해제 = r2 엔진 재취소 · r1 체결 감지 종결 기록. `tossctl order cancel`(order.go:295 `tradingService.Cancel` 직행)·브로커 앱 취소는 그 자체로 해제가 아니고 r1 로만 이어짐. `attempt-resolve`·원장 손 편집·`flatten-all` 은 해제 수단 아님. 경보 = 기존 `exit.liquidation_delayed`(`noteDelay`·`noteClearFailure`, 코드 변경 0); latch 만 남은 상태의 경보 없음은 잔여(새 채널·게이트웨이 경보 배선 안 함). 운영 문서 이식은 2.3 | design D12, issues R-a165-5·6, tasks 2.3 |

- 새 AST(HEAD `d64620457`, `go run ./tools/logic-map`): `analysis/evidence/journal--attempt.transition.ast.json`(branches 23 — B17(635)·B19(642)·B20(649)·B23(668), defer `tx.Rollback` 557) ·
  `analysis/evidence/journal--journal.prepare.ast.json`(branches 16 — B14(450) RECORDED transition INSERT, Commit B16(466)).
- 손으로 읽은 근거(AST 아님): lineage.go:157→174·strategy_dispatch_runtime.go:787→795·928→935 의 UPDATE→transition INSERT 쌍, filldetect 추적 목록(ledger.go:112–113), 콘솔·CLI 표면의 latch 미표시, `exit.liquidation_delayed` 두 경로의 문구·임계.
- 실행하지 않은 것: 시험 전부·coverprofile, `check_analysis.py`, `make sdd-sync`·`make gate`. 진입 거절 사유가 운영자 화면까지 가는 경로는 추적하지 않았다(D12 ① 표).
- **다음**: GPT 보이스 재리뷰. 꼭 볼 곳 — r2 를 기록 시작 순번으로 비교한 선택(정착 순번 대비), 질의 1·2 가 코드 판정과 갈릴 가능성(후보 목록으로만 쓰는가), latch 만 남은 상태의 경보 부재 수락.

## 1.3 3차 재리뷰 (GPT-5.6 Terra, 2026-10-28)

### 결론: **FAIL (구현 전 blocker 1)**

R3-1의 `attempt_transitions.id` 순서 비교와 NULL fail-closed는 맞다. 같은 attempt의 재시도·재시작 `RecoverPending`·두 프로세스의 journal 쓰기는 이 판정을 우회하지 않는다. 그러나 latch는 거래일을 넘어서도 진입을 막는데, 다음 날/세션의 재취소는 현재 거래일로 기록되어 r2 후보에서 제외된다. D12의 자동 r2 해제 주장이 성립하지 않아 2.1 착수 금지다.

### 발견

| id | 심각도 | 발견 | 근거 | 구현 전 처분 |
|---|---|---|---|---|
| R3-3 | **P1** | **거래일 경계 뒤 재취소가 r2로 latch를 풀 수 없다.** `checkSymbolFree`의 latch 검색은 현재 plan의 거래일을 받지 않고 시장·종목만 비교하므로 D0의 latch는 D1에도 남는다. 반면 `Gateway.prepareRequest`는 재취소의 intent 거래일을 실행 시각의 `market.TradingDay(g.clk.Now())`로 기록한다. D11 r2는 latch L의 D0 범위를 후보 C에도 강제하고, 심지어 RED (h)는 같은 target X의 다른 거래일 C를 거절으로 고정한다. 따라서 D0 R3 → 거래일 경계 → D1 엔진 재취소 CONFIRMED는 `r_C > s_L`이어도 범위 불일치로 r2가 아니며, 체결 감지 r1이 지연·실패하면 진입 차단이 계속된다. | `design.md:196,198,213-214,232`; `internal/execgw/gateway.go:798-840,1011-1041`; `internal/journal/fills.go:1854-1918`; `internal/filldetect/ledger.go:125-148`; `internal/filldetect/payload.go:105-109,144-159`. | D11의 시간축을 정한다. 권장: CANCEL attempt의 기록일이 아니라 **target order의 canonical scope**(L이 읽은 D0 account·market·trading_day·symbol)를 r2 비교에 보존한다. 즉 재취소 intent가 target의 원 거래일을 상속하거나, 질의가 C의 현재 intent가 아닌 target의 durable owner scope로 같은 대상임을 증명해야 한다. broker order id가 거래일을 넘어 재사용될 수 있으므로 거래일 비교 자체를 제거하면 안 된다. 어느 쪽이든 D0→D1 r2 양성, D0 latch와 다른 범위/다른 target 음성, D0 terminal r1 양성을 RED·변이에 추가한다. |
| R3-4 | P2 | **D12의 읽기 전용 질의 2는 코드 r2 범위를 재현하지 못한다.** Q1은 account를 출력하지 않고, Q2는 `target_order_id`만 거르며 account·market·symbol·latch 거래일을 출력·비교하지 않는다. 재사용된 order id의 다른 범위 CANCEL도 `recorded_seq > latch_seq`처럼 보일 수 있다. 코드 판정은 fail-closed라 잘못 풀리지는 않지만, 직전 R3-2를 해소하려는 운영 조회가 잘못된 후보를 제시한다. | `design.md:253-267`; `design.md:213-214`; `internal/journal/exit_proposal_release.go:150-173`; `internal/journal/schema.go:205-207,223-255`. | Q2를 `latch_attempt`에서 시작해 C와 account·market·trading_day·symbol·byte-exact target id를 모두 조인·출력하고 `recorded_seq > latch_seq`를 SQL에서 직접 판정하게 고친다. P1의 거래일 정책을 먼저 확정한 뒤 같은 조건을 사용한다. |

### 재검증한 처분

- **R3-1 순서 경계: 통과.** `attempt_transitions.id`는 AUTOINCREMENT append-only이며 BEGIN IMMEDIATE 쓰기로 직렬화된다. `Prepare`의 RECORDED와 `transition`의 FAILED_CONFIRMED는 각각 상태 행과 같은 트랜잭션에서 기록된다. 누락 transition은 SQL NULL 비교로 해제되지 않는다. `FAILED_CONFIRMED`는 `PendingAttempts` 대상이 아니므로 `RecoverPending`도 latch를 변경하지 않는다.
- **동시 engine/flatten: 통과.** 같은 journal을 쓰는 후행 CONFIRMED CANCEL만 r2 후보가 되고, latch 이전에 RECORDED된 뒤 늦게 CONFIRMED된 attempt는 `r_C > s_L`을 만족하지 않는다. `flatten-all`은 비상 청산이지 일반 해제 수단이 아니라는 D12 처분도 유지된다.
- **r1·운영 경로: 수락 가능한 잔여.** `JournalTracked`는 terminal이 아닌 주문과 CONFIRMED attempt가 명명한 주문을 추적한다. broker 종료인데 감지가 종결 snapshot을 쓰지 못하면 r1은 오지 않고, `order cancel`·`attempt-resolve`은 journalled 해제가 아니다. D12는 엔진을 계속 돌려 r2를 기다리고 감지 장애 runbook을 따르도록 하며, 진입만 막고 exit·손절·flatten은 열어 둔다. 별도 수동 해제 명령을 추가하지 않는 것은 보수 방향으로 수락한다. 단, 위 P1 때문에 거래일 경계를 넘긴 r2는 현재 그 잔여의 완화책이 아니다.
- **사유 가시성: 수락 가능한 잔여.** 기존 `ReasonSymbolInFlight` 재사용은 API detail의 `already-processing`, attempt id, target order id로 일반 in-flight와 구분된다. 콘솔까지 detail이 전달되는지는 추적하지 않았고 D12 읽기 전용 조회가 보완한다. 새 reason code·UI는 이 라운드 범위 밖이다.

### 구현 착수 전 필수 조치

1. R3-3 거래일 경계 계약·r2 SQL·D11 RED/변이를 수리한다.
2. R3-4 Q2를 확정한 r2 범위와 동일하게 고친다.
3. 수리 뒤 별도 재리뷰를 통과하기 전 2.1을 시작하지 않는다.

### 검증·스킬·미실행

- 정적 확인: HEAD `e10f45500`; 지정된 design/spec/tasks/issues/review 및 journal·engine·filldetect·CLI source 대조, `git diff --check` 통과.
- `ai-memory-retrieval`: MCP 조회 도구가 없어 CLI 색인으로 a165를 찾았으나 결과 0. `search-first`: 저장소 심볼·기존 구현을 우선 검색했고 새 기능·의존성·통합을 만들지 않아 외부 패키지/레지스트리 채널은 N/A. `grill-me`/`grilling`: 기존 사람 결정과 승인 범위가 확정돼 새 질문 N/A. `impeccable: N/A — backend/documentation-only`.
- 지시상 소스·시험 편집·시험 실행·운영 콘솔·docker·sqlite3 운영 원장·주문·mutating 명령 0. `openspec validate`, `check_analysis.py`, `make sdd-sync`, `make gate`도 실행하지 않았다.

### 1.3 3차 재리뷰 처분 (2026-10-28, 작성자)

소스·시험 편집 0, 시험 실행 0, 운영 접근 0. `openspec validate --strict` 통과.

| 발견 | 처분 | 반영 위치 |
|---|---|---|
| R3-3 P1 거래일 경계 뒤 재취소가 r2 로 안 풀림 | **r2 비교 축에서 거래일을 뺀다.** 대상 주문의 정규 범위 = L 의 계좌(`intents.account_ref`, schema.go:206)·시장(204)·종목(207)·`target_order_id`(230)를 질의 안에서 읽어 C 와 비교(정규화는 `ConfirmedCancelOf` 와 같음, exit_proposal_release.go:156–158). 순번 `r_C > s_L` 그대로. 근거: C 의 intent 거래일은 C 를 준비한 날(`Gateway.prepareRequest` 1011 `market.TradingDay(g.clk.Now())` → 1039, evidence AST B3(1012) 만 실패 반환)이고 재취소 대상은 거래일 인자 없는 `LiveOrdersForSymbol`(1630)의 `order.OrderID`(`clearTheSymbol` 1686). 「latch 거래일 이후(`>=`)」 안은 순번이 이미 정한 순서를 벽시계로 다시 묻는 판정이라 버림. **번호 재사용은 거래일 대신 주인 유일성으로 막는다** — L 의 계좌·시장·종목에서 `broker_order_id = target` 인 접수 확정 PLACE/AMEND(schema.go:231)의 `COUNT(DISTINCT intent_id) = 1`; `LiveOrdersForSymbol` 의 주인 술어(fills.go:1871–1877)를 거래일·매매방향 없이 센 것. 0·2 이상이면 해제 아님(r1 은 그대로). 재리뷰가 권한 「target 원 거래일 상속」은 CANCEL intent 에 target 의 거래일 자리가 없어(schema 36) 새 열이 필요하므로 택하지 않음 — 주인 행에서 유도. RED (h) 갱신·(i1) D0 latch → D1 재취소 = 해제 ·(i2) D1 terminal(r1) = 해제 ·(j1~j3) 다른 계좌·시장·종목의 같은 id = 거절 ·(k1) 주인 둘 ·(k2) 주인 0 = 거절 ·(k3) 대조군. 변이 ⑩ 교체(거래일 같음 복원 → i1) ·⑪a/b/c(계좌·시장·종목 비교 제거 → j1/j2/j3) ·⑫(주인 유일성 제거 → k1, `<= 1` → k2) | design 머리말·D11(해제 조건·「범위·대상」 축별 표·「거래일을 뺀 이유」·「번호 재사용」·「r2 질의 모양」 SQL·RED·변이), delta R3 bullet + 시나리오 둘, tasks 1.3·2.2·3.1, issues R-a165-5·7 |
| R3-4 P2 D12 질의 2 가 코드 r2 범위를 재현 못 함 | 질의 1 에 `account_ref` 열. 질의 2 를 latch 하나(`<latch_attempt>`)에서 시작해 r2 와 **같은** 범위 조인(계좌·시장·종목·바이트 같은 대상 id)·주인 수·`recorded_seq > latch_seq` 를 SQL 에서 판정하고 범위 열(계좌·시장·종목·대상 id·latch 거래일·취소 거래일·두 순번·`owner_count`·`r2`)을 출력. 결과는 **후보 목록이지 판정이 아님**을 다시 적음(코드가 정본, 2.3 이식 때 구현 SQL 과 대조) | design D12 ①·②-1, issues R-a165-6 |

- 새 AST(HEAD `5e1ae6260`, `go run ./tools/logic-map`): `analysis/evidence/execgw--gateway.preparerequest.ast.json`(1002–1086, branches 10 — B3(1012) 거래일 실패 반환, 1011 `:=` 대입·`market.TradingDay`/`g.clk.Now` 호출).
  기존 `journal--journal.liveordersforsymbol.ast.json` 은 HEAD 에서 다시 뽑아 `source_sha256`·범위·branches 7 이 같음을 확인(파일 그대로). `clearTheSymbol` 번들 ast 도 HEAD sha 일치(호출 1630·1650·1686).
- 손으로 읽은 근거(AST 아님): `LiveOrdersForSymbol` 의 SQL(`confirmed_orders` 1860–1867·주인 술어 1871–1877), `confirmedOrderTerminalEvidence`(exit_proposal_release.go:281–294 — terminal 증거는 주인의 거래일로 범위를 잡아 r1 은 거래일을 넘겨도 맞음), 재사용 시험 이름 fills_test.go:635·985.
- 임시 검증(증거 아님): design 의 r2 SQL·D12 질의 2 를 메모리 SQLite 최소 표에 돌려 (f) 앞선 C1 → 없음, (j1~j3) → 없음, (i1) D1 C → 찾음, (k1) 주인 둘 → 없음을 봤다(구문·조인 확인용, 시험 아님 — RED 는 2.2).
- 새 잔여 R-a165-7: 원장 밖 주문의 번호 재사용은 주인 유일성이 못 본다(측정 0, 수락 사유 issues).
- 실행하지 않은 것: 시험 전부·coverprofile, `check_analysis.py`, `make sdd-sync`·`make gate`.
- **다음**: GPT 보이스 재리뷰. 꼭 볼 곳 — 거래일을 뺀 뒤의 번호 재사용 방어(주인 유일성이 거래일 비교를 대신하는가, 원장 밖 재사용 R-a165-7 수락), 주인 술어를 거래일·매매방향 없이 센 선택, 「target 거래일 상속」을 택하지 않은 이유(schema 36).
