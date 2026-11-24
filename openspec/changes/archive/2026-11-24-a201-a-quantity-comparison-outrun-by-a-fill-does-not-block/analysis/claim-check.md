# a201 claim check — proposal · design 의 주장 대조(tasks 1.11)

HEAD `86a131a51`. 근거: `cited-ast/`(36) · `repro/`(실제 RunOnce 6) · `probe-d5/`(scratch 사본 탐침) · `evidence-phase0.md` · `residual-window.md` · `spec-contradictions.md`. 판정 = 참 / 거짓 / 부분 거짓 / 미확인.

## 결정에 닿는 것(멈춤 조건 — Manager 로)

| id | 주장(원문 요지) | 판정 | 근거 |
|---|---|---|---|
| **C-D6** | design D6 · proposal 비목표 「해제 규칙 · release cause 무편집, 이 change 는 『세우지 않는다』만 한다」 | **거짓(효과)** | 뺀 결과 비교가 깨끗해지면 `Observe` 가 해제 갈래로 감(B2 if@650 거짓 → B6 range@666: 활성 QM 블록마다 `creditUsableBy` → 해제, 또는 a192 `orderTerminalCandidate` call@682). 뺀 종목은 `disputed`(call@644)에서 빠져 그 종목의 앞 크레딧이 반박되지 않음(B29 if@841). **탐침 실측**: 앞 주기 31→20 조정 크레딧이 있던 종목에 W2 → D5 켜면 그 주기에 `ADJUSTMENT_APPLIED` 해제(`probe-d5/run-d5-on.txt`). 규칙 문언은 안 바뀌지만 이 change 가 해제를 일으킴 |
| **C-D4** | D4 「그 주기의 다른 종목 판정에 영향 없음」 | **거짓** | 같은 기전: 뺀 종목이 그 비교의 유일한 차단 항목이면 비교 전체가 깨끗해져 **다른 종목**의 크레딧 있는 QM 블록 해제 · a192 해제 판정 · `clearStreaks`(call@655) · 영구 승격 재시도 권위 확인(call@659)이 그 주기에 일어남. 지금 코드는 같은 일을 「다음 깨끗한 주기」에 함 — 규칙상 정당한 해제가 한 주기 당겨지는 쪽이지만 「영향 없음」은 아님 |
| **C-R1** | design §4 · proposal 비목표 「잔여는 다시 읽은 직후부터 Observe 기록까지의 메모리 구간」, §2 정상 입력 4 「stale 로 판이 멈춘 경우 뒤 종목들도 각자 재판정된다」 | **부분 거짓(창을 잘못 짚음)** | W2 · W3 의 체결은 정의상 조정 tx 전 = 재판정 전 → 그 구간에 떨어지는 W2 · W3 는 0. 실제로 남는 것은 **그 주기에 조정이 시도되지 않은 종목**(판 중단 G1 뒤 · 복수 시장 거절 · (b) 단독 stale)의 체결이 재판정 **뒤**(다음 로컬 읽기 전 ≈ 60초+)에 오는 경우 — 재현 `TestA201CurrentBehaviourAStaleStopsTheTrailingSymbolWhoseFillLandsAfterObserve`, 탐침 D5 켜고 끄고 모두 B 고착. 여러 손절이 한꺼번에 체결되는 날(이 change 가 겨냥한 날) 앞 종목 W3 이 판을 멈추면 뒤 종목들이 계좌-앞섬 상태로 이 모양이 됨. 구간 안 I/O 도 0 이 아님(EnterReconcile tx · 조건부 SELECT — residual-window.md §1) |
| **C-D5** | D5 판단 기준 「a192 로 2주기 안에 풀리면 D5 를 빼고」 vs §3-1 「그 주기 뒤 그 종목 차단 0, `reconcile_states` 행 0, 진입 · 편입 허용」 | **design 내부 모순(측정으로 드러남)** | 탐침 D5 끔: 네 모양 모두 그 주기 missing-order 블록(행 1 · 진입 거절) → 일치 2주기째 `ORDER_TERMINAL_CONFIRMED` 해제 → D5 기준으로는 빼야 하지만 §3-1 의 「행 0 · 진입 허용」은 깨짐(약 2분 진입 · 편입 거절). 또 **앞 크레딧 위의 W2 는 D5 없이 고착 그대로**(QM 블록이 키를 쥐어 결속 없음) |

## 사실 정정(결정에 안 닿음 — 문서 표현)

