**판정: APPROVE**

P0·P1은 없습니다. P2가 넷, 문서 정정(T)이 하나 있습니다. 네 P2 모두 오늘 생산 동작은 바꾸지 않습니다. 서명 매니페스트가 0건이라 모든 시장이 `dispatchHandoffs` 첫 갈래(:39–40)를 타고, 활성화된 시장에서도 하류 B2 개수 관문이 주문을 막기 때문입니다. 다만 P2-2는 5.2.2.2가 B2를 걷어 내기 **전에** 닫아야 합니다.

**실험 조건**
- 사본: `/tmp/claude-1000/a112-rev-87ll`(리뷰 트리 `a112-review-00e1b9bd` 전체 복사). 끝나고 지웠습니다(`ls` → No such file).
- 조작 없는 대조를 먼저 돌렸고 전부 GREEN이었습니다.
  - strategyhandoff: 무태그·태그 모두 ok.
  - 엔진: census 7개와 a112 소유자 범위 시험 6개가 무태그·태그 모두 PASS.
- 실험 하나를 끝낼 때마다 `diff -rq internal <리뷰트리>/internal` 결과가 "internal identical"임을 확인했습니다.
- 실제 저장소는 건드리지 않았습니다. 시작 때와 같은 untracked 3개뿐입니다. 도중에 HEAD가 `6699e4ce`(a090)로 바뀌었는데, 이는 병행 세션의 커밋입니다.

## 발견 표

