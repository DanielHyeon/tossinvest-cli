# Design: extend-execution-contract

> 2026-07-26 3판. 1·2판의 발견 전체는 `review.md`. 3판의 경계 원칙: **이 change는 측정 없이 확정 가능한 것만 담는다.** 조건주문·보호주문·발동 주문 귀속·청산 수량 예약은 어떤 형태로도 이 change에 들어오지 않는다(2c `add-protection-orders`, 2b 측정 후 작성). 브로커 동작에 관한 모든 서술은 `docs/migration/openapi.latest.json` 인용을 달거나 `[미측정]`으로 표시한다.

## Context

P1은 place/cancel/amend에 대해 journal 선기록 → DISPATCH_STARTED → 분류 → IN_DOUBT 해소, GuardianDecision 재검증, 심볼당 in-flight 1개, raw mutator 봉인을 만들었다. 리뷰 3라운드가 확정한 사실:

- **브로커에 멱등키가 있다**: `OrderCreateRequest.clientOrderId` — "멱등성 키로 사용됩니다 … 동일 값으로 재요청 시 이전 주문 결과를 그대로 재반환합니다 … 멱등성 키는 10분간 유효" (openapi). P1 스펙의 "브로커 멱등성 키가 없으므로"는 사실 오류이고, 엔진 경로는 이 필드를 쓰지 않는다(`PlaceIntent`·`orderCreateV0/V1`에 필드 부재).
- **어떤 조회 응답도 `clientOrderId`를 싣지 않는다**(`Order`·`OrderResponse`만 echo): 멱등키는 조회 매칭자가 될 수 없고 재요청 응답으로만 정체를 회수한다.
- **`orderId`는 opaque token**: 계약에 형태·패턴이 없다(64자 난수 example). 형태 검증 금지.
- **OrderStatus는 10개가 문서화**되어 있고, P1 스펙의 "OPEN/CLOSED 수준" 전제보다 풍부하다. 특히 `CANCEL_REJECTED`/`REPLACE_REJECTED`는 "별도 주문 레코드로 생성됨"(openapi) — 로컬 intent 없는 레코드가 정상 경로에서 생긴다.
- **체결은 누적 모델**: `OrderExecution{filledQuantity, averageFilledPrice(부분 체결 시 평균), filledAmount, filledAt, commission}` — 개별 체결 ID 없음. P1 journal이 이미 watermark 규칙(delta>0 전진 / 동일 no-op / delta<0 fail-closed)을 구현하고 있다(fills.go).

설계 원칙(StockOS 실행 무결성 분석에서 채택, 토스 계약에 맞춰 재구현): **불확실성은 상태로 격리하고, 불일치는 계산하지 않고 RECONCILE로 멈추고, 결정은 영속 후 실행하며, 브로커가 보증하지 않는 것은 코드 타입으로 표시한다.**

## Goals / Non-Goals

**Goals**: 결정 계약 강화(safety class·RiskIntent preimage·generation), 멱등키 기계(발급·영속·TTL 마진·재생 골격), 진입 측 위험 예약, RECONCILE 상태 의미, 한도 fail-closed + 총계 계산 계약, 브로커 식별자·상태 취급 정정, 엔진 Gateway 배선·봉인, journal NonceStore. 전부 httptest + 합성 결정으로 테스트 가능하다.

**Non-Goals**: 조건주문 일체(2c), 발동 주문 귀속·격리 원장(2c), 청산 수량 예약·PROTECTION_WEAKENING(2c), 멱등 재생의 실사용 활성화(2b attestation 후), Guardian 판정 체인·한도 수치(2d), 전략(P3).

## Decisions

### D1. 결정은 영속 후 실행 — preimage와 generation을 싣는다

GuardianDecision은 제출 전에 journal에 영속된다: RiskIntent preimage(원문 JSON), 그 canonical hash, safety class, 한도 스냅샷, nonce, 만료, generation. Gateway는 제출 직전 **journal에서 읽은** preimage로 해시를 재계산해 주문 파라미터와 대조한다 — 호출자 공급 값으로 재검증하면 순환한다(2라운드 C7).

영속이 닫는 것은 결정→제출 사이 TOCTOU다. "위험 입력이 제출자가 통제하지 않는 권위에서 온다"는 것은 발급자(2d Guardian)의 계약이며, 이 change는 그 자리를 인터페이스로 남긴다.

`generation`은 같은 intent의 교체·재발급 순서를 매긴다(StockOS OcoAggregate.generation 채택). 결정·attempt·예약이 같은 generation을 참조한다.

### D2. 멱등키 = 결정에 결속된 결정적 `clientOrderId`

