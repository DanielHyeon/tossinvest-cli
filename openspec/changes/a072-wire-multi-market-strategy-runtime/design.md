## Context

Candidate discovery, market scheduler, pure lanes, horizon router, risk owner, Guardian,
strategy dispatch와 official ExecutionGateway는 독립된 계약으로 존재하지만 production engine의
하나의 supervised entry path로 조립되지 않았다. 기존 runtime은 safety loops 중심이고 strategy
lane에는 production caller가 없다. 이번 change는 KR 안정화를 US 구현의 선행조건으로 두지 않고
두 시장을 같은 delivery wave에서 독립적으로 평가한다.

2026-08-03 read-only production baseline에서 `/strategy-runtime`은 runtime seam `미배선`,
KR Parker lane 하나만 존재, runtime `UNOBSERVED`, lane/autostart/entry `OFF`, source manifest
`NOT_CONFIGURED`, candidate/scheduler/Guardian/reconciliation `MISSING`, ProtectionReady
`UNWIRED`, activation absent를 반환했다. `/strategy-runtime/market-schedule`은 scheduler
`DISABLED/OFF`, 선택 시장 없음, exchange calendar unverified와 `NOT_ACTIVATED`를 반환했다.
Server choice에 KR과 US가 보여도 current calendar/activation binding을 증명하지 않으므로
운영 권한으로 해석하지 않는다.

기존 single-writer journal/Gateway, official-only mutation, human activation과 OFF-default 경계는
유지한다. Entry 평가 concurrency가 주문 mutation concurrency나 권한 복제를 뜻하지 않는다.

## Goals / Non-Goals

**Goals:**

- KR과 US evaluation을 동시에 진행하되 시장별 calendar, activation, budget과 failure를 격리한다.
- approved candidate에서 official Gateway까지의 lineage와 첫 refusal을 영속한다.
- dispatch 직전 immutable safety evidence를 비가역 durable lease와 fenced owner로 재검증한다.
- entry OFF 또는 시장 장애 중에도 fill, reconciliation, protection과 exit를 계속한다.
- 중앙 무결성 장애 시 외부 supervisor가 bounded RTO의 entry-incapable safety fallback을 기동한다.

**Non-Goals:**

- lane·autostart·automation gate·LIVE approval 기본값 변경 또는 자동 활성화.
- Guardian/ExecutionGateway를 우회하는 paper, shadow, canary 또는 동시 mutation path.
- KR/US를 하나의 calendar나 activation scope로 합치는 설정.
- 사람 승인 없는 live order, 운영 토글 변경 또는 capability 실측.
- broker가 증명하지 않은 idempotency 또는 symbol/time 추정 reconciliation.

## Decisions

### 1. 한 coordinator 아래 시장별 evaluation worker를 둔다

KR과 US는 각각 scheduler/calendar binding, activation manifest, evidence cursor, budget key와
typed state를 가진다. coordinator는 두 worker를 동시에 감독하지만 한 worker의
`WAIT_MARKET`, `BUDGET_DEFERRED` 또는 cycle-level refusal을 다른 worker의 입력으로 사용하지
않는다.

Market worker가 panic, 비정상 return, watchdog deadline 초과 또는 반복 crash threshold에
도달하면 supervisor는 그 market의 effective entry만 OFF로 durable latch하고 typed fault를
남긴다. 해당 worker는 bounded backoff/attempt policy 안에서만 재기동하며 latch 해제는 fresh
market authority와 명시적 recovery 조건을 요구한다. Peer market worker, fill detector,
reconciliation, protection, exit와 emergency reduction은 같은 process에서 계속 실행한다.

KR 성공 뒤 US worker를 시작하는 직렬 pipeline은 사용자 결정과 시장 독립성을 위반하므로
배제한다. `KR+US` 결합 activation은 어느 calendar와 approval이 권한을 부여했는지 숨기므로
배제한다.

### 2. 평가 concurrency와 fenced mutation single-writer를 분리한다

시장 worker는 immutable evidence로 router/lane의 순수 결정을 만들고 dispatch request만
발행한다. 하나의 dispatch owner가 journal transaction, Guardian generation, risk reservation과
ExecutionGateway 호출을 직렬화한다. Worker에는 broker mutator나 operating-setting writer를
주입하지 않는다.

Dispatch ownership은 journal에 저장된 monotonically increasing `owner_epoch`과 epoch마다
유일한 fencing token을 갖는다. Claim, 상태 transition과 Gateway 호출은 모두 current epoch/token을
compare-and-set하며 stale owner는 broker transport 전에 거부된다. 재기동 owner는 durable epoch를
증가시키고 이전 token을 영구 fence한다. 시장별 Gateway를 복제하는 대안은 account-wide
exposure와 nonce/reservation 권위가 갈라지므로 배제한다.

