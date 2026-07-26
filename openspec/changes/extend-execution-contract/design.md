# Design: extend-execution-contract

> 2026-07-26 4판(접합부 사양화). 판 이력·발견 108+36건은 `review.md`. 경계 원칙: **측정 없이 확정 가능한 것만** — 조건주문·보호주문·발동 주문·청산 예약은 어떤 형태로도 없다(2c). 브로커 동작 서술은 `docs/migration/openapi.latest.json` 인용 또는 `[미측정]` 태그 필수.

## Context

P1은 place/cancel/amend에 journal 선기록 → DISPATCH_STARTED → 분류 → IN_DOUBT 해소, GuardianDecision 재검증, 심볼당 in-flight 1개, raw mutator 봉인을 만들었다. 리뷰가 확정한 사실:

- **멱등키 실재**: `OrderCreateRequest.clientOrderId` — "멱등성 키로 사용됩니다 … 동일 값으로 재요청 시 이전 주문 결과를 그대로 재반환 … 10분간 유효"(openapi). P1의 "멱등키 없음" 전제는 오류이고 엔진 경로는 필드를 쓰지 않는다. **어떤 조회 응답도 키를 싣지 않는다** — 회수는 재요청 응답으로만.
- **`orderId`는 opaque token**(형태 계약 없음, 64자 난수 example).
- **OrderStatus 10개 문서화**. `CANCEL_REJECTED`/`REPLACE_REJECTED`는 "별도 주문 레코드로 생성됨". **status 파라미터는 그룹 라벨**이며 `PARTIAL_FILLED`는 OPEN·CLOSED **양쪽 그룹에 속한다**(openapi) — 양쪽 목록 대조는 orderId dedup 없이는 부분 체결 주문을 이중 매칭한다.
- **`422 opposite-pending-order-exists`**: "동일 종목에 반대 방향의 체결 대기 주문이 있습니다"(openapi) — 반대 방향 동시 제출은 브로커가 거부한다.
- **`409 request-in-progress`**: "동일 주문 키에 대해 처리 중인 요청이 있습니다"(openapi) — 재생의 최빈 응답.
- **체결은 누적 모델**(`OrderExecution` — 개별 체결 ID 없음, 평균가는 부분 체결마다 변동). P1 journal이 watermark(delta>0 전진/동일 no-op/감소 fail-closed)를 이미 구현.
- 미체결 만료가 어떤 status로 나타나는지는 **문서에 없다** `[미측정 — 2b 2.1]`. P1 파생표는 "CLOSED+fill0+무취소 → UNKNOWN(주석: Expiry would look like this)"로 이미 보수 처리한다.

원칙(StockOS 분석 채택, 토스 계약으로 재구현): 불확실성은 상태로 격리, 불일치는 RECONCILE로 중단, 결정은 영속 후 실행, 브로커가 보증하지 않는 것은 타입으로 표시.

## Goals / Non-Goals

**Goals**: 결정 계약(class·preimage·generation·멱등키 결속), 멱등 재생 골격(자기 방어 진입점), 진입 측 위험 예약, RECONCILE 영속 상태, 총계 계산 계약(2a 구조 결정 포함), opaque 식별자·상태 파생 재작성·EXECUTION_CORRECTION, 엔진 배선·봉인. 전부 httptest + 합성 결정으로 테스트 가능.

**Non-Goals**: 조건주문 일체(2c), 발동 주문 귀속·격리 원장(2c), 청산 수량 예약·PROTECTION_WEAKENING 발급(2c), 재생 실사용 활성화(2b 후), Guardian 발급자·한도 수치(2d), 전략(P3), filldetect 상시 루프 배선(엔진 루프를 소유하는 change).

## Decisions

### D1. 결정은 발급자가 영속하고, attempt가 결정을 가리킨다

발급자(2a에서는 flatten과 테스트; 2d에서 Guardian)가 `decisions` 행을 자기 트랜잭션으로 **Gateway 호출 전에** 기록한다. `decisions`는 intent를 참조하지 않는다 — intent id는 Gateway 내부에서 발급되므로(gateway.go:583) 발급 시점에 존재하지 않는다. 대신 **attempt가 `decision_id`·`safety_class`·`generation`을 가진다**: `PrepareRequest`를 확장해(공개 API 변경 — 태스크 명시) Prepare 트랜잭션이 attempt→결정 결속을 기록한다. Gateway `verifyDecision`은 journal에서 decision 행을 읽어 preimage 해시를 재계산·대조한다(호출자 공급 값 금지).