| # | 등급 | 주장 | 증거 | 권고 |
|---|---|---|---|---|
| 1 | P2 | 문 유도(`strategyHandoffAdmitDoors`, guard_test.go:664–692)는 결과 타입 문자열에 `Handoff`가 들어간 **수신자 없는 func 선언**만 셉니다. 동결 표면 표에 새 이름을 정식으로 적어 넣으면, 아래 다섯 모양은 census가 전혀 보지 못합니다: 수신자 메서드, 포인터 인자로 채우기, `Handoff`가 아닌 이름의 별칭, 공개 var, 제네릭. 커밋 문구 "다음에 문이 늘어도 목록을 손으로 고칠 필요가 없고"는 과장입니다. 유도 함수 자체를 시험하는 fixture도 없습니다. `TestAdmitCensusSeesEverySpellingItClaims`(:764)는 문 목록을 손으로 넘깁니다. | **실험 A**: strategyhandoff에 다섯 모양을 추가했습니다(`Handoff.Split`, `AdmitInto(dst *[]Handoff,…)`, `type Batch = []Handoff`+`AdmitBatch`, `var AdmitVar = AdmitEachOwnerScope`, `AdmitAs[T]`). 무태그·태그 모두 `escape_test.go:136: … undeclared export AdmitAs … AdmitBatch … AdmitInto … AdmitVar … Batch … Handoff.Split`로 FAIL → 모두 (i) 단계에서 먼저 막힙니다. **실험 B**: 표에 여섯 줄을 추가하고 엔진 생산 파일 `zz_leak.go`에서 다섯 문을 모두 호출했습니다. strategyhandoff는 ok이고, 엔진 `TestExactlyOneProductionSiteAdmitsIntoTheSeam`·`TestTheSeamFilesStay…`·`TestNoProductionSiteDiscards…` 등이 무태그·태그 모두 PASS → 둘째 주조 자리가 보이지 않습니다. 부수 관찰: 표면 표의 서명 문자열은 타입 매개변수를 잃습니다(`AdmitAs` → `func(…) T`). | 유도(보이지 않으면 통과)를 허용 목록(모르면 실패)으로 바꿉니다. 엔진 생산 코드의 모든 `strategyhandoff.X` 선택자에 대해, 주조하지 않는 이름(`Handoff`·`Delivered`·`Refusal`과 상수·`Capacity`·`ErrNoDelivery`·`Delivered.Result`)의 명시 목록과 문 목록 `{Admit, AdmitEachOwnerScope}` 중 어디에도 없으면 실패하게 합니다. 이렇게 하지 않을 거라면 커밋·review 문구에서 "자동으로 새 문을 본다"를 빼고 fixture를 추가합니다. |
| 2 | P2 | 입장 census는 식별자 **언급 횟수**만 셉니다. 그래서 활성화된 시장의 소유자 범위당 상한(같은 범위 둘이면 시장 전체 OverCapacity)이 엔진 쪽에서 행동으로 고정되어 있지 않습니다. `dispatchHandoffs` 안에서 `AdmitEachOwnerScope`를 한 번만 언급하고 원소마다 따로 부르면 중복 검사(handoff.go:231–238)가 무력화되는데, 모든 스위트가 초록입니다. 변이 원장 F03은 strategyhandoff 단위 시험만 잡습니다. | **실험 C2**: strategy_dispatch_handoff.go:47을 `admit := strategyhandoff.AdmitEachOwnerScope; if !ready‖len==0 {return admit(…)}; for _, one := range selected { …admit(true, []Result{one})… }`로 바꿨습니다. 결과: strategyhandoff 무태그·태그 ok, 엔진 무태그·태그의 census 7개와 a112 시험 6개 전부 PASS. (준비 상태 갈래까지 쪼갠 첫 변이 C는 `TestOnlyAnActivatedMarket…`의 닫힘 갈래(:77)에 우연히 걸렸습니다.) **제안 핀을 실측했습니다**: 활성화된 KR 시장에서 첫 항목과 같은 계좌·종목·세대(8)로 봉인한 제안을 하나 더 싣고 `dispatchHandoffs()`가 `len==1 && OverCapacity`인지 단언합니다. C2 변이에서는 `2 handoffs (first ""), want one OverCapacity`로 FAIL, 변이 없이는 PASS였습니다. | 위 시험(`a112_owner_scope_handoff_test.go`, 태그 `tossos_testseams`)을 5.2.2.1에 추가합니다. 5.2.2.2가 B2를 걷어 내면 이 경로가 곧바로 주문으로 이어지므로, 5.2.2.2 착수 조건에 넣을 것을 권합니다. |
| 3 | P2 | `TestNoProductionSiteDiscardsTheSeamsAdmissionAnswer`(:838–863)의 인식기 `seamAdmissionAnswers`(:867–916)는 `X.dispatchHandoff().Single()` 한 철자만 인식합니다. 5.2.2.1이 만든 `dispatchHandoffs()[i].Single()`은 보지 못하고, 넷째 자리가 생겨도 `checked == 3`이 그대로 통과합니다. 편집 전부터 있던 중간 바인딩 철자(`h := …dispatchHandoff(); r, _ := h.Single()`)도 같은 방식으로 새어 나갑니다. 즉 기존 결함 부류에 새 철자가 하나 더 붙은 것입니다. | **실험 D**: 엔진 생산 파일에 `result, _ := authority.dispatchHandoffs()[0].Single()`를 넣었습니다. 무태그·태그 모두 해당 시험과 `TestExactlyOneProductionSiteAdmits…`·`TestTheSingleProposal…`·`TestTheProductionCycleDelivers…`가 PASS. **D′**(편집 전 철자, 중간 바인딩)도 PASS. | Single을 부르는 자리를 수신자 식 모양으로 거르지 말고, 엔진 생산 코드의 `.Single()` 선택자 **전부**를 세서 자리 목록(3개)과 대조합니다. 그리고 `_`로 버리는지 검사합니다. 영값 반환과 `ValidProposal` 재확인이 뒤를 받치므로 노출은 늘지 않아 P2입니다. |
| 4 | P2(낮음) | `ownerScopeOf`(handoff.go:252–268)와 `strategyrouter.OwnerKey`(router.go:13–40)를 묶는 시험이 없습니다. 한쪽 정규화만 바뀌거나 축 하나가 빠져도 초록입니다. 오늘 실질적 차이는 0입니다: 계보는 `validScope`(flow.go:165–168, `canonical == key`와 시장 검증)를 거친 정규 키에서 `candidateLineage`로 만들어지고, `dispatchHandoffs`는 시장 하나 단위라 시장 축이 상수입니다. 골든 queue 블록은 소유자 범위를 따로 정의하지 않고, `dedup_key_fields`의 앞 넷(account·market·symbol·position_generation)과 일치합니다. | **(a)** `NewOwnerKey`에만 계좌 `ToUpper`를 추가 → strategyhandoff 무태그·태그 ok. **(b)** `ownerScopeOf`에서 시장 축을 `""`로 → strategyhandoff 무태그·태그 ok, 엔진 태그 a112 시험 ok(실질 동등 변이이지만 고정되어 있지 않음). | strategyhandoff `_test.go`에서 strategyrouter를 import합니다. 시험 import는 허용 목록 검사(productionOnly) 대상이 아니고, strategyrouter는 이미 strategyflow를 통해 전이 폐포 안에 있습니다. (가) `reflect`로 `OwnerKey` 필드 이름·개수가 {AccountRef, Market, Symbol, PositionGeneration}인지 확인하고, (나) 유효한 계보 쌍에 대해 `ownerScopeOf(a)==ownerScopeOf(b) ⇔ NewOwnerKey(a)==NewOwnerKey(b)`를 표 기반으로 확인합니다(표기 변형과 세대·종목·계좌 차이 포함). |
| 5 | (T) | "Admit이 유일한 문"이라는 주석이 낡았습니다. | 세 곳: dependency_closure_test.go:157 "`Handoff`를 만드는 문은 `Admit` 하나뿐", guard_test.go:611–612 "Admit을 부르는 자리를 하나로 고정", :618 "`Admit`이라는 이름이 나오는 자리를 전부 센다". 과거 서술로 읽을 여지가 있는 한 곳: strategy_entry_supervisor.go:479 "`dispatchHandoff().Deliver`로 내보내고". | 문 둘(`Admit`·`AdmitEachOwnerScope`)과 자리 둘로 고쳐 씁니다. |

