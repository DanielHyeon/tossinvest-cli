# order-execution Specification (delta)

## ADDED Requirements

### Requirement: 원자적 위험 예약
계좌 전체에 걸친 한도(총 개방 노출, 일일 손실, 현금)의 판정과 그 결과의 예약은 하나의 journal 트랜잭션 안에서 수행되어야 한다(SHALL). 서로 다른 심볼에 대한 동시 결정이 같은 스냅샷을 각각 통과해 합산 한도를 초과하는 것은 허용되지 않는다(SHALL NOT).

브로커 조회를 이 트랜잭션 안에서 수행해서는 안 된다(SHALL NOT — journal은 단일 커넥션이므로 네트워크 왕복 동안 모든 mutation 기록이 막힌다). 스냅샷은 트랜잭션 밖에서 수집하고, 안에서 as-of·staleness를 검증한 뒤 예약을 삽입하며, 불충족이면 롤백하고 재수집한다(SHALL). 재수집은 횟수 상한(기본 3회)과 총 데드라인을 가지며 초과 시 fail-closed로 거부한다(SHALL). 예약 산술은 decimal 문자열 연산이며 float 누적을 사용하지 않는다(SHALL NOT).

예약 해제의 정본은 **attempt의 브로커 종결 상태 도달**이다(SHALL): FILLED·CANCELED·REJECTED·NOT_DISPATCHED·FAILED_CONFIRMED — 체결 없이 장 마감에 만료되어 종결된 주문을 포함한다. 그 외 해제는 nonce 미소비 상태의 결정 만료뿐이며, nonce 소비 후 만료는 예약을 해제하지 않는다(SHALL NOT — 주문이 접수됐을 수 있다). UNRESOLVED_IN_DOUBT의 예약은 운영자 해소로만 풀린다(SHALL). 일일 손실 예약은 시장별 거래일 경계에서 소멸한다(SHALL).

#### Scenario: 동시 다심볼 결정
- **WHEN** 총 개방 노출 한도의 잔여분이 1건분만 남은 상태에서 서로 다른 두 심볼의 결정이 동시에 요청되면
- **THEN** 하나만 예약에 성공하고 다른 하나는 한도 초과로 거부된다

#### Scenario: 미체결 만료 주문의 예약 해제
- **WHEN** 예약을 보유한 주문이 체결 없이 장 마감에 만료되어 종결 상태로 관측되면
- **THEN** 예약이 BROKER_TERMINAL 사유로 해제되어 한도가 누수되지 않는다

#### Scenario: nonce 소비 후 만료
- **WHEN** nonce가 소비된 뒤 응답이 유실되고 결정 유효 시간이 지나면
- **THEN** 예약은 만료를 이유로 해제되지 않고 해소 완료까지 유지된다

#### Scenario: 재수집 상한 초과
- **WHEN** 스냅샷 as-of 검증이 상한 횟수까지 연속 실패하면
- **THEN** 결정 발급이 fail-closed로 거부되고 사유가 기록된다

### Requirement: RECONCILE 상태
권위 값의 불일치는 산식으로 보정하지 않고 RECONCILE 상태로 전이해야 한다(SHALL). 진입 조건: 브로커 보유·매도가능 조회가 불가하거나 staleness 한계를 초과, 로컬 파생 수량과 브로커 스냅샷의 불일치, 같은 브로커 식별자가 상충하는 계좌·심볼 컨텍스트에 출현.

RECONCILE 상태에서는 신규 진입과 수량 확대가 금지되고(SHALL NOT), 읽기·계좌 동기화·운영자 확인은 허용되며, 위험 축소는 **확정 하한 수량**으로만 허용된다(SHALL — 수량이 불확실해도 과소 청산은 안전한 방향이며, §0.3 손절 즉시성은 유지되어야 한다). 해제는 재조회 일치와 원인 기록을 요구한다(SHALL).

