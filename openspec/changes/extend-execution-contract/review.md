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
