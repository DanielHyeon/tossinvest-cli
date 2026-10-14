# a138 design — 최대 보유 포지션 수

근거 표기: 「AST:」는 `analysis/ast-evidence/<번들>.ast.json`(`tools/logic-map` 산출물, HEAD a1d66f13, 편집 전)이다. 그 밖의 줄 번호는 같은
HEAD 를 읽은 것이다. CodeGraph 근거는 `analysis/code-context/`.

## 사용자 결정 (2026-10-15 — Manager 전달 원문)

> a138-max-open-positions (user decision 2026-10-15: max 6)
> - New entries are refused when the account would hold more than 6 open positions (count rule to be defined from existing code: engine-managed open
>   positions + pending entry reservations so concurrent entries can't overshoot; decide whether adopted/manual holdings count by reading what the
>   journal/positions model already distinguishes — if it's a genuine human choice, STOP after the proposal and report options with a recommendation).
>   Sells, stop-loss, emergency exits, flatten, reconcile are never blocked or delayed (safety invariant 4).
> - Configurable like other automation_gate limits (config field, default 6, audited settings seam, startup interlock cross-check — follow how a136 did
>   max_daily_buy_notional …). Fail closed for entries if the count can't be computed.

사람 결정(2026-10-15, Manager 전달 — 사용자 답 「권장대로」):

> User decided a138 (2026-10-15, "권장대로"): D1 = B (count the whole account: managed + reconciled external holdings + HELD entry reservations + this
> symbol, distinct (market, symbol)); D2 = scale-in ALLOWED at the cap (an entry for an already-held symbol does not raise the count).

반영: D1 → D1-B(spec 델타 「최대 보유 포지션 수 한도」 ① 그대로). D2 → 허용(같은 요구의 「이미 집합에 있는 종목에 더 사는 진입은 … 거부되지 않는다」 문장 ·
시나리오 「든 종목에 더 사는 진입」 그대로).

## D0. 저장소가 이미 가진 것

| 사실 | 근거 |
|---|---|
| 계좌 총계 한도의 최종 권위는 예약 트랜잭션(BEGIN IMMEDIATE, 단일 커넥션) | risk-management 「발급 절차와 예약의 권위」, reservations.go 머리 주석 |
| 모든 예약 tx 는 `reserveRows` 를 지남, 하루 매수 판정이 그 안 5b | CodeGraph callers(raw/callers-reserveRows.txt) · AST `journal--reserverows` B4(339) |
| 하루 매수 판정은 노출을 올리지 않는 결정을 통과시키고, 상한 없는 EXPOSURE_RAISING 예약을 거부 | AST `journal--checkdailybuynotional` B2(73) · B3(77) |
| 위험 축소 발급은 예약 tx 를 지나지 않음 | AST `execgw--riskguardian.issuereduction` 호출 목록(`g.journal.RecordDecision` 만) |
| 포지션 투영의 세 종류: 엔진 진입(`entry_decision_id`) · 편입(`adoption_id`) · 외부(둘 다 NULL) | core_domain.go:127-151 · adoption.go:77-95 |
| 외부 보유는 대사가 조정 이벤트로 접어 넣음(기본 주기 60초 `DefaultReconcilePeriod`, 안정 스냅숏이 잡힌 사이클에서만 — reconcileloop.go:95 · :414 · :431), 발견 시 알림 | reconcile/external.go 머리 주석 · `ApplyPositionAdjustment` |
| 관리 여부 단일 술어 `position.ExitEligible`(= 진입 결정 또는 편입) | position_projection.go:517 · a137 `a137_daily_loss_rows.go` 주석(정적 시험이 지킴) |
| 투영은 FLAT 행을 쓰지 않음 — 살아 있는 행 = CLOSED 가 아닌 행(OPENING · OPEN · SCALING · CLOSING) | position/position.go:21-31 |
| 결정 preimage 에 `market`(소문자) · `symbol`(대문자) 이 정규화돼 저장됨 | decision.go:63-73 · `canonicalVenue`(556) |
| 종결 체결은 같은 tx 에서 HELD 예약을 해제한 뒤(RecordFill B28 512) 투영 훅을 부르고(B37 551), 투영이 전이를 거절하면 포지션 행 없이 종목 범위 RECONCILE 만 남김(ProjectPosition B5 101 → `enterReconcileInTx`) — 그 종목은 ① · ② 어디에도 없음 | AST `journal--journal.recordfill` · `journal--projectposition`, reconcile_states DDL(execution_contract.go:202) + `scope_market`(risk_bucket_owner.go:57, 투영 거절은 NULL 로 씀) |
| Gateway 는 EXPOSURE_RAISING 결정에 종류와 무관하게 HELD 예약이 하나라도 있으면 제출을 인가함 | gateway.go `checkReservation`(freeze 리뷰 P1-2) |
| 미소비 만료 · 일 경계 보유는 예약 tx 앞단(`reservePrecheck` 3)이 같은 tx 에서 해제 | reservations.go:300-317 |
| 정책 유도는 직선 함수 하나 | AST `engine--riskpolicyfromautomationgate`(분기 0, 반환 2 — 39 는 `number` 클로저) |
| 인터록 단일 출처 검사는 `verifyGate` 7(B7) · 7b(B8 — 하루 매수) | AST `engine--verifygate` B7(530) · B8(534) |

## D1. 무엇을 「보유 포지션」으로 세나 — **사람 결정: B(2026-10-15)**

저장소는 세 종류를 칼럼으로 구별하고, 계좌 단위 한도의 두 최근 선례가 반대로 답한다.

- a136 하루 매수 총액: 엔진 밖 수동 매수도 **센다**(사용자 결정 D9=b — 저자 추천과 반대로 고름).
- a137 일일 손실: **관리 포지션만** 잰다(외부 포지션의 손실은 엔진이 값을 잴 수 없어 측정 대상 밖 — 명시 한계).

그래서 저자가 고르지 않는다.

| 안 | 세는 포지션 | 장점 | 단점 · 거절되는 정상 입력 |
|---|---|---|---|
| D1-A | 관리 포지션(`ExitEligible` — 엔진 진입 + 편입) + 진행 중 엔진 진입 | 엔진이 손절로 보호하는 범위와 같음(a137 과 같은 경계). 운영자가 손으로 오래 든 종목이 자동 진입을 막지 않음 | 계좌는 6 개보다 많이 들 수 있음(외부 보유 + 엔진 6). 「계좌가 들게 될」 문구와 다름 |
| **D1-B (추천)** | 계좌 전체의 살아 있는 포지션(관리 + 외부 — 대사가 접어 넣은 것) + 진행 중 엔진 진입 | 문구(「the account would hold」) 그대로. a136 D9=b(수동 활동도 계좌 한도에 셈)와 같은 방향. 더 보수적(D1-A 가 허용하는 진입의 부분집합만 허용) | 손으로 든 종목이 칸을 차지 — **외부 보유가 6 개 이상이면 자동 진입이 전부 거부됨**(LIMIT_REACHED 상세에 개수가 남음, 운영자가 상한을 올리거나 편입/정리). 외부 매수는 대사(60초 이상, 안정 스냅숏 사이클) 뒤에야 보임 — 그 사이 한 칸 과소 집계 가능 |
| D1-C | D1-B + 미체결 수동 매수 주문의 종목(a136 수동 매수 스냅숏을 종목 단위로 확장) | 대사 전 수동 매수까지 셈 | 수동 매수 보드를 종목 단위로 바꿔야 함(a136 refresher · 보드 · journal 칸 편집 — 표면 큼). 수동 주문은 어차피 막지 않으므로 얻는 것이 작음 |

추천 이유: 사용자 문구와 a136 의 선택(계좌 한도는 엔진 밖 활동도 셈)에 맞고, 두 안 중 거부 쪽(보수)이다. D1-A 를 고르면 바뀌는 것은 포지션 질의의
술어 하나(`Position.ExitEligible` 로 거름)뿐이고 나머지 설계는 같다. **spec 델타는 D1-B 로 써 두었고, 다른 답이면 「최대 보유 포지션 수 한도」 요구의 ① 문장만 바꾼다.**

## D2. 경계 — 「6 개보다 많이」

- 세는 단위는 서로 다른 (시장, 종목) 하나다. 같은 종목의 여러 인스턴스 · 부분 체결 중인 진입과 그 포지션은 하나로 센다.
- 판정: 집합 S = ① 살아 있는 포지션(D1) ∪ ② HELD 예약을 **종류와 무관하게**(닫힌 집합 OPEN_EXPOSURE · DAILY_LOSS · CASH 를 `IN` 으로 — 색인
  `idx_reservations_held` 를 그대로 탐) 가진 EXPOSURE_RAISING 결정의 종목(Gateway 인가 술어와 같음, freeze 리뷰 P1-2 (a)) ∪ ③ 이번 결정의 종목
  ∪ ④ 종목 범위 활성 RECONCILE(`reconcile_states.symbol` 이 있고 `released_at` 이 NULL) 중 ①~③ 에 없는 것 — 한 칸씩
  (freeze 리뷰 P1-1: 투영이 거절된 종결 체결은 예약을 해제하고 포지션 행을 남기지 않음). `scope_market`(KR/US, risk_bucket_owner.go:57 이 더한 칸)이 있으면
  (시장, 종목)으로 맞추고, NULL(전역 종목 범위 — 투영 거절이 쓰는 모양, position_projection.go `enterReconcileScopeInTx`)이면 종목만으로 ①~③ 의 어느 원소와 맞춤 —
  맞는 것이 없으면 한 칸(과대 집계 쪽). 리뷰는 「시장 칸 없음」이라 했으나 v5 뒤 칸이 더해졌음(HEAD 정정).
  **|S| > 상한이면 거부**.
- 귀결: 이미 6 개를 든 계좌에서 그중 한 종목에 더 사는 진입(scale-in · 캠페인 추가 다리)은 |S| = 6 이라 허용. 문구 「6 개보다 많이 들게 될 진입」 그대로다
  (사람 결정 D2 = 허용, 2026-10-15).
- 돈 한도의 「도달 시 거부(≥)」 규칙과 다른 이유: 그 규칙은 금액 총계의 경계이고, 이 상한은 **허용되는 최대 개수**다(max 6 = 6 은 허용). spec 은 이 차이를 요구 안에 적는다.

## D3. 판정 위치 — 예약 트랜잭션 안

a136 D3 와 같은 이유: 체인은 호출자가 모은 값으로 한 번 돌기 때문에, 동시 진입 둘이 같은 「5 개 보유」를 보고 함께 통과할 수 있다. 예약 tx 는 단일 커넥션
BEGIN IMMEDIATE 라 직렬이므로 그 안에서 원장을 다시 읽으면 두 번째 진입은 첫 번째의 HELD 예약(②)을 본다. 자리는 `reserveRows` 5c(하루 매수 5b 다음,
insert 앞) 한 줄, 판정 함수는 새 파일 `internal/journal/open_positions.go` 의 `checkOpenPositions`.

- 노출을 올리지 않는 결정(RISK_REDUCING 등)은 즉시 통과(a136 B2 와 같은 모양). 그런 결정은 애초에 예약 tx 에 오지 않지만(D0), 이 함수가 그 사실에 기대지 않게 함.
- EXPOSURE_RAISING 인데 요청에 상한이 없으면 `ErrInvalidRequest`(한도를 빠뜨린 발급 문이 조용히 통과하지 못함 — a136 B3 와 같은 규칙).
- 이번 결정의 종목은 tx 안에서 `decisions.risk_preimage` 를 다시 읽어 얻음(호출자 값 불신). preimage 종류가 RISK_INTENT 가 아니거나 해석 불가면 오류(거부).
- ② 의 종목도 각 결정의 preimage 에서 얻음. 해석 불가 행이 하나라도 있으면 판정 실패(거부) — 0 으로 두지 않음.
- 결과: |S| > 상한 → `ErrReservationLimitExceeded`(LIMIT_REACHED, 재수집 안 함). 상세 문구 `MAX_OPEN_POSITIONS` · 개수 · 상한 · 이번 종목.
  LIMIT_REACHED 는 총 노출 · 하루 매수 · 개수 셋이 공유하고 원인은 상세 문구로만 갈림(a136 선례와 같은 모양 — sentinel 문서에 개수 상한은 「초과(>)」임을 적음).

## D4. 설정 · 정책

- `config.AutomationGate.MaxOpenPositions float64 \`json:"max_open_positions,omitempty"\``. 숫자를 float 로 받는 이유: 형제 칸과 같고, `int` 로 받으면
  `6.5` 같은 값이 **게이트 OFF 에서도** config 읽기 전체를 실패시켜 OFF = upstream(§0.2)을 깸. 정수성은 게이트 ON 정책 검증이 봄.
  한정(freeze 리뷰 P3-5): 이것은 **소수** 값에 대해서만 참 — `"6"`(문자열) · `1e400` 은 형제 칸과 똑같이 OFF 에서도 config 해석을 실패시킴(기존 선례).
- 해석 `EffectiveMaxOpenPositions()`(새 파일 `internal/config/open_positions.go`): 0 이 아니면 그대로(음수 · 비정수 · NaN · Inf 도 그대로 넘겨 검증이 거부), 0 이면 기본값 6.
  통화와 무관. Guardian 정책 유도와 인터록 대조가 모두 이 함수를 씀(단일 출처).
- `risk.Policy.MaxOpenPositions int`, `DefaultPolicy` = 6(출처: 사용자 결정 2026-10-15, a138 — 이전에 상한이 없던 새 검사라 완화 아님),
  `Validate` 는 `< 1` 거부. 정수 판정은 **한 헬퍼** `config.WholeOpenPositions(v) (int, bool)`(NaN · ±Inf · 소수 · < 1 · int32 범위 밖 → false — `math.Trunc(+Inf) == +Inf`
  이므로 IsInf 를 먼저 봄, freeze 리뷰 P2-3)이고 정책 유도 · 인터록 · 콘솔 검증이 모두 그것을 씀. 유도(`riskPolicyFromAutomationGate`)는 false 면 0 을 넣어
  `Validate` 가 거부하게 하고, 생산 factory(`newProductionRiskGuardian`)는 Validate 앞에서 같은 헬퍼로 먼저 거부해 **운영자가 적은 값**을 문구에 담음
  (`ErrLimitsRequired`). 생산 기동은 factory 가 인터록보다 먼저라(engine.go:536-558) 7c 의 정수 거부는 주입 Guardian seam 에서만 닿음 — 그래도 두는 이유는 단일 출처 대조.
- 판정은 체인이 읽지 않음(체인 단계 아님) — 발급자가 예약 요청에 그대로 실음.

## D5. 단일 출처 · 감사 · 콘솔

- Guardian 진술 `(*RiskGuardian).MaxOpenPositionsLimit() int`, 인터록 `checkGuardianOpenPositionLimit`(verifyGate 7c): 해석 값이 유효 정수가 아니면
  `ErrLimitsRequired`, Guardian 이 진술하지 않거나 다르면 `ErrGuardianLimitsMismatch`.
- `AutomationStatus.MaxOpenPositions`(해석 값 운반), `acceptanceDetail` 에 `max_open_positions=`, `recordGateSettings` 끝에 audit 한 줄
  (`engine.automation_gate.max_open_positions` = 해석 값 — config 수동 편집의 이전 · 새 값이 여기 남음; 순서는 끝에 덧붙여 기존 순서 무변경).
- `LimitsDigest` 입력에 상한을 더함 — 이 값만 바꿔도 전략 바인딩이 무효화됨(a136 P2-4 와 같은 규칙).
- 콘솔: `config.GuardianLimits.OpenPositions *float64`(nil = 키 무접촉, 티어 프리셋은 갖지 않음). 고급 폼 정수 칸 · 현재 값 표 행(미설정이면 「기본 6 적용」) ·
  저장 뒤 audit(파일 값). 통화 전환 저장은 이 키를 지우지 않음(개수는 통화로 재해석되지 않음 — a136 P1-2 의 지움 규칙은 금액 전용).
  확인 문구 · 추가 승인 없음(기억 규칙). 규칙 12(ui-skills 1~3) · 13(렌더 확인, 버튼 미클릭).

## D6. 매도는 여기 오지 않음

`IssueReduction` 은 예약 tx 를 지나지 않고(D0), 예약 tx 안의 판정도 노출을 올리지 않는 결정을 먼저 통과시킨다. 손절 · 비상 청산 · flatten 은 RISK_REDUCING
발급 또는 Issuer 직행이고, 대사 · 체결 감지는 예약 tx 를 열지 않는다. 판정은 거부 방향만 추가한다.

## D7. 거절되는 정상 입력 (fail-closed 가 무엇을 거부하는가)

- 브로커에서 이미 취소됐지만 체결 감지가 종결을 아직 관측하지 않은 진입 주문 — 그 HELD 예약이 한 칸을 차지(관측 뒤 해소).
- UNRESOLVED_IN_DOUBT · UNKNOWN_BROKER_STATE · NONCE_SPENT_EXPIRED 로 운영자 해제를 기다리는 HELD 예약 — 기한 없이 한 칸(`OperatorReleaseReservation` 뒤 해소).
- CLOSING(청산 주문이 나갔지만 아직 수량이 남은) 포지션, 엔진 밖에서 팔았거나 매도 체결이 아직 관측되지 않은 포지션(대사 수렴 전), 소수 잔량(dust) 포지션 — 한 칸씩.
- D1-B: 손으로 든 종목 — 한 칸씩. 외부 보유가 상한 이상이면 신규 종목 진입 전부 거부.
- 종목 범위 RECONCILE(투영 거절 · 수량 불일치 등 원인 무관) — S 에 없으면 한 칸(④, 시장 없는 범위는 종목만으로 맞춰 과대 집계 쪽).
- KR 과 US 는 한 계좌에서 합산 — 두 시장 합계 6.
- 옛 결정 하나의 preimage 가 해석 불가 — 그 결정이 HELD 예약을 들고 있는 동안 모든 진입 거부(원장 결함이 조용히 0 이 되지 않게).
- 원장 읽기 오류 — 그 진입 거부.
- 거부하지 **않는** 것(permissive 쪽): 대사가 접어 넣기 전의 외부 매수(60초 이상, 안정 스냅숏 사이클까지 — 상한 없음), 미체결 수동 매수 주문(D1-C 를 고르지 않음),
  운영자가 HELD 예약을 해제했지만 주문이 살아 있거나 체결된 경우(체결 감지 전), 엔진 진입 주문을 엔진 밖에서 정정한 경우(자식 주문에 로컬 attempt 가 없어 부모 종결 시
  예약이 해제되고 대사 전까지 빈 창). 투영 거절 갭(P1-1)은 ④ 로 닫음.

## D8. 원장 · 스키마 · 성능

스키마 변경 0. 질의 셋: `positions` 는 색인 `idx_positions_symbol`(account_ref, market, symbol, state) 의 선두 account_ref 범위라 **그 계좌의 CLOSED 이력 인스턴스까지**
훑는다(freeze 리뷰 P3-1 정정 — 행 수는 계좌가 거쳐 간 인스턴스 수, 작음). HELD 예약은 `idx_reservations_held`(account_ref, kind IN 셋, state) 범위, 활성 RECONCILE 은
`idx_reconcile_active`(부분 색인) 범위. 결정 preimage 는 HELD 예약의 결정 수만큼(상한 근처라 작음) 읽음. 진입 tx 가 잠금을 잡는 동안 매도 기록이 기다리므로(§0.3 간접)
지연 실측을 검증 영수증에 남김.

## D9. 공유 파일 순서 (a139)

`riskPolicyFromAutomationGate` · `risk.DefaultPolicy` · risk-management 「Automation gate limits …」 요구는 a139(건당 위험 예산 = 일일 손실 / 3)도 고친다.
Manager 지시대로 a138 의 Go · spec 을 먼저 전부 커밋하고, a139 델타는 a138 이 남긴 문장 위에 쓴다. a138 이 먼저 archive 되어야 a139 의 MODIFIED 가 충돌 없이 합쳐진다.
