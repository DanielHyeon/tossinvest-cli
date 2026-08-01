## Context

공통 policy ID는 새 관리 포지션에 snapshot되지만 종목별 override·release의 명시적 CAS 계약이 없다. 설정과 실제 exit lifecycle을 분리해 기존 포지션을 조용히 재해석하지 않아야 한다.

## Goals / Non-Goals

**Goals:** 종목별 preview/apply/release/re-adopt를 versioned audit와 함께 제공한다.

**Non-Goals:** 최적화 추천, LIVE 토글, 자동 신규 매수, broker protection 구현.

## Decisions

1. override는 account+market+symbol+adoption generation에 귀속된 별도 record로 둔다. 기존
   `exit_states.position_id` 수명을 재정의하지 않고 `(position_id, adoption_generation)`을 key로 갖는
   lifecycle record와 generation별 event를 추가한다.
2. 모든 write는 CAS와 before/after audit를 사용한다. 전역 설정 파일을 종목별로 확장하지 않는다.
   server-rendered console은 JS로 `If-Match` header를 만들지 않고 current version·scope·action·expiry가
   서명된 opaque preview token을 CSRF와 함께 제출한다. 서버는 token을 검증한 뒤 같은 version CAS를
   적용하며 stale이면 412를 반환한다. a051 HTTP API adapter만 표준 `If-Match` header를 사용할 수 있다.
3. release는 exit state를 명시적으로 닫되 미체결 reduce-only 주문·긴급 청산 상태가 있으면 거부한다.
4. re-adopt는 새 generation과 새 t0를 만들며 과거 high-water/rung을 재사용하지 않는다.
5. 기존 포지션 일괄 rebind는 별도 명시 작업 없이는 금지한다.
6. UI 소유 카테고리는 `position-management`다. `종목별 정책`과 `외부 매수 자동편입`을 하위 구획으로 분리하고 각 control에 쉬운 설명, 기본값, 현재 desired, effective, 적용 대상과 적용 시점을 표시한다.
7. 외부 편입의 authoritative defaults는 enabled OFF, default stop 5%, allowed 2~20%/step 0.5%, include/exclude empty다. exclude 우선과 `1주는 중간 익절 없이 보호선만 승격` 규칙을 control 가까이에 설명한다.
8. `internal/console`은 journal read-only 불변을 유지한다. write handler는 cmd/runtime가 주입한 좁은
   `PositionPolicyCommander`만 호출하고 command service가 writable journal transaction, CAS, audit와
   serialization을 단독 소유한다. console이나 template에 `journal.Open`/SQL/write method를 노출하지 않는다.
9. 콘솔은 arbitrary text/number/textarea/contenteditable, 임의 symbol·reason 입력을 제공하지 않는다.
   policy는 server registry preset/inherit, auto-adoption은 OFF/ON, stop은 2~20%의 0.5% option ID,
   include/exclude는 현재 server-rendered 보유 행 action, reason은 server enum chip만 사용한다. hidden
   field는 CSRF와 opaque preview/action token만 허용한다.
10. release/re-adopt는 engine-owned local command transport를 통해서만 수행한다. console 프로세스가
    writable journal을 열거나 전역 설정 파일을 우회 변경하지 않으며 a050의 canonical control plane과
    동일 command service를 공유한다.

## Risks / Trade-offs

- [release가 보호 공백 생성] → active exit/protection 확인, 3초 대기와 명시 checkbox/button을 요구하되 typed phrase 입력은 금지한다.
- [동시 콘솔 쓰기] → version-bound opaque preview token, CAS 412와 현재 version 반환으로 해결한다.
- [공통 정책·종목 override·자동편입 혼동] → 한 카테고리 안에서 별도 section과 범위 badge를 사용하고 저장 preview에 영향받는 종목/generation을 명시한다.

## Migration Plan

nullable override/adoption generation schema를 추가하고 read-only preview부터 연결한 뒤 write routes를 추가한다. rollback 시 기존 snapshot은 유지하고 새 override 적용만 중지한다.

## Open Questions

없음.
