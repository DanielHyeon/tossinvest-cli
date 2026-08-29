# engine-safety — delta (interlock-gates-entry-not-exit)

## MODIFIED Requirements

### Requirement: 자동화 게이트 기동 인터록
자동 주문 게이트는 기본 OFF이며(SHALL), 게이트 ON 설정 시 다음이 모두 검증되지 않으면 기동을 거부한다(SHALL):

1. 필수 한도 전부가 명시적으로 설정되고 양수·유한하며 통화 일치 — 주문 수량, 주문 notional, 총 개방 노출, 일일 손실 절대액, 일일 손실 자본 비율 중 **하나라도** 누락·0·NaN·Inf이면 거부(부분적으로 무제한인 게이트는 허가된 게이트가 아니다)
2. 유효한 capability attestation(만료·계좌 식별·성공 endpoint 집합 — verify-execution-capability change가 생성) 존재·미만료·계좌 일치. attestation endpoint 집합은 엔진 자동 경로가 실제 사용하는 호출 전부와 drift 가드로 동기화한다(SHALL — 목록을 확장하는 change는 가드를 함께 갱신한다)
3. 거래 정책이 매도와 실주문 실행을 허용 — 매수는 가능한데 청산이 불가능한 조합으로는 기동할 수 없다(SHALL NOT)
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

### Requirement: 엔진 런타임 수명주기

엔진 런타임(`tossctl engine run`)의 루프 집합은 **reconcile driver·exit observer·체결 감지(= `filldetect.Detector` 폴링 루프)**다(SHALL — 체결 감지 없는 런타임은 발의 pending이 영구 미해소로 남아 exit 수명주기 계약을 위반한다; 힌트 라우팅(`Hints`)을 포함하는 경우 Refresh 미배선은 감독이 아니라 **조립 시점 검증**으로 거부한다(SHALL); exit 관측의 SLO 양보 지점은 엔진 배선의 어댑터로 체결 감지 상태에 연결한다). 기동 순서(SHALL): ① **journal 디렉터리 flock 획득이 기동의 첫 동작이다** — 실패(다른 인스턴스 보유)면 즉시 거부한다(SHALL — journal은 단일 writer 설계이고, flock이 첫 동작이어야 journal open·마이그레이션 전체가 배타 안에 들어온다; 자문 마커로는 경합이 닫히지 않는다). ② 게이트 OFF면 기동할 루프 집합이 없으므로 거부한다(기동 인터록은 게이트 ON에만 정의된다). ③ 게이트 ON이면 기동 인터록 검증을 소비하고, 미충족이면 인터록이 반환한 미충족 항목을 열거하며 거부한다(fail-closed). ④ verify runlock이 신선하면 거부한다. 이와 별도로 **엔진 활성 마커**(갱신 1분·stale 5분 — runlock 선례 수치)를 유지해 콘솔의 엔진 상태 표시·autostart의 사전 확인이 소비한다(SHALL — 자문 신호이며 배타는 flock이 담당함을 명시).

③의 "미충족 항목"은 기동 인터록 조항 1~5다(SHALL). ProtectionReady는 기동을 거부하지 않으므로 이 열거에 들어가지 않으며, 대신 기동 출력과 상태 보고가 그 상태를 표시한다.

감독 계약은 두 층이다(SHALL): ① **방어적 종료 계약** — 루프가 컨텍스트 취소 외의 사유로 반환하면 전체 런타임이 정지하고 critical 알림이 발송된다. 컨텍스트 취소에 의한 반환은 정상 종료이며 critical을 발송하지 않는다(SHALL NOT). ② **지속 열화 임계** — 루프가 살아 있으나 사이클이 연속 실패하는 상태를 각 루프에 정의한다: exit 관측은 landed 60초 두절 계약 유지, reconcile driver와 체결 감지는 연속 5주기 실패 시 critical 알림 + ENTRY_BLOCKED 자동 강화(SHALL — 루프는 계속 재시도한다).

종료 시그널은 루프 취소·완주 대기·journal 정합 close로 처리하고(SHALL), 두 번째 시그널은 즉시 종료한다. 재기동 복구는 landed 계약(pending 복원·편입 완결·nonce 재사용 금지)을 소비하며 새 복구 경로를 만들지 않는다(SHALL NOT).

#### Scenario: 게이트 OFF 기동
- **WHEN** 게이트 OFF 상태로 `engine run`을 실행하면
- **THEN** "기동할 루프 집합이 없다(게이트 OFF)"로 거부되고 실패 종료한다 — 인터록 조항 열거는 없다

#### Scenario: 게이트 ON + 인터록 미충족 기동
- **WHEN** 게이트 ON + 조항 1~5 중 하나가 미충족인 상태로 `engine run`을 실행하면
- **THEN** 루프가 하나도 시작되지 않고 인터록의 미충족 항목이 열거되며 실패 종료한다

#### Scenario: 게이트 ON + 보호 미배선 기동
- **WHEN** 게이트 ON + 조항 1~5 충족 + ProtectionReady 미충족으로 `engine run`을 실행하면
- **THEN** 세 루프가 전부 기동하고, 보호가 이 프로세스가 사는 동안만 유효하다는 사실이 출력된다

#### Scenario: 두 번째 인스턴스 기동
- **WHEN** 엔진이 실행 중인 머신에서 `engine run`(또는 autostart·콘솔 버튼)이 다시 실행되면
- **THEN** 실행 중 인스턴스가 안내되고 기동이 거부된다

#### Scenario: 루프의 비정상 반환
- **WHEN** 기동된 루프 중 하나가 컨텍스트 취소가 아닌 사유로 반환하면
- **THEN** 나머지 루프도 정지하고 critical 알림이 발송되며 프로세스가 실패로 종료한다
