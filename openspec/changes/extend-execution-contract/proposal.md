# Change: extend-execution-contract

> 2026-07-26 설계 재작성. 초판(조건주문 = 새 MutationKind)은 proposal-freeze 리뷰에서 P1과 조립되지 않음이 확인되어 폐기했다 — `review.md`.

## Why

add-core-domain의 리뷰가 드러낸 구조적 오류: 무인 자동매매의 안전 근거 전체가 "네이티브 조건주문이 브로커에 상주한다"에 걸려 있는데, **조건주문 경로는 P1이 만든 안전 아키텍처 바깥에 통째로 있다.** 이 change의 리뷰는 그 위에 두 가지를 더 확정했다.

**조건주문은 주문이 아니다.** 식별자 네임스페이스·수명주기·응답 형태가 다르고, 그 실행은 **별개의 주문을 만든다**. 조건주문 식별자를 일반 주문의 브로커 주문번호 자리에 넣으면 그 값이 체결 감지 추적 집합(`journal.TrackedFillOrders` → `filldetect`)과 reconciliation의 로컬 미체결 목록으로 흘러가는데, 그 식별자는 일반 주문 조회에 유효하지 않다 — **보호주문 하나가 전 종목 체결 감지를 죽이고 진입을 영구 차단한다.**

**브로커에 멱등키가 있다.** P1 확정 스펙은 "자동 재제출 절대 금지 — 브로커 멱등성 키가 없으므로 무조건"이라고 쓰고 있지만, `docs/migration/openapi.latest.json`은 두 생성 엔드포인트 모두에 `clientOrderId`를 멱등키로 문서화한다(동일 키·동일 본문 재요청은 이전 결과를 그대로 반환, 10분 유효). 그런데 `internal/execgw`는 이 필드를 한 번도 쓰지 않는다 — 응답 유실이라는 정확히 그 상황을 위해 설계된 장치를 자동 경로가 비워둔 채, 훨씬 약한 대체물(pagination·안정화 조회·fingerprint 휴리스틱)을 만들어 놓았다.

그 밖에 코드로 확인된 fail-open: 게이트가 거래 정책을 보지 않아 **매수만 되고 손절·청산은 불가한 naked long**으로 기동할 수 있고(`verifyGate`), 한도는 수량·notional이 **둘 다** 0일 때만 기동을 거부하며(`Limits.IsZero()`), `Context.TradingService`가 exported라 확인 토큰만으로 조건주문을 낼 수 있고, 성공한 조건주문 취소·정정은 어댑터가 식별자를 반환하지 않아 전부 IN_DOUBT로 간다.

이 change는 강제 장치(레일)만 다룬다. 판단 정책(Guardian 판정 체인·한도 수치)은 후행 add-core-domain이 맡는다. 경계 원칙: **여기는 실패해도 안전한 레일, 거기는 그 레일 위의 판단.** P1이 오늘 그렇게 하듯 이 change는 합성 GuardianDecision만으로 완전히 테스트된다.

## What Changes

