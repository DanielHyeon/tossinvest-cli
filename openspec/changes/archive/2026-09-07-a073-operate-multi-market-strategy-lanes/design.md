## Context

a072 runtime은 시장별 execution state와 durable lineage를 만들지만 운영 표면이 이를 독립적으로
읽지 못하면 KR의 상태가 US를 가리거나 dormant 배포를 entry-ready로 오인할 수 있다. 기존
console과 private API는 shared registry/runtime-only Unix endpoint 패턴을 이미 사용하고,
lane-performance는 결정적 identifier chain만 허용한다. 이 change는 그 패턴을 multi-market
strategy runtime으로 확장하고 Compose 배포와 activation을 분리한다.

2026-08-03 read-only production baseline에서 `/strategy-runtime`은 runtime seam `미배선`,
KR Parker lane만 표시, runtime `UNOBSERVED`, lane/autostart/entry `OFF`, source manifest
`NOT_CONFIGURED`, candidate/scheduler/Guardian/reconciliation `MISSING`, ProtectionReady
`UNWIRED`, activation absent였다. 시장·일정 화면도 scheduler `DISABLED/OFF`, 선택 시장 없음,
calendar unverified와 `NOT_ACTIVATED`를 표시했다. 따라서 이 change의 첫 operational truth는
두 시장을 ready로 꾸미는 것이 아니라 이 blocker와 unavailable 상태를 분리해 정직하게
투영하는 것이다.

## Goals / Non-Goals

**Goals:**

- console과 private API에 동일한 KR·US runtime projection을 제공한다.
- evidence, campaign/leg, risk, scheduler, protection/reconciliation과 first refusal을 시장별로
  정직하게 표시한다.
- market/lane/version/campaign identifier lineage로만 성과를 귀속하고 partial fill/staged close
  수량·비용·FX·PnL 보존을 검증한다.
- OFF 설정을 보존한 digest-pinned Compose replace, compatibility gate와 부분 rollback을 정의한다.

**Non-Goals:**

- 새로운 LIVE, gate, lane activation 또는 protection-weakening mutation route.
- 배포 과정의 autostart, automation, lane desired 또는 LIVE approval 변경.
- runtime unavailable 값을 default/zero/다른 시장 값으로 대체.
- symbol/time 추정 성과 귀속, 누락 비용/FX의 0 처리 또는 post-deploy live order 검증.
- schema compatibility가 증명되지 않은 binary rollback.

## Decisions

### 1. Server-owned multi-market status projection을 하나만 둔다

Projection key는 market이며 lane desired/effective, evidence freshness/digest, campaign/leg,
risk bucket, scheduler/calendar, activation, protection/reconciliation, first refusal과 observed-at을
포함한다. Console HTML과 HTTP JSON/SSE adapter는 같은 projection/descriptor를 소비한다.
ProtectionReady 외부 enum은 정확히 `WIRED`/`UNWIRED`만 사용하고 실패·unavailable 상세는 별도
typed refusal/error envelope에 둔다. `READY`, `DEGRADED`, `UNKNOWN` 같은 제3 readiness enum은
만들지 않는다.

각 adapter가 값을 재계산하는 대안은 default와 refusal 의미가 drift하므로 배제한다.

### 2. Runtime read는 existing authenticated Unix endpoint를 확장한다

Compose sidecar는 shared private engine directory의 runtime-only authenticated Unix endpoint에서
projection을 읽는다. Endpoint는 GET/read stream만 제공하고 preview/apply, lane/LIVE activation,
order, protection mutation capability를 노출하지 않는다. 한 market read가 실패하면 그 market만
typed unavailable envelope가 되고 다른 market snapshot은 유지된다. ProtectionReady를 읽지
못한 경우 readiness 값은 `UNWIRED`와 `runtime_unavailable` typed refusal로 fail-closed하며,
일반 runtime availability 자체는 readiness enum에 섞지 않는다.

Journal 파일을 sidecar가 임의 조인하는 대안은 runtime effective state와 first refusal을
재구성할 수 없고 writer boundary를 흐리므로 배제한다.

### 3. 성과 귀속은 persisted composite lineage와 보존식을 사용한다

Performance key는 market, lane/version, campaign/leg, candidate, decision, attempt, order, fill,
position/close identifier와 policy version을 요구한다. 한 링크라도 없으면 `link_missing`으로
분리하며 symbol/time proximity나 다른 market의 동일 ticker로 보정하지 않는다. Metric observation
부재는 `not_measured`다.

Projector는 cumulative order quantity를 합산하지 않고 deduplicated fill event의 signed quantity
delta만 적용한다. Partial entry/exit fill과 staged close 각각은 고유 fill/close-leg identity를
가지며 correction/bust는 원 fill을 가리키는 별도 delta로 반영한다. Entry acquired quantity는
`attributed closed quantity + authoritative residual position quantity`와 같아야 하고, staged close는
그 delta만 realized로 이동하며 잔여 수량은 open 상태에 남는다. Position이 완전히 닫힐 때까지
부분 close row를 하나의 추정 closed trade로 합치지 않는다.