`clientOrderId = deterministic(intent_id, generation)` (≤36자, `[a-zA-Z0-9\-_]`, openapi 패턴). **RECORDED 단계에서** attempt 행에 영속되고, canonical wire body(직렬화 결과 그대로) + serializer version도 함께 불변 저장된다 — 재생은 저장된 body만 사용한다(2라운드 N8: 구조화 필드에서 재구성하면 바이너리 버전에 따라 본문이 달라질 수 있다).

**canonical 딜레마의 해소(2라운드 R4)**: 멱등키를 `CanonicalPlace`(확인 토큰의 입력)에 넣지 않는다 — 넣으면 모든 CLI confirm token이 바뀐다(§0.2). 대신 **결정이 키를 명시 필드로 싣고**, Gateway가 `attempt.client_order_id == decision.client_order_id`를 재검증한다. 키는 canonical 밖에서 인가된다.

일반 주문 경로 배선: `orderintent.PlaceIntent`·`official.orderCreateV0/V1`에 필드 추가(upstream 수정 — Pre-Edit 대상), CLI 경로는 무변경(키 미전달 = 멱등성 미적용, openapi가 허용).

### D3. 멱등 재생은 TTL 마진 안에서만, attestation 뒤에서만

IN_DOUBT 해소의 1차 절차는 동일 키·저장된 wire body의 재요청이고 응답의 `orderId`가 권위다. 단:

- **시간 규칙**: `elapsed(dispatch_started_at) < TTL − margin`일 때만. TTL은 문서상 10분 `[미측정 — 2b]`, margin 기본 60초(왕복 p99 + 시계 오차, 2b에서 실측 조정). 경과 근거가 전송 시작의 로컬 시계이므로 마진 없는 경계 사용은 창 밖 재생 = 새 주문이 된다(2라운드 N4).
- **활성화 규칙**: 재생 실동작(원본 결과 재반환·유효 창·계좌 스코프)이 2b attestation으로 확인되기 전에는 이 경로를 타지 않는다. 코드·테스트는 지금 만들고 활성화 플래그는 attestation이 켠다.
- **폴백**: 창 초과·미검증·`idempotency-key-conflict`·키를 받지 않는 mutation(취소·정정)은 P1 조회 절차(fingerprint·pagination·안정화). 두 절차는 서로를 대체하지 못한다 — 조회 응답에 키가 없기 때문.
- **재생 자체의 응답 유실**: 같은 attempt에 재생 시도 횟수·시각을 append 기록하고, 상한(기본 2회) 초과 시 조회 절차로 전환. 재생은 새 attempt를 만들지 않는다 — 같은 attempt의 해소 단계다(2라운드 N7: "재생 진행 중"은 attempt 상태가 아니라 해소 절차의 내부 단계이며, journal에는 replay_attempts 카운터·시각만 기록).
- **재생 writer의 봉인**: Resolver는 P1 불변식대로 mutator를 갖지 않는다. 재생은 Gateway의 해소 전용 진입점(저장된 body 전송만 가능, 새 본문 구성 불가)으로 수행하고, 그 진입점이 두 번째 제출 문이 되지 않음을 정적 테스트로 증명한다.

### D4. Safety class는 이 change에서 둘, enum은 셋

결정은 `safety_class`를 명시 필드로 싣는다: **EXPOSURE_RAISING**(진입 제출) / **RISK_REDUCING**(reduce-only 청산, 미체결 진입의 취소). enum 값 **PROTECTION_WEAKENING**은 예약만 한다 — 발급자·소비자는 2c(보호주문이 존재해야 의미가 있다).

한도 면제는 `KindCancel` 리터럴이 아니라 class 기준으로 재작성한다(guardian.go:181-183). 빈 한도 스냅샷은 "의도적 면제"의 표지가 아니다 — class가 양성 표지다: EXPOSURE_RAISING은 필수 한도 전부 설정된 스냅샷 없이는 거부되고, RISK_REDUCING은 스냅샷을 싣지 않는다.

직렬화는 P1 규칙 유지: 심볼당 in-flight 1개. 이 change에는 조건주문이 없으므로 클래스별 latch 분리가 필요 없다 — RISK_REDUCING(청산·취소)이 EXPOSURE_RAISING의 IN_DOUBT에 막히지 않아야 한다는 §0.3 요구만 반영한다: **미해소 EXPOSURE_RAISING은 같은 심볼의 RISK_REDUCING을 차단하지 않되, RISK_REDUCING의 수량은 RECONCILE 규칙(D6)을 따른다.**

### D5. 진입 측 위험 예약 — 네트워크는 트랜잭션 밖

