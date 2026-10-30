## ADDED Requirements

### Requirement: 초기 데이터 종류 투영
프로브는 현재 공개 상태로부터 initial-fields-v1 입력을 생성해야 한다(SHALL). 기본 가격·랭킹·거래량·최상단 호가·종목·시각·세션 정보만 유지하고 확장 데이터 및 그 오류 흔적을 제외해야 한다(SHALL). 뉴스 미제공은 고정 문구로 표시하며 전체 원본 상태를 변경해서는 안 된다(MUST NOT).

#### Scenario: 모든 확장 자료가 있는 상태
- **WHEN** initial 입력을 생성한다
- **THEN** HTTP 요청에 분봉·일봉·시장 맥락·뉴스 내용·종목 수급·경고 유형·가격제한 블록이 없고 원본은 보존된다

#### Scenario: 기본 자료와 확장 자료에 결측이 함께 있다
- **WHEN** initial 입력을 생성한다
- **THEN** 기본 자료의 실제 결측과 의도적 뉴스 미제공만 전달되고 확장 자료 조회 실패는 전달되지 않는다

### Requirement: 자동 비교 실행과 운영 원장 격리
프로브는 compare를 기본값으로 하여 별도 옵션 없이 current와 initial을 동일 회차에 측정해야 한다(SHALL). 명시적 current 및 initial 단독 모드를 지원해야 한다(SHALL). 출력 생략 시 compare·initial은 각각 별도 자동 경로를 사용하고 current는 옛 기본 경로를 유지해야 한다(SHALL). 잘못된 mode 및 initial/compare의 옛 운영 출력 경로 사용은 외부 호출·쓰기 전에 거부해야 한다(SHALL).

#### Scenario: 평소 실행 명령을 그대로 사용한다
- **WHEN** input-profile과 out을 지정하지 않는다
- **THEN** 같은 수집 상태에서 current와 initial을 자동 판단하고 shadow-ledger-compare에 저장한다

#### Scenario: 현재 입력만 명시적으로 선택한다
- **WHEN** input-profile current를 지정하고 out을 생략한다
- **THEN** 기존 shadow-ledger 경로·입력·판단 ID·직렬화 계약이 유지된다

#### Scenario: 초기 입력만 명시적으로 선택한다
- **WHEN** input-profile initial을 지정하고 out을 생략한다
- **THEN** 초기 입력만 판단하여 shadow-ledger-initial에 저장한다

#### Scenario: 비교 모드를 운영 원장에 출력하려 한다
- **WHEN** compare 또는 initial이 기본 출력 경로 또는 그 별칭을 지정한다
- **THEN** 모델·브로커 호출과 원장 변경 없이 오류가 반환된다

### Requirement: 동일 스냅숏 비교와 실패 보존
compare는 동일 수집 결과로 current와 initial을 판단해야 한다(SHALL). 종목·시각·기준가·질문·모델·horizon을 고정하고 screening과 실패를 profile별로 기록해야 한다(SHALL). 실패 응답을 다른 군의 성공으로 대체해서는 안 된다(MUST NOT).

#### Scenario: 양쪽 판단 성공
- **WHEN** 동일 후보의 두 profile 판단이 성공한다
- **THEN** 두 판단은 동일 조건을 공유하고 서로 다른 입력 digest·profile·충돌 없는 ID로 기록된다

#### Scenario: 한쪽 판단 실패
- **WHEN** initial 또는 current만 실패한다
- **THEN** 성공군과 실패군을 모두 추적할 수 있고 정상 paired 표본으로 세지 않는다

#### Scenario: 두 군 모두 판단 실패
- **WHEN** 같은 스냅숏의 current와 initial이 모두 실패한다
- **THEN** 각 gap의 cohort·profile·source digest·관측 식별자로 하나의 비교 시도와 두 실패를 셀 수 있다

### Requirement: 입력 추적과 공유 라벨
원장은 실제 전송 상태 digest와 전체 source 상태 digest를 분리하고 비교 cohort와 profile을 식별해야 한다(SHALL). 동일 관측의 결과 가격은 profile 간 공유하며 재시작 이후에도 판단 ID 연결을 보존해야 한다(SHALL).

#### Scenario: 초기 입력에는 분봉이 없다
- **WHEN** 초기 입력군의 사후 가격을 평가한다
- **THEN** source 상태의 분봉을 사용하고 전송 상태는 별도로 검증한다

#### Scenario: 오전 신호 이후의 회차에 청산 가격이 관측된다
- **WHEN** 후속 회차 source에 필요한 청산 분봉이 있고 그 회차의 판단은 실패한다
- **THEN** 저장된 source와 gap 연결로 가격을 복원하고 오전 신호 source 하나의 미래 가격 부재를 전체 가격 결측으로 오인하지 않는다

#### Scenario: 라벨 기한 전에 재시작
- **WHEN** 두 profile의 판단이 미완료 상태로 복원된다
- **THEN** 동일 기준가와 기한을 보존하고 한 번 조회한 결과 가격을 각 판단에 연결한다

### Requirement: 혼합 없는 비교 보고
보고서는 profile별 모집단·실패·라벨 결측을 구분하고 동일 비교 조건의 쌍만 판단 차이에 사용해야 한다(SHALL). J2 p>=0.6 교차·확률 차이·선택 성과를 표시하되 날짜/종목 중복·가격 결측·비용 가정을 명시해야 한다(SHALL). 기존 사전등록 판정을 소급 변경해서는 안 된다(MUST NOT).

#### Scenario: 정상 쌍과 단독 행과 중복 행이 섞인다
- **WHEN** 비교 보고서를 생성한다
- **THEN** 정상 쌍만 paired 분석에 포함하고 나머지 수와 이유를 별도로 표시한다

#### Scenario: 옛 원장만 보고한다
- **WHEN** profile 메타데이터가 없는 기존 행을 읽는다
- **THEN** 기존 current 보고는 유지되고 비교 실험이 실행된 것으로 해석하지 않는다

#### Scenario: 기본 보고서를 요청한다
- **WHEN** report를 ledger 옵션 없이 실행한다
- **THEN** 자동 비교 원장을 읽고 입력군별 calibration 및 paired 요약을 출력한다

### Requirement: 오프라인 검증과 운영 불변
구현 검증은 대상 도구 테스트 및 합성 HTTP/원장 fixture로 수행해야 한다(SHALL). 이 변경의 검증 과정에서 실제 모델·브로커 API 또는 운영 설정 변경을 실행해서는 안 된다(MUST NOT).

#### Scenario: 구현 검증과 종료
- **WHEN** Terra가 구현·리뷰·종결을 수행한다
- **THEN** 대상 테스트와 Manager 검토 근거를 남기고 전체 저장소 gate·운영 실행 여부를 정확히 구분한다
