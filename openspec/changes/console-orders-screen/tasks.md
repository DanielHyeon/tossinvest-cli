# Tasks: console-orders-screen

각 `[T]`는 RED → GREEN → REFACTOR → VERIFY를 거친다. 읽기 전용 화면이므로 실계좌 mutating
단계가 없다 — 검증은 fixture와 fake seam으로 한다.

> **선행 change**: `console-operator-overview` — 라우트·주입 능력 정적 검사의 패키지 전체 확장,
> 미측정 자료형, 미측정 사유 다섯. 그 확장 없이 이 change의 새 파일에 등록된 라우트는 라우트 표
> 검사에 **보이지 않는다**.
> **위험 등급**: 이 change는 High-risk 표면 둘을 건드린다 — `internal/journal`(원장 읽기),
> `internal/official`(계좌 게이트웨이). 둘 다 **가산이며 기존 서명·반환·에러 매핑을 바꾸지
> 않는다**. Function Logic Map 면제 없음.
> 파일 표면: `internal/official`·`internal/journal`·`internal/console`·`cmd/tossctl`.
> 활성 change `add-candidate-discovery`(`internal/candidate`·`internal/candidatesrc`)와
> `internal/official`이 겹친다 — 그쪽은 `ratebudget.go`, 이쪽은 주문 읽기로 파일이 다르다.

## 1. 원문 보존 주문 읽기 (internal/official, 가산)

