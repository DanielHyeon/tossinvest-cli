## ADDED Requirements

### Requirement: private API는 콘솔과 같은 market runtime projection을 제공한다
Private read API는 콘솔과 같은 market runtime projection을 제공해야 한다 (SHALL). API와 SSE는
console과 동일한 shared projection에서 KR과 US 각각의 lane desired/effective, evidence
freshness/digest, campaign/leg, horizon risk bucket, scheduler/calendar, activation,
ProtectionReady, reconciliation health, first typed refusal와 observed-at을 stable versioned
schema로 반환해야 한다 (SHALL). Market별 status/error envelope를 사용해 한 시장 unavailable이
다른 시장 snapshot을 제거해서는 안 된다 (MUST NOT). ProtectionReady는 정확히
`WIRED`/`UNWIRED`만 사용하고 failure/unavailable을 typed refusal로 분리해야 하며 (SHALL),
default/zero/제3 readiness enum 또는 LIVE/gate/lane activation/autostart/order/protection mutation
route를 추가해서는 안 된다 (MUST NOT).

#### Scenario: 부분 market failure
- **WHEN** KR runtime projection은 current이고 US projection read가 실패한다
- **THEN** API는 KR snapshot과 US typed unavailable plus `UNWIRED` ProtectionReady를 반환하고 0값 성공이나 `UNKNOWN` readiness로 만들지 않는다

#### Scenario: 웹·API drift 검사
- **WHEN** KR과 US runtime fixture를 console adapter와 API adapter로 각각 렌더한다
- **THEN** market key, desired/effective, exact readiness, first refusal, provenance와 unavailable 의미가 동일하다

#### Scenario: mutation surface 부재
- **WHEN** private API route table과 OpenAPI schema를 검사한다
- **THEN** multi-market runtime resource는 read/SSE만 존재하고 LIVE/gate/lane activation/order/protection mutation endpoint는 0건이다

### Requirement: dormant deployment health는 activation 없이 검증된다
Service health는 dormant deployment를 activation 없이 검증해야 한다 (SHALL). Console/API
schema, authenticated runtime-only Unix projection 연결, KR/US dormant state와 config/activation
digest 보존을 read-only로 검증해야 한다 (SHALL). Health check는 engine entry process를
시작하거나 autostart, automation, lane desired, LIVE approval와 protection 설정을 변경해서는
안 되며 (MUST NOT), broker mutation을 수행해서도 안 된다 (MUST NOT).

Replacement 전 service별 current/target exact image digest, rendered Compose digest,
config/activation/protection digest, environment key set, volume/mount identity, current schema와
target/rollback image schema compatibility range 및 baseline health의 immutable preimage를
요구해야 한다 (SHALL). Compatibility gate는 첫 replacement 전에 target read/write와 rollback
post-replace readability를 증명해야 하며 (SHALL), mutable tag, unknown range 또는 불완전
preimage이면 replacement를 수행해서는 안 된다 (MUST NOT). Service는 frozen order로 한 번에
하나씩, service별 최대 5분 health bound 안에서 교체해야 한다 (SHALL).

#### Scenario: OFF 상태 service replace
- **WHEN** Compose services를 새 image로 replace하고 저장된 activation states가 모두 OFF/미승인이다
- **THEN** post-deploy health는 두 market dormant state와 동일한 config digest를 확인하고 어떤 activation이나 order도 만들지 않는다

#### Scenario: activation drift
- **WHEN** Compose render 또는 replacement가 pre-deploy activation/config digest를 바꾸려 한다
- **THEN** deploy verification은 실패하고 변경된 service를 healthy로 판정하지 않는다

#### Scenario: runtime endpoint 미기동
- **WHEN** entry runtime은 OFF라 Unix projection이 아직 생성되지 않았다
- **THEN** API service health는 transport 정상과 typed not-configured를 구분하며 entry-ready로 위장하거나 engine을 자동 시작하지 않는다

#### Scenario: schema compatibility gate 실패
- **WHEN** target image는 current schema를 쓰지만 rollback image가 post-replace schema를 읽는다고 증명할 수 없다
- **THEN** 첫 replacement 전에 배포를 중단하고 running images와 preimage를 그대로 유지한다

#### Scenario: 두 번째 service health 실패
- **WHEN** 첫 service는 교체됐지만 두 번째 service가 frozen health timeout 안에 통과하지 못한다
- **THEN** 이후 service는 건드리지 않고 첫 service만 exact preimage digest로 역순 rollback하며 config/volumes/journal/protection을 변경하지 않는다

#### Scenario: rollback schema 비호환 발견
- **WHEN** replacement 뒤 current schema가 old image compatibility range 밖임을 감지한다
- **THEN** destructive rollback을 금지하고 new service를 유지하며 entry OFF, safety continuity와 typed `ROLLBACK_INCOMPATIBLE` 상태를 보존한다
