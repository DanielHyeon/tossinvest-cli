## ADDED Requirements

### Requirement: KR과 US evaluation은 독립 scope로 동시에 감독된다
Production strategy runtime은 KR과 US evaluation을 독립 scope로 동시에 감독해야 한다 (SHALL). KR과 US 각각에 독립된 calendar, activation manifest, evidence
cursor, budget key와 typed runtime state를 가진 evaluation worker를 동시에 실행해야 한다
(SHALL). 한 시장의 `WAIT_MARKET`, `DISABLED`, `BUDGET_DEFERRED`, stale evidence 또는
cycle failure가 다른 시장의 eligible evaluation을 중단해서는 안 된다 (MUST NOT). 결합
KR+US calendar 또는 activation scope를 runtime authority로 사용해서는 안 된다 (MUST NOT).
Worker의 panic, abnormal return, watchdog expiry 또는 반복 crash는 해당 market의 effective
entry만 durable OFF로 latch하고 bounded restart해야 하며 (SHALL), peer worker와 모든 safety-class
loop를 취소해서는 안 된다 (MUST NOT).

#### Scenario: KR 휴장과 US 정규장
- **WHEN** KR worker는 holiday로 WAIT_MARKET이고 US worker는 current calendar와 activation으로 ENTRY_ALLOWED다
- **THEN** KR 신규 평가만 대기하고 US evaluation은 같은 runtime에서 계속되며 두 시장의 safety loop도 계속된다

#### Scenario: US evidence 장애
- **WHEN** US evidence adapter가 한 cycle에서 실패하고 KR evidence는 current다
- **THEN** US는 typed refusal과 bounded retry를 기록하고 KR candidate evaluation과 dispatch eligibility를 중단하지 않는다

#### Scenario: 시장별 activation
- **WHEN** KR activation은 유효하고 US activation은 OFF다
- **THEN** KR worker만 EntryDecision을 만들 수 있고 US 신규 entry/scale-in은 0건이며 US exit와 reconciliation은 계속된다

#### Scenario: KR worker 비정상 반환
- **WHEN** KR worker가 비정상 반환하지만 US worker와 safety-class loop는 정상이다
- **THEN** KR entry만 effective OFF로 latch하고 bounded restart하며 US evaluation과 두 시장 fill/reconciliation/protection/exit를 계속한다

### Requirement: supervised entry path는 완전한 lineage를 영속한다
Runtime은 supervised entry path의 완전한 lineage를 영속해야 한다 (SHALL). ApprovedCandidate를 immutable evidence snapshot, market scheduler, horizon router,
selected lane, campaign/leg, risk sizing, Guardian, strategy dispatch와 official
ExecutionGateway 순서로만 연결해야 한다 (SHALL). candidate/evidence, router, lane/version,
campaign/leg, risk policy/reservation, Guardian decision, attempt와 order identifier chain을
journal에 영속해야 하며 (SHALL), lane 또는 worker가 broker mutator를 직접 소유해서는 안 된다
(MUST NOT).

#### Scenario: 유효한 KR continuation candidate
- **WHEN** KR ApprovedCandidate가 current evidence와 scheduler를 통과해 continuation lane에서 수락된다
- **THEN** candidate부터 official Gateway attempt까지 모든 identifier가 같은 market lineage로 저장되고 broker mutation은 central dispatch owner만 수행한다

#### Scenario: lane refusal
- **WHEN** selected US lane이 입력을 거부한다
- **THEN** first typed refusal과 evidence/lane version을 저장하고 Guardian, attempt와 broker request는 0건이다

#### Scenario: lineage 누락
- **WHEN** campaign/leg 또는 lane version을 durable lineage에 결합할 수 없다
- **THEN** runtime은 exposure-raising dispatch를 거부하고 symbol/time 추정으로 누락 식별자를 채우지 않는다

### Requirement: exposure-raising dispatch는 fenced 비가역 durable lease를 요구한다
Runtime은 모든 exposure-raising dispatch에 fenced 비가역 durable lease를 요구해야 한다 (SHALL). Exposure-raising attempt마다 account/market/symbol, candidate/evidence digest,
router/lane/version, campaign/leg, activation manifest와 generation, calendar generation,
ProtectionReady attestation/serial, reconciliation generation, risk reservation/policy, Guardian
decision/generation, build digest, monotonic owner epoch/fencing token와 expiry에 결합된 lease를
발급·영속해야 한다 (SHALL). ProtectionReady enum은 정확히 `WIRED`/`UNWIRED`만 소비하고 실패
상세는 typed refusal로 분리해야 한다 (SHALL).