총 개방 노출·일일 손실·현금의 판정과 예약은 하나의 journal 트랜잭션에서 원자적으로 수행한다. 브로커 스냅샷은 트랜잭션 **밖에서** 수집하고(journal은 `SetMaxOpenConns(1)` — 안에 네트워크를 넣으면 모든 mutation 기록이 막힌다), 안에서 스냅샷 as-of·staleness를 검증한 뒤 삽입하며, 불충족이면 롤백·재수집한다. **재수집은 상한(기본 3회)·총 데드라인을 갖고 초과 시 fail-closed 거부**(2라운드 N15).

예약 해제 트리거(닫힌 목록이 아니라 **브로커 종결 상태 도달**이 정본):
- attempt가 브로커 종결 상태(FILLED / CANCELED / REJECTED / NOT_DISPATCHED / FAILED_CONFIRMED)에 도달 — **미체결 만료 포함**(2라운드 N3: 당일 주문의 장 마감 만료가 어느 트리거에도 없어 일상 누수)
- nonce **미소비** 상태의 결정 만료 (소비 후 만료는 주문이 접수됐을 수 있으므로 해제하지 않는다)
- UNRESOLVED_IN_DOUBT의 예약은 운영자 해소로만
- 일일 손실 예약은 거래일 경계(시장별, P1 시간 규율)에서 소멸

예약 산술은 decimal 문자열 연산(P1 journal 관례)이며 float 누적을 쓰지 않는다.

### D6. RECONCILE은 행동 제한 상태다

권위 값 불일치는 산식으로 보정하지 않고 RECONCILE 상태로 전이한다(StockOS `evaluate_synthetic_oco`의 원칙 채택). 이 change에서의 진입 조건: 브로커 보유·매도가능 조회 불가 또는 stale 초과, 로컬 파생과 브로커 스냅샷의 수량 불일치, 같은 식별자가 상충하는 계좌·심볼에 출현.

RECONCILE 중 허용/금지:
- 금지: 신규 진입, 수량 확대, (2c 예약) 보호 임의 취소
- 허용: 읽기·계좌 동기화·운영자 확인, **확정 하한 수량의 위험 축소**(과소 보호·과소 청산은 안전한 방향 — 수량을 정확히 몰라도 확정 하한으로는 행동 가능, §0.3)
- 해제: 재조회 일치 + 원인 기록

### D7. 브로커 식별자는 opaque, 상태는 문서화된 10개

식별자 취급(KIS 형태 검증의 교훈을 원칙으로만 채택): null·빈 문자열 거부, 응답 원문 그대로 저장(trim·변환 금지), 계좌 스코프와 함께 저장, 생성 응답의 id를 상세조회 round-trip으로 확인, 예상 밖 형식도 파싱하지 않음, 같은 id가 상충 컨텍스트에 나타나면 RECONCILE. **정규식·prefix 검증 금지** — `orderId`는 opaque token이다(openapi).

브로커 상태 파생을 문서화된 enum 전체로 확장한다: PENDING / PENDING_CANCEL / PENDING_REPLACE / PARTIAL_FILLED / FILLED / CANCELED / REJECTED / CANCEL_REJECTED / REPLACE_REJECTED / REPLACED (openapi OrderStatus). 미지 값은 P1대로 UNKNOWN_BROKER_STATE fail-closed. `CANCEL_REJECTED`/`REPLACE_REJECTED`는 "별도 주문 레코드로 생성됨"(openapi) — 취소·정정 해소 절차가 이 레코드를 인지해야 하며, 그 레코드의 구체 형태(원주문 링크 여부)는 `[미측정 — 2b]`이므로 인지 실패는 외부 분류가 아니라 RECONCILE로 처리한다.

체결 정정 이벤트: 누적 수량 동일 + `averageFilledPrice`/`filledAmount` 변경은 수량 재반영 없이 EXECUTION_CORRECTION 이벤트로 기록한다(P1 watermark의 "identical snapshots are a no-op"을 세분화 — 평균가는 부분 체결마다 바뀌는 값이므로(openapi) dedup 키에 넣지 않는다).

### D8. 엔진 Gateway 배선과 봉인

엔진 프로필이 ExecutionGateway를 구성한다(현재 `execgw.New`는 flatten CLI에만 존재, 엔진 Context에 Gateway 필드 없음). `Context.TradingService`(exported, 확인 토큰만으로 mutation 가능)를 봉인한다 — 엔진 컨텍스트는 mutation 메서드를 가진 값을 노출하지 않으며 정적 테스트로 증명. `runInterlock`과의 구성 순서, EntryGate·Resolver·NonceStore·예약 저장소 연결을 명시한다.

