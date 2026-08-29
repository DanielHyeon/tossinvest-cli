# Change: a069-add-kr-us-weekly-value-lanes

## Why

현재 engine에는 공시 시점에 고정된 fundamental evidence를 사용해 주간 가치 재평가를 수행하는
lane이 없다. KR OpenDART와 US SEC EDGAR는 출처·정정·희석 신호가 다르므로 시장별 계약이 필요하다.
KR 운영 안정 후 US를 시작하는 순차 접근을 피하고 두 시장의 weekly lane을 같은 delivery wave에서
독립적으로 구현해야 한다.

## What Changes

- OpenDART point-in-time evidence를 쓰는 KR weekly value lane과 SEC EDGAR point-in-time evidence를
  쓰는 US weekly value lane을 별도 ID·version·market scope로 동시에 추가한다. 어느 시장도 다른
  시장의 개발 완료나 운영 활성화를 기다리지 않는다.
- 입력은 filing/revision identity, as-of/observed/ingested/evaluated/fresh-until, currency/unit,
  diluted shares, dilution facts, model/config digest와 fair value를 포함하는 시장별 strict versioned
  schema로 고정한다.
- 공식 exchange calendar와 IANA timezone이 만든 market-week마다 atomic unique reservation을
  하나만 허용하고 전체 최대 일곱 planned leg를 적용한다. zero-fill cancel/expiry만 reservation을
  release하고 어떤 positive fill도 그 주를 소비하며 idempotent retry는 새 leg가 아니다.
- campaign risk budget과 일곱 planned quantity ceiling은 생성 시 고정하고 actual fill·cancel·retry로
  현재/후속 leg를 상향 재배분하지 않는다. 각 leg는 a066 cap 이하이며 매 leg 전에 value thesis,
  새 공시, dilution, projected post-trade risk를 다시 검증한다. Actual fill price·fees·FX의 monetary
  risk가 immutable campaign budget을 넘으면 overage latch로 후속 신규 leg를 차단하되 fill/common
  exit 처리는 계속한다.
- scale-in으로 effective stop을 낮추지 않고, 필요한 structural stop이 승인된 risk cap보다 넓으면
  entry를 typed refusal로 막는다.
- target은 `min(staged_target, fair_value)`로 제한하고 비용 후 minimum RR이 실패하면 entry를
  거부한다. lane은 exit decision을 만들지 않으며 typed invalidation/refusal만 반환하고 공통 손절,
  emergency exit, reconciliation과 exit engine의 독립 권한을 약화시키지 않는다.
- 시장, lane ID/version, disclosure/evidence digest와 campaign/leg lineage를 분리 기록하고 lane
  추가만으로 LIVE 또는 automation을 활성화하지 않는다.

## Capabilities

### Modified Capabilities

- `strategy-engine`: KR·US weekly value lane의 순수 평가, 7-leg 주간 progression, 재검증,
  stop/target 규율과 독립 lineage 요구사항을 추가한다.

## Impact

- Affected code: strategy lane registry와 KR/US weekly value evaluator, disclosure/value fixtures,
  campaign/leg 연동 및 lane별 attribution tests
- Dependencies: `a064-add-multi-market-strategy-evidence`, `a065-add-position-campaign-leg-core`,
  `a066-add-multi-horizon-risk-buckets`
- Safety: stale/missing/corrected disclosure는 fail-closed refusal이며, fundamental adapter의 secret은
  주문 권한을 주지 않는다. lane 기본 상태는 OFF이고 기존 exit 안전 경로를 보존한다.
