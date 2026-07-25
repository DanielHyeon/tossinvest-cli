# Change: extend-execution-contract

## Why

add-core-domain의 proposal-freeze 리뷰(45건, `../add-core-domain/review.md`)가 하나의 구조적 오류를 드러냈다: 무인 자동매매의 안전 근거 전체가 "네이티브 조건주문이 브로커에 상주한다"에 걸려 있는데, **조건주문 경로는 P1이 만든 안전 아키텍처 바깥에 통째로 있다.**

- `trading.Service.ConditionalPlace`(internal/trading/conditional.go:123)는 확인 토큰 검사 후 브로커를 직접 호출한다 — journal 선기록·MutationAttempt·nonce·IN_DOUBT·in-flight 제한이 전부 없다. 응답이 유실되면 saga는 "에러"만 보고, 재제출은 브로커측 stop을 중복시켜 발동 시 oversell을 만든다.
- `journal.NetPositions`(internal/journal/fills.go:457-463)는 `JOIN intents`로 체결을 귀속한다. 조건주문 발동은 로컬 intent 없는 신규 브로커 주문이므로 **외부 주문으로 분류되어 포지션이 영원히 CLOSED에 도달하지 못한다.** "브로커 상주라서 안전하다"와 "로컬 intent 없는 주문은 외부"가 구조적으로 양립하지 않는다.
- `engine.RequiredEndpoints()`(internal/app/engine/interlock.go:81-91)에 조건주문 엔드포인트가 없다. 손절 메커니즘이 한 번도 실증되지 않은 채 완전히 유효한 attestation으로 자동매매가 켜질 수 있다.
- `verifyGate`(interlock.go:194-229)는 trading policy를 보지 않는다. `policy.Sell=false` 또는 `policy.Conditional=false`여도 게이트는 ON+verified가 되어 **매수는 하는데 손절도 청산도 못 하는 naked long**이 성립한다.

이 change는 강제 장치(레일)만 다룬다. 판단 정책(Guardian 판정 체인·사이징·한도 수치)은 후행 add-core-domain이 맡는다. 경계 원칙: **여기는 실패해도 안전한 레일, 거기는 그 레일 위의 판단.** P1이 오늘 그렇게 하듯 이 change는 합성 GuardianDecision만으로 완전히 테스트된다.

## What Changes

- **조건주문의 Gateway 편입**: conditional PLACE/CANCEL/MODIFY를 `journal.MutationKind`로 승격. journal 선기록 → DISPATCH_STARTED → 분류 → IN_DOUBT 해소를 place/cancel/amend와 동일하게 적용하고, 조건주문 전용 fingerprint(계좌·심볼·트리거가·수량·유형·제출 창)로 OPEN/CLOSED 양쪽을 pagination 대조한다.
- **브로커 발동 주문의 lineage 편입**: 조건주문 등록 시 `conditionalOrderId`를 기록하고, 발동으로 생성된 브로커 주문번호를 lineage와 추적 집합에 편입해 그 체결이 포지션에 귀속되게 한다. `NetPositions`의 귀속 규칙을 확장한다.
- **mutation safety class**: 노출 증가(entry)와 위험 감소(보호 생성·증량, reduce-only 청산, cancel)를 별도 클래스로 정의하고 직렬화 규칙을 분리한다. 진입 IN_DOUBT가 같은 심볼의 손절 제출을 막지 않는다(§0.3). HALT_ALL에서도 위험 감소 클래스는 통과한다.
- **RiskIntent 결합**: entry·stop·target·방향·계좌·시장·정책 버전을 담은 불변 `RiskIntent`의 canonical hash를 GuardianDecision과 주문 provenance에 결합해, 손절 데이터를 바꿔치기한 제출이 Gateway 재검증을 통과하지 못하게 한다.
- **한도 fail-closed**: 필수 한도별 명시적 configured 비트 + 양수·유한 검증 + 통화 일치. 하나라도 누락·0·NaN·Inf면 기동 거부(현재는 수량·notional이 **둘 다** 0일 때만 거부). 총 개방 노출·일일 손실을 한도 집합에 추가한다.
- **원자적 위험 예약**: 결정 발급과 노출·현금 예약을 하나의 `BEGIN IMMEDIATE` 트랜잭션으로 묶어 서로 다른 심볼의 동시 결정이 합산 한도를 뚫지 못하게 한다. nonce 소비·실패·만료·체결·취소에 따른 예약 수명주기를 정의한다.
- **청산·보호 발급자 한도 계약**: 위험 감소 mutation의 결정은 빈 한도 스냅샷을 싣는다(`flatten.decisionFor` 패턴의 계약화). 큰 청산이 주문 한도로 거부되지 않음을 테스트한다.
- **journal 기반 NonceStore**: 프로세스 수명을 넘는 결정에 대비해 in-memory 저장소를 교체한다(`guardian.go:124` 주석이 이미 지시).
- **게이트 전제조건 강화**: trading policy(`Sell`·`Conditional`·`AllowLiveOrderActions`) 검증, Guardian 한도를 감사된 `gateLimits(gate)` 단일 출처에서 구성, `RequiredEndpoints()`에 조건주문 엔드포인트 추가.

## Capabilities

### Modified Capabilities

- `order-execution`: MutationAttempt 수명주기와 IN_DOUBT 해소를 조건주문·safety class로 확장
- `engine-safety`: ExecutionGateway 봉인과 자동화 게이트 기동 인터록을 강화

### New Capabilities

(없음 — 기존 두 capability의 계약 확장이다)

## Impact

- Affected code: `internal/execgw`(신규 mutation kind·safety class·RiskIntent·예약·nonce), `internal/journal`(조건주문 attempt·발동 주문 lineage·예약 테이블·nonce 테이블, v5 additive), `internal/app/engine`(인터록 강화·Gateway 배선), `internal/trading`(조건주문의 Gateway 경유 경로 — CLI 표면의 기존 확인 토큰 경로는 보존)
- 선행: P1 archive 완료(됨)
- 후행: `add-core-domain`(재범위)이 이 계약 위에서 Guardian 판정을 구현. `verify-execution-capability`는 조건주문 등록·트리거 관측·modify 원자성·시장별 지원을 attestation 계약에 추가
- **upstream 파일 수정 예정**: `internal/trading/conditional.go`(엔진 경로 추가, CLI 경로 무변경). High-risk → Pre-Edit 전문 선언 필요
- 이 계약이 완성되기 전에는 자동 **진입**이 불가하다. 게이트는 기본 OFF를 유지한다
