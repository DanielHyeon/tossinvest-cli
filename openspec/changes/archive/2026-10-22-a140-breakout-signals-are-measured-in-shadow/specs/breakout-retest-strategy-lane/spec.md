## ADDED Requirements

### Requirement: breakout 파생 지표는 저장된 공식 닫힌 1분봉에서 결정적으로 만들어진다
breakout v1 의 네 파생 지표 — 봉별 `RVOLPPM` · `UpperWickRangePPM` · `VolumeExpanded` 와 스냅숏의 `ATRMinor` — 는 evidence 저장소가 봉인한 같은 세션 정규장 닫힌 1분봉 묶음만으로, 판본이 붙은 순수 함수가 정수 산술로 만들어야 한다 (SHALL).
같은 봉 묶음과 같은 판본은 언제나 같은 값을 내야 하며 (SHALL), 부동소수점 · 시계 · 네트워크 · 랭킹 · 계좌 입력을 써서는 안 된다 (MUST NOT).
입력 봉은 그 세션 정규장 개장 분부터 빈 분 없이 이어져야 하며 (SHALL), 빈 분이 있으면 typed refusal 이어야 한다 (SHALL — 봉 번호가 분과 어긋나면 opening range 와 timeout 의 뜻이 무너진다).
판본 `breakoutsignal-v1` 의 정의는 다음이어야 한다 (SHALL).
`UpperWickRangePPM = ceil((high − max(open, close)) × 1,000,000 / (high − low))` 이고 high = low 이면 0.
`ATRMinor` 는 opening range 15 봉의 Wilder ATR(14) — TR₁ = high − low, TRₖ = max(high − low, |high − 이전 close|, |low − 이전 close|), ATR₁₄ = (TR₁ + … + TR₁₄)/14, ATR₁₅ = (13 × ATR₁₄ + TR₁₅)/14 를 유리수로 계산한 뒤 minor 단위 내림 — 이고 그 세션 동안 고정이다 (SHALL — 뒤 봉이 ATR 을 바꾸면 이미 판정된 setup 이 소급 재해석된다).
`RVOLPPM` 은 opening range 뒤의 봉 i 에 대해 `floor(volume_i × 1,000,000 × 15 / (직전 15 봉 거래량 합))`, `VolumeExpanded` 는 `RVOLPPM ≥ 1,000,000`.
lane 입력 봉의 신원은 판(revision)과 무관한 `세션 ID:여는 시각(ms)` 이어야 한다 (SHALL — 봉 정정 하나로 setup 신원이 갈리지 않게), 판은 따로 넘긴다.
opening range 15 봉은 RVOL 이 정의되지 않는다: 그 봉의 RVOL 을 요청하면 typed refusal 이어야 하고 (SHALL), lane 입력에서 그 봉의 RVOL · 거래량 확장 칸은 0 · 거짓으로 두되 평가기는 그 값을 읽지 않는다 (SHALL — 평가기는 opening range 뒤 봉에서만 RVOL · 거래량 확장을 본다).
값을 만들 수 없으면 — opening range 미완성, 빈 분, 분모 0, ATR 이 0 으로 내려감, 범위를 넘는 값 — typed refusal 을 내야 하며 (SHALL), 0 · 직전 값 · 추정값으로 대신해서는 안 된다 (MUST NOT).
이 판본은 breakout v1 의 동결 상수(opening range 15 · RVOL 1,500,000 ppm · upper-wick 350,000 ppm · timeout KR 8/US 10)를 바꾸지 않는다 (MUST NOT).

#### Scenario: 직전 봉이 모자란 봉
- **WHEN** 세션의 열 번째 닫힌 봉의 RVOL 을 요청한다
- **THEN** 직전 15 봉이 없으므로 typed refusal 이 나오고 RVOL 0 은 만들어지지 않는다

#### Scenario: 같은 봉 묶음의 재계산
- **WHEN** 같은 봉 묶음으로 프로세스를 다시 시작해 지표를 다시 만든다
- **THEN** 네 지표와 그 digest 가 바이트 단위로 같다

#### Scenario: opening range 뒤에 봉이 더 쌓인다
- **WHEN** 세션의 40 번째 닫힌 봉까지 쌓인 묶음과 20 번째까지의 묶음으로 ATR 을 만든다
- **THEN** 두 ATR 은 같다(opening range 15 봉으로 고정)

#### Scenario: 빈 분이 있는 세션
- **WHEN** 봉 묶음에서 개장 뒤 어느 한 분의 봉이 빠져 있다
- **THEN** 지표 생산은 빈 분을 이름으로 든 typed refusal 이고 lane 입력은 만들어지지 않는다

#### Scenario: opening range 에 진폭이 없다
- **WHEN** opening range 15 봉의 고가 · 저가 · 종가가 모두 같다
- **THEN** ATR 은 0 이 아니라 typed refusal(ATR 0)이고 lane 입력은 만들어지지 않는다

