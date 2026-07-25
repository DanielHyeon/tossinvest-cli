# Design: extend-execution-contract

> 2026-07-26 재작성. 초판 설계(조건주문 = 새 MutationKind)는 proposal-freeze 리뷰에서 P1과 조립되지 않음이 확인되어 폐기했다. 근거·발견 전문은 `review.md`.

## Context

P1은 place/cancel/amend에 대해 완결된 안전 계약을 만들었다: journal 선기록 → DISPATCH_STARTED → 분류 → IN_DOUBT 해소, GuardianDecision 재검증, 심볼당 in-flight 1개, raw mutator 봉인. 조건주문은 그 계약 밖에 있고, 무인 자동 보호 경로로 쓰려면 안으로 들여와야 한다.

리뷰가 두 가지를 새로 확정했다. 첫째, **조건주문은 주문이 아니다** — 식별자 네임스페이스·수명주기·응답 형태가 다르고 그 실행은 별개의 주문을 만든다. 둘째, **브로커에 멱등키가 있다** — P1이 "없다"고 전제한 것과 달리 `clientOrderId`가 두 생성 엔드포인트 모두에 문서화되어 있고, 엔진은 그것을 쓰지 않고 있다.

## Goals / Non-Goals

**Goals**: 조건주문의 형제 수명주기와 발동 주문 다리, 멱등 재생 기반 정체 회수, 3-클래스 safety 분류와 클래스별 직렬화, 청산 수량 예약, 결정의 위험 입력 결합, 한도·게이트 fail-closed.

**Non-Goals**: Guardian 판정 체인·한도 **수치**(add-core-domain), 포지션 aggregate·보호 saga(add-core-domain), 실계좌 능력 검증(verify-execution-capability), MCP 표면의 Gateway 우회(P4까지 문서화된 잔존 리스크).

## Decisions

### D1. 조건주문은 형제 수명주기다 — 같은 테이블에 얹지 않는다

조건주문 mutation은 자체 journal 테이블(`conditional_attempts`)과 자체 식별자 컬럼(`conditional_order_id`)을 갖는다. `conditionalOrderId`를 `broker_order_id`에 넣지 않는다.

넣으면 안 되는 이유가 코드로 확인됐다: `TrackedFillOrders`가 CONFIRMED attempt의 `broker_order_id`를 체결 감지 추적 집합에 UNION하고(fills.go:502) 감지기는 추적 id 조회 실패를 사이클 전체 실패로 처리한다(detect.go:386-389). `conditionalOrderId`는 `GET /orders/{id}`에 유효하지 않으므로 **첫 보호주문 하나가 전 종목 체결 감지를 죽인다.** 같은 값이 `LocalState.OpenOrders`를 거쳐 `MissingOrders`가 되고 영구 진입 차단을 만든다(compare.go:85-108, 299-306, 181).

수명주기 **단계**는 P1과 같다(RECORDED → DISPATCH_STARTED → ACKED/IN_DOUBT → 종결) — 그 단계는 "요청은 보냈는데 응답을 못 받았다"를 다루는 것이고 그 문제는 주문 유형과 무관하다. 공유하는 것은 단계와 원칙이지 테이블이 아니다.

조건주문 intent는 자체 컬럼을 갖는다: 유형(SINGLE/OCO/OTO), 만료일, leg별 (방향·트리거가·주문가·유형). `intents` 행은 단일 side·수량·가격만 가지며 `prepareRequest`가 BUY/SELL 아닌 side를 거부하므로(gateway.go:565-569) OTO를 표현할 수 없고, fingerprint는 불변 intent 행에서 필드를 되읽어 매칭하므로(fingerprint.go:35-38) 컬럼 없이는 해소 시점 재계산이 불가능하다.

### D2. 발동 주문 다리 — 발동된 것은 일반 주문이고 일반 기계에 속한다

