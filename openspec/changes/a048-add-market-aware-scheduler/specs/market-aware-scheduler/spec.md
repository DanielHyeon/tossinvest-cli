## ADDED Requirements

### Requirement: scheduler는 시장별 entry 가능 상태를 결정한다
시스템은 exchange calendar, IANA timezone, session, holiday, early close와 DST를 사용해 `ENTRY_ALLOWED`, `WAIT_MARKET`, `DISABLED`, `BUDGET_DEFERRED` 중 하나를 반환해야 한다 (SHALL).
calendar는 official typed adapter의 canonical response digest와 fetched-at을 가져야 하며 (SHALL), 6시간 이상 stale, session 전 refresh 실패, missing 또는 parse failure이면 신규 entry는 `WAIT_MARKET`여야 한다 (SHALL).

#### Scenario: 미국 DST 전환
- **WHEN** 미국 정규장 시각이 DST 경계를 지난다
- **THEN** local machine timezone과 무관하게 exchange session 기준으로 entry window를 계산한다

#### Scenario: 휴장일
- **WHEN** 대상 시장이 휴장이다
- **THEN** 신규 entry/candidate entry cadence를 대기시키되 reconcile·exit·filldetect는 계속된다

### Requirement: desired state 복원은 기존 사람 승인을 넘지 않는다
재시작은 저장된 actor, approval time, market scope와 config version이 유효할 때만 entry desired state를 복원해야 하며 새로운 LIVE 승인을 만들어서는 안 된다 (MUST NOT).
복원은 a047 activation manifest의 scheduler, calendar, market/session, config/build digest와 expiry가 모두 현재 값과 일치할 때만 가능해야 한다 (SHALL).

#### Scenario: 승인된 자동 복원
- **WHEN** 운영자가 이전에 auto-resume을 승인했고 모든 gate가 여전히 유효하다
- **THEN** 해당 market lane만 복원한다

#### Scenario: 설정 version 변경
- **WHEN** 승인 뒤 high-risk config version이 변경됐다
- **THEN** entry는 재승인을 요구하고 자동 복원하지 않는다

#### Scenario: manifest 만료
- **WHEN** desired state가 ON이어도 activation manifest가 만료됐거나 calendar digest가 바뀌었다
- **THEN** auto-resume은 거부되고 신규 진입만 OFF를 유지한다

### Requirement: polling은 API 예산을 침범하지 않는다
scheduler는 entry/candidate 호출이 exit, fill detection과 reconciliation의 예약 예산을 소비하지 않도록 cadence를 제한해야 한다 (SHALL).
endpoint budget-key별 safety reserve는 reported remaining의 50%를 올림한 값과 5 calls 중 큰 값이어야 하며 (SHALL), budget provenance가 missing/stale이면 entry/candidate/analytics 추가 poll을 수행해서는 안 된다 (MUST NOT).

#### Scenario: 예산 압력
- **WHEN** reserved safety budget을 제외한 호출 여유가 없다
- **THEN** 신규 entry polling을 BUDGET_DEFERRED로 미루고 안전 루프를 유지한다

#### Scenario: budget header 부재
- **WHEN** endpoint의 remaining/reset provenance가 없거나 stale하다
- **THEN** 신규 entry/candidate/analytics poll은 0건이고 emergency exit, reconcile, fill detection과 protection supervision은 기존 cadence를 유지한다
