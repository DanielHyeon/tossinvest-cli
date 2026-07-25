# order-execution Specification (delta)

## ADDED Requirements

### Requirement: 조건주문의 형제 수명주기
조건주문의 등록·취소·정정은 일반 주문과 **같은 단계**를 거치되 **자체 저장소**를 가져야 한다(SHALL): journal 선기록 → DISPATCH_STARTED → 결과 분류 → 필요 시 해소. 조건주문 식별자를 일반 주문의 브로커 주문번호 컬럼에 저장해서는 안 된다(SHALL NOT — 그 컬럼은 체결 감지 추적 집합과 reconciliation의 로컬 미체결 목록으로 흘러가며, 조건주문 식별자는 일반 주문 조회에 유효하지 않아 체결 감지 사이클 전체와 진입 경로를 영구히 실패시킨다). 조건주문 intent는 유형(SINGLE/OCO/OTO)·만료일·leg별 방향·트리거가·주문가를 자체 컬럼으로 보존해야 하며(SHALL), 해소 시점의 fingerprint 재계산이 이 컬럼들로 가능해야 한다. 엔진 경로에서 확인 토큰만 검사하고 브로커를 직접 호출하는 제출은 존재해서는 안 된다(SHALL NOT). CLI/MCP 표면의 기존 확인 토큰 경로는 변경 대상이 아니다.

#### Scenario: 조건주문 등록 후 체결 감지 계속 동작
- **WHEN** 보호주문이 성공적으로 등록되고 다음 체결 감지 사이클이 수행되면
- **THEN** 감지 사이클이 정상 완료되며, 조건주문 식별자로 인한 조회 실패가 발생하지 않는다

#### Scenario: 조건주문 등록 후 reconciliation
- **WHEN** 활성 조건주문이 있는 상태에서 reconciliation이 수행되면
- **THEN** 조건주문이 로컬 미체결 주문으로 오인되어 진입이 차단되지 않는다

#### Scenario: OTO 조건주문의 해소 fingerprint
- **WHEN** 두 leg의 방향이 서로 다른 조건주문의 해소가 필요하면
- **THEN** 저장된 leg별 컬럼에서 fingerprint를 재계산할 수 있다

### Requirement: 발동 주문의 편입
조건주문 발동으로 생성된 주문은 일반 주문으로서 기존 추적·체결·reconciliation 경로에 편입되어야 한다(SHALL). 조건주문 등록이 확정되면 leg별 예상 주문 레코드(조건주문 식별자·leg 식별·심볼·방향·최대 수량·상태)를 기록하고(SHALL), 조건주문 상태 조회를 주기적으로 수행해 leg의 상태와 발동 주문 식별자를 관측한다(SHALL). 발동 주문 식별자가 관측되면 그 주문을 조건주문으로의 lineage와 함께 추적 대상에 편입한다(SHALL). 발동 주문에 대해 intent를 소급 생성해서는 안 된다(SHALL NOT — 내지 않은 주문에 의도를 부여하면 provenance가 거짓이 된다).

예상 주문 레코드는 일회성이고 유계여야 한다(SHALL): leg가 발동하면 그 레코드는 발동 주문 식별자로 소비되어 이후 매칭에 사용되지 않고, 브로커측 자동 취소(OCO 반대편)와 만료는 상태 조회로 관측해 종결시키며, 종결은 체결 귀속 완료 이후에 수행하고 tombstone을 보존 기간 동안 유지한다. 낡은 예상 주문이 실제 외부 주문을 엔진 포지션으로 흡수해서는 안 된다(SHALL NOT — 위험한 실패 방향은 매칭 실패가 아니라 거짓 매칭이다).

조건주문 상태 조회는 체결 감지와 rate limit 그룹을 공유하므로 예산을 명시해야 하며(SHALL), 이 조회의 실패가 보호 경로를 차단해서는 안 된다(SHALL NOT).

#### Scenario: 브로커측 손절 발동으로 포지션 청산
- **WHEN** 브로커에 등록된 stop 조건주문이 발동해 전량 체결되면
- **THEN** 발동 주문이 추적에 편입되어 그 체결이 포지션에 반영되고, 영구 불일치로 보고되지 않는다

#### Scenario: OCO 반대편 자동 취소
- **WHEN** OCO의 한쪽 leg가 발동해 반대편이 브로커측에서 자동 취소되면
- **THEN** 상태 조회가 이를 관측해 해당 예상 주문 레코드를 종결시킨다

#### Scenario: 만료된 예상 주문과 수동 매도
- **WHEN** 조건주문이 만료된 뒤 운영자가 같은 심볼을 수동 매도하면
- **THEN** 그 체결이 엔진 포지션으로 흡수되지 않고 외부 거래로 분류된다

