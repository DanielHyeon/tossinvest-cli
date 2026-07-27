# Change: extend-execution-contract

> 2026-07-26 4판(접합부 사양화 — design.md D1~D9가 정본). 판 이력·리뷰 4라운드(발견 108+36건, 폐쇄 감사 포함)는 `review.md`. 경계 원칙: **측정 의존성** — 이 change(2a)는 실계좌 측정 없이 확정 가능한 강제 장치만 담고, 조건주문·보호주문 관련은 전부 2b(측정) 이후의 2c로 이동했다.

## Why

리뷰가 확정한 사실 위에서 P1 실행 계약을 정정·강화한다:

- **P1 스펙의 사실 오류**: "자동 재제출 절대 금지 — 브로커 멱등성 키가 없으므로 무조건". 실제로는 `clientOrderId` 멱등키가 문서화되어 있고(동일 키·본문 재요청은 이전 결과 재반환, 10분 유효 — openapi), 엔진 경로는 이 필드를 쓰지 않는다. 응답 유실을 위해 설계된 장치를 비워두고 더 약한 대체물(조회 휴리스틱)만 만든 상태다. 단, 어떤 조회 응답도 키를 싣지 않으므로 회수는 재요청 응답으로만 가능하다.
- **fail-open 실재**: 게이트가 거래 정책을 보지 않아 매수만 되고 청산 불가한 naked long으로 기동 가능(`verifyGate`), 한도는 수량·notional 둘 다 0일 때만 거부(`Limits.IsZero()`), nonce가 in-memory, 결정이 위험 입력(손절·노출)에 결속되지 않음, `Context.TradingService`가 exported라 확인 토큰만으로 mutation 가능.
- **총계 한도가 제출 시점에 강제 불가**: 서로 다른 심볼의 동시 결정이 같은 스냅샷을 각각 통과해 합산 한도를 초과한다(in-flight 락은 심볼 단위).
- **P1의 브로커 모델이 계약과 다르다**: 실제 payload의 status는 10값 enum인데 파생기는 OPEN/CLOSED 2값을 전제한다 — 확장이 아니라 **재작성**이 필요하다. `CANCEL_REJECTED`/`REPLACE_REJECTED`는 별도 주문 레코드를 만들고(형태 `[미측정]`), `orderId`는 opaque token(형태 검증 금지), 평균 체결가는 부분 체결마다 바뀌며(정정 이벤트 필요), 목록 조회의 status는 그룹 라벨이라 `PARTIAL_FILLED`가 OPEN·CLOSED 양쪽에 속한다(해소 대조에 orderId dedup 필수), 반대 방향 동시 제출은 `422 opposite-pending-order-exists`로 거부된다.

채택 원칙(StockOS 실행 무결성 분석 → 토스 계약으로 재구현): 불확실성은 상태로 격리, 불일치는 RECONCILE로 중단(산식 보정 금지), 결정은 영속 후 실행, 브로커가 보증하지 않는 것은 타입으로 표시.

## What Changes

- **결정 계약**: safety class(EXPOSURE_RAISING/RISK_REDUCING + 형태 일치 검증으로 위조 차단, PROTECTION_WEAKENING enum 예약) + 클래스별 preimage(RiskIntent/ReductionIntent — flatten 포함) journal 영속·Gateway의 journal 기반 재검증(만료 시각 재검증 유지) + generation(2a에서 0 고정). 한도 면제를 `KindCancel` 리터럴에서 검증된 class 기준으로 재작성
- **멱등키 기계**: `f(decision_id, generation)` 유도(발급자 소유 값), RECORDED에 canonical wire body·serializer version 불변 저장, attempt→결정 결속(`PrepareRequest` 공개 계약 확장), 자기 방어 재생 진입점(state·attestation·회당 TTL−margin·상한·간격), 409/422 전용 응답 규칙(dispatch 분류기 금지), 조회 폴백의 orderId dedup·관측 창 오염 규칙. confirm token(canonical)은 무변경
- **진입 측 위험 예약**: 판정·예약 원자 트랜잭션(네트워크는 밖), 해제는 파생된 종결 상태·nonce 미소비 만료·운영자(UNKNOWN·UNRESOLVED의 유일 출구)·거래일 소멸만 — 만료 추정 해제 금지 `[미측정]`, nonce 소비는 전송 시작 트랜잭션에 병합, decimal 산술
- **RECONCILE 영속 상태**: `reconcile_states`가 권위(기존 in-memory 래치는 투영), 확정 하한 공식 `max(0, min(신선 보유, 신선 매도가능) − 미체결 SELL)`, 부재 시 자동 경로 매도 0
- **총계 한도 계산 계약**(수치는 2d가 보수 방향으로만 대체): 자동 진입 LIMIT 전용, gross long, 실현 손실 기준, 보수 staleness 기본값(계좌 스냅샷 10초·환율 60초)
- **브로커 취급 정정**: opaque 식별자 규칙(공백 검사 후 원문 저장 — 위반 3개소 수정, round-trip 실패→IN_DOUBT), 상태 파생 **재작성**(10값 OrderStatus), EXECUTION_CORRECTION(`fill_snapshots.filled_amount` + payload 확장 + 동일 트랜잭션)
- **엔진 배선·봉인**(순서 고정): 계좌 해석 무조건화 → journal 편입(파일시스템 allowlist가 기동 조건이 됨) → Gateway 구성 → `TradingService` 봉인 + flatten 자체 배선 전환 → 인터록 강화(항목별 한도 fail-closed·거래 정책·단일 출처·Gateway 확인·키 미지원 transport 거부)
- **journal v5 단일 원자 마이그레이션**: design D9 표의 전사 + durable NonceStore

## Capabilities

### Modified Capabilities

- `order-execution`: IN_DOUBT 해소 재정의(멱등키 사실 정정·재생 1차·orderId dedup 조회 폴백), MutationAttempt 수명주기(결정 결속·wire body 영속), 브로커 상태 파생(10값 재작성), **Retry Matrix 산출물**(재생=재시도가 아니라 정체 회수 — 단일 서술)
- `engine-safety`: ExecutionGateway 봉인(엔진 배선 순서·TradingService·flatten 자체 배선), 자동화 게이트 기동 인터록(fail-closed 한도·거래 정책·단일 출처)

### New Capabilities

(없음)

## Impact

- Affected code: `internal/execgw`(결정 영속·class 형태 검증·멱등키·재생·예약·RECONCILE), `internal/journal`(v5·NonceStore·정정 이벤트·PrepareRequest 확장), `internal/brokerstate`(파생 재작성), `internal/app/engine`(배선·봉인·인터록), `internal/flatten`(ReductionIntent 결정), `internal/filldetect`(payload filledAmount)
- **upstream 수정(Pre-Edit 필수)**: `internal/orderintent/intent.go`, `internal/official/orders_write.go`, `internal/trading/`(엔진 진입점), `internal/app/engine/engine.go`, `internal/journal/durability.go`(PrepareRequest), `cmd/tossctl/flatten.go`(자체 배선 전환)
- 후행: 2b `verify-execution-capability`(멱등키 실동작·TTL 마진·만료 status·CANCEL_REJECTED 형태·조건주문 속성·매도가능 의미·비용표) → 2c `add-protection-orders`(2b 결과 위에서 작성) → 2d `add-core-domain`(동결, 재작성 대기)
- 이 change 완료 후에도 자동 진입은 불가하다(Guardian 발급자는 2d). 게이트 기본 OFF 유지
