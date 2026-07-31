## Context

공통 policy ID는 새 관리 포지션에 snapshot되지만 종목별 override·release의 명시적 CAS 계약이 없다. 설정과 실제 exit lifecycle을 분리해 기존 포지션을 조용히 재해석하지 않아야 한다.

## Goals / Non-Goals

**Goals:** 종목별 preview/apply/release/re-adopt를 versioned audit와 함께 제공한다.

**Non-Goals:** 최적화 추천, LIVE 토글, 자동 신규 매수, broker protection 구현.

## Decisions

1. override는 account+market+symbol+adoption generation에 귀속된 별도 record로 둔다.
2. 모든 write는 `If-Match`와 before/after audit를 사용한다. 전역 설정 파일을 종목별로 확장하지 않는다.
3. release는 exit state를 명시적으로 닫되 미체결 reduce-only 주문·긴급 청산 상태가 있으면 거부한다.
4. re-adopt는 새 generation과 새 t0를 만들며 과거 high-water/rung을 재사용하지 않는다.
5. 기존 포지션 일괄 rebind는 별도 명시 작업 없이는 금지한다.
6. UI 소유 카테고리는 `position-management`다. `종목별 정책`과 `외부 매수 자동편입`을 하위 구획으로 분리하고 각 control에 쉬운 설명, 기본값, 현재 desired, effective, 적용 대상과 적용 시점을 표시한다.
7. 외부 편입의 authoritative defaults는 enabled OFF, default stop 5%, allowed 2~20%/step 0.5%, include/exclude empty다. exclude 우선과 `1주는 중간 익절 없이 보호선만 승격` 규칙을 control 가까이에 설명한다.
8. `internal/console`은 journal read-only 불변을 유지한다. write handler는 cmd/runtime가 주입한 좁은
   `PositionPolicyCommander`만 호출하고 command service가 writable journal transaction, CAS, audit와
   serialization을 단독 소유한다. console이나 template에 `journal.Open`/SQL/write method를 노출하지 않는다.

## Risks / Trade-offs

- [release가 보호 공백 생성] → active exit/protection 확인, 3초 대기와 명시 checkbox/button을 요구하되 typed phrase 입력은 금지한다.
- [동시 콘솔 쓰기] → CAS 412와 현재 version 반환으로 해결한다.
- [공통 정책·종목 override·자동편입 혼동] → 한 카테고리 안에서 별도 section과 범위 badge를 사용하고 저장 preview에 영향받는 종목/generation을 명시한다.

## Migration Plan

nullable override/adoption generation schema를 추가하고 read-only preview부터 연결한 뒤 write routes를 추가한다. rollback 시 기존 snapshot은 유지하고 새 override 적용만 중지한다.

## Open Questions

없음.
