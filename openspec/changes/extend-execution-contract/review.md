# Review: extend-execution-contract (proposal-freeze)

보이스 2개 — codex CLI(ENG/SAFETY/SPEC-QUALITY, 13건), Claude Eng 적대(P1 코드 대조, 25건). Manager가 모든 구조적 주장을 코드·openapi 문서에서 직접 재검증했다.

## 판정

**구현 착수 불가. 설계 재작성 필요.** Eng 보이스의 평결 "P1과 조립되지 않는다(Composes with P1? No)"가 옳다. 세 개의 독립적인 hard break가 코드로 확인됐고, 그 위에 openapi 문서 확인으로 **P1의 핵심 전제 하나가 사실이 아님**이 드러났다.

이 change의 원래 D1("조건주문은 새 MutationKind일 뿐 새 경로가 아니다")이 틀렸다. 조건주문은 주문이 아니다 — 식별자 네임스페이스가 다르고, 수명주기가 다르고, 응답 형태가 다르며, 그 실행은 **별개의 주문을 만든다**. 기존 attempt·fingerprint·추적 집합에 밀어넣으면 체결 감지와 reconcile이 깨진다.

## 최중대 발견: 브로커에 멱등키가 있다

P1 확정 스펙 `order-execution` "IN_DOUBT 해소" 1항: *"자동 재제출 절대 금지 — **브로커 멱등성 키가 없으므로 무조건**"*. `internal/execgw/indoubt.go:9-12`도 같은 문장을 근거로 삼는다.

`docs/migration/openapi.latest.json` 확인 결과 **사실이 아니다**:

- `OrderCreateRequest.clientOrderId` — "클라이언트 지정 주문 식별자. **멱등성 키로 사용됩니다.** 미전달: 멱등성 미적용. **전달: 동일 값으로 재요청 시 이전 주문 결과를 그대로 재반환합니다.** 서버는 자동 생성하지 않습니다. 최대 36자, `[a-zA-Z0-9\-_]`. **멱등성 키는 10분간 유효**하며, 이후 동일 값으로 재요청 시 새 주문으로 처리됩니다."
- `ConditionalOrderCreateRequest.clientOrderId` — "멱등키 (선택) — 주문 생성 API 와 동일. 동일한 값으로 재요청 시 중복 생성을 방지합니다."
- 두 생성 응답 모두 키를 그대로 되돌려주고, 본문이 다르면 `idempotency-key-conflict` 오류.

그리고 **엔진은 이 키를 쓰지 않는다**: `ClientOrderID` grep에 `internal/execgw`가 한 번도 나오지 않는다. 사람이 CLI에서 직접 넘길 때만(`cmd/tossctl/order.go:555`) 쓰인다. 즉 응답 유실이라는 정확히 그 상황을 위해 설계된 장치를 자동 경로가 비워둔 채, 대신 pagination·안정화 조회·fingerprint 휴리스틱이라는 훨씬 약한 대체물을 만들어 놓았다.

**결정적 비대칭**: 어떤 조회 응답도 `clientOrderId`를 싣지 않는다 — `Order` 스키마와 `ConditionalOrderDetailResponse` 모두 없다. 따라서 멱등키는 **조회 매칭자가 될 수 없고**, 오직 **재요청의 응답**으로만 정체를 회수할 수 있다. 이것이 실제 설계 제약이며 스펙이 명시해야 한다.

**Manager 결정**: 멱등 재생(idempotent replay)을 IN_DOUBT 해소의 1차 절차로 승격한다. 이것은 "오류 시 재시도"가 아니다 — 같은 키의 재요청은 새 주문을 만들 수 없으므로, **정체 회수(identity recovery)**다. P1의 조회 기반 절차는 (a) 유효 창 경과, (b) 능력 검증 미완료, (c) 키 충돌 시의 폴백으로 남는다. 문서 기술은 실측이 아니므로 **능력 attestation 항목**으로 만들고, 검증 전에는 기존 절차만 쓴다.

이 발견은 P1 아카이브 스펙의 사실 오류이므로 `order-execution`을 MODIFIED로 정정한다.

## A. P1과 조립되지 않는 hard break (Eng)