### Requirement: Mutation Safety Class와 직렬화
모든 mutation은 세 safety class 중 하나로 분류되어야 한다(SHALL):

- **EXPOSURE_RAISING** — 진입 제출, 노출 증가
- **RISK_REDUCING** — 보호주문 생성·증량, reduce-only 청산, 미체결 진입의 취소
- **PROTECTION_WEAKENING** — 활성 보호주문의 취소·수량 축소, 청산 주문의 취소

취소를 일괄로 위험 감소로 분류해서는 안 된다(SHALL NOT — 활성 보호주문의 취소는 보호를 제거하므로 위험을 증가시킨다). 분류 판별자는 mutation 종류와 방향이 아니라 대상이 해당 포지션의 보호인지 여부이며, 예상 주문 레코드가 그 판정의 권위다(SHALL).

직렬화는 클래스별로 분리된다(SHALL): EXPOSURE_RAISING은 심볼당 1건, RISK_REDUCING도 심볼당 1건이되 EXPOSURE_RAISING의 in-flight·IN_DOUBT에 의해 차단되지 않으며(SHALL NOT — §0.3), PROTECTION_WEAKENING은 대상 조건주문 식별자 단위로 직렬화한다. 클래스별 심볼 단위 제한은 해소 시 유일 매칭의 근거이므로 제거할 수 없다(SHALL NOT — 제거하면 해소 절차가 다른 주문을 이 attempt의 결과로 확정하는 거짓 CONFIRMED가 가능해진다).

PROTECTION_WEAKENING은 가장 엄격한 규칙을 받는다(SHALL): 원자적 교체의 일부이고 무보호 창이 측정·유계일 때만 허용되며, audit 기록을 요구한다.

#### Scenario: 진입 IN_DOUBT 중 손절 제출
- **WHEN** 같은 심볼의 진입 attempt가 IN_DOUBT인 상태에서 보호주문 제출이 요청되면
- **THEN** 제출이 차단되지 않고 수행된다

#### Scenario: 활성 보호의 취소 시도
- **WHEN** 포지션의 유일한 활성 보호주문을 취소하는 mutation이 요청되면
- **THEN** PROTECTION_WEAKENING으로 분류되어 위험 감소 경로의 면제를 받지 못한다

#### Scenario: 동시 위험 감소 mutation
- **WHEN** 같은 심볼에 보호주문 생성과 청산 제출이 동시에 요청되면
- **THEN** 하나만 in-flight로 진행하고 다른 하나는 대기한다

### Requirement: 청산 수량 예약
매도 mutation의 수량은 원자적 예약으로 통제되어야 하며(SHALL), 단건 상한으로는 충분하지 않다(SHALL NOT — 개별 주문이 각각 상한을 통과해도 합계가 보유 수량을 초과할 수 있다). 가용 매도 수량은 `보유 수량 − 미체결 매도 주문 − 대기 중 조건주문의 예약 수량 − 유효한 동시 예약`으로 계산한다(SHALL). 브로커 주문 요청에 reduce-only 필드가 없으므로 이 계산이 유일한 oversell 방어다.

수량의 권위는 가장 최근 브로커 스냅샷이며(SHALL), 로컬 파생 보유수량을 상한을 높이는 근거로 사용해서는 안 된다(SHALL NOT — 로컬 파생은 외부 주문을 제외하므로 실제보다 클 수 있다). 브로커 스냅샷이 staleness 한계를 넘으면 critical 알림과 함께 가장 최근 값을 보수적으로 사용하고, 스냅샷이 전혀 없으면 로컬 매도를 수행하지 않는다(SHALL NOT — 그 구간의 보호는 브로커측 조건주문이 담당한다).

#### Scenario: 보호와 청산의 합계 초과
- **WHEN** 보유 100주에 대해 100주 보호주문이 등록된 상태에서 100주 청산이 요청되면
- **THEN** 가용 매도 수량이 예약에 의해 0이므로 청산이 거부되거나 보호 취소 후로 순서화된다

#### Scenario: 동시 청산 요청
- **WHEN** 같은 심볼에 대해 두 경로가 동시에 전량 청산을 요청하면
- **THEN** 하나만 예약에 성공하고 다른 하나는 가용 수량 부족으로 거부된다

#### Scenario: 계좌 스냅샷 부재
- **WHEN** 브로커 계좌 스냅샷을 한 번도 얻지 못한 상태에서 로컬 청산이 요청되면
- **THEN** 로컬 매도가 수행되지 않고 critical 알림이 발송된다

