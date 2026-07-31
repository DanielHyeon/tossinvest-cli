# a042 · 익절·손절 기준선 snapshot 영속화

## 0. 백로그 계층 추적

- **Initiative**: `INIT-TOS-002`
- **Epic**: `EPIC-TOS-003`
- **Feature**: `FEAT-TOS-009`
- **Story**: `STORY-TOS-a042`

## Why

a041의 권위 기준선이 프로세스 수명에만 머물면 재시작 때 더 낮은 보호선으로 재해석될 수 있다. 보호 상태와 정책 provenance를 journal에 원자적으로 보존해 crash 뒤에도 같은 기준을 회수해야 한다.

## What Changes

- exit snapshot에 정책 버전, active rung, high-water, 현재·다음 기준선, 마지막 관측을 영속한다.
- additive-nullable migration과 rollback·백업 계약을 정의한다.
- 재시작은 저장된 보호선 이상으로만 상태를 전진시킨다.
- 손상·누락 상태는 해당 포지션을 fail-closed로 격리하되 다른 긴급 청산은 지연하지 않는다.
- a050 최적화 화면이 저장 snapshot과 현재 재계산값을 혼동하지 않도록 version, source, observed-at, stale/unknown reason을 가진 read model을 제공한다.
- **비목표**: UI 자체 구현, 정책 변경 API, 브로커 조건주문.

## Capabilities

### New Capabilities

- 없음.

### Modified Capabilities

- `position-ledger`: 익절·손절 기준선 snapshot과 crash recovery provenance를 추가한다.
- `exit-policy`: 저장된 snapshot을 단조적으로 복구하는 규칙을 추가한다.

## Impact

- `internal/journal`, exit state migration, engine recovery와 crash/fault-injection 테스트.
- 원장 스키마 변경이므로 rollback 및 독립 검증이 필수다.