| # | 발견 | 코드 근거 (Manager 검증) | 판정 |
|---|---|---|---|
| A1 | **조건주문 ID가 `broker_order_id`에 들어갈 자리가 없다** | `journal.Dispatch`는 broker order id 없는 ACK를 IN_DOUBT로 돌린다(dispatch.go:126-136) → 무언가는 넣어야 한다. 그런데 `TrackedFillOrders`(fills.go:502)가 CONFIRMED attempt의 `broker_order_id`를 전부 체결 감지 추적 집합에 UNION하고, `filldetect`는 추적 id 조회 실패를 **사이클 전체 실패**로 처리한다(detect.go:386-389). `conditionalOrderId`는 `GET /orders/{id}`에 유효하지 않다 → **첫 보호주문 등록이 전 종목 체결 감지를 영구히 죽인다.** 비워두면 모든 성공한 등록이 IN_DOUBT에 주차된다 | **확인 · 설계 변경** |
| A2 | 같은 ID가 reconcile을 영구 진입 차단으로 몬다 | `LocalStateFromJournal`이 `TrackedFillOrders`로 `OpenOrders`를 만들고(compare.go:85-108), 브로커 미체결 목록에 없는 로컬 open order를 `MissingOrders`로 분류하며(compare.go:299-306) `BlocksEntry()`가 이를 수량 불일치와 동일 취급(compare.go:181) | **확인 · 설계 변경** |
| A3 | **latch 면제의 실패 모드는 주차가 아니라 거짓 CONFIRMED다** | gateway.go:333-345가 심볼 latch를 유일 매칭의 직접 근거로 명시하고 indoubt.go:312-322가 `len(found) > 1`을 UNRESOLVED로 주차한다. 반례: MARKET 청산의 응답 유실 후 stop이 발동하면, matcher는 price=0이라 가격을 판별자로 쓰지 못하고(indoubt.go:609) side·수량·시각 창이 일치해 **발동된 stop의 주문을 청산 attempt의 결과로 CONFIRMED 처리**한다 — 일어나지 않은 매도가 확정되고 그 주문번호가 `NetPositions` 조인에 들어간다 | **확인 · 설계 변경** |
| A4 | RISK_REDUCING 직렬화 키가 정작 필요한 mutation에 존재하지 않는다 | 조건주문 등록은 응답 전까지 조건주문 ID가 없고(conditional_writes.go:43-49), reduce-only 청산 `KindPlace`는 대상 주문번호가 없다(gateway.go:264-277은 cancel/amend에만 설정) | **확인 · 설계 변경** |
| A5 | **"멱등키 없음" 전제가 거짓** | 위 최중대 발견 참조 | **확인 · 스펙 정정** |
| A6 | Open Question 2는 저장소가 이미 답한다 + 조건주문 상태 폴러 태스크가 없다 | `ConditionalOrderCondition.TriggeredOrderID`가 leg별로 존재(conditional_reads.go:20,95; models.go:771-793). 그러나 이를 읽는 주기적 폴러 태스크가 없어 2.2·2.3·2.4·3.3이 전부 공중에 뜬다. §0.4 rate budget도 언급되지 않음 | **확인 · 수용** |
| A7 | 조건주문 cancel/modify가 confirmable ACK를 만들 수 없다 | `CancelConditionalOrder`·`ModifyConditionalOrder`가 out 파라미터에 nil을 넘기고 error만 반환(conditional_writes.go:11-13, 63-65). `classifyMutation`은 `domain.MutationResult`에서 id를 읽으므로(classify.go:22-27) 성공한 취소·정정이 전부 IN_DOUBT로 간다. 어댑터·인터페이스 확장 태스크 없음 | **확인 · 수용** |
| A8 | 봉인 시나리오가 오늘도 거짓이고 태스크로도 참이 되지 않는다 | `Context.TradingService`가 **exported**(engine.go:121-123)이고 `*trading.Service`가 `ConditionalPlace/Cancel/Modify`를 노출한다. 확인 토큰은 호출자가 로컬에서 계산 가능(gateway.go:316-324). `TradingService`를 unexport하지 않으면 정적 증명이 불가능 | **확인 · 수용** |
| A9 | 예약 트랜잭션이 엔진 전체를 단일 커넥션 뒤에 직렬화한다 | `journal.Open`이 `SetMaxOpenConns(1)`(journal.go:108-110). 브로커 조회를 트랜잭션 안에 넣으면 HTTP 왕복 동안 단일 writer를 점유해 `Prepare`·`MarkDispatchStarted`·`Settle`이 전부 막힌다 — 이 change가 지키려는 보호 경로 포함. 태스크 5.2의 `-race` 테스트는 `SetMaxOpenConns(1)` 때문에 트랜잭션 유무와 무관하게 통과하므로 아무것도 증명하지 못한다 | **확인 · 설계 변경** |
| A10 | `KindCancel` 한도 면제가 리터럴에 묶여 있다 | `verifyLimits`가 `plan.kind == journal.KindCancel`로 면제(guardian.go:181-183). 새 kind를 추가하면 조건주문 취소가 한도 검사에 걸려 **§0.3 위반을 이 change가 새로 만든다** | **확인 · 수용** |
| A11 | 조건주문 modify는 새 id를 발급하지 않아 amend 기계가 이전되지 않는다 | amend_indoubt.go:9-14는 "정정은 새 주문번호로 답한다"에 기반. `ConditionalOrderResponse`는 `conditionalOrderId` 하나만 반환 — 후속 주문도 lineage도 잔여도 없다 | **확인 · 설계 변경** |
| A12 | `intents` 행이 조건주문을 담을 수 없다 | schema.go:58-74에 triggerPrice·type·expireDate·second leg 컬럼 없음. `prepareRequest`가 BUY/SELL 아닌 side를 거부(gateway.go:565-569) → OTO 불가. fingerprint는 불변 intent 행에서 필드를 되읽어 매칭하므로(fingerprint.go:35-38, indoubt.go:564-591) **해소 시점에 재계산 불가**. 태스크 1.1의 "additive, v4 호환"은 마이그레이션 하나만큼 과소평가 | **확인 · 설계 변경** |