### Requirement: 원자적 위험 예약
계좌 전체에 걸친 한도(총 개방 노출, 일일 손실, 현금)의 판정과 그 결과의 예약은 하나의 journal 트랜잭션 안에서 수행되어야 한다(SHALL). 서로 다른 심볼에 대한 동시 결정이 같은 스냅샷을 각각 통과해 합산 한도를 초과하는 것은 허용되지 않는다(SHALL NOT).

브로커 조회를 이 트랜잭션 안에서 수행해서는 안 된다(SHALL NOT — journal은 단일 커넥션이므로 네트워크 왕복 동안 모든 mutation 기록이 막히며, 여기에는 이 계약이 지키려는 보호 경로가 포함된다). 스냅샷은 트랜잭션 밖에서 수집하고, 트랜잭션 안에서는 스냅샷의 as-of 조건과 staleness 한계를 검증한 뒤 예약을 삽입하며, 조건 불충족이면 롤백하고 재수집한다(SHALL).

예약 해제는 다음에서만 일어난다(SHALL): 체결 또는 취소 확정, 제출 실패 확정, 결정 만료(단 nonce가 소비되지 않은 경우에 한한다 — 소비 후에는 주문이 접수됐을 수 있으므로 만료가 예약을 풀어서는 안 된다). IN_DOUBT는 해소 전까지 예약을 유지하고(SHALL), UNRESOLVED_IN_DOUBT의 예약은 운영자 해소로만 해제된다(SHALL). 일일 손실 예약은 거래일 경계에서 소멸해야 한다(SHALL — 그러지 않으면 주차된 attempt 하나가 다음 거래일을 조용히 정지시킨다).

#### Scenario: 동시 다심볼 결정
- **WHEN** 총 개방 노출 한도의 잔여분이 1건분만 남은 상태에서 서로 다른 두 심볼의 결정이 동시에 요청되면
- **THEN** 하나만 예약에 성공하고 다른 하나는 한도 초과로 거부된다

#### Scenario: nonce 소비 후 만료
- **WHEN** nonce가 소비된 뒤 응답이 유실되고 결정 유효 시간이 지나면
- **THEN** 예약은 만료를 이유로 해제되지 않고 해소 완료까지 유지된다

#### Scenario: 거래일 경계
- **WHEN** 일일 손실 예약을 보유한 채 거래일이 바뀌면
- **THEN** 그 예약은 소멸하고 새 거래일의 한도가 온전히 사용 가능하다

### Requirement: 총계 한도의 계산 계약
총 개방 노출·일일 손실·현금은 계산 계약이 정의되어야 하며(SHALL), 정의되지 않은 양에 예약을 걸어서는 안 된다(SHALL NOT). 계약은 다음을 모두 포함한다: 각 값의 권위 데이터, 미체결 주문과 대기 중 조건주문의 평가 가격, 통화 정규화와 환율 권위·staleness 허용치, 실현 손익과 미실현 손익의 포함 범위, 수수료·세금 반영, 시장별 거래일 경계(P1 시간 규율 준수), 예약의 합산 방식, 외부 수동 거래의 취급. 입력 중 하나라도 stale하거나 미지이면 fail-closed로 진입을 거부한다(SHALL). 수치 자체는 이 계약의 대상이 아니다.

#### Scenario: 환율 stale
- **WHEN** 외화 자산의 원화 환산에 필요한 환율이 staleness 한계를 넘으면
- **THEN** 총 개방 노출 판정이 fail-closed로 진입을 거부한다

## MODIFIED Requirements

### Requirement: IN_DOUBT 해소
IN_DOUBT 해소의 목적은 **정체 회수**다 — 주문이 접수됐는지, 접수됐다면 어떤 식별자인지를 확정하는 것이며, 정체를 모르는 채 다시 주문을 내는 것은 금지된다(SHALL NOT).

해소는 다음 순서를 따른다(SHALL):

