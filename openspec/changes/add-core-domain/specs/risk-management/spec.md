# risk-management Specification (delta)

## ADDED Requirements

### Requirement: Guardian 판정 체인
모든 자동 진입 의도는 Guardian 판정 체인을 통과해야 하며(SHALL), 체인은 고정 순서로 평가되고 첫 실패에서 정지한다: kill switch/운영 모드 → 게이트 상태(진입 차단 latch: 401/403·SLO 위반·reconcile 불일치·recovery 미완료) → 주문 크기 한도 → 구조적 손절 계약 → 최소 RR → 현금·비용 검증 → 중복·재진입 규칙 → 총 개방 노출 한도 → 일일 손실 한도 → ALLOW. 각 거부는 안정적 reason-code로 기록된다(SHALL). 판정 순서는 StockOS `evaluate_guardian`의 검증된 순서를 보존하며 순서 변경은 spec 개정을 요구한다(SHALL).

#### Scenario: 체인 첫 실패 정지
- **WHEN** kill switch가 활성인 상태에서 진입 의도가 평가되면
- **THEN** 후속 판정 없이 KILL_SWITCH_ACTIVE로 즉시 거부된다

#### Scenario: 일일 손실 한도 도달
- **WHEN** 당일 실현 손실이 절대액 또는 계좌자본 % 한도 중 하나라도 도달한 상태에서 진입 의도가 평가되면
- **THEN** DAILY_LOSS_LIMIT_REACHED로 거부된다 (계좌자본 0 이하이면 즉시 차단)

### Requirement: No Stop = No Trade
손절가가 없거나 보호적이지 않은(매수 기준 stop ≥ entry) 진입 의도는 수량 계산 이전 단계에서 거부되어야 한다(SHALL). 위험 기반 수량은 `floor(위험예산 × 등급배수 / (entry − stop))`로 계산하고 stop 폭이 0 이하이면 수량 0(fail-closed)이다(SHALL). 최소 RR(기본 1.5) 미달·RR 계산 불가는 거부한다(SHALL — 0 대체 금지).

#### Scenario: 손절 없는 진입
- **WHEN** stop이 없는 진입 의도가 평가되면
- **THEN** STOP_MISSING으로 거부되고 수량 계산은 수행되지 않는다

#### Scenario: stop 폭 0
- **WHEN** entry와 stop이 같은 의도가 평가되면
- **THEN** 수량이 0으로 계산되어 INVALID_ORDER_SIZE로 거부된다

### Requirement: 한도 수치의 provenance
모든 한도·정책 수치는 코드에 출처(StockOS 파일·검증 상태)와 함께 기록되어야 하며(SHALL), 사용자 미확정 시 보수 기본값(small_live: 주문당 1,000,000 KRW / 총 노출 10,000,000 KRW / 일일 손실 100,000 KRW 또는 1%)을 사용한다. 수치 변경은 §0.9(보수 방향만) 검토와 audit 기록을 요구한다(SHALL).

#### Scenario: 한도 완화 시도
- **WHEN** 일일 손실 한도를 높이는 설정 변경이 적용되면
- **THEN** audit 로그에 이전·새 값이 기록되고 변경은 사람 승인 절차를 거친다

### Requirement: Kill switch와 운영 모드
kill switch는 신규 진입 차단 전용(BLOCK-ONLY)이며 어떤 소비자도 강제청산을 유발하지 않는다(SHALL NOT). 운영 모드 축은 NORMAL / ENTRY_BLOCKED / EXIT_ONLY / HALT_ALL로 정의되고(SHALL): ENTRY_BLOCKED=진입만 차단, EXIT_ONLY=진입 차단+청산 허용+신규 보호주문 허용, HALT_ALL=모든 자동 제출 중단(수동 flatten-all은 예외). 모드는 journal에 영속되어 재시작 후 유지되고, 전환은 audit 기록·critical 알림·사람 승인(§0.7)을 요구한다(SHALL).

#### Scenario: HALT_ALL 중 flatten-all
- **WHEN** HALT_ALL 모드에서 운영자가 flatten-all을 실행하면
- **THEN** 수동 명령은 정상 수행된다 (§0.3 즉시성 보존)

#### Scenario: 재시작 후 모드 유지
- **WHEN** EXIT_ONLY 모드에서 프로세스가 재시작되면
- **THEN** 복구 후에도 EXIT_ONLY가 유지된다

### Requirement: GuardianDecision 발급
Guardian ALLOW는 execgw 계약(intent 해시, 한도 스냅샷, 만료 시각, one-shot nonce)을 만족하는 GuardianDecision으로 발급되어야 하며(SHALL), 결정 만료(기본 5초) 후 제출·nonce 재사용은 Gateway에서 거부된다. 청산·보호주문 의도는 진입 한도 판정을 면제하되 kill switch BLOCK-ONLY 원칙과 운영 모드 규칙은 적용된다(SHALL).

#### Scenario: 만료된 결정으로 제출
- **WHEN** 발급 후 만료 시간이 지난 GuardianDecision으로 제출하면
- **THEN** Gateway가 거부하고 재평가를 요구한다

### Requirement: 자동화 게이트 활성화
자동화 게이트 ON은 다음이 모두 충족될 때만 가능하다(SHALL): 유효 attestation(P1 인터록), 0이 아닌 한도의 Guardian 주입, 운영 모드 NORMAL 또는 ENTRY_BLOCKED, 사람 승인 flip(§0.7 + audit). 이 배선의 통합 테스트는 미충족 조합 전부의 기동 거부를 검증한다(SHALL).

#### Scenario: Guardian 한도 전부 0으로 기동
- **WHEN** 게이트 ON + 모든 한도가 0인 Guardian으로 기동하면
- **THEN** 기동이 거부된다