## B. 남은 fail-open (양 보이스)

| # | 발견 | 판정 |
|---|---|---|
| B1 | **수량 상한이 브로커에 걸린 보호 수량을 빼지 않아 headline hazard가 그대로 남는다** — WATCHING 조건주문은 `triggeredOrderId`가 null이라 브로커 주문이 없고 따라서 예약도 없다. 보유 100 + stop 100 등록 + 청산 100 제출 = 살아있는 SELL 200주 | **확인 · 설계 변경** |
| B2 | 동시 RISK_REDUCING 둘이 각각 상한을 통과한다 — 예약이 진입 측 총계만 덮고 **청산 수량 예약이 없다**. 모호한 진입이 없어도 성립하므로 스펙의 시나리오는 쉬운 경우만 테스트한다 | **확인 · 설계 변경** |
| B3 | **"모든 취소는 RISK_REDUCING"** — 활성 보호주문의 취소는 위험 **증가**인데 latch·한도·HALT_ALL을 모두 면제받는 가장 덜 통제된 mutation이 된다. 판별자는 kind×방향이 아니라 "취소 대상이 이 포지션의 보호인가"이며, 그 레지스트리는 예상 주문 기록이 제공한다 | **확인 · 설계 변경** (codex도 동일 지적) |
| B4 | 낡은 예상 주문이 진짜 외부 주문을 흡수한다 — OCO 패배 leg의 브로커측 자동취소와 `expireDate` 만료는 로컬 mutation 없이 일어나고 폴러 없이는 안 보인다. D2의 위험한 방향은 매칭 실패가 아니라 **거짓 매칭**이다 | **확인 · 수용** |
| B5 | 강화된 인터록이 여전히 진입 가능·청산 불가로 기동한다 — `RequiredEndpoints()`에 `GET /conditional-orders`(해소 조회의 입력)와 `GET /sellable-quantity`(상한의 입력)가 빠졌다. 선행 리뷰 A4와 같은 구멍이 한 층 아래에서 재발 | **확인 · 수용** |
| B6 | 발동 주문이 *주문* 비교 차원에서 영구 외부로 남아 reconcile이 영구히 깨끗하지 않다 → 운영자가 §0.5가 의존하는 채널을 무시하도록 학습된다 | **확인 · 수용** |
| B7 | UNRESOLVED_IN_DOUBT에 예약 해제 경로가 없고, 일일 손실 예약이 거래일 경계에서 소멸하지 않는다 → 주차된 attempt 하나가 다음 날 거래를 조용히 정지시킨다 | **확인 · 수용** |
| B8 | stale 폴백이 로컬 추정치를 권위로 삼아 **더 큰** 수량을 허용한다 — `NetPositions`는 외부 주문을 제외하므로 로컬이 stale-high일 수 있고, reconciliation 메인 계약은 브로커가 최종 권위라고 규정한다 | **확인 · 설계 변경** |

## C. 스펙 품질·태스크