#### Scenario: 수량 불일치 시 청산 요청
- **WHEN** RECONCILE 상태에서 청산이 요청되면
- **THEN** 확정 하한 수량까지만 제출되고 초과분은 해소 후로 보류된다

#### Scenario: 진입 시도
- **WHEN** RECONCILE 상태에서 진입 의도가 평가되면
- **THEN** 거부되고 RECONCILE 사유가 기록된다

### Requirement: 총계 한도의 계산 계약
총 개방 노출·일일 손실·현금은 계산 계약이 정의되어야 하며(SHALL), 정의되지 않은 양에 예약을 걸어서는 안 된다(SHALL NOT). 계약은 다음을 포함한다: 각 값의 권위 데이터, 미체결 주문의 평가 가격, 통화 정규화와 환율 권위·staleness 허용치, 실현/미실현 손익의 포함 범위, 수수료·세금 반영, 시장별 거래일 경계(P1 시간 규율 준수), 예약 합산 방식, 외부 수동 거래의 취급. 입력이 stale하거나 미지이면 fail-closed로 진입을 거부한다(SHALL). 수치는 이 계약의 대상이 아니다(2d).

#### Scenario: 환율 stale
- **WHEN** 외화 자산의 원화 환산에 필요한 환율이 staleness 한계를 넘으면
- **THEN** 총 개방 노출 판정이 fail-closed로 진입을 거부한다

### Requirement: 브로커 식별자의 opaque 취급
브로커가 발급하는 주문 식별자는 opaque token으로 취급해야 한다(SHALL — openapi는 `orderId`에 형태·패턴을 계약하지 않는다). 클라이언트는 형태·prefix·길이 패턴을 검증하거나 해석해서는 안 되며(SHALL NOT), null·빈 문자열만 거부한다(SHALL). 식별자는 응답 원문 그대로(변환 없이) 계좌 스코프와 함께 저장하고(SHALL), 생성 응답의 식별자는 상세조회 round-trip으로 실재를 확인한다(SHALL). 같은 식별자가 상충하는 계좌·심볼 컨텍스트에 나타나면 RECONCILE로 전이한다(SHALL).

#### Scenario: 예상 밖 형식의 식별자
- **WHEN** 생성 응답이 지금까지와 전혀 다른 형식의 orderId를 반환하면
- **THEN** 파싱·검증 없이 opaque 값으로 저장되고 round-trip 조회로 실재가 확인된다

#### Scenario: 빈 식별자
- **WHEN** 생성 응답의 orderId가 비어 있으면
- **THEN** ACK로 처리되지 않고 IN_DOUBT 해소가 시작된다

### Requirement: 체결 정정 이벤트
누적 체결 수량이 동일한데 평균 체결가 또는 체결 금액이 변경된 관측은 수량을 재반영하지 않고 EXECUTION_CORRECTION 이벤트로 기록해야 한다(SHALL — 평균 체결가는 부분 체결마다 바뀌는 값이므로(openapi) 중복 판정 키에 포함해서는 안 된다(SHALL NOT)). 누적 수량 감소는 P1 규칙대로 fail-closed를 유지한다.

#### Scenario: 수량 동일·평균가 변경
- **WHEN** 같은 주문의 누적 수량이 동일하고 평균 체결가만 달라진 스냅샷이 관측되면
- **THEN** 수량 delta 없이 정정 이벤트가 기록되고 포지션 수량은 변하지 않는다

## MODIFIED Requirements

### Requirement: IN_DOUBT 해소
IN_DOUBT 해소의 목적은 **정체 회수**다 — 주문이 접수됐는지, 접수됐다면 어떤 식별자인지 확정하는 것이며, 정체를 모르는 채 다시 주문을 내는 것은 금지된다(SHALL NOT).

공식 API의 주문 생성은 클라이언트 제공 멱등키를 지원한다(openapi `clientOrderId`: "동일 값으로 재요청 시 이전 주문 결과를 그대로 재반환합니다", 유효 10분). 해소는 다음 순서를 따른다(SHALL):

