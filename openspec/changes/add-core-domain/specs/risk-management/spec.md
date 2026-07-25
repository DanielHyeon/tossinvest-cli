# risk-management Specification (delta)

## ADDED Requirements

### Requirement: Guardian 판정 체인
모든 자동 진입 의도는 Guardian 판정 체인을 통과해야 하며(SHALL), 체인은 고정 순서로 평가되고 첫 실패에서 정지한다: kill switch/운영 모드 → 게이트 상태(진입 차단 latch: 401/403·SLO 위반·reconcile 불일치·recovery 미완료) → 주문 크기 한도 → 구조적 손절 계약 → 최소 RR → 현금·비용 검증 → 중복·재진입 규칙 → 총 개방 노출 한도 → 일일 손실 한도 → ALLOW. 각 거부는 안정적 reason-code로 기록된다(SHALL). 판정 순서는 StockOS `evaluate_guardian`의 검증된 순서를 보존하며 순서 변경은 spec 개정을 요구한다(SHALL). 이식 범위는 문서화되어야 한다(SHALL): 제외 항목(KIS 고유·LLM 게이트·capital stage·미국장 진입 시간창)과 판단 보류 항목(레버리지/인버스 클래스·ETF/ETN 클래스·당일 재진입 쿨다운)을 명시하고, 보류 항목을 임의로 이식하거나 누락해서는 안 된다(SHALL NOT).

#### Scenario: 체인 첫 실패 정지
- **WHEN** kill switch가 활성인 상태에서 진입 의도가 평가되면
- **THEN** 후속 판정 없이 KILL_SWITCH_ACTIVE로 즉시 거부된다

#### Scenario: 일일 손실 한도 도달
- **WHEN** 당일 실현 손실이 절대액 또는 계좌자본 % 한도 중 하나라도 도달한 상태에서 진입 의도가 평가되면
- **THEN** DAILY_LOSS_LIMIT_REACHED로 거부된다 (계좌자본 0 이하이면 즉시 차단)

### Requirement: No Stop = No Trade
손절가가 없거나 보호적이지 않은 진입 의도는 수량 계산 이전 단계에서 거부되어야 한다(SHALL). TossOS는 long-only이므로 진입은 매수이고 보호적 손절은 `stop < entry`를 뜻하며, 매도는 보유수량 이하 reduce-only로만 허용되고 short 노출은 구조적으로 금지된다(SHALL NOT). 위험 기반 수량은 `floor(위험예산 × 등급배수 / (entry − stop))`로 계산하고 stop 폭이 0 이하이면 수량 0(fail-closed)이다(SHALL). 최소 RR(기본 1.5) 미달·RR 계산 불가는 거부한다(SHALL — 0 대체 금지).

#### Scenario: 손절 없는 진입
- **WHEN** stop이 없는 진입 의도가 평가되면
- **THEN** STOP_MISSING으로 거부되고 수량 계산은 수행되지 않는다

#### Scenario: stop 폭 0
- **WHEN** entry와 stop이 같은 의도가 평가되면
- **THEN** 수량이 0으로 계산되어 INVALID_ORDER_SIZE로 거부된다

#### Scenario: 보유수량 초과 매도
- **WHEN** 보유수량을 초과하는 매도 의도가 평가되면
- **THEN** short 노출이 되므로 거부된다

### Requirement: 총계 한도의 계산 계약
일일 손실과 총 개방 노출은 구현마다 달라질 수 없도록 계산 계약이 명시되어야 한다(SHALL). 각 값에 대해 다음을 정의한다: 권위 데이터 출처, 통화 정규화 규칙(USD 자산의 KRW 환산 시세와 그 staleness 허용치), 거래일 경계(시장별 timezone·DST — P1 시간 규율 준수), 실현 손실과 미실현 손실의 포함 여부, 수수료·거래세·환전 비용 반영, 외부 수동 거래의 취급, 미체결 매수와 위험 예약의 노출 포함 여부, 계좌자본 분모의 측정 시점. 입력 중 하나라도 stale하거나 미지이면 판정은 fail-closed(진입 거부)여야 한다(SHALL).

#### Scenario: 환율 시세 stale
- **WHEN** USD 보유 자산의 KRW 환산에 필요한 시세가 staleness 임계를 초과하면
- **THEN** 총 개방 노출 판정이 fail-closed로 진입을 거부한다

#### Scenario: 미체결 매수의 노출 포함
- **WHEN** 미체결 매수 주문과 유효한 위험 예약이 존재하는 상태에서 총 개방 노출을 계산하면
- **THEN** 그 금액이 노출에 포함되어 한도가 이중으로 사용되지 않는다

### Requirement: 정책 수치의 provenance
모든 한도·정책 수치는 코드에 출처(StockOS 파일·검증 상태)와 함께 기록되어야 하며(SHALL), 사용자 미확정 시 보수 기본값(small_live: 주문당 1,000,000 KRW / 총 노출 10,000,000 KRW / 일일 손실 100,000 KRW 또는 1%)을 사용한다. 등급 배수와 비용 bps는 개별 값으로 열거되어야 하고(SHALL), 비용은 과대 추정이 보수 방향이다(SHALL — 과소 추정은 Guardian의 현금·비용 검증과 R 배수를 동시에 낙관적으로 만든다). 수치의 "Toss 검증됨" 표시는 실계좌 검증 결과에만 근거해 전환할 수 있다(SHALL NOT — 미검증 상태에서 검증됨으로 표기 금지). 수치 변경은 §0.9(보수 방향만) 검토와 audit 기록을 요구한다(SHALL).

