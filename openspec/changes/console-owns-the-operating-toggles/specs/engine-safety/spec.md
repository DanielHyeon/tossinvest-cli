# engine-safety Specification (delta)

선행 change `interlock-gates-entry-not-exit`(커밋 b50f991, 미archive)의 본문을 이어받아
3절만 개정한다.

## MODIFIED Requirements

### Requirement: 자동화 게이트 기동 인터록
자동 주문 게이트는 기본 OFF이며(SHALL), 게이트 ON 설정 시 다음이 모두 검증되지 않으면 기동을 거부한다(SHALL):

1. 필수 한도 전부가 명시적으로 설정되고 양수·유한하며 통화 일치 — 주문 수량, 주문 notional, 총 개방 노출, 일일 손실 절대액, 일일 손실 자본 비율 중 **하나라도** 누락·0·NaN·Inf이면 거부(부분적으로 무제한인 게이트는 허가된 게이트가 아니다)
2. 유효한 capability attestation(만료·계좌 식별·성공 endpoint 집합 — verify-execution-capability change가 생성) 존재·미만료·계좌 일치. attestation endpoint 집합은 엔진 자동 경로가 실제 사용하는 호출 전부와 drift 가드로 동기화한다(SHALL — 목록을 확장하는 change는 가드를 함께 갱신한다)
3. 거래 정책이 **청산 경로가 실제로 요구하는 것 전부**를 허용 — `place`·`sell`·`cancel`·`allow_live_order_actions`(SHALL). 매수는 가능한데 청산이 불가능한 조합으로는 기동할 수 없다(SHALL NOT)는 이 절이 처음부터 주장한 것이고, 검사 대상이 그 주장에 미치지 못했다 — 매도 제출도 `place`이며(`internal/trading` 정책 검사가 side와 무관하게 요구한다), exit 관측 루프는 자기 발의를 취소하므로 `cancel`도 쓴다. 둘이 꺼진 채 기동하면 엔진은 뜨고 **첫 손절에서 거부된다**. 검사에 넣지 않은 요구는 요구가 아니다
4. Guardian이 인터록이 감사한 설정 한도와 같은 출처에서 구성됨 — 동등성 검증은 EXPOSURE_RAISING 결정의 한도에만 적용
5. 엔진 프로필에 ExecutionGateway가 구성됨(round-trip용 주문 조회 배선 포함)

게이트 flip은 사람 승인 절차(§0.7)와 audit 기록을 요구한다(SHALL).

**브로커측 보호 실행 배선(ProtectionReady)은 기동 조건이 아니라 진입 허가 조건이다**(SHALL).
조항 1~5가 기동을 결정하고, ProtectionReady는 기동한 런타임에서 **무엇이 허가되는지**를 결정한다.
이 분리의 근거는 실패의 비대칭이다 — 진입 후 프로세스가 죽으면 손절 없는 **새** 포지션이 남고,
청산 전에 죽으면 보호 미배선 상태의 **기존** 포지션이 남는다. 후자는 런타임을 거부해도 동일하게
발생하는 상태이므로, 그것을 이유로 청산 루프를 거부하는 것은 같은 결함을 더 넓은 범위에 유지한다.

ProtectionReady 미충족 프로필에서:

- 게이트 ON + 조항 1~5 충족이면 런타임은 **기동한다**(SHALL). 루프 집합은 변하지 않는다 —
  reconcile driver·exit observer·체결 감지.
- 노출을 **증가**시키는 mutation은 거부된다(SHALL). 집행은 mutation chokepoint에서 이루어지며,
  판정 근거는 호출자가 선언한 Safety Class가 아니라 **mutation 자체의 형태에서 계산된 사실**이다
  (매수 여부). 클래스를 잘못 붙여 우회할 수 없다.
- 집행 지점은 **하나**여야 한다(SHALL — 두 곳에서 판정하면 답을 틀릴 곳이 두 곳이 된다).
- ProtectionReady 표지는 config 키·Options 필드가 아닌 **컴파일 타임 상수**로만 존재한다
  (SHALL NOT — 설정으로 만족시킬 수 있는 표지는 짓지 않고 준비됐다고 주장하는 길이다).
  보호주문 도입 change가 자기 작업의 마지막 단계로 이 상수를 뒤집는다.
- 운영자에게 보호가 **프로세스 수명에 묶여 있다**는 사실이 기동 시점에 전달되어야 한다(SHALL —
  읽는 문장이며, 타이핑 확인·추가 승인 마찰이어서는 안 된다(SHALL NOT)).

#### Scenario: attestation 만료 상태 기동
- **WHEN** 게이트 ON + attestation 만료 상태로 기동하면
- **THEN** 기동이 거부되고 재검증 안내가 출력된다

#### Scenario: 한도 일부만 설정
- **WHEN** 주문 수량 한도만 양수이고 총 개방 노출 한도가 설정되지 않은 상태로 기동하면
- **THEN** 기동이 거부된다

#### Scenario: 청산 불가 정책으로 기동
- **WHEN** 매도가 비활성인 거래 정책으로 게이트 ON 기동하면
- **THEN** 기동이 거부된다

#### Scenario: 한도 출처 불일치
- **WHEN** 주입된 Guardian이 인터록이 검증한 설정 한도와 다른 한도로 EXPOSURE_RAISING 결정을 찍으면
- **THEN** 기동(또는 해당 결정)이 거부된다

#### Scenario: 보호 미배선 기동
- **WHEN** ProtectionReady 미충족 프로필에서 게이트 ON + 조항 1~5 충족으로 기동하면
- **THEN** 런타임이 기동하고, 보호가 프로세스 수명에 묶여 있다는 사실이 출력되며, 상태 보고의 protection은 UNWIRED로 남는다

#### Scenario: 보호 미배선에서 노출 증가 mutation
- **WHEN** ProtectionReady 미충족 상태에서 매수 mutation이 제출되면
- **THEN** 거부되고 사유가 보호 미배선으로 열거된다 — 결정이 EXPOSURE_RAISING으로 찍혀 있든 아니든 판정은 mutation의 형태에서 나온다

#### Scenario: 보호 미배선에서 노출 감소 mutation
- **WHEN** ProtectionReady 미충족 상태에서 매도 mutation이 제출되면
- **THEN** 통과한다 — 청산은 이 조항이 지목하는 실패를 만들지 않는다

#### Scenario: 청산에 필요한 토글 누락
- **WHEN** `trading.place` 또는 `trading.cancel`이 꺼진 채로 게이트 ON 기동하면
- **THEN** 기동이 거부되고 꺼진 토글이 이름으로 열거된다 — 손절을 낼 수 없는 구성으로 기동하지 않는다