| # | 발견 | 판정 |
|---|---|---|
| C1 | **태스크 4.4와 6.3이 서로 모순** — "위험 감소 결정은 빈 한도" vs "주입된 결정 한도가 감사된 설정 한도와 같음을 테스트". `flatten.decisionFor`가 이미 빈 한도를 쓰므로 실재하는 충돌 | **확인 · 수용** |
| C2 | `Limits`가 "한도 없음"과 "한도 항목 누락"을 구별할 수 없다 — configured 비트만으로는 부족하고 **양성 표지**(결정의 safety class)가 필요 | **확인 · 수용** |
| C3 | 태스크 3.3이 세 개의 미결정 계약을 한 줄에 숨긴다 (보유수량 권위·staleness 임계·보호 수량 차감) | **확인 · 수용** |
| C4 | 태스크 1.5가 존재하지 않는 엔진 Gateway 배선을 전제한다 | **확인 · 수용** |
| C5 | 태스크 1.3이 저장소가 이미 답한 질문을 구현 시점으로 미뤄 잘못된 설계가 나갈 여지를 준다 | **확인 · 수용** |
| C6 | 총계 한도 계산 계약이 이 change에 없는데 이 change가 그 한도를 필수화하고 예약을 걸게 한다 (codex) | **확인 · 수용** — 계산 계약을 이 change로 이동, 수치는 add-core-domain |
| C7 | RiskIntent 재검증의 신뢰 경계 미정의 — 제출자가 위험 데이터를 공급하면 재검증이 순환한다 (codex) | **확인 · 수용** — preimage를 결정과 함께 journal에 영속, Gateway는 journal에서 읽는다 |
| C8 | v5를 세 태스크로 쪼갠 것이 immutable version 계약과 충돌 (codex) | **확인 · 수용** — 단일 원자적 마이그레이션 태스크 |
| C9 | 태스크 2.5의 무조건 green 요구가 이 change가 바꿔야 할 동작을 고정한다 | **확인 · 수용** — 바뀌어야 할 단언과 불변인 단언을 열거 |

## Manager 설계 재작성

**D1 폐기 → 형제 수명주기**: 조건주문은 자체 테이블(`conditional_attempts`, 자체 fingerprint 컬럼, `conditional_order_id`)을 갖는다. `broker_order_id`에는 절대 넣지 않는다. **다리**: leg의 `triggeredOrderId`가 관측되면 **그 주문 id**가 일반 주문 추적·체결·reconcile 경로에 조건주문 lineage와 함께 편입된다. 발동 주문은 일반 주문이고 일반 기계에 속한다. (A1·A2·A11·A12·B6 동시 해소)

**D2 신설 → 멱등 재생**: Gateway가 attempt별 결정적 `clientOrderId`를 발행·영속하고, 해소 1차 절차는 동일 본문 재요청이다. 조회 기반 절차는 폴백. 능력 attestation으로 실측 전에는 사용하지 않는다. (A5·A3·A4 대부분 해소 — 유일성이 latch가 아니라 키에서 온다)

**D3 재정의 → 3 클래스**: EXPOSURE_RAISING / RISK_REDUCING / **PROTECTION_WEAKENING**(활성 보호의 취소·축소, 청산 주문의 취소). 마지막 클래스가 가장 엄격하다 — HALT_ALL에서 금지, 원자적 교체의 일부이고 무보호 창이 유계일 때만 허용. 판별자는 예상 주문 레지스트리다. (B3)

**D4 신설 → 청산 수량 예약**: 진입 측 노출 예약과 대칭. 상한은 `보유 − 미체결 SELL − WATCHING 조건주문 예약 − 동시 예약`의 원자적 총량이며 단건 `min`이 아니다. (B1·B2)

**D5 → 예약 트랜잭션에 네트워크 없음**: 브로커 스냅샷을 트랜잭션 밖에서 수집하고 as-of/버전과 staleness 한계를 트랜잭션 안에서 검증한다. 락 순서와 트랜잭션 스코프 API를 명시한다. 동시성 테스트는 `SetMaxOpenConns(1)`에 의존하지 않는 방식이어야 한다. (A9)

**D6 → 결정에 safety class를 싣는다**: "의도적 면제"와 "부주의한 미설정"을 구별하는 양성 표지. 단일 출처 동등성 테스트는 EXPOSURE_RAISING 결정에만 적용한다. (C1·C2)

**D7 → stale 시 권위는 브로커 스냅샷**: 로컬 추정치를 상한으로 올리지 않는다. 브로커 스냅샷이 아예 없으면 로컬 매도를 하지 않는다 — 그 구간의 보호는 이미 브로커측 조건주문이 담당한다. (B8)

