# a041 · 익절·손절 기준선 계약 완성

## 0. 백로그 계층 추적

- **Initiative**: `INIT-TOS-002`
- **Epic**: `EPIC-TOS-003`
- **Feature**: `FEAT-TOS-009`
- **Story**: `STORY-TOS-a041`

## Why

현재 엔진에는 단조 보호선과 ladder가 있지만 진입가·최초 손절·현재 보호선·다음 익절·예상 수량을 하나의 권위 있는 결과로 설명하는 계약이 없다. 기준선 계산과 주문 판단을 먼저 통합해야 화면·브로커 보호·향후 최적화가 같은 사실을 사용할 수 있다.

## What Changes

- 포지션별 `ExitLineSnapshot` 계약을 도입한다.
- 기준선, high-water, rung, 다음 목표와 다음 보호선의 단조·우선순위를 고정한다.
- 1주 보유는 중간 부분익절 주문을 만들지 않되 단계와 보호선은 승격하고, 최종 익절 또는 보호선 이탈 시 1주 전량을 청산한다.
- 0수량 주문을 절대 생성하지 않는다.
- 자체 진입과 외부 편입이 같은 계산 계약을 사용한다.
- 세 공통 정책의 label, 설명, rung 기본값, 단위와 1주 예상 동작을 server-authoritative descriptor로 제공해 a050 `/optimization`이 임의 수치를 만들지 않게 한다.
- **비목표**: 원장 스키마 변경, 화면 자체 구현, 조건주문 생성, 전략 진입. 설정 화면은 a050이 이 change의 descriptor를 소비한다.

## Capabilities

### New Capabilities

- 없음.

### Modified Capabilities

- `exit-policy`: 익절·손절 기준선의 권위 snapshot과 1주 중간 익절 생략 불변식을 추가한다.

## Impact

- `internal/exitpolicy`, `internal/app/engine/exitloop.go`와 해당 단위·분기 테스트.
- High-risk exit 계산 변경이므로 Function Logic Map과 적대적 Eng 리뷰가 필요하다.
