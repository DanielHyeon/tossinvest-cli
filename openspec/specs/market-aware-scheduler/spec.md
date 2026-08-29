# market-aware-scheduler Specification

## Purpose
TBD - created by archiving change a048-add-market-aware-scheduler. Update Purpose after archive.
## Requirements
### Requirement: scheduler는 시장별 entry 가능 상태를 결정한다
시스템은 exchange calendar, IANA timezone, session, holiday, early close와 DST를 사용해 `ENTRY_ALLOWED`, `WAIT_MARKET`, `DISABLED`, `BUDGET_DEFERRED` 중 하나를 반환해야 한다 (SHALL).
calendar는 official typed adapter의 canonical response digest와 응답 완료 시각인 fetched-at을 가져야 하며 (SHALL), 6시간 이상 stale, session 전 refresh 실패, missing 또는 parse failure이면 신규 entry는 `WAIT_MARKET`여야 한다 (SHALL).
market scope는 단일 calendar/activation binding으로 증명 가능한 `none`, `KR`, `US`만 허용해야 하며 (SHALL), per-market binding이 없는 결합 scope를 저장하거나 선택지로 광고해서는 안 된다 (MUST NOT).

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
허용된 low-priority poll은 암호학적 난수로 발급되고 coordinator, endpoint key, poll class와 reset generation에 결합된 불투명 capability commitment로 계산해야 하며 (SHALL), 예측·위조·다른 key/class/coordinator/generation의 capability가 commitment를 완료하거나 해제해서는 안 된다 (MUST NOT). endpoint/reset generation별 capability 발급 수는 reported limit와 독립된 절대 상한을 가져야 하고 (SHALL), 같은 window reconciliation은 issued capability 기억을 지우거나 재사용 가능하게 만들어서는 안 된다 (MUST NOT). 호출 성공, 오류 또는 취소 뒤 completion은 commitment를 completed/unreconciled로 표시할 뿐 capacity를 복원해서는 안 된다 (MUST NOT). budget request는 시작 전에 coordinator, endpoint key, reset generation과 그 시점의 monotonic completion watermark에 결합된 불투명 one-shot observation cycle을 발급받아야 하며 (SHALL), 같은 window observation은 자신의 cycle이 시작되기 전에 완료된 commitment만 reconcile하고 아직 in-flight이거나 cycle 시작 뒤 완료된 commitment는 보존해야 한다 (SHALL). wall observed-at/completed-at 비교, 처리 순서, 초기/manual observation, 위조·재생·다른 key/coordinator/generation cycle, missing/stale/invalid/conflicting observation은 reconciliation authority가 되어서는 안 된다 (MUST NOT). 신뢰 가능한 다음 reset boundary도 valid nonnil cycle이 모든 기존 commitment를 account한 뒤에만 이전 generation을 지우고 절대 발급 상한과 observation-cycle 상한을 reset해야 하며 (SHALL), commitment가 비어 있어도 manual observation이 generation 또는 어떤 발급 기억도 초기화해서는 안 된다 (MUST NOT). reset raw/kind/instant는 official parser와 동일한 exact `1_000_000_000` epoch threshold, observed-at 기준 inclusive `[-1m,+24h]` plausibility와 overflow-safe delta derivation을 통과해야 하고 raw-kind mismatch는 거부해야 한다 (SHALL). delta reset은 고정 anchor에 대한 bounded tolerance 안의 pre-boundary drift를 같은 window로 처리하되 가장 이른 deadline을 유지하고 duration absolute-value overflow 없이 비교해야 하며 (SHALL), epoch reset은 exact identity를 유지하고 다음 generation은 이전 boundary가 지났다는 양의 증거 뒤에만 만들어야 한다 (SHALL).

#### Scenario: 예산 압력
- **WHEN** reserved safety budget을 제외한 호출 여유가 없다
- **THEN** 신규 entry polling을 BUDGET_DEFERRED로 미루고 안전 루프를 유지한다

#### Scenario: budget header 부재
- **WHEN** endpoint의 remaining/reset provenance가 없거나 stale하다
- **THEN** 신규 entry/candidate/analytics poll은 0건이고 emergency exit, reconcile, fill detection과 protection supervision은 기존 cadence를 유지한다

#### Scenario: 동일 window에 in-flight poll이 있다
- **WHEN** low-priority poll completion 전 같은 reset window의 새 remaining observation이 도착한다
- **THEN** outstanding commitment를 계속 차감하여 safety reserve를 침범하지 않고, 이후 completion만으로 capacity를 복원하지 않으며 그 completion 뒤 시작된 authoritative one-shot request cycle 또는 그 cycle이 account하는 신뢰 가능한 다음 reset boundary에서만 reconcile한다