추가 태스크: 조건주문 상태 폴러(+rate budget), cancel/modify 어댑터의 id 반환, `TradingService` unexport, `RequiredEndpoints()`에 두 조회 추가, UNRESOLVED 예약 해제·일일 손실 거래일 경계 소멸, reconcile의 조건주문·발동 주문 취급.

## 후속

- 재작성 후 **다시 proposal-freeze 리뷰**를 받는다. 이번 리뷰는 폐기된 설계에 대한 것이다.
- 멱등키 실측은 `verify-execution-capability`의 필수 항목이다 — 재생 응답이 원본 결과를 돌려주는지, 유효 창이 실제 10분인지, 계좌 스코프인지.
- P1 아카이브 스펙 `order-execution`의 "멱등성 키가 없으므로" 문장은 사실 오류이므로 이 change에서 MODIFIED로 정정한다.

---

# Review 2라운드 (재작성본)

보이스 2개 — codex(8건, critical 4), Eng 적대(20건, critical 8). Manager가 핵심 사실 주장을 코드·openapi에서 재검증.

## 판정

**여전히 착수 불가.** 다만 성격이 1라운드와 다르다 — Eng 평결 그대로: *"round-1은 설계 방향이 틀렸고, 이번은 방향은 옳은데 접합부가 미지정이다. 재작성이 아니라 접합부 사양화로 수렴 가능하다."*

## 1라운드 발견 중 닫히지 않은 것

| # | 내용 | 판정 |
|---|---|---|
| R1 | **발동 주문의 포지션 귀속이 여전히 불가능** — `TrackedFillOrders`(fills.go:504)와 `NetPositions`(fills.go:466)는 `JOIN intents`로 부호를 얻는데, 재작성본 스펙은 발동 주문의 intent 소급 생성을 SHALL NOT으로 금지하면서 시나리오로는 "포지션에 반영된다"를 요구한다. **스펙이 자기모순이다.** 대체 조인(세 번째 UNION arm, `expected_orders`의 leg 방향에서 부호 도출)이 어디에도 없다 | **확인 · 미해소** |
| R2 | **lineage edge를 삽입할 수 없다** — `lineage_edges.intent_id`·`attempt_id`가 둘 다 NOT NULL FK(schema.go:117-128)이고 발동 주문은 둘 다 없다. `relation != "replaces"`도 거부된다(lineage.go:81). v5 태스크가 이 테이블을 건드리지 않는다. 또한 D1의 불변식이 `broker_order_id`에만 걸려 있어 `parent_order_id` 오염을 막지 못한다 | **확인 · 미해소** |
| R3 | **거짓 CONFIRMED가 남는다** — 클래스별 latch는 엔진이 제출하는 mutation만 제한할 뿐 **브로커가 자율 발동시키는 주문을 제한하지 못한다.** 갭 하락에서 reduce-only MARKET SELL 응답 유실 + 같은 수초 내 stop 발동 → 심볼·SELL·수량 일치, 가격은 양쪽 0이라 판별 불가, 시간 창은 **두 사건이 같은 가격 움직임에서 나왔기에** 일치. 해법은 matcher가 예상/발동 주문 id를 배제하는 것인데, 레지스트리는 폴러가 돈 뒤에야 그 id를 알므로 정작 필요한 창에서 경합한다 | **확인 · 미해소** |
| R4 | **일반 주문 경로에 `clientOrderId`가 없다** — `PlaceIntent`·`CanonicalPlace`·`orderCreateV0/V1`·`apiOrderCreateResponse` 전부 필드 없음(조건주문 경로에만 존재). 태스크 1.1이 "이미 필드 존재"라고 **사실과 반대로** 적었다. 귀결: `internal/orderintent/intent.go`·`internal/official/orders_write.go` Pre-Edit 필요(목록에 없음), 그리고 **canonical 딜레마** — 키를 `CanonicalPlace`에 넣으면 모든 CLI confirm token이 바뀌고(§0.2 위반), 빼면 `IntentHash`가 멱등키를 결속하지 않아 재생을 안전하게 만드는 유일한 필드가 인가되지 않은 채 남는다 | **확인 · 미해소** |

## 새로 생긴 구멍

