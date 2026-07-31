## Context

engine runtime에는 entry loop가 없고 Tracer는 production에서 호출되지 않는다. a046 verdict를 a045 보호 전제 아래 공식 주문 경로에 연결하는 최소 lane architecture가 필요하다.

## Goals / Non-Goals

**Goals:** 독립 lane, durable decision provenance, Guardian-only LIVE 진입을 구현한다.

**Non-Goals:** 다수 전략 동시 이식, scheduler, 자동 최적화, paper/shadow/canary 경로.

## Decisions

1. lane는 순수 `Evaluate(ApprovedCandidate) EntryDecision`을 반환하고 broker/journal을 직접 받지 않는다.
2. orchestrator만 RiskIntent→Guardian→durable attempt→official gateway를 수행한다.
3. 첫 lane는 StockOS에서 검증된 하나의 정책만 이식하고 모든 상수에 출처·시장·검증 상태를 붙인다.
4. lane desired state와 effective state를 분리한다. OFF는 신규 entry만 차단하고 exit loop에 전달되지 않는다.
5. 자동 재시작은 운영자가 승인한 desired state와 유효 gate/protection을 모두 재확인한다.
6. UI 소유 카테고리는 `strategy-runtime`이다. `전략 파라미터`, `lane 상태`, `자동 기동`, `LIVE 주문 승인`을 별도 section과 별도 save/confirm action으로 분리하며 `한 번에 모두 활성화` control을 금지한다.
7. lane desired 기본값과 auto-start 기본값은 OFF다. 첫 lane source/market/constants가 동결되기 전 descriptor는 `not_configured`, read-only이며 숫자 0이나 임의 StockOS 값을 UI default로 넣지 않는다.
8. 모든 설정 descriptor는 label/help/default/desired/effective/unit/range/provenance/apply timing과 effective refusal reason을 포함한다.
9. 신규 진입 권한은 immutable activation manifest 하나로만 표현한다. manifest는 account/profile,
   build/commit digest, lane ID/version/source/market/constants digest, threshold/evidence digest,
   settings digest, capability-attestation digest/expiry, Guardian version/limits, reconciliation watermark,
   protection profile, operating policy, scheduler/calendar scope, lane/scheduler/autostart/gate/LIVE의 개별
   승인, actor/issued-at/expires-at/generation/audit ID를 묶는다.
10. orchestrator는 durable dispatch 직전에 manifest 전 필드를 다시 검증하고 decision/attempt에 digest를
    기록한다. mismatch, expiry, kill switch, reconcile degradation 또는 high-risk config 변경은 entry만
    OFF로 만들며 reduce-only exit/reconcile에는 전달하지 않는다.

## Risks / Trade-offs

- [전략이 gateway를 우회] → package dependency/static test로 lane의 broker import를 금지한다.
- [재시작 중 중복 진입] → deterministic decision/client order identity와 duplicate Guardian check를 사용한다.
- [직접 LIVE 위험] → 기본 OFF, 전체 gate, 단일 운영 승인과 kill switch를 유지한다.
- [서로 다른 시점의 승인이 조합됨] → 원자적 activation manifest와 submit-time TOCTOU 재검증을 사용한다.
- [전략값 저장과 엔진/LIVE 활성화 혼동] → 독립 section·버전·확인 흐름을 사용하고 설정 저장이 운영 토글을 바꾸지 않음을 preview에 고정한다.

## Migration Plan

lane interface와 dormant runtime을 배포하고 전체 gate 후 운영자가 LIVE를 한 번 승인한다. canary는 없다. rollback은 entry desired state OFF 후 exit/reconcile을 계속 실행한다.

## Open Questions

첫 lane의 정확한 StockOS source policy와 적용 시장은 proposal-freeze에서 하나로 고정한다.