1. **멱등 재생**: RECORDED 단계에 영속된 동일 키·동일 wire body의 재요청 응답에서 식별자를 회수한다. 재생은 재시도가 아니다 — 같은 키는 유효 창 안에서 새 주문을 만들 수 없다. 적용 조건: (a) 실동작이 능력 attestation으로 확인됨 `[미측정 — 2b 전 비활성]`, (b) `elapsed(전송 시작) < TTL − margin`(margin 기본 60초 — 경과 근거가 로컬 시계이므로 마진 없는 경계 사용 금지(SHALL NOT)), (c) 재생 횟수 상한(기본 2회) 이내. 재생은 새 attempt를 만들지 않고 같은 attempt의 해소 기록(횟수·시각)으로 남는다(SHALL).
2. **조회 대조 (폴백)**: 창 초과, 미검증, `idempotency-key-conflict`, 멱등키를 받지 않는 mutation(취소·정정)은 journal fingerprint로 미체결·종결 양쪽 목록을 pagination 완주하며 대조한다. 조회 응답은 멱등키를 싣지 않으므로 키로 매칭할 수 없다(SHALL NOT — 두 절차는 서로를 대체하지 못한다).
3. **부재 판정**: 최소 관찰 기간에 걸친 연속 N회(기본 3회) 안정화 조회 + 매수가능금액·보유수량 delta 교차 확인 후에만.
4. **해소 불능**: UNRESOLVED_IN_DOUBT로 해당 심볼 신규 진입 영구 차단, 운영자 해소만 허용.

재생 경로는 Gateway의 해소 전용 진입점으로만 수행되어야 하며(SHALL), 그 진입점은 저장된 wire body의 재전송만 가능하고 새 본문을 구성할 수 없어야 한다(SHALL NOT — 두 번째 제출 문이 되지 않음을 정적 테스트로 증명). 조회 대조의 유일 매칭을 위해 심볼당 in-flight mutation 1개 제한을 유지한다(SHALL). 단, 미해소 EXPOSURE_RAISING attempt가 같은 심볼의 RISK_REDUCING mutation을 차단해서는 안 되며(SHALL NOT — §0.3), 그 경우 RISK_REDUCING 수량은 RECONCILE 상태 규칙(확정 하한)을 따른다.

#### Scenario: 멱등 재생으로 정체 회수
- **WHEN** 능력 검증 완료 상태에서 주문 제출 응답이 유실되고 TTL−margin 이내에 해소가 시작되면
- **THEN** 저장된 wire body의 재요청 응답에서 orderId를 회수해 CONFIRMED로 종결하며 두 번째 주문은 생성되지 않는다

#### Scenario: 마진 경계 초과
- **WHEN** 전송 시작 후 경과 시간이 TTL−margin을 넘긴 뒤 해소가 시작되면
- **THEN** 재생을 사용하지 않고 조회 대조로 진행한다

#### Scenario: 재생 응답도 유실
- **WHEN** 재생 요청의 응답이 상한 횟수까지 유실되면
- **THEN** 재생 기록이 남고 조회 대조로 전환되며 자동 재제출은 발생하지 않는다

#### Scenario: 단발 부재 조회
- **WHEN** 첫 조회에서 주문이 보이지 않으면
- **THEN** FAILED로 판정하지 않고 안정화 재조회를 계속한다

#### Scenario: 해소 불능
- **WHEN** 관찰 기간 내 존재도 부재도 증명되지 않으면
- **THEN** UNRESOLVED_IN_DOUBT로 표기되어 해당 심볼의 신규 진입이 영구 차단되고 운영자 알림이 발송된다 (위험 축소 경로는 계속 동작)