Lease 상태는 `ISSUED`, `CLAIMED`, `SUBMITTING`, `SUBMITTED`, `AMBIGUOUS`, `REFUSED`로 제한해야
한다 (SHALL). Current fenced owner의 atomic claim/validation은 `ISSUED→CLAIMED` 후 모든
authority를 current durable state에서 비교해야 한다 (SHALL). 성공은
`CLAIMED→SUBMITTING→SUBMITTED`, validation/dispatch refusal은 terminal `REFUSED`, definitive
broker rejection 또는 authoritative no-accept/no-fill은 terminal `REFUSED`, durable transport
uncertainty만 terminal `AMBIGUOUS`로 영속해야 한다 (SHALL). 모든 claim 또는 validation
시도는 성공 dispatch 또는 terminal refusal로 lease를 비가역 소비해야 하며
(SHALL), terminal state를 `ISSUED`로 되돌리거나 재사용해서는 안 된다 (MUST NOT).

각 authority의 monotonic generation을 비교해야 하므로 값이 A→B→A로 복귀해도 이전 lease를
유효하게 만들어서는 안 된다 (MUST NOT). Owner 재기동은 durable epoch를 증가시키고 이전
fencing token의 journal transition과 Gateway call을 거부해야 한다 (SHALL).

#### Scenario: decision 뒤 protection drift
- **WHEN** EntryDecision 뒤 dispatch 전에 해당 시장 ProtectionReady generation이 바뀐다
- **THEN** lease와 exact reservation을 같은 transaction에서 `REFUSED + RELEASED`로 만들고 broker request는 0건이며 typed protection refusal을 기록한다

#### Scenario: lease 재사용
- **WHEN** 이미 소비됐거나 만료된 dispatch lease로 두 번째 exposure-raising attempt를 요청한다
- **THEN** replay attempt는 typed `REFUSED`, retry attempt가 별도로 만든 exact HELD reservation만 원자 `RELEASED`, 원래 terminal lease/disposition과 broker request는 불변이며 새로운 current decision/lease를 요구한다

#### Scenario: Guardian과 risk generation 불일치
- **WHEN** lease의 risk reservation 또는 Guardian generation이 current durable state와 다르다
- **THEN** lease와 exact reservation을 원자 `REFUSED + RELEASED`로 소비하고 broker request 없이 stale authority를 다시 계산해 같은 lease를 자동 승인하지 않는다

#### Scenario: authority A에서 B를 거쳐 A로 복귀
- **WHEN** lease 발급 뒤 activation 또는 Guardian authority가 A→B→A로 값은 복귀했지만 generation은 증가했다
- **THEN** 이전 lease는 terminal `REFUSED`이고 fresh decision/lease 없이는 제출할 수 없다

#### Scenario: CLAIMED 뒤 pre-transport crash
- **WHEN** lease가 CLAIMED지만 durable transport-start marker 전에 process가 crash한다
- **THEN** recovery는 broker request가 없음을 증명해 lease와 exact reservation을 원자 `REFUSED + RELEASED`로 확정하고 `AMBIGUOUS`를 사용하지 않는다

#### Scenario: stale owner fencing
- **WHEN** 이전 owner epoch/token을 가진 process가 새 owner 선출 뒤 lease transition 또는 Gateway call을 시도한다
- **THEN** journal compare-and-set과 Gateway 경계가 broker transport 전에 거부하고 current owner만 진행한다

### Requirement: broker unknown outcome은 attested idempotency 없이는 재제출하지 않는다
Runtime은 broker unknown outcome을 attested idempotency 없이는 재제출하지 않아야 한다 (SHALL). `SUBMITTING` 뒤 timeout, connection loss, crash 또는 malformed response는
exact broker outcome을 먼저 분류해야 한다 (SHALL). Definitive broker rejection 또는 attested
exact query가 증명한 authoritative no-accept/no-fill은 broker outcome evidence와 함께 terminal
`REFUSED + RELEASED`, exact acceptance는 `SUBMITTED + TRANSFERRED`, 어느 쪽도 durable하게
증명할 수 없는 transport uncertainty만 `AMBIGUOUS + HELD`로 같은 transaction에서 영속해야
한다 (SHALL). 동일 operation key 재제출은 current a071 attestation이
client key 전달/echo, exact broker lookup, uniqueness scope, pending/terminal query와
duplicate-submit idempotency를 모두 증명할 때만 허용해야 한다 (SHALL). Capability가 없거나
불완전하면 exact reconciliation만 수행하고 자동 resubmit, 새 lease 생성과 symbol/time dedup을
수행해서는 안 된다 (MUST NOT).

#### Scenario: attested idempotent recovery
- **WHEN** `SUBMITTING` 뒤 응답이 유실됐지만 current market/order attestation이 exact lookup과 same-key idempotency를 모두 증명한다
- **THEN** reconciliation은 exact operation key로 조회하고 attested bounded same-key recovery만 수행하며 새 lease나 다른 identity를 만들지 않는다

#### Scenario: unattested idempotency
- **WHEN** `SUBMITTING` outcome이 unknown이고 broker attestation에 exact dedup 또는 idempotency가 없다
- **THEN** attempt는 terminal `AMBIGUOUS`, reservation은 `HELD`, 자동 resubmit은 0건이며 entry는 exact reconciliation 완료 전까지 차단된다

