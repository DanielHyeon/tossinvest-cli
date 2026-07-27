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