preimage는 **클래스별 스키마**다: EXPOSURE_RAISING → RiskIntent(계좌·시장·심볼·방향·진입가·손절가·목표가·수량·정책 버전), RISK_REDUCING → ReductionIntent(계좌·시장·심볼·방향·상한 수량·사유). flatten은 ReductionIntent를 기록하므로 퇴화 carve-out 없이 `risk_preimage NOT NULL`이 성립한다.

`generation`은 같은 논리 의도의 재발급 순서. **2a에서는 항상 0** — 전진시키는 주체(보호 교체·재발급)는 2c/2d가 정의한다.

### D2. 멱등키 = `f(decision_id, generation)`

발급자가 소유한 값에서만 유도한다(intent id는 발급 시점에 미존재). ≤36자, `[a-zA-Z0-9\-_]`(openapi 패턴). 결정 행에 저장되고, Prepare 시 attempt에 복사되며, Gateway가 둘의 일치를 재검증한다. RECORDED 단계에 canonical wire body + serializer version이 불변 영속된다 — 재생은 저장본만 사용(구조화 필드 재구성 금지: 직렬화 규칙 변화가 key-conflict를 만든다).

`CanonicalPlace`(확인 토큰 입력)에는 넣지 않는다 — CLI confirm token 무변경(§0.2). 일반 주문 경로 배선: `orderintent.PlaceIntent`·`official.orderCreateV0/V1`·응답 파서에 필드 추가(Pre-Edit). Gateway는 키를 실을 수 없는 transport로의 place를 거부한다(키가 전송됐다는 전제 위에 재생이 서 있다 — 엔진은 official 직결이므로 정적으로 성립, 테스트로 고정).

### D3. 멱등 재생 — 자기 방어 진입점

해소 1차 절차. **진입점 자신의 의무**(호출자 전제 금지): attempt state == IN_DOUBT, attestation 플래그 ON `[2b 전 비활성]`, **재생 1회마다** `elapsed(dispatch_started_at) < TTL(10분 문서값) − margin(기본 60초, 설정 주입)` 재검사, 회수 상한(기본 2회)·최소 간격, 저장된 wire body 외 전송 불가(새 본문 구성 API 자체가 없음 — 정적으로 증명 가능한 형태: 진입점 입력은 attempt id뿐).

**응답 규칙 — 재생 경로에 `ClassifyHTTPMutation` 사용 금지**(dispatch 분류기는 422를 확정 거부로 종결하는데 재생의 422는 의미가 다르다):
- 2xx + orderId → 정체 회수, CONFIRMED
- `409 request-in-progress`(openapi) → 원 요청 처리 중 — 대기 후 재시도, 상한 미소비
- `422 idempotency-key-conflict`(openapi) → 같은 키에 다른 본문이 전송된 적 있음 — **FAILED_CONFIRMED가 아니다**(원 주문에 대해 아무것도 말하지 않는다). 프로그램 오류로 간주, UNRESOLVED + critical 알림
- 응답 유실 → replay_count·last_replay_at 기록, 상한 후 조회 폴백

재생은 새 attempt를 만들지 않고 nonce를 소비하지 않는다(이미 소비된 nonce의 attempt에 대한 정체 회수).

**복구 호출 그래프**: 재시작 복구가 소유한다 — PendingAttempts 순회 → IN_DOUBT이고 재생 적격이면 Gateway 재생 진입점 → 부적격·실패 시 Resolver 조회 폴백. Resolver 자신은 P1 불변식대로 writer를 갖지 않는다(진입점은 Gateway 소속).

**조회 폴백의 정정**: `status` 파라미터는 그룹 라벨이고 `PARTIAL_FILLED`는 양쪽 그룹에 속하므로(openapi), 양쪽 목록 대조는 **orderId 기준 dedup 후** 유일성을 판정한다 — dedup 없이는 부분 체결 + 응답 유실(최빈 사례)이 이중 매칭으로 영구 주차된다. 또한 관측 창 동안 같은 심볼에 다른 mutation이 전송됐다면 잔고·매수가능금액 delta 교차확인은 무효이며 자동 FAILED_CONFIRMED는 금지된다(부재를 증명할 수 없다 — UNRESOLVED로).

### D4. Safety class — 형태 검증으로 위조 차단, 동시 carve-out 폐기

클래스: EXPOSURE_RAISING(진입) / RISK_REDUCING(reduce-only 청산·취소). PROTECTION_WEAKENING은 enum·CHECK에 예약만(2c 발급 — 예약 값은 미도달 상태로 남으며 이는 의도된 forward-compat이다).