| # | 내용 | 판정 |
|---|---|---|
| N1 | **조건주문 정정 설계가 API를 정반대로 읽었다** — openapi는 정정이 "기존 조건주문을 취소하고 새 조건주문을 생성하며 **새 ID가 발급되고 기존 ID는 무효화**"된다고 명시한다. 1라운드 Eng의 "같은 id 반환" 주장이 틀렸고 내가 그것을 설계에 넣었다. 정정은 일반 amend와 같은 의미이므로 lineage 기계가 이전되며, 응답 유실 시 구·신 ID 양쪽을 다뤄야 한다 | **확인 · 설계 정정** |
| N2 | **청산 예약 + PROTECTION_WEAKENING = 청산 불능**, 그리고 스펙이 그것을 시나리오로 승인한다("청산이 거부되거나"). 전량 보호가 걸린 포지션은 매도 수량이 0이고, 수량을 풀 취소는 허용 기준이 미정이거나 HALT_ALL에서 금지 → **WORKFLOW §0.3 직접 위반** | **확인 · 최중대** |
| N3 | **예약이 일상적으로 누수한다** — 해제 목록이 닫힌 집합인데 "체결되지 않고 장 마감에 브로커에서 만료되는 당일 주문"이 어느 항목에도 없다. 정상 운영 며칠이면 한도가 0까지 조용히 래칫되고, 징후는 "진입이 그냥 멈춤"이다. 필요한 트리거는 **브로커 종결 상태 도달** | **확인 · 최중대** |
| N4 | **재생 유효창 경계에서 중복 주문이 생긴다** — 경과 근거가 `DispatchStartedAt`(전송 *시작*의 로컬 시계)이라 왕복 지연과 시계 오차가 실제 만료를 앞으로 민다. 안전 마진도 시계 권위 규칙도 없다. design.md의 "같은 키의 재요청은 새 주문을 만들 수 없으므로"가 창 밖에서 무조건 거짓 | **확인 · 수용** |
| N5 | **OTO는 하나의 mutation에 두 class가 공존한다**(first BUY, second SELL). 스펙은 mutation당 단일 class만 허용 → RISK_REDUCING이면 flat 계좌의 OTO BUY가 진입 한도와 예약을 우회하고, EXPOSURE_RAISING이면 §0.3 면제를 표현 못 한다 | **확인 · 수용** |
| N6 | **형제 테이블이 latch·재시작 복구에서 빠진다** — `checkSymbolFree`→`PendingAttempts`(recovery.go:30), `UnresolvedAttempts`(resolution.go:85), `RecoverPending`(recovery.go:86)이 전부 `mutation_attempts` 전용. RISK_REDUCING은 두 저장소에 걸쳐 있는데 UNION 요구가 없고, DISPATCH_STARTED에서 멈춘 조건주문 attempt는 복구되지도 않는다 | **확인 · 수용** |
| N7 | **멱등 재생이 Resolver의 구조적 불변식을 뒤집는다** — `indoubt.go:9-12`가 "the Resolver has no trading service, no broker writer, no submit path"를 파일 전체의 근거로 명시한다. 재생은 정확히 그 writer를 요구하고, 봉인 논증의 두 번째 문이 된다. 또 attempt 상태기계에 "재생 진행 중" 상태가 없다(IN_DOUBT는 재진입 불가) | **확인 · 수용** |
| N8 | **재생이 "동일 본문"을 요구하면서 정확한 wire body를 영속하지 않는다** — 구조화 intent에서 재구성하면 바이너리 버전·기본값·직렬화 규칙 변화로 본문이 달라질 수 있다. RECORDED에 canonical wire body 또는 digest+serializer version을 불변 저장해야 한다 | **확인 · 수용** |
| N9 | OCO 이중 예약(수량은 조건주문 단위인데 `expected_orders`는 leg 단위), 발동 leg 이중 계상(예상 주문 소비가 폴러 관측 시점이라 그 창에서 두 번 차감), 미체결 매도의 잔량/원수량 규정 부재 — 셋 다 방향이 같다: **청산 거부** | **확인 · 수용** |
| N10 | `sellableQuantity`가 공식에서 빠졌다 — 같은 change가 그 endpoint를 기동 필수로 요구하면서 계산에서 버린다. `flatten/liquidate.go:329-343`은 담보·미결제로 매도가능이 보유보다 작을 수 있어 `min(sellable, held)`를 쓴다고 이미 명시 | **확인 · 수용** |
| N11 | PROTECTION_WEAKENING 분류 권위가 불완전 — 일반 Gateway가 낸 미체결 청산 주문은 `expected_orders`에 없고, 발동→관측 창의 보호 주문도 없고, 이 change 이전 등록분과 운영자가 앱에서 만든 것도 없다. **불완전한 레지스트리를 권위로 선언한 것 자체가 fail-open.** 필요한 규칙: "보호가 아님을 증명할 수 없는 취소는 PROTECTION_WEAKENING" | **확인 · 수용** |
| N12 | 폴러에 pagination 완주·status 커버리지 요구가 없는데 이제 포지션 종결의 임계 경로에 있다. 2페이지 이후 발동 주문이나 조회하지 않는 status는 영영 관측되지 않아 포지션이 닫히지 않는다 | **확인 · 수용** |
| N13 | 태스크 0.1이 테이블 **이름 5개**만 주면서 3.x·5.x·6.x의 입력인 컬럼 사양을 전부 숨긴다. immutable migration이라 잘못 만들면 못 고친다 — **스키마는 design.md에 있어야 하고 0.1은 전사여야 한다** | **확인 · 수용** |
| N14 | 태스크 8.2가 결정 영속 배선 전체를 숨기고, 명시된 해법이 순환을 제거하지 않고 **이동**시킨다(제출자는 여전히 결정 시점에 위험 데이터를 공급한다 — 영속은 TOCTOU만 닫는다). "위험 입력은 제출자가 통제하지 않는 권위에서 온다"는 요구가 없다 | **확인 · 수용** |
| N15 | 예약 재수집 루프에 상한·데드라인·종단 fail-closed 없음. 예약 산술이 float64인데 안전 속성이 "합이 보유를 넘지 않는다"라 오차가 누적된다(소수점 US 주문에서 실재) | **확인 · 수용** |
| N16 | `verify-execution-capability/tasks.md`의 교차 참조가 어긋남(6.1 → 실제 9.3). 그리고 태스크 2.7의 "유효 창 10분 확인"은 **의도적으로 라이브 중복 주문을 만드는 절차**인데 그 사실도 범위도 거부 시 결과도 적혀 있지 않다 | **확인 · 수용** |