조건주문 등록이 확정되면 leg별 예상 주문 레코드를 남긴다(조건주문 ID·leg 식별·심볼·방향·최대 수량·상태). 조건주문 상태 폴러가 `GET /api/v1/conditional-orders`를 주기적으로 읽어 leg의 `first.status`/`second.status`/`triggeredOrderId`를 관측한다. `triggeredOrderId`가 나타나면 **그 주문 id**를 일반 주문 추적 집합·체결 경로·reconcile에 편입하고 조건주문으로의 lineage를 기록한다.

`TriggeredOrderID`는 leg별로 실재하며(conditional_reads.go:20, models.go:771-793) openapi가 "일반 주문 API에 그대로 사용할 수 있습니다"라고 명시한다. 따라서 휴리스틱 매칭은 필요 없다 — 초판 설계의 Open Question은 저장소가 이미 답하고 있었다.

**예상 주문은 일회성이고 유계다**: leg가 발동하면 그 레코드는 발동 주문 id로 소비되어 더 이상 매칭에 쓰이지 않는다. OCO 패배 leg의 브로커측 자동 취소와 `expireDate` 만료는 폴러가 관측해 레코드를 종결시킨다. 종결은 **체결 귀속이 끝난 뒤** 수행하고 tombstone을 보존 기간 동안 남긴다. 이렇게 하지 않으면 낡은 예상 주문이 운영자의 진짜 수동 매도를 엔진 포지션으로 흡수한다 — D2의 위험한 방향은 매칭 실패가 아니라 거짓 매칭이다.

폴러는 `ORDER_INFO` rate limit 그룹을 체결 감지 루프와 공유하므로 예산을 명시하고(§0.4), 폴러 실패가 보호 경로를 막지 않는다.

### D3. 멱등 재생을 정체 회수의 1차 절차로 쓴다

Gateway는 attempt별 결정적 `clientOrderId`(attempt 식별자에서 파생, ≤36자, `[a-zA-Z0-9\-_]`)를 발행해 **DISPATCH_STARTED 이전에 영속**한다. 응답이 유실되면 해소의 첫 단계는 **동일 본문·동일 키 재요청**이고, 응답의 `orderId`/`conditionalOrderId`가 권위다.

이것은 재시도가 아니다. 같은 키의 재요청은 새 주문을 만들 수 없다("동일 값으로 재요청 시 이전 주문 결과를 그대로 재반환"). P1이 금지한 것은 **정체를 모르는 채 다시 주문을 내는 것**이고, 멱등 재생은 정확히 그 반대다.

제약이 하나 있고 그것이 설계를 결정한다: **어떤 조회 응답도 `clientOrderId`를 싣지 않는다**(`Order`, `ConditionalOrderDetailResponse` 모두). 따라서 멱등키는 조회 매칭자가 될 수 없고 오직 재요청 응답으로만 정체를 회수한다.

폴백 조건(P1의 fingerprint·pagination·안정화 절차를 그대로 사용): 유효 창(문서 기준 10분) 경과, 능력 검증 미완료, `idempotency-key-conflict`. cancel/modify는 멱등키를 받지 않으므로 항상 조회 절차를 쓴다.

**문서는 실측이 아니다.** 재생 응답이 원본 결과를 돌려주는지, 창이 실제로 10분인지, 키가 계좌 스코프인지는 `verify-execution-capability`의 필수 항목이며, 검증 전에는 재생을 사용하지 않고 P1 절차만 쓴다.

### D4. Safety class는 셋이다

- **EXPOSURE_RAISING**: 진입 제출, 노출 증가
- **RISK_REDUCING**: 보호주문 생성·증량, reduce-only 청산, **미체결 진입의 취소**
- **PROTECTION_WEAKENING**: 활성 보호주문의 취소·수량 축소, 청산 주문의 취소

초판의 "모든 취소는 RISK_REDUCING"은 틀렸다. 활성 손절의 취소는 보호를 제거하므로 위험 **증가**이며, 그 분류대로면 시스템에서 가장 덜 통제된 mutation이 된다(latch·한도·HALT_ALL 전부 면제). 판별자는 kind×방향이 아니라 **취소 대상이 이 포지션의 보호인가**이고, 그 레지스트리가 D2의 예상 주문 기록이다.