- **조건주문의 형제 수명주기**: 자체 journal 테이블과 자체 식별자 컬럼. 단계 모델(RECORDED→DISPATCH_STARTED→ACKED/IN_DOUBT→종결)과 원칙은 공유하되 저장소는 분리한다. 조건주문 intent는 유형·만료일·leg별 방향·트리거가를 자체 컬럼으로 보존해 해소 시점 fingerprint 재계산을 가능하게 한다
- **발동 주문 다리**: 등록 시 leg별 예상 주문 기록 → 조건주문 상태 폴러가 `triggeredOrderId` 관측 → 그 **주문 id**를 일반 추적·체결·reconcile 경로에 lineage와 함께 편입. 예상 주문은 일회성·유계이며 종결은 체결 귀속 완료 후
- **멱등 재생**: attempt별 결정적 멱등키를 전송 전에 영속하고, 해소의 1차 절차를 동일 본문 재요청으로 한다. 재시도가 아니라 **정체 회수**다 — 같은 키는 새 주문을 만들 수 없다. 어떤 조회 응답도 멱등키를 싣지 않으므로 조회는 키로 매칭할 수 없고, 이 비대칭이 1차 절차와 폴백이 서로를 대체하지 못하는 이유다. 실계좌 능력 검증 전에는 사용하지 않는다
- **3-클래스 safety**: EXPOSURE_RAISING / RISK_REDUCING / **PROTECTION_WEAKENING**. 취소를 일괄로 위험 감소로 분류하지 않는다 — 활성 보호의 취소는 위험 증가다. 클래스당 심볼 latch를 유지해 해소의 유일 매칭을 보장하되, RISK_REDUCING은 진입의 in-flight·IN_DOUBT에 막히지 않는다(§0.3)
- **청산 수량 예약**: `보유 − 미체결 SELL − 대기 조건주문 예약 − 동시 예약`의 원자적 총량. 단건 상한은 보호 100 + 청산 100이 각각 통과해 매도 200이 되는 것을 막지 못한다. 권위는 최근 브로커 스냅샷이며 로컬 파생을 상한 상향 근거로 쓰지 않는다
- **진입 측 위험 예약**: 결정 발급과 같은 트랜잭션. 브로커 조회는 트랜잭션 밖에서 수집하고 안에서 as-of·staleness를 검증한다(journal은 단일 커넥션이므로 네트워크를 트랜잭션에 넣으면 보호 경로까지 막힌다). nonce 소비 후 만료는 예약을 풀지 않고, 일일 손실 예약은 거래일 경계에서 소멸한다
- **총계 한도의 계산 계약**: 정의되지 않은 양에 예약을 걸 수 없으므로 계산 계약을 여기서 정의한다(수치는 add-core-domain)
- **결정 계약**: RiskIntent 해시 + **preimage를 journal에 영속**(제출자 공급 값으로 재검증하면 순환한다), 결정에 safety class 탑재, 한도 면제를 종류 리터럴이 아닌 class 기준으로 재작성, 한도 항목별 fail-closed, journal 기반 NonceStore
- **엔진 배선·인터록**: 엔진 프로필의 Gateway 구성(현재 존재하지 않는다), 조건주문 mutation 메서드 노출 봉인, `RequiredEndpoints()`에 조건주문 등록·취소·정정 + 목록 조회 + 매도가능수량 조회 추가, 거래 정책 검증, 한도 단일 출처

## Capabilities

### Modified Capabilities

- `order-execution`: IN_DOUBT 해소를 멱등 재생 우선으로 재정의(**"멱등키가 없다"는 사실 오류 정정**), MutationAttempt 수명주기를 조건주문·safety class로 확장
- `engine-safety`: ExecutionGateway 봉인과 자동화 게이트 기동 인터록을 강화

### New Capabilities

(없음 — 기존 두 capability의 계약 확장이다)

## Impact

- Affected code: `internal/execgw`(멱등키·safety class·RiskIntent·예약·nonce), `internal/journal`(조건주문 테이블·예상 주문·예약·nonce, v5 단일 원자 마이그레이션), `internal/app/engine`(Gateway 구성·인터록 강화·봉인), `internal/trading`·`internal/official`(조건주문 어댑터의 식별자 반환, 엔진용 Gateway 경유 진입점), `internal/filldetect`·`internal/reconcile`(발동 주문 편입)
- 선행: P1 archive 완료(됨)
- 후행: `add-core-domain`(재범위)이 이 계약 위에서 Guardian 판정을 구현
- 병행: `verify-execution-capability`가 조건주문 능력 + **멱등키 실동작**(재생 응답 내용·유효 창·계좌 스코프)을 attestation에 추가
- **upstream 파일 수정 예정**: `internal/trading/conditional.go`, `internal/official/conditional_writes.go`, `internal/app/engine/engine.go`. 전부 High-risk → Pre-Edit 전문 선언 필요
- 이 계약이 완성되기 전에는 자동 **진입**이 불가하다. 게이트는 기본 OFF를 유지한다