Cost basis allocation은 authoritative journal에 저장된 position cost-basis policy/version과
rounding rule만 사용한다. 각 close delta는 allocated entry basis, exit proceeds, broker fee,
exchange fee, tax와 기타 persisted cost를 원 currency로 보존한다. Reporting currency 환산은
persisted FX source/rate/as-of/quote currency와 rounding version을 요구한다. 보존식은 각 currency와
reporting currency에서 `gross_pnl - entry_fees - exit_fees - taxes - fx_cost = net_pnl`이고,
모든 close delta의 allocated basis와 quantity 합은 authoritative position totals를 초과하거나
잃어서는 안 된다. Fee 또는 FX evidence가 없으면 해당 metric은 `not_measured`이며 0 또는 최신
환율로 대체하지 않는다.

### 4. 배포와 activation을 서로 다른 단계로 고정한다

Compose build/replace는 config와 private engine data volume을 보존하고 activation 관련 환경값을
덮어쓰지 않는다. Post-deploy check는 service health, console/API schema, Unix projection 연결과
두 market의 dormant OFF/typed reason을 read-only로 확인한다. Engine entry process 시작,
operating toggle 변경과 live order는 health check에 포함하지 않는다.

Replace 전 immutable deployment preimage를 기록한다. Preimage에는 service별 current/target image
digest, rendered Compose digest, config/activation/lane/autostart/automation/LIVE approval/protection
digest, environment key set, volume/mount identity와 mode, current data/journal schema version,
target image의 readable/writable schema range, rollback image의 current/post-replace schema
compatibility range와 baseline health가 포함된다. Mutable tag만 기록한 preimage는 유효하지 않다.

Compatibility gate는 첫 service를 중단하기 전에 target images가 current schema를 읽고 쓰며
rollback images도 replacement 뒤 가능한 schema를 읽을 수 있음을 증명해야 한다. Migration은
additive/backward-compatible 범위만 허용한다. 어느 범위라도 불명확하면 replacement는 0건이다.

### 5. Replacement와 rollback은 bounded subset transaction이다

Deployment manifest는 service 순서와 service별 health timeout을 동결하며 한 번에 한 service만
교체한다. 각 timeout은 양수이고 5분을 초과할 수 없다. 새 service가 schema/API/Unix health와
preimage state digest를 통과해야 다음 service를 교체한다. 실패하면 아직 교체하지 않은 service는
건드리지 않는다.

Rollback은 실제로 교체된 subset만 exact preimage image digest로 역순 복구하고 config,
approval, journal, volumes/mounts와 broker-resident protection을 변경하지 않는다. Partial rollback
각 단계에도 같은 bounded health gate를 적용한다. Old image가 current schema와 호환되지 않으면
강제 downgrade하지 않고 해당 new service를 유지하며 entry를 OFF로 latch하고 safety services와
read-only health를 계속한 뒤 typed `ROLLBACK_INCOMPATIBLE` recovery를 요구한다. Blanket Compose
down/up 또는 schema downgrade는 허용하지 않는다.

## Risks / Trade-offs

- [한 market unavailable이 전체 payload를 실패시킴] → market별 status/error envelope로 부분
  snapshot을 반환한다.
- [Console/API descriptor drift] → shared registry golden contract와 schema test를 사용한다.
- [부분체결/분할청산 이중 집계] → deduplicated fill delta, authoritative cost policy와 quantity/PnL
  conservation assertion을 사용한다.
- [FX/fee 근거 누락] → 0이나 current rate로 보정하지 않고 `not_measured`로 분리한다.
- [부분 배포 뒤 health 실패] → 미교체 service는 보존하고 교체 subset만 digest-pinned 역순 rollback한다.
- [Rollback binary가 current schema를 못 읽음] → compatibility gate에서 사전 중단하거나 new
  service+entry OFF를 유지하며 destructive downgrade를 금지한다.

## Migration Plan

1. Shared market-keyed projection과 read-only adapters를 backward-compatible nullable field로
   배포한다.
2. Console/API contract, exact readiness enum과 partial-unavailable behavior를 fixture로 검증한다.
3. Performance projector에 campaign/leg composite lineage와 fill-delta quantity/fee/FX/PnL
   conservation을 추가하고 기존 누락 row는 `link_missing`/`not_measured`로 유지한다.
4. Exact image/schema/config/volume preimage와 compatibility matrix를 동결한다. Gate 실패 시
   replacement는 0건이다.
5. Manifest 순서대로 한 service씩 bounded replace하고 매 단계 Console/API/Unix projection
   dormant health와 state digest를 검증한다. 운영 activation은 별도의 사람 승인 단계다.
6. 실패 시 교체 subset만 exact digest로 역순 rollback한다. Schema incompatibility면 new service,
   entry OFF와 safety continuity를 유지하고 typed recovery로 전환한다.

## Open Questions

없음. Runtime projection이 아직 생성되지 않은 설치는 두 시장을 typed unavailable로 표시하고
ProtectionReady는 `UNWIRED`로 fail-closed하며 readiness나 성과를 추정하지 않는다.
