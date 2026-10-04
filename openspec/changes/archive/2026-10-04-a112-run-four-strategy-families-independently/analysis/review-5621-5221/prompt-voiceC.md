# 목소리 C — census · 표면 · 경계의 완전성

공통 브리프(`prompt-common.md`)를 먼저 읽는다. 너의 축은 **경계를 지키는 세기(census)와 표면 고정이 새 문을 빠짐없이 보는가**다.

필수 판정:

1. **census 소스 유도의 완전성.** `strategy_dispatch_handoff_guard_test.go` 의 `strategyHandoffAdmitDoors` 는 strategyhandoff 생산 소스에서
   「수신자 없는 공개 함수 중 결과 타입 문자열에 `Handoff` 가 나오는 것」을 경계 문으로 유도하고, `admitSites` 가 엔진 생산 파일에서 그 식별자를 센다.
   반증하라: 경계 값을 **만드는** 다른 방법이 이 유도에 안 걸리는가 — 결과가 `Handoff` 를 품은 다른 타입(구조체 · 맵 · 채널 · 함수 타입 · 별칭 · 제네릭),
   공개 변수(`var X = Admit`), 수신자 있는 공개 메서드가 새 Handoff 를 돌려주는 경우, 결과 없이 포인터 인자로 채우는 경우, `Handoff` 가 아닌 이름의 타입 별칭.
   각각이 (i) strategyhandoff 의 동결 표면 표(`escape_test.go` `exportedSurface`)에서 먼저 막히는지, (ii) 막히지 않고 census 도 못 보는지 — (ii) 가 있으면 P1.
   사본에서 실제로 그 모양을 추가해 두 스위트(무태그 · 태그)가 빨개지는지 재라.
   또 `admitSites` 가 엔진에서 문 식별자를 세는 범위(선언 전체 · 테스트 파일 제외)가 우회되는가: 다른 패키지를 한 다리 거쳐 `AdmitEachOwnerScope` 를 부르고 결과를 엔진에 넘기기
   (`TestOnlyTheEngineImportsThisSeam` 이 막는가), 엔진 안 dot import, 함수 값 반환.
2. **동결 표면 · import 폐포.** `AdmitEachOwnerScope` 선언 문자열이 정확한가, 비공개 `ownerScope` 가 표면 · 필드 검사에 새지 않는가, `strings` 추가가 전이 폐포 검사
   (`dependency_closure_test.go` 의 금지 전이 목록)와 충돌 없는가.
3. **소유자 범위의 정의.** `ownerScopeOf` 의 네 축이 `strategyrouter.OwnerKey`(`internal/strategyrouter/router.go`)와 같은가 — 정규화 차이(OwnerKey 는 market 을 검증,
   `ownerScopeOf` 는 `string(Market)` 그대로), Lineage 필드가 봉인 과정에서 이미 정규화되는지. 두 정의가 갈라질 때(한쪽만 고쳐질 때) 잡는 시험이 있는가 — 없으면 P2 로 적고 핀 모양을 제안.
   골든 `analysis/goldens/four-family-runtime-v1.json` queue 블록의 「owner scope」 정의와 대조.
4. **다른 경계 census 의 정합.** `TestExactlyOneProductionCallSiteTurnsAHandoffIntoADispatch` · `TestNoProductionSiteDiscardsTheSeamsAdmissionAnswer` ·
   `singleProposalAssumptionCensus` 가 새 코드(`dispatchHandoffs` 의 `entries` 순회, `deliverEachStrategyHandoff` 의 `Deliver`)를 올바르게 세는가 — 특히
   `dispatchHandoffs` 안의 `for _, entry := range authority.entries` 가 「단일 제안 가정」 census 나 「entries 읽기」 금지와 충돌하지 않는지, 그리고 그 census 들이
   새 함수를 **보지 못해서** 통과한 것은 아닌지.