---

# 3판 재범위 결정 (2026-07-26)

2라운드 평결("방향은 옳으나 접합부 미지정")과 사용자 검토를 반영해 경계를 **측정 의존성**으로 재설정했다. 이 change(2a)에서 조건주문·보호주문·발동 주문 귀속·청산 수량 예약을 전부 제거 — 2b 측정(멱등키 실동작·조건주문 속성·CANCEL_REJECTED 형태·매도가능수량 의미) 후 2c `add-protection-orders`로 작성한다. 2라운드 발견 중 R1~R3·N1~N2·N5·N9~N12(조건주문 관련)는 2c 소관으로 이관, N3(예약 누수)·N4(TTL 마진)·N6~N8·N13~N16과 R4(clientOrderId 배선 부재)는 3판에 반영했다. 추가 채택: opaque 식별자 규칙, OrderStatus 10개 파생(CANCEL_REJECTED 별도 레코드 fail-closed), EXECUTION_CORRECTION 이벤트, RECONCILE 행동 제한 상태(확정 하한 위험 축소 허용), 브로커 종결 상태 기반 예약 해제.

---

# Review 3라운드 (3판)

보이스 2 — codex(12건), Eng 적대(24건 + 최소 수정 11항목). Manager 표본 검증: 만료≠OrderStatus(derive.go 17행이 이미 "Expiry would look like this → UNKNOWN"), TrimSpace 저장(resolution.go:47), filledAmount 미저장(payload.go), **PARTIAL_FILLED가 OPEN·CLOSED 양쪽 그룹에 명시**(openapi status param — 가장 흔한 IN_DOUBT 사례가 dedup 없이는 2중 매칭→영구 주차), **422 opposite-pending-order-exists**(동시 RISK_REDUCING 조항을 브로커가 거부), **409 request-in-progress**(재생의 최빈 응답, 규칙 부재) — 전부 사실.

## 판정

**경계는 지켜졌다**(조건주문 의무 0건 — Eng C-boundary 확인). 그러나 접합부 결함으로 착수 불가 유지. Eng 평결: "round 1과 달리 고칠 수 있는 판". 4판은 재설계가 아니라 접합부 사양화다.

## 4판 반영 결정 (36건 통합)

