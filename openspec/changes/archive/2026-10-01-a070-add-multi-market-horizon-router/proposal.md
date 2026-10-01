# Change: a070-add-multi-market-horizon-router

## Why

KR·US에 short와 weekly lane을 추가해도 후보를 시장·시간축·기존 campaign ownership에 따라 한
소유 lane으로 결정적으로 보내는 계층이 없으면 중복 진입과 API 예산 경합이 발생한다. 또한 KR와
US를 결합 activation으로 묶으면 한 시장의 휴장·장애·비활성화가 다른 시장 평가까지 막는다.
두 시장은 같은 delivery wave에서 배선하되 런타임 상태는 독립이어야 한다.

## What Changes

- candidate의 account, market, symbol, position generation, horizon, eligible lane version과 기존
  campaign ownership을 평가해 ownership key당 최대 한 owning lane을 선택하는 deterministic
  router를 추가한다. owner key는 `(account, market, symbol, position_generation)`이며 horizon은
  attribution/admission 정보일 뿐 owner identity가 아니다.
- router는 모든 horizon의 active owner를 조회한다. 이미 campaign owner가 있으면 새 scoring
  결과가 소유권을 빼앗지 않으며, 모호함·충돌·필수 evidence 부재는 typed refusal로 fail closed한다.
- KR와 US calendar, IANA timezone, DST, regular-session scope, activation binding을 독립 평가한다.
  KR 안정화·활성화가 US의 개발 또는 eligible evaluation의 선행 조건이 아니며 그 역도 같다.
- 한 시장이나 lane이 disabled, closed, stale 또는 실패해도 다른 시장의 eligible evaluation을
  차단하지 않는다. 결합 KR+US activation manifest는 만들지 않는다.
- short와 weekly, KR와 US low-priority capability는 admission/anti-replay subscope로 분리하되 같은
  physical endpoint/reset generation의 authoritative remaining, commitments와 absolute issuance cap을
  공유한다. scope 분할로 quota가 늘어나서는 안 되며 어느 cadence도 safety reserve를 소비하지 않는다.
- scheduler desired state는 시장별 revision, lock과 activation binding으로 원자 저장한다. legacy
  disabled는 양 시장 OFF로, 검증된 single-market state는 해당 시장만 이관하고 peer는 OFF로
  남기며 independent CAS/rollback과 crash recovery를 적용한다.
- router는 결정·refusal만 만들고 broker, journal writer 또는 운영 토글을 직접 변경하지 않는다.

## Capabilities

### Added Capabilities

- `multi-market-horizon-router`: 시장·시간축·campaign ownership 기반의 단일 lane 선택과
  typed-refusal 계약을 정의한다.

### Modified Capabilities

- `market-aware-scheduler`: KR·US의 독립 calendar/activation 평가와 short/weekly 저우선순위 예산
  격리를 정의한다.

## Impact

- Affected code: candidate-to-lane router, lane registry/ownership lookup, scheduler market bindings,
  cadence budget classes, deterministic routing and isolation tests
- Dependencies: `a067-add-kr-us-continuation-lanes`, `a068-add-kr-us-reversal-lanes`,
  `a069-add-kr-us-weekly-value-lanes`
- Safety: router는 주문 mutation 권한이 없고 activation을 생성하지 않는다. disabled/OFF는 신규
  entry만 억제하며 exit, reconciliation, fill detection과 protection supervision은 계속된다.
