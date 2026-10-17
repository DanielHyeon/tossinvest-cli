# Story Mapping

- Story: `STORY-TOS-a144`
- Feature: `FEAT-TOS-003`

## Why

현재 M0 trigger-only 검증은 기본 생략된 `idempotency-ttl-edge`까지 PASS를 요구한다. TTL 관측은 trigger·child 주문·체결 귀속의 입력이 아니므로, 별도 두 번째 라이브 주문을 요구하는 TTL 측정이 단일 1주 trigger 측정을 불필요하게 차단한다.

## What Changes

- M0 trigger-only의 사전 PASS 집합에서 `idempotency-ttl-edge`만 제외한다.
- TTL step의 선택 실행·기록·reporting은 유지한다. PASS, skipped, failed, refused, missing verdict를 변환하거나 합산 PASS로 만들지 않는다.
- TTL 외 모든 procedure step의 최신 PASS, strict trigger-only redo, receipt, per-mutation confirmation, outstanding/M0 checkpoint HOLD, engine exclusion을 유지한다.
- 이번 변경은 설계·검증 계약만 정의한다. 구현과 라이브 검증은 후속 task다.

## Capabilities

### New Capabilities
- 없음.

### Modified Capabilities
- `order-execution`: user-supervised conditional-trigger redo의 prior-record 조건에 TTL-edge 단일 예외를 추가한다.

## Impact

예정 구현 범위는 `internal/verifylive/m0_recovery.go`의 M0 prerequisite 판정과 해당 CLI/direct-Runner tests다. 새 설정, framework, generic dependency policy, broker API, engine policy, orderscope는 추가·변경하지 않는다. A100의 기존 all-PASS 문구는 이 후속 계약 범위에서만 TTL 예외로 대체되며 역사 문서는 수정하지 않는다.