**스키마(D9 확장)**: decisions에서 intents FK 제거(발급 시점에 intent 미존재 — codex#1·Eng A1/A3), decision_id는 발급자 발행, 멱등키는 `f(decision_id, generation)`로 재유도(발급자가 계산 가능한 값만 사용); mutation_attempts에 decision_id·safety_class·generation 컬럼(attempt→결정 결속 — codex#2); risk_reservations에 attempt_id(예약↔종결 관측 조인 — codex#3); fill_snapshots에 filled_amount(정정 감지의 prev — Eng B6); reconcile_states 테이블(영속 — codex#5); execution_corrections 완전 사양(PK·FK·dedup·동일 트랜잭션); decisions.account_ref; spent_nonces 보존 불변식(보존 ≥ 최대 결정 TTL); generation은 2a에서 0 고정(전진 주체는 2c/2d).

**동시 carve-out 폐기**: "미해소 EXPOSURE_RAISING이 RISK_REDUCING을 차단하지 않는다"를 2a에서 제거. 근거: (1) 422 opposite-pending — 브로커가 동시 반대 주문을 거부하고 P1 분류기는 422를 확정 거부로 종결해 §0.3 의도가 반전됨(Eng B2), (2) 부재 증명 오염 — 동시 청산이 잔고 delta 교차확인을 무효화(codex#8), (3) AttemptRecord에 class가 없어 데이터 경로 부재(Eng A4). 2a의 §0.3 = flatten의 순서화된 saga(P1 보존) + 해소 우선·유계(재생이 이를 빠르게 함). 보호주문의 동시성 요구는 2c에서 브로커 계약 위에 재설계.

**caller-forgeable class 차단**: Gateway가 mutation 형태에서 raisesExposure를 독립 계산해 class와 일치 검증(EXPOSURE_RAISING ⇔ 노출 증가) — 발급자 부재 상태에서 class 선언만으로 한도 면제 불가(Eng B3).

**flatten의 1급 답**: RISK_REDUCING preimage는 축소 의도(계좌·심볼·방향·상한 수량·사유)로 클래스별 스키마를 갖는다 — 퇴화 carve-out 없이 preimage NOT NULL 유지. flatten이 자기 결정 행을 journal에 기록 후 제출(Eng A2).

**해소 절차**: 조회 폴백에 orderId dedup 의무(PARTIAL_FILLED 양쪽 그룹 인용 — Eng B1); 재생 진입점의 자기 방어 의무(state==IN_DOUBT·회당 시간 재검사·상한·attestation — Eng B4); 재생 경로에 ClassifyHTTPMutation 금지, 409=대기·상한 미소비, 422 키충돌=FAILED 아님·UNRESOLVED+알림(Eng B5); 동일 심볼 mutation이 관측 창에 개입하면 잔고 교차확인 무효→자동 FAILED 금지(codex#8); 복구 호출 그래프 명시(recovery→Gateway 재생→Resolver 폴백 — codex#6).

**예약**: 해제는 파생된 종결 상태만 — "미체결 만료 포함" 문구 삭제(만료의 status 표현은 [미측정 — 2b 2.1]; CLOSED+fill0+무취소는 UNKNOWN 유지); UNKNOWN이 잡은 예약의 운영자 해제 경로 신설(Eng B-high); nonce 소비를 MarkDispatchStarted 트랜잭션에 병합(핫패스 직렬 쓰기 1회 — Eng B-medium); 종결 관측의 생산자는 journal 트랜잭션 결합으로 정의, filldetect 상시 배선은 엔진 루프 소유 change로(Eng A-high).

**RECONCILE**: journal이 권위, EntryGate는 기동 시 재구성되는 in-memory 투영으로 관계 명시(Eng B-medium); 확정 하한 = `max(0, min(신선한 브로커 보유, 매도가능) − 로컬 미체결 SELL)`, 스냅샷 부재 시 0(자동 경로; 수동 flatten은 자체 신선 조회로 무변경). 매도가능 의미는 [미측정]이나 min()에서 하한을 낮추는 방향으로만 사용 — 보수적(Eng B-high·C-high).

**기타**: opaque 규칙 재서술(trim은 공백 검사만, 저장은 원문 — 위반 3개소 명시 수정), round-trip 실패 상태 = IN_DOUBT, brokerstate 6.2는 확장이 아니라 재작성(2값 raw status 전제 — 4.2의 선행 의존), Retry Matrix 요구를 MODIFIED에 추가("재생은 정체 회수" 단일 서술 + retry.go 앵커), 총계 계약에 2a 구조 결정 명시(자동 진입 LIMIT 전용·gross long·실현 손실 기준·보수 기본 staleness — 2d가 보수 방향으로만 대체), 8.1 분할(계좌 해석·journal 편입·Gateway 구성·인터록 순서·flatten 자체 배선), 2c 선행 의무 문구를 인터록 SHALL 밖으로.
