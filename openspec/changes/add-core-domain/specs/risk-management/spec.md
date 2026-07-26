# risk-management Specification (delta)

## ADDED Requirements

### Requirement: Guardian 판정 체인
모든 자동 진입 의도는 Guardian 판정 체인을 통과해야 하며(SHALL), 체인은 고정 순서로 평가되고 첫 실패에서 정지한다: kill switch/운영 모드 → 게이트 상태(진입 차단 latch: 401/403·SLO 위반·RECONCILE·recovery 미완료) → 심볼 allowlist → 구조적 손절 계약(No Stop=No Trade·보호적 stop·실질 본전 미달 target 거부) → 주문 크기 한도 → 최소 RR(신규 검사) → 현금 검증(비용 포함) → 당일 재진입 규칙 → 총 개방 노출 한도 → 일일 손실 한도 → 중복 주문 검사 → ALLOW. 각 거부는 안정적 reason-code로 기록된다(SHALL).

**순서의 권위는 이 표다**(SHALL — StockOS `evaluate_guardian`과의 대응은 `docs/guardian-chain.md` 참고 산출물로 남기되 순서 보존을 규범으로 주장하지 않는다: 원 체인에는 미이식 단계가 섞여 있고 최소 RR은 원 체인에 없는 신규 검사다). 이식 분류의 열거(SHALL):

- **이식**: kill switch·모드, 진입 latch, 손절 계약(reason: STOP_MISSING·STOP_NOT_BELOW_ENTRY·TARGET_NOT_ABOVE_ENTRY·INVALID_TARGET_STOP·TARGET_BELOW_BREAK_EVEN), 주문 크기(INVALID_ORDER_SIZE·MAX_ORDER_EXCEEDED), 현금(INSUFFICIENT_CASH·비용 포함), 당일 재진입(쿨다운 기본 30분 `[미검증]`·당일 심볼당 최대 진입 수 기본 2회 `[미검증]`·**미체결 매수 존재 시 진입 차단 PENDING_BUY_ORDER_BLOCKED**), 총 노출, 일손실, 중복 주문
- **신규**: 최소 RR, 심볼 allowlist, STOP_MISSING(원본은 생성자 강제라 카운터파트 없음)
- **제외**: KIS 고유(CASH_ONLY 등), LLM 게이트, capital stage, ARM/DASHBOARD 확인(운영 UI 소관 — 게이트 flip 승인이 대체), LIVE_DISABLED(인터록이 대체), BUY_PAUSED/SELL_ONLY(운영 모드가 대체), 미국장 진입 시간창, DAILY_TURNOVER·MAX_POSITIONS·CANCEL_RATE(P3), SELL_COST_BUFFER(§0.3)
- **구조 대체**: 레버리지/인버스·ETF/ETN 클래스 차단 → 심볼 allowlist(분류 소스 `[미측정]`, P3 재설계)

체인의 모든 입력은 의도 필드·브로커 스냅샷·journal 상태에서 온다(SHALL — 신호 계층 산물 없음: 구조적 RR 계산·등급배수는 P3). 총계 한도의 경계값은 도달 시 차단(≥)이며(SHALL — 예약 트랜잭션의 실장과 일치), 주문 단위 한도는 포함 상한(초과 시 차단)을 유지한다.

#### Scenario: 체인 첫 실패 정지
- **WHEN** kill switch가 활성인 상태에서 진입 의도가 평가되면
- **THEN** 후속 판정 없이 KILL_SWITCH_ACTIVE로 즉시 거부된다

#### Scenario: allowlist 밖 심볼
- **WHEN** allowlist에 없는 심볼의 진입 의도가 평가되면
- **THEN** SYMBOL_NOT_ALLOWED로 거부된다

#### Scenario: 총계 경계값 도달
- **WHEN** 당일 실현 손실이 한도에 정확히 도달한 상태에서 진입 의도가 평가되면
- **THEN** DAILY_LOSS_LIMIT_REACHED로 거부된다 (계좌자본 0 이하이면 즉시 차단)

### Requirement: No Stop = No Trade와 위험 기반 수량
손절가가 없거나 보호적이지 않은 진입 의도는 수량 계산 이전 단계에서 거부되어야 한다(SHALL). TossOS는 long-only이므로 진입은 매수이고 보호적 손절은 `stop < entry`이며, 매도는 보유수량 이하 reduce-only, short 노출은 구조적으로 금지된다(SHALL NOT). 위험 기반 수량은 `floor(위험예산 / (entry − stop))`이고(등급배수는 P3까지 1.0 고정 — 보수 하한), stop 폭 0 이하는 수량 0(fail-closed)이다(SHALL). 최소 RR은 의도 필드의 순수 산술 `(target − entry) / (entry − stop)`로 검사하며(신규 검사 — 세션 구조 기반 산출은 P3), 기본값 2.0 미달·계산 불가는 거부한다(SHALL — 0 대체 금지. provenance: StockOS parker_vwap §22 lock 2.0; 1.5는 최저 티어 값이라 기각). 진입 손절가는 사전 검사로 끝나지 않고 포지션 개시 시 exit 정책의 t0 기준선이 된다(SHALL — exit-policy).

