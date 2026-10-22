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