PROTECTION_WEAKENING은 가장 엄격하다: HALT_ALL에서 금지, 원자적 교체의 일부이며 무보호 창이 유계로 측정될 때만 허용, audit 필수.

직렬화:
- EXPOSURE_RAISING: 심볼당 1건 (P1 규칙 유지)
- RISK_REDUCING: EXPOSURE_RAISING에 막히지 않는다(§0.3). **자기들끼리는 심볼당 1건** — 초판이 "대상 식별자 단위"라 했지만 보호주문 생성과 reduce-only 청산에는 제출 시점에 그런 식별자가 없다. 두 개의 독립된 심볼 latch(클래스당 하나)가 정답이다
- PROTECTION_WEAKENING: 대상 조건주문 ID 단위(이 클래스는 항상 대상이 있다)

멱등 재생(D3)이 활성이면 유일 매칭이 키에서 오므로 latch는 유일성 근거가 아니라 동시성 제어로 남는다. 재생이 미검증이면 latch가 P1과 같은 유일성 근거로 작동한다 — 그래서 클래스당 1건이 필요하다.

### D5. 청산 수량 예약 — 상한이 아니라 예약이다

`min(확정 보유, 매도가능)`은 **단건** 상한이라 총 매도 의무를 막지 못한다. 보유 100주에 stop 100주 등록과 청산 100주 제출이 각각 상한을 통과해 살아있는 매도 200주가 된다. WATCHING 조건주문은 `triggeredOrderId`가 null이라 브로커 주문이 존재하지 않고 따라서 브로커가 예약하는 것도 없다.

진입 측 노출 예약과 대칭으로 **청산 수량 예약**을 둔다:

```
가용 매도 수량 = 보유 − 미체결 SELL − WATCHING 조건주문의 예약 수량 − 유효한 동시 예약
```

원자적 총량이며 단건 min이 아니다. 브로커 응답에 reduce-only 필드가 없으므로(orders_write.go) 이 계산이 유일한 방어다.

**stale 시 권위는 브로커 스냅샷이다.** 초판은 계좌 조회가 stale이면 로컬 확정 보유수량을 쓰라고 했는데 방향이 반대다 — `NetPositions`는 외부 주문을 제외하므로 로컬이 실제보다 클 수 있고, reconciliation 메인 계약은 브로커를 최종 권위로 규정한다. 규칙: 가장 최근 브로커 스냅샷을 쓰고 staleness 한계와 critical 알림을 붙인다. 브로커 스냅샷이 아예 없으면 로컬 매도를 하지 않는다 — 그 구간의 보호는 이미 브로커측 조건주문이 담당한다(§0.3 만족).

### D6. 예약 트랜잭션에 네트워크를 넣지 않는다

`journal.Open`은 `SetMaxOpenConns(1)`이다(journal.go:108-110). 브로커 조회를 `BEGIN IMMEDIATE` 안에 넣으면 HTTP 왕복 동안 단일 writer를 점유해 `Prepare`·`MarkDispatchStarted`·`Settle`이 전부 막힌다 — 이 change가 지키려는 보호 경로까지.

절차: 브로커 스냅샷을 트랜잭션 **밖에서** 수집 → 트랜잭션 안에서 as-of/버전과 staleness 한계를 검증하고 예약을 삽입 → 조건 불충족이면 롤백하고 재수집. 트랜잭션 스코프 내부 API(기존 journal 메서드 재진입 금지)와 락 순서를 명시한다. 동시성 테스트는 `SetMaxOpenConns(1)`이 이미 모든 문장을 직렬화하므로 `-race`만으로는 아무것도 증명하지 못한다 — 스냅샷 as-of 조건이 실제로 재검증을 유발하는지를 직접 검증해야 한다.

### D7. 결정은 safety class와 RiskIntent preimage를 싣는다