### Requirement: MutationAttempt 수명주기
각 MutationAttempt는 RECORDED → DISPATCH_STARTED → (ACKED | IN_DOUBT) → 종결(CONFIRMED | NOT_DISPATCHED | FAILED_CONFIRMED | UNRESOLVED_IN_DOUBT) 단계를 가져야 한다(SHALL). RECORDED는 fsync 완료 후에만 DISPATCH_STARTED로 진행하며(SHALL), RECORDED 단계에서 멱등키(`clientOrderId`)와 canonical wire body·serializer version이 함께 불변 영속된다(SHALL — 재생은 저장본만 사용하며 구조화 필드에서 본문을 재구성해서는 안 된다(SHALL NOT)). 멱등키는 결정에 결속된 결정적 값이며 확인 토큰의 canonical 입력에는 포함되지 않는다(SHALL NOT — CLI confirm token 무변경). 재시작 시 RECORDED에서 멈춘 attempt는 NOT_DISPATCHED로 안전 종결하고, DISPATCH_STARTED 이후는 해소 완료 전까지 차단 대상으로 취급하되 차단 범위는 safety class 규칙을 따른다(SHALL).

#### Scenario: 전송 시작 전 크래시
- **WHEN** RECORDED까지만 기록된 attempt가 재시작 시 발견되면
- **THEN** NOT_DISPATCHED로 종결되고 어떤 차단도 발생하지 않는다

#### Scenario: 전송 중 크래시
- **WHEN** DISPATCH_STARTED로 기록된 attempt가 재시작 시 발견되면
- **THEN** 영속된 멱등키·wire body로 해소 절차가 시작된다

#### Scenario: 직렬화 규칙 변경 후 재생
- **WHEN** 바이너리 업데이트로 직렬화 규칙이 바뀐 뒤 이전 attempt의 재생이 필요하면
- **THEN** 저장된 wire body가 그대로 사용되어 본문 불일치(idempotency-key-conflict)가 발생하지 않는다

### Requirement: 브로커 상태 파생
주문 종결 상태는 공식 API의 원시 status만이 아니라 `(status, canceledAt, execution.filledQuantity, quantity, lineage)`에 대한 우선순위 파생 함수로 결정해야 한다(SHALL). 파생은 문서화된 OrderStatus 전체를 다룬다(SHALL — openapi): PENDING, PENDING_CANCEL, PENDING_REPLACE, PARTIAL_FILLED, FILLED, CANCELED, REJECTED, CANCEL_REJECTED, REPLACE_REJECTED, REPLACED. 미지의 status 값은 UNKNOWN_BROKER_STATE로 fail-closed 처리한다(SHALL: 해당 심볼 신규 진입 차단 + 알림).

`CANCEL_REJECTED`·`REPLACE_REJECTED`는 "별도 주문 레코드로 생성됨"(openapi — 원주문은 이전 상태로 복귀). 취소·정정의 해소 절차는 이 별도 레코드의 존재를 인지해야 하며(SHALL), 레코드의 구체 형태는 `[미측정 — 2b]`이므로 인지·귀속에 실패한 레코드는 외부 주문으로 분류하지 않고 RECONCILE로 처리한다(SHALL — fail-closed).

#### Scenario: CLOSED + canceledAt 존재
- **WHEN** status=CANCELED이고 canceledAt이 설정된 주문을 파생하면
- **THEN** CANCELLED로 판정된다 (filledQuantity>0이면 부분체결 후 취소로 기록)

#### Scenario: 취소 거부 레코드 관측
- **WHEN** 취소 요청 후 CANCEL_REJECTED 상태의 별도 주문 레코드가 관측되면
- **THEN** 원주문은 이전 상태로 복귀한 것으로 파생되고, 별도 레코드는 외부 주문으로 분류되지 않는다

#### Scenario: 미지의 status 값
- **WHEN** 파생 함수가 알 수 없는 status 문자열을 받으면
- **THEN** UNKNOWN_BROKER_STATE로 fail-closed 처리되고 알림이 발송된다