#### Scenario: 한도 완화 시도
- **WHEN** 일일 손실 한도를 높이는 설정 변경이 적용되면
- **THEN** audit 로그에 이전·새 값이 기록되고 변경은 사람 승인 절차를 거친다

#### Scenario: 미검증 비용 수치
- **WHEN** Toss 실측 검증이 완료되지 않은 비용 bps로 판정이 수행되면
- **THEN** 해당 수치는 미검증으로 표시되고 과대 추정 값이 사용된다

### Requirement: Kill switch와 운영 모드
kill switch는 신규 진입 차단 전용(BLOCK-ONLY)이며 어떤 소비자도 강제청산을 유발하지 않는다(SHALL NOT). 운영 모드 축은 NORMAL / ENTRY_BLOCKED / EXIT_ONLY / HALT_ALL로 정의되고(SHALL): ENTRY_BLOCKED=진입만 차단, EXIT_ONLY=진입 차단+청산 허용+신규 보호주문 허용, HALT_ALL=모든 **노출 증가** 중단. HALT_ALL에서도 위험 감소 mutation(보호주문 생성·증량, reduce-only 청산, 취소)은 계속 허용되어야 한다(SHALL — 그렇지 않으면 재시작 시 발견된 미보호 포지션에 손절을 걸 수 없어 No Stop = No Trade와 충돌한다). kill switch와 모드가 동시에 적용될 때는 더 보수적인 쪽이 이기며, 조합표를 산출물로 문서화한다(SHALL).

모드 전환의 승인 규칙은 방향 비대칭이다(SHALL): 보수 방향 전환(NORMAL→ENTRY_BLOCKED→EXIT_ONLY→HALT_ALL)은 사람 승인 없이 자동·즉시 수행되고 즉시 durable하게 기록되며, 완화·해제만 사람 승인(§0.7)과 audit를 요구한다. 모드와 kill switch 상태는 journal에 영속되어 재시작 후 유지된다(SHALL).

#### Scenario: HALT_ALL 중 미보호 포지션 발견
- **WHEN** HALT_ALL 상태에서 재시작 복구가 보호 없는 포지션을 발견하면
- **THEN** 보호주문 제출이 허용되어 손절이 등록된다

#### Scenario: 손실 한도 도달 시 자동 강화
- **WHEN** 일일 손실 한도 도달로 ENTRY_BLOCKED 전환이 필요하면
- **THEN** 사람 승인을 기다리지 않고 즉시 전환·기록되며 알림이 발송된다

#### Scenario: 모드 완화 시도
- **WHEN** EXIT_ONLY에서 NORMAL로 되돌리려 하면
- **THEN** 사람 승인 절차와 audit 기록을 거쳐야 한다

#### Scenario: 재시작 후 모드 유지
- **WHEN** EXIT_ONLY 모드에서 프로세스가 재시작되면
- **THEN** 복구 후에도 EXIT_ONLY가 유지된다

### Requirement: GuardianDecision 발급
Guardian ALLOW는 실행 계약이 요구하는 GuardianDecision(주문 intent 해시, RiskIntent 해시, 한도 스냅샷, 만료 시각, one-shot nonce)으로 발급되어야 하며(SHALL), 총계 한도를 소비하는 진입 결정은 위험 예약과 같은 트랜잭션에서 발급된다(SHALL). 위험 감소 의도(보호주문·청산)의 결정은 빈 한도 스냅샷을 실어 청산이 주문 한도로 거부되지 않게 한다(SHALL NOT 한도 스냅샷 부여). 위험 감소 의도는 진입 한도 판정을 면제하되 kill switch BLOCK-ONLY 원칙과 운영 모드 규칙은 적용된다(SHALL).

#### Scenario: 만료된 결정으로 제출
- **WHEN** 발급 후 만료 시간이 지난 GuardianDecision으로 제출하면
- **THEN** Gateway가 거부하고 재평가를 요구한다

#### Scenario: 청산 결정의 빈 한도
- **WHEN** 주문당 최대 수량을 초과하는 포지션의 청산 결정이 발급되면
- **THEN** 한도 스냅샷이 비어 있어 Gateway가 수량을 이유로 거부하지 않는다

### Requirement: 게이트 활성화 배선
엔진은 실행 계약의 기동 인터록 전제조건을 실제로 충족시키도록 배선되어야 한다(SHALL): ExecutionGateway 구성, 인터록이 감사한 설정 한도에서 Guardian 구성(단일 출처), 운영 모드 스냅샷 주입, 위험 예약·nonce 저장소 연결. 게이트는 기본 OFF이며 ON은 사람 승인 flip을 요구한다(SHALL). 미충족 조합에서의 기동 거부는 통합 테스트로 검증한다(SHALL).

#### Scenario: Guardian 한도가 감사된 설정과 불일치
- **WHEN** 주입된 Guardian이 인터록이 검증한 설정 한도와 다른 한도를 찍으면
- **THEN** 기동이 거부된다

#### Scenario: Gateway 미배선 기동
- **WHEN** 게이트 ON 상태인데 엔진에 ExecutionGateway가 구성되지 않았으면
- **THEN** 기동이 거부된다
