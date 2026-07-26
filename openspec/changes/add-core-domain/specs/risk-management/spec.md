# risk-management Specification (delta)

## ADDED Requirements

### Requirement: Guardian 판정 체인
모든 자동 진입 의도는 Guardian 판정 체인을 통과해야 하며(SHALL), 체인은 고정 순서로 평가되고 첫 실패에서 정지한다: kill switch/운영 모드 → 게이트 상태(진입 차단 latch: 401/403·SLO 위반·RECONCILE·recovery 미완료) → 심볼 allowlist → 구조적 손절 계약(No Stop = No Trade) → 주문 크기 한도 → 최소 RR → 현금·비용 검증(실질 본전 미달 target 거부 포함) → 중복·재진입 규칙(당일 재진입 쿨다운 포함) → 총 개방 노출 한도 → 일일 손실 한도 → ALLOW. 각 거부는 안정적 reason-code로 기록된다(SHALL).

순서의 권위는 이 표이며, 이식된 검사에 한해 StockOS `evaluate_guardian`의 상대 순서를 보존하고 그 매핑을 문서 산출물로 남긴다(SHALL — StockOS 원 체인에는 이식하지 않는 단계가 섞여 있으므로 "원 순서 보존"을 전체에 주장하지 않는다). 이식 범위는 열거되어야 한다(SHALL): **이식** — kill switch·모드, 진입 latch, 손절 계약, 주문 크기, 현금·비용(TARGET_BELOW_BREAK_EVEN 포함), 중복·재진입 쿨다운, 총 노출, 일일 손실. **제외** — KIS 고유(CASH_ONLY_REQUIRED 등), LLM 게이트, capital stage, 미국장 진입 시간창, DAILY_TURNOVER·MAX_POSITIONS·CANCEL_RATE(P3 스케줄러·슬롯 소관), SELL_COST_BUFFER(§0.3 — 비용은 청산 게이트가 될 수 없다). **구조 대체** — 레버리지/인버스·ETF/ETN 클래스 차단은 상품 분류 데이터 소스가 없으므로(`[미측정]`) 심볼 allowlist로 대체하며, P3 유니버스 확장 시 분류 게이트를 재설계한다.

체인의 모든 입력은 의도 필드·브로커 스냅샷·journal 상태에서 온다(SHALL — 신호 계층 산물을 요구하는 검사는 이 change에 없다: 구조적 RR 계산·등급배수는 P3).

#### Scenario: 체인 첫 실패 정지
- **WHEN** kill switch가 활성인 상태에서 진입 의도가 평가되면
- **THEN** 후속 판정 없이 KILL_SWITCH_ACTIVE로 즉시 거부된다

#### Scenario: allowlist 밖 심볼
- **WHEN** allowlist에 없는 심볼의 진입 의도가 평가되면
- **THEN** SYMBOL_NOT_ALLOWED로 거부된다

#### Scenario: 일일 손실 한도 도달
- **WHEN** 당일 실현 손실이 절대액 또는 계좌자본 % 한도 중 하나라도 도달한 상태에서 진입 의도가 평가되면
- **THEN** DAILY_LOSS_LIMIT_REACHED로 거부된다 (계좌자본 0 이하이면 즉시 차단)