**class는 caller 선언만으로 효력이 없다**: Gateway가 mutation 형태에서 노출 증가 여부를 독립 계산해 `EXPOSURE_RAISING ⇔ raisesExposure` 일치를 검증한다. 불일치 거부 — 발급자가 없는 2a에서 class 선언으로 한도를 우회하는 경로를 구조적으로 차단한다(BUY에 RISK_REDUCING을 붙여도 거부). 한도 면제는 `KindCancel` 리터럴 → class 기준으로 재작성하되, 이 형태 검증이 선행한다.

**3판의 동시 carve-out("미해소 EXPOSURE_RAISING이 RISK_REDUCING을 차단하지 않는다")은 폐기한다.** 근거: (1) 브로커가 반대 방향 동시 주문을 `422 opposite-pending-order-exists`로 거부하고, P1 분류기는 422를 확정 거부로 종결하므로 조항의 §0.3 의도가 브로커 경계에서 반전된다. (2) 동시 청산이 잔고 delta를 바꿔 부재 증명을 오염시킨다. (3) attempt에 class가 없어 데이터 경로도 없었다. **2a의 §0.3**: 심볼 latch는 전 클래스 유지(유일 매칭·422 회피), 대신 IN_DOUBT 해소가 우선·유계이고(재생이 이를 빠르게 한다), 수동 flatten의 순서화된 saga(취소 확정 → 매도)는 P1 그대로 무영향. 보호주문의 동시성 요구는 2c가 브로커 계약 위에서 재설계한다.

### D5. 진입 측 위험 예약 — 해제는 파생된 종결 상태만

판정·예약은 하나의 journal 트랜잭션(네트워크는 밖: 스냅샷 사전 수집 → 안에서 as-of·staleness 검증 → 삽입, 불충족 롤백·재수집 상한 3회·데드라인·초과 fail-closed). decimal 문자열 산술.

해제 트리거:
- **attempt가 파생된 브로커 종결 상태에 도달**(FILLED/CANCELLED/REJECTED 계열 + NOT_DISPATCHED/FAILED_CONFIRMED). 판정 주체는 brokerstate 파생 함수이며, 만료 추정으로 해제하지 않는다 — 만료의 status 표현은 `[미측정 — 2b 2.1]`이고 P1 파생은 CLOSED+fill0+무취소를 UNKNOWN으로 보존한다
- nonce **미소비** 결정의 만료 (소비 후 만료는 미해제)
- **운영자 해제**: UNKNOWN_BROKER_STATE·UNRESOLVED가 잡은 예약의 유일한 출구 — 근거 기록 필수, audit. fail-closed 래칫은 옳지만 운영자가 볼 수 없는 래칫은 아니다
- 일일 손실 예약의 거래일 경계 소멸(시장별, P1 clock)

예약↔관측 조인: `risk_reservations.attempt_id`(Prepare 시 backfill)로 attempt 종결과 결합한다. 해제는 종결 기록과 같은 journal 트랜잭션에서 수행한다 — 생산자가 filldetect든 해소 절차든 운영자든, **기록 시점에 해제가 원자적으로 따라온다**. filldetect의 상시 루프 배선은 엔진 루프를 소유하는 change의 몫이며, 2a는 트랜잭션 결합만 정의한다.

nonce 소비는 `MarkDispatchStarted` 트랜잭션에 병합한다 — 단일 커넥션 핫패스에 독립 쓰기를 추가하지 않는다.

### D6. RECONCILE — journal이 권위, 확정 하한은 공식이다

`reconcile_states` 테이블로 영속한다(scope·원인·진입 시각·증거·해제 시각·해제 원인). **기존 3개 in-memory 차단 기계(EntryGate 래치, checkSymbolFree의 UNRESOLVED 차단, reconcile.Tracker)와의 관계**: journal의 RECONCILE 상태가 권위이고, EntryGate의 reconcile 계열 래치는 기동 시 journal에서 재구성되는 투영이다. Tracker의 메모리 상태는 이 테이블로 이전한다.

진입 조건: 브로커 보유·매도가능 조회 불가 또는 stale 초과(소비자가 그 값을 필요로 하는 시점 기준), 로컬 파생과 브로커 스냅샷의 수량 불일치, 같은 식별자의 상충 컨텍스트 출현.

**확정 하한의 공식**:

```
확정 하한 = max(0, min(신선한 브로커 보유수량, 신선한 매도가능수량) − 로컬 미체결 SELL 수량)
신선 = 스냅샷 나이 ≤ staleness 한계
스냅샷 부재 → 0 (자동 경로의 로컬 매도 없음)
```

