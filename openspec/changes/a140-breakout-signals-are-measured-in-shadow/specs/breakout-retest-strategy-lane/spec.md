## ADDED Requirements

### Requirement: breakout 파생 지표는 저장된 공식 닫힌 1분봉에서 결정적으로 만들어진다
breakout v1 의 네 파생 지표 — 봉별 `RVOLPPM` · `UpperWickRangePPM` · `VolumeExpanded` 와 스냅숏의 `ATRMinor` — 는 evidence 저장소가 봉인한 같은 세션 정규장 닫힌 1분봉 묶음만으로, 판본이 붙은 순수 함수가 정수 산술로 만들어야 한다 (SHALL). 같은 봉 묶음과 같은 판본은 언제나 같은
값을 내야 하며 (SHALL), 부동소수점 · 시계 · 네트워크 · 랭킹 · 계좌 입력을 써서는 안 된다 (MUST NOT). 판본 `breakoutsignal-v1` 의 정의는
다음이어야 한다 (SHALL): `UpperWickRangePPM = ceil((high − max(open, close)) × 1,000,000 / (high − low))` 이고 high = low 이면 0,
`ATRMinor` 는 같은 세션 닫힌 1분봉의 Wilder ATR(14)(첫 봉의 TR = high − low, minor 단위 내림), `RVOLPPM` 은 그 봉 거래량 × 1,000,000 을
같은 세션 직전 15 봉 평균 거래량으로 나눈 값의 내림, `VolumeExpanded` 는 `RVOLPPM ≥ 1,000,000`.
값을 만들 수 없으면 — ATR 에 봉 14 개 미만, RVOL 에 직전 봉 15 개 미만, 분모 0, 범위를 넘는 값 — typed refusal 을 내야 하며 (SHALL),
0 · 직전 값 · 추정값으로 대신해서는 안 된다 (MUST NOT). 이 판본은 breakout v1 의 동결 상수(opening range 15 · RVOL 1,500,000 ppm ·
upper-wick 350,000 ppm · timeout KR 8/US 10)를 바꾸지 않는다 (MUST NOT).

#### Scenario: 직전 봉이 모자란 봉
- **WHEN** 세션의 열 번째 닫힌 봉의 RVOL 을 요청한다
- **THEN** 직전 15 봉이 없으므로 typed refusal 이 나오고 RVOL 0 은 만들어지지 않는다

#### Scenario: 같은 봉 묶음의 재계산
- **WHEN** 같은 봉 묶음으로 프로세스를 다시 시작해 지표를 다시 만든다
- **THEN** 네 지표와 그 digest 가 바이트 단위로 같다

#### Scenario: 고가와 저가가 같은 봉
- **WHEN** high = low 인 닫힌 봉의 윗꼬리 비율을 만든다
- **THEN** 값은 0 이고 0 으로 나누기 오류나 거절이 아니다

### Requirement: breakout 그림자 신호 측정은 주문 권위가 없다
breakout 신호의 그림자 측정은 브로커에 읽기 요청만 보내야 하며 (SHALL), 주문 생성 · 정정 · 취소, 토글 · 레인 활성화 · 운영 모드 쓰기,
엔진 저널 쓰기를 해서는 안 된다 (MUST NOT). 측정은 자기 전용 evidence 저장소에만 써야 하고 (SHALL), 엔진이 생산 제안 재생에 쓰는
`evidence.db` 와 같은 파일(심볼릭 링크 해소 뒤)이면 시작을 거부해야 한다 (SHALL). 측정은 서명 매니페스트를 읽거나 만들지 않아야 하고
(MUST NOT), 외부 판단 API(Jev)를 호출해서는 안 된다 (MUST NOT). 측정이 만든 「냈을 제안」은 어떤 발주 경로 · 조정자 · Guardian 에도
전달되어서는 안 된다 (MUST NOT).

#### Scenario: 생산 evidence DB 경로를 준다
- **WHEN** 측정 도구의 evidence 저장소 경로가 엔진 저널 옆 `evidence.db` 를 가리킨다(직접 또는 심볼릭 링크)
- **THEN** 도구는 브로커 호출 없이 시작을 거부하고 사유를 말한다

#### Scenario: 측정 도구의 의존 금지
- **WHEN** 측정 도구 패키지의 import 와 호출 이름을 정적으로 검사한다
- **THEN** 주문 · 계좌 · 보유 · 매수 가능 · 토글 · 활성화 쓰기 이름이 0 건이다

#### Scenario: ARMED 에 닿은 setup
- **WHEN** 한 종목의 setup 이 ARMED 에 닿아 새 호가로 PROPOSED 가 판정된다
- **THEN** 그 결과는 원장 행으로만 남고 Guardian 판정 · 원장 예약 · 브로커 주문은 0 건이다

### Requirement: 그림자 신호 원장은 모집단과 결손을 복원 가능하게 적는다
각 신호 행은 시각 · 시장 · 종목 · setup ID · 단계 · 거절 코드 · 전이 목록 · 지표 판본 · 설정 digest · 스냅숏 digest · RVOL 반사실(1.2×/2.0×/2.5×)을 포함해야 하며 (SHALL), PROPOSED 행은 가상 진입 · 손절 · 목표 · 수량과 그 가상 사이징의 출처를 포함해야 한다 (SHALL).
라벨 행은 신호 행을 참조하고 결과(손절 · 목표 · 마감)와 비용 포함 R 배수 또는 라벨 결손 사유를 포함해야 한다 (SHALL). 봉 · 호가 · 랭킹 읽기
포기, 세션 · 달력 결손은 결손 행으로 남아야 하며 (SHALL), 조용히 생략되어서는 안 된다 (MUST NOT). 보고서는 표본이 30 미만인 결과 구간의
판정을 보류해야 한다 (SHALL).

#### Scenario: 한 분의 봉 읽기가 rate limit 으로 포기된다
- **WHEN** 한 종목의 그 분 봉 읽기가 재시도 한도를 넘어 포기된다
- **THEN** 그 종목 · 그 분의 결손 행이 원장에 남고 다음 분의 평가는 계속된다

#### Scenario: 제안이 몇 건 없는 측정
- **WHEN** 측정 기간 동안 PROPOSED 가 12 건이다
- **THEN** 보고서는 퍼널과 거절 분포는 보이되 승률 · 평균 R 의 판정은 표본 부족으로 보류한다
