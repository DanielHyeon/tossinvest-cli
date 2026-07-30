# Issues: console-orders-screen

WORKFLOW §예외 경로 기록. 구현 중 발견한 스펙·설계 결함과 그 처리.

## I-1 — P0-3의 같은 결함이 `cmd/tossctl` 쪽에도 있다 (분류: safe local, 승인됨)

**발견**: review.md P0-3은 `OrdersReader{Orders}`가 금지 동사에 걸리는 문제를 `internal/console`의
seam별 allowlist로 푼다고 정했고, tasks 3.7도 그렇게 적었다. 그런데 **같은 금지가 `cmd/tossctl`에도
따로 있다** — `cmd/tossctl/console_test.go`의 `TestTheConsoleIsHandedOneCapabilityAndNotABroker`가
`console.Options` 리터럴의 **필드 이름**을 `{"broker","client","order","place","cancel"}`로 막는다.
실제로 깨졌다:

```text
--- FAIL: TestTheConsoleIsHandedOneCapabilityAndNotABroker
    console_test.go:267: console.Options is given a Orders field; the dashboard's whole
    broker is console.HoldingsReader
```

리뷰 둘도, design도, tasks도 이 두 번째 가드를 보지 못했다. 이것은 overview change의 I-1이
기록한 것과 **같은 부류**다: 새 표면을 설계하면서 그 표면에 걸리는 가드를 확인하지 않았다.

**처리**: 콘솔 쪽과 **같은 방식**으로 풀었다 — 금지 목록은 손대지 않고, 이유를 명시한
`consoleFieldExemptions{"Orders": <근거>}`를 두고 그 예외에 **더 강한 검사를 붙였다**:
`consoleOrdersSeam(...)`이 돌려주는 값의 메서드 집합이 정확히 `[Orders]`임을 reflect로 확인한다
(기존 `HoldingsReader` 검사와 같은 형태). 예외에 이유 문자열이 없으면 테스트가 실패한다.

`"order"`를 목록에서 지우지 않은 이유는 D3과 같다 — 지우면 미래의 `PlaceOrder` 필드가 같은 순간에
통과한다.

**Manager 확인 요청**: tasks 3.7의 문장("선행 change의 seam별 allowlist에 등록")은
`internal/console`만 가리킨다. `cmd/tossctl` 쪽 예외도 같은 근거로 승인할 것인지 판정 요망.

## I-2 — `reading`이라는 이름이 이미 다른 뜻으로 쓰이고 있다 (분류: safe local, 개명 판정)