#### Scenario: completion 응답에 신뢰 가능한 budget이 없다
- **WHEN** low-priority poll이 성공, 오류 또는 취소로 끝났지만 missing, stale, invalid 또는 conflicting budget provenance만 있다
- **THEN** completed commitment를 consumed/unreconciled로 계속 차감하여 반복 acquire-complete가 같은 stale capacity를 재사용하지 못한다

#### Scenario: commitment capability가 scope와 다르다
- **WHEN** capability가 위조됐거나 발급 coordinator, endpoint key, poll class 또는 reset generation과 다르다
- **THEN** 어떤 commitment도 완료하거나 capacity를 복원하지 않는다

#### Scenario: 동일 시각 provenance가 충돌한다
- **WHEN** 같은 observed-at correction의 reported, reset, reset kind, reset raw 또는 limit가 기존 근거와 다르다
- **THEN** strictly newer observation 전까지 entry/candidate/analytics poll을 0건으로 제한한다

#### Scenario: 완료 전 시작된 response가 wall clock rollback 뒤 늦게 도착한다
- **WHEN** budget request cycle이 commitment completion 전에 시작됐고 wall clock이 rollback한 뒤 response의 observed-at만 completion보다 뒤로 보인다
- **THEN** 해당 response는 commitment를 reconcile하지 않으며 completion 뒤 시작된 one-shot cycle의 authoritative response만 reconcile한다

#### Scenario: capability 발급 기억이 절대 상한에 도달한다
- **WHEN** reported limit가 매우 커도 같은 endpoint/reset generation의 capability 발급 수가 절대 상한에 도달한다
- **THEN** entry/candidate/analytics 발급은 fail-closed이고 safety class는 계속되며, same-window reconcile은 상한을 되돌리지 않고 proven reset만 새 generation을 연다

#### Scenario: commitment가 비어 있는 상태의 manual 새 window 관측
- **WHEN** 같은 generation에서 256개 capability가 발급·reconcile되어 commitment map은 비었지만 issued 기억과 observation-cycle 상한이 남아 있고 manual `Observe`가 다음 reset을 보고한다
- **THEN** generation과 두 발급 기억은 그대로이며 low-priority는 fail-closed이고 safety class는 계속되고, 모든 이전 commitment를 account하는 valid nonnil cycle만 새 generation을 연다

#### Scenario: official reset parser와 다른 forged provenance
- **WHEN** reset raw가 delta duration을 wrap하거나 `1_000_000_000` threshold의 kind와 다르거나 observed-at 기준 1분 이전·24시간 이후이거나 raw/kind/derived instant가 일치하지 않는다
- **THEN** 해당 evidence는 reconciliation/generation authority를 얻지 못하고 low-priority poll은 fail-closed이며 safety class는 계속된다

#### Scenario: delta reset derived instant가 response latency로 흔들린다
- **WHEN** 같은 reset window의 유효한 delta header들이 subsecond latency와 정수 seconds 때문에 서로 조금 다른 Reset instant를 만든다
- **THEN** raw/kind/고정 tolerance가 일치하는 pre-boundary observation은 같은 generation으로 reconcile하고 가장 이른 deadline을 유지하며, 이전 boundary가 지난 뒤 새 reset을 관측할 때만 generation을 전환한다

### Requirement: KR과 US scheduler binding은 동시에 독립 진행한다
Runtime scheduler는 KR과 US binding을 동시에 독립 진행해야 한다 (SHALL). KR과 US 각각에 별도 calendar generation, activation manifest, session
decision과 endpoint budget binding을 유지하면서 두 시장을 동시에 진행해야 한다 (SHALL).
한 시장의 closed/stale/refused/budget-deferred state를 다른 시장 decision의 입력으로 사용해서는
안 되며 (MUST NOT), per-market binding을 결합한 하나의 승인 scope를 만들거나 복원해서도 안
된다 (MUST NOT). Safety class budget과 exit/reconciliation cadence는 두 entry binding보다
우선해야 한다 (SHALL).

#### Scenario: 동시 상이 상태
- **WHEN** KR은 current calendar에서 WAIT_MARKET이고 US는 ENTRY_ALLOWED다
- **THEN** 두 decision을 같은 cycle에 독립 반환하고 US candidate cadence를 KR 상태 때문에 지연하지 않는다

#### Scenario: 한 시장 budget provenance 누락
- **WHEN** US entry endpoint budget provenance가 stale이고 KR endpoint budget은 current다
- **THEN** US entry poll만 BUDGET_DEFERRED이며 KR entry poll과 두 시장 safety class 호출은 계속된다

#### Scenario: 재시작 복원 범위
- **WHEN** KR auto-resume manifest만 유효하고 US manifest는 만료됐다
- **THEN** KR scheduler binding만 복원하고 US entry는 OFF로 유지하며 새로운 LIVE approval을 만들지 않는다
