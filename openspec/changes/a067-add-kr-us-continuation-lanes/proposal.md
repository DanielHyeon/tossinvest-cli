# Change: a067-add-kr-us-continuation-lanes

## Why

현재 production strategy lane은 KR 전용 Parker VWAP 한 개뿐이고, KR·US의 단기 추세 지속을
각 시장의 증거와 세션 규율로 평가하는 소유 lane이 없다. KR 안정화를 US 개발의 선행 조건으로
두면 US 경로가 다시 미완성 상태로 남으므로, 두 시장의 continuation lane을 같은 delivery wave에서
독립적으로 설계·구현해야 한다.

## What Changes

- KR flow continuation과 US participation continuation을 서로 다른 ID·version·market scope를 가진
  순수 lane으로 동시에 추가한다. 두 lane 중 하나의 완료나 운영 상태가 다른 lane의 개발·평가를
  gate하지 않는다.
- KR flow와 US participation 입력은 generic metric map이 아니라 source lineage, units,
  effective/observed/ingested/evaluated/fresh-until, integer-normalized metric, threshold와
  schema/config digest를 가진 서로 다른 strict versioned schema로 고정한다. evaluator는 그
  integer arithmetic과 비교 순서를 그대로 사용하고 float 또는 누락값 추정을 하지 않는다.
- 각 lane은 승인 후보와 point-in-time 시장 증거를 받아 결정 또는 typed refusal만 반환하며 broker,
  journal writer, 운영 토글을 직접 호출하지 않는다.
- 8:4:2 leg plan은 campaign 생성 시 immutable risk budget과 planned quantity를 고정하고
  weights/14, 앞 leg floor, 마지막 leg remainder로 계산한다. partial fill·cancel·retry의 미사용
  수량을 현재나 후속 leg로 상향 재배분하지 않으며 각 leg는 a066 cap 이하이다. Actual fill
  price·fees·FX로 monetary risk를 재계산해 immutable campaign budget 초과 시 overage latch로
  후속 신규 leg를 차단하되 fill/exit 처리는 계속한다.
- 시장별 continuation 재확인 뒤에만 다음 leg를 허용하고 scale-in으로 effective stop을 낮추지
  않는다. lane은 exit decision을 만들지 않고 exit/structural invalidation을 typed invalidation으로
  반환하며, 공통 exit engine의 독립적인 risk-reducing 권한을 대신하거나 지연하지 않는다.
- KR와 US 결정은 lane ID/version, market, evidence digest와 campaign/leg lineage를 분리 보존해
  성과를 독립 귀속한다.
- lane 추가만으로 LIVE entry나 automation을 켜지 않는다. activation manifest, Guardian, protection
  readiness와 사람 승인은 후속 runtime change의 책임이다.

## Capabilities

### Modified Capabilities

- `strategy-engine`: KR·US 단기 continuation lane의 순수 결정 계약, 8:4:2 progression, refusal 및
  독립 lineage 요구사항을 추가한다.

## Impact

- Affected code: strategy lane registry와 KR/US continuation evaluator, decision/refusal fixtures,
  campaign/leg 연동 및 lane별 attribution tests
- Dependencies: `a064-add-multi-market-strategy-evidence`, `a065-add-position-campaign-leg-core`,
  `a066-add-multi-horizon-risk-buckets`
- Safety: 주문·토글 mutation을 추가하지 않고 lane 기본 상태는 OFF다. 기존 exit, reconciliation,
  emergency liquidation 경로와 승인 경계는 변경하지 않는다.
