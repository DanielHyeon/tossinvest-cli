## Why

KR·US lane를 동시에 구현해도 운영자가 시장별 evidence, campaign, risk, scheduler와
첫 거부 사유를 분리해 확인할 수 없으면 dormant 배포 상태와 실제 entry-ready 상태를
구분할 수 없다. 또한 시장·lane·version·campaign lineage가 가시성 계층에서 합쳐지면
성과가 symbol/time 추정으로 잘못 귀속될 수 있다. runtime activation과 배포를 분리한
read-only 운영 표면과 dormant deployment 검증이 필요하다.

## What Changes

- 로컬 콘솔과 private read API에서 KR·US별 lane desired/effective, evidence freshness,
  campaign/leg, horizon risk bucket, scheduler/calendar, protection/reconciliation 및 첫 typed
  refusal을 독립된 상태로 제공한다.
- candidate부터 lane/version, campaign/leg, decision, attempt, order, fill, position/close까지
  persisted identifiers로만 시장별 성과를 귀속하고 누락은 `link_missing` 또는
  `not_measured`로 표시한다.
- 부분체결과 staged close를 fill delta 단위로 귀속하고, authoritative position cost-basis
  policy/version에 따라 closed/residual quantity, gross PnL, fees, FX와 net PnL의 보존식을
  강제한다. 누락 FX/fee 근거를 0으로 꾸미지 않는다.
- 화면과 API는 동일한 server-owned descriptor/status projection을 사용하고, 한 시장이
  unavailable이어도 다른 시장 상태와 health를 계속 반환한다.
- Compose services를 build/replace하기 전에 exact image digest, compose/config/activation/
  protection digest, volume/mount와 schema compatibility range를 rollback preimage로 동결한다.
  한 service씩 bounded replace하고 실패 시 이미 교체된 subset만 역순 부분 rollback한다.
- 배포는 저장된 autostart, automation gate, lane desired state, LIVE approval 또는 보호
  설정을 변경하지 않는다. post-deploy check는 주문 mutation과 activation 없이 수행하며,
  실제 운영 전환은 별도의 사람 승인 절차로 남긴다.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `operator-console`: KR·US lane runtime의 시장별 evidence, campaign, risk, scheduler,
  safety readiness와 첫 refusal을 같은 registry에서 읽기 전용으로 설명한다.
- `http-api-service`: private read API가 콘솔과 동일한 시장별 runtime projection과 dormant
  health를 제공하며 LIVE/activation mutation surface를 추가하지 않는다.
- `lane-performance`: market, lane/version과 campaign/leg의 완전한 identifier lineage만으로
  성과를 귀속하고 시장 간 또는 symbol/time 추정 귀속을 금지한다.

## Impact

- `internal/console` 및 status registry: multi-market runtime views와 dormant health 표시.
- `internal/httpapi`: private read resources/SSE projection과 unavailable-market isolation.
- performance/query projection: market/lane/version/campaign lineage와 missing-state reporting.
- `compose.yaml`, build/deploy verification 및 운영 문서: OFF 상태를 보존한 service replace와
  compatibility gate, digest-pinned partial rollback과 post-deploy no-mutation health check.