1. **멱등 재생**: 공식 API의 생성 엔드포인트는 클라이언트 제공 멱등키를 지원하며 동일 키·동일 본문의 재요청은 새 주문을 만들지 않고 이전 결과를 반환한다. 엔진은 mutation attempt마다 결정적 멱등키를 발행해 전송 시작 이전에 journal에 영속해야 하고(SHALL), 응답 유실 시 동일 본문·동일 키의 재요청 응답에서 식별자를 회수한다. 이 재요청은 재시도가 아니다 — 새 주문을 만들 수 없기 때문이다. 멱등키의 실제 동작(재생 응답 내용·유효 창·계좌 스코프)이 실계좌 능력 검증으로 확인되기 전에는 이 단계를 사용하지 않는다(SHALL NOT).
2. **조회 대조 (폴백)**: 멱등키 유효 창 경과, 능력 미검증, 키 충돌, 또는 멱등키를 받지 않는 mutation(취소·정정)인 경우 journal에 저장된 fingerprint로 미체결과 종결 **양쪽** 목록을 pagination 완주하며 대조한다. 어떤 조회 응답도 멱등키를 싣지 않으므로 조회는 키로 매칭할 수 없다(SHALL NOT — 이것이 1단계와 2단계가 서로를 대체하지 못하는 이유다).
3. **부재 판정**: 최소 관찰 기간에 걸친 연속 N회(기본 3회) 안정화 조회 + 매수가능금액·보유수량 delta 교차 확인 후에만 부재로 판정한다.
4. **해소 불능**: 증명 불가 시 UNRESOLVED_IN_DOUBT로 해당 심볼 신규 진입을 영구 차단하고 운영자 해소만 허용한다.

2단계의 유일 매칭을 보장하기 위해 엔진은 safety class별로 심볼당 in-flight mutation을 1개로 제한한다(SHALL). 이 제한을 제거하면 다른 주문이 이 attempt의 결과로 확정되는 거짓 CONFIRMED가 발생할 수 있다(SHALL NOT 제거).

#### Scenario: 멱등 재생으로 정체 회수
- **WHEN** 능력 검증이 완료된 상태에서 주문 제출 응답이 유실되면
- **THEN** 동일 키·동일 본문 재요청의 응답에서 주문 식별자를 회수해 CONFIRMED로 종결하며, 두 번째 주문은 생성되지 않는다

#### Scenario: 멱등 유효 창 경과
- **WHEN** 멱등키 유효 창이 지난 뒤 해소를 시작하면
- **THEN** 재생을 사용하지 않고 조회 대조 절차로 진행한다

#### Scenario: 제출 응답 유실 후 주문이 2페이지에 존재
- **WHEN** 조회 대조에서 대상 주문이 목록 2페이지 이후에 있으면
- **THEN** pagination 완주로 발견되어 CONFIRMED로 종결된다

#### Scenario: 단발 부재 조회
- **WHEN** 첫 조회에서 주문이 보이지 않으면
- **THEN** FAILED로 판정하지 않고 안정화 재조회를 계속한다

#### Scenario: 해소 불능
- **WHEN** 관찰 기간 내 존재도 부재도 증명되지 않으면
- **THEN** UNRESOLVED_IN_DOUBT로 표기되어 해당 심볼의 신규 진입이 영구 차단되고 운영자 알림이 발송된다 (보호·청산 경로는 계속 동작)

### Requirement: MutationAttempt 수명주기
각 MutationAttempt는 RECORDED → DISPATCH_STARTED → (ACKED | IN_DOUBT) → 종결(CONFIRMED | NOT_DISPATCHED | FAILED_CONFIRMED | UNRESOLVED_IN_DOUBT) 단계를 가져야 한다(SHALL). 이 단계 모델은 조건주문 mutation에도 동일하게 적용되나 저장소는 분리된다(SHALL — 조건주문의 형제 수명주기 참조). RECORDED는 fsync 완료 후에만 DISPATCH_STARTED로 진행하며(SHALL), 멱등키는 RECORDED 단계에서 함께 영속된다(SHALL). 재시작 시 RECORDED 단계에서 멈춘 attempt는 NOT_DISPATCHED로 안전 종결하고, DISPATCH_STARTED 이후 단계는 해소 절차 완료 전까지 차단 대상으로 취급한다(SHALL). 다만 차단의 범위는 safety class 규칙을 따르며, 미해소 EXPOSURE_RAISING attempt가 같은 심볼의 RISK_REDUCING mutation을 차단해서는 안 된다(SHALL NOT).

#### Scenario: 전송 시작 전 크래시
- **WHEN** RECORDED까지만 기록된 attempt가 재시작 시 발견되면
- **THEN** NOT_DISPATCHED로 종결되고 어떤 차단도 발생하지 않는다

#### Scenario: 전송 중 크래시
- **WHEN** DISPATCH_STARTED로 기록된 attempt가 재시작 시 발견되면
- **THEN** 영속된 멱등키로 해소 절차가 시작되고 완료 전까지 같은 클래스의 신규 mutation이 차단된다

#### Scenario: 조건주문 attempt의 동일 단계
- **WHEN** 조건주문 등록 attempt가 DISPATCH_STARTED로 기록된 뒤 재시작되면
- **THEN** 같은 단계 모델의 해소 절차가 조건주문 저장소에서 수행된다