`Limits`의 configured 비트만으로는 "의도적 면제"(위험 감소)와 "부주의한 미설정"(진입)이 구별되지 않는다 — 둘 다 모든 비트가 false다. 결정에 **safety class**를 실어 양성 표지로 삼는다. 단일 출처 동등성 검증(주입된 결정 한도 == 감사된 설정 한도)은 EXPOSURE_RAISING 결정에만 적용한다. 초판의 태스크 4.4와 6.3은 이 구분 없이는 서로 모순이었다.

`KindCancel` 리터럴에 묶인 한도 면제(guardian.go:181-183)를 **safety class 기준으로 재작성**한다. 그러지 않으면 새 조건주문 취소 kind가 한도 검사에 걸려 이 change가 §0.3 위반을 새로 만든다.

RiskIntent(계좌·시장·심볼·방향·진입가·손절가·목표가·수량·정책 버전)는 canonical 해시로 결정에 결합되고, **preimage를 결정과 함께 journal에 영속**한다. Gateway는 제출 시 호출자가 준 위험 데이터가 아니라 **journal에서 읽은 preimage**로 재검증한다 — 제출자가 공급한 값으로 재계산하면 검증이 순환한다.

### D8. 총계 한도의 계산 계약은 여기에 속한다

정의되지 않은 양에 예약을 걸 수는 없다. 총 개방 노출·일일 손실·현금의 **계산 계약**(권위 데이터, 미체결·조건주문의 평가 가격, 통화 정규화와 FX 권위·staleness, 실현/미실현 범위, 시장별 거래일 경계, 예약 합산, stale 시 fail-closed)을 이 change에서 정의한다. **수치**는 add-core-domain이 정한다.

일일 손실 예약은 거래일 경계에서 소멸해야 한다. 그러지 않으면 주차된 attempt 하나가 다음 날 거래를 조용히 정지시킨다.

### D9. journal v5는 하나의 원자적 마이그레이션

`conditional_intents`, `conditional_attempts`, `expected_orders`, `risk_reservations`(진입·청산 양방향), `spent_nonces`를 **한 번의** v5 마이그레이션으로 추가한다. 세 태스크로 쪼개면 같은 스키마 버전이 서로 다른 구조를 뜻하게 되어 immutable version 계약과 충돌한다.

**롤백은 구버전 바이너리 실행이 아니다** — `ErrSchemaTooNew`로 기동이 거부된다. 복구 경로는 마이그레이션 직전 자동 백업으로의 복원이며 절차와 테스트를 만든다.

## Risks / Trade-offs

- [멱등 재생이 문서와 다르게 동작] → 능력 검증 항목, 미검증 시 P1 절차만 사용. 재생 없이도 설계가 성립하도록 클래스당 심볼 latch 유지
- [조건주문 상태 폴러가 rate budget을 소비] → 예산 명시, 폴러 실패가 보호 경로를 막지 않음, 체결 감지 SLO와의 우선순위 정의
- [D2가 reconcile의 외부 주문 판정을 바꾼다] → 바뀌어야 할 단언과 불변인 단언을 태스크에 열거(무조건 green 요구 금지)
- [`TradingService` unexport가 기존 소비자를 깬다] → 엔진 프로필 한정, CLI 표면 무영향을 characterization 테스트로 고정

## Migration Plan

journal v5 단일 원자 마이그레이션. 직전 자동 백업 → 실패 시 복원. 스키마 계약 테스트로 v4→v5 전이와 구버전 거부를 고정.

## Open Questions

- 멱등 재생의 실제 응답(원본 결과 vs 오류), 유효 창, 계좌 스코프 — verify-execution-capability 필수 항목
- 조건주문 modify는 같은 `conditionalOrderId`를 반환하므로 amend lineage가 이전되지 않는다. 정정 IN_DOUBT 해소는 leg의 트리거가·수량 재조회로 판정하는 별도 절차가 된다 — 구현 시 절차 확정
