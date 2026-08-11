## ADDED Requirements

### Requirement: 보호는 journal에 커밋된 체결에서만 유도된다
시스템은 브로커 상주 보호주문을 **journal에 커밋된 체결로부터만** 계획해야 한다 (SHALL).
보호 계획은 체결의 durable 커밋이 반환된 뒤에 시작해야 하며 (SHALL), 커밋 이전이나 커밋과
동시에 브로커 보호 mutation을 발행해서는 안 된다 (MUST NOT). 공식 API는 per-fill 식별자를
제공하지 않으므로 촉발 단위는 개별 체결 이벤트가 아니라 **한 관측 주기가 커밋한 양(+)의 누적
delta 하나**여야 한다 (SHALL). 시스템은 그 delta 하나당 정확히 한 번 수렴을 시도해야 하며
(SHALL), 여러 delta를 모으는 추가 배칭 window를 도입해서는 안 된다 (MUST NOT).

보호 수량·trigger·만료는 journal에 기록된 값에서 정확히 유도해야 하며 (SHALL), 심볼·시각
근사나 브로커 응답 추정으로 대체해서는 안 된다 (MUST NOT). 누적 수량이 변하지 않은 체결
정정(execution correction)은 보호 교체를 촉발해서는 안 된다 (MUST NOT). 이미 관측한 상태와
모순되는 스냅샷은 보호를 변경하지 않고 해당 포지션의 진입을 닫아야 한다 (SHALL).

#### Scenario: 커밋 이후에만 계획한다
- **WHEN** 체결 스냅샷이 durable 커밋을 반환하기 전에 프로세스가 사라진다
- **THEN** 브로커 보호 mutation은 0건이고 재기동은 커밋된 체결만을 보호 계획의 입력으로 삼는다

#### Scenario: 한 주기의 누적 delta
- **WHEN** 한 관측 주기 사이에 거래소 부분체결이 여러 건 발생해 누적 수량이 한 번에 증가한다
- **THEN** 수렴은 그 증가분 전체에 대해 한 번만 일어나고 부분체결 건수만큼 브로커 왕복이 발생하지 않는다

#### Scenario: 수량이 변하지 않은 정정
- **WHEN** 브로커가 이미 보고한 체결을 누적 수량 변화 없이 재진술한다
- **THEN** 보호 교체와 브로커 mutation은 0건이고 기존 보호주문은 그대로 유지된다

#### Scenario: 모순되는 스냅샷
- **WHEN** 직전 관측보다 작은 누적 체결 수량이 도착한다
- **THEN** 보호주문을 취소·교체하지 않고 해당 포지션의 신규 진입을 닫으며 typed reconcile reason을 기록한다

### Requirement: 보호 상태는 additive-nullable journal 컬럼으로만 영속된다
시스템은 보호 lifecycle 상태와 브로커 order id를 기존 trading journal에 영속해야 한다 (SHALL).
별도의 보호 전용 데이터베이스를 기동 의존성으로 도입해서는 안 된다 (MUST NOT).
스키마 변경은 additive이고 새 컬럼은 nullable이어야 하며 (SHALL), 기존 컬럼의 의미를 바꾸어서는
안 된다 (MUST NOT). 값이 없는 행은 「보호 미설치」로 읽혀야 하고 (SHALL), 새 컬럼을 모르는
이전 바이너리에서도 기존 체결·대사·청산 경로가 동일하게 동작해야 한다 (SHALL).

브로커 보호주문 등록이 확인되면 그 order id를 journal에 커밋해야 한다 (SHALL). 롤백과
스키마 하향은 신규 진입을 OFF로 유지해야 하며 (SHALL), **이미 브로커에 상주하는 보호주문을
취소해서는 안 된다** (MUST NOT).

#### Scenario: 새 컬럼을 모르는 바이너리로 롤백
- **WHEN** 보호 컬럼이 채워진 journal을 그 컬럼을 모르는 이전 바이너리가 연다
- **THEN** 체결 감지·대사·reduce-only 청산이 그대로 동작하고 브로커의 기존 보호주문은 취소되지 않으며 신규 진입은 OFF다

#### Scenario: 등록 응답과 커밋 사이의 프로세스 손실
- **WHEN** 브로커가 보호주문을 수락한 뒤 order id를 journal에 커밋하기 전에 엔진이 종료된다
- **THEN** 재기동은 stable operation identity와 exact broker 조회로 기존 주문을 귀속하고 attested idempotency 증명 없이 재제출하지 않는다

#### Scenario: 보호 미설치 행
- **WHEN** 보호 컬럼이 NULL인 기존 포지션 행을 읽는다
- **THEN** 「보호 미설치」로 해석되어 그 포지션의 신규 진입이 닫히고 기존 청산 경로는 영향을 받지 않는다

### Requirement: coverage 부족은 그 포지션의 신규 진입만 닫는다
시스템은 보호 coverage를 **포지션 단위로** 판정해야 한다 (SHALL). 브로커가 확인한 보호 수량과
그 포지션의 다른 매도 청구권의 합이 보유 수량과 정확히 같고 해당 보호주문이 ACTIVE이며 어떤
latch도 걸려 있지 않을 때만 그 포지션의 신규 진입을 열어야 한다 (SHALL). 그 외의 모든 상태
— 미보호, pending, unknown, reconciling, terminal — 에서는 닫아야 한다 (SHALL).

coverage 판정은 시장 단위 readiness 판정과 **독립적으로 AND** 되어야 하며 (SHALL), 시장 단위
readiness 요청에 포지션 식별자를 추가하거나 두 판정을 하나의 상태로 합쳐서는 안 된다
(MUST NOT). 한 포지션의 coverage 부족이 같은 시장 다른 포지션의 판정이나 reduce-only 청산·대사를
바꾸어서는 안 된다 (MUST NOT).

#### Scenario: 부분 보호 상태
- **WHEN** 보유 수량 100 중 60만 브로커 보호주문으로 덮여 있다
- **THEN** 그 포지션의 신규 진입은 닫히고 reduce-only 청산과 대사는 계속되며 다른 포지션은 영향을 받지 않는다

#### Scenario: 두 latch의 독립성
- **WHEN** 시장 readiness는 WIRED지만 한 포지션의 coverage가 미달이다
- **THEN** 그 포지션만 진입이 닫히고 같은 시장의 완전 보호된 포지션은 열린 상태를 유지한다

#### Scenario: coverage 충족 뒤에도 lane은 OFF다
- **WHEN** 보호 수량과 다른 매도 청구권의 합이 보유 수량과 정확히 일치한다
- **THEN** 그 포지션의 coverage latch는 열리지만 운영자가 승인하지 않은 lane은 여전히 OFF이고 자동 진입은 발생하지 않는다
