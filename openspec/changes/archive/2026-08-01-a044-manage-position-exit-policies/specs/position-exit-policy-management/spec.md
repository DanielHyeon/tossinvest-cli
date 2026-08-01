## ADDED Requirements

### Requirement: 종목별 정책 변경은 preview와 CAS 승인을 사용한다
시스템은 account·market·symbol·generation 범위의 정책 변경을 preview한 뒤 일치하는 version과 운영자 승인으로만 적용해야 한다 (SHALL).

#### Scenario: 정상 override
- **WHEN** 운영자가 current version을 가진 preview를 승인한다
- **THEN** 새 policy snapshot과 before/after/actor/server-defined reason audit가 한 transaction에 기록된다

#### Scenario: stale version
- **WHEN** 다른 write 뒤 오래된 version을 담은 opaque console preview token 또는 API `If-Match`로 적용한다
- **THEN** 412 version mismatch를 반환하고 정책과 audit effective state를 바꾸지 않는다

### Requirement: release와 re-adopt는 lifecycle을 분리한다
release는 active 보호·미체결 청산과 충돌하지 않을 때만 허용해야 하며 (SHALL), re-adopt는 새 generation과 새 관측 t0를 만들어야 한다 (MUST).

#### Scenario: 안전한 release
- **WHEN** 미체결 exit가 없고 운영자가 현재 version으로 release를 승인한다
- **THEN** 현재 lifecycle을 닫고 이후 자동 exit 관리를 중지한다

#### Scenario: re-adopt
- **WHEN** release된 외부 보유를 다시 편입한다
- **THEN** 과거 high-water/rung을 재사용하지 않고 새 generation으로 시작한다