**발견**: design D3은 wrapper 이름을 `reading(next)`으로 못박았다. 그런데 `internal/console`에는
이미 `type reading struct` — `console-operator-overview`가 도입한 `(값, 측정됨, 사유)` 자료형 —
이 있다([overview.go:165](../../../internal/console/overview.go#L165)).

Go에서 메서드 이름은 타입의 네임스페이스에 있으므로 `func (c *Console) reading(...)`과
패키지 수준 `type reading`은 **충돌 없이 컴파일된다**. 실제로 빌드·vet·전체 테스트가 통과한다.
그러나 한 패키지 안에서 같은 단어가 두 가지를 뜻하게 된다.

**처리**: design의 이름을 **그대로 따랐다**. 근거 셋:
① design은 승인된 계약이고 이름 변경은 구현자가 임의로 할 일이 아니다,
② 라우트 표 검사가 `fn.Sel.Name == "reading"`으로 읽으므로 이름이 계약의 일부다,
③ 두 뜻이 문법적으로 결코 만나지 않는다(하나는 selector 뒤, 하나는 타입 위치).
`console.go`의 wrapper 주석에 이 사실을 적어 두지는 않았다 — 적으면 그것이 이름을 바꿔야 할
이유처럼 읽힌다.

**반영(아래 Manager 판정에 따라 개명)**: wrapper 메서드 `reading` → `readOnly`, 라우트 레코드
필드 `route.Reading` → `route.ReadOnly`, AST 인식 분기 `case "reading"` → `case "readOnly"`,
등록 `c.session0(c.readOnly(c.handleOrders))`, 테스트
(`TestTheOrdersRouteIsRegisteredReadOnlyAndNothingElseIs` 포함)와 doc comment 전부 갱신.
`console.go`의 wrapper 주석에 **이름을 그렇게 고른 이유**를 남겼다 — 적지 않으면 다음 사람이
`mutating`과의 대칭을 이유로 `reading`으로 되돌린다.

개명 뒤 AST 분기가 실제로 새 이름을 인식하는지는 변이로 재확인했다(wrapper를 떼면 가드가
실패한다). 개명이 가드를 조용히 죽이는 것이 이 작업의 유일한 실패 방식이다.

## I-3 — 한 페이지의 크기를 design이 정하지 않았다 (분류: safe local)

**발견**: D5는 "`HasNext`가 참이면 N건 이상"을 정했지만 **한 번에 몇 건을 가져오는지**는 정하지
않았다. 그 값이 곧 "언제 '이상'이 붙는가"를 결정하므로 어딘가에는 있어야 한다.

**처리**: `cmd/tossctl`의 어댑터에 `consoleOrdersPageLimit = 100`을 두고 두 목록 모두에 적용했다.
페이지를 끝까지 걷지 **않는** 쪽을 골랐다 — 한 번의 페이지 로드 뒤에 브로커 호출 루프가 생기고,
그것은 화면당 2콜이라는 계약을 깨는 유일한 방법이기 때문이다. 잘림은 감추지 않고 "N건 이상"으로
드러내므로 이 절단은 손실이 아니라 정직한 상한이다.

## I-4 — 조건주문의 상태는 콘솔이 판정하지 않는다 (분류: safe local)

**발견**: spec은 "미체결 주문과 종결된 주문은 구분해 표시한다"를 요구하고, design D1은 상태 판정을
`internal/brokerstate`에서 가져오라고 했다. 그런데 `brokerstate`는 **일반 주문의 10-값 enum만**
모델링한다. 조건주문의 자기 상태(`WATCHING`)는 다른 어휘이고
([brokerstate/derive.go:18](../../../internal/brokerstate/derive.go#L18)이 그 둘을 섞지 말라고
명시한다) 이 저장소 어디에도 조건주문 상태의 전이표가 없다.

**처리**: 조건주문에 대해서는 **판정하지 않고 질문을 보고한다**. 어댑터는 브로커의 `OPEN` 그룹
(`verifylive.ConditionalStatusOpen`)만 요청하므로 돌아온 것은 전부 살아 있는 조건주문이고,
화면은 그것을 미체결로 세면서 브로커의 원문 상태 문자열을 상태 열에 그대로 적는다.
없는 전이표를 지어내는 것이 이 change가 막으려는 그 일이다.

부수 결과: 조건주문은 `종결` 필터에 절대 걸리지 않는다. 그리고 조건주문 payload에는 방향(side)이
없으므로 **방향 필터는 조건주문에 적용되지 않는다**(제외가 아니라 미적용) — 제외하면 매수 필터를
걸었을 때 잔여 조건주문이 조용히 사라진다.

## I-5 — 조건주문 id는 이 빌드의 `mutation_attempts`에 도달하지 않는다 (분류: 조사 결과, P1-1 근거)

**질문**: 리뷰 P1-1은 두 수리 중 하나를 고르라고 했다 — ① `firstNonBlank(rec.Triggered, rec.ID)`로
조인하거나, ② 조건주문 id가 실제로 원장에 기록되지 않는다면 `불명`으로 렌더한다.

**조사**(현재 HEAD, `rg` + 호출부 추적):

- `mutation_attempts`를 **쓰는** 경로는 `internal/journal`의 `Attempt` 수명주기뿐이고, 그 `kind`는
  `KindPlace`·`KindCancel`·`KindAmend` 셋이다([durability.go:62](../../../internal/journal/durability.go#L62)).
  `broker_order_id`는 `MarkAcked`가 넣으며 값은 `execgw`가 `PlaceOrder`/`CancelOrder`/`ModifyOrder`
  응답에서 얻은 **일반 주문** 번호다([lineage.go:122](../../../internal/journal/lineage.go#L122)).
- `internal/execgw` 전체에 `Conditional`이라는 식별자가 **하나도 없다**. 조건주문은 그 게이트웨이를
  지나지 않는다.
- 조건주문 등록의 실제 경로는 둘이다: `trading.Service.ConditionalPlace`
  ([conditional.go:123](../../../internal/trading/conditional.go#L123), `cmd/tossctl/order.go`가 호출)와
  `internal/verifylive/mutate.go`. **둘 다 journal attempt를 열지 않으며**, `internal/verifylive`의
  비테스트 파일에는 `journal` import 자체가 없다.

**결론**: 조건주문 id는 이 빌드에서 그 칼럼에 기록되지 않는다. 따라서 조인의 **불일치는 증거가
아니다** — 원장에 물어본 적 없는 질문에 "그 밖"이라고 답하는 것이다.

**처리(두 수리를 다 적용)**: `conditionalOriginOf`를 따로 두고 `Triggered → ID` 순으로 조회하되,
**불일치는 `originOther`가 아니라 `originUnknown`**으로 판정한다. 근거:

- 불일치를 `불명`으로 두는 것이 위 조사 결과가 요구하는 정직함이다.
- 그럼에도 조회를 살려 두는 이유는, 일치는 **진짜 증거**이기 때문이다. 언젠가 조건주문 등록이
  원장을 남기게 되면 화면은 그날 바로 사실을 말한다. `불명` 상수로 굳혀 두면 영원히 침묵한다.
- 리뷰가 시연한 fixture(조건주문 자신의 id가 `mutation_attempts`에 있는 원장)를 테스트에 그대로
  넣었다 — 그 행은 이제 `엔진 발주`로 렌더된다.
- 원장이 읽히는데도 `불명`이 나오는 이유는 **페이지 안내 1회**로 설명한다. 설명 없는 `불명`이
  설명 있는 `엔진 발주` 옆에 있으면 조인 버그로 읽히고, 다음 사람이 열을 "정리"한다.

## I-6 — `PARTIAL_FILLED`는 두 그룹 양쪽에 속한다 (분류: safe local)

**발견**: design D2 개정으로 일반 주문이 `OPEN`·`CLOSED` 두 호출이 되면서, openapi의 그룹 정의를
다시 읽어야 했다. `status` 파라미터 설명은 `PARTIAL_FILLED`를 **양쪽 집합 모두에** 열거한다:

```text
OPEN   ∈ {PENDING, PARTIAL_FILLED, PENDING_CANCEL, PENDING_REPLACE}
CLOSED ∈ {FILLED, CANCELED, REJECTED, REPLACED, CANCEL_REJECTED, REPLACE_REJECTED, PARTIAL_FILLED}
```

한 주문이 두 응답에 모두 나올 수 있고, 그러면 표에 두 행·집계에 두 건이 된다. 부분 체결은 흔한
상태이므로 예외가 아니다.

**처리**: 주문번호로 중복을 제거하고 **미체결 사본이 이긴다**(브로커가 방금 "대기 중"이라고 말한
쪽이다). 테스트로 고정했고, 제거를 빼면 그 테스트가 실패한다.

## I-7 — 변이 검증 표 (리뷰가 요구한 증명)

리뷰의 핵심 관찰은 "다섯 개의 변이가 전체 스위트를 통과했다"였다. 그래서 이번 수정은 각각에 대해
**결함을 다시 넣고 새 테스트가 실패하는지** 확인했다. 10건 전부가 물었다.

- 라이브 읽기에서 `Status: OPEN` 제거 →
  `TestOneRefreshAsksTheOpenGroupAndTheClosedGroupSeparatelyAndTheLiveOneWhole`
- `CLOSED` 호출 통째로 제거 → 같은 테스트(2콜로 떨어진다)
- 두 일반 그룹이 잘림 플래그를 공유 → `TestTheLiveCountIsANumberEvenWhenTheClosedPageWasTruncated`
- 중복 제거 삭제 → `TestAnOrderInBothGroupsIsOneRowAndIsCountedOnce`
- 개요에서 시각·나이·TTL 표시 제거 → `TestTheOverviewNeverRendersAStaleOrdersReadingAsAMeasuredNumber`
- `/orders` 합계를 굵은 숫자만으로 되돌리고 출처를 다른 문단에 →
  `TestTheOrdersCountsCarryTheirOwnProvenanceInTheSameBreath`
- 조건주문 조인을 `rec.Triggered`로 되돌림 →
  `TestAWatchingConditionalIsNeverLabelledOtherByAJoinThatCannotSucceed`
- 방향 필터에서 `!r.Conditional &&` 삭제 → `TestASideFilterNeverHidesAWatchingConditional`
- 시장 필터에서 `—` 예외 삭제 → `TestAMarketFilterNeverHidesAnOrderWhoseMarketIsUnknown`
- `OrdersRaw`가 받은 `status`를 버리게 함 → `TestBothOrderReadsSendTheGroupTheyWereGiven`
- `OrdersRaw`의 status 가드를 무력화(조건이 결코 참이 되지 않게) →
  `TestTheRawReadsRefuseARequestWithNoStatusGroup`, `TestOrdersFilterEmptyOmits…`
- `ConditionalOrdersRaw`의 status 가드를 무력화 → `TestTheRawReadsRefuseARequestWithNoStatusGroup`

> 마지막 두 변이는 처음에 가드를 **삭제**하는 형태로 넣었더니 미사용 import 때문에 **빌드가**
> 깨져 테스트가 실행조차 되지 않았다. 컴파일러가 잡은 것은 증명이 아니므로 조건을 절대 참이 될
> 수 없는 값으로 바꾸는 형태로 다시 넣었고, 그때 두 테스트가 실제로 물었다.

## I-8 — status 가드는 클라이언트에 있어야 한다 (분류: Manager 판정에 따른 번복)

**처음 판단(틀림)**: `OrdersRaw`에 status 가드를 넣지 않고 `cmd/tossctl`에만 두었다. 이유로
"`internal/official`의 additive 확인 grep(`git diff -U0 … | grep '^-[^-]'`)에 삭제 줄이 잡힌다"를
들었다.

**Manager 판정**: 그것은 **보증이 아니라 증명을 최적화한 것**이다. 이 change의 additive 주장은
**High-risk 패키지의 기존 호출자**에 대한 것이다 — `Orders`·`OrderByID`·`adaptOrder`·
`ConditionalOrders`와 그 호출자의 동작이 같다는 것. `OrdersRaw`와 `ConditionalOrdersRaw`는
**이 change가 만든 심볼**이므로 그것을 고쳐서 깨질 선행 호출자가 없다. grep은 주장의 증거이지
주장 자체가 아니며, **검사와 검사 대상이 어긋나면 움직여야 하는 것은 검사다.**

그리고 이것이 중요한 이유는 방금 고친 P0의 모양 그 자체다 — **호출자가 잊었다.** 호출부에만
사는 규칙은 다음 호출부에서, 이 건을 리뷰하지 않는 change에서 똑같이 잊힌다.

**처리**: 두 raw 읽기 모두 빈 `status`를 거부한다(`ErrOrderStatusRequired`). 에러는 파라미터
이름과 이유를 말하고, **요청을 보내기 전에** 거부한다 — 잊은 호출자가 rate limit 슬롯을 써 가며
알아낼 일이 아니다. `ShouldFallback`은 이 에러에 false를 답한다(호출자 실수이지 장애가 아니다).
`cmd/tossctl`의 와이어 테스트도 그대로 둔다 — 두 계층이 맞고, 라이브 읽기가 OPEN 그룹을
묻는다는 것을 증명하는 것은 그쪽이다.

두 엔드포인트의 문구는 **다르게** 썼다. 일반 주문은 `status=OPEN`이 전량 반환이라 생략이
"거부 아니면 조용한 절단"이지만, **조건주문은 두 그룹 모두 페이지네이션한다**(`limit` 기본 20,
최대 100). 조건주문 쪽에 "전량 반환"이라고 적었으면 이 change가 막으려는 바로 그 종류의
사실 아닌 문장이 됐을 것이다.

**additive 확인 형식도 바뀐다.** 이제 `git diff … | grep '^-[^-]'`에는 가드의 줄이 잡히므로,
**선행 심볼이 그대로임**을 보이는 형식으로 확인한다 — 두 파일의 최상위 선언을 `9969238`과
바이트 비교해 `OrdersRaw`·`ConditionalOrdersRaw`(이 change가 만든 것)와 import 블록(삽입만)
외에는 전부 unchanged임을 보인다. `internal/journal`은 여전히 변경 0이다.

---

## Manager 판정 (2026-07-28)

**I-1 — 승인, 그리고 범위 밖이 아니다.** `cmd/tossctl/console_test.go`의 금지 목록은 **같은
seam의 반대편에 있는 같은 가드**다. task 3.7이 콘솔 쪽만 적은 것은 내가 그쪽만 알았기 때문이지
cmd 쪽을 제외하려던 것이 아니다. 같은 처리를 적용하고 그 김에 검사를 더 강하게 만든 것이 옳다.

**I-2 — wrapper 이름을 고친다.** 구현자는 design D3을 그대로 따랐고 **D3이 틀렸다** —
`internal/console`에 이미 `type reading struct`가 있는 것을 확인하지 않고 `reading(next)`이라
이름 붙였다. 한 단어가 한 패키지에서 두 뜻을 갖는 것은 쓴 사람에게만 자연스럽고 이후 모두를
오도한다. 값 타입이 그 단어에 대한 권리가 더 크고 이미 커밋됐으므로 **wrapper를 `readOnly`로**
바꾼다. `c.session0(c.readOnly(c.handleOrders))`가 호출 지점에서 보증을 말한다. D3 정정 완료.

**I-3 — 승인.** 페이지 크기 100보다 **페이지네이션 순회를 두지 않은 것**이 중요하다.
한 페이지 로드 뒤에 브로커 호출 루프를 두면 "갱신당 2콜" 계약이 무한대로 바뀌면서도 모든
테스트가 계속 통과한다. 거부와 그 근거를 doc 주석에 유지할 것.

**I-4 — 승인, 근거가 정확하다.** 조건주문에 방향 필터를 "제외"가 아니라 "해당 없음"으로 둔
것이 맞다. 방향이 없는 payload가 `매수` 필터에 걸려 빠지면 **살아 있는 잔여물이 필터 뒤에
숨는다** — 0으로 세는 것과 같은 실패가 한 상호작용 뒤에 일어나는 것이다.
조건주문 상태기계를 만들지 않은 것도 맞다. 브로커의 상태 문자열을 그대로 보고하는 것이
정직하고, 유도하면 **잔여물이 다음 검증을 막는지 결정하는 상태에 두 번째 정본**이 생긴다.

---

## I-8 — `openOrdersPanelFrom`의 doc 주석이 rate budget을 **둘**이라고 적고 있었다 (2026-07-28, 수정 완료)

`internal/console/overview.go`의 `openOrdersPanelFrom` doc 주석이 “주문 화면의 갱신은 **둘**이
든다”고 적었다. design D2 개정(I-6)으로 일반 주문이 `OPEN`·`CLOSED` 두 그룹으로 갈라지면서 실제
비용은 **셋**(미체결 그룹 + 종결 그룹 + 조건주문)이고, spec delta와 `internal/console/console.go:199`,
`console.go:275`, `TestOneRefreshAsksTheOpenGroupAndTheClosedGroupSeparatelyAndTheLiveOneWhole`가
모두 셋이라고 말한다. 동작은 처음부터 셋이었으므로 계약 위반은 아니지만, **주석에 적힌 틀린
rate budget 숫자는 다음 사람이 예산을 잡는 근거**가 된다 — 이 콘솔에서 계정 하나의 rate budget은
엔진·실계좌 검증과 공유된다(spec Requirement 7).

수정: 주석을 “셋 — 미체결 그룹 1 + 종결 그룹 1 + 조건주문 1”로 고치고, 갈라 세는 이유(D2:
한 엔드포인트의 침묵을 다른 쪽의 0으로 보고하지 않는다)를 함께 적었다. 본문은 무변경이므로
`analysis/function-logic/internal-console--console.openorderspanelfrom/`의 분기표(B1·B2)는 그대로이고,
같은 문서의 “문서 drift(발견, 미수정)” 항목을 수정 완료로 갱신했다. 위 §Manager 판정 I-3에 남은
“갱신당 2콜”은 D2 개정 **이전에** 내려진 판정의 기록이므로 그대로 둔다 — 그 판정의 요지(페이지네이션
순회를 두지 않았다)는 콜 수와 무관하게 유효하다.

부수: `internal/console/static_test.go`의 `TestNoCapabilityReachesTheConsoleAroundOptions` 수정
(add-candidate-discovery/issues.md §11 G-1)으로 이 change의 diff hunk가 그 함수와 교차하게 되어
`internal-console--testnocapabilityreachestheconsolearoundoptions` target을 이 change에도 신설했다.
이 change의 base commit에는 그 함수가 이미 존재하므로 revision은 `current`다.

---

## I-9 — 콘솔이 계좌를 **화면마다** 해석하고 있었다 (2026-07-28, 수정 완료)

`runConsole`은 `Holdings: newConsoleHoldings(root)`와 `Orders: consoleOrdersSeam(root)`를 서로
독립적으로 배선했고, 두 seam이 **각자** 첫 사용에서 `verifyBrokerFactory`를 불렀다. 그 함수는
`buildVerifyBroker` — `resolveVerifyAccount`, 즉 2026-07-26에 429를 세 번 받아 검증 실행 3스텝을
잃게 한 `/api/v1/accounts` 읽기(measurements.md M4)를 하는 곳이다. 따라서 포지션 화면과 `/orders`를
모두 여는 콘솔 세션은 그 읽기를 **프로세스당 2회** 했고, 검증 실행이 세 번째를 더했다.

**주석이 과잉 주장하고 있었다.** `consoleOrdersSeam`의 doc 주석은 두 번째 클라이언트를 만들면
"계좌 시퀀스를 다시 해석한다"고 적었는데, 콘솔 전체로 보면 그 해석은 이미 두 번 일어나고 있었다.
`TestTheOrdersSeamResolvesTheAccountOnceAndBuildsNoSecondClient`는 **seam 하나 안의** 구축만 세므로
그 주석이 틀린 내내 통과했다 — 이 브랜치의 리뷰 다섯 라운드가 걷어내 온 바로 그 모양이다.

수정:

- `consoleBroker` + `newConsoleBroker` + `resolve()` — 콘솔이 공유하는 계좌 해석 1회. 구축은 락
  안에서 일어나 동시 개시 2건도 1회로 묶인다. 실패는 캐시하지 않는다(`openapi login` 이후 재시도).
- 두 읽기 seam은 `*rootOptions` 대신 이 공유 resolver를 받는다. **넘어가는 능력은 그대로다** —
  `lazyHoldings`/`lazyOrders`는 브로커도 method value도 필드로 갖지 않고 호출마다 지역 변수로
  꺼내 쓴다. `TestTheConsoleIsHandedOneCapabilityAndNotABroker`와 internal/console의 능력 순회는
  약화 없이 그대로 통과한다.
- `TestOpeningEveryConsoleReadScreenResolvesTheAccountOnce` — **콘솔 단위**의 주장. 모든 읽기 화면을
  열어(순차 2회 + 동시 1회) 해석 1회를 확인하고, 소스에서 ① `verifyBrokerFactory`를 부르는 함수
  집합이 이유가 적힌 `consoleBrokerBuildSites`와 정확히 일치하는지, ② `runConsole`이 공유 resolver를
  정확히 하나 만들어 모든 읽기 seam에 넘기는지 검사한다. RED는 2회를 보였고, 수정 후 seam별
  resolver를 되돌리는 변이에서 다시 2회로 실패한다(같은 변이에서 seam 단위 테스트는 계속 통과 —
  이 테스트가 무엇을 새로 잡는지가 그것이다).
- seam 단위 테스트는 남기되 doc 주석에 **범위가 seam 하나**라고 적었다.

**검증 실행은 공유하지 않는다(의도).** `consoleVerifyStarter`는 실행마다 자기 클라이언트를 만들며,
이유를 그 함수의 주석에 적었다: 기록에 적히는 계좌는 실주문 직전에 그 실행이 확인한 계좌여야 하고,
읽기 화면이 언젠가(자격증명 교체 이전일 수도 있다) 해석해 둔 값을 물려받으면 기록이 아무도
재확인하지 않은 계좌를 이름 붙이게 된다. 해석 실패의 계약도 다르다 — 읽기는 화면 위 문장이고
검증은 실행 전 치명이다. 비용은 실행당 1회로 묶여 있다.

증거: `cmd/tossctl/console.go`·`console_test.go`에 묶인 이 change의 모든 target을 재생성했고,
`newConsoleBroker`·`consoleBroker.resolve`·`newConsoleHoldings`·
`TestOpeningEveryConsoleReadScreenResolvesTheAccountOnce` 4종을 신설했다. `lazyHoldings.Holdings`는
본문이 바뀌었으므로 revision이 `base`에서 `current`로 올라갔다.
