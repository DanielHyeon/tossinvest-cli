# a044 · 종목별 익절·손절 정책 수명주기 관리

## 0. 백로그 계층 추적

- **Initiative**: `INIT-TOS-002`
- **Epic**: `EPIC-TOS-003`
- **Feature**: `FEAT-TOS-009`
- **Story**: `STORY-TOS-a044`

## Why

현재 공통 정책은 신규 관리 포지션에만 적용되고 운영자가 한 종목을 override·release·재편입하는 계약이 부족하다. 외부 매수 자동관리와 기존 포지션 불변성을 함께 유지하려면 종목별 정책 수명주기를 버전·감사와 함께 제공해야 한다.

## What Changes

- 종목별 정책 preview, 승인, release, re-adopt와 충돌 resolve를 정의한다.
- version-bound opaque preview token 기반 console CAS(HTTP API는 `If-Match`)와 before/after/actor/server-defined reason 감사를 요구한다.
- 공통 정책 변경은 기존 포지션을 자동 rebind하지 않는다.
- release/re-adopt는 과거 high-water·rung을 재사용하지 않는다.
- 정책 변경은 LIVE/automation toggle을 바꾸지 않는다.
- a050의 `position-management` 카테고리에 종목별 관리와 외부 매수 자동편입을 함께 배치하고, 계좌·시장·종목·generation 단위의 현재값·기본값·실효값·설명을 제공한다.
- 외부 매수 자동편입 기본값은 OFF, 합성 손절폭은 5%(허용 2~20%, 0.5% 단위), include/exclude 목록은 비어 있음이며 exclude가 우선한다. 종목별 정책 기본값은 공통 정책 상속이다.
- UI는 server-defined preset과 현재 보유 행 action만 사용하며 자유 텍스트·숫자·symbol·reason 입력과 typed confirmation을 제공하지 않는다.
- **비목표**: 정책 수치 자동 최적화, 신규 매수, 브로커 조건주문.

## Capabilities

### New Capabilities

- `position-exit-policy-management`: 종목별 정책 override와 관리 수명주기 계약.

### Modified Capabilities

- `exit-policy`: 포지션별 정책 snapshot과 rebind 금지 규칙을 확장한다.
- `position-ledger`: adoption release/re-adopt의 새 lifecycle 규칙을 추가한다.
- `operator-console`: 정책 관리 표면과 CAS 오류 표현을 추가한다.

## Impact

- `internal/config`, `internal/journal`, `internal/app/engine`, `internal/console`.
- High-risk 정책·원장 변경이며 기존 포지션과 emergency exit 회귀 테스트가 필수다.
