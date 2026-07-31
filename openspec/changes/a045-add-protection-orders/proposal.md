# a045 · 브로커 상주 보호주문 추가

## 0. 백로그 계층 추적

- **Initiative**: `INIT-TOS-002`
- **Epic**: `EPIC-TOS-002`
- **Feature**: `FEAT-TOS-003`
- **Story**: `STORY-TOS-a045`

## Why

현재 보호선은 엔진 프로세스가 살아 있는 동안의 로컬 판단이고 `ProtectionReady`는 UNWIRED다. 자동 진입 전에 검증된 Toss 조건주문을 durable saga로 연결해 프로세스·네트워크 장애 중에도 손실 보호가 브로커에 남아야 한다.

## What Changes

- attested 시장에 SINGLE+MARKET 손절 중심의 broker-resident protection saga를 도입한다.
- 체결·부분체결·정정·발동·취소·재시작을 durable identifier로 연결한다.
- 한 심볼의 브로커측 매도 청구권을 하나로 제한하고 oversell을 차단한다.
- 보호선은 안전한 방향으로만 정정하며 flatten은 보호주문을 먼저 해제한다.
- attestation과 실제 capability가 일치하는 경우에만 `ProtectionReady=WIRED`다.
- paper/shadow/canary 주문 경로는 만들지 않는다. 실제 활성화는 전체 gate 후 운영자 승인으로만 한다.
- a050의 `exit-protection` 카테고리에 브로커 보호 준비상태, trigger·수량·broker ID·갱신시각·reconcile 사유를 읽기 쉽게 표시하고 보호 약화 변경은 별도 위험 확인을 거친다.
- attestation 전 기본 상태는 `OFF / 지원 확인 전 사용 불가`다. attestation이 SINGLE+MARKET을 확정해도 활성화 기본값은 OFF이며 미검증 주문 유형·수치를 UI 기본값으로 만들지 않는다.

## Capabilities

### New Capabilities

- `protection-orders`: 브로커 상주 손절의 상태기계, 수량 소유권과 복구 계약.

### Modified Capabilities

- `engine-safety`: protection readiness와 exposure-raising interlock을 확정한다.
- `order-execution`: 조건주문 mutation·idempotency·귀속 규칙을 추가한다.
- `reconciliation`: 브로커 보호주문과 로컬 saga의 불일치를 회수한다.
- `operator-console`: 보호주문 상태와 보호 약화 위험 확인 표면을 추가한다.

## Impact

- official Open API conditional endpoints, journal schema, engine gateway/recovery/reconciliation.
- 최상위 high-risk change이며 `verify-execution-capability`와 `verify-observes-the-trigger`의 attestation을 선행 입력으로 사용한다.
