## ADDED Requirements

### Requirement: Campaign과 leg는 전략 중립 identity와 순서를 가진다

PositionCampaign은 account, market, symbol, owning lane identity/version, originating decision/evidence, prospective generation token 및 실제 position generation을 명시적으로 연결해야 한다(SHALL). 각 CampaignLeg는 campaign 내 불변 sequence, 계획 identity, 요청 수량, intent/attempt/fill lineage와 상태를 가져야 한다(SHALL). 코어는 특정 lane의 비율, 최대 leg 수 또는 cadence 상수를 포함해서는 안 된다(MUST NOT).

#### Scenario: 서로 다른 scale-in 전략
- **WHEN** 두 lane가 각각 8:4:2와 2:4:8 계획을 CampaignLeg 명령으로 변환한다
- **THEN** core는 동일한 ordered-leg 계약으로 기록하고 두 비율을 도메인 상수로 해석하거나 저장하지 않는다

#### Scenario: 청산 후 재진입
- **WHEN** CLOSED campaign과 position generation의 symbol에 새 진입이 계획된다
- **THEN** 이전 campaign을 재사용하지 않고 새 campaign identity와 새 position generation lineage를 요구한다

### Requirement: Campaign은 prospective position generation을 CAS로 예약한다

첫 fill 전 campaign 생성은 `(account, market, symbol, expected_position_generation, expected_position_version)`을 compare-and-swap하고 유일한 prospective generation token을 같은 journal transaction에서 예약해야 한다(SHALL). stale expectation, active campaign 또는 이미 예약된 prospective generation이 있으면 `GENERATION_CONFLICT`로 거부해야 한다(SHALL). 첫 accepted entry fill은 기존 Position apply transaction 안에서 token을 실제 successor position generation에 set-once 결합해야 하며(SHALL), mismatch를 임의 generation에 귀속해서는 안 된다(MUST NOT).

#### Scenario: 동시 prospective campaign 생성
- **WHEN** 두 lane가 같은 account/market/symbol과 expected generation/version으로 campaign을 동시에 생성한다
- **THEN** 한 CAS만 prospective token을 얻고 다른 생성은 GENERATION_CONFLICT이며 두 번째 campaign은 없다

#### Scenario: 첫 fill의 successor mismatch
- **WHEN** first fill apply가 prospective token이 기대한 successor와 다른 Position generation을 관측한다
- **THEN** campaign은 RECONCILE로 격리되고 fill을 추정 campaign에 귀속하지 않는다

### Requirement: Campaign 명령과 fill 적용은 멱등하고 원자적이다

모든 leg plan, submit-link, cancel 및 fill-application 명령은 deterministic command key를 가져야 하고(SHALL), 같은 key의 retry는 기존 결과를 반환하며 상태·수량·event를 두 번 전진시켜서는 안 된다(MUST NOT). expected campaign version, 해당 broker order identity의 cumulative fill watermark, event append와 projection 갱신은 하나의 journal transaction에서 검증·commit되어야 한다(SHALL). fill watermark는 leg aggregate 하나가 아니라 broker order identity별로 유지되어야 한다(SHALL).

#### Scenario: 부분체결 관측 재전송
- **WHEN** 동일 cumulative fill watermark의 broker observation이 retry 또는 restart 뒤 다시 적용된다
- **THEN** CampaignLeg와 Position 수량은 추가 증가하지 않고 기존 적용 결과가 반환된다

#### Scenario: commit 직전 crash
- **WHEN** leg fill transaction이 commit 전에 process crash를 겪는다
- **THEN** event, campaign projection과 Position apply는 모두 이전 상태이며 재시도가 한 번만 반영한다

#### Scenario: 동시 다음-leg 계획
- **WHEN** 같은 campaign version에 대해 두 worker가 다음 sequence leg를 동시에 계획한다
- **THEN** 하나만 commit되고 다른 명령은 VERSION_CONFLICT 또는 기존 idempotent 결과를 받는다

#### Scenario: Replacement order의 cumulative fill
- **WHEN** replacement order가 predecessor carry baseline과 새 cumulative watermark를 보고한다
- **THEN** predecessor에 이미 반영된 fill은 다시 더하지 않고 새 order identity의 검증된 증가분만 한 번 적용한다

#### Scenario: 서로 다른 order watermark 혼합
- **WHEN** 한 leg의 두 broker order identity가 서로 모순되거나 합계가 leg requested quantity를 초과한다
- **THEN** 이미 관측된 fill과 Position delta는 보존하고 산술로 잘라 맞추지 않으며 campaign/leg를 RECONCILE로 격리해 신규 exposure만 차단한다

### Requirement: terminal predecessor의 late fill은 Position에 exactly once 보존된다

replaced 또는 cancelled predecessor order의 immutable broker-order cumulative watermark가 뒤늦게 증가하면 시스템은 그 positive delta, fill evidence와 authoritative Position delta를 같은 journal transaction에서 exactly once 적용해야 한다(SHALL). 같은 transaction은 successor replacement의 remaining quantity와 leg aggregate filled/residual quantity를 재계산해야 한다(SHALL). late delta가 leg/order cap을 초과하거나 predecessor/replacement lineage가 ambiguous해도 fill 또는 Position apply를 버리거나 truncate/rollback해서는 안 되며(MUST NOT), campaign을 `RECONCILE`로 latch하고 신규 exposure만 차단해야 한다(SHALL). stop, emergency exit, reconciliation과 fill detection은 계속되어야 한다(SHALL).