로컬 파생 수량은 하한을 **올리는** 근거가 될 수 없다(외부 주문 제외로 실제보다 클 수 있다). 매도가능수량의 정확한 의미는 `[미측정 — 2b 2.8]`이나 min()에서 하한을 낮추는 방향으로만 쓰므로 보수적이다. 수동 flatten은 자체 신선 조회를 수행하므로 무변경(P1 동작 보존 — §0.3).

### D7. 브로커 취급 — opaque 식별자, 상태 파생 재작성, 정정 이벤트

**opaque 규칙(재서술)**: 공백 제거 후 비면 거부하고, 저장은 **수신 원문 그대로**(변환 없음), 비교는 바이트 동일. trim은 검사에만 쓰고 저장에 쓰지 않는다 — 현재 위반 3개소를 명시 수정한다(`execgw/classify.go:149`, `journal/resolution.go:42·47·126`, `execgw/indoubt.go:512·516`). 정규식·prefix 검증을 추가하지 않는다(리뷰 확인 항목 — `orderId`는 opaque, openapi). 생성 응답 식별자의 상세조회 round-trip: 실패·부재 시 CONFIRMED가 아니라 **IN_DOUBT**로 남겨 해소 절차가 확정한다(MarkAcked 후·Settle 전).

**상태 파생은 확장이 아니라 재작성이다**: 현재 파생기는 raw status 2값(OPEN/CLOSED) 전제로 18행 표를 만들었고 실제 payload의 10값 enum은 전부 UNKNOWN으로 떨어진다(derive.go:105-106, 279-281). 문서화된 10값(PENDING/PENDING_CANCEL/PENDING_REPLACE/PARTIAL_FILLED/FILLED/CANCELED/REJECTED/CANCEL_REJECTED/REPLACE_REJECTED/REPLACED — openapi)에 대한 새 우선순위 표를 만들고, `canceledAt`·수량·lineage와의 모순 조합·미지 값은 UNKNOWN fail-closed를 유지한다. **D5의 예약 해제·terminal 판정·추적 집합이 전부 이 재작성의 하류다**(선행 의존). CANCEL_REJECTED/REPLACE_REJECTED "별도 주문 레코드"(openapi)는 인지 실패 시 외부 분류가 아니라 RECONCILE `[형태 미측정 — 2b 2.1]`.

**EXECUTION_CORRECTION**: `fill_snapshots`에 `filled_amount` 컬럼을 추가하고(prev 없이는 금액-only 정정을 감지할 수 없다) filldetect payload가 `filledAmount`를 읽도록 확장한다. 동일 누적 수량 + 평균가/금액 변경은 **RecordFill의 같은 `BEGIN IMMEDIATE` 안에서**(prev가 존재하는 유일한 지점) 정정 이벤트를 삽입하고 스냅샷을 갱신한다 — 수량 delta 없음. dedup: 스냅샷과 동일하면 no-op이므로 반복 poll은 자연 멱등.

### D8. 엔진 배선 — 분할과 순서

8.x는 하나의 태스크가 아니다. 순서와 결과를 명시한다:

1. **계좌 해석의 무조건화**: 엔진 프로필은 게이트 OFF에서도 기동 시 계좌를 해석한다(`GET /api/v1/accounts` — 읽기 전용 엔진도 필요로 하는 값). 현재는 게이트 ON 검증 내부에만 존재(interlock.go:217-222)
2. **journal 편입**: 엔진 프로필이 journal을 연다 — **결과 명시**: 파일시스템 allowlist(ext4/xfs/btrfs)와 무결성 검사가 엔진 기동 조건이 된다(P1 journal 계약의 의도된 상속). config-dir 격리(flatten의 기존 동작)는 보존
3. **Gateway 구성**: journal·EntryGate(journal RECONCILE 투영 재구성 포함)·Resolver·NonceStore·예약 저장소. `runInterlock`은 이 구성 **후** 실행되어 "Gateway 구성됨"을 검증한다
4. **봉인**: `Context.TradingService` 노출 제거. 유일한 생산 소비자인 flatten(flatten.go:223)은 엔진 Context가 아니라 **자체 배선**으로 trading.Service를 구성한다(이미 자체 EntryGate·journal·Source를 가진다) — flatten의 P1 동작 무변경을 characterization으로 고정
5. 인터록 강화: 한도 항목별 fail-closed·거래 정책(매도+실행)·단일 출처(EXPOSURE_RAISING 한정)·Gateway 구성 확인

