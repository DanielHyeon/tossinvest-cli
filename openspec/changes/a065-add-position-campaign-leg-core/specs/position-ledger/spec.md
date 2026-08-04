## ADDED Requirements

### Requirement: Campaign lineage는 Position 단일 권위를 확장한다

엔진이 campaign을 통해 만든 Position 변화는 `PROSPECTIVE_GENERATION → CAMPAIGN → LEG → ORDER_WATERMARK → DECISION → INTENT → MUTATION_ATTEMPT → FILL → POSITION_GENERATION`의 명시적 journal 참조로 연결되어야 한다(SHALL). Campaign과 Leg는 계획·상태 및 broker order identity별 누적 fill watermark를 소유하지만 Position 수량, 평균단가와 체결 적용 권위를 대체하거나 직접 변경해서는 안 된다(MUST NOT). prospective token 결합, campaign projection과 Position 투영은 기존 tx-scoped first/follow-up fill apply 지점에서 함께 전진하거나 함께 rollback되어야 한다(SHALL).

replaced/cancelled predecessor의 late positive fill은 campaign terminal 상태, leg cap 초과 또는 ambiguous replacement lineage를 이유로 거부되어서는 안 된다(MUST NOT). immutable broker-order watermark와 Position delta를 same transaction에서 exactly once 보존하고 successor remaining/leg aggregate 및 campaign `RECONCILE` projection을 함께 commit하거나 함께 rollback해야 한다(SHALL).

#### Scenario: Campaign provenance 질의
- **WHEN** campaign으로 형성된 OPEN Position의 provenance를 질의한다
- **THEN** 각 leg의 계획, 결정, 주문 시도, fill과 Position delta가 시간창 추정 없이 명시적 참조로 반환된다

#### Scenario: Campaign projection hook 실패
- **WHEN** Position fill apply transaction 안에서 campaign projection 갱신이 실패한다
- **THEN** fill snapshot, Position delta와 campaign watermark가 모두 commit되지 않는다

#### Scenario: replacement predecessor의 late fill
- **WHEN** replacement가 제출된 뒤 cancelled predecessor에서 새로운 authoritative positive fill delta가 도착한다
- **THEN** Position과 predecessor watermark는 한 번 전진하고 replacement remaining/leg aggregate가 재계산되며 ambiguous 또는 cap 초과이면 fill을 보존한 채 campaign 신규 entry만 차단한다

### Requirement: Campaign persistence migration은 additive이고 legacy를 추정하지 않는다

Campaign/Leg event, prospective-generation token, per-order watermark, projection과 lineage column은 additive-nullable migration으로 추가되어야 하며(SHALL), 기존 Position row에 임의 campaign identity, generation token, lane owner, leg sequence 또는 stop을 backfill해서는 안 된다(MUST NOT). 더 새로운 schema를 모르는 바이너리는 ErrSchemaTooNew로 쓰기 전에 거부해야 한다(SHALL).

#### Scenario: Legacy Position 조회
- **WHEN** campaign migration 전에 생성된 Position을 새 read model로 조회한다
- **THEN** campaign lineage는 명시적 legacy-unknown으로 반환되고 합성 campaign이나 leg는 생성되지 않는다

#### Scenario: 구버전 바이너리 기동
- **WHEN** campaign migration이 적용된 journal을 이전 schema 바이너리가 연다
- **THEN** ErrSchemaTooNew로 종료되고 journal mutation은 0건이다