| id | 주장 | 판정 | 근거 · 정정 |
|---|---|---|---|
| C-D3a | D3 (가) 「사람의 앱 거래는 … 워터마크를 움직이지 않는다 — 외부 거래로 생긴 불일치는 **여기서** 걸러진다」 | 부분 거짓 | 종목 워터마크는 계좌 조건 없음(position_adjustments.go:190-198 [손 SQL]) — 다른 계좌의 같은 종목 체결 · 같은 창의 엔진 체결이 움직임(W3-b 재현이 다른 계좌 체결로 (b) stale 을 냄). 외부 거래가 섞인 불일치를 거르는 것은 (가)가 아니라 **(다)**. 안전성은 (다)로 유지. tossctl 수동 주문이 attempt 를 써서 감지기 추적 대상이 되는지 — 미확인 |
| C-D3c | D3 (다) 「`Compare` 가 쓰는 같은 일치 판정」 | 참(조건부) | 로컬 정의는 같음(`HeldBySymbol` — LocalStateFromJournal call@197 과 같은 비-CLOSED 종목 합). 판정 = `canonicalPresentQuantity`(부재 → "0") + `quantitiesAgree` — 둘 다 **비공개** → 재판정 함수는 `internal/reconcile` 안이어야 함. `Compare` 의 분류 갈래 중 「로컬 부재 · 계좌 양수 → ExternalPos(비차단)」 는 이 판정으로 「불일치 유지」가 됨 — 보수 방향, 기록 |
| C-W3b | §2 정상 입력 1 「W3 의 stale 두 종류(수량 재검증 · 워터마크 재검증)」 | 부분 거짓 | 엔진 체결 자신은 tx 직전에 와도 (a) 수량으로 걸림(재검증 순서 B15@253 → B17@267, 재현). (b) 단독 stale 은 **수량 불변 fill_events 행**(다른 계좌 체결 · 투영 거절 체결)이 tx 직전에 들어갈 때만 생기고, 그 뒤 고착은 엔진 체결이 따로 와야 생김(합성 — W3-b 재현) |
| C-P6 | proposal 「정본은 이미 이것을 금한다 :79 · :91」 | 부분 참 | W3(stale)는 :79 대상, **W2(refused)는 :79 의 「어긋나면」(재검증 실패)에 해당 안 함** — 조정 시도 전 |
| C-P10 | proposal 「종목을 닫을 때마다 『외부 청산』 알림을 동기 HTTP 로 → k 번째 종목의 창이 넓어진다」 | 참(조건부) | 동기 사슬 AST: ConvergeQuantities call@254 → notifierAlerter call@141 → Notify call@154 → publishBestEffort call@184(고루틴 0). 조건: 앞 종목이 **엔진 관리 포지션을 수렴으로 닫았을 때**(`ClosedExitState`)만, Publisher 가 있을 때만. 그리고 앞 종목이 W3 stale 이면 판이 멈춰 k 번째는 W2/W3 가 아니라 C-R1 모양이 됨 |
| C-S4 | design §4 「브로커 호출 · I/O 없는 메모리 구간」 | 부분 거짓 | 브로커 · 네트워크 0 은 참, 장부 I/O 1~3(residual-window.md §1) |

## 참으로 확인된 것

| id | 주장 | 근거 |
|---|---|---|
| C-P1 | 체결 반영과 대사가 잠금을 공유하지 않음 | RunOnce AST 에 잠금 호출 0, 장부는 문장 · tx 단위로만 직렬(단일 연결). 재현이 주기 안 임의 지점 커밋을 만듦 |
| C-P2~P5 | W2 건너뜀 · W3 stale · 같은 diff 로 Observe · 5주기 미해제 · 진입 · 편입 거절 | 실제 RunOnce 재현(`repro/`) — 재매수 adopted 0 |
| C-P7 | add-core-domain issues.md:652 「항상 무언가를 쓴다」 | archive/2026-07-26-add-core-domain/issues.md:652-654 원문 |
| C-D1 | 창은 로컬 읽기~조정 커밋, `Observe` 의 다른 생산 호출자 없음 | AST · grep 1 자리(:464) |
| C-D2 | 워터마크를 로컬 읽기 전에 읽으면 보수 | id 순서 = 커밋 순서(단일 연결 journal.go:174, BEGIN IMMEDIATE) → 전역 최대보다 큰 종목 워터마크 = 그 뒤 커밋. 읽기~로컬 읽기 사이 체결은 (가)만 참으로 만들고 (다)가 거름. 뒤에 읽으면 놓침 — 참 |
| C-D3b | (나)가 W4 · W5 · W6 을 건드리지 않음 | 세 창 모두 그 주기 조정이 커밋됨 → `Converged` 에 그 종목(Applied 참/거짓 무관) → (나) 거짓 → 무변경. 전제 (a): `ConvergeReport.Converged` 로 판정 가능, 판 중단에도 커밋분 보존(defer@172) |
| C-D4b | 뺀 항목은 승격 횟수에 안 들어감 | `promotionKeys`(call@645)는 받은 diff 에서만 키를 뽑고 `advanceStreaks`(B1 range@1092)는 이번 키만 남김, 깨끗하면 `clearStreaks` |
| C-D8 | 읽기 실패면 안 뺌 | design 결정(fail-closed) — 구현 전 |
| C-E | a200 과 편집 파일 겹침 없음, a199 는 같은 파일 다른 함수 | edit-surface.md §3 |

## 미확인

- 운영 집계(ops-aggregates.md)의 수치 — Manager 제공, 운영 저널 접근 금지라 재현 안 함.
- 「첫 종목이면 창은 장부 읽기 · 쓰기 몇 번의 폭」 · 구간 폭 ms — 측정 안 함.
- 운영 장부가 단일 계좌인지(종목 워터마크의 계좌 무관성이 실제로 다른 계좌 체결에 노출되는지).
- tossctl 수동 주문이 attempt 를 써 감지기 추적 대상이 되는지.
- ntfy Publisher 의 실제 시한과 운영 설정(켜짐 여부).