### 3. Dispatch lease는 비가역 권한 교집합과 상태기계다

Lease preimage는 account/market/symbol, candidate/evidence digest, router/lane/version,
campaign/leg, activation manifest와 generation, calendar generation, protection readiness
attestation/serial, reconciliation generation, risk policy/reservation, Guardian
decision/generation, build digest, owner epoch/fencing token와 expiry를 포함한다. 각 mutable
authority는 digest뿐 아니라 durable monotonic generation을 포함하므로 값이 A→B→A로 돌아와도
원래 lease와 일치하지 않는다.

Durable 상태는 `ISSUED`, `CLAIMED`, `SUBMITTING`, `SUBMITTED`, `AMBIGUOUS`, `REFUSED`로 제한한다.
정상 제출은 `ISSUED → CLAIMED → SUBMITTING → SUBMITTED`다. 하나의 atomic claim/validation
transaction이 current `ISSUED`를 current owner epoch/token으로 `CLAIMED`한 뒤 모든 current
authority를 검증한다. 검증 실패 또는 dispatch 취소는 `CLAIMED → REFUSED`, broker 호출 직전은
durable transport-start marker와 함께 `CLAIMED → SUBMITTING`으로 전이한다. Exact broker acceptance는
`SUBMITTING → SUBMITTED`, definitive broker rejection 또는 exact identity/query가 증명한
authoritative no-accept/no-fill은 `SUBMITTING → REFUSED`, acceptance 여부를 durable하게 확정할 수
없는 transport uncertainty만 `SUBMITTING → AMBIGUOUS`다.
`SUBMITTED`, `AMBIGUOUS`, `REFUSED`는 outgoing transition이 없는 terminal state다.

검증 실패도 lease를 소비한다. Current `ISSUED` lease의 bound authority가
missing/changed/expired/stale이거나 fence/scope가 맞지 않거나 pre-transport cancel이면 모두
broker request 0건의 `REFUSED`이며 같은 transaction에서 그 lease의 exact reservation을
`RELEASED`한다. 이미 소비된 terminal lease를 replay하면 원래 lease와 reservation disposition은
변경하지 않고 retry attempt만 typed `REFUSED`로 기록하며, retry attempt가 별도 exact HELD
reservation을 잘못 만들었다면 그것만 원자 `RELEASED`한다. 어떠한 claim/validation 시도도 lease를
`ISSUED`로 되돌리거나 다른 owner가 재검증하게 해서는 안 된다. `CLAIMED` 뒤 crash는 durable
transport-start marker가 없으므로 반드시 `REFUSED + RELEASED`로 복구한다. `SUBMITTING` 뒤
crash만 exact broker outcome을 조회해 `SUBMITTED + TRANSFERRED`, `REFUSED + RELEASED` 또는
durable uncertainty인 `AMBIGUOUS + HELD` 중 하나로 원자 확정한다.
따라서 authority가 A→B→A가 되거나 실패 원인이 사라져도 같은 lease는 부활하지 않으며 fresh
decision과 fresh lease만 새 권한을 만들 수 있다.

Lease terminal state와 reservation disposition은 서로 다른 durable record지만 outcome
classification transaction에서 함께 commit한다. Broker transport 전 `REFUSED`, definitive broker
rejection 또는 post-transport exact query가 증명한 authoritative no-accept/no-fill은 broker outcome
code, operation/order identity, query/evidence digest와 observed-at을 기록하면서 exact risk/campaign
reservation을 같은 journal transaction에서 `RELEASED`로 바꾼다. `SUBMITTED`면 reservation을
attempt/order fill·cancel lifecycle로 `TRANSFERRED`한다. Durable transport uncertainty인
`AMBIGUOUS`만 reservation을 `HELD`로 동결하고 같은 campaign/risk capacity에
대한 release, reuse와 fresh lease 발급을 금지한다. Exact broker reconciliation이 authoritative
`NOT_SUBMITTED`를 증명한 때만 별도 reconciliation outcome이 `HELD→RELEASED`, acceptance를
증명하면 `HELD→TRANSFERRED`를 기록한다. 이 disposition 변경은 terminal lease state를
되돌리거나 다시 제출 가능하게 만들지 않는다.

### 4. Broker unknown outcome은 attested capability가 재시도 상한이다