- [x] 1.1 [T] 브로커의 decimal 문자열을 **보존하는** 주문 읽기. `RawHolding`의 선례를 그대로
      따른다([asset_reads.go:75](../../../internal/official/asset_reads.go#L75)의 근거가 그대로
      적용된다). RED: `parseDecimal("")→0`이므로 `domain.Order`로는 부재·해석불가·진짜 0이
      **콘솔이 보기 전에 하나가 된다** → GREEN(원문 문자열 보존, 부재는 부재로).
- [x] 1.2 [T] `Orders`·`adaptOrder`·기존 호출자 **동작 무변경**을 테스트로 고정.
      RED: 기존 경로의 반환·에러 매핑이 바뀌면 실패 → GREEN.
- [x] 1.3 [T] `currency`에서 시장을 유도해 함께 나른다. RED: 현행은 디코드 후 **버린다**
      (`orders_reads.go:51`) → GREEN. 종목명은 원문에도 없으므로 **만들지 않는다**(D7).
- [x] 1.4 [T] `HasNext`·`NextCursor`를 seam 밖으로 노출(D5). RED: 현행 `Orders`가 둘 다
      버려서 **잘린 첫 페이지가 확신에 찬 짧은 건수**가 된다 → GREEN.
- [x] 1.5 [T] 조건주문 읽기도 같은 규율로(`ConditionalOrders`). RED: 일반 주문만 부르면
      조건주문 잔여물이 살아 있는데 미체결 0건이 렌더된다 → GREEN.

## 2. 발주 주체 (internal/journal, 가산)

- [x] 2.1 [T] `journal.ReadOnly`에 `mutation_attempts.broker_order_id` 읽기 접근자 가산.
      `mode=ro` 연결만 쓴다. RED: 현행 일곱 메서드 중 이 테이블을 읽는 것이 없다 → GREEN.
- [x] 2.2 [T] `readOnlyTables`에 그 테이블 등록. RED: 등록하지 않으면 `OpenReadOnly`가 성공한
      뒤 **질의가 하나씩 실패하고 0행을 돌려주며**, 0행은 "전부 수동 주문"으로 읽힌다 —
      그 목록이 존재하는 이유가 정확히 그것이다 → GREEN(열 때 한 번 명확히 거부).
- [x] 2.3 [T] `TestTheReadOnlyHandleHasNoWriteMethods`의 메서드 열거 갱신. RED: 새 메서드가
      쓰기 판정에 걸리거나 열거가 낡아 가드가 무의미해진다 → GREEN.

## 3. 라우트 가드 — 예외를 열되 크기를 잰다

- [x] 3.1 [T] 판정부를 **순수 함수**로 추출. RED: 현행 `registeredRoutes`는 디스크의 소스를
      파싱하므로 테스트가 가짜 라우트를 "등록"할 수 없고, 그러면 구현자는 "allowlist에 그
      경로들이 없다"는 **아무것도 재지 않는 테스트**를 쓰게 된다 → GREEN.
- [x] 3.2 [T] `reading(next)` wrapper — `mutating`의 거울. GET/HEAD 외 405이고 같은
      `ast.Inspect` 분기가 인식한다(D3 결정 2). RED: 현행 `route`는 `{Path, Session, CSRFGated}`
      뿐이라 **"GET"이 "CSRF 보호가 없다"로 퇴화하고**, 그 상태에서 `POST /orders`가 세션만으로
      CSRF 없이 통과한다 → GREEN.
- [x] 3.3 [T] 정확 경로 allowlist `{"/orders"}` — **바이트 일치**. RED: 접두 일치면
      `/orders/cancel`, `ToLower`면 `/Orders`, `TrimSuffix`면 `/orders/`가 통과하고
      후행 슬래시 패턴은 **서브트리 패턴이라 하위 경로를 그 핸들러로 라우팅한다** → GREEN.
- [x] 3.4 [T] 예외를 **두 판정 루프 모두**에서 참조. RED: `actVerbs` 루프가 `accountVerbs`를
      포함하므로 `/orders`가 양쪽에 걸린다 → GREEN. **`consoleStateChanging`에 넣지 않는다** —
      넣으면 CSRF 게이트가 요구되어 읽기 라우트가 열리지 않는다.
- [x] 3.5 [T] 예외 조건 검사 — `Path == "/orders" && Reading && !CSRFGated`. 셋 중 하나라도
      깨지면 예외가 적용되지 않음을 고정.
- [x] 3.6 [T] 예외의 크기를 예외 자신이 잰다 — `/orders/cancel`·`/orders/new`·`/orders/amend`·
      `/Orders`·`/orders/` 다섯이 등록되면 가드가 **실패함을 직접 확인**(3.1의 순수 함수 위에서).
- [x] 3.7 [T] `OrdersReader`의 메서드 이름이 금지 동사(`"order"`)에 걸리는 문제.
      **금지 목록에서 `"order"`를 지우지 않는다** — 지우면 `/order/place`류가 통과한다.
      선행 change의 **seam별 allowlist**에 `OrdersReader→{Orders}`를 등록하는 방식으로 푼다.

## 4. 주문 조회 seam과 캐시

- [x] 4.1 [T] `OrdersReader` — 콘솔이 `internal/official`을 import하지 않는다(기존
      `TestTheConsoleHoldsNoBrokerOfItsOwn` 유지). 필터 타입은 **콘솔 로컬**이다 —
      `official.OrdersFilter`를 명명할 수 없다.
- [x] 4.2 [T] 캐시 TTL 15초 이상, lazy, 서버측 폴러 없음. 갱신 1회 = **3콜**
      (미체결 + 종결 + 조건주문 — design D2 개정, §8.1).
- [x] 4.3 [T] 필터는 **캐시 위에서 in-process**(D6). RED: 브로커 파라미터로 넘기면
      `?state=live`와 `?state=closed`가 별개 캐시 키가 되어 TTL당 3콜이 6콜이 된다 → GREEN.
- [x] 4.4 [T] 검증 중 갱신 보류를 holdings와 **같은 판정**으로 상속(in-process run 신호 +
      다른 프로세스 runlock mtime 5분 상한). RED: 검증 중에도 조회가 나간다 → GREEN.
- [x] 4.5 [T] **부분 실패는 합산하지 않는다**. RED: 조건주문 조회만 실패했는데 총계가 일반
      주문 건수로 렌더되면 **조용히 낮은 확신에 찬 숫자**가 된다 → GREEN("N건 + 조건주문 미측정").
- [x] 4.6 [T] `cmd/tossctl` 배선. `verifyBrokerFactory`가 돌려주는 `verifylive.Broker`에는
      `Orders`가 없으므로 경로를 **명시적으로 고른다** — 계정 seq 해석을 중복하지 않는 쪽으로.
      RED: 두 번째 `*official.Client`를 만들면 계정 seq 해석이 중복되고 429 위험이 는다.
- [x] 4.7 [T] seam 미배선 빌드에서 `/orders`는 `seam_unwired`를 말하고 나머지 화면은 영향받지
      않는다.

## 5. `/orders` 화면

- [x] 5.1 [T] 라우트 등록 — GET, `session0` 안, `reading` wrapper 적용, CSRF 밖. §3 가드 통과.
- [x] 5.2 [T] 표 렌더 — 시각·심볼·시장·매수/매도·상태·주문수량·체결수량·주문가·평균체결가·
      주문번호·발주 주체. **브로커가 주지 않은 값은 0이 아니라 "—"**. RED: 미체결 주문은 체결
      정보 전체가 null이므로 **모든 미체결 행이 평균체결가 0으로 렌더된다** → GREEN.
      시각은 `SubmittedAt`(실제 순간)을 쓰고 파싱 실패 시의 처리를 명시한다.
- [x] 5.3 [T] 미체결과 종결의 구분. 조건주문을 **미체결에 포함**한다.
- [x] 5.4 [T] 발주 주체 3-상태 — `엔진 발주`/`그 밖`/`원장 미판독 — 불명`. RED: 원장을 못
      읽었을 때 전 행이 "그 밖"이면 엔진이 아무 일도 안 한 것처럼 보인다 → GREEN.
- [x] 5.5 [T] 원장 미판독 안내는 페이지 1회, 행마다 반복하지 않는다.
- [x] 5.6 [T] 필터 UI + `N/M건`. RED: 필터 적용 시 총건수가 사라지면 "주문이 이것뿐"으로
      읽힌다 → GREEN. **목록이 미측정이면 필터를 비활성**으로 — `0/—건`은 "0건이 일치"로 읽힌다.
- [x] 5.7 [T] 잘린 페이지는 "N건 이상". RED: `HasNext`가 참인데 숫자로 단정 → GREEN.
- [x] 5.8 [T] meta refresh 주기 = 주문 캐시 TTL, 상수에서 파생.
- [x] 5.9 [T] 화면에 주문을 내거나 정정·취소하는 폼·링크가 없고 확인 문자열 입력란이 없음을
      렌더 결과로 고정(D8 — 사용자 지시 2026-07-27).
- [x] 5.10 [T] 선행 change가 `seam_unwired`로 남긴 개요의 미체결 패널이 실제 값을 갖는다.
      RED: 배선 후에도 미측정으로 남으면 배선이 안 된 것이다 → GREEN.

## 6. 배선과 내비게이션

- [x] 6.1 [T] 내비게이션에 `/orders` 추가, 404 본문의 화면 목록 갱신, 라우트 수 하한 상향.
- [x] 6.2 [T] 기존 콘솔·원장·official 테스트 전부 통과.

## 7. 게이트

> §1-§6 구현 중 **내부 로직을 편집한 기존 함수**(7.1의 대상 범위):
> `console.Console.routes`(라우트 1건 등록 추가), `console.Console.New`(캐시 1개 생성 추가),
> `console.Console.overview`(미체결 패널 1줄 — `seam_unwired` 상수 → `openOrdersPanelFrom`),
> `journal.readOnlyTables`(값 추가 — `checkSchema`의 루프 입력).
> `internal/official`·`internal/journal`의 새 심볼은 전부 신규 leaf이며 기존 함수 본문은
> 한 줄도 바뀌지 않았다. 구현 중 발견한 스펙 편차 4건은 `issues.md`에 있다.

- [x] 7.1 Function Logic Map + `check_analysis.py` — **면제 없음**. `internal/journal`·`internal/official`
      가산 함수와 라우트 가드 판정부를 전부 매핑했고, 대부분의 결론이 `High-risk impact: yes`다.
- [x] 7.2 PM registry allowlist + fixture 등록
- [x] 7.3 `make sdd-sync && make sdd-check && make gate CHANGE=console-orders-screen`
- [x] 7.4 독립 리뷰(별도 컨텍스트) — 완료. 세 물음 전부 실행으로 답했다.
      ① httptest 기록으로 네 호출의 요청이 바이트 동일하고 계정 해석 1회 공유, 401 재시도·에러 분류·
         429의 rate 예산 기록까지 같은 경로임을 확인. 스키마 이력상 v6 테이블이 있는데
         `mutation_attempts`만 없는 버전은 존재하지 않음(전진 전용·drop/rename 금지).
      ② `/orders/cancel`·`/orders/new`·`/orders/amend`·`/Orders`·`/orders/` 다섯이 전부 실패,
         `POST /orders`는 405. 접두 일치·`ToLower`로 바꾸면 각각 통과함을 변이로 확인.
      ③ 조건주문만 실패하면 "N건 + 미측정"이고 합산되지 않음. 그리고 리뷰가 **P0 둘을 더 찾았다** —
         `status` 누락으로 잔여물이 숨을 수 있었고, 개요가 오래된 읽기를 측정값으로 냈다.

## 8. 구현 리뷰 대응 (적대적 리뷰 P0×2 / P1×3, 2026-07-28)

> 리뷰가 찾아낸 가장 날카로운 사실: **다섯 개의 변이가 전체 스위트를 통과했다.** 그래서 각
> 수정마다 결함을 다시 넣어 새 테스트가 실제로 실패하는지 확인했다 — 변이 12건 전부가
> 물었고, 목록은 `issues.md` I-7에 있다.
> `internal/journal`은 변경 0이고, `internal/official`은 이 change가 만든 두 raw 읽기
> (`OrdersRaw`·`ConditionalOrdersRaw`)에만 status 가드가 들어갔다 — 선행 심볼은 전부
> `9969238`과 바이트 동일하다(§8.10, `issues.md` I-8).

- [x] 8.1 [T] **P0-1** 갱신 1회를 3콜로: `status=OPEN` + `status=CLOSED` + 조건주문.
      RED(`TestOneRefreshAsksTheOpenGroupAndTheClosedGroupSeparatelyAndTheLiveOneWhole`):
      실제 와이어가 `GET /api/v1/orders?limit=100` 한 건이었다 — `status`는 `required: true`이고
      생략하면 거부되거나 **전 기간 한 페이지**가 돌아와 101번째 살아 있는 주문이 표에도 집계에도
      없이 `0건 이상`이 된다 → GREEN. `OPEN`에는 `limit`·`cursor`를 **보내지 않는다**(API가 무시).
- [x] 8.2 [T] `OrdersReading`을 세 목록으로 분리(`Open`/`Closed`/`Conditional`)하고 잘림 플래그도
      각각. RED(`TestTheLiveCountIsANumberEvenWhenTheClosedPageWasTruncated`): 잘림을 공유하면
      **미체결 건수가 하한이 된다** — `OPEN`이 전량 반환이라 얻은 정확성을 종결 페이지가 도로
      깎는다 → GREEN. 미체결 건수는 이제 하한이 아니라 숫자다.
- [x] 8.3 [T] 두 그룹에 겹쳐 오는 주문 중복 제거. openapi는 `PARTIAL_FILLED`를 **OPEN과 CLOSED
      양쪽**에 정의하므로 한 주문이 두 번 온다. RED(`TestAnOrderInBothGroupsIsOneRowAndIsCountedOnce`):
      한 주문이 두 행·두 건이 된다 → GREEN(미체결 사본이 이긴다).
- [x] 8.4 [T] **P0-2 / D9** 개요의 미체결 건수에 캐시 시각·경과·TTL 초과 표시를 값과 **같은 셀**에.
      RED(`TestTheOverviewNeverRendersAStaleOrdersReadingAsAMeasuredNumber`): 개요는 설계상
      갱신하지 않으므로 세 시간 전 빈 계좌의 `0건`을 나이도 표시도 없이 측정값으로 렌더했다 →
      GREEN. 검증 중에는 "주문 화면 갱신도 보류된다"를 함께 적는다(`verify_suspended` 사유를
      쓰지 않는 이유는 보유 패널의 `brokerReadable`과 같다 — 이 화면은 보류될 것이 없다).
- [x] 8.5 [T] **P1-3** `/orders` 합계도 같은 규율. RED
      (`TestTheOrdersCountsCarryTheirOwnProvenanceInTheSameBreath`): 굵은 합계와 나이·실패가
      다른 문단이면 읽는 사람이 둘을 잇지 않는다 → GREEN(조회 시각은 합계 `<dd>` 안, 실패·보류
      문단은 같은 `<section>` 안).
- [x] 8.6 [T] **P1-1** 조건주문 발주 주체 조인. RED
      (`TestAWatchingConditionalIsNeverLabelledOtherByAJoinThatCannotSucceed`):
      `rec.Triggered`는 감시 중인 조건주문에서 항상 비어 있고 어댑터는 OPEN 그룹만 부르므로
      **화면에 나오는 모든 행이 `engineIDs[""]`를 조회했고**, 그 결과가 상수 `그 밖`이었다 → GREEN.
      이 빌드에서 조건주문 id는 `mutation_attempts`에 **기록되지 않으므로**(§8.7) 불일치는
      `불명`이고, 일치는 증거이므로 `엔진 발주`로 적는다. 페이지 안내 1회로 이유를 말한다.
- [x] 8.7 조사 결과 기록: `mutation_attempts`를 쓰는 것은 `internal/execgw`의 PLACE/CANCEL/AMEND
      경로뿐이고 `broker_order_id`는 `PlaceOrder`/`CancelOrder`/`ModifyOrder`가 돌려준 **일반 주문**
      번호다. 조건주문 등록은 `trading.Service.ConditionalPlace`와 `internal/verifylive`를 지나며
      **둘 다 journal attempt를 열지 않는다**(`internal/verifylive`에는 journal import 자체가 없다).
      → 조건주문 id는 이 빌드에서 그 칼럼에 도달하지 않는다.
- [x] 8.8 [T] **P1-2 / P2-1** 필터의 "판정할 수 없는 행은 걸러 내지 않는다" 규칙을 테스트로 고정.
      RED(`TestASideFilterNeverHidesAWatchingConditional`, `TestAMarketFilterNeverHidesAnOrderWhoseMarketIsUnknown`):
      `!r.Conditional &&`를 지워도 전체 스위트가 통과했고, 시장 축은 시장이 `—`인 살아 있는 주문을
      **실제로 제외하고 있었다** → GREEN(시장도 방향과 같은 "해당 없음"으로).
- [x] 8.9 [T] `TestOrdersFilterEmpty` 개명·재작성 →
      `TestOrdersFilterEmptyOmitsEveryParameterIncludingTheRequiredOne`. 빈 필터가 파라미터를
      하나도 보내지 않는다는 **클라이언트 동작**은 그대로 고정하되, 그것이 쓸 수 있는 호출 형태라는
      뜻이 아님을 이름과 주석에 적었다 — 이 테스트를 허가로 읽은 것이 P0-1의 출처다.
      `TestBothOrderReadsSendTheGroupTheyWereGiven`를 추가해 두 읽기가 받은 그룹을 그대로
      와이어에 싣는지 고정한다.
- [x] 8.10 [T] **가드를 클라이언트로 옮긴다**(Manager 판정, `issues.md` I-8). `OrdersRaw`와
      `ConditionalOrdersRaw`가 빈 `status`를 `ErrOrderStatusRequired`로 거부하고, **요청을 보내기
      전에** 거부한다. RED(`TestTheRawReadsRefuseARequestWithNoStatusGroup`): 호출부에만 사는
      규칙은 다음 호출부에서 똑같이 잊힌다 — 방금 고친 P0의 모양이 정확히 그것이다 → GREEN.
      두 엔드포인트의 문구는 다르다: 조건주문은 두 그룹 모두 페이지네이션하므로 "전량 반환"이라
      쓰지 않는다. `cmd/tossctl`의 와이어 테스트는 그대로 둔다(두 계층).
