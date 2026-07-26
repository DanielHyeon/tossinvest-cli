# Change: extend-execution-contract

> 2026-07-26 3판. 1·2판과 리뷰 3라운드(108건)의 기록은 `review.md`. 3판의 재범위 원칙: **측정 의존성으로 경계를 긋는다** — 이 change(2a)는 실계좌 측정 없이 확정 가능한 강제 장치만 담고, 조건주문·보호주문 관련은 전부 2b(측정) 이후의 2c로 이동했다.

## Why

리뷰 3라운드가 확정한 사실 위에서 P1 실행 계약을 정정·강화한다:

- **P1 스펙의 사실 오류**: "자동 재제출 절대 금지 — 브로커 멱등성 키가 없으므로 무조건". 실제로는 `clientOrderId` 멱등키가 문서화되어 있고(동일 키·본문 재요청은 이전 결과 재반환, 10분 유효 — openapi), 엔진 경로는 이 필드를 쓰지 않는다. 응답 유실을 위해 설계된 장치를 비워두고 더 약한 대체물(조회 휴리스틱)만 만든 상태다.
- **fail-open 실재**: 게이트가 거래 정책을 보지 않아 매수만 되고 청산 불가한 naked long으로 기동 가능(`verifyGate`), 한도는 수량·notional 둘 다 0일 때만 거부(`Limits.IsZero()`), nonce가 in-memory, 결정이 위험 입력(손절·노출)에 결속되지 않음, `Context.TradingService`가 exported라 확인 토큰만으로 mutation 가능.
- **총계 한도가 제출 시점에 강제 불가**: 서로 다른 심볼의 동시 결정이 같은 스냅샷을 각각 통과해 합산 한도를 초과한다(in-flight 락은 심볼 단위).
- **P1의 브로커 모델이 계약보다 좁다**: OrderStatus는 10개가 문서화("OPEN/CLOSED 수준" 전제보다 풍부)되어 있고 `CANCEL_REJECTED`/`REPLACE_REJECTED`는 별도 주문 레코드를 만든다. `orderId`는 opaque token(형태 검증 금지). 평균 체결가는 부분 체결마다 바뀐다(정정 이벤트 필요).

채택 원칙(StockOS 실행 무결성 분석 → 토스 계약으로 재구현): 불확실성은 상태로 격리, 불일치는 RECONCILE로 중단(산식 보정 금지), 결정은 영속 후 실행, 브로커가 보증하지 않는 것은 타입으로 표시.

## What Changes

- **결정 계약**: safety class(EXPOSURE_RAISING/RISK_REDUCING, PROTECTION_WEAKENING은 enum 예약) + RiskIntent preimage journal 영속(Gateway는 journal에서 재검증) + generation. 한도 면제를 `KindCancel` 리터럴에서 class 기준으로 재작성
- **멱등키 기계**: 결정 결속 결정적 `clientOrderId`, RECORDED 단계에 canonical wire body·serializer version 불변 저장, TTL−margin 시간 규칙, 재생 골격(2b attestation 전 비활성), 재생 상한·폴백. confirm token(canonical)은 무변경
- **진입 측 위험 예약**: 판정·예약 원자 트랜잭션(네트워크는 밖), 해제는 브로커 종결 상태 정본(미체결 만료 포함), nonce 소비 후 만료는 미해제, 일일 손실 예약의 거래일 소멸, 재수집 상한·fail-closed, decimal 산술
- **RECONCILE 상태**: 불일치·조회 불가 시 행동 제한 상태로 전이(진입·수량 확대 금지, 확정 하한 위험 축소 허용)
- **총계 한도 계산 계약**(수치는 2d) + 한도 항목별 fail-closed
- **브로커 취급 정정**: opaque 식별자 규칙(round-trip 확인·상충 시 RECONCILE), OrderStatus 10개 파생 확장, EXECUTION_CORRECTION 이벤트
- **엔진 배선·봉인**: Gateway 구성, `TradingService` 봉인, 인터록 강화(한도 fail-closed·거래 정책·단일 출처·Gateway 확인), flatten 결정에 class 부여
- **journal v5 단일 원자 마이그레이션**: design.md D9 표의 전사 + durable NonceStore

## Capabilities

### Modified Capabilities

- `order-execution`: IN_DOUBT 해소 재정의(멱등키 사실 정정·재생 1차·조회 폴백), MutationAttempt 수명주기(wire body·키 영속), 브로커 상태 파생(문서화된 10개 enum)
- `engine-safety`: ExecutionGateway 봉인(엔진 배선·TradingService), 자동화 게이트 기동 인터록(fail-closed 한도·거래 정책·단일 출처)

### New Capabilities

(없음)

## Impact

- Affected code: `internal/execgw`(결정 영속·class·멱등키·재생·예약·RECONCILE), `internal/journal`(v5·NonceStore·정정 이벤트), `internal/app/engine`(Gateway 구성·봉인·인터록), `internal/brokerstate`(상태 enum 확장), `internal/flatten`(결정 class)
- **upstream 수정(Pre-Edit 필수)**: `internal/orderintent/intent.go`, `internal/official/orders_write.go`(clientOrderId 배선), `internal/trading`(엔진 진입점), `internal/app/engine/engine.go`(봉인)
- 후행: 2b `verify-execution-capability`(멱등키 실동작·TTL 마진·CANCEL_REJECTED 레코드 형태·조건주문 속성 측정) → 2c `add-protection-orders`(조건주문 형제 수명주기·발동 주문 다리·격리 원장·청산 수량 예약·PROTECTION_WEAKENING·flatten 조건주문 취소 — 2b 결과 위에서 작성) → 2d `add-core-domain`(판단 정책, 재작성 대기)
- 이 change 완료 후에도 자동 진입은 불가하다(Guardian 발급자는 2d). 게이트 기본 OFF 유지