#### Scenario: cancelled predecessor late positive delta
- **WHEN** partial fill 뒤 cancelled되고 replacement가 연결된 predecessor가 더 높은 cumulative fill watermark를 보고한다
- **THEN** predecessor의 새 delta와 Position이 한 transaction에서 한 번 전진하고 replacement remaining 및 leg aggregate가 재계산되며 campaign은 RECONCILE로 신규 entry를 차단한다

#### Scenario: late fill retry
- **WHEN** 같은 predecessor cumulative watermark가 process restart 또는 observation retry로 다시 도착한다
- **THEN** per-order watermark가 이미 반영된 delta를 0으로 만들고 Position, replacement remaining과 leg aggregate는 추가로 변하지 않는다

#### Scenario: late fill transaction crash
- **WHEN** predecessor watermark는 검증됐지만 Position 또는 replacement projection commit 전에 process가 crash한다
- **THEN** watermark, fill, Position, replacement remaining, leg aggregate와 campaign state가 모두 rollback되고 retry가 전체 delta를 정확히 한 번 적용한다

#### Scenario: late fill cap 초과 또는 lineage ambiguity
- **WHEN** late positive delta가 leg cap을 초과하거나 predecessor와 replacement의 lineage가 상충한다
- **THEN** fill evidence와 authoritative Position delta는 그대로 보존되고 campaign은 RECONCILE/entry-block이며 실제 체결 수량을 cap에 맞춰 자르거나 폐기하지 않는다

### Requirement: Campaign 상태기계는 EXIT FIRST를 강제한다

Campaign은 PLANNED, ACTIVE, EXITING, CLOSED, RECONCILE 상태를, Leg는 PLANNED, SUBMITTED, PARTIAL, FILLED, CANCELLED, RECONCILE 상태를 가져야 한다(SHALL). 허용 전이와 terminal/recovery 처리는 design D4의 완전한 versioned 표와 정확히 일치해야 하며(SHALL), 표에 없는 정상 전이를 허용해서는 안 된다(MUST NOT). EXITING, CLOSED, RECONCILE, Position CLOSING 또는 unresolved risk-reducing intent가 있는 동안 exposure-raising leg plan/submit을 허용해서는 안 된다(MUST NOT). stop, emergency exit, reduce-only fill, reconciliation과 fill detection은 entry 상태나 손실 latch 때문에 지연되어서는 안 된다(MUST NOT).

#### Scenario: scale-in과 stop 동시 발생
- **WHEN** 다음 entry leg와 stop exit가 같은 campaign version을 대상으로 경쟁한다
- **THEN** exit 전이가 우선 commit되고 entry leg는 EXIT_FIRST_BLOCKED로 거부된다

#### Scenario: RECONCILE 중 체결 관측
- **WHEN** campaign이 RECONCILE 상태에서 기존 주문의 fill observation이 도착한다
- **THEN** fill detection과 Position 투영은 계속되고 새로운 exposure-raising leg만 거부된다

#### Scenario: CLOSED 뒤 새로운 fill 사실
- **WHEN** CLOSED campaign에 idempotent retry가 아닌 새 positive fill 또는 order 사실이 도착한다
- **THEN** campaign을 reopen하지 않고 event를 격리하며 account reconciliation과 신규 entry block을 발동한다

### Requirement: Effective stop은 불리한 방향으로 후퇴하지 않는다

long-only campaign의 effective stop은 이미 저장된 유효 stop보다 낮아지거나 NULL이 되어서는 안 된다(MUST NOT). 새 candidate는 유효성 검사를 통과한 경우에만 `max(saved_effective_stop, candidate_stop)`으로 합성되어야 하고(SHALL), source, policy/version, observed-at과 선택 provenance를 보존해야 한다(SHALL).

#### Scenario: 더 낮은 새 stop
- **WHEN** 저장된 effective stop보다 낮은 유효 candidate가 제안된다
- **THEN** 저장 stop이 유지되고 candidate와 거부 provenance가 기록된다

#### Scenario: stop evidence 누락
- **WHEN** 새 leg 평가에서 stop evidence가 missing 또는 invalid다
- **THEN** 기존 effective stop은 변경되지 않고 새 exposure-raising leg는 fail closed된다

### Requirement: Offline reconstruction은 주문 없이 campaign을 재현한다

시스템은 append-only journal evidence만 사용해 prospective-generation CAS/binding, campaign 상태, ordered legs, command 결과, broker order identity별 cumulative fill watermarks와 replacement lineage, position generation lineage와 effective stop을 결정적으로 재구성해야 한다(SHALL). reconstruction은 snapshot과의 불일치를 stable reason과 마지막 valid event로 보고해야 하며(SHALL), broker 호출, 주문 intent 생성, 상태 자동 보정 또는 운영 토글 변경을 수행해서는 안 된다(MUST NOT).

#### Scenario: 재시작 전후 동일 replay
- **WHEN** 같은 ordered journal events를 빈 projection에 replay한다
- **THEN** process restart 전 snapshot과 동일한 campaign/leg 상태와 digest가 생성된다

#### Scenario: 순서가 끊긴 leg event
- **WHEN** replay 중 campaign-local leg sequence gap이 발견된다
- **THEN** reconstruction은 LEG_SEQUENCE_GAP과 마지막 valid event를 보고하고 누락 leg를 추정하지 않는다
