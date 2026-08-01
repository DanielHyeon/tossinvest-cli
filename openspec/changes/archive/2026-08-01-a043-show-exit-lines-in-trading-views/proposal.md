# a043 · 거래 화면 익절·손절 기준선 표시

## 0. 백로그 계층 추적

- **Initiative**: `INIT-TOS-002`
- **Epic**: `EPIC-TOS-003`
- **Feature**: `FEAT-TOS-009`
- **Story**: `STORY-TOS-a043`

## Why

운영자는 `/positions`와 `/orders`에서 현재 보호선 일부만 볼 수 있어 다음 익절 목표와 실제 예상 동작을 추론해야 한다. a041/a042의 동일 snapshot을 두 화면에 노출해 판단과 표시의 불일치를 제거한다.

## What Changes

- `/positions`에 최초 손절, 현재 보호선, 다음 익절, 다음 보호선, rung, 예상 수량과 평가 시각을 표시한다.
- 1주 보유는 중간 매도 없음과 보호선 승격을 명확히 표시한다.
- `/orders`의 exit 주문에 발생 기준선, 관측가, 정책·rung·사유를 연결한다.
- 누락·stale 값은 `—`와 사유로 표시하고 0으로 꾸미지 않는다.
- 두 거래 화면은 설정 입력을 복제하지 않고 관련 문맥에서 `/optimization?category=exit-protection` 또는 `/optimization?category=position-management`로 이동하는 설명 링크를 제공한다.
- **비목표**: 정책 계산·저장 변경, 주문 제출, 설정 쓰기.

## Capabilities

### New Capabilities

- 없음.

### Modified Capabilities

- `operator-console`: 거래 화면이 권위 exit snapshot을 일관되게 표현하도록 요구한다.

## Impact

- `internal/console/portfolio.go`, `orders.go`, templates와 반응형/CSP 테스트.
- read-only UI change이며 broker 호출 예산과 engine 상태를 변경하지 않는다.