## 필수 항목별 판정

1. **census 소스 유도의 완전성: P2(#1, #2).** (ii) "표에 막히지 않고 census도 못 보는 문"은 **없습니다**. 다섯 모양 모두 `exportedSurface`에서 먼저 막힙니다(실험 A). 표를 정식으로 늘린 뒤에는 유도가 다섯을 모두 놓칩니다(실험 B). 엔진 쪽 우회 경로는 이렇게 판정했습니다.
   - 다른 패키지를 경유: `TestOnlyTheEngineImportsThisSeam`이 막습니다. 실험 E에서 `internal/zzrelay`가 `AdmitEachOwnerScope`를 부르자 `dependency_closure_test.go:234 … imports the handoff seam; only the engine may`로 FAIL.
   - dot import와 함수 값 반환: `admitSites`가 선언 안의 모든 Ident를 보므로 잡힙니다. dot import는 fixture로 확인됩니다. 함수 값 반환은 코드 판독에 근거하며 실행하지 않았습니다.
   - 잡히지 않는 것: 한 번만 언급하고 여러 번 호출하는 경우(#2).
2. **동결 표면·import 폐포: 통과.**
   - `AdmitEachOwnerScope` 선언 문자열이 정확합니다(대조 GREEN).
   - 비공개 `ownerScope`는 표면 검사와 필드 검사에 새지 않습니다. 두 검사 모두 공개 타입만 봅니다.
   - `strings` 추가는 직접 import 허용 목록에만 영향을 줍니다. 전이 금지 목록은 모듈 안 변경 패키지만 대상이라 충돌이 없습니다.
   - `gofmt -l`(/usr/local/go/bin/gofmt) 결과 변경 파일의 위반은 0입니다.
3. **소유자 범위의 정의: P2 낮음(#4).**
   - 네 축과 정규화(TrimSpace, 종목 ToUpper)는 `NewOwnerKey`와 같습니다.
   - 시장 축은 검증 없이 원문을 쓰지만, 봉인된 계보가 이미 정규화·검증되어 있어 실질 차이는 없습니다.
   - 두 정의가 갈라지는 것을 잡는 핀이 없습니다. 핀 모양은 #4에 적었습니다.
4. **다른 경계 census의 정합.**
   - `singleProposalAssumptionCensus`: 정상. `range authority.entries`는 셈 대상인 두 모양(`len(x.entries) <비교> 1`, `x.entries[0]`)에 들지 않고 판본 전과 같은 순회입니다. census는 새 파일을 스캔하며, 보지 못해서 통과한 것이 아닙니다.
   - `TestExactlyOneProductionCallSiteTurnsAHandoffIntoADispatch`: 정상. `dispatch` 호출은 여전히 `runProductionStrategyMarketCycle`의 closure 한 곳뿐입니다. `deliverEachStrategyHandoff`의 `Deliver`는 몸통만 넘깁니다.
   - `TestNoProductionSiteDiscardsTheSeamsAdmissionAnswer`: 새 문 `dispatchHandoffs`를 **보지 못합니다**(#3, 실험 D).

**범위 밖**: 3260f4eb 사유 코드 census와 36ade9b2(5.6.2.1)는 이 목소리의 필수 항목이 아니어서 실행으로 검증하지 않았습니다.

Recommendation: APPROVE because P0·P1이 없고 네 P2는 오늘 생산 경로(활성화 0건, B2 개수 관문)에서 주문 동작을 바꾸지 않기 때문 — 단 #2의 엔진 쪽 중복 범위 핀(실측 완료)은 5.2.2.2가 B2를 걷어 내기 전에 넣고, #1은 유도 대신 허용 목록으로 바꾸거나 "자동으로 새 문을 본다"는 주장을 철회할 것.