인터록 강화: (1) 필수 한도 항목별 configured·양수·유한·통화 일치 — 하나라도 누락이면 거부, (2) 거래 정책이 매도·실주문 실행을 허용하지 않으면 거부(naked long 방지; 조건주문 정책 검증은 2c), (3) Guardian 한도를 감사된 설정 한도 단일 출처에서 구성 — 동등성 검증은 EXPOSURE_RAISING 결정에만, (4) Gateway 미구성 시 거부. flatten의 결정 발급(`decisionFor`)은 RISK_REDUCING class를 싣도록 갱신한다 — class 도입이 비상 청산을 깨지 않아야 한다(2라운드 2d-C2의 절반; 조건주문 취소는 2c).

### D9. journal v5 — 단일 원자 마이그레이션, 스키마는 여기에

태스크는 이 표의 전사다. 전 컬럼 decimal은 TEXT.

**`decisions`** — 결정 영속(D1)
| 컬럼 | 타입/제약 |
|---|---|
| id | TEXT PK |
| intent_id | TEXT NOT NULL REFERENCES intents(id) |
| generation | INTEGER NOT NULL DEFAULT 0 |
| safety_class | TEXT NOT NULL CHECK(EXPOSURE_RAISING\|RISK_REDUCING\|PROTECTION_WEAKENING) |
| risk_preimage | TEXT NOT NULL (canonical JSON 원문) |
| risk_hash | TEXT NOT NULL |
| client_order_id | TEXT (place만, ≤36자) |
| limits_json | TEXT (EXPOSURE_RAISING 필수, RISK_REDUCING NULL) |
| nonce | TEXT NOT NULL UNIQUE |
| issued_at / expires_at | TEXT NOT NULL |
| UNIQUE(intent_id, generation) | |

**`risk_reservations`** — 진입 측 예약(D5)
| 컬럼 | 타입/제약 |
|---|---|
| id | INTEGER PK AUTOINCREMENT |
| decision_id | TEXT NOT NULL REFERENCES decisions(id) |
| account_ref | TEXT NOT NULL |
| kind | TEXT NOT NULL CHECK(OPEN_EXPOSURE\|DAILY_LOSS\|CASH) |
| amount / currency | TEXT NOT NULL |
| trading_day | TEXT (DAILY_LOSS만, 시장별 거래일) |
| snapshot_as_of | TEXT NOT NULL (검증에 쓴 브로커 스냅샷 시각) |
| state | TEXT NOT NULL CHECK(HELD\|RELEASED) DEFAULT HELD |
| released_at / release_reason | TEXT (해제 시 필수, reason enum: BROKER_TERMINAL\|EXPIRED_UNCONSUMED\|OPERATOR\|DAY_BOUNDARY) |

**`spent_nonces`** — durable NonceStore
| nonce TEXT PK · decision_id TEXT · consumed_at TEXT NOT NULL | 보존 기간 정책 포함 |

**`mutation_attempts` 확장(additive 컬럼)** — client_order_id TEXT, wire_body TEXT, serializer_version TEXT, replay_count INTEGER DEFAULT 0, last_replay_at TEXT. UNIQUE 인덱스(account_ref, client_order_id) WHERE client_order_id IS NOT NULL.

**`execution_corrections`** — order_id·prev/new averageFilledPrice·filledAmount·observed_at (D7).

마이그레이션 직전 자동 백업 + 복원 절차·테스트. 구버전 바이너리는 ErrSchemaTooNew — 롤백 수단이 아니며 복구는 백업 복원.

## Risks / Trade-offs

- [멱등 재생이 문서와 다르게 동작] → 2b attestation 전 비활성, 폴백은 P1 절차 그대로. 재생 없이도 전체가 성립
- [upstream 4파일 수정(orderintent·orders_write·trading·engine)] → Pre-Edit 전문 선언, CLI 무영향 characterization 테스트
- [CANCEL_REJECTED 별도 레코드의 형태 미측정] → 인지 실패는 RECONCILE(fail-closed), 2b 측정 항목
- [예약 재수집 루프와 폴링 경합] → 상한·데드라인·fail-closed, §0.4 rate budget 내

## Migration Plan

journal v5 단일 원자 마이그레이션(D9 표). 직전 백업 → 실패 시 복원. 스키마 계약 테스트.

## Open Questions

- margin 기본 60초의 타당성 — 2b 왕복 p99 실측으로 조정
- `CANCEL_REJECTED` 별도 레코드가 목록 조회에 어떻게 나타나는가 — 2b
