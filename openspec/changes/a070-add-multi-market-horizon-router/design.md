## Context

a067–a069 lane 사이에 account/market/symbol generation ownership과 cadence를 조정할 router가 없다.
기존 scheduler는 dormant `DISABLED/OFF`, market 미선택이며 runtime은 `UNOBSERVED`다. 이 change는
a066 단일-owner invariant를 보존하고 market records와 budget admission을 확장하지만 market 선택,
activation 또는 토글을 만들지 않는다.

## Goals / Non-Goals

**Goals:**

- 모든 horizon에서 하나의 generation owner를 결정하고 competing horizon 우회를 차단한다.
- KR/US durable scheduler record, CAS/lock/activation과 legacy migration을 독립화한다.
- market/horizon anti-replay scope를 유지하면서 physical endpoint quota를 공유한다.
- OFF/failure를 해당 market에 격리하고 common safety cadence를 보존한다.

**Non-Goals:**

- lane scoring, campaign/owner persistence, broker dispatch 또는 activation 생성
- combined KR+US record/lock/approval
- physical endpoint quota, safety reserve 또는 reset authority의 복제/변경

## Decisions

### 1. ownership key에는 horizon이 없다

Owner key는 `(account, market, canonical symbol, position_generation)`이다. Horizon은 candidate
attribution과 a066 admission에만 포함한다. Router는 모든 horizon active owner를 먼저 읽고 하나가
있으면 보존한다. `(market,symbol,horizon)` 대안은 short와 weekly가 같은 generation을 동시에
소유하게 하므로 배제한다.

### 2. ambiguity와 stale owner snapshot은 refusal이다

Multiple owner, scope conflict, unresolved tie와 owner snapshot drift는 typed refusal이다. Router는
pure read/decision boundary이고 a066 transaction이 owner acquisition을 소유한다. Registry order나
score로 existing owner를 덮는 대안은 재시작 attribution을 깨므로 제외한다.

### 3. market별 durable record와 CAS를 사용한다

KR/US 각각 desired state, monotonic revision, lock identity, calendar와 activation binding을 하나의
transactional record로 둔다. CAS/rollback은 그 market만 건드린다. Legacy disabled는 both OFF,
verified single market은 exact market만 migrate하고 peer OFF, unknown/corrupt는 both OFF다.
Global desired record는 독립 rollback과 권한 provenance를 잃으므로 배제한다.

### 4. market/horizon은 quota owner가 아니라 admission subscope다

Capability는 endpoint/reset generation의 하나인 reported remaining, reserve, commitment set,
absolute count와 observation-cycle authority에서 발급된다. Market/horizon은 anti-replay scope에
포함되지만 별도 counters를 만들지 않는다. Per-scope quota 복제 대안은 네 scope가 동일 endpoint
quota를 최대 네 배로 쓰게 하므로 배제한다.

### 5. router와 scheduler는 mutation authority를 갖지 않는다

Router는 decision/refusal만 반환하고 scheduler state writer도 order/campaign/toggle writer를
주입받지 않는다. OFF는 entry routing만 막고 common exit/fill/reconcile/protection cadence는 유지한다.

## Risks / Trade-offs

- [모든 horizon owner lookup 비용] → ownership key index와 bounded single-row invariant를 사용하고 multiple result는 fail closed한다.
- [legacy migration이 기존 선택을 잃음] → 검증된 exact single-market evidence만 보존하고 불명확 권한보다 OFF를 우선한다.
- [shared quota 때문에 한 scope가 다른 scope를 소진] → 이는 실제 provider quota를 반영하며 fairness는 bounded admission policy로 해결하고 quota를 복제하지 않는다.
- [market CAS 충돌이 운영자에게 복잡함] → market/revision별 typed conflict와 rollback provenance를 노출한다.

## Migration Plan

1. Legacy fixture를 disabled, verified KR, verified US, corrupt/combined로 분류하고 migration RED tests를 만든다.
2. Market records/CAS/locks와 shared endpoint quota authority를 dormant 상태로 추가한다.
3. Cross-horizon owner, concurrent CAS/acquire, crash/restart와 quota exhaustion tests를 통과시킨다.
4. a072 전에는 routing decision을 dispatch에 연결하지 않는다. Rollback은 새 routing wiring을 제거하되 migrated authority를 넓히거나 peer market을 켜지 않는다.

## Open Questions

없음. Live activation manifest와 dispatch lease는 a072가 소유한다.
