## ADDED Requirements

### Requirement: KR와 US scheduler record는 독립 revision과 activation을 가진다
Scheduler는 KR와 US 각각에 독립된 durable desired-state record, monotonic revision, lock identity와 market-scoped activation binding을 유지해야 한다 (SHALL).
각 record는 market, desired state, calendar generation/digest, IANA timezone/session scope, activation
manifest digest/expiry, config version과 updated actor/time을 포함해야 하고 (SHALL), compare-and-swap은
한 시장의 expected revision만 검사·증가시켜야 한다 (SHALL). 한 시장의 CAS conflict, lock, rollback,
disabled, stale 또는 unapproved 상태가 peer market record/revision/activation을 변경하거나 차단해서는
안 된다 (MUST NOT). Combined KR+US record, lock 또는 activation binding을 저장·복원해서는 안 된다
(MUST NOT).

#### Scenario: KR와 US 동시 CAS
- **WHEN** 서로 다른 worker가 같은 시각에 current KR revision과 current US revision을 갱신한다
- **THEN** 두 transaction은 각 market record에서 독립 commit될 수 있고 revision/lock이 서로 충돌하지 않는다

#### Scenario: 한 시장 stale revision
- **WHEN** KR update는 stale expected revision이고 US update는 current revision이다
- **THEN** KR만 VERSION_CONFLICT로 rollback되고 US record/activation은 commit될 수 있다

#### Scenario: crash 전후 durable replay
- **WHEN** market record와 activation binding transaction의 commit 전후에 crash하고 재시작한다
- **THEN** 해당 market은 old 또는 new complete revision 중 하나로 복원되고 partial binding이나 peer market rollback은 없다

### Requirement: legacy scheduler state는 권한을 넓히지 않고 시장별 record로 이관된다
Scheduler migration은 legacy state를 fail-closed market records로 결정적으로 이관해야 한다 (SHALL).
Legacy disabled/unselected state는 KR OFF와 US OFF로 이관해야 하며 (SHALL), current verified
single-market state는 그 exact market의 desired state, calendar/activation evidence와 revision만
이관하고 peer market을 OFF/not-activated로 만들어야 한다 (SHALL). Unknown, combined, corrupt 또는
unverified legacy state는 두 시장 OFF와 typed migration refusal이어야 하고 (SHALL), migration이
새 activation, LIVE approval 또는 auto-resume 권한을 합성해서는 안 된다 (MUST NOT).

#### Scenario: legacy disabled
- **WHEN** legacy scheduler가 disabled이고 selected market이 없다
- **THEN** migration 뒤 KR와 US record가 모두 OFF/not-activated이고 entry는 0건이다

#### Scenario: verified legacy US state
- **WHEN** legacy state가 current signed/verified US calendar와 activation binding만 가진다
- **THEN** US record만 exact evidence로 이관되고 KR은 OFF/not-activated이며 US 권한으로 KR을 켜지 않는다

#### Scenario: migration 중 crash/retry
- **WHEN** legacy-to-market migration commit 전후 crash 뒤 같은 migration을 retry한다
- **THEN** 하나의 migration version과 동일한 KR/US records로 멱등 수렴하고 activation/revision을 중복 발급하지 않는다

### Requirement: KR와 US session decision은 결합 gate 없이 독립 평가된다
Scheduler는 KR와 US session decision을 독립 current record에서 계산해야 한다 (SHALL).
한 시장의 closed, stale, failed 또는 unapproved state가 다른 시장의 `ENTRY_ALLOWED`를 바꾸어서는
안 되고 (MUST NOT), safety class budget과 exit/reconciliation cadence는 두 entry binding보다
우선해야 한다 (SHALL).

#### Scenario: 두 시장 상태가 다르다
- **WHEN** KR은 official calendar에서 WAIT_MARKET이고 US는 valid record로 regular session에 있다
- **THEN** KR entry cadence만 대기하고 US는 ENTRY_ALLOWED가 될 수 있으며 두 시장 safety cadence는 계속된다

#### Scenario: 한 시장 activation만 승인된다
- **WHEN** US market binding만 사람 승인을 가지고 KR binding은 OFF다
- **THEN** US 승인으로 KR을 켜지 않고 KR OFF로 US를 끄지 않으며 각 state/revision을 독립 반환한다

### Requirement: market과 horizon capability는 공유 endpoint quota의 admission subscope다
Scheduler는 market/horizon low-priority capability를 anti-replay admission subscope로 발급하되 physical endpoint/reset-generation quota authority를 하나만 유지해야 한다 (SHALL).
모든 KR/US와 short/weekly subscope는 같은 endpoint/reset generation의 authoritative reported
remaining, safety reserve, outstanding/completed commitments, absolute issuance count와 observation
cycle cap을 공유해야 하며 (SHALL), subscope 생성, retry, CAS 또는 generation alias로 capacity를
복제하거나 quota를 곱해서는 안 된다 (MUST NOT). Capability는 endpoint, reset generation, market,
horizon, poll class와 coordinator에 결합되어 다른 subscope에서 재사용될 수 없고 (MUST NOT), 어느
entry cadence도 exit, fill detection, reconciliation과 protection supervision reserve를 소비해서는
안 된다 (MUST NOT).

#### Scenario: shared endpoint exhaustion
- **WHEN** KR short가 shared endpoint/reset-generation의 low-priority allowance를 모두 commitment로 보유한다
- **THEN** US short와 KR/US weekly도 같은 physical quota에서 BUDGET_DEFERRED이고 새 subscope가 capacity를 만들지 않으며 safety reserve는 유지된다

#### Scenario: 서로 다른 endpoint 독립성
- **WHEN** endpoint A의 shared quota가 고갈되고 endpoint B에는 authoritative allowance가 남는다
- **THEN** endpoint A의 모든 market/horizon subscope만 deferred되고 endpoint B admission은 계속될 수 있다

#### Scenario: horizon capability replay
- **WHEN** KR short capability를 US weekly 또는 같은 scope의 다른 reset generation에서 완료하려 한다
- **THEN** scope mismatch/replay를 거부하고 shared commitment와 capacity는 변경되지 않는다

#### Scenario: concurrent subscope acquire
- **WHEN** KR/US short/weekly worker가 shared remaining의 마지막 한 slot을 동시에 acquire한다
- **THEN** physical endpoint authority에서 하나만 commit되고 나머지는 deferred되며 총 issued commitment는 shared cap을 넘지 않는다