### Requirement: No Stop = No Trade와 위험 기반 수량
손절가가 없거나 보호적이지 않은 진입 의도는 수량 계산 이전 단계에서 거부되어야 한다(SHALL). TossOS는 long-only이므로 진입은 매수이고 보호적 손절은 `stop < entry`이며, 매도는 보유수량 이하 reduce-only, short 노출은 구조적으로 금지된다(SHALL NOT). 위험 기반 수량은 `floor(위험예산 / (entry − stop))`로 계산하고(등급배수는 P3 신호 계층이 등급을 생산할 때까지 1.0 고정 — 보수 하한), stop 폭이 0 이하이면 수량 0(fail-closed)이다(SHALL). 최소 RR은 의도 필드의 순수 산술 `(target − entry) / (entry − stop)`로 검사하며(SHALL — 세션 구조 기반 산출은 P3), 기본값 2.0 미달·계산 불가는 거부한다(SHALL — 0 대체 금지. provenance: StockOS parker_vwap §22 lock #5의 잠금값 2.0. 1.5는 StockOS 내 최저 티어 값이므로 기본값으로 쓰지 않는다).

#### Scenario: 손절 없는 진입
- **WHEN** stop이 없는 진입 의도가 평가되면
- **THEN** STOP_MISSING으로 거부되고 수량 계산은 수행되지 않는다

#### Scenario: RR 미달
- **WHEN** (target−entry)/(entry−stop)이 2.0 미만인 의도가 평가되면
- **THEN** MIN_RR_NOT_MET으로 거부된다

#### Scenario: 보유수량 초과 매도
- **WHEN** 보유수량을 초과하는 매도 의도가 평가되면
- **THEN** short 노출이 되므로 거부된다

### Requirement: 판정 체인과 예약의 권위 관계
체인은 스냅샷 위의 순수 사전 검사이고, 총계 한도(총 노출·일손실·현금)의 **최종 권위는 원자적 위험 예약 트랜잭션이다**(SHALL — order-execution 메인 스펙). 발급 절차: 체인 ALLOW → 예약 트랜잭션(as-of 재검증·삽입) → 결정 영속·발급. 체인이 ALLOW를 냈어도 예약이 거부하면 RESERVATION_CONFLICT reason-code로 거부되며(SHALL), 재수집 상한 초과 시 fail-closed다. 재수집 시 체인을 다시 돌지 않는다(SHALL — 체인 입력 중 예약 트랜잭션이 재검증하지 않는 것이 바뀌었을 가능성은 결정 만료(기본 5초)가 담보한다).

#### Scenario: 체인 통과 후 예약 거부
- **WHEN** 체인이 ALLOW를 낸 뒤 동시 결정이 한도 잔여를 소진해 예약이 거부되면
- **THEN** RESERVATION_CONFLICT로 거부되고 진입은 발생하지 않는다

### Requirement: GuardianDecision 발급자
Guardian은 engine-safety 메인 스펙의 결정 계약을 구현하는 발급자다(SHALL): 체인 ALLOW를 EXPOSURE_RAISING 결정(RiskIntent preimage — 계좌·시장·심볼·방향·진입가·손절가·목표가·수량·정책 버전 — journal 영속, `f(decision_id, generation)` 멱등키, 인터록이 감사한 설정 한도 스냅샷, 만료 기본 5초, nonce)으로 변환하고, 위험 감소 의도는 ReductionIntent preimage의 RISK_REDUCING 결정으로(한도 스냅샷 없음) 발급한다. 발급자는 `ExposureLimiter`를 구현해 자기 한도를 진술한다(SHALL — 인터록 단일 출처 검증). 위험 입력(진입가·손절가·목표가)은 제출자가 아니라 발급 요청의 원 의도에서 오며, 발급 후 변조는 Gateway의 journal 기반 재검증이 차단한다.

#### Scenario: 만료된 결정으로 제출
- **WHEN** 발급 후 만료 시간이 지난 결정으로 제출하면
- **THEN** Gateway가 브로커 호출 전에 거부하고 재평가를 요구한다

#### Scenario: 청산 결정의 한도 면제
- **WHEN** 주문당 최대 수량을 초과하는 포지션의 청산 결정이 발급되면
- **THEN** RISK_REDUCING 결정이므로 한도로 거부되지 않는다

### Requirement: 정책 수치의 provenance
모든 한도·정책 수치는 코드에 출처(StockOS 파일·검증 상태)와 함께 기록되어야 하며(SHALL), 사용자 미확정 시 보수 기본값(small_live: 주문당 1,000,000 KRW / 총 노출 10,000,000 KRW / 일일 손실 100,000 KRW 또는 1%)을 사용한다. "Toss 검증됨" 표시는 실계좌 검증 결과에만 근거해 전환한다(SHALL NOT — 미검증 상태에서 검증됨 표기 금지). 수치 변경은 §0.9(보수 방향만) 검토와 audit 기록을 요구한다(SHALL).

#### Scenario: 한도 완화 시도
- **WHEN** 일일 손실 한도를 높이는 설정 변경이 적용되면
- **THEN** audit 로그에 이전·새 값이 기록되고 변경은 사람 승인 절차를 거친다

### Requirement: Kill switch와 운영 모드 — 모드×클래스 표
kill switch는 신규 진입 차단 전용(BLOCK-ONLY)이며 어떤 소비자도 강제청산을 유발하지 않는다(SHALL NOT). 운영 모드의 의미는 mutation safety class에 대한 허용 표로 정의된다(SHALL — 어휘는 engine-safety의 3-클래스):

| 모드 | EXPOSURE_RAISING | RISK_REDUCING | PROTECTION_WEAKENING* |
|---|---|---|---|
| NORMAL | 허용 | 허용 | 허용(audit) |
| ENTRY_BLOCKED | 거부 | 허용 | 허용(audit) |
| EXIT_ONLY | 거부 | 허용 | 허용(audit — 부분 청산에 수반되는 보호 축소) |
| HALT_ALL | 거부 | 허용 | **거부** |

*PROTECTION_WEAKENING의 발급·소비는 보호주문 도입 change(2c)가 정의하며, 이 표는 그 열을 예약한다. 수동 flatten-all은 모든 모드에서 통과한다(§0.3 — RISK_REDUCING).

모드 전환의 승인 규칙은 방향 비대칭이다(SHALL): 보수 방향(NORMAL→ENTRY_BLOCKED→EXIT_ONLY→HALT_ALL)은 사람 승인 없이 자동·즉시·durable하게 수행되고, 완화·해제만 사람 승인(§0.7)과 audit를 요구한다. 자동 강화 트리거는 열거된다(SHALL): 일일 손실 한도 도달, 자격증명 실패(401/403), **critical 알림 outbox** 전달 실패 지속 — 분석·성과 작업의 실패는 트리거가 아니다(SHALL NOT — trade-analytics 격리). 모드·kill switch·전환 이력은 journal에 영속되어 재시작 후 유지된다(SHALL). kill switch와 모드가 동시 적용될 때는 더 보수적인 쪽이 이긴다(SHALL).

#### Scenario: 손실 한도 도달 시 자동 강화
- **WHEN** 일일 손실 한도 도달로 ENTRY_BLOCKED 전환이 필요하면
- **THEN** 사람 승인 없이 즉시 전환·영속되고 알림이 발송된다

#### Scenario: HALT_ALL 중 청산
- **WHEN** HALT_ALL 상태에서 RISK_REDUCING 청산이 요청되면
- **THEN** 허용된다 (수동 flatten-all 포함)

#### Scenario: 분석 작업 실패는 비트리거
- **WHEN** 성과 집계 outbox 작업이 반복 실패하면
- **THEN** 운영 모드는 변하지 않고 분석 재시도만 수행된다

#### Scenario: 재시작 후 모드 유지
- **WHEN** EXIT_ONLY 모드에서 프로세스가 재시작되면
- **THEN** 복구 후에도 EXIT_ONLY가 유지된다

#### Scenario: 모드 완화 시도
- **WHEN** EXIT_ONLY에서 NORMAL로 되돌리려 하면
- **THEN** 사람 승인 절차와 audit 기록을 거쳐야 한다
