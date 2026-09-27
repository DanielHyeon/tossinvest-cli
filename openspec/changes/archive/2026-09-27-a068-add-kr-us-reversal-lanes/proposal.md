# Change: a068-add-kr-us-reversal-lanes

## Why

KR 수급 흡수와 US 급락·유동성 이탈은 서로 다른 증거를 요구하지만, 현재 engine에는 이를
구조적 반전으로 확인해 단계 진입하는 전용 lane이 없다. 가격 하락만으로 물타기를 허용하지 않는
명시적 계약이 필요하며, KR 완성 뒤 US를 착수하는 순차 계획이 아니라 두 시장을 같은 delivery
wave에서 독립적으로 완성해야 한다.

## What Changes

- KR absorption reversal과 US dislocation reversal을 별도 ID·version·market scope의 순수 lane으로
  동시에 추가한다. KR의 안정화·활성화 여부는 US 구현·검증의 선행 gate가 아니다.
- KR absorption과 US dislocation은 source lineage, units, effective/observed/ingested/evaluated/
  fresh-until, integer-normalized metrics, thresholds와 schema/config digest를 가진 별도 strict
  versioned input schema를 사용한다. sweep, market-structure break, reclaim은 같은 scope에서
  `sweep_at <= break_at <= reclaim_at <= evaluated_at`과 versioned bounded window를 만족해야 한다.
- 2:4:8 leg plan을 정의하되 최종 leg는 단순 가격 하락으로 해제하지 않고, 시장별 sweep,
  market-structure break, reclaim 증거가 모두 유효할 때만 허용한다.
- 2:4:8 plan은 immutable campaign risk budget과 planned quantity에서 weights/14, 앞 leg floor,
  마지막 leg remainder로 계산한다. actual fill·cancel·retry는 어떤 leg도 상향 재배분하지 못하고
  각 leg는 a066 cap 이하이다. Actual fill price·fees·FX의 monetary risk가 immutable campaign
  budget을 넘으면 overage latch로 후속 신규 leg를 차단하고 fill/common exit 처리는 유지한다.
- exit 또는 structural invalidation이 관측되면 lane은 typed invalidation/refusal만 반환해 모든
  scale-in을 억제한다. 실제 exit decision은 공통 exit engine만 발급하며 lane은 그 독립 권한을
  대신하거나 지연하지 않는다. scale-in은 effective stop을 낮출 수 없다.
- 결정과 refusal에 market, lane ID/version, evidence digest, campaign/leg lineage를 기록해 KR·US
  성과와 거부 사유를 독립 귀속한다.
- lane 추가는 LIVE 주문이나 운영 토글을 실행하지 않으며 후속 runtime wiring 전까지 OFF다.

## Capabilities

### Modified Capabilities

- `strategy-engine`: KR·US 단기 reversal lane, 2:4:8 progression, 구조 확인, exit-first 및 독립
  lineage 요구사항을 추가한다.

## Impact

- Affected code: strategy lane registry와 KR/US reversal evaluator, structural evidence fixtures,
  campaign/leg 연동 및 lane별 attribution tests
- Dependencies: `a064-add-multi-market-strategy-evidence`, `a065-add-position-campaign-leg-core`,
  `a066-add-multi-horizon-risk-buckets`
- Safety: broker·journal·toggle mutation은 lane 밖에 유지하며, 가격 하락만으로 추가 매수하지 않고
  exit 및 emergency liquidation의 우선순위를 보존한다.