#### Scenario: 손절 없는 진입
- **WHEN** stop이 없는 진입 의도가 평가되면
- **THEN** STOP_MISSING으로 거부되고 수량 계산은 수행되지 않는다

#### Scenario: RR 미달
- **WHEN** (target−entry)/(entry−stop)이 2.0 미만인 의도가 평가되면
- **THEN** MIN_RR_NOT_MET으로 거부된다

#### Scenario: 보유수량 초과 매도
- **WHEN** 보유수량을 초과하는 매도 의도가 평가되면
- **THEN** short 노출이 되므로 거부된다

### Requirement: 발급 절차와 예약의 권위
발급 절차는 실장 가능한 순서를 따른다(SHALL): 체인 ALLOW → **결정 영속과 예약 삽입을 하나의 journal 트랜잭션으로 수행**(신규 원자 API — 예약의 결정 FK가 만족되면서 결정과 예약 사이 크래시 창이 없다) → Gateway 제출. 총계 한도의 최종 권위는 예약 트랜잭션이며(SHALL), 예약이 거부되면 결정도 같은 트랜잭션에서 함께 롤백되어 제출 가능한 결정이 남지 않는다(SHALL — Gateway의 HELD 예약 검증이 이를 이중으로 강제한다: engine-safety delta). 예약 실패는 원인별 안정 reason-code로 기록된다(SHALL): LIMIT_REACHED(한도 도달·즉시), SNAPSHOT_RECOLLECTION_EXHAUSTED(재수집 상한·데드라인 소진), VERSION_CONFLICT(원장 버전 경합), DECISION_EXPIRED(재수집 중 결정 만료). 재수집 시 체인은 재실행하지 않는다(SHALL — 모드·latch는 Gateway 제출 시 EntryGate 재검사가 커버하고, 결정 만료(60초 — 실장 `DefaultDecisionTTL`)가 나머지 입력의 신선도 상한이다. 재수집 예산 10초 < TTL).

#### Scenario: 예약 거부 시 결정 부재
- **WHEN** 체인 ALLOW 후 동시 결정이 한도 잔여를 소진해 예약이 거부되면
- **THEN** LIMIT_REACHED가 기록되고, 롤백으로 제출 가능한 결정이 존재하지 않는다

#### Scenario: 발급과 제출 사이 모드 강화
- **WHEN** 발급 직후 자동 강화로 ENTRY_BLOCKED가 되면
- **THEN** Gateway 제출 시 EntryGate 재검사가 진입을 거부한다

### Requirement: GuardianDecision 발급자
Guardian은 engine-safety 메인 스펙의 결정 계약을 구현하는 발급자다(SHALL): EXPOSURE_RAISING 결정(RiskIntent preimage·`f(decision_id, generation)` 멱등키·인터록이 감사한 설정 한도 스냅샷·만료 60초·nonce)과 위험 감소의 ReductionIntent 결정(한도 없음)을 발급하고, `ExposureLimiter`를 구현해 자기 한도를 진술한다(SHALL — 인터록 단일 출처 검증 통과). 위험 입력은 발급 요청의 원 의도에서 오고, 발급 후 변조는 Gateway의 journal 기반 재검증이 차단한다.

#### Scenario: 발급자 한도 진술
- **WHEN** 인터록이 Guardian의 ExposureLimits를 감사된 설정 한도와 대조하면
- **THEN** 필드·Set 비트 단위로 일치한다

### Requirement: 정책 수치의 provenance
모든 한도·정책 수치는 코드에 출처(StockOS 파일·검증 상태)와 함께 기록되어야 하며(SHALL), 사용자 미확정 시 보수 기본값 전체 집합을 사용한다(SHALL — 인터록 5필드를 전부 충족): 주문당 notional 1,000,000 KRW·주문당 수량 100주·총 노출 10,000,000 KRW·일일 손실 100,000 KRW·일일 손실 자본비 1%·통화 KRW. "Toss 검증됨" 표시는 실계좌 검증 결과에만 근거해 전환한다(SHALL NOT — 미검증 상태에서 검증됨 표기 금지). 수치 변경은 §0.9(보수 방향만) 검토와 audit 기록을 요구한다(SHALL).

