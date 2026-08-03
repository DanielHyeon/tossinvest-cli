## Why

후보, scheduler, strategy lane, risk, Guardian과 official execution 구성요소는 각각
존재하지만 production에서 하나의 supervised entry loop로 이어져 있지 않다. 특히
KR 운영 안정화를 US 구현의 선행조건으로 두면 한 시장의 일정·evidence·거부가 다른
시장의 개발과 평가까지 직렬화한다. KR과 US를 같은 delivery wave에서 독립적인
activation/calendar scope로 연결하되, durable dispatch 직전에는 모든 안전 근거를 다시
검증하는 runtime 계약이 필요하다.

## What Changes

- approved candidate를 evidence snapshot, market scheduler, horizon router, selected lane,
  risk sizing, Guardian decision, strategy dispatch와 official ExecutionGateway로 연결하는
  supervised production entry loop를 구성한다.
- KR과 US evaluation worker를 동시에 실행하고 calendar, activation, budget, refusal과
  failure domain을 시장별로 격리한다. 한 시장의 휴장·거부·지연은 다른 시장의 eligible
  evaluation을 정지시키지 않는다.
- exposure-raising mutation마다 durable dispatch lease에서 lane/version, candidate/evidence,
  activation manifest, market calendar, protection readiness, reconciliation health, risk와
  Guardian generation을 다시 대조한다. Lease는 durable owner epoch/fencing token과
  `ISSUED → CLAIMED → SUBMITTING → SUBMITTED | AMBIGUOUS | REFUSED` 상태로 영속하며,
  claim/validation 시도는 성공 dispatch 또는 terminal refusal로 비가역 소비한다.
- Broker submit 결과가 unknown이면 a071 attestation이 exact identity/query/dedup과
  idempotency를 증명할 때만 같은 operation key 재제출을 허용한다. 증명이 없으면
  `AMBIGUOUS` reconciliation으로 남기고 추정 재제출하지 않는다.
- Risk/campaign reservation은 lease terminal state와 별도 durable disposition으로 보존한다.
  broker 전 `REFUSED`는 같은 transaction에서 release, `SUBMITTED`는 order lifecycle로 transfer,
  `AMBIGUOUS`는 exact reconciliation 전까지 `HELD`로 동결해 release나 새 lease를 금지한다.
- lane 또는 automation OFF는 신규 entry와 scale-in만 중단한다. fill detection,
  reconciliation, broker protection, stop/exit와 emergency reduction loop는 계속 실행한다.
- 시장 worker의 비정상 반환은 그 시장만 effective OFF로 latch하고 bounded restart하며
  peer market과 safety loop를 계속한다. Journal/Gateway/dispatch-owner/fence 무결성 장애는
  전체 신규 entry를 닫고 외부 supervisor가 bounded RTO 안에 fenced safety-only fallback을
  기동하도록 한다.
- runtime을 구현·배포해도 KR/US lane, autostart, automation gate와 LIVE approval의 기본값은
  OFF/미승인으로 유지한다. 이 change는 운영 activation이나 live order를 수행하지 않는다.

## Capabilities

### New Capabilities

- `strategy-runtime`: KR·US의 독립적이면서 동시에 실행되는 supervised evaluation과
  evidence-to-official-gateway 연결, dispatch lease 재검증 및 safety-loop 생존 계약.

### Modified Capabilities

- `strategy-engine`: approved candidate와 선택된 market/horizon lane을 production dispatch
  lineage로 연결하고 lane OFF에서 신규 진입만 0건으로 만든다.
- `market-aware-scheduler`: 결합 market scope를 만들지 않고 KR·US별 calendar/activation/
  budget binding을 동시에 운용하며 한 시장 상태가 다른 시장을 막지 않게 한다.
- `engine-safety`: exposure-raising dispatch 직전에 모든 immutable safety evidence를 같은
  lease에서 재검증하고, entry loop 장애 중에도 protection·exit·reconciliation을 유지한다.

## Impact

- `internal/app/engine`: supervised runtime assembly, lifecycle, worker isolation과 shutdown.
- `internal/strategyengine`, scheduler와 multi-market router: candidate-to-lane dispatch,
  independent KR/US evaluation과 typed refusal lineage.
- `internal/risk`, Guardian, journal, protection과 `internal/execgw`: 기존 권한 경계를
  유지한 lease 검증 및 official-only submission integration.
- `cmd/tossctl` engine profile과 runtime tests: dormant defaults, failure propagation,
  concurrent market progress와 safety-loop continuity 검증.