**Retry Matrix와의 관계**: 아카이브 스펙 "주문 mutation은 어떤 오류에도 자동 재시도 금지"와 재생의 공존은 정의가 아니라 스펙으로 해소한다 — Retry Matrix 요구를 MODIFIED에 포함해 "멱등 재생은 재시도가 아니라 정체 회수(같은 키는 유효 창 안에서 새 주문을 만들 수 없다)"를 한 곳에 명시하고, `execgw/retry.go`의 동일 서술을 갱신한다.

### D9. journal v5 — 단일 원자 마이그레이션 (태스크는 이 표의 전사)

전 decimal은 TEXT. 마이그레이션 직전 자동 백업, 복구는 백업 복원(구버전 실행은 ErrSchemaTooNew — 롤백 아님).

**`decisions`**

| 컬럼 | 제약 |
|---|---|
| id | TEXT PK (발급자 발행) |
| account_ref | TEXT NOT NULL |
| generation | INTEGER NOT NULL DEFAULT 0 (2a에서 항상 0) |
| safety_class | TEXT NOT NULL CHECK(EXPOSURE_RAISING\|RISK_REDUCING\|PROTECTION_WEAKENING) |
| preimage_kind | TEXT NOT NULL CHECK(RISK_INTENT\|REDUCTION_INTENT) |
| risk_preimage | TEXT NOT NULL (canonical JSON 원문) |
| risk_hash | TEXT NOT NULL |
| client_order_id | TEXT (place만; `f(id, generation)`) |
| limits_json | TEXT (EXPOSURE_RAISING 필수, RISK_REDUCING NULL) |
| nonce | TEXT NOT NULL UNIQUE |
| issued_at / expires_at | TEXT NOT NULL |

**`mutation_attempts` additive 컬럼**: decision_id TEXT REFERENCES decisions(id), safety_class TEXT, generation INTEGER, client_order_id TEXT, wire_body TEXT, serializer_version TEXT, replay_count INTEGER DEFAULT 0, last_replay_at TEXT. UNIQUE(account_ref, client_order_id) WHERE client_order_id IS NOT NULL. `PrepareRequest` 확장으로 기록(공개 API 변경 — durability_test 고정 대상).

**`risk_reservations`**: id PK, decision_id NOT NULL REFERENCES decisions(id), attempt_id TEXT REFERENCES mutation_attempts(id) (Prepare 시 backfill), account_ref NOT NULL, kind CHECK(OPEN_EXPOSURE|DAILY_LOSS|CASH), amount/currency NOT NULL, trading_day (DAILY_LOSS만), snapshot_as_of NOT NULL, state CHECK(HELD|RELEASED) DEFAULT HELD, released_at/release_reason (enum: BROKER_TERMINAL|EXPIRED_UNCONSUMED|OPERATOR|DAY_BOUNDARY).

**`spent_nonces`**: nonce PK, decision_id, consumed_at NOT NULL. **보존 불변식: 보존 기간 ≥ 최대 결정 TTL** — 어떤 결정도 자기 소비 기록보다 오래 살 수 없다.

**`reconcile_states`**: id PK, account_ref NOT NULL, symbol (NULL=계좌 전역), cause enum NOT NULL, evidence TEXT, entered_at NOT NULL, released_at, release_cause. 활성 상태 부분 unique(account_ref, symbol) WHERE released_at IS NULL.

**`execution_corrections`**: id PK, account_ref NOT NULL, order_id NOT NULL, prev_avg_price/new_avg_price, prev_filled_amount/new_filled_amount, cumulative_qty NOT NULL, observed_at NOT NULL. RecordFill 트랜잭션 내 삽입.

**`fill_snapshots` additive 컬럼**: filled_amount TEXT (정정 감지의 prev).

## Risks / Trade-offs

- [재생이 문서와 다르게 동작] → 2b 전 비활성, 폴백 P1 절차. 재생 없이도 전체 성립
- [upstream 수정 4+2파일] → Pre-Edit 전문, CLI confirm token 무변경 characterization
- [brokerstate 재작성이 P1 파생 소비자에 파급] → 기존 테스트 중 바뀌어야 할 단언 사전 열거(무조건 green 금지)
- [journal-in-engine의 파일시스템 조건] → 의도된 상속임을 명시, tmpfs 테스트는 격리 경로 사용
- [staleness 수치가 2a에 필요] → 7.1이 보수 기본값을 명시 정의(2d는 보수 방향으로만 대체)

## Migration Plan

v5 단일 원자(D9 표). 직전 백업 → 실패 시 복원. 스키마 계약 테스트로 전이·거부 고정.

## Open Questions

- margin 60초 타당성 — 2b 왕복 p99
- 만료 주문의 실제 status 표현 / CANCEL_REJECTED 레코드 형태 — 2b 2.1