`SUBMITTING` 뒤 timeout, connection loss, process crash 또는 malformed response는 즉시 성공이나
거부로 추정하지 않고 exact outcome classification을 수행한다. Definitive rejection response 또는
attested exact query의 authoritative no-accept/no-fill은 `REFUSED + RELEASED`, exact acceptance는
`SUBMITTED + TRANSFERRED`다. 이 둘을 증명할 수 없는 durable transport uncertainty만
`AMBIGUOUS + HELD`다. a071 protection/broker capability attestation이 client operation key 전달/echo,
exact lookup identity, uniqueness scope, pending/terminal query와 duplicate-submit idempotency를
모두 현재 market/order scope에 대해 증명한 경우에만 reconciliation이 같은 operation key를
조회하고 필요 시 bounded 재제출할 수 있다. 이 경우에도 새 lease를 만들거나 owner fence를
우회하지 않는다.

Capability가 없거나 일부만 증명돼 acceptance 여부가 확정되지 않으면 `AMBIGUOUS`를 유지하고 exact query가 가능한 범위에서만
대사하며 자동 resubmit은 0건이고 reservation은 `HELD`다. Broker에 존재할 수 있는 주문을 not-found로 추정하거나
symbol/time으로 dedup하지 않는다. Cancel/replace unknown도 exact order identity로 terminal
상태를 확인하기 전 성공으로 간주하지 않는다.

### 5. Safety loops와 entry workers는 서로 다른 생존 등급을 가진다

Fill detection, reconciliation, protection supervisor, exit observer와 emergency reduction은
safety class로 계속 실행된다. Lane/automation OFF, market close, stale evidence, entry budget
고갈과 한 market worker 비정상은 해당 시장의 신규 entry/scale-in만 latch OFF한다. Safety
class는 entry worker와 다른 cancellation context, queue와 reserved API budget을 사용한다.

Journal corruption, Gateway invariant violation, owner epoch/fence CAS 불능 또는 둘 이상의 current
owner 같은 central integrity fault는 모든 신규 entry를 차단하고 process를 fail-closed할 수 있다.
이 경우 in-process recovery를 safety 보장으로 간주하지 않는다. 별도 deployment domain의
external supervisor는 heartbeat 상실을 감지하고, 이전 owner token을 fence한 새 epoch로
entry capability가 없는 safety-only fallback을 `safety_fallback_rto` 안에 시작해야 한다.
`safety_fallback_rto`는 versioned deployment manifest에 고정된 양의 값이며 60초를 초과할 수
없다. Fallback은 fill/reconciliation/protection/reduce-only exit/emergency reduction만 소유하고
entry lease를 발급할 수 없다. Fallback 기동도 실패하면 broker-resident protection은 건드리지
않고 지속 critical alert와 명시적 `SAFETY_FALLBACK_UNAVAILABLE` 상태를 발행한다.

### 6. Runtime은 dormant 배포와 사람 activation을 분리한다

Code와 descriptor를 배포해도 저장값이 없는 lane, scheduler, autostart와 automation은 OFF이고
LIVE approval은 미승인이다. Runtime은 existing human-approved per-market activation manifest만
소비하며 승인 생성 API를 제공하지 않는다. 테스트는 fake clock, fixed evidence와 isolated
official broker를 사용하며 live mutation을 구조적으로 차단한다.

### 7. First-leg authority는 additive v26 companion으로 원자 결합한다

기존 q_final, strategy lineage, PositionCampaign과 risk owner API를 순서대로 호출하면 prospective
token과 decision/campaign 선행조건이 순환하고 중간 crash window가 생긴다. Journal이 prospective
token을 직접 mint하고 q_final decision/aggregate와 정확히 다섯 bucket HELD, strategy
decision/attempt, campaign/claim, risk owner, leg 1을 한 transaction에서 commit한다. 이 API는
dispatch lease나 execution lineage `DISPATCH_START`를 만들지 않으므로 제출 권한이 아니다.

Released v20-v25 schema는 수정하지 않는다. v26은 immutable
`strategy_first_leg_bindings` companion을 추가해 q_final decision, aggregate reservation,
strategy decision/attempt, campaign/prospective claim, risk owner/token과 first leg sequence/plan을
exact scope 및 request digest에 포함된 고정 `router_id`/`router_version`으로 묶는다. BEFORE INSERT trigger는 모든 referenced row의 account/market/symbol,
lane, decision, token과 quantity가 일치하는지 검증한다. 별도 additive lease trigger는 future
dispatch lease의 Guardian decision, aggregate reservation, campaign, `leg_id`(first-leg plan id),
lane/router와 scope가 이 binding과 일치하지 않거나 `operation_id`가 bound strategy attempt의 실제
`client_order_id`가 아니면 거부한다. Historical row는 backfill하지 않으며,
binding이 없는 legacy authority는 신규 exposure-raising lease를 만들 수 없다.

