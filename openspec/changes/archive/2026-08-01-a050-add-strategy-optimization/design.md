## Context

공통 exit policy 선택 UI는 있지만 versioned parameter lifecycle과 성과 evidence 연결이 없다.

## Goals / Non-Goals

**Goals:** snapshot→candidate→preview→apply→history→rollback의 감사 가능한 설정 수명주기와 사용자가 설정 범위·기본값·효과를 혼동하지 않는 카테고리 UI를 만든다.

**Non-Goals:** 자동 LIVE 토글, 자동 lane 승격, paper/shadow/canary 단계, 근거 없는 최적값 생성.

## Decisions

1. a041의 transport-neutral `internal/settingmeta` contract를 사용해 parameter registry가 type/control/options/market/owner/apply timing/provenance와 safety direction을 선언한다. 각 domain change가 descriptor 값과 behavior를 소유하고 a050은 category composition과 lifecycle만 소유한다.
2. candidate setting은 base version과 a049 evidence digest를 참조하며 immutable하다.
3. apply/rollback은 동일 CAS command의 정방향/역방향이며 actor/reason/before/after를 기록한다.
4. active position policy snapshot은 설정 apply로 재해석하지 않는다.
5. 추천 불가 상태를 first-class로 노출한다. a049 evidence 부족은 자동 추천과 evidence-backed candidate 생성을 막지만, registry와 safety validation을 통과한 보수적 server preset의 사람 선택 자체를 막지 않는다.
6. `/optimization`의 정본 정보구조와 필드 표현은 `ui-design.md`를 따른다. 각 parameter는 소유 change가 descriptor/default/range/help를 제공하고 a050은 임의 default를 만들지 않는다.
7. 한 카테고리 저장은 그 카테고리의 changed subset만 전송한다. 미저장 다른 카테고리 draft, LIVE/lane/automation state를 묶지 않는다.
8. UI는 server registry의 stable option ID만 선택한다. text/textarea/number/contenteditable,
   임의 symbol 입력과 typed confirmation phrase를 금지한다. decimal/integer도 registry가 제공하는
   discrete choice 또는 bounded step option으로만 바꾼다.
9. StockOS 최신 lane-console의 화면 단위 navigation, 상태·파라미터·실행품질·성과 구획,
   changed-key-only 저장, 저장 후 effective mismatch 확인을 재사용한다. 구형 긴 slider matrix와
   전역 전체 저장은 복제하지 않는다.
10. high-risk 확인은 before/after·영향 범위·3초 대기·명시 checkbox/button을 사용한다. audit
    reason은 자유 입력이 아니라 server-defined reason code/chip으로 기록한다.
11. `internal/console`과 `internal/httpapi`는 journal read-only 경계를 유지한다. write는 두 transport가
    공유하는 좁은 `OptimizationCommander`를 통해 canonical control-plane owner에서만 실행한다. command는
    durable/idempotent이고 engine이 generation/version/approval/Guardian/broker state를 재검증한 뒤 journal의
    유일한 trading-state writer로 남는다. CAS/audit 실패는 transaction 전체를 중단한다.
12. high-risk key apply나 rollback은 desired state를 보존할 수 있지만 immutable activation manifest digest가
    바뀌면 effective entry는 OFF가 된다. rollback은 과거 row를 다시 쓰지 않고 current registry/capability/
    protection monotonicity를 재검증한 새 version이며 새 manifest 승인 전 LIVE authority를 복원하지 않는다.
13. `ui-design.md`의 표와 수치는 owner descriptor에서 생성되는 예시일 뿐 별도 default source가 아니다.

## Risks / Trade-offs

- [설정과 engine effective 값 drift] → desired/effective version을 함께 표시하고 restart 필요 여부를 명시한다.
- [위험 확대 설정] → 별도 high-risk 승인과 정책 상한을 요구한다.
- [설정이 많아져 혼란] → 여섯 고정 카테고리, 기본/현재/effective 동시 표시, sticky 변경 bar와 category-scoped save를 사용한다.
- [임의 입력으로 오타·범위 우회] → registry option ID만 전송하고 자유 입력 control 부재를 정적·렌더링 테스트로 고정한다.

## Migration Plan

a041~a049 owner descriptor를 조합한 read-only registry/snapshot/history를 먼저 배포하고 같은 change 내에서 CAS write를 연결한다. 최종 배포는 직접 production 경로지만 LIVE lane는 자동 활성화하지 않는다.

## Open Questions

자동 추천 알고리즘은 범위 밖이며 첫 버전은 사람이 만든 candidate 설정만 다룬다.
