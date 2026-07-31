## 1. 선행 증거와 설계 동결

- [ ] 1.1 `verify-execution-capability`·`verify-observes-the-trigger` attestation에서 account/profile/market/type/session/trigger/quantity/persistence/reservation/idempotency/atomic-replace/tool-build/evidence digest를 strict versioned matrix로 고정한다.
- [ ] 1.2 conditional client, gateway, reservations, reconciliation, flatten의 CodeGraph impact와 모든 high-risk Function Logic·Branch Test Map을 작성한다.
- [ ] 1.3 saga·crash window·oversell·1주 보호·replace·orphan·2초 flatten deadline·initial fill 1초 arm/2초 ACTIVE RED 테스트를 추가한다.

## 2. Durable protection saga

- [ ] 2.1 additive schema, canonical conditional body와 saga repository를 구현한다.
- [ ] 2.2 create/replace/cancel/trigger recovery를 official-only gateway에 구현한다.
- [ ] 2.3 fill quantity와 하나의 sell claim으로 수렴하는 reservation lifecycle을 구현한다.

## 3. Engine과 reconciliation

- [ ] 3.1 a041 보호선을 broker trigger에 연결하고 단방향 replace를 구현한다.
- [ ] 3.2 orphan/missing/duplicate/mismatch reconciliation과 flatten cancel-first를 구현한다.
- [ ] 3.3 attested profile만 `ProtectionReady=WIRED`로 평가하고 미지원 entry는 차단한다.

## 4. 검증과 직접 LIVE 인계

- [ ] 4.1 `exit-protection`에 capability/activation 기본 OFF, unavailable reason, effective trigger·qty·broker ID·updated-at·reconcile reason과 쉬운 설명을 구현한다.
- [ ] 4.2 protection 약화 before/after·영향범위·3초 지연 확인, 강화/약화 분류와 모바일·접근성·CSP RED 테스트를 통과한다.
- [ ] 4.3 fault-injection/race/restart/full test·vet·validate와 security/adversarial review를 통과한다.
- [ ] 4.4 paper/shadow/canary 경로가 없고 기본 운영 토글 OFF임을 정적으로 검증한다.
- [ ] 4.5 `make gate CHANGE=a045-add-protection-orders` 후 운영자의 별도 LIVE 승인 절차를 문서화한다.
