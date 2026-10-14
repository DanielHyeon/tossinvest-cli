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
| 외부 보유는 대사가 조정 이벤트로 접어 넣음(30초 루프), 발견 시 알림 | reconcile/external.go 머리 주석 · `ApplyPositionAdjustment` |
| 관리 여부 단일 술어 `position.ExitEligible`(= 진입 결정 또는 편입) | position_projection.go:517 · a137 `a137_daily_loss_rows.go` 주석(정적 시험이 지킴) |
| 투영은 FLAT 행을 쓰지 않음 — 살아 있는 행 = CLOSED 가 아닌 행(OPENING · OPEN · SCALING · CLOSING) | position/position.go:21-31 |
| 결정 preimage 에 `market`(소문자) · `symbol`(대문자) 이 정규화돼 저장됨 | decision.go:63-73 · `canonicalVenue`(556) |
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
| **D1-B (추천)** | 계좌 전체의 살아 있는 포지션(관리 + 외부 — 대사가 접어 넣은 것) + 진행 중 엔진 진입 | 문구(「the account would hold」) 그대로. a136 D9=b(수동 활동도 계좌 한도에 셈)와 같은 방향. 더 보수적(D1-A 가 허용하는 진입의 부분집합만 허용) | 손으로 든 종목이 칸을 차지 — **외부 보유가 6 개 이상이면 자동 진입이 전부 거부됨**(LIMIT_REACHED 상세에 개수가 남음, 운영자가 상한을 올리거나 편입/정리). 외부 매수는 대사(≤30초) 뒤에야 보임 — 그 사이 한 칸 과소 집계 가능 |
| D1-C | D1-B + 미체결 수동 매수 주문의 종목(a136 수동 매수 스냅숏을 종목 단위로 확장) | 대사 전 수동 매수까지 셈 | 수동 매수 보드를 종목 단위로 바꿔야 함(a136 refresher · 보드 · journal 칸 편집 — 표면 큼). 수동 주문은 어차피 막지 않으므로 얻는 것이 작음 |

추천 이유: 사용자 문구와 a136 의 선택(계좌 한도는 엔진 밖 활동도 셈)에 맞고, 두 안 중 거부 쪽(보수)이다. D1-A 를 고르면 바뀌는 것은 포지션 질의의
술어 하나(`Position.ExitEligible` 로 거름)뿐이고 나머지 설계는 같다. **spec 델타는 D1-B 로 써 두었고, 다른 답이면 「최대 보유 포지션 수」 요구의 ① 문장만 바꾼다.**

## D2. 경계 — 「6 개보다 많이」

- 세는 단위는 서로 다른 (시장, 종목) 하나다. 같은 종목의 여러 인스턴스 · 부분 체결 중인 진입과 그 포지션은 하나로 센다.
- 판정: 집합 S = ① 살아 있는 포지션(D1) ∪ ② HELD OPEN_EXPOSURE 예약을 가진 EXPOSURE_RAISING 결정의 종목 ∪ ③ 이번 결정의 종목. **|S| > 상한이면 거부**.
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
- 결과: |S| > 상한 → `ErrReservationLimitExceeded`(LIMIT_REACHED, 재수집 안 함). 상세 문구에 개수 · 상한 · 이번 종목이 새 종목인지.

## D4. 설정 · 정책

- `config.AutomationGate.MaxOpenPositions float64 \`json:"max_open_positions,omitempty"\``. 숫자를 float 로 받는 이유: 형제 칸과 같고, `int` 로 받으면
  `6.5` 같은 값이 **게이트 OFF 에서도** config 읽기 전체를 실패시켜 OFF = upstream(§0.2)을 깸. 정수성은 게이트 ON 정책 검증이 봄.
- 해석 `EffectiveMaxOpenPositions()`(새 파일 `internal/config/open_positions.go`): 0 이 아니면 그대로(음수 · 비정수 · NaN · Inf 도 그대로 넘겨 검증이 거부), 0 이면 기본값 6.
  통화와 무관. Guardian 정책 유도와 인터록 대조가 모두 이 함수를 씀(단일 출처).
- `risk.Policy.MaxOpenPositions int`, `DefaultPolicy` = 6(출처: 사용자 결정 2026-10-15, a138 — 이전에 상한이 없던 새 검사라 완화 아님),
  `Validate` 는 `< 1` 거부. 유도(`riskPolicyFromAutomationGate`)는 해석 값이 정수이고 1 이상이며 int 범위일 때만 그 값을 쓰고, 아니면 0 을 넣어
  `Validate` 가 `ErrLimitsRequired` 로 기동을 멈추게 함(a136 이 음수를 넘겨 검증이 거부하게 한 것과 같은 길).
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
- CLOSING(청산 주문이 나갔지만 아직 수량이 남은) 포지션 — 한 칸으로 셈.
- D1-B: 손으로 든 종목 — 한 칸씩. 외부 보유가 상한 이상이면 신규 종목 진입 전부 거부.
- 옛 결정 하나의 preimage 가 해석 불가 — 그 결정이 HELD 예약을 들고 있는 동안 모든 진입 거부(원장 결함이 조용히 0 이 되지 않게).
- 원장 읽기 오류 — 그 진입 거부.
- 거부하지 **않는** 것(permissive 쪽): D1-B 에서 대사 전의 외부 매수(≤30초 창), 미체결 수동 매수 주문(D1-C 를 고르지 않으면).

## D8. 원장 · 스키마 · 성능

스키마 변경 0. 질의는 `positions`(색인 `idx_positions_symbol` 의 선두 account_ref) 와 HELD 예약(`idx_reservations_held`) 두 범위 — 쓰기 잠금 안에서
계좌 평생 행을 훑지 않음(§0.3 간접: 진입 tx 가 잠금을 잡는 동안 매도 기록이 기다림). 결정 preimage 는 HELD 예약의 결정 수만큼(상한 근처라 작음) 읽음.

## D9. 공유 파일 순서 (a139)

`riskPolicyFromAutomationGate` · `risk.DefaultPolicy` · risk-management 「Automation gate limits …」 요구는 a139(건당 위험 예산 = 일일 손실 / 3)도 고친다.
Manager 지시대로 a138 의 Go · spec 을 먼저 전부 커밋하고, a139 델타는 a138 이 남긴 문장 위에 쓴다. a138 이 먼저 archive 되어야 a139 의 MODIFIED 가 충돌 없이 합쳐진다.
