# Review: a041-complete-exit-line-contract

- Date: 2026-07-31
- Voices: Security, Test Architecture, Maintainability
- Scope: exit snapshot, policy identity, setting metadata foundation

## Findings and decisions

1. Stable policy ID만으로 mutable rung을 식별할 수 없다. ID/version/canonical digest와 deterministic snapshot/decision identity를 계약에 추가했다.
2. UI 계약이 a050에 늦게 생기면 upstream descriptor 소유권이 순환한다. transport-neutral `internal/settingmeta` 최소 계약은 a041이 소유하고 각 domain은 자기 값을 제공한다.
3. 1주 partial은 0수량 주문이 아니라 state-only promotion이어야 하고 final/breach만 정확히 1주를 청산한다.
4. 구현 전 기존 evaluator/exitloop의 Function Logic Map과 경계·race Branch Test Map이 필수다.

## Verification evidence

- OpenSpec strict validation: pass.
- Existing full Go suite: pass (proposal-freeze test review).
- LIVE/order authority change: none; defaults remain conservative.

## Verdict

계약 구현을 승인한다. default-OFF/zero-order 회귀, deterministic identity, same-snapshot consumer와 independent implementation review를 gate 조건으로 한다.
