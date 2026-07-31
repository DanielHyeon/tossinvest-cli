# Review: a047-add-strategy-engine

- Date: 2026-07-31
- Voices: Security, Test Architecture, Maintainability

## Security decision

pure lane/orchestrator와 immutable activation-manifest 기반 default-OFF 구현만 승인한다. exact StockOS source policy/market/constants golden fixture와 a045/a046 선행 gate 없이는 exposure-raising dispatch를 연결하지 않는다.

## Findings and decisions

1. manifest는 account/build/lane/threshold/settings/attestation/Guardian/reconciliation/protection/policy/scheduler와 개별 human approvals를 하나의 digest로 묶는다.
2. dispatch 직전에 모든 field를 재검증하며 mismatch/expiry/kill/reconcile degradation/high-risk config change는 effective entry를 OFF로 만든다. exit는 유지한다.
3. identity와 lineage는 candidate life, lane, threshold, settings, manifest digest를 끝까지 전달한다.
4. lane package는 broker/journal/config를 import하지 않는 pure domain이다.

## Verification evidence

- OpenSpec strict validation: pass.
- Exact source lane and external prerequisites: not yet frozen; activation remains blocked.

## Verdict

구조와 dormant/default-OFF scaffold는 승인한다. 첫 lane provenance 동결 전 runtime implementation/gate 완료를 주장하지 않는다.
