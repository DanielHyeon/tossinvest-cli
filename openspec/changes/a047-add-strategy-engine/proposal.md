# a047 · 전략 진입 엔진 추가

## 0. 백로그 계층 추적

- **Initiative**: `INIT-TOS-002`
- **Epic**: `EPIC-TOS-004`
- **Feature**: `FEAT-TOS-010`
- **Story**: `STORY-TOS-a047`

## Why

현재 production runtime은 reconcile·exit·filldetect만 실행하고 후보 verdict를 매수 결정으로 소비하지 않는다. 승인 후보를 Guardian과 공식 Open API 주문 경로에만 연결하는 독립 전략 레인 계약이 필요하다.

## What Changes

- `EntryLane`과 `ApprovedCandidate → EntryDecision → RiskIntent` 계약을 추가한다.
- 모든 진입은 Guardian, operating mode, protection readiness, durable journal과 official gateway를 통과한다.
- lane OFF는 신규 진입만 차단하고 기존 포지션 exit는 계속한다.
- 결정·주문·체결 provenance에 lane와 candidate 식별자를 저장한다.
- 계좌·build·lane·threshold·settings·capability attestation·Guardian·reconciliation·protection·scheduler와 사람 승인 상태를 하나의 immutable activation manifest로 묶고 모든 신규 진입 직전에 재검증한다.
- paper/shadow/canary 주문 경로 없이 최종 LIVE 경로만 구현하되 기본 운영 토글은 OFF다.
- a050의 `strategy-runtime` 카테고리에 lane 설정, desired/effective 상태, 자동 기동과 LIVE 승인을 서로 분리해 제공하고 각 필드의 출처·기본값·설명·적용 시점을 표시한다.
- lane와 자동 기동의 기본값은 OFF다. 첫 lane 정책·시장·상수가 proposal-freeze에서 확정되기 전에는 임의 숫자 기본값을 만들지 않고 `미구성 / 읽기 전용`으로 표시한다.

## Capabilities

### New Capabilities

- `strategy-engine`: 독립 진입 레인, 후보 소비, provenance와 LIVE 승인 계약.

### Modified Capabilities

- `engine-safety`: 레인 OFF·Guardian·protection readiness 진입 차단을 추가한다.
- `order-execution`: 전략 진입 intent의 durable submit 계약을 추가한다.
- `operator-console`: 전략 lane 설정과 운영 승인 분리 표면을 추가한다.

## Impact

- 신규 `internal/strategy`, engine runtime, risk/order/journal integration과 console operating controls.
- LIVE 진입 경로이므로 a045 완료와 적대적 Eng/security review가 선행된다.
- activation manifest가 없거나 하나의 digest/version/expiry라도 불일치하면 effective entry는 OFF다.