#### Scenario: 한도 완화 시도
- **WHEN** 일일 손실 한도를 높이는 설정 변경이 적용되면
- **THEN** audit 로그에 이전·새 값이 기록되고 변경은 사람 승인 절차를 거친다

#### Scenario: 기본값 게이트 ON 기동
- **WHEN** 사용자 미확정 기본값 집합으로 게이트 ON 기동하면
- **THEN** 인터록 한도 검증(5필드·통화)이 통과한다

### Requirement: Kill switch와 운영 모드 — 모드×클래스 표
kill switch는 신규 진입 차단 전용(BLOCK-ONLY)이며 어떤 소비자도 강제청산을 유발하지 않는다(SHALL NOT). 운영 모드는 mutation safety class 허용 표로 정의된다(SHALL):

| 모드 | EXPOSURE_RAISING | RISK_REDUCING | PROTECTION_WEAKENING* |
|---|---|---|---|
| NORMAL | 허용 | 허용 | 허용(audit) |
| ENTRY_BLOCKED | 거부 | 허용 | 허용(audit) |
| HALT_ALL | 거부 | 허용 | **거부** |

*PROTECTION_WEAKENING의 발급·소비는 보호주문 도입 change가 정의하며 이 표는 열을 예약한다 — **현재는 landed `RecordDecision`이 전 모드에서 이 class를 거부하므로 허용(audit) 셀은 2c 발급 도입 후에 효력을 갖는다**. HALT_ALL은 이 change 안에서는 ENTRY_BLOCKED와 행동이 같지만, 운영자 전용 진입(자동 강화 없음)이라는 승인 의미와 2c의 PROTECTION_WEAKENING 거부로 구별되므로 유지한다. EXIT_ONLY는 두지 않는다 — ENTRY_BLOCKED와 행동이 동일해지므로(구별 근거 없는 모드는 §0.7 승인 사다리의 무의미 단계다), 실제 행동 차이가 생기는 change가 재도입한다. 수동 flatten-all은 모든 모드에서 통과한다(§0.3).

모드의 강제 지점은 EntryGate 투영이다(SHALL — 모드 전환은 journal 영속과 동시에 EntryGate 계좌 latch로 투영되고, Gateway의 기존 제출 시 재검사가 이를 소비한다; 봉인된 제출 시퀀스는 변경되지 않는다). 전환 승인은 방향 비대칭이다(SHALL): 보수 방향(NORMAL→ENTRY_BLOCKED→HALT_ALL)은 자동·즉시·durable, 완화는 사람 승인(§0.7)+audit. 자동 강화 트리거와 목적 상태의 열거(SHALL): 일일 손실 한도 도달 → ENTRY_BLOCKED, 자격증명 실패(401/403) → ENTRY_BLOCKED, critical 알림 outbox 전달 실패 지속 → ENTRY_BLOCKED, exit 관측 두절 임계 초과 → ENTRY_BLOCKED — 전부 메인 스펙의 "신규 진입 차단"과 정합하며 HALT_ALL 자동 진입은 없다(SHALL NOT — HALT_ALL은 운영자 결정). 분석·성과 작업 실패는 트리거가 아니다(SHALL NOT). 모드·kill switch·이력은 journal 영속·재시작 유지(SHALL), 동시 적용 시 보수 우선(SHALL).

#### Scenario: 손실 한도 도달 시 자동 강화
- **WHEN** 일일 손실 한도 도달로 자동 강화가 발동하면
- **THEN** 사람 승인 없이 ENTRY_BLOCKED로 즉시 전환·영속되고 EntryGate에 투영되며 알림이 발송된다

#### Scenario: HALT_ALL 중 청산
- **WHEN** HALT_ALL 상태에서 RISK_REDUCING 청산이 요청되면
- **THEN** 허용된다 (수동 flatten-all 포함)

#### Scenario: 분석 작업 실패는 비트리거
- **WHEN** 성과 집계 작업이 반복 실패하면
- **THEN** 운영 모드는 변하지 않고 분석 재시도만 수행된다

#### Scenario: 재시작 후 모드 유지
- **WHEN** ENTRY_BLOCKED에서 프로세스가 재시작되면
- **THEN** journal에서 복원되어 EntryGate 투영과 함께 유지된다

#### Scenario: 모드 완화 시도
- **WHEN** ENTRY_BLOCKED에서 NORMAL로 되돌리려 하면
- **THEN** 사람 승인 절차와 audit 기록을 거쳐야 한다
