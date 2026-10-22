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