#### Scenario: definitive broker rejection
- **WHEN** broker가 operation identity에 결합된 definitive rejection과 no-accept/no-fill을 반환한다
- **THEN** broker outcome/evidence와 lease `REFUSED`, reservation `RELEASED`가 한 transaction에 기록되고 `AMBIGUOUS` 또는 retry는 0건이다

#### Scenario: SUBMITTING crash 뒤 authoritative not-submitted
- **WHEN** restart recovery의 attested exact query가 operation이 accepted되지 않았고 fill도 없음을 authoritative하게 증명한다
- **THEN** exact query evidence와 `REFUSED + RELEASED`를 원자 기록하고 같은 lease를 재제출하지 않는다

### Requirement: terminal lease의 reservation disposition은 별도로 대사된다
Runtime은 terminal lease의 risk/campaign reservation disposition을 별도 durable record로 대사해야 한다 (SHALL). Broker transport 전 refusal, definitive broker rejection과 post-transport authoritative no-accept/no-fill은 lease `REFUSED` 및 exact reservation `RELEASED`를 broker outcome evidence와 같은 journal transaction에서 commit해야 하고 (SHALL), `SUBMITTED`는 reservation을 attempt/order의
fill/cancel lifecycle로 `TRANSFERRED`해야 한다 (SHALL). `AMBIGUOUS`는 reservation을 `HELD`로
동결해 release, reuse와 같은 capacity의 fresh lease를 금지해야 한다 (SHALL).

Exact broker reconciliation이 authoritative `NOT_SUBMITTED`를 증명한 때만 별도 outcome으로
`HELD→RELEASED`, acceptance를 증명한 때만 `HELD→TRANSFERRED`를 기록해야 한다 (SHALL).
Reservation disposition 변경은 terminal lease state를 되돌리거나 lease를 다시 제출 가능하게
해서는 안 된다 (MUST NOT).

#### Scenario: broker 전 refusal
- **WHEN** current authority validation이 broker transport 전에 실패한다
- **THEN** lease는 terminal `REFUSED`이고 같은 transaction에서 exact risk/campaign reservation만 `RELEASED`된다

#### Scenario: broker 뒤 definitive no-accept
- **WHEN** exact broker outcome이 요청을 거부했고 accepted order와 fill이 없음을 증명한다
- **THEN** outcome identity/digest와 terminal `REFUSED`, exact reservation `RELEASED`를 원자 commit하고 AMBIGUOUS/HELD를 남기지 않는다

#### Scenario: ambiguous reservation freeze
- **WHEN** broker transport 뒤 outcome이 unknown이다
- **THEN** lease는 terminal `AMBIGUOUS`, reservation은 `HELD`이며 release/reuse와 같은 capacity의 new lease가 모두 차단된다

#### Scenario: authoritative not-submitted reconciliation
- **WHEN** exact broker identity/query가 해당 ambiguous operation이 제출되지 않았음을 authoritative하게 증명한다
- **THEN** 별도 reconciliation outcome이 reservation을 `RELEASED`로 바꾸지만 lease는 terminal `AMBIGUOUS`로 유지된다

#### Scenario: authoritative acceptance reconciliation
- **WHEN** exact broker identity/query가 ambiguous operation acceptance를 증명한다
- **THEN** reservation을 attempt/order lifecycle로 `TRANSFERRED`하고 lease는 terminal `AMBIGUOUS`인 채 새 order link만 별도 기록한다

### Requirement: entry OFF는 safety lifecycle을 중단하지 않는다
Runtime은 entry OFF 중에도 safety lifecycle을 계속해야 한다 (SHALL). Lane, market scheduler, autostart 또는 automation effective state가 OFF가 되면 runtime은 해당
scope의 신규 entry와 scale-in을 즉시 0건으로 만들어야 한다 (SHALL). Fill detection,
reconciliation, protection supervision, stop/exit observation과 emergency risk reduction은 OFF,
market close, stale entry evidence 또는 entry budget 부족과 무관하게 계속되어야 한다 (SHALL).
OFF 전환은 운영자 LIVE approval을 만들거나 삭제해서는 안 된다 (MUST NOT).

#### Scenario: lane OFF 전환
- **WHEN** open campaign이 있는 시장의 lane effective state가 OFF로 바뀐다
- **THEN** 신규 leg와 scale-in은 중단되지만 미체결/체결 해소, broker protection과 exit supervision은 계속된다

#### Scenario: entry budget 고갈
- **WHEN** low-priority entry budget은 없지만 safety budget은 유효하다
- **THEN** 신규 evaluation은 BUDGET_DEFERRED이고 fill, reconciliation, protection과 exit 호출은 기존 cadence를 유지한다

#### Scenario: kill switch
- **WHEN** kill switch가 entry를 차단한다
- **THEN** 두 시장의 신규 exposure는 즉시 0건이 되고 existing protection과 reduce-only exit는 계속된다