KR/KRW와 US/USD는 같은 request/transaction 구현과 paired RED/GREEN suite를 공유한다. 한 시장
구현의 안정화를 다른 시장 시작 조건으로 두지 않으며 두 시장이 모두 통과하기 전에는 이 slice를
완료로 표시하지 않는다.

### 8. KR과 US는 하나의 account-base Guardian과 frozen FX authority를 공유한다

계좌마다 하나의 `RiskGuardian`과 하나의 account-wide exposure/daily-loss cap만 둔다. 기존
`limit_currency`는 시장 선택 통화가 아니라 account base currency로 정의한다. KR과 US별
quote-currency Guardian을 두면 각 Guardian이 같은 account-wide cap을 전부 소비할 수 있으므로
배제한다.

KR은 same-currency라도 봉인된 identity FX authority를 사용하고, US는 request-scoped official
quote-to-base FX와 conservative haircut을 사용한다. 그 하나의 authority가 Guardian sizing,
aggregate와 정확히 다섯 bucket reservation, versioned decision limits envelope,
`CLAIMED→SUBMITTING` fence와 Gateway validation까지 바뀌지 않고 전달된다. Cash와 broker order
cost는 quote 통화로 비교하고 exposure, loss, equity, policy limits는 base 통화로 비교한다.
Base reservation은 ceil, admissible quantity는 floor한다.

이 결정도 paired-delivery gate의 일부다. KR identity와 US conversion의 RED/GREEN, production
adapter, Gateway mismatch와 failure-isolation 검증을 같은 wave에서 완료하지 않으면 구현은
미완료다. FX 장애는 해당 시장의 exposure-raising entry만 닫고 protection, fill,
reconciliation과 reduce-only exit를 닫지 않는다.

## Risks / Trade-offs

- [두 market worker가 account-wide risk를 경합] → fenced central dispatch owner와 한 journal
  transaction의 reservation으로 직렬화한다.
- [한 worker의 느린 API가 전체 runtime을 막음] → market queue, context와 budget key를
  분리하고 bounded handoff만 공유한다.
- [Dispatch 직전 authority drift 또는 ABA] → generation을 포함한 lease를 비가역 소비하고
  mismatch는 terminal `REFUSED`로 만든다.
- [Broker acceptance 뒤 응답 손실] → attested exact identity/idempotency가 있을 때만 bounded
  same-key recovery, 없으면 `AMBIGUOUS` no-resubmit과 reservation `HELD`다.
- [Central integrity fault로 process 종료] → external fenced safety-only fallback을 최대 60초
  RTO 안에 기동하고 entry capability는 계속 제거한다.
- [시장별 Guardian이 account cap을 중복 소비] → 계좌당 하나의 base-currency Guardian과 하나의
  base reservation domain만 사용한다.
- [FX TOCTOU·역방향·단위 혼합] → official evidence에서 mint한 frozen authority 하나를 sizing부터
  Gateway까지 결합하고 stale/pair/digest mismatch는 transport 전에 거부한다.

## Migration Plan

1. Runtime graph와 KR/US market worker를 같은 wave의 fake inputs에서 구성하고 Gateway를 닫은 채 검증한다.
2. Additive v26 first-leg binding과 router/client-order lease authority 및 KR/US atomic admission을 함께 적용하고 v25 preservation,
   migration rollback과 old-build refusal을 검증한다.
3. KR identity와 US official conversion을 같은 wave에서 account-base Guardian, aggregate/five
   reservation과 versioned decision envelope에 연결한다.
4. Durable lease state machine, owner epoch/fence와 single dispatch owner를 journal/Gateway에
   연결한다.
5. Existing safety loops, market worker isolation과 external safety-fallback supervisor를 fault
   injection으로 통합 테스트한다.
6. OFF/미승인 defaults와 60초 이하 safety fallback RTO manifest로 binary/services를 배포한다.
   배포 자체는 engine entry를 시작하지 않는다.
7. Rollback은 entry worker를 제거하거나 OFF로 유지하되 existing safety loops와 journal
   lineage를 계속 읽고 `AMBIGUOUS` attempt는 capability-attested reconciliation으로만 해소한다.

## Open Questions

없음. 실제 activation manifest와 capability evidence가 없으면 구현은 해당 시장을 OFF로
유지하며 승인, broker idempotency 또는 근거를 합성하지 않는다.
