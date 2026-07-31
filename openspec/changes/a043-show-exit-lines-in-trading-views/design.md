## Context

positions는 baseline과 initial stop 일부를 표시하고 orders는 broker/journal 주문을 나눠 표시한다. 두 화면 모두 a041/a042 snapshot을 각자 재해석하지 않고 read model로 받아야 한다.

## Goals / Non-Goals

**Goals:** 운영자가 데스크톱·모바일에서 현재 보호와 다음 동작을 한 번에 읽게 한다.

**Non-Goals:** 계산, journal mutation, 정책 설정, 주문 버튼 추가.

## Decisions

1. console/httpapi에 의존하지 않는 `internal/operatorview`가 snapshot을 canonical `ExitLineView`로 한 번
   변환하고 console과 이후 a051 adapter가 공유한다. transport별 DTO가 수치·상태를 재해석하지 않는다.
2. 1주는 `중간 매도 없음 · 보호선 승격`을 명시하며 projected quantity 0을 수량 0 주문처럼 표시하지 않는다.
3. orders는 exit decision ID로 snapshot trigger를 연결하고 결정적 연결이 없으면 `근거 미연결`로 표시한다.
4. CSP inline handler 없이 server-rendered semantic table/card를 사용한다.
5. `/positions`와 `/orders`는 결과를 설명하는 read-only 화면이다. 설정은 a050의 카테고리 화면 한 곳에서만 변경하고 거래 화면에는 문맥형 deep link만 둔다.

## Risks / Trade-offs

- [정보량 증가로 다시 복잡해짐] → 기본 행은 현재/다음 핵심값만, provenance는 details로 접는다.
- [stale snapshot 오인] → 평가 시각과 stale reason을 같은 시각 계층에 표시한다.
- [거래 화면과 최적화 화면에 설정이 중복됨] → 거래 화면의 입력 control을 금지하고 canonical category deep link를 사용한다.

## Migration Plan

read model과 fixture를 추가한 뒤 positions, orders 순서로 template를 전환한다. rollback은 template/read adapter만 되돌리며 엔진에는 영향이 없다.

## Open Questions

없음.