#### Scenario: 봉 하나가 정정된다
- **WHEN** opening range 안의 봉 하나가 판 2 로 정정된다
- **THEN** lane 입력 봉 신원은 바뀌지 않아 setup 신원도 같고, 지표 digest 는 바뀐다

#### Scenario: 고가와 저가가 같은 봉
- **WHEN** high = low 인 닫힌 봉의 윗꼬리 비율을 만든다
- **THEN** 값은 0 이고 0 으로 나누기 오류나 거절이 아니다

### Requirement: breakout 그림자 신호 측정은 주문 권위가 없다
breakout 신호의 그림자 측정은 브로커에 읽기 요청만 보내야 하며 (SHALL), 주문 생성 · 정정 · 취소, 토글 · 레인 활성화 · 운영 모드 쓰기,
엔진 저널 쓰기를 해서는 안 된다 (MUST NOT). 측정의 evidence 쓰기는 자기 전용 evidence 저장소에만 해야 하고 (SHALL), 엔진이 생산 제안 재생에 쓰는
`evidence.db` 와 같은 파일(심볼릭 링크 해소 뒤)이면 시작을 거부해야 한다 (SHALL). 측정은 서명 매니페스트를 읽거나 만들지 않아야 하고
(MUST NOT), 외부 판단 API(Jev)를 호출해서는 안 된다 (MUST NOT). ARMED 에 닿은 setup 은 처음 본 분부터 최대 5 분 동안 분마다 새 호가로 다시 판정해야 하며 (SHALL), 처음 PROPOSED, 5 번째 거절, 또는 처음 본 뒤 5 분이 지남(호가 읽기가 실패해도) 중 먼저 오는 것에서 그 setup 의 판정을 끝내고 시도마다 호가 나이 · 스프레드 · 진입 괴리 실측값을 원장에 남겨야 한다 (SHALL).
측정이 만든 「냈을 제안」은 어떤 발주 경로 · 조정자 · Guardian 에도
전달되어서는 안 된다 (MUST NOT).

#### Scenario: 생산 evidence DB 경로를 준다
- **WHEN** 측정 도구의 evidence 저장소 경로가 엔진 저널 옆 `evidence.db` 를 가리킨다(직접 또는 심볼릭 링크)
- **THEN** 도구는 브로커 호출 없이 시작을 거부하고 사유를 말한다

#### Scenario: 측정 도구의 의존 금지
- **WHEN** 측정 도구 패키지의 import 와 호출 이름을 정적으로 검사한다
- **THEN** 주문 · 계좌 · 보유 · 매수 가능 · 토글 · 활성화 쓰기 이름이 0 건이다

#### Scenario: ARMED 의 첫 호가가 넓다
- **WHEN** ARMED 첫 분의 새 호가가 SPREAD_TOO_WIDE 이고 둘째 분의 새 호가는 한계 안이다
- **THEN** 둘째 분에 PROPOSED 가 판정되고 두 시도의 실측값이 원장에 남으며, 그 setup 은 다시 평가되지 않는다

#### Scenario: ARMED 에 닿은 setup
- **WHEN** 한 종목의 setup 이 ARMED 에 닿아 새 호가로 PROPOSED 가 판정된다
- **THEN** 그 결과는 원장 행으로만 남고 Guardian 판정 · 원장 예약 · 브로커 주문은 0 건이다

### Requirement: 그림자 신호 원장은 모집단과 결손을 복원 가능하게 적는다
각 신호 행은 시각 · 시장 · 종목 · setup ID · 단계 · 거절 코드 · 전이 목록 · 지표 판본 · 설정 digest · 스냅숏 digest · RVOL 반사실(1.2×/2.0×/2.5×)을 포함해야 하며 (SHALL), PROPOSED 행은 가상 진입 · 손절 · 목표 · 수량과 그 가상 사이징의 출처를 포함해야 한다 (SHALL).
라벨은 제안 시각 뒤에 연 봉만으로 매겨야 하며 (SHALL — 진입 전 가격으로 결과를 매기지 않게), 라벨 행은 신호 행을 참조하고 결과(손절 · 목표 · 마감)와 비용 포함 R 배수 또는 라벨 결손 사유를 포함해야 한다 (SHALL). 봉 · 호가 · 랭킹 읽기
포기, 세션 · 달력 결손은 결손 행으로 남아야 하며 (SHALL), 조용히 생략되어서는 안 된다 (MUST NOT). 보고서는 표본이 30 미만인 결과 구간의
판정을 보류해야 한다 (SHALL).

#### Scenario: 한 분의 봉 읽기가 rate limit 으로 포기된다
- **WHEN** 한 종목의 그 분 봉 읽기가 재시도 한도를 넘어 포기된다
- **THEN** 그 종목 · 그 분의 결손 행이 원장에 남고 다음 분의 평가는 계속된다

#### Scenario: 제안이 몇 건 없는 측정
- **WHEN** 측정 기간 동안 PROPOSED 가 12 건이다
- **THEN** 보고서는 퍼널과 거절 분포는 보이되 승률 · 평균 R 의 판정은 표본 부족으로 보류한다
